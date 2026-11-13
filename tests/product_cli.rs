mod common;

use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    process::{Command, Stdio},
};
use tempfile::TempDir;
/// v2 profile input (REQ-03A-01/08): `project-plan` accepts schema_version 2 only;
/// v1 files stay readable through `read_profile` with defaulted new fields.
fn profile(pin: Option<String>) -> Value {
    let mut v = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
    if let Some(pin) = pin {
        v["package_pin"] = json!(pin);
    }
    v
}
fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    // The child may validate its arguments and exit before reading stdin; on a fast
    // runner that surfaces as EPIPE here. The tests assert the child's exit status and
    // output, never that the stdin write completed, so only a broken pipe is tolerated.
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
/// Result of an internal operation: asserts the success envelope (01-C #1990) and unwraps it.
fn ok_result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}
#[test]
fn stdin_broken_pipe_does_not_panic() {
    // `package-inspect` never reads stdin and fails fast on a missing path, so a
    // payload larger than the pipe buffer reproduces the runner-only EPIPE.
    let big = json!({"pad": "x".repeat(256 * 1024)});
    let out = cli(&["package-inspect", "/nonexistent/moai-rc-pipe-test"], &big);
    assert!(!out.status.success());
}
fn create(root: &TempDir, input: &Value) -> Value {
    let p = root.path().to_str().unwrap();
    let preview = cli(&["project-plan", p], input);
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
    plan
}
#[test]
fn create_repeat_and_unknown_user_keys_preserved() {
    let root = tempfile::Builder::new()
        .prefix("moai project ")
        .tempdir()
        .unwrap();
    let input = profile(None);
    create(&root, &input);
    let file = root.path().join(".moai-cli/project.yaml");
    let mut value: Value = serde_json::from_slice(&fs::read(&file).unwrap()).unwrap();
    value["custom"] = json!({"user":true});
    let mut bytes = serde_json::to_vec_pretty(&value).unwrap();
    bytes.push(b'\n');
    fs::write(&file, &bytes).unwrap();
    // 03A-AC-4: the identical input plans `unchanged`, and the second apply leaves the
    // file — unknown user key included — byte-for-byte identical.
    let second = ok_result(&cli(
        &["project-plan", root.path().to_str().unwrap()],
        &input,
    ));
    assert_eq!(second["action"], "unchanged");
    assert!(second["delta"]["conflict"].as_array().unwrap().is_empty());
    create(&root, &input);
    assert_eq!(fs::read(&file).unwrap(), bytes);
}

fn apply_result(root: &TempDir, plan: &Value) -> Value {
    let out = cli(
        &[
            "project-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan,
    );
    assert!(out.status.success(), "{:?}", out.stderr);
    ok_result(&out)
}

fn v2_input() -> Value {
    let mut v = profile(None);
    v["decisions"] = json!([
        {"id":"d1","topic":"tools","value":"Rust","source":"user","scope":"project"}
    ]);
    v
}

fn json_file(path: &std::path::Path) -> Value {
    serde_json::from_slice(&fs::read(path).unwrap()).unwrap()
}

fn write_json(path: &std::path::Path, value: &Value) -> Vec<u8> {
    let mut bytes = serde_json::to_vec_pretty(value).unwrap();
    bytes.push(b'\n');
    fs::write(path, &bytes).unwrap();
    bytes
}

#[test]
fn profile_v2_create_and_secret_rejection() {
    // 03A-AC-1: a v2 profile creates the file with the new keys, and a secret-shaped
    // decision value is refused and echoed nowhere.
    let root = TempDir::new().unwrap();
    let input = v2_input();
    let plan = ok_result(&cli(
        &["project-plan", root.path().to_str().unwrap()],
        &input,
    ));
    assert_eq!(plan["action"], "create");
    let added = plan["delta"]["added"].as_array().unwrap();
    assert!(added.contains(&json!("decisions")) && added.contains(&json!("interview_revision")));
    assert_eq!(added.len(), 8);
    let result = apply_result(&root, &plan);
    assert_eq!(result["state"], "created");
    assert_eq!(result["interview_revision"], json!("rc1-fixture-1"));
    assert_eq!(result["host_verified"], json!(false));
    assert!(result["next_actions"][0]
        .as_str()
        .unwrap()
        .contains("generation.json"));
    let file = root.path().join(".moai-cli/project.yaml");
    let value = json_file(&file);
    assert_eq!(value["schema_version"], json!(2));
    assert_eq!(value["interview_revision"], json!("rc1-fixture-1"));
    assert_eq!(value["decisions"][0]["id"], json!("d1"));
    let baseline = json_file(
        &root
            .path()
            .join(".moai-cli/manifests/profile-baseline.json"),
    );
    assert_eq!(baseline["schema_version"], json!(1));
    assert_eq!(
        baseline["profile"]["interview_revision"],
        json!("rc1-fixture-1")
    );
    let mut secret = v2_input();
    secret["decisions"] = json!([
        {"id":"d2","topic":"credentials","value":"password=fixture-only","source":"user","scope":"project"}
    ]);
    let out = cli(&["project-plan", root.path().to_str().unwrap()], &secret);
    assert!(!out.status.success());
    let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(failure["error"]["kind"], "invalid_input");
    assert!(!String::from_utf8_lossy(&out.stdout).contains("password=fixture-only"));
    assert!(!String::from_utf8_lossy(&out.stderr).contains("password=fixture-only"));
}

#[test]
fn profile_reconfigure_delta_preserves_user_keys() {
    // 03A-AC-2: a reconfigure that changes other keys applies as `update`; the
    // user-edited key and the unknown user key survive with their values intact.
    let root = TempDir::new().unwrap();
    let input = v2_input();
    create(&root, &input);
    let file = root.path().join(".moai-cli/project.yaml");
    let mut value = json_file(&file);
    value["purpose"] = json!("Build a better editor"); // user edit away from the baseline
    value["custom"] = json!({"user": true});
    write_json(&file, &value);
    let mut next = input.clone(); // purpose stays at the baseline value; tools change
    next["tools"] = json!(["Python"]);
    let plan = ok_result(&cli(
        &["project-plan", root.path().to_str().unwrap()],
        &next,
    ));
    assert_eq!(plan["action"], "update");
    assert_eq!(plan["delta"]["changed"], json!(["tools"]));
    assert!(plan["delta"]["conflict"].as_array().unwrap().is_empty());
    assert!(plan["delta"]["unchanged"]
        .as_array()
        .unwrap()
        .contains(&json!("purpose")));
    apply_result(&root, &plan);
    let after = json_file(&file);
    assert_eq!(after["tools"], json!(["Python"]));
    assert_eq!(after["purpose"], json!("Build a better editor"));
    assert_eq!(after["custom"], json!({"user": true}));
    assert!(String::from_utf8_lossy(&fs::read(&file).unwrap()).contains("Build a better editor"));
    // The refreshed baseline makes the same input `unchanged` on the next run.
    let again = ok_result(&cli(
        &["project-plan", root.path().to_str().unwrap()],
        &next,
    ));
    assert_eq!(again["action"], "unchanged");
}

#[test]
fn profile_conflict_writes_nothing() {
    // 03A-AC-3: a key the user changed and the input changes too is a conflict; apply
    // writes nothing and the file hash is identical before and after.
    let root = TempDir::new().unwrap();
    create(&root, &v2_input());
    let file = root.path().join(".moai-cli/project.yaml");
    let mut value = json_file(&file);
    value["purpose"] = json!("User choice");
    write_json(&file, &value);
    let mut clashing = v2_input();
    clashing["purpose"] = json!("Different purpose");
    let plan = ok_result(&cli(
        &["project-plan", root.path().to_str().unwrap()],
        &clashing,
    ));
    assert_eq!(plan["action"], "conflict");
    assert_eq!(plan["delta"]["conflict"], json!(["purpose"]));
    let before = fs::read(&file).unwrap();
    let baseline_before = fs::read(
        root.path()
            .join(".moai-cli/manifests/profile-baseline.json"),
    )
    .unwrap();
    let out = cli(
        &[
            "project-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        &plan,
    );
    assert!(!out.status.success());
    let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(failure["error"]["kind"], "conflict");
    // REQ-03A-06: no input value is echoed in the refusal.
    assert!(!String::from_utf8_lossy(&out.stdout).contains("User choice"));
    assert!(!String::from_utf8_lossy(&out.stderr).contains("Different purpose"));
    assert_eq!(fs::read(&file).unwrap(), before);
    assert_eq!(
        fs::read(
            root.path()
                .join(".moai-cli/manifests/profile-baseline.json")
        )
        .unwrap(),
        baseline_before
    );
}

#[test]
fn general_yaml_is_preserved_by_refusal() {
    // 03A-AC-5 / REQ-03A-04: until the lossless-editing HOLD is decided, a general YAML
    // file (comments, block style) is read-only: plans refuse and the file is preserved.
    let root = TempDir::new().unwrap();
    let dir = root.path().join(".moai-cli");
    fs::create_dir_all(&dir).unwrap();
    let file = dir.join("project.yaml");
    let yaml = b"# my comment\npurpose: keep me\nhosts:\n  - claude\n";
    fs::write(&file, yaml).unwrap();
    let out = cli(
        &["project-plan", root.path().to_str().unwrap()],
        &v2_input(),
    );
    assert!(!out.status.success());
    let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(failure["error"]["kind"], "conflict");
    assert!(String::from_utf8_lossy(&out.stdout).contains("preservation review"));
    assert_eq!(fs::read(&file).unwrap(), yaml);
}

#[test]
fn profile_v1_file_still_readable_under_v2() {
    // REQ-03A-01: v1 files stay readable with defaulted new fields, and a matching v2
    // input plans the upgrade instead of stalling on the schema marker (audit #2052).
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    let file = root.path().join(".moai-cli/project.yaml");
    fs::create_dir_all(file.parent().unwrap()).unwrap();
    let v1 = json!({
        "schema_version": 1, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"]
    });
    let bytes = write_json(&file, &v1);
    // Readable through the v2 reader: agents-plan succeeds on the v1 file.
    assert!(cli(
        &["agents-plan", root.path().to_str().unwrap()],
        &agents_request("en", "none")
    )
    .status
    .success());
    let plan = ok_result(&cli(
        &["project-plan", root.path().to_str().unwrap()],
        &v2_input(),
    ));
    assert_eq!(plan["action"], "update");
    assert!(plan["delta"]["conflict"].as_array().unwrap().is_empty());
    assert!(plan["delta"]["changed"]
        .as_array()
        .unwrap()
        .contains(&json!("schema_version")));
    assert!(plan["delta"]["added"]
        .as_array()
        .unwrap()
        .contains(&json!("interview_revision")));
    // The plan is read-only: the v1 file waits for the apply.
    assert_eq!(fs::read(&file).unwrap(), bytes);
}

#[test]
fn profile_v1_file_without_baseline_upgrades_to_v2() {
    // Audit #2052 FAIL fix (leader decision a): schema_version is MoAI-owned, so a v1
    // file with no baseline record upgrades to v2 as an `update`. Other keys keep the
    // no-baseline rule — a value the input would change stays a conflict.
    let root = TempDir::new().unwrap();
    let file = root.path().join(".moai-cli/project.yaml");
    fs::create_dir_all(file.parent().unwrap()).unwrap();
    let v1 = json!({
        "schema_version": 1, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"]
    });
    write_json(&file, &v1);
    let input = v2_input();
    let plan = ok_result(&cli(
        &["project-plan", root.path().to_str().unwrap()],
        &input,
    ));
    assert_eq!(plan["action"], "update");
    assert!(plan["delta"]["conflict"].as_array().unwrap().is_empty());
    let result = apply_result(&root, &plan);
    assert_eq!(result["state"], "updated");
    let after = json_file(&file);
    assert_eq!(after["schema_version"], json!(2));
    assert_eq!(after["interview_revision"], json!("rc1-fixture-1"));
    assert_eq!(after["purpose"], json!("Build a small editor"));
    assert_eq!(after["tools"], json!(["Rust"]));
    assert_eq!(after["decisions"][0]["id"], json!("d1"));
    // The apply recorded the baseline, so the same input is `unchanged` afterwards.
    let again = ok_result(&cli(
        &["project-plan", root.path().to_str().unwrap()],
        &input,
    ));
    assert_eq!(again["action"], "unchanged");
    // A v1 value the input would change is still a conflict without a baseline; the
    // schema marker alone never unlocks the other keys.
    let root2 = TempDir::new().unwrap();
    let file2 = root2.path().join(".moai-cli/project.yaml");
    fs::create_dir_all(file2.parent().unwrap()).unwrap();
    write_json(&file2, &v1);
    let mut clashing = v2_input();
    clashing["purpose"] = json!("Different purpose");
    let plan = ok_result(&cli(
        &["project-plan", root2.path().to_str().unwrap()],
        &clashing,
    ));
    assert_eq!(plan["action"], "conflict");
    assert_eq!(plan["delta"]["conflict"], json!(["purpose"]));
    assert!(String::from_utf8_lossy(&fs::read(&file2).unwrap()).contains("Build a small editor"));
}
#[test]
fn stale_and_tampered_plans_do_not_overwrite() {
    let root = TempDir::new().unwrap();
    let p = root.path().to_str().unwrap();
    let out = cli(&["project-plan", p], &profile(None));
    let mut plan: Value = ok_result(&out);
    let digest = plan["plan_digest"].as_str().unwrap().to_owned();
    plan["profile"]["purpose"] = json!("Tampered");
    assert!(!cli(
        &[
            "project-apply",
            p,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture"
        ],
        &plan
    )
    .status
    .success());
    assert!(!root.path().join(".moai-cli").exists());
    plan["profile"] = profile(None);
    fs::create_dir(root.path().join(".moai-cli")).unwrap();
    fs::write(
        root.path().join(".moai-cli/project.yaml"),
        b"# user comment\npurpose: user choice\n",
    )
    .unwrap();
    assert!(!cli(
        &[
            "project-apply",
            p,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture"
        ],
        &plan
    )
    .status
    .success());
    assert_eq!(
        fs::read(root.path().join(".moai-cli/project.yaml")).unwrap(),
        b"# user comment\npurpose: user choice\n"
    );
}
#[test]
fn pin_precedence_integrity_and_no_fallback() {
    let repo = TempDir::new().unwrap();
    let user = TempDir::new().unwrap();
    fs::write(repo.path().join("SKILL.md"), "repository package").unwrap();
    fs::write(user.path().join("SKILL.md"), "user package").unwrap();
    let pin = moai_cli::package_digest(repo.path()).unwrap();
    let root = TempDir::new().unwrap();
    create(&root, &profile(Some(pin.clone())));
    let args = [
        "package-resolve",
        root.path().to_str().unwrap(),
        "--repo-package",
        repo.path().to_str().unwrap(),
        "--user-package",
        user.path().to_str().unwrap(),
    ];
    let result = cli(&args, &Value::Null);
    assert!(result.status.success());
    let value: Value = ok_result(&result);
    assert_eq!(value["source"], "repo_pin");
    assert_eq!(value["digest"], pin);
    fs::write(repo.path().join("SKILL.md"), "modified").unwrap();
    assert!(!cli(&args, &Value::Null).status.success());
    assert!(!cli(
        &[
            "package-resolve",
            root.path().to_str().unwrap(),
            "--user-package",
            user.path().to_str().unwrap()
        ],
        &Value::Null
    )
    .status
    .success());
    let unpinned = TempDir::new().unwrap();
    create(&unpinned, &profile(None));
    let result = cli(
        &[
            "package-resolve",
            unpinned.path().to_str().unwrap(),
            "--user-package",
            user.path().to_str().unwrap(),
        ],
        &Value::Null,
    );
    assert!(result.status.success());
    assert_eq!(ok_result(&result)["source"], "user");
}
#[test]
fn nonportable_or_secret_shaped_input_is_rejected_without_echo() {
    for value in [
        "/Users/example/private",
        "C:\\private",
        "sk-example-secret",
        "password=fixture-only",
    ] {
        let root = TempDir::new().unwrap();
        let mut input = profile(None);
        input["purpose"] = json!(value);
        let out = cli(&["project-plan", root.path().to_str().unwrap()], &input);
        assert!(!out.status.success());
        // stdout carries the failure envelope (01-C #1990); neither stream echoes the value.
        let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(failure["ok"], json!(false));
        assert_eq!(failure["error"]["kind"], "invalid_input");
        assert!(!String::from_utf8_lossy(&out.stdout).contains(value));
        assert!(!String::from_utf8_lossy(&out.stderr).contains(value));
        assert!(!root.path().join(".moai-cli").exists());
    }
}
#[cfg(unix)]
#[test]
fn symlinks_are_preserved_and_not_followed() {
    use std::os::unix::fs::symlink;
    let root = TempDir::new().unwrap();
    let outside = TempDir::new().unwrap();
    symlink(outside.path(), root.path().join(".moai-cli")).unwrap();
    assert!(!cli(
        &["project-plan", root.path().to_str().unwrap()],
        &profile(None)
    )
    .status
    .success());
    assert!(fs::read_dir(outside.path()).unwrap().next().is_none());
    fs::write(outside.path().join("SKILL.md"), "package").unwrap();
    symlink(
        outside.path().join("SKILL.md"),
        outside.path().join("alias"),
    )
    .unwrap();
    assert!(moai_cli::package_digest(outside.path()).is_err());
}

#[test]
fn plan_cannot_be_replayed_in_another_project() {
    let first = TempDir::new().unwrap();
    let second = TempDir::new().unwrap();
    let result = cli(
        &["project-plan", first.path().to_str().unwrap()],
        &profile(None),
    );
    let plan: Value = ok_result(&result);
    let result = cli(
        &[
            "project-apply",
            second.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture",
        ],
        &plan,
    );
    assert!(!result.status.success());
    assert!(!second.path().join(".moai-cli").exists());
}
#[test]
fn package_inspection_is_stable_and_covers_support_files() {
    let package = TempDir::new().unwrap();
    fs::write(package.path().join("SKILL.md"), "router").unwrap();
    fs::create_dir(package.path().join("resources")).unwrap();
    fs::write(package.path().join("resources/guide.md"), "guide").unwrap();
    let args = ["package-inspect", package.path().to_str().unwrap()];
    let first = cli(&args, &Value::Null);
    assert!(first.status.success());
    assert_eq!(first.stdout, cli(&args, &Value::Null).stdout);
    fs::write(package.path().join("resources/guide.md"), "changed").unwrap();
    assert_ne!(first.stdout, cli(&args, &Value::Null).stdout);
}

// ---- AGENTS.md managed-section generation -------------------------------------------------

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
fn agents_project(language: &str, local_file: &str) -> (TempDir, Value) {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    create(&root, &profile(None));
    let out = agents_plan(&root, &agents_request(language, local_file));
    assert!(
        out.status.success(),
        "{:?}",
        String::from_utf8_lossy(&out.stderr)
    );
    let plan: Value = ok_result(&out);
    let applied = agents_apply(&root, &plan);
    assert!(
        applied.status.success(),
        "{:?}",
        String::from_utf8_lossy(&applied.stderr)
    );
    (root, plan)
}
fn actions(plan: &Value) -> Vec<(String, String)> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["path"].as_str().unwrap().into(),
                f["action"].as_str().unwrap().into(),
            )
        })
        .collect()
}

#[test]
fn agents_generation_is_idempotent_and_preserves_user_body() {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    create(&root, &profile(None));
    let user_body = "# My project\r\n\r\nUser rules stay as written.\r\n";
    fs::write(root.path().join("AGENTS.md"), user_body).unwrap();
    fs::write(root.path().join(".gitignore"), "target/").unwrap(); // no trailing newline
    let request = agents_request("en", "ignored");
    let out = agents_plan(&root, &request);
    assert!(out.status.success());
    let plan: Value = ok_result(&out);
    assert_eq!(
        actions(&plan),
        vec![
            ("AGENTS.md".into(), "append".into()),
            ("AGENTS.local.md".into(), "create".into()),
            (".gitignore".into(), "append".into()),
            (
                ".moai-cli/manifests/generation.json".into(),
                "create".into()
            ),
        ]
    );
    assert!(agents_apply(&root, &plan).status.success());
    let agents = fs::read_to_string(root.path().join("AGENTS.md")).unwrap();
    assert!(agents.starts_with(user_body));
    assert!(agents.contains("<!-- moai:managed:begin id=moai-workflow -->"));
    assert!(agents.contains("skills/moai/workflows/"));
    let ignore = fs::read_to_string(root.path().join(".gitignore")).unwrap();
    assert!(ignore.starts_with("target/\n"));
    assert!(ignore.contains("\nAGENTS.local.md\n"));
    assert!(root.path().join("AGENTS.local.md").is_file());
    let manifest: Value = serde_json::from_slice(
        &fs::read(root.path().join(".moai-cli/manifests/generation.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(manifest["schema_version"], 1);
    assert_eq!(manifest["owned_files"].as_array().unwrap().len(), 2);
    assert_eq!(manifest["created_once"], json!(["AGENTS.local.md"]));
    assert!(!manifest.to_string().contains(root.path().to_str().unwrap()));

    // Second run: diff 0 on every file.
    let before: Vec<Vec<u8>> = [
        "AGENTS.md",
        "AGENTS.local.md",
        ".gitignore",
        ".moai-cli/manifests/generation.json",
    ]
    .iter()
    .map(|p| fs::read(root.path().join(p)).unwrap())
    .collect();
    let out = agents_plan(&root, &request);
    let plan: Value = ok_result(&out);
    assert!(actions(&plan)
        .iter()
        .all(|(_, a)| a == "unchanged" || a == "merge"));
    assert!(
        actions(&plan)
            .iter()
            .filter(|(p, a)| a == "merge" && p != ".moai-cli/manifests/generation.json")
            .count()
            == 0
    );
    let result = agents_apply(&root, &plan);
    assert!(result.status.success());
    let value: Value = ok_result(&result);
    assert_eq!(value["state"], "unchanged");
    assert_eq!(value["host_loaded"], false);
    let after: Vec<Vec<u8>> = [
        "AGENTS.md",
        "AGENTS.local.md",
        ".gitignore",
        ".moai-cli/manifests/generation.json",
    ]
    .iter()
    .map(|p| fs::read(root.path().join(p)).unwrap())
    .collect();
    assert_eq!(before, after);
}

#[test]
fn agents_user_edits_are_preserved_or_reported_as_conflict() {
    let (root, _) = agents_project("en", "none");
    let file = root.path().join("AGENTS.md");
    let generated = fs::read_to_string(&file).unwrap();

    // Edit outside the managed block: still unchanged, bytes preserved.
    let outside = format!("{generated}\n## User appendix\nkeep me\n");
    fs::write(&file, &outside).unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("en", "none")));
    assert_eq!(actions(&plan)[0].1, "unchanged");
    assert!(agents_apply(&root, &plan).status.success());
    assert_eq!(fs::read_to_string(&file).unwrap(), outside);

    // Edit inside the managed block while the template is unchanged: preserved, no rewrite.
    let edited = outside.replace(
        "Read only the workflow you need",
        "Read ONLY the workflow you need",
    );
    assert_ne!(edited, outside);
    fs::write(&file, &edited).unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("en", "none")));
    assert_eq!(actions(&plan)[0].1, "unchanged");

    // Template changes (language) after a user edit inside the block: conflict, nothing applied.
    let out = agents_plan(&root, &agents_request("ko", "none"));
    assert!(out.status.success());
    let plan: Value = ok_result(&out);
    assert_eq!(actions(&plan)[0].1, "conflict");
    let result = agents_apply(&root, &plan);
    assert!(!result.status.success());
    assert!(String::from_utf8_lossy(&result.stderr).contains("conflict"));
    assert_eq!(fs::read_to_string(&file).unwrap(), edited);

    // Unedited block + template change: merge replaces only the block.
    fs::write(&file, &outside).unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("ko", "none")));
    assert_eq!(actions(&plan)[0].1, "merge");
    assert!(agents_apply(&root, &plan).status.success());
    let merged = fs::read_to_string(&file).unwrap();
    assert!(merged.contains("MoAI 작업 흐름"));
    assert!(merged.ends_with("\n## User appendix\nkeep me\n"));
    assert!(merged.starts_with("# demo\n\n"));

    // User removes the managed block while the template is unchanged: the deletion is
    // respected (never re-appended silently). Once the template changes, it is a conflict.
    fs::write(&file, "# demo\n\nno block\n").unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("ko", "none")));
    assert_eq!(actions(&plan)[0].1, "unchanged");
    // The deletion is reported, not silent.
    let absent = "managed block absent; user deletion preserved";
    assert!(plan["files"][0]["detail"]
        .as_str()
        .is_some_and(|d| d.contains(absent)));
    assert_eq!(
        plan["warnings"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|w| w.as_str().is_some_and(|w| w.contains(absent)))
            .count(),
        1
    );
    assert!(agents_apply(&root, &plan).status.success());
    assert_eq!(fs::read_to_string(&file).unwrap(), "# demo\n\nno block\n");
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("en", "none")));
    assert_eq!(actions(&plan)[0].1, "conflict");

    // Malformed markers: conflict.
    fs::write(
        &file,
        "<!-- moai:managed:begin id=moai-workflow -->\nno end\n",
    )
    .unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("ko", "none")));
    assert_eq!(actions(&plan)[0].1, "conflict");
}

#[test]
fn agents_local_file_is_opt_in_and_never_rewritten() {
    let (root, plan) = agents_project("en", "none");
    assert!(!root.path().join("AGENTS.local.md").exists());
    assert!(!root.path().join(".gitignore").exists());
    assert!(actions(&plan)
        .iter()
        .all(|(p, _)| p != "AGENTS.local.md" && p != ".gitignore"));

    let (root, plan) = agents_project("ko", "tracked");
    assert!(root.path().join("AGENTS.local.md").is_file());
    assert!(!root.path().join(".gitignore").exists());
    assert!(actions(&plan).iter().all(|(p, _)| p != ".gitignore"));
    fs::write(root.path().join("AGENTS.local.md"), "personal\n").unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("ko", "tracked")));
    assert!(agents_apply(&root, &plan).status.success());
    assert_eq!(
        fs::read(root.path().join("AGENTS.local.md")).unwrap(),
        b"personal\n"
    );
    // Deleted by the user: not recreated.
    fs::remove_file(root.path().join("AGENTS.local.md")).unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("ko", "tracked")));
    assert!(agents_apply(&root, &plan).status.success());
    assert!(!root.path().join("AGENTS.local.md").exists());
}

#[test]
fn agents_plan_rejects_replay_tamper_missing_reference_and_shadowing_is_reported() {
    let root = TempDir::new().unwrap();
    create(&root, &profile(None));
    // Missing workflow reference blocks the plan.
    assert!(!agents_plan(&root, &agents_request("en", "none"))
        .status
        .success());
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::write(root.path().join("CLAUDE.md"), "existing\n").unwrap();
    fs::write(root.path().join("AGENTS.override.md"), "codex override\n").unwrap();
    let out = agents_plan(&root, &agents_request("en", "none"));
    assert!(out.status.success());
    let plan: Value = ok_result(&out);
    let notes = plan["host_notes"].as_array().unwrap();
    assert_eq!(notes.len(), 2);
    // Replay in another project is refused.
    let other = TempDir::new().unwrap();
    create(&other, &profile(None));
    fs::create_dir_all(other.path().join("skills/moai/workflows")).unwrap();
    assert!(!agents_apply(&other, &plan).status.success());
    assert!(!other.path().join("AGENTS.md").exists());
    // Tampered plan is refused.
    let mut tampered = plan.clone();
    tampered["request"]["language"] = json!("ko");
    assert!(!agents_apply(&root, &tampered).status.success());
    assert!(!root.path().join("AGENTS.md").exists());
    // Target changed after preview is refused.
    fs::write(root.path().join("AGENTS.md"), "late user file\n").unwrap();
    assert!(!agents_apply(&root, &plan).status.success());
    assert_eq!(
        fs::read(root.path().join("AGENTS.md")).unwrap(),
        b"late user file\n"
    );
    assert_eq!(
        fs::read(root.path().join("CLAUDE.md")).unwrap(),
        b"existing\n"
    );
    assert_eq!(
        fs::read(root.path().join("AGENTS.override.md")).unwrap(),
        b"codex override\n"
    );
    // Tampered manifest is refused.
    agents_project("en", "none");
    let (root2, _) = agents_project("en", "none");
    fs::write(
        root2.path().join(".moai-cli/manifests/generation.json"),
        b"{\"schema_version\":2}",
    )
    .unwrap();
    assert!(!agents_plan(&root2, &agents_request("en", "none"))
        .status
        .success());
}

#[cfg(unix)]
#[test]
fn agents_apply_refuses_symlinked_parents_and_recovers_from_partial_apply() {
    use std::os::unix::fs::symlink;
    let root = TempDir::new().unwrap();
    let outside = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    create(&root, &profile(None));
    symlink(outside.path(), root.path().join("AGENTS.md")).unwrap();
    assert!(!agents_plan(&root, &agents_request("en", "none"))
        .status
        .success());
    fs::remove_file(root.path().join("AGENTS.md")).unwrap();

    // Force a failure after AGENTS.md and AGENTS.local.md are written: the manifest
    // directory is read-only, so the final manifest write fails and the apply is partial.
    use std::os::unix::fs::PermissionsExt;
    let manifests = root.path().join(".moai-cli/manifests");
    fs::create_dir_all(&manifests).unwrap();
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
    assert!(root.path().join("AGENTS.local.md").is_file());
    assert!(!manifests.join("generation.json").exists());
    // Resume: the already-written block is adopted as baseline and the rest completes.
    fs::set_permissions(&manifests, fs::Permissions::from_mode(0o755)).unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("en", "tracked")));
    assert_eq!(actions(&plan)[0].1, "unchanged");
    assert!(agents_apply(&root, &plan).status.success());
    let manifest: Value = serde_json::from_slice(
        &fs::read(root.path().join(".moai-cli/manifests/generation.json")).unwrap(),
    )
    .unwrap();
    let owned: Vec<&str> = manifest["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["path"].as_str().unwrap())
        .collect();
    assert!(owned.contains(&"AGENTS.md"));
    assert_eq!(manifest["created_once"], json!(["AGENTS.local.md"]));
}

#[test]
fn interrupted_update_refreshes_baseline_instead_of_false_conflict() {
    // B1: an update that wrote the block but died before the manifest must not leave a stale
    // baseline that turns the next template change into a false conflict.
    let (root, _) = agents_project("en", "none");
    let manifest = root.path().join(".moai-cli/manifests/generation.json");
    let v1_manifest = fs::read(&manifest).unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("ko", "none")));
    assert_eq!(actions(&plan)[0].1, "merge");
    assert!(agents_apply(&root, &plan).status.success());
    // Simulate the crash between the AGENTS.md write and the manifest write.
    fs::write(&manifest, &v1_manifest).unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("ko", "none")));
    assert_eq!(actions(&plan)[0].1, "unchanged");
    assert!(agents_apply(&root, &plan).status.success());
    let refreshed: Value = serde_json::from_slice(&fs::read(&manifest).unwrap()).unwrap();
    let baseline = refreshed["owned_files"][0]["baseline_sha256"]
        .as_str()
        .unwrap();
    assert_eq!(
        baseline,
        plan["files"][0]["rendered_sha256"].as_str().unwrap()
    );
    // A further template change is a merge, not a conflict.
    fs::create_dir_all(root.path().join("docs/wf")).unwrap();
    let mut request = agents_request("ko", "none");
    request["workflow_reference"] = json!("docs/wf/");
    let plan: Value = ok_result(&agents_plan(&root, &request));
    assert_eq!(actions(&plan)[0].1, "merge");
    assert!(agents_apply(&root, &plan).status.success());
}

#[test]
fn profile_values_cannot_inject_managed_markers() {
    // B2: a profile value carrying marker syntax must be refused before anything is written.
    for (field, value) in [
        (
            "purpose",
            json!("x <!-- moai:managed:end id=moai-workflow --> y"),
        ),
        (
            "tools",
            json!(["Rust", "<!-- moai:managed:begin id=moai-workflow -->"]),
        ),
        ("development_methods", json!(["TDD -->"])),
        ("purpose", json!("mentions moai:managed: syntax")),
    ] {
        let root = TempDir::new().unwrap();
        fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
        let mut input = profile(None);
        input[field] = value;
        create(&root, &input);
        let out = agents_plan(&root, &agents_request("en", "none"));
        assert!(!out.status.success(), "{field}");
        assert!(String::from_utf8_lossy(&out.stderr).contains("marker"));
        assert!(!root.path().join("AGENTS.md").exists());
    }
}

#[test]
fn legacy_manifest_with_personal_file_in_owned_files_is_still_read() {
    // M2 compatibility: manifests written before `created_once` listed AGENTS.local.md in owned_files.
    let (root, _) = agents_project("en", "tracked");
    let manifest = root.path().join(".moai-cli/manifests/generation.json");
    let mut value: Value = serde_json::from_slice(&fs::read(&manifest).unwrap()).unwrap();
    assert_eq!(value["created_once"], json!(["AGENTS.local.md"]));
    let local_digest = moai_cli::digest(&fs::read(root.path().join("AGENTS.local.md")).unwrap());
    value["owned_files"]
        .as_array_mut()
        .unwrap()
        .push(json!({"path":"AGENTS.local.md","section_id":null,"baseline_sha256":local_digest}));
    value.as_object_mut().unwrap().remove("created_once");
    fs::write(&manifest, serde_json::to_vec_pretty(&value).unwrap()).unwrap();
    fs::remove_file(root.path().join("AGENTS.local.md")).unwrap();
    let plan: Value = ok_result(&agents_plan(&root, &agents_request("en", "tracked")));
    // Created once before: the user's deletion is respected and the file is not recreated.
    assert!(actions(&plan)
        .iter()
        .any(|(p, a)| p == "AGENTS.local.md" && a == "unchanged"));
    assert!(agents_apply(&root, &plan).status.success());
    assert!(!root.path().join("AGENTS.local.md").exists());
    let migrated: Value = serde_json::from_slice(&fs::read(&manifest).unwrap()).unwrap();
    assert_eq!(migrated["created_once"], json!(["AGENTS.local.md"]));
    assert!(migrated["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .all(|f| !f["section_id"].is_null()));
}

fn session(args: &[&str]) -> std::process::Output {
    Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("session")
        .args(args)
        .output()
        .unwrap()
}
#[test]
fn session_dry_run_composes_documented_claude_flags() {
    let out = session(&[
        "--role",
        "leader",
        "--name",
        "leader",
        "--advisor",
        "fable",
        "--dry-run",
    ]);
    assert!(out.status.success(), "{:?}", out.stderr);
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["executed"], json!(false));
    assert_eq!(
        v["argv"],
        json!([
            "claude",
            "--name",
            "leader",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--advisor",
            "fable"
        ])
    );
    let plain = session(&["--name", "agent-3", "--dry-run"]);
    let v: Value = serde_json::from_slice(&plain.stdout).unwrap();
    assert_eq!(v["role"], json!("agent"));
    assert!(!v["argv"].to_string().contains("advisor"));
}
#[test]
fn session_rejects_lower_ranked_advisor_and_unknown_host() {
    let out = session(&[
        "--name",
        "a",
        "--model",
        "fable",
        "--advisor",
        "opus",
        "--dry-run",
    ]);
    assert!(!out.status.success());
    assert!(String::from_utf8_lossy(&out.stderr).contains("rank"));
    let out = session(&["--name", "a", "--host", "codex", "--dry-run"]);
    assert!(!out.status.success());
    let out = session(&["--name", "a", "--effort", "turbo", "--dry-run"]);
    assert!(!out.status.success());
}

/// Runs the launcher syntax (00-F #2028). Tests use `--dry-run` only and never start a host.
fn launcher(args: &[&str], token: Option<&str>) -> std::process::Output {
    // A fresh git project per call: `-a` previews its `agent-N` from that project's counter
    // (#2359), never from this repository's.
    let project = tempfile::tempdir().unwrap();
    assert!(Command::new("git")
        .args(["init", "-q"])
        .current_dir(project.path())
        .status()
        .unwrap()
        .success());
    let mut c = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    c.args(args)
        .current_dir(project.path())
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        // No home, so no key file (#2359) can stand in for the token.
        .env(
            "HOME",
            std::env::temp_dir().join("moai-cli-test-no-such-home"),
        );
    match token {
        Some(t) => c.env("ANTHROPIC_AUTH_TOKEN", t),
        None => c.env_remove("ANTHROPIC_AUTH_TOKEN"),
    };
    c.output().unwrap()
}
fn launched(out: &std::process::Output) -> Value {
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["executed"], json!(false));
    v
}
#[test]
fn launcher_leader_default_claude() {
    let expected = json!([
        "claude",
        "--name",
        "leader",
        "--model",
        "opus",
        "--effort",
        "medium",
        "--dangerously-skip-permissions"
    ]);
    for args in [
        &["-l", "--dry-run"][..],
        &["cc", "-l", "--dry-run"][..],
        &["--leader", "--dry-run"][..],
    ] {
        let v = launched(&launcher(args, None));
        assert_eq!(v["argv"], expected, "{args:?}");
        assert_eq!(
            (v["host"].clone(), v["role"].clone()),
            (json!("claude"), json!("leader"))
        );
        assert_eq!(v["env_keys"], json!([]));
    }
    // `-a` without a host is the Claude Code agent named from the project counter (#2359); a
    // fresh project previews `agent-1` and the dry run takes no number.
    let v = launched(&launcher(&["-a", "--dry-run"], None));
    assert_eq!(v["role"], json!("agent"));
    assert_eq!(v["argv"][2], json!("agent-1"));
    assert!(v["notes"].to_string().contains("--dry-run takes no number"));
    assert!(!v["notes"].to_string().contains("HOLD"));
    // Ambiguous or empty invocations fail instead of guessing.
    assert!(!launcher(&["-l", "-a", "--dry-run"], None).status.success());
    assert!(!launcher(&["-l", "cc", "--dry-run"], None).status.success());
    // A host word without -l|-a is a plain host session since #2053 (tests/launcher_roleless.rs).
    assert!(launcher(&["cc", "--dry-run"], None).status.success());
    assert!(!launcher(&["-a", "zcode", "--dry-run"], None)
        .status
        .success());
    assert!(!launcher(&[], None).status.success());
}
#[test]
fn launcher_flags_can_be_disabled() {
    let v = launched(&launcher(
        &[
            "cc",
            "-a",
            "--name",
            "agent-3",
            "--no-advisor",
            "--permissions",
            "--dry-run",
        ],
        None,
    ));
    assert_eq!(
        v["argv"],
        json!(["claude", "--name", "agent-3", "--model", "opus", "--effort", "medium"])
    );
    let v = launched(&launcher(
        &[
            "-a",
            "--name",
            "agent-4",
            "--model",
            "sonnet",
            "--effort",
            "high",
            "--dry-run",
        ],
        None,
    ));
    // No advisor by default since the user decision of 2026-11-12 (#2476).
    assert_eq!(
        v["argv"],
        json!([
            "claude",
            "--name",
            "agent-4",
            "--model",
            "sonnet",
            "--effort",
            "high",
            "--dangerously-skip-permissions"
        ])
    );
    assert!(!launcher(&["-l", "--effort", "turbo", "--dry-run"], None)
        .status
        .success());
    assert!(!launcher(&["-l", "--name", "two words", "--dry-run"], None)
        .status
        .success());
}

/// #2164 (A5) warned that `-l|-a` were deprecated; the user decision of 2026-11-10 (#2359,
/// comment 6760976883) keeps them as the way to name a session `leader` / `agent-N`, so no
/// shape warns any more and each still composes.
#[test]
fn launcher_role_flags_are_kept_without_a_deprecation_warning() {
    for args in [
        &["-l", "--dry-run"][..],
        &["-a", "--name", "agent-7", "--dry-run"][..],
        &["cc", "-l", "--dry-run"][..],
        &["codex", "-a", "--dry-run"][..],
        &["cc", "--dry-run"][..],
    ] {
        let out = launcher(args, None);
        let v = launched(&out);
        assert!(
            !String::from_utf8_lossy(&out.stderr).contains("deprecated"),
            "stderr: {args:?}"
        );
        assert!(
            !v["notes"].to_string().contains("deprecated"),
            "notes: {args:?}"
        );
    }
}

#[test]
fn launcher_glm_requires_token_name_only() {
    let sentinel = "moai-test-sentinel-not-a-real-token";
    let out = launcher(&["glm", "-a", "--dry-run"], None);
    assert!(!out.status.success());
    let err = String::from_utf8_lossy(&out.stderr);
    assert!(
        err.contains("ANTHROPIC_AUTH_TOKEN") && err.contains("blocked"),
        "{err}"
    );
    assert!(out.stdout.is_empty());
    let out = launcher(&["glm", "-l", "--dry-run"], Some(sentinel));
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(
        !text.contains(sentinel),
        "the token value must never be printed"
    );
    let v = launched(&out);
    assert_eq!(v["host"], json!("glm"));
    assert_eq!(
        v["env_keys"],
        json!([
            "ANTHROPIC_BASE_URL",
            "API_TIMEOUT_MS",
            "ANTHROPIC_DEFAULT_OPUS_MODEL",
            "ANTHROPIC_DEFAULT_SONNET_MODEL",
            "ANTHROPIC_DEFAULT_HAIKU_MODEL",
            "ANTHROPIC_DEFAULT_FABLE_MODEL",
            "CLAUDE_CODE_AUTO_COMPACT_WINDOW",
            "CLAUDE_CODE_MAX_CONTEXT_TOKENS",
            "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS",
            "ANTHROPIC_AUTH_TOKEN",
            "Z_AI_API_KEY"
        ])
    );
    assert_eq!(
        v["argv"],
        json!([
            "claude",
            "--name",
            "leader",
            "--model",
            "opus",
            "--effort",
            "medium",
            "--dangerously-skip-permissions"
        ])
    );
    // No advisor by default since #2476; an explicit one is dropped on glm with a note.
    assert!(!v["notes"].to_string().contains("--advisor omitted"));
    let v = launched(&launcher(
        &["glm", "-l", "--advisor", "fable", "--dry-run"],
        Some(sentinel),
    ));
    assert!(!v["argv"].to_string().contains("advisor"));
    assert!(v["notes"].to_string().contains("--advisor omitted on glm"));
}
#[test]
fn launcher_codex_argv() {
    // Full access since #2359 (user decision 2026-11-10), leader and agent alike.
    let v = launched(&launcher(&["codex", "-l", "--dry-run"], None));
    assert_eq!(
        v["argv"],
        json!([
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "MoAI session name: leader"
        ])
    );
    assert_eq!(v["host"], json!("codex"));
    assert!(v["notes"].to_string().contains("cross-session"));
    let v = launched(&launcher(
        &[
            "-a",
            "codex",
            "--name",
            "agent-2",
            "--model",
            "some-model",
            "--dry-run",
        ],
        None,
    ));
    assert_eq!(
        v["argv"],
        json!([
            "codex",
            "-s",
            "danger-full-access",
            "-a",
            "never",
            "-m",
            "some-model",
            "MoAI session name: agent-2"
        ])
    );
    let v = launched(&launcher(
        &[
            "codex",
            "-a",
            "--name",
            "agent-5",
            "--dry-run",
            "first task",
        ],
        None,
    ));
    assert_eq!(
        v["argv"].as_array().unwrap().last().unwrap(),
        &json!("MoAI session name: agent-5\nfirst task")
    );
    assert!(!v["argv"].to_string().contains("dangerously"));
    // A PROMPT on a Claude host is not passed on, and the output says so.
    let v = launched(&launcher(&["cc", "-l", "--dry-run", "first task"], None));
    assert!(!v["argv"].to_string().contains("first task"));
    assert!(v["notes"]
        .to_string()
        .contains("PROMPT is used by codex only"));
}

fn git_in(dir: &std::path::Path, args: &[&str]) {
    let out = Command::new("git")
        .args([
            "-c",
            "user.name=fixture",
            "-c",
            "user.email=fixture@example.invalid",
        ])
        .args(["-c", "commit.gpgsign=false", "-c", "core.autocrlf=false"])
        .args(["-c", "core.excludesFile=/dev/null"])
        .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
        .arg("-C")
        .arg(dir)
        .args(args)
        .output()
        .unwrap();
    assert!(out.status.success(), "git {args:?}: {:?}", out.stderr);
}
fn write_file(root: &std::path::Path, rel: &str, bytes: &[u8]) {
    let path = root.join(rel);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, bytes).unwrap();
}
/// Every path and byte under `dir`, `.git` included, without following symlinks.
fn tree_state(dir: &std::path::Path) -> Vec<(String, Vec<u8>)> {
    fn walk(root: &std::path::Path, dir: &std::path::Path, out: &mut Vec<(String, Vec<u8>)>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            let rel = p.strip_prefix(root).unwrap().to_string_lossy().into_owned();
            let m = fs::symlink_metadata(&p).unwrap();
            if m.file_type().is_symlink() {
                let t = fs::read_link(&p).unwrap();
                out.push((rel, t.to_string_lossy().as_bytes().to_vec()));
            } else if m.is_dir() {
                out.push((rel, Vec::new()));
                walk(root, &p, out);
            } else {
                out.push((rel, fs::read(&p).unwrap()));
            }
        }
    }
    let mut out = Vec::new();
    walk(dir, dir, &mut out);
    out.sort();
    out
}
/// Fixture repo: a Go-style baseline commit with `.claude` and template `.claude`, then an
/// ignored installed `.claude/` that partly differs, a lane worktree directory inside it,
/// CRLF bytes, a Unix symlink, one modified and one untracked file.
fn baseline_fixture() -> (TempDir, String) {
    let root = tempfile::Builder::new()
        .prefix("moai baseline ")
        .tempdir()
        .unwrap();
    let r = root.path();
    git_in(r, &["init", "-q", "-b", "main"]);
    write_file(r, ".claude/same.md", b"same\n");
    write_file(r, ".claude/changed.md", b"old\n");
    write_file(r, ".claude/gone.md", b"gone\n");
    write_file(r, "internal/template/templates/.claude/same.md", b"same\n");
    write_file(r, "internal/template/templates/.claude/tpl.md", b"tpl\n");
    write_file(
        r,
        "internal/template/templates/.claude/settings.local.json",
        b"{}\n",
    );
    write_file(r, "crlf.txt", b"a\r\nb\r\n");
    write_file(r, "edited.txt", b"before\n");
    #[cfg(unix)]
    std::os::unix::fs::symlink("crlf.txt", r.join("link")).unwrap();
    git_in(r, &["add", "-A"]);
    git_in(r, &["commit", "-q", "-m", "baseline"]);
    let out = Command::new("git")
        .arg("-C")
        .arg(r)
        .args(["rev-parse", "HEAD"])
        .output()
        .unwrap();
    let commit = String::from_utf8(out.stdout).unwrap().trim().to_string();
    git_in(r, &["rm", "-q", "-r", "--cached", ".claude"]);
    write_file(r, ".gitignore", b".claude/\n");
    git_in(r, &["add", ".gitignore"]);
    git_in(r, &["commit", "-q", "-m", "install"]);
    fs::remove_file(r.join(".claude/gone.md")).unwrap();
    write_file(r, ".claude/changed.md", b"new\n");
    write_file(r, ".claude/mine.md", b"mine\n");
    write_file(r, ".claude/worktrees/lane/x.md", b"lane\n");
    write_file(r, ".moai-cli/worktrees/lane/y.md", b"lane\n");
    write_file(r, "edited.txt", b"after\n");
    write_file(r, "new.txt", b"new\n");
    (root, commit)
}
fn baseline(root: &TempDir, extra: &[&str]) -> (std::process::Output, Value) {
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("baseline-inventory")
        .arg(root.path())
        .args(extra)
        .output()
        .unwrap();
    // Success carries the inventory under `result`; a failure envelope has none (Null).
    let v = serde_json::from_slice::<Value>(&out.stdout)
        .map(|v| v["result"].clone())
        .unwrap_or(Value::Null);
    (out, v)
}
#[test]
fn baseline_inventory_is_read_only_repeatable_and_relative() {
    let (root, commit) = baseline_fixture();
    let before = tree_state(root.path());
    let (out1, v1) = baseline(&root, &["--go-baseline-commit", &commit]);
    assert!(out1.status.success(), "{:?}", out1.stderr);
    let (_, v2) = baseline(&root, &["--go-baseline-commit", &commit]);
    assert_eq!(
        before,
        tree_state(root.path()),
        "collector wrote to the tree"
    );
    let strip = |v: &Value| {
        let mut v = v.clone();
        for k in ["", "baseline", "inventory", "claude_diff"] {
            let o = if k.is_empty() { &mut v } else { &mut v[k] };
            o.as_object_mut().unwrap().remove("collected_at");
        }
        v
    };
    assert_eq!(strip(&v1), strip(&v2));
    let raw = String::from_utf8(out1.stdout).unwrap();
    let canon = fs::canonicalize(root.path()).unwrap();
    assert!(!raw.contains(canon.to_str().unwrap()));
    assert!(!raw.contains(root.path().to_str().unwrap()));
    if let Some(home) = std::env::var_os("HOME") {
        assert!(!raw.contains(home.to_str().unwrap()));
    }
    let head = Command::new("git")
        .arg("-C")
        .arg(root.path())
        .args(["rev-parse", "HEAD"])
        .output()
        .unwrap();
    assert_eq!(
        v1["baseline"]["head_sha"],
        json!(String::from_utf8(head.stdout).unwrap().trim())
    );
    assert_eq!(
        v1["baseline"]["remote_heads"]["status"],
        json!("unavailable")
    );
    assert!(v1["baseline"]["home"]
        .as_array()
        .unwrap()
        .iter()
        .all(|h| h["path"].as_str().unwrap().starts_with("~/")));
    let files = v1["inventory"]["files"].as_array().unwrap();
    let file = |p: &str| {
        files
            .iter()
            .find(|f| f["path"] == json!(p))
            .unwrap()
            .clone()
    };
    assert_eq!(
        file("crlf.txt")["sha256"],
        json!(moai_cli::digest(b"a\r\nb\r\n"))
    );
    #[cfg(unix)]
    {
        assert_eq!(file("link")["kind"], json!("symlink"));
        assert_eq!(file("link")["sha256"], json!(moai_cli::digest(b"crlf.txt")));
    }
    assert_eq!(
        file("edited.txt")["sha256"],
        json!(moai_cli::digest(b"after\n"))
    );
    let changes = v1["inventory"]["changes"].as_array().unwrap();
    let paths: Vec<&str> = changes
        .iter()
        .map(|c| c["path"].as_str().unwrap())
        .collect();
    assert_eq!(paths, vec!["edited.txt", "new.txt"]);
    assert_eq!(changes[1]["status"], json!("untracked"));
    assert_eq!(v1["inventory"]["excluded_changes"], json!(1));
    let sums = v1["checksums"].as_str().unwrap();
    assert!(sums.contains("  crlf.txt\n"));
    #[cfg(unix)]
    assert!(!sums.contains("  link\n"), "symlinks are not followed");
    assert!(sums.contains("  new.txt\n"));
}
#[test]
fn baseline_claude_diff_classifies_every_path_once() {
    let (root, commit) = baseline_fixture();
    let (out, v) = baseline(&root, &["--go-baseline-commit", &commit]);
    assert!(out.status.success(), "{:?}", out.stderr);
    let d = &v["claude_diff"];
    assert_eq!(d["status"], json!("collected"));
    assert_eq!(d["source"], json!("project"));
    assert_eq!(d["excluded_active_files"], json!(1));
    let rows = d["rows"].as_array().unwrap();
    assert_eq!(d["total"], json!(rows.len()));
    let class = |p: &str| {
        let r = rows
            .iter()
            .find(|r| r["path"] == json!(p))
            .unwrap_or_else(|| panic!("{p} missing from {rows:?}"));
        (r["class"].as_str().unwrap().to_string(), r.clone())
    };
    assert_eq!(class("same.md").0, "identical");
    assert_eq!(class("changed.md").0, "differing");
    assert_eq!(class("gone.md").0, "baseline-only");
    assert_eq!(class("tpl.md").0, "template-only");
    assert_eq!(class("mine.md").0, "active-only");
    let (_, local) = class("settings.local.json");
    assert_eq!(local["scope"], json!("private"));
    assert_eq!(local["origin"], json!("generated"));
    let (_, mine) = class("mine.md");
    assert_eq!(mine["tracking"], json!("ignored"));
    assert_eq!(mine["origin"], json!("source"));
    assert!(rows
        .iter()
        .all(|r| !r["path"].as_str().unwrap().starts_with("worktrees/")));
}
#[test]
fn tree_digest_stable_sensitive_and_excludes() {
    let root = tempfile::Builder::new()
        .prefix("moai tree ")
        .tempdir()
        .unwrap();
    let r = root.path();
    write_file(r, "a.txt", b"a\r\n");
    write_file(r, "dir/b.txt", b"b\n");
    write_file(r, "skip/c.txt", b"c\n");
    #[cfg(unix)]
    std::os::unix::fs::symlink("a.txt", r.join("link")).unwrap();
    let before = tree_state(r);
    let d1 = moai_cli::tree_digest(r, &["skip/"]).unwrap();
    assert_eq!(d1, moai_cli::tree_digest(r, &["skip/"]).unwrap());
    assert_eq!(before, tree_state(r), "tree_digest wrote to the tree");
    assert!(d1.starts_with("sha256:") && d1.len() == 71);
    write_file(r, "skip/c.txt", b"changed\n");
    write_file(r, "skip/new.txt", b"new\n");
    assert_eq!(d1, moai_cli::tree_digest(r, &["skip"]).unwrap());
    assert_ne!(d1, moai_cli::tree_digest(r, &[]).unwrap());
    write_file(r, "dir/b.txt", b"B\n");
    let d2 = moai_cli::tree_digest(r, &["skip/"]).unwrap();
    assert_ne!(d1, d2);
    write_file(r, "dir/b.txt", b"b\n");
    assert_eq!(d1, moai_cli::tree_digest(r, &["skip/"]).unwrap());
    #[cfg(unix)]
    {
        fs::remove_file(r.join("link")).unwrap();
        std::os::unix::fs::symlink("dir/b.txt", r.join("link")).unwrap();
        assert_ne!(d1, moai_cli::tree_digest(r, &["skip/"]).unwrap());
        fs::remove_file(r.join("link")).unwrap();
        std::os::unix::fs::symlink("a.txt", r.join("link")).unwrap();
    }
    fs::create_dir(r.join("empty")).unwrap();
    assert_eq!(d1, moai_cli::tree_digest(r, &["skip/"]).unwrap());
    fs::rename(r.join("a.txt"), r.join("z.txt")).unwrap();
    assert_ne!(d1, moai_cli::tree_digest(r, &["skip/"]).unwrap());
}
#[test]
fn baseline_lists_paths_the_checksum_format_cannot_carry() {
    let (root, _) = baseline_fixture();
    write_file(root.path(), "back\\slash.txt", b"x\n");
    let (out, v) = baseline(&root, &[]);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(
        v["inventory"]["checksum_unrepresentable"],
        json!(["back\\slash.txt"])
    );
    assert!(!v["checksums"].as_str().unwrap().contains("slash"));
    assert!(v["inventory"]["changes"]
        .as_array()
        .unwrap()
        .iter()
        .any(|c| c["path"] == json!("back\\slash.txt")));
}
#[test]
fn baseline_without_or_with_unknown_commit_degrades_without_error() {
    let (root, _) = baseline_fixture();
    let (out, v) = baseline(&root, &[]);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["claude_diff"]["status"], json!("not_collected"));
    let (out, v) = baseline(&root, &["--go-baseline-commit", &"a".repeat(40)]);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["claude_diff"]["status"], json!("commit_not_found"));
    let (out, _) = baseline(&root, &["--go-baseline-commit", "HEAD"]);
    assert!(!out.status.success());
}

// ---- Current protocol contract fixtures (Issue #1987, 00-D) ------------------------------
//
// Records what every subcommand prints today (exit code, stdout, stderr) as golden files under
// tests/fixtures/protocol/v1/current/. This is the baseline the protocol draft is measured
// against; it describes behavior, it does not choose a contract. Run with
// MOAI_UPDATE_FIXTURES=1 to rewrite the files instead of comparing.

const PROTOCOL_FIXTURES: &str = "tests/fixtures/protocol/v1/current";
/// Values that depend on the temporary directory or the clock, not on the contract.
const MASKED_KEYS: [&str; 4] = ["target_binding", "plan_digest", "collected_at", "receipt"];

fn masked_values(v: &Value, out: &mut Vec<String>) {
    match v {
        Value::Object(m) => {
            for (k, x) in m {
                match (MASKED_KEYS.contains(&k.as_str()), x.as_str()) {
                    (true, Some(s)) => out.push(s.to_string()),
                    _ => masked_values(x, out),
                }
            }
        }
        Value::Array(a) => a.iter().for_each(|x| masked_values(x, out)),
        _ => {}
    }
}
fn mask_text(s: &str, masks: &[(String, String)]) -> String {
    masks.iter().fold(s.to_string(), |acc, (from, to)| {
        acc.replace(from.as_str(), to)
    })
}
fn mask_value(v: &Value, masks: &[(String, String)]) -> Value {
    match v {
        Value::String(s) => json!(mask_text(s, masks)),
        Value::Array(a) => Value::Array(a.iter().map(|x| mask_value(x, masks)).collect()),
        Value::Object(m) => Value::Object(
            m.iter()
                .map(|(k, x)| (mask_text(k, masks), mask_value(x, masks)))
                .collect(),
        ),
        other => other.clone(),
    }
}
fn as_json_or_text(bytes: &[u8]) -> Value {
    serde_json::from_slice(bytes)
        .unwrap_or_else(|_| json!(String::from_utf8_lossy(bytes).into_owned()))
}

struct FixtureRun {
    roots: Vec<std::path::PathBuf>,
    home: TempDir,
    seen: Vec<String>,
}
impl FixtureRun {
    /// Runs one case and compares (or rewrites) its golden file. Returns parsed stdout.
    fn case(&mut self, name: &str, args: &[&str], stdin: &[u8]) -> Value {
        let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
            .args(args)
            .env("HOME", self.home.path())
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_NOSYSTEM", "1")
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .unwrap();
        // Same rule as cli(): a child that exits before reading stdin surfaces as EPIPE on
        // fast runners. The fixture records exit, stdout and stderr, never the write itself.
        match child.stdin.take().unwrap().write_all(stdin) {
            Ok(()) => {}
            Err(e) if e.kind() == std::io::ErrorKind::BrokenPipe => {}
            Err(e) => panic!("stdin write failed: {e}"),
        }
        let out = child.wait_with_output().unwrap();
        let stdin_v = if stdin.is_empty() {
            Value::Null
        } else {
            as_json_or_text(stdin)
        };
        let stdout_v = if out.stdout.is_empty() {
            Value::Null
        } else {
            as_json_or_text(&out.stdout)
        };
        let mut secrets = Vec::new();
        masked_values(&stdin_v, &mut secrets);
        masked_values(&stdout_v, &mut secrets);
        let mut masks: Vec<(String, String)> = Vec::new();
        for r in &self.roots {
            for spelling in [r.clone(), fs::canonicalize(r).unwrap()] {
                masks.push((spelling.to_str().unwrap().to_string(), "<root>".into()));
            }
        }
        masks.extend(secrets.into_iter().map(|s| (s, "<masked>".to_string())));
        masks.sort_by_key(|a| std::cmp::Reverse(a.0.len()));
        let record = json!({
            "case": name,
            "args": mask_value(&json!(args), &masks),
            "stdin": mask_value(&stdin_v, &masks),
            "exit": out.status.code(),
            "stdout": mask_value(&stdout_v, &masks),
            "stderr": mask_text(&String::from_utf8_lossy(&out.stderr), &masks),
            "masked_keys": MASKED_KEYS,
        });
        let file = format!(
            "{}/{:02}-{name}.json",
            PROTOCOL_FIXTURES,
            self.seen.len() + 1
        );
        self.seen.push(file.clone());
        let path = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join(&file);
        if std::env::var_os("MOAI_UPDATE_FIXTURES").is_some() {
            fs::create_dir_all(path.parent().unwrap()).unwrap();
            let mut text = serde_json::to_string_pretty(&record).unwrap();
            text.push('\n');
            fs::write(&path, text).unwrap();
        } else {
            let golden: Value = serde_json::from_slice(
                &fs::read(&path).unwrap_or_else(|_| panic!("missing fixture {file}")),
            )
            .unwrap();
            assert_eq!(golden, record, "protocol fixture drift: {file}");
        }
        stdout_v
    }
}

#[test]
fn protocol_fixtures_record_current_contract() {
    let project = TempDir::new().unwrap();
    let p = project.path().to_str().unwrap().to_string();
    let package = TempDir::new().unwrap();
    let pkg = package.path().to_str().unwrap().to_string();
    let empty = TempDir::new().unwrap();
    let e = empty.path().to_str().unwrap().to_string();
    let repo = TempDir::new().unwrap();
    let r = repo.path().to_str().unwrap().to_string();
    fs::create_dir_all(project.path().join("skills/moai/workflows")).unwrap();
    fs::write(package.path().join("SKILL.md"), "router\n").unwrap();
    let mut run = FixtureRun {
        roots: [&project, &package, &empty, &repo]
            .iter()
            .map(|t| t.path().to_path_buf())
            .collect(),
        home: TempDir::new().unwrap(),
        seen: Vec::new(),
    };
    run.roots.push(run.home.path().to_path_buf());
    let prof = profile(None).to_string();

    run.case("usage-unknown-subcommand", &["no-such-command"], b"");
    run.case("usage-missing-argument", &["project-plan"], b"");
    let mut extra = profile(None);
    extra["unexpected"] = json!(1);
    run.case(
        "project-plan-unknown-field",
        &["project-plan", &p],
        extra.to_string().as_bytes(),
    );
    run.case("project-plan-malformed-json", &["project-plan", &p], b"{");
    run.case(
        "project-plan-input-over-limit",
        &["project-plan", &p],
        &vec![b' '; 65537],
    );
    let plan = run.case(
        "project-plan-create",
        &["project-plan", &p],
        prof.as_bytes(),
    )["result"]
        .clone();
    let digest = plan["plan_digest"].as_str().unwrap().to_string();
    let mut v2 = plan.clone();
    v2["protocol_version"] = json!(2);
    run.case(
        "project-apply-protocol-version-2",
        &[
            "project-apply",
            &p,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
        v2.to_string().as_bytes(),
    );
    run.case(
        "project-apply-placeholder-permission",
        &[
            "project-apply",
            &p,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "placeholder",
        ],
        plan.to_string().as_bytes(),
    );
    run.case(
        "project-apply-created",
        &[
            "project-apply",
            &p,
            "--plan-digest",
            &digest,
            "--permission-reference",
            "fixture-approved",
        ],
        plan.to_string().as_bytes(),
    );
    run.case(
        "project-plan-unchanged",
        &["project-plan", &p],
        prof.as_bytes(),
    );
    run.case("package-inspect-ok", &["package-inspect", &pkg], b"");
    run.case(
        "package-inspect-missing-skill",
        &["package-inspect", &e],
        b"",
    );
    run.case(
        "package-resolve-user",
        &["package-resolve", &p, "--user-package", &pkg],
        b"",
    );
    run.case("package-resolve-no-package", &["package-resolve", &p], b"");
    let mut req2 = agents_request("ko", "none");
    req2["protocol_version"] = json!(2);
    run.case(
        "agents-plan-protocol-version-2",
        &["agents-plan", &p],
        req2.to_string().as_bytes(),
    );
    let aplan = run.case(
        "agents-plan-create",
        &["agents-plan", &p],
        agents_request("ko", "none").to_string().as_bytes(),
    )["result"]
        .clone();
    let adigest = aplan["plan_digest"].as_str().unwrap().to_string();
    run.case(
        "agents-apply-ok",
        &[
            "agents-apply",
            &p,
            "--plan-digest",
            &adigest,
            "--permission-reference",
            "fixture-approved",
        ],
        aplan.to_string().as_bytes(),
    );
    run.case(
        "session-dry-run",
        &["session", "--name", "agent-1", "--dry-run"],
        b"",
    );
    run.case(
        "session-invalid-effort",
        &[
            "session",
            "--name",
            "agent-1",
            "--effort",
            "turbo",
            "--dry-run",
        ],
        b"",
    );
    run.case(
        "baseline-inventory-not-a-repo",
        &["baseline-inventory", &e],
        b"",
    );
    // Fixed identity and dates make the fixture repository's commit id reproducible.
    write_file(repo.path(), "a.txt", b"a\n");
    for args in [
        &["init", "-q", "-b", "main"][..],
        &["add", "a.txt"][..],
        &["commit", "-q", "-m", "fixture"][..],
    ] {
        let out = Command::new("git")
            .args([
                "-c",
                "user.name=fixture",
                "-c",
                "user.email=fixture@example.invalid",
            ])
            .args(["-c", "commit.gpgsign=false"])
            .args(["-c", "gc.auto=0", "-c", "maintenance.auto=false"])
            .arg("-C")
            .arg(repo.path())
            .args(args)
            .env("GIT_CONFIG_GLOBAL", "/dev/null")
            .env("GIT_CONFIG_NOSYSTEM", "1")
            .env("GIT_AUTHOR_DATE", "2026-11-08T00:00:00Z")
            .env("GIT_COMMITTER_DATE", "2026-11-08T00:00:00Z")
            .output()
            .unwrap();
        assert!(out.status.success(), "git {args:?}");
    }
    run.case("baseline-inventory-ok", &["baseline-inventory", &r], b"");

    let dir = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join(PROTOCOL_FIXTURES);
    let mut on_disk: Vec<String> = fs::read_dir(&dir)
        .unwrap()
        .map(|e| {
            format!(
                "{PROTOCOL_FIXTURES}/{}",
                e.unwrap().file_name().to_string_lossy()
            )
        })
        .collect();
    on_disk.sort();
    assert_eq!(on_disk, run.seen, "fixture set differs from the case list");
}

fn repo_json(rel: &str) -> Value {
    let path = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join(rel);
    serde_json::from_slice(&fs::read(&path).unwrap_or_else(|_| panic!("missing {rel}"))).unwrap()
}

#[test]
fn status_crosswalk_is_complete_and_never_promotes() {
    let x = repo_json("schemas/protocol/v1/status-crosswalk.json");
    let mut source: Vec<String> = Vec::new();
    for s in x["sources"].as_array().unwrap() {
        for t in s["tokens"].as_array().unwrap() {
            source.push(format!(
                "{}:{}",
                s["domain"].as_str().unwrap(),
                t.as_str().unwrap()
            ));
        }
    }
    source.sort();
    let from = x["forbidden_promotions"]["from"].as_array().unwrap();
    let to = x["forbidden_promotions"]["to"].as_array().unwrap();
    let variants = x["variants"].as_object().unwrap();
    assert_eq!(variants.keys().collect::<Vec<_>>(), ["A", "B"]);
    for (name, v) in variants {
        let allowed = v
            .get("canonical_enum")
            .or_else(|| v.get("common_enum"))
            .unwrap()
            .as_array()
            .unwrap();
        let mut keys = Vec::new();
        for row in v["rows"].as_array().unwrap() {
            let key = format!(
                "{}:{}",
                row["domain"].as_str().unwrap(),
                row["source_token"].as_str().unwrap()
            );
            let (canonical, reason) = (&row["canonical"], &row["no_map_reason"]);
            assert!(
                canonical.is_null() != reason.is_null(),
                "{name} {key}: exactly one of canonical / no_map_reason"
            );
            if !canonical.is_null() {
                assert!(
                    allowed.contains(canonical),
                    "{name} {key}: value outside enum"
                );
                assert!(
                    !(from.contains(&row["source_token"]) && to.contains(canonical)),
                    "{name} {key}: forbidden promotion to {canonical}"
                );
            }
            keys.push(key);
        }
        keys.sort();
        assert_eq!(
            keys, source,
            "{name}: rows must cover every source token once"
        );
    }
}

#[test]
fn protocol_schemas_declare_2020_12_and_cover_current_fixtures() {
    let dir = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("schemas/protocol/v1");
    let mut ids = Vec::new();
    for sub in ["", "variants"] {
        for e in fs::read_dir(dir.join(sub)).unwrap() {
            let name = e.unwrap().file_name().to_string_lossy().into_owned();
            if name.ends_with(".schema.json") {
                let s = repo_json(&format!("schemas/protocol/v1/{sub}/{name}").replace("//", "/"));
                assert_eq!(
                    s["$schema"],
                    json!("https://json-schema.org/draft/2020-12/schema"),
                    "{name}"
                );
                ids.push(s["$id"].as_str().unwrap().to_string());
            }
        }
    }
    let unique: std::collections::BTreeSet<_> = ids.iter().collect();
    assert_eq!(unique.len(), ids.len(), "duplicate $id");
    // Key coverage only, not schema validation (the validator is HOLD-00D-2): every current
    // success output carries exactly the required keys and nothing outside properties. Since
    // 01-C #1990 the operation result sits under the envelope's `result`.
    for (schema, fixture) in [
        ("plan", "06-project-plan-create"),
        ("plan", "10-project-plan-unchanged"),
        ("agents-plan", "16-agents-plan-create"),
        ("package-result", "11-package-inspect-ok"),
        ("package-result", "13-package-resolve-user"),
    ] {
        let s = repo_json(&format!("schemas/protocol/v1/{schema}.schema.json"));
        let out = &repo_json(&format!("{PROTOCOL_FIXTURES}/{fixture}.json"))["stdout"]["result"];
        let props = s["properties"].as_object().unwrap();
        for k in s["required"].as_array().unwrap() {
            assert!(
                out.get(k.as_str().unwrap()).is_some(),
                "{fixture}: missing {k}"
            );
        }
        for k in out.as_object().unwrap().keys() {
            assert!(
                props.contains_key(k),
                "{fixture}: {k} not in {schema} schema"
            );
        }
    }
}

#[test]
fn exit_code_table_matches_current_fixtures() {
    let t = repo_json("schemas/protocol/v1/variants/hold-00d-4-exit-codes.json");
    let kinds: Vec<&str> = t["draft_kinds"]
        .as_array()
        .unwrap()
        .iter()
        .map(|k| k.as_str().unwrap())
        .collect();
    let mut failing = Vec::new();
    for e in fs::read_dir(std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join(PROTOCOL_FIXTURES))
        .unwrap()
    {
        let name = e.unwrap().file_name().to_string_lossy().into_owned();
        let f = repo_json(&format!("{PROTOCOL_FIXTURES}/{name}"));
        if f["exit"] != json!(0) {
            failing.push(name.trim_end_matches(".json").to_string());
        }
    }
    failing.sort();
    let rows = t["current_observed"]["rows"].as_array().unwrap();
    let mut listed: Vec<String> = rows
        .iter()
        .map(|r| r["fixture"].as_str().unwrap().to_string())
        .collect();
    listed.sort();
    assert_eq!(
        listed, failing,
        "every failing fixture is in the observed table"
    );
    for r in rows {
        let f = repo_json(&format!(
            "{PROTOCOL_FIXTURES}/{}.json",
            r["fixture"].as_str().unwrap()
        ));
        assert_eq!(f["exit"], r["exit"], "{}", r["fixture"]);
        assert!(f["stderr"]
            .as_str()
            .unwrap()
            .starts_with(r["stderr_start"].as_str().unwrap()));
        assert!(kinds.contains(&r["draft_kind"].as_str().unwrap()));
    }
    for (name, opt) in t["options"].as_object().unwrap() {
        let table = opt["table"].as_object().unwrap();
        assert_eq!(table["ok"], json!(0), "{name}");
        for k in &kinds {
            let code = table[*k]
                .as_u64()
                .unwrap_or_else(|| panic!("{name} lacks {k}"));
            assert!(code > 0 && code < 256, "{name} {k}");
        }
    }
    // Option A is the only one today's binary already satisfies.
    let a = t["options"]["A"]["table"].as_object().unwrap();
    for r in rows {
        assert_eq!(
            a[r["draft_kind"].as_str().unwrap()],
            r["exit"],
            "{}",
            r["fixture"]
        );
    }
}

/// AC-01C-1 (Issue #1990): every internal operation answers with the decided envelope and a
/// typed error; the launchers and words that name no operation do not. Runs over the golden
/// fixtures, which `protocol_fixtures_record_current_contract` binds to the binary's output.
#[test]
fn envelope_fields_and_error_codes() {
    // `identity` is covered by identity_primary_checkout and the refusal test; it has no
    // golden fixture because its output carries the fixture HOME path.
    const OPS: [&str; 8] = [
        "package-inspect",
        "project-plan",
        "project-apply",
        "package-resolve",
        "agents-plan",
        "agents-apply",
        "baseline-inventory",
        "identity",
    ];
    const KINDS: [&str; 14] = [
        "usage",
        "invalid_input",
        "input_too_large",
        "unsupported_protocol_version",
        "stale_plan",
        "permission_reference_missing",
        "not_found",
        "conflict",
        "partial",
        "io",
        "internal",
        "git_failed",
        "blocked",
        "transport",
    ];
    let dir = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join(PROTOCOL_FIXTURES);
    let mut seen = (0, 0, 0);
    for e in fs::read_dir(dir).unwrap() {
        let name = e.unwrap().file_name().to_string_lossy().into_owned();
        let f = repo_json(&format!("{PROTOCOL_FIXTURES}/{name}"));
        let word = f["args"][0].as_str().unwrap_or_default();
        let out = &f["stdout"];
        if !OPS.contains(&word) {
            assert!(
                out.get("ok").is_none(),
                "{name}: launcher or unknown word got an envelope"
            );
            continue;
        }
        let op = word.replacen('-', ".", 1);
        assert_eq!(out["protocol_version"], json!(1), "{name}");
        assert_eq!(out["operation"], json!(op), "{name}");
        let exit = f["exit"].as_i64().unwrap();
        assert_eq!(out["ok"], json!(exit == 0), "{name}");
        if exit == 0 {
            assert_eq!(out.as_object().unwrap().len(), 4, "{name}");
            assert!(out["result"].is_object(), "{name}");
            seen.0 += 1;
            continue;
        }
        assert_eq!(out.as_object().unwrap().len(), 4, "{name}");
        let err = &out["error"];
        let kind = err["kind"].as_str().unwrap();
        assert!(KINDS.contains(&kind), "{name}: {kind}");
        assert!(err["message"].is_string(), "{name}");
        assert_eq!(err["retryable"], json!(kind == "io"), "{name}");
        assert_eq!(
            exit == 2,
            kind == "usage",
            "{name}: exit {exit} kind {kind}"
        );
        let line = f["stderr"].as_str().unwrap();
        assert_eq!(line.lines().count(), 1, "{name}: stderr is one human line");
        if kind == "usage" {
            seen.2 += 1;
        } else {
            assert_eq!(line.trim_end(), err["message"].as_str().unwrap(), "{name}");
            seen.1 += 1;
        }
    }
    assert!(
        seen.0 > 0 && seen.1 > 0 && seen.2 > 0,
        "success, failure and usage cases all present: {seen:?}"
    );
}

#[test]
fn digest_normalization_examples_are_reproducible() {
    let dir = std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("tests/fixtures/protocol/v1/variants/hold-00d-5");
    let tuple = fs::read(dir.join("serde-tuple-example.json")).unwrap();
    assert_eq!(tuple, serde_json::to_vec(&(1u32, "create")).unwrap());
    let jcs = fs::read(dir.join("jcs-example.json")).unwrap();
    let n = repo_json("schemas/protocol/v1/variants/hold-00d-5-version-and-normalization.json");
    let opts = &n["normalization_options"];
    assert_eq!(
        moai_cli::digest(&tuple),
        format!(
            "sha256:{}",
            opts["serde_tuple"]["example_sha256"].as_str().unwrap()
        )
    );
    assert_eq!(
        moai_cli::digest(&jcs),
        format!(
            "sha256:{}",
            opts["rfc8785_jcs"]["example_sha256"].as_str().unwrap()
        )
    );
    assert_ne!(
        tuple, jcs,
        "the two normalizations disagree on the same data"
    );
}

// ---- Common skill package structure (02-A #1991) -------------------------------------------
//
// Run with `cargo test --offline --locked skill_package`. Reference depth (REQ-02A-08) is
// counted outside the route indirection: SKILL.md may point only at references/routes.json,
// routes.json points at the workflows, and a workflow points at no other package file.
mod skill_package {
    use super::cli;
    use serde_json::{json, Value};
    use std::{
        fs,
        path::{Path, PathBuf},
    };
    use tempfile::TempDir;

    const INTENTS: [&str; 8] = [
        "READ", "SETUP", "UPDATE", "PLAN", "RUN", "VERIFY", "AUDIT", "SYNC",
    ];
    /// Mode routes (rc.2 A1 #2160). Kept apart from INTENTS, which also counts the legacy
    /// asset map's rows by intent.
    const MODES: [&str; 2] = ["MODE-LEADER", "MODE-AGENT"];
    /// The `/moai project` route (#2496). Kept apart from INTENTS for the same reason as
    /// MODES: the legacy asset map has no rows for it.
    const PROJECT: &str = "PROJECT";
    const SPEC: &str = "docs/design-review/autonomy-modes-spec.md";
    const MANIFEST: &str = "references/source-manifest.json";
    // Claude Code-only frontmatter (https://code.claude.com/docs/en/skills, read 2026-11-08).
    const CLAUDE_ONLY: [&str; 9] = [
        "allowed-tools",
        "argument-hint",
        "disable-model-invocation",
        "user-invocable",
        "context",
        "agent",
        "hooks",
        "paths",
        "model",
    ];
    const ALLOWED_KEYS: [&str; 4] = ["name", "description", "license", "metadata"];

    fn root() -> PathBuf {
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai")
    }
    fn files(dir: &Path, base: &Path, out: &mut Vec<String>) {
        for entry in fs::read_dir(dir).unwrap() {
            let path = entry.unwrap().path();
            if path.is_dir() {
                files(&path, base, out);
            } else {
                out.push(
                    path.strip_prefix(base)
                        .unwrap()
                        .to_str()
                        .unwrap()
                        .replace('\\', "/"),
                );
            }
        }
    }
    fn package_files() -> Vec<String> {
        let mut out = Vec::new();
        files(&root(), &root(), &mut out);
        out.sort();
        out
    }
    /// Top-level frontmatter keys; nested `metadata:` lines are indented and skipped.
    fn frontmatter(text: &str) -> (Vec<(String, String)>, &str) {
        let rest = text
            .strip_prefix("---\n")
            .expect("SKILL.md starts with ---");
        let end = rest.find("\n---\n").expect("frontmatter is closed");
        let keys = rest[..end]
            .lines()
            .filter(|l| !l.starts_with(' ') && !l.trim().is_empty())
            .map(|l| {
                let (k, v) = l.split_once(':').expect("key: value");
                (k.trim().to_string(), v.trim().to_string())
            })
            .collect();
        (keys, &rest[end + 5..])
    }
    fn package_paths_in(text: &str) -> Vec<String> {
        ["references/", "workflows/", "resources/", "agents/"]
            .iter()
            .flat_map(|p| text.match_indices(p).map(|(i, _)| i))
            .map(|i| {
                text[i..]
                    .chars()
                    .take_while(|c| !c.is_whitespace() && !matches!(c, '`' | ')' | ',' | '"'))
                    .collect::<String>()
                    .trim_end_matches('.')
                    .to_string()
            })
            .collect()
    }

    #[test]
    fn skill_package_frontmatter_is_host_neutral() {
        let text = fs::read_to_string(root().join("SKILL.md")).unwrap();
        let (keys, body) = frontmatter(&text);
        let names: Vec<&str> = keys.iter().map(|(k, _)| k.as_str()).collect();
        for k in &names {
            assert!(
                ALLOWED_KEYS.contains(k),
                "frontmatter key {k} is not host-neutral"
            );
            assert!(!CLAUDE_ONLY.contains(k));
        }
        let get = |k: &str| keys.iter().find(|(n, _)| n == k).map(|(_, v)| v.as_str());
        // Agent Skills: name 1-64 [a-z0-9-] matching the directory; description 1-1024.
        let name = get("name").expect("name");
        assert_eq!(name, root().file_name().unwrap().to_str().unwrap());
        assert!(!name.is_empty() && name.len() <= 64);
        assert!(name
            .chars()
            .all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '-'));
        let description = get("description").expect("description");
        assert!(!description.is_empty() && description.chars().count() <= 1024);
        // No Claude argument substitution or dynamic shell context in the shared body.
        assert!(!text.contains("$ARGUMENTS"));
        assert!(!text.contains("!`"));
        assert!(
            text.lines().count() <= 500,
            "SKILL.md stays under 500 lines"
        );
        // One hop: the router names only the route table.
        assert_eq!(package_paths_in(body), vec!["references/routes.json"]);
        // REQ-02A-05, narrowed by the #2592 user decisions (6840661050, 6840838042): no host
        // agent files are shipped; the base agents ship only as tool-neutral sources under
        // references/agents/ (decision A), rendered per host at install time.
        assert!(package_files()
            .iter()
            .all(|p| !p.starts_with("agents/") && !p.ends_with(".toml")));
        for p in package_files().iter().filter(|p| p.contains("agents/")) {
            assert!(
                p.starts_with("references/agents/") && p.ends_with(".md"),
                "agent material outside references/agents/: {p}"
            );
        }
    }

    #[test]
    fn routes_reference_existing_workflows() {
        let routes: Value =
            serde_json::from_slice(&fs::read(root().join("references/routes.json")).unwrap())
                .unwrap();
        assert_eq!(routes["schema_version"], json!(1));
        assert_eq!(routes["classification_is_authorization"], json!(false));
        let table = routes["routes"].as_object().unwrap();
        let mut keys: Vec<&str> = table.keys().map(String::as_str).collect();
        keys.sort();
        let mut want: Vec<&str> = INTENTS
            .iter()
            .chain(MODES.iter())
            .chain(std::iter::once(&PROJECT))
            .copied()
            .collect();
        want.sort();
        assert_eq!(keys, want);
        for (intent, route) in table {
            let reference = route["reference"].as_str().unwrap();
            assert_eq!(
                reference,
                format!("../workflows/{}.md", intent.to_ascii_lowercase()),
                "{intent}"
            );
            assert!(
                root().join("references").join(reference).is_file(),
                "{intent}"
            );
            assert!(route["requires_binary"].is_boolean(), "{intent}");
            // A workflow loads no further package file (depth beyond the route table).
            let body = fs::read_to_string(root().join("references").join(reference)).unwrap();
            assert_eq!(package_paths_in(&body), Vec::<String>::new(), "{intent}");
            assert!(
                !body.contains("$ARGUMENTS") && !body.contains("!`"),
                "{intent}"
            );
            assert!(body.lines().count() <= 500, "{intent}");
        }
    }

    /// #2592 S4: each step workflow says which base agent it hands a step to, with one
    /// shared rule block, and names only agents that ship as sources.
    #[test]
    fn step_workflows_name_their_base_agent() {
        const HEAD: &str = "## Base agents\n";
        let common = |body: &str| -> String {
            let start = body.find(HEAD).expect("Base agents section");
            let rest = &body[start..];
            let end = rest.find("\n\n").map(|i| i + 2).unwrap();
            // The shared block runs to the end of its bullet list.
            const LAST: &str = "this session and say so.\n";
            let list_end = rest[end..]
                .find(LAST)
                .map(|i| end + i + LAST.len())
                .expect("shared block ends with the not-installed rule");
            rest[..list_end].to_string()
        };
        let mut sources: Vec<String> = fs::read_dir(root().join("references/agents"))
            .unwrap()
            .map(|e| {
                let text = fs::read_to_string(e.unwrap().path()).unwrap();
                let (keys, _) = frontmatter(&text);
                keys.into_iter().find(|(k, _)| k == "name").unwrap().1
            })
            .collect();
        sources.sort();
        let first = fs::read_to_string(root().join("workflows/run.md")).unwrap();
        let shared = common(&first);
        for name in &sources {
            assert!(shared.contains(&format!("`{name}`")), "{name} listed");
        }
        let named = shared.matches("`moai-").count() - shared.matches("`moai-cli`").count();
        assert_eq!(
            named,
            sources.len(),
            "the shared block lists exactly the sources"
        );
        for (route, agent) in [
            ("plan", "moai-plan"),
            ("read", "moai-research"),
            ("run", "moai-impl"),
            ("audit", "moai-audit"),
            ("sync", "moai-docs"),
        ] {
            let body = fs::read_to_string(root().join(format!("workflows/{route}.md"))).unwrap();
            assert_eq!(body.matches(HEAD).count(), 1, "{route}");
            assert_eq!(common(&body), shared, "{route}: shared block drifted");
            let own = &body[body.find(HEAD).unwrap() + shared.len()..];
            let own = &own[..own.find("\n## ").unwrap_or(own.len())];
            assert!(own.contains(&format!("`{agent}`")), "{route} names {agent}");
            // No stand-in for a missing host feature (user correction 2026-11-12T22:3xZ).
            assert!(!own.contains("moai-cli"), "{route}");
            // Only the five supported hosts (user decision 6841979953, #2592): Claude Code,
            // Codex CLI, OMP, Kiro, Cursor CLI. Removed hosts never come back here.
            let words: Vec<String> = own
                .split(|c: char| !c.is_ascii_alphanumeric())
                .map(str::to_ascii_lowercase)
                .collect();
            for removed in ["opencode", "goose", "zcode", "pi"] {
                assert!(
                    !words.iter().any(|w| w == removed),
                    "{route} names the unsupported host {removed}"
                );
            }
        }
    }

    /// The mode workflow behind a route, after checking the route itself (AC-1, #2160).
    fn mode_workflow(route: &str) -> String {
        let routes: Value =
            serde_json::from_slice(&fs::read(root().join("references/routes.json")).unwrap())
                .unwrap();
        let r = &routes["routes"][route];
        assert_eq!(r["requires_binary"], json!(false), "{route}");
        let reference = r["reference"].as_str().unwrap();
        let skill = fs::read_to_string(root().join("SKILL.md")).unwrap();
        assert!(skill.contains(route), "SKILL.md names {route}");
        fs::read_to_string(root().join("references").join(reference)).unwrap()
    }
    /// AC-2: every `SPEC §N` a mode workflow cites is a section of the SPEC, and the
    /// workflow does not restate the claim procedure the SPEC defines.
    fn cites_spec_sections_only(body: &str, route: &str) {
        let spec = fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(SPEC)).unwrap();
        let cited: Vec<&str> = body
            .match_indices("SPEC §")
            .map(|(i, m)| {
                let rest = &body[i + m.len()..];
                &rest[..rest
                    .find(|c: char| !c.is_ascii_digit())
                    .unwrap_or(rest.len())]
            })
            .collect();
        assert!(!cited.is_empty(), "{route} cites the SPEC by section");
        for n in cited {
            assert!(
                spec.lines().any(|l| l.starts_with(&format!("## {n}. "))),
                "{route} cites SPEC §{n}, which the SPEC does not have"
            );
        }
        for restated in ["moai-claim:", "lease_until", "journal", "comment id"] {
            assert!(!body.contains(restated), "{route} restates {restated}");
        }
    }

    /// AC-1 (#2334): the mode workflows bind the /goal step to the tested
    /// generator instead of free-form wording — leader feeds the state it read
    /// to tools/gen_goal_block.py and prints its `leader` line, with the
    /// keep-set exclusion rule stated as the generator's rule; agent binds its
    /// card-scoped lines to the generator's `agent` kind.
    #[test]
    fn goal_step_is_bound_to_the_goal_block_generator() {
        let leader = mode_workflow("MODE-LEADER");
        assert!(
            leader.contains("`tools/gen_goal_block.py`"),
            "leader workflow must name the generator: {leader}"
        );
        assert!(
            leader.contains("`leader` goal line"),
            "leader workflow must print the generator's leader line: {leader}"
        );
        assert!(
            leader.contains("keep-set") && leader.contains("reason"),
            "leader workflow must state the keep-set exclusion with reasons: {leader}"
        );
        assert!(
            leader.contains("never hand-edit the generated line"),
            "leader workflow must forbid editing the generated line: {leader}"
        );
        let agent = mode_workflow("MODE-AGENT");
        assert!(
            agent.contains("`tools/gen_goal_block.py`"),
            "agent workflow must name the generator: {agent}"
        );
        assert!(
            agent.contains("`agent` goal lines"),
            "agent workflow must print the generator's agent lines: {agent}"
        );
        assert!(
            agent.contains("never hand-edit the generated lines"),
            "agent workflow must forbid editing the generated lines: {agent}"
        );
        // The generator itself still enforces the rule: a keep-set entry without
        // a reason is refused, and excluded numbers leave the condition.
        let run = |input: &str| {
            let mut out = std::process::Command::new("python3")
                .args(["-I", "tools/gen_goal_block.py"])
                .stdin(std::process::Stdio::piped())
                .stdout(std::process::Stdio::piped())
                .stderr(std::process::Stdio::piped())
                .spawn()
                .unwrap();
            crate::common::feed_stdin(&mut out, input.as_bytes());
            out.wait_with_output().unwrap()
        };
        let bad = run(r#"{"milestone":"m","keep_set":[{"number":7}]}"#);
        assert!(
            !bad.status.success(),
            "keep-set entry without a reason must be refused"
        );
        let good = run(
            r#"{"milestone":"m1","issues":[{"number":1,"state":"open"},{"number":2,"state":"open"}],"prs":[{"number":3,"state":"open"}],"keep_set":[{"number":2,"reason":"operator-held"}]}"#,
        );
        let v: Value = serde_json::from_slice(&good.stdout).unwrap();
        // The excluded entry drops out of the counts (2 issues -> 1) and the
        // condition carries its number and reason.
        assert_eq!(v["counts"]["open_issues"], json!(1));
        assert_eq!(v["counts"]["open_prs"], json!(1));
        assert_eq!(v["counts"]["keep_set"], json!(1));
        let cond = v["leader"]["condition"].as_str().unwrap();
        assert!(cond.contains("#2(operator-held)"), "{cond}");
        assert!(v["keep_set_excluded"][0]["reason"] == json!("operator-held"));
    }

    #[test]
    fn mode_leader_route_loads_its_workflow() {
        let body = mode_workflow("MODE-LEADER");
        let skill = fs::read_to_string(root().join("SKILL.md")).unwrap();
        assert!(
            skill.contains("`--mode-leader`") && skill.contains("`--autonomy full|semi`"),
            "the canonical leader flag is --mode-leader (#2286): {skill}"
        );
        // REQ-A1-01: the loaded mode ignores the other mode's instructions.
        assert!(body.contains("ignore instructions for any other mode"));
        assert!(
            body.contains("`/goal <condition>`") && body.contains("Never run `/goal` yourself")
        );
        cites_spec_sections_only(&body, "MODE-LEADER");
    }

    #[test]
    fn mode_agent_route_loads_its_workflow() {
        let body = mode_workflow("MODE-AGENT");
        let skill = fs::read_to_string(root().join("SKILL.md")).unwrap();
        assert!(
            skill.contains("`--mode-agent`") && skill.contains("`$moai --mode-agent …`"),
            "the canonical agent flag is --mode-agent (#2286): {skill}"
        );
        assert!(body.contains("ignore instructions for any other mode"));
        assert!(
            body.contains("`$moai --mode-agent`"),
            "Codex may enter agent mode"
        );
        cites_spec_sections_only(&body, "MODE-AGENT");
    }

    /// #2619: an open question is sorted before anyone is asked — a fact is answered, a
    /// one-way or five-category choice waits for the user, any other choice is the leader's
    /// recorded decision, and a handed-over reversible choice runs on its default after its
    /// deadline. A HOLD may name only the five categories or `irreversible`.
    #[test]
    fn open_questions_are_sorted_and_holds_name_five_categories() {
        let skill = fs::read_to_string(root().join("SKILL.md")).unwrap();
        let start = skill
            .find("the five HOLD categories")
            .expect("SKILL.md names the five HOLD categories");
        let clause = &skill[start..start + skill[start..].find(", or that changes").unwrap()];
        let named: Vec<&str> = clause.split('`').skip(1).step_by(2).collect();
        assert_eq!(
            named,
            ["money", "legal", "external_contract", "secret", "real_data"],
            "{clause}"
        );
        for line in [
            "a question that a read, a command or a document answers is not a decision",
            "goes to the user and waits — no deadline applies it by default",
            "the leader decides it within the chosen autonomy level, records it, and the user may reverse it afterwards",
            "a lane asks the leader, not the user",
            "the default applies once the deadline passes unanswered",
            "A HOLD names one of the five categories, `guarded` or `irreversible`; no other reason declares one.",
            // Audit B2 on #2620: what stays with the user although it is reversible.
            "or that changes a `guarded` setting — permissions, keys, the real HOME and global settings, deletions or repository protection settings — goes to the user and waits",
            "Cost, permissions, external transmission and destructive effects beyond the authorized scope are case (2).",
            // Audit B1 on #2620: the leader's own decision stays inside the autonomy level.
            "what the chosen autonomy level still reserves for the user (new Issues, lifting a HOLD, merge-policy changes and releases at `semi`) goes to the user as case (4)",
            "Independent audit and the CI gate are never delegated.",
            "its options, the default, whether it is reversible, its category (`none`, `guarded` or one of the five), its deadline and who decided; a decision without a record does not count.",
            // #2678: a reported figure comes from a named command, and silence is not zero.
            "Take every number and state in a report from a command run for that report, and name the command next to it (Issue #2678).",
            "An empty output is not zero: say whether the command prints nothing for \"none\" or also on a failure, and count a paged list only after reading every page.",
            "After a write, read the written thing back and report what the read shows, not what was sent.",
            "Write a time only from a clock read in the same step (`date -u`).",
        ] {
            assert!(skill.contains(line), "SKILL.md lacks: {line}");
        }
        let leader = mode_workflow("MODE-LEADER");
        let agent = mode_workflow("MODE-AGENT");
        for body in [&leader, &agent] {
            assert!(
                body.contains("sort it in the order SKILL.md gives (Issue #2619)")
                    || body.contains("sort the question in the order SKILL.md gives (Issue #2619)")
            );
        }
        assert!(leader.contains("never decide a HOLD for the lane"));
        // #2678: every posted table row names its source command; silence is "no result".
        assert!(leader.contains("Every row of a status, close or decision table you post carries a source: the command whose output gave its number or state"));
        assert!(leader.contains(
            "A row whose command printed nothing says \"no result\" and why, never a bare 0."
        ));
        for line in [
            "within the autonomy level SPEC §6 gives — at `semi` the choices listed below for that level still go to the user",
            "record it with every field SKILL.md names",
            "At `semi`, new Issues, lifting a HOLD, merge-policy changes and releases also go to the user.",
        ] {
            assert!(leader.contains(line), "mode-leader.md lacks: {line}");
        }
        assert!(agent.contains("send a choice to the leader — not to the user"));
        assert!(agent.contains(
            "never declare a HOLD outside the five categories, `guarded` and `irreversible`"
        ));
    }

    /// #2516 (leader decision B, comment 6823474405): the Orca session-control rules live in
    /// both mode workflows as one section, kept identical, with the rules a lane and a leader
    /// rely on pinned by sentence.
    #[test]
    fn both_modes_carry_the_same_orca_session_control_section() {
        let section = |route: &str| {
            let body = mode_workflow(route);
            let start = body
                .find("## Session control with Orca")
                .unwrap_or_else(|| panic!("{route} lacks the Orca section"));
            let rest = &body[start..];
            let end = rest[2..].find("\n## ").map_or(rest.len(), |i| i + 2);
            rest[..end].to_string()
        };
        let agent = section("MODE-AGENT");
        assert_eq!(
            agent,
            section("MODE-LEADER"),
            "the Orca section differs between mode-agent.md and mode-leader.md"
        );
        for line in [
            "installed is not chosen",
            "report the session-control step `not_run`",
            "`orca skills get orca-cli`",
            "`orca agent-context --json`",
            "the schema wins",
            "never overwrite another session's unsubmitted input",
            "`accepted: true` proves only that the input was accepted",
            "A send you could not confirm is `uncertain`; never resend on silence",
            "paging while the read is `limited`",
            "Never answer another session's permission or trust prompt",
            "without the user's confirmation",
            "silence, a timeout or an `unverifiable` state never authorises recovery",
            "Name every terminal by its handle on every command",
            "never by its title",
            "never send the same input to an old and a new handle",
            "Close only a terminal you created",
        ] {
            assert!(agent.contains(line), "Orca section lacks: {line}");
        }
    }

    /// #2636 (parent #2634): the orca and aside adapters carry when to use each tool, the
    /// commands that print its version-matched guide (never a copied guide), the lines MoAI
    /// keeps, and the fallback when the tool is missing. Each adapter names its official web
    /// page and the CLI version it was checked against.
    #[test]
    fn adapters_point_to_version_matched_guides_and_keep_the_safety_lines() {
        let read = |name: &str| {
            fs::read_to_string(root().join(format!("references/adapters/{name}.md")))
                .unwrap_or_else(|e| panic!("adapter {name}: {e}"))
        };
        let mut adapters: Vec<String> = package_files()
            .into_iter()
            .filter(|p| p.starts_with("references/adapters/"))
            .collect();
        adapters.sort();
        assert_eq!(
            adapters,
            [
                "references/adapters/aside.md",
                "references/adapters/orca.md"
            ]
        );
        for (name, tool) in [("orca", "Orca"), ("aside", "Aside")] {
            let text = read(name);
            let missing = format!("## When {tool} is missing");
            for heading in [
                "## When to use it",
                "## Open the version-matched guide first",
                "## Lines MoAI keeps",
                missing.as_str(),
                "## Sources",
            ] {
                assert!(text.contains(heading), "{name} lacks {heading}");
            }
            assert!(text.contains("MoAI does not copy"), "{name}");
            assert!(text.contains("Never invent a flag from memory."), "{name}");
            assert!(text.contains("record"), "{name}");
            assert!(text.contains("`not_run`"), "{name}");
            assert!(text.contains("(read 2026-11-13"), "{name}");
        }
        let orca = read("orca");
        for line in [
            "installed is not chosen",
            "`orca skills get orca-cli`",
            "`orca agent-context --json`",
            "the schema wins",
            "never overwrite another session's unsubmitted input",
            "Name every terminal by its handle",
            "Never answer another session's permission or trust prompt",
            "never resend on silence",
            "Use `git worktree` and the host's own terminal instead",
            "https://orca.dev/docs/cli/overview",
            "orca 1.4.257",
            "MoAI uses Orca through its CLI only",
        ] {
            assert!(orca.contains(line), "orca adapter lacks: {line}");
        }
        let aside = read("aside");
        for line in [
            "`aside guide`",
            "`aside guide repl` before the first `aside repl`",
            "`aside skills list`",
            "Keep the session permission at its default, Guard",
            "`--permission full-access`",
            "needs the user's confirmation every time",
            "`aside session delete`, which cannot be undone",
            "never edit Aside's memory files",
            "Never print, upload, commit or share the CLI credential file.",
            "Use a read-only web fetch instead",
            // User decision #2634 6845653775: Aside through its CLI only, never MCP.
            "## CLI only, never MCP",
            "MoAI never runs `aside mcp`, never registers it in a host's MCP settings and offers no path to add it",
            "https://docs.aside.com/help/security",
            "Aside CLI 1.26.1111.1046",
        ] {
            assert!(aside.contains(line), "aside adapter lacks: {line}");
        }
        // User decision #2634 6845653775 (audit 6845704355 B1): no MCP install path in either
        // adapter — no host command that adds, reads or removes an MCP entry.
        for (name, text) in [("orca", &orca), ("aside", &aside)] {
            for gone in ["mcp add", "mcp get", "mcp remove", "mcp_servers"] {
                assert!(!text.contains(gone), "{name} adapter still names `{gone}`");
            }
        }
    }

    #[test]
    fn mode_leader_refuses_codex_with_decision_d3() {
        // REQ-A1-02: leader mode is Claude Code only (D-3); the refusal comes first.
        let body = mode_workflow("MODE-LEADER");
        let procedure = &body[body.find("## Procedure").unwrap()..];
        let first = procedure.lines().find(|l| l.starts_with("1. ")).unwrap();
        assert!(
            first.contains("Codex CLI") && first.contains("refuse"),
            "{first}"
        );
        assert!(
            first.contains("D-3") && first.contains("Claude Code only"),
            "{first}"
        );
        assert!(
            first.contains("`$moai --mode-agent`"),
            "the refusal points to agent mode"
        );
    }

    fn retired_mode_spelling(text: &str) -> Option<&'static str> {
        // Legacy aliases remain aliases when their operands use tabs or repeated spaces.
        let normalized = text.split_whitespace().collect::<Vec<_>>().join(" ");
        [
            "--agent-mode",
            "--leader-mode",
            "--mode agent",
            "--mode leader",
        ]
        .into_iter()
        .find(|retired| normalized.contains(*retired))
    }

    #[test]
    fn retired_mode_aliases_cannot_hide_behind_tabs_or_repeated_spaces() {
        for (text, retired) in [
            (
                "For compatibility, `--mode\tagent` also selects MODE-AGENT.",
                "--mode agent",
            ),
            (
                "For compatibility, `--mode  leader` also selects MODE-LEADER.",
                "--mode leader",
            ),
            ("`--mode  agent`", "--mode agent"),
            ("`--mode\tleader`", "--mode leader"),
        ] {
            assert_eq!(retired_mode_spelling(text), Some(retired), "{text:?}");
        }
        assert_eq!(
            retired_mode_spelling("`--mode-agent` or `--mode-leader`"),
            None
        );
    }

    /// AC-2 (#2286, user decision 2026-11-10): the routing line sends exactly the two
    /// spellings `--mode-agent` and `--mode-leader` to their routes; the earlier
    /// spellings (`--agent-mode`, `--leader-mode`, `--mode agent|leader`) appear nowhere in
    /// the package. The routes file must carry both MODE routes for those picks to resolve.
    #[test]
    fn mode_routing_flags_map_to_routes() {
        let skill = fs::read_to_string(root().join("SKILL.md")).unwrap();
        assert!(
            skill.contains("`--mode-agent` or `--mode-leader`"),
            "SKILL.md must route the canonical flags: {skill}"
        );
        // Every package file is scanned except the source manifest, whose `reason` strings
        // keep the history of earlier spellings on purpose (#2286 pre-audit B1).
        for file in package_files()
            .into_iter()
            .filter(|p| p != "references/source-manifest.json")
        {
            let text = fs::read_to_string(root().join(&file)).unwrap();
            assert_eq!(
                retired_mode_spelling(&text),
                None,
                "retired mode spelling is still in the package file {file}"
            );
        }
        // Both canonical picks must resolve to their MODE routes.
        assert!(
            skill.contains("pick the MODE-AGENT or MODE-LEADER route")
                || (skill.contains("MODE-AGENT") && skill.contains("MODE-LEADER")),
            "SKILL.md must name both routes for the flags: {skill}"
        );
        let routes: Value =
            serde_json::from_slice(&fs::read(root().join("references/routes.json")).unwrap())
                .unwrap();
        let route = |k: &str| routes["routes"][k]["reference"].as_str().unwrap_or("");
        assert_eq!(route("MODE-AGENT"), "../workflows/mode-agent.md");
        assert_eq!(route("MODE-LEADER"), "../workflows/mode-leader.md");
        assert_eq!(
            routes["routes"]["MODE-AGENT"]["requires_binary"],
            json!(false)
        );
        assert_eq!(
            routes["routes"]["MODE-LEADER"]["requires_binary"],
            json!(false)
        );
    }

    /// #2286 CLOSE-TABLE item 1: the repository's own instruction files carry no retired mode
    /// spelling either, so a lane reading them types only `--mode-agent` / `--mode-leader`.
    #[test]
    fn repository_instruction_files_carry_no_retired_mode_spelling() {
        let repo = Path::new(env!("CARGO_MANIFEST_DIR"));
        for file in ["AGENTS.md", "AGENTS.local.md"] {
            let text = fs::read_to_string(repo.join(file)).unwrap();
            assert_eq!(
                retired_mode_spelling(&text),
                None,
                "retired mode spelling is still in {file}"
            );
            assert!(
                text.contains("--mode-agent") || file == "AGENTS.md",
                "{file}"
            );
        }
    }

    /// #2496 AC-1: `/moai project` is a route, and its workflow carries the three stages —
    /// detect and interview, the index-build suggestion, and detection-only LSP·MCP·DB.
    #[test]
    fn project_route_names_its_three_stages() {
        let routes: Value =
            serde_json::from_slice(&fs::read(root().join("references/routes.json")).unwrap())
                .unwrap();
        let r = &routes["routes"][PROJECT];
        assert_eq!(r["reference"], json!("../workflows/project.md"));
        assert_eq!(r["requires_binary"], json!(true));
        let skill = fs::read_to_string(root().join("SKILL.md")).unwrap();
        assert!(skill.contains("PROJECT"), "SKILL.md lists the route");
        assert!(
            skill.contains("`/moai project`"),
            "SKILL.md names the entry"
        );
        let body = fs::read_to_string(root().join("workflows/project.md")).unwrap();
        for stage in [
            "**Detect.** Run `moai-cli project-detect <project>`",
            "**Interview by kind.**",
            "New project:",
            "Existing project:",
            "\"value\": \"HOLD\"",
            "**Offer the code map.**",
            "run it only after the user agrees",
            "`lsp`",
            "`mcp_servers`",
            "`databases`",
            "MoAI never creates or edits an MCP configuration",
            "no product, structure or tech documents are written",
            // #2535 AC-1 (items 1 and 3) and AC-2 (items 2 and 4).
            "with the plan output's `result` object on stdin, not the whole response",
            "`invalid typed JSON input`",
            "When `result.profile` is `null`",
            "create the profile first",
            "Detection reads the project root only",
            "`project_kind` speaks of code only",
            // #2535 item 7: a no-baseline conflict is taken in two steps, never by hand.
            "has no baseline yet",
            "Apply in two steps",
            "Never edit `project.yaml` by hand to get past the conflict",
            // #2639: read before asking, three or four at a time, the five common axes, a
            // file preview before approval, and hand edits never overwritten on a re-run.
            "Read before you ask (Issue #2639)",
            "is confirmed with the user, not asked",
            "three or four questions at a time and at most three rounds, one axis per question",
            "what must never be touched, what counts as done (a check the user will run themselves), which mistakes a review should always stop, and what may run without asking against what must be asked",
            "sorted the way SKILL.md sorts an open question",
            "each new file with its first lines and each changed file as its difference",
            "a managed part the user edited by hand is planned as `conflict` and never overwritten",
            // #2639 REVIEW.md draft from the interview answers.
            "The same plan carries a `REVIEW.md` section when the interview recorded `review_blocks` or `protected`",
            "an unanswered axis is never written as a rule",
            // #2639 project agents: from the answers, never a base name, host files only.
            "**Offer project agents (Issue #2639).**",
            "its `role_id` is `prj-<name>` for the project's own job",
            "a project agent never shadows a user-layer `moai-*` agent",
            "for any other host, say that none is written",
            "apply with `generation-apply` only after the user agrees",
            "planned `unchanged` while the answers stay the same, and `conflict` once they change too",
        ] {
            assert!(body.contains(stage), "project.md lacks: {stage}");
        }
        // #2535 item 5: skill files are read with the host's file tool, not `ls`/`cat`.
        assert!(
            skill.contains("with the host's file-reading tool, not with a shell command"),
            "SKILL.md names the file-reading tool"
        );
    }

    #[test]
    fn source_manifest_covers_every_package_file() {
        let manifest: Value =
            serde_json::from_slice(&fs::read(root().join(MANIFEST)).unwrap()).unwrap();
        assert_eq!(manifest["schema_version"], json!(1));
        let entries = manifest["files"].as_object().unwrap();
        let mut covered: Vec<&str> = entries.keys().map(String::as_str).collect();
        covered.sort();
        let mut present = package_files();
        present.retain(|p| p != MANIFEST);
        assert_eq!(
            covered, present,
            "every package file except the manifest has an entry"
        );
        for (path, entry) in entries {
            let bytes = fs::read(root().join(path)).unwrap();
            assert_eq!(
                format!("sha256:{}", entry["sha256"].as_str().unwrap()),
                moai_cli::digest(&bytes),
                "{path}: manifest sha256 is stale"
            );
            for key in ["kind", "template_id", "adoption", "reason"] {
                assert!(entry["origin"][key].is_string(), "{path}: origin.{key}");
            }
        }
        let legacy = &manifest["legacy_source"];
        assert_eq!(legacy["license"], json!("Apache-2.0"));
        assert_eq!(legacy["revision"].as_str().unwrap().len(), 40);
        assert_eq!(legacy["blob_check"]["mismatched"], json!([]));
        assert_eq!(legacy["blob_check"]["missing_at_revision"], json!([]));
        // The full legacy list lives outside the package (installs never carry it), pinned
        // by sha256; the package keeps only counts.
        let map_ref = &legacy["asset_map"];
        let map_path =
            Path::new(env!("CARGO_MANIFEST_DIR")).join(map_ref["path"].as_str().unwrap());
        assert!(!map_path.starts_with(root()));
        let map_bytes = fs::read(&map_path).unwrap();
        assert_eq!(
            format!("sha256:{}", map_ref["sha256"].as_str().unwrap()),
            moai_cli::digest(&map_bytes),
            "legacy asset map changed without updating the manifest"
        );
        let map: Value = serde_json::from_slice(&map_bytes).unwrap();
        assert_eq!(map["revision"], legacy["revision"]);
        assert_eq!(map["inventory_sha256"], legacy["inventory_sha256"]);
        let assets = map["assets"].as_array().unwrap();
        assert_eq!(map_ref["rows"], json!(assets.len()));
        assert_eq!(legacy["blob_check"]["rows"], json!(assets.len()));
        let intents: Vec<String> = INTENTS.iter().map(|i| i.to_ascii_lowercase()).collect();
        for i in &intents {
            let n = assets
                .iter()
                .filter(|a| a["intents"].as_array().unwrap().contains(&json!(i)))
                .count();
            assert_eq!(map_ref["rows_by_intent"][i], json!(n), "{i}");
        }
        for a in assets {
            // Unmapped originals stay unchecked; nothing is marked done or dropped.
            assert_eq!(a["status"], json!("unchecked"), "{}", a["path"]);
            assert_eq!(a["sha256"].as_str().unwrap().len(), 64);
            let mapped = a["intents"].as_array().unwrap();
            assert!(mapped
                .iter()
                .all(|i| intents.contains(&i.as_str().unwrap().to_string())));
            if mapped.is_empty() {
                assert!(a["proposed_route"].is_string(), "{}", a["path"]);
            }
        }
        // Pointers only: no legacy bytes are copied into the package.
        assert!(package_files()
            .iter()
            .all(|p| !p.contains("source-assets/")));
    }

    #[test]
    fn package_inspect_rejects_missing_skill_md() {
        let real = root();
        let args = ["package-inspect", real.to_str().unwrap()];
        let first = cli(&args, &Value::Null);
        assert!(
            first.status.success(),
            "{}",
            String::from_utf8_lossy(&first.stderr)
        );
        assert_eq!(
            first.stdout,
            cli(&args, &Value::Null).stdout,
            "digest is repeatable"
        );
        let copy = TempDir::new().unwrap();
        fs::create_dir(copy.path().join("workflows")).unwrap();
        fs::write(copy.path().join("workflows/read.md"), "# READ").unwrap();
        let out = cli(
            &["package-inspect", copy.path().to_str().unwrap()],
            &Value::Null,
        );
        assert!(!out.status.success());
        assert!(String::from_utf8_lossy(&out.stderr).contains("SKILL.md"));
    }
}

// ---- identity (Issue #1988, 01-A) ---------------------------------------------------------

/// Runs `moai-cli identity <path>` with a fixture HOME and no user git config. `path_env`
/// replaces PATH when given (used to make git unavailable).
fn ident(
    path: &std::path::Path,
    home: &std::path::Path,
    path_env: Option<&str>,
) -> (std::process::Output, Value) {
    ident_env(path, home, path_env, &[])
}
/// Like `ident`, with extra caller environment variables (Issue #2033).
fn ident_env(
    path: &std::path::Path,
    home: &std::path::Path,
    path_env: Option<&str>,
    envs: &[(&str, &str)],
) -> (std::process::Output, Value) {
    let mut cmd = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    cmd.arg("identity")
        .arg(path)
        .env("HOME", home)
        .env("GIT_CONFIG_GLOBAL", "/dev/null")
        .env("GIT_CONFIG_NOSYSTEM", "1")
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .envs(envs.iter().copied());
    if let Some(p) = path_env {
        cmd.env("PATH", p);
    }
    let out = cmd.output().unwrap();
    // The identity sits under the envelope's `result` (01-C #1990); a failure has none.
    let v = serde_json::from_slice::<Value>(&out.stdout)
        .map(|v| v["result"].clone())
        .unwrap_or(Value::Null);
    (out, v)
}
fn canon(p: &std::path::Path) -> String {
    fs::canonicalize(p).unwrap().to_str().unwrap().to_string()
}
fn committed_repo(dir: &std::path::Path) {
    git_in(dir, &["init", "-q", "-b", "main"]);
    write_file(dir, "a.txt", b"a\n");
    git_in(dir, &["add", "a.txt"]);
    git_in(dir, &["commit", "-q", "-m", "fixture"]);
}
/// REQ-01A-05: no returned string may point at the Go generation's state.
fn assert_no_go_state(v: &Value) {
    for (k, x) in v.as_object().unwrap() {
        if let Some(s) = x.as_str() {
            assert!(
                !s.ends_with("/.moai") && !s.contains("/.moai/") && !s.contains("MOAI_HOME"),
                "{k} points at Go state: {s}"
            );
        }
    }
}

#[test]
fn identity_primary_checkout() {
    let repo = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    committed_repo(repo.path());
    fs::create_dir(repo.path().join("sub")).unwrap();
    let before = (tree_state(repo.path()), tree_state(home.path()));
    let (out, v) = ident(repo.path(), home.path(), None);
    assert!(out.status.success(), "{:?}", out.stderr);
    // AC-01C-1: identity answers with the same envelope as the other internal operations.
    let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(envelope["protocol_version"], json!(1));
    assert_eq!(envelope["operation"], json!("identity"));
    assert_eq!(envelope["ok"], json!(true));
    let git_dir = canon(&repo.path().join(".git"));
    assert_eq!(v["identity_status"], json!("ok"));
    assert_eq!(v["workspace_kind"], json!("primary"));
    assert_eq!(v["clone_anchor"], json!(git_dir));
    assert_eq!(v["workspace_git_dir"], json!(git_dir));
    assert_eq!(v["clone_id"], v["workspace_id"]);
    assert_eq!(v["clone_id"], json!(moai_cli::digest(git_dir.as_bytes())));
    assert_eq!(v["work_tree"], json!(canon(repo.path())));
    assert_eq!(
        v["project_state_root"],
        json!(format!("{}/.moai-cli", canon(repo.path())))
    );
    assert_eq!(
        v["home_state_root"],
        json!(home.path().join(".moai-cli").to_str().unwrap())
    );
    assert_eq!(v["profile_status"], json!("missing"));
    assert_eq!(v["project_id"], Value::Null);
    assert_eq!(v["writes"], json!(false));
    assert_no_go_state(&v);
    // A subdirectory resolves to the same workspace; two runs print identical bytes.
    let (sub, _) = ident(&repo.path().join("sub"), home.path(), None);
    assert_eq!(sub.stdout, out.stdout);
    assert_eq!(ident(repo.path(), home.path(), None).0.stdout, out.stdout);
    assert_eq!(
        before,
        (tree_state(repo.path()), tree_state(home.path())),
        "identity wrote to the repository or HOME"
    );
    assert!(!home.path().join(".moai-cli").exists());
    // The profile's project_id is reported separately from the path-derived ids.
    create(&repo, &profile(None));
    let (_, v) = ident(repo.path(), home.path(), None);
    assert_eq!(v["profile_status"], json!("present"));
    assert_eq!(v["project_id"], json!("demo"));
}

#[test]
fn identity_linked_worktree_shares_clone_id() {
    let repo = TempDir::new().unwrap();
    let side = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    committed_repo(repo.path());
    let wt = side.path().join("wt");
    git_in(
        repo.path(),
        &["worktree", "add", "-q", "-b", "lane", wt.to_str().unwrap()],
    );
    let (o1, primary) = ident(repo.path(), home.path(), None);
    let (o2, linked) = ident(&wt, home.path(), None);
    assert!(o1.status.success() && o2.status.success());
    assert_eq!(primary["workspace_kind"], json!("primary"));
    assert_eq!(linked["workspace_kind"], json!("linked"));
    assert_eq!(primary["clone_id"], linked["clone_id"]);
    assert_eq!(primary["clone_anchor"], linked["clone_anchor"]);
    assert_ne!(primary["workspace_id"], linked["workspace_id"]);
    assert_ne!(linked["clone_anchor"], linked["workspace_git_dir"]);
    assert_eq!(linked["work_tree"], json!(canon(&wt)));
    assert_eq!(
        linked["project_state_root"],
        json!(format!("{}/.moai-cli", canon(&wt)))
    );
    assert_no_go_state(&linked);
}

#[test]
fn identity_bare_separate_gitdir_gitfile() {
    let base = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    let b = base.path();
    // Bare repository: no work tree, so no project state root is offered.
    git_in(b, &["init", "-q", "--bare", "bare.git"]);
    let (out, v) = ident(&b.join("bare.git"), home.path(), None);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["workspace_kind"], json!("bare"));
    assert_eq!(v["clone_anchor"], json!(canon(&b.join("bare.git"))));
    assert_eq!(v["work_tree"], Value::Null);
    assert_eq!(v["project_state_root"], Value::Null);
    // --separate-git-dir: `.git` is a gitfile, the repository lives elsewhere.
    git_in(
        b,
        &[
            "init",
            "-q",
            "--separate-git-dir",
            b.join("sep.git").to_str().unwrap(),
            b.join("sep-wt").to_str().unwrap(),
        ],
    );
    assert!(b.join("sep-wt/.git").is_file());
    let (_, v) = ident(&b.join("sep-wt"), home.path(), None);
    assert_eq!(v["workspace_kind"], json!("primary"));
    assert_eq!(v["clone_anchor"], json!(canon(&b.join("sep.git"))));
    assert_eq!(v["work_tree"], json!(canon(&b.join("sep-wt"))));
    // Hand-written gitfile pointing at a moved repository.
    let m = b.join("moved");
    fs::create_dir(&m).unwrap();
    committed_repo(&m);
    fs::rename(m.join(".git"), b.join("moved.git")).unwrap();
    fs::write(
        m.join(".git"),
        format!("gitdir: {}\n", b.join("moved.git").to_str().unwrap()),
    )
    .unwrap();
    let (_, v) = ident(&m, home.path(), None);
    assert_eq!(v["workspace_kind"], json!("primary"));
    assert_eq!(v["clone_anchor"], json!(canon(&b.join("moved.git"))));
    assert_eq!(v["work_tree"], json!(canon(&m)));
    // Inside a non-bare git dir there is no work tree either.
    let (_, v) = ident(&b.join("moved.git"), home.path(), None);
    assert_eq!(v["workspace_kind"], json!("git_dir"));
    assert_eq!(v["project_state_root"], Value::Null);
}

#[test]
fn identity_outside_git_and_missing_git_binary() {
    let plain = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    let (out, v) = ident(plain.path(), home.path(), None);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["identity_status"], json!("no_git"));
    for k in [
        "workspace_kind",
        "clone_anchor",
        "workspace_git_dir",
        "work_tree",
        "clone_id",
        "workspace_id",
        "project_id",
        "project_state_root",
    ] {
        assert_eq!(v[k], Value::Null, "{k} must not be guessed");
    }
    let repo = TempDir::new().unwrap();
    committed_repo(repo.path());
    let (out, v) = ident(repo.path(), home.path(), Some(""));
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["identity_status"], json!("git_unavailable"));
    assert_eq!(v["clone_id"], Value::Null);
    assert_eq!(v["project_state_root"], Value::Null);
}

/// Issue #2033 AC-IDF-1: caller environment that would redirect git or change discovery
/// must not move any reported boundary.
#[test]
fn identity_ignores_injected_git_environment() {
    let repo = TempDir::new().unwrap();
    let elsewhere = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    committed_repo(repo.path());
    fs::create_dir(repo.path().join("sub")).unwrap();
    let x = elsewhere.path().to_str().unwrap();
    let worktree_params = format!("'core.worktree'='{x}'");
    let ceiling = repo.path().to_str().unwrap();
    // Measured on git 2.54: injected core.bare=true turns a git dir into "bare" through
    // either config channel, and a ceiling at the repository hides it from a subdirectory.
    let cases: [&[(&str, &str)]; 6] = [
        &[
            ("GIT_CONFIG_COUNT", "1"),
            ("GIT_CONFIG_KEY_0", "core.worktree"),
            ("GIT_CONFIG_VALUE_0", x),
        ],
        &[
            ("GIT_CONFIG_COUNT", "1"),
            ("GIT_CONFIG_KEY_0", "core.bare"),
            ("GIT_CONFIG_VALUE_0", "true"),
        ],
        &[("GIT_CONFIG_PARAMETERS", &worktree_params)],
        &[("GIT_CONFIG_PARAMETERS", "'core.bare'='true'")],
        &[("GIT_CEILING_DIRECTORIES", ceiling)],
        &[("GIT_DISCOVERY_ACROSS_FILESYSTEM", "1")],
    ];
    for path in [repo.path().join("sub"), repo.path().join(".git")] {
        let (out, clean) = ident(&path, home.path(), None);
        assert!(out.status.success(), "{:?}", out.stderr);
        assert_eq!(clean["identity_status"], json!("ok"));
        for envs in cases {
            let (out, v) = ident_env(&path, home.path(), None, envs);
            assert!(out.status.success(), "{envs:?}: {:?}", out.stderr);
            assert_eq!(v, clean, "{path:?} {envs:?} moved the boundary");
        }
    }
}

/// Issue #2033 AC-IDF-2: git present but failing is a `git_failed` error envelope, not the
/// `no_git` status, and the message carries git's summary without any path.
#[test]
fn identity_reports_git_failure_separately_from_no_git() {
    let repo = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    committed_repo(repo.path());
    // safe.directory refusal; GIT_CONFIG_GLOBAL=/dev/null leaves no exception configured.
    let (out, _) = ident_env(
        repo.path(),
        home.path(),
        None,
        &[("GIT_TEST_ASSUME_DIFFERENT_OWNER", "1")],
    );
    assert_eq!(out.status.code(), Some(1), "{:?}", out.stderr);
    let env: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(env["operation"], json!("identity"));
    assert_eq!(env["ok"], json!(false));
    assert!(env.get("result").is_none());
    let err = &env["error"];
    assert_eq!(err["kind"], json!("git_failed"));
    assert_eq!(err["retryable"], json!(false));
    let message = err["message"].as_str().unwrap();
    assert!(message.contains("dubious ownership"), "{message}");
    let stderr = String::from_utf8_lossy(&out.stderr);
    for text in [message, stderr.as_ref()] {
        assert!(
            !text.contains('/') && !text.contains('\''),
            "path echoed: {text}"
        );
    }
}

#[test]
fn identity_refuses_symlinks_and_bad_paths_without_echo() {
    let base = TempDir::new().unwrap();
    let home = TempDir::new().unwrap();
    let repo = base.path().join("repo");
    fs::create_dir(&repo).unwrap();
    committed_repo(&repo);
    #[cfg(unix)]
    let link = base.path().join("link");
    #[cfg(unix)]
    std::os::unix::fs::symlink(&repo, &link).unwrap();
    let missing = base.path().join("does-not-exist");
    let file = repo.join("a.txt");
    #[cfg(unix)]
    let bad_paths = [&link, &missing, &file];
    #[cfg(not(unix))]
    let bad_paths = [&missing, &file];
    for p in bad_paths {
        let (out, _) = ident(p, home.path(), None);
        assert!(!out.status.success(), "{p:?} must be refused");
        // A failure envelope on stdout (01-C #1990); neither stream echoes the path.
        let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(failure["ok"], json!(false));
        assert_eq!(failure["operation"], json!("identity"));
        let expected = if p == &missing {
            "not_found"
        } else {
            "invalid_input"
        };
        assert_eq!(failure["error"]["kind"], json!(expected), "{p:?}");
        for stream in [&out.stdout, &out.stderr] {
            let text = String::from_utf8_lossy(stream);
            assert!(
                !text.contains(base.path().to_str().unwrap()),
                "path echoed: {text}"
            );
        }
    }
    #[cfg(unix)]
    {
        // A symlinked project state root is refused like target() refuses it.
        let elsewhere = TempDir::new().unwrap();
        std::os::unix::fs::symlink(elsewhere.path(), repo.join(".moai-cli")).unwrap();
        let (out, _) = ident(&repo, home.path(), None);
        assert!(!out.status.success());
        assert!(String::from_utf8_lossy(&out.stderr).contains("symlink"));
        fs::remove_file(repo.join(".moai-cli")).unwrap();
        // So is a symlinked home state root.
        std::os::unix::fs::symlink(elsewhere.path(), home.path().join(".moai-cli")).unwrap();
        let (out, _) = ident(&repo, home.path(), None);
        assert!(!out.status.success());
    }
}

// ---- claude-doctor (Issue #1992, 02-B) --------------------------------------------------

/// A PATH directory holding a fake `claude` that prints `version_line` (or nothing on PATH
/// when `None`), so no real Claude Code runs.
fn fake_claude(version_line: Option<&str>) -> TempDir {
    let bin = TempDir::new().unwrap();
    if let Some(line) = version_line {
        let script = bin.path().join("claude");
        common::write_executable(&script, &format!("#!/bin/sh\necho '{line}'\n"));
    }
    bin
}
/// Runs `moai-cli claude-doctor <project>` with a fixture HOME and fake claude; returns the
/// envelope's `result` (Null on failure).
fn doctor(
    project: &std::path::Path,
    home: &std::path::Path,
    bin: &TempDir,
) -> (std::process::Output, Value) {
    let path = format!("{}:/usr/bin:/bin", bin.path().display());
    let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .arg("claude-doctor")
        .arg(project)
        .env("HOME", home)
        .env("PATH", path)
        .output()
        .unwrap();
    let v = serde_json::from_slice::<Value>(&out.stdout)
        .map(|v| v["result"].clone())
        .unwrap_or(Value::Null);
    (out, v)
}

/// A killed orphan can remain a zombie under a non-reaping container PID 1.
/// Only absence or a terminated state proves that the fixture is no longer running.
fn assert_fixture_process_stopped(pid: &str) {
    for _ in 0..40 {
        let out = Command::new("ps")
            .args(["-p", pid, "-o", "stat="])
            .output()
            .unwrap();
        let zombie = out.status.success()
            && String::from_utf8_lossy(&out.stdout)
                .trim_start()
                .starts_with('Z');
        let absent = out.status.code() == Some(1) && out.stdout.is_empty() && out.stderr.is_empty();
        if absent || zombie {
            return;
        }
        std::thread::sleep(std::time::Duration::from_millis(50));
    }
    panic!("fixture child {pid} is still running");
}

/// #2349 (e): a `claude` that never answers `--version` is bounded. The doctor returns
/// `unknown` with a timeout reason, and the process group ends with it — the hanging
/// script's own child is gone too.
#[test]
fn claude_doctor_bounds_a_hanging_claude() {
    let home = TempDir::new().unwrap();
    let bin = TempDir::new().unwrap();
    let pidfile = bin.path().join("child.pid");
    common::write_executable(
        &bin.path().join("claude"),
        &format!(
            "#!/bin/sh\nsleep 120 &\necho $! > '{}'\nwait\n",
            pidfile.display()
        ),
    );
    let project = TempDir::new().unwrap();
    write_file(project.path(), "AGENTS.md", b"# rules\n");
    let started = std::time::Instant::now();
    let (out, v) = doctor(project.path(), home.path(), &bin);
    let elapsed = started.elapsed();
    assert!(out.status.success(), "{:?}", out.stderr);
    assert!(
        // #2511: far below the 120s stub, far above the 5s probe (VERSION_TIMEOUT is pinned
        // by a unit test in src/claude.rs), so load cannot fail it.
        elapsed < std::time::Duration::from_secs(60),
        "doctor waited {elapsed:?} for a hanging claude"
    );
    assert_eq!(v["agents_md_status"], json!("unknown"));
    assert_eq!(v["claude_version"], json!(null));
    assert!(
        v["reason"].as_str().unwrap().contains("timed out"),
        "{}",
        v["reason"]
    );
    let pid = fs::read_to_string(&pidfile).unwrap().trim().to_string();
    assert_fixture_process_stopped(&pid);
}

#[test]
fn claude_doctor_waits_for_inherited_stdout_within_the_deadline() {
    let home = TempDir::new().unwrap();
    let bin = TempDir::new().unwrap();
    common::write_executable(
        &bin.path().join("claude"),
        "#!/bin/sh\nprintf '2.1.319\\n'\nsleep 1 &\nexit 0\n",
    );
    let project = TempDir::new().unwrap();
    write_file(project.path(), "AGENTS.md", b"# rules\n");
    let (out, v) = doctor(project.path(), home.path(), &bin);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["claude_version"], json!("2.1.319"));
    assert_eq!(v["agents_md_status"], json!("loaded"));
}

#[test]
fn claude_doctor_stops_inherited_stdout_after_the_launcher_exits() {
    let home = TempDir::new().unwrap();
    let bin = TempDir::new().unwrap();
    let pidfile = bin.path().join("child.pid");
    common::write_executable(
        &bin.path().join("claude"),
        &format!(
            "#!/bin/sh\nprintf '2.1.319\\n'\nsleep 120 &\necho $! > '{}'\nexit 0\n",
            pidfile.display()
        ),
    );
    let project = TempDir::new().unwrap();
    write_file(project.path(), "AGENTS.md", b"# rules\n");
    let started = std::time::Instant::now();
    let (out, v) = doctor(project.path(), home.path(), &bin);
    assert!(out.status.success(), "{:?}", out.stderr);
    // #2511: below the 120s grandchild, generous above the 5s probe.
    assert!(started.elapsed() < std::time::Duration::from_secs(60));
    assert_eq!(v["claude_version"], json!(null));
    assert_eq!(v["agents_md_status"], json!("unknown"));
    assert!(v["reason"].as_str().unwrap().contains("timed out"));
    let pid = fs::read_to_string(pidfile).unwrap();
    assert_fixture_process_stopped(pid.trim());
}

#[cfg(unix)]
#[test]
fn claude_doctor_cleans_groups_without_trusting_path_kill() {
    for shadow_kill in [false, true] {
        let home = TempDir::new().unwrap();
        let bin = TempDir::new().unwrap();
        let pidfile = bin.path().join("child.pid");
        common::write_executable(
            &bin.path().join("claude"),
            &format!(
                "#!/bin/sh\n/bin/sleep 120 &\necho $! > '{}'\nwait\n",
                pidfile.display()
            ),
        );
        if shadow_kill {
            common::write_executable(&bin.path().join("kill"), "#!/bin/sh\nexit 1\n");
        }
        let project = TempDir::new().unwrap();
        write_file(project.path(), "AGENTS.md", b"# rules\n");
        let started = std::time::Instant::now();
        let out = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
            .args(["claude-doctor", project.path().to_str().unwrap()])
            .env("HOME", home.path())
            .env("PATH", bin.path())
            .output()
            .unwrap();
        let elapsed = started.elapsed();
        assert!(out.status.success(), "{:?}", out.stderr);
        let envelope: Value = serde_json::from_slice(&out.stdout).unwrap();
        let result = &envelope["result"];
        assert_eq!(result["agents_md_status"], json!("unknown"));
        assert_eq!(result["claude_version"], json!(null));
        assert!(
            // #2511: the five-second value itself is pinned by
            // `claude::tests::version_probe_deadline_is_five_seconds`; here only "it did not
            // wait for the 120s grandchild", with room for machine load.
            elapsed < std::time::Duration::from_secs(60),
            "the documented five-second probe took {elapsed:?}"
        );
        let pid = fs::read_to_string(pidfile).unwrap();
        assert_fixture_process_stopped(pid.trim());
    }
}

/// HA-03-CLAUDE: the four AGENTS/CLAUDE combinations, judged without writing anything.
#[test]
fn claude_doctor_agents_loading_combinations() {
    let home = TempDir::new().unwrap();
    let bin = fake_claude(Some("2.1.319 (Claude Code)"));
    let case = |files: &[(&str, &str)]| {
        let p = TempDir::new().unwrap();
        write_file(p.path(), "AGENTS.md", b"# rules\n");
        for (rel, body) in files {
            write_file(p.path(), rel, body.as_bytes());
        }
        let before = tree_state(p.path());
        let (out, v) = doctor(p.path(), home.path(), &bin);
        assert!(out.status.success(), "{:?}", out.stderr);
        assert_eq!(before, tree_state(p.path()), "doctor wrote to the project");
        assert_eq!(v["writes"], json!(false));
        assert_eq!(v["host_loaded"], json!("not_run"));
        assert_eq!(v["claude_version"], json!("2.1.319"));
        (p, v)
    };
    let (_p1, agents_only) = case(&[]);
    assert_eq!(agents_only["agents_md_status"], json!("loaded"));
    assert_eq!(agents_only["load_path"], json!("direct"));
    assert_eq!(agents_only["instruction_files_source"], json!("default"));
    let (_p2, claude_md) = case(&[("CLAUDE.md", "# mine\n")]);
    assert_eq!(
        claude_md["agents_md_status"],
        json!("suppressed_by_claude_md")
    );
    assert_eq!(claude_md["suggestions"].as_array().unwrap().len(), 2);
    let (_p3, local) = case(&[("CLAUDE.local.md", "# private\n")]);
    assert_eq!(local["agents_md_status"], json!("suppressed_by_claude_md"));
    assert!(local["suppressing_files"][0]
        .as_str()
        .unwrap()
        .ends_with("/CLAUDE.local.md"));
    let (_p4, import) = case(&[("CLAUDE.md", "# mine\n@AGENTS.md\n")]);
    assert_eq!(import["agents_md_status"], json!("loaded"));
    assert_eq!(import["load_path"], json!("import"));
    // An import inside a code fence or a code span is not an import.
    let (_p5, quoted) = case(&[("CLAUDE.md", "```\n@AGENTS.md\n```\nsee `@AGENTS.md`\n")]);
    assert_eq!(quoted["agents_md_status"], json!("suppressed_by_claude_md"));
    // Also when the span holds more than the import, or never closes (Codex review #2044).
    for body in ["`See @AGENTS.md for rules`\n", "see `x @AGENTS.md\n"] {
        let (_p, v) = case(&[("CLAUDE.md", body)]);
        assert_eq!(
            v["agents_md_status"],
            json!("suppressed_by_claude_md"),
            "{body}"
        );
    }
    // A suppressing file above the start directory counts too.
    let (p6, _) = case(&[(".claude/CLAUDE.md", "# team\n")]);
    fs::create_dir(p6.path().join("sub")).unwrap();
    let (_, sub) = doctor(&p6.path().join("sub"), home.path(), &bin);
    assert_eq!(sub["agents_md_status"], json!("suppressed_by_claude_md"));
}

/// A project under HOME: the user's own ~/.claude/CLAUDE.md does not suppress AGENTS.md,
/// while a CLAUDE.md in HOME itself is an ordinary directory above and does (agent-2 #2044).
#[test]
fn claude_doctor_ignores_the_user_claude_md() {
    let home = TempDir::new().unwrap();
    let bin = fake_claude(Some("2.1.319 (Claude Code)"));
    let project = home.path().join("work/repo");
    write_file(&project, "AGENTS.md", b"# rules\n");
    write_file(home.path(), ".claude/CLAUDE.md", b"# my global notes\n");
    let (out, v) = doctor(&project, home.path(), &bin);
    assert!(out.status.success(), "{:?}", out.stderr);
    assert_eq!(v["agents_md_status"], json!("loaded"));
    assert_eq!(v["suppressing_files"], json!([]));
    write_file(home.path(), "CLAUDE.md", b"# home dir file\n");
    let (_, v) = doctor(&project, home.path(), &bin);
    assert_eq!(v["agents_md_status"], json!("suppressed_by_claude_md"));
}

#[test]
fn claude_doctor_versions_and_user_settings() {
    let home = TempDir::new().unwrap();
    let p = TempDir::new().unwrap();
    write_file(p.path(), "AGENTS.md", b"# rules\n");
    write_file(p.path(), "CLAUDE.md", b"# mine\n");
    for (line, status) in [
        (None, "unknown"),
        (Some("not a version"), "unknown"),
        (Some("2.1.276 (Claude Code)"), "unsupported_version"),
        (Some("2.1.277 (Claude Code)"), "suppressed_by_claude_md"),
    ] {
        let (out, v) = doctor(p.path(), home.path(), &fake_claude(line));
        assert!(out.status.success(), "{:?}", out.stderr);
        assert_eq!(v["agents_md_status"], json!(status), "{line:?}");
    }
    let bin = fake_claude(Some("2.1.319 (Claude Code)"));
    let settings = |body: &str| write_file(home.path(), ".claude/settings.json", body.as_bytes());
    settings(
        r#"{"pluginConfigs":{"cc-plugin-agents-md@builtin":{"options":{"instructionFiles":"claude-md-and-agents-md"}}}}"#,
    );
    let (_, v) = doctor(p.path(), home.path(), &bin);
    assert_eq!(v["agents_md_status"], json!("loaded"));
    assert_eq!(v["instruction_files_source"], json!("user_settings"));
    // Before 2.1.285 the plugin id was agents-md@builtin; the new id is not read there.
    let (_, v) = doctor(p.path(), home.path(), &fake_claude(Some("2.1.280")));
    assert_eq!(v["agents_md_status"], json!("suppressed_by_claude_md"));
    settings(
        r#"{"pluginConfigs":{"agents-md@builtin":{"options":{"instructionFiles":"claude-md-and-agents-md"}}}}"#,
    );
    let (_, v) = doctor(p.path(), home.path(), &fake_claude(Some("2.1.280")));
    assert_eq!(v["agents_md_status"], json!("loaded"));
    // From 2.1.285 an entry under the older id is still read.
    let (_, v) = doctor(p.path(), home.path(), &bin);
    assert_eq!(v["agents_md_status"], json!("loaded"));
    // Project settings are ignored for this entry.
    fs::remove_file(home.path().join(".claude/settings.json")).unwrap();
    write_file(
        p.path(),
        ".claude/settings.json",
        br#"{"pluginConfigs":{"cc-plugin-agents-md@builtin":{"options":{"instructionFiles":"claude-md-and-agents-md"}}}}"#,
    );
    let (_, v) = doctor(p.path(), home.path(), &bin);
    assert_eq!(v["agents_md_status"], json!("suppressed_by_claude_md"));
    // Unreadable user settings and an excluding value are reported, never guessed.
    settings("{not json");
    let (out, v) = doctor(p.path(), home.path(), &bin);
    assert!(out.status.success());
    assert_eq!(v["agents_md_status"], json!("unknown"));
    settings(
        r#"{"pluginConfigs":{"cc-plugin-agents-md@builtin":{"options":{"instructionFiles":"claude-md"}}}}"#,
    );
    let (_, v) = doctor(p.path(), home.path(), &bin);
    assert_eq!(v["agents_md_status"], json!("unknown"));
    // No AGENTS.md at all is unknown, not loaded.
    let bare = TempDir::new().unwrap();
    let (_, v) = doctor(bare.path(), home.path(), &bin);
    assert_eq!(v["agents_md_status"], json!("unknown"));
    // Bad paths are refused through the envelope.
    let (out, _) = doctor(&p.path().join("missing"), home.path(), &bin);
    assert_eq!(out.status.code(), Some(1));
    let env: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(env["operation"], json!("claude.doctor"));
    assert_eq!(env["error"]["kind"], json!("not_found"));
}

// ---- Project-specific generation (03-C #1997) --------------------------------------------

fn generation_project() -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let profile = json!({"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"],"interview_revision":"rc1-fixture-1"});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.path().join(".moai-cli/project.yaml"), &bytes).unwrap();
    root
}

fn generation_request(local_file: &str) -> Value {
    json!({
        "protocol_version": 1,
        "language": "en",
        "workflow_reference": "skills/moai/workflows",
        "local_file": local_file,
        "workflow_id": "review-flow",
        "native_roles": [],
        "skill_projections": []
    })
}

fn generation_plan_cli(root: &TempDir, request: &Value) -> std::process::Output {
    cli(&["generation-plan", root.path().to_str().unwrap()], request)
}

fn generation_apply_cli(root: &TempDir, plan: &Value) -> std::process::Output {
    cli(
        &[
            "generation-apply",
            root.path().to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan,
    )
}

fn actions_of(plan: &Value) -> Vec<(String, String)> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| {
            (
                f["path"].as_str().unwrap().to_string(),
                f["action"].as_str().unwrap().to_string(),
            )
        })
        .collect()
}

#[test]
fn generation_default_has_no_native_roles() {
    // 03C-AC-1: the default request generates the AGENTS.md managed section, the
    // portable workflow and the .gitignore section; nothing under the native role
    // folders exists or is planned (SETUP-03).
    let root = generation_project();
    let request = generation_request("ignored");
    let plan = ok_result(&generation_plan_cli(&root, &request));
    let paths: Vec<&str> = plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["path"].as_str().unwrap())
        .collect();
    assert!(paths.iter().any(|p| p.starts_with("AGENTS.md")));
    assert!(paths
        .iter()
        .any(|p| p.starts_with(".moai-cli/workflows/review-flow/")));
    assert!(paths.contains(&".gitignore"));
    assert!(!paths.iter().any(|p| p.starts_with(".claude/agents/")));
    assert!(!paths.iter().any(|p| p.starts_with(".codex/agents/")));
    // The plan wrote nothing.
    assert!(!root.path().join("AGENTS.md").exists());
    assert!(!root.path().join(".moai-cli/workflows").exists());
    let applied = ok_result(&generation_apply_cli(&root, &plan));
    assert_eq!(applied["host_loaded"], json!(false));
    assert_eq!(applied["execution_verified"], json!(false));
    assert!(root.path().join("AGENTS.md").exists());
    assert!(root
        .path()
        .join(".moai-cli/workflows/review-flow/SKILL.md")
        .exists());
    assert!(root
        .path()
        .join(".moai-cli/workflows/review-flow/references/workflow.md")
        .exists());
    assert!(root.path().join(".gitignore").exists());
    assert!(!root.path().join(".claude/agents").exists());
    assert!(!root.path().join(".codex/agents").exists());
    // The manifest records the generation baselines and the profile linkage.
    let manifest: Value = serde_json::from_slice(
        &fs::read(root.path().join(".moai-cli/manifests/generation.json")).unwrap(),
    )
    .unwrap();
    assert_eq!(manifest["workflow_id"], json!("review-flow"));
    assert_eq!(manifest["interview_revision"], json!("rc1-fixture-1"));
}

#[test]
fn generation_rerun_is_diff_zero() {
    // 03C-AC-2: the second plan is unchanged for every managed file, and the target
    // files plus generation.json hash identically after both applies (SETUP-05).
    let root = generation_project();
    let request = generation_request("ignored");
    let plan = ok_result(&generation_plan_cli(&root, &request));
    ok_result(&generation_apply_cli(&root, &plan));
    let sha = |p: &std::path::Path| moai_cli::digest(&fs::read(p).unwrap());
    let agents_before = sha(&root.path().join("AGENTS.md"));
    let workflow_before = sha(&root.path().join(".moai-cli/workflows/review-flow/SKILL.md"));
    let manifest_before = sha(&root.path().join(".moai-cli/manifests/generation.json"));
    let plan2 = ok_result(&generation_plan_cli(&root, &request));
    for (path, action) in actions_of(&plan2) {
        if path == ".moai-cli/manifests/generation.json" {
            // agents.rs reports the manifest row as merge by convention: its apply
            // rewrites the manifest bytes, and the generation linkage below restores
            // them — the equality assertions here carry the requirement.
            continue;
        }
        assert_eq!(action, "unchanged", "{path}");
    }
    ok_result(&generation_apply_cli(&root, &plan2));
    assert_eq!(sha(&root.path().join("AGENTS.md")), agents_before);
    assert_eq!(
        sha(&root.path().join(".moai-cli/workflows/review-flow/SKILL.md")),
        workflow_before
    );
    assert_eq!(
        sha(&root.path().join(".moai-cli/manifests/generation.json")),
        manifest_before
    );
}

#[test]
fn generation_preserves_user_files_and_reports_conflict() {
    // 03C-AC-3: user bytes outside the managed section (CRLF, no final newline),
    // an unrelated native role file and user .gitignore rules survive byte for
    // byte; a user edit inside the section plus a template change reports
    // conflict and writes nothing (SETUP-04).
    let root = generation_project();
    fs::write(
        root.path().join("AGENTS.md"),
        b"# user header\r\n\n# user footer without final newline",
    )
    .unwrap();
    fs::create_dir_all(root.path().join(".claude/agents")).unwrap();
    fs::write(
        root.path().join(".claude/agents/other.md"),
        b"user's own role",
    )
    .unwrap();
    fs::write(root.path().join(".gitignore"), b"target/\n*.tmp\n").unwrap();
    let request = generation_request("ignored");
    let plan = ok_result(&generation_plan_cli(&root, &request));
    ok_result(&generation_apply_cli(&root, &plan));
    assert!(
        String::from_utf8_lossy(&fs::read(root.path().join("AGENTS.md")).unwrap())
            .starts_with("# user header\r\n")
    );
    assert_eq!(
        fs::read(root.path().join(".claude/agents/other.md")).unwrap(),
        b"user's own role"
    );
    let gitignore_bytes = fs::read(root.path().join(".gitignore")).unwrap();
    let gitignore = String::from_utf8_lossy(&gitignore_bytes);
    assert!(gitignore.contains("*.tmp"));
    assert!(gitignore.contains("moai:managed"));
    // The user edits inside the managed section; a template change (language) then
    // reports a conflict and nothing is written.
    let agents = fs::read_to_string(root.path().join("AGENTS.md")).unwrap();
    let edited = agents.replace(
        "Confirm the approved SPEC and Issue scope",
        "The user rewrote this managed line",
    );
    assert_ne!(
        agents, edited,
        "the fixture string must exist in the section"
    );
    fs::write(root.path().join("AGENTS.md"), edited).unwrap();
    let before = fs::read(root.path().join("AGENTS.md")).unwrap();
    let mut changed = generation_request("ignored");
    changed["language"] = json!("ko");
    let plan2 = ok_result(&generation_plan_cli(&root, &changed));
    assert!(actions_of(&plan2)
        .iter()
        .any(|(p, a)| p == "AGENTS.md" && a == "conflict"));
    let out = generation_apply_cli(&root, &plan2);
    assert!(!out.status.success());
    let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(failure["error"]["kind"], json!("conflict"));
    assert_eq!(fs::read(root.path().join("AGENTS.md")).unwrap(), before);
}

#[test]
fn generation_native_roles_render_official_shapes() {
    // 03C-AC-4: Claude roles render name/description/tools frontmatter, Codex roles
    // only name/description/developer_instructions (never sandbox_mode), and the
    // read-only intent surfaces as a plan warning.
    let root = generation_project();
    let mut request = generation_request("none");
    request["native_roles"] = json!([
        {"role_id": "prj-fixer", "host": "claude", "purpose": "Fix lint and test failures",
         "tools": ["Bash", "Edit"]},
        {"role_id": "prj-watcher", "host": "codex", "purpose": "Review diffs only",
         "read_only_intent": true}
    ]);
    let plan = ok_result(&generation_plan_cli(&root, &request));
    assert!(plan
        .to_string()
        .as_str()
        .contains("cannot create a sandbox"));
    ok_result(&generation_apply_cli(&root, &plan));
    let claude_role = String::from_utf8_lossy(
        &fs::read(root.path().join(".claude/agents/prj-fixer.md")).unwrap(),
    )
    .into_owned();
    assert!(claude_role.contains("name: prj-fixer"));
    assert!(claude_role.contains("description: Fix lint and test failures"));
    assert!(claude_role.contains("tools: Bash, Edit"));
    let codex_role_bytes = fs::read(root.path().join(".codex/agents/prj-watcher.toml")).unwrap();
    let codex_role = String::from_utf8_lossy(&codex_role_bytes).into_owned();
    assert!(codex_role.contains("name = \"prj-watcher\""));
    assert!(codex_role.contains("developer_instructions = \"Review diffs only\""));
    assert!(!codex_role.contains("sandbox_mode"));
    assert!(!codex_role.contains("approval_policy"));
    assert!(!codex_role.contains("mcp_servers"));
}

/// #2639 (project agents): a project role carrying the confirmed test command is written for
/// both hosts, a re-run plans it `unchanged`, a hand edit is kept (`unchanged` while the
/// answers stay, `conflict` once they change too) and never overwritten, and a role named with the base agents' `moai-` prefix is refused.
#[test]
fn generation_project_agent_keeps_hand_edits_and_never_takes_a_base_name() {
    let root = generation_project();
    let mut request = generation_request("none");
    let purpose = "Run `cargo test --locked` and report the first failing test";
    request["native_roles"] = json!([
        {"role_id": "prj-demo-tester", "host": "claude", "purpose": purpose},
        {"role_id": "prj-demo-tester", "host": "codex", "purpose": purpose}
    ]);
    let plan = ok_result(&generation_plan_cli(&root, &request));
    ok_result(&generation_apply_cli(&root, &plan));
    let claude = root.path().join(".claude/agents/prj-demo-tester.md");
    let codex = root.path().join(".codex/agents/prj-demo-tester.toml");
    assert!(fs::read_to_string(&claude).unwrap().contains(purpose));
    assert!(fs::read_to_string(&codex)
        .unwrap()
        .contains("Run `cargo test --locked`"));

    let again = ok_result(&generation_plan_cli(&root, &request));
    for (path, action) in actions_of(&again) {
        if path.ends_with("prj-demo-tester.md") || path.ends_with("prj-demo-tester.toml") {
            assert_eq!(action, "unchanged", "{path}");
        }
    }

    let edited = "---\nname: prj-demo-tester\ndescription: mine\n---\n\nmine\n";
    fs::write(&claude, edited).unwrap();
    // Only the user changed it: the plan keeps the edit and applying writes nothing over it.
    let kept = ok_result(&generation_plan_cli(&root, &request));
    assert!(actions_of(&kept)
        .iter()
        .any(|(p, a)| p == ".claude/agents/prj-demo-tester.md" && a == "unchanged"));
    ok_result(&generation_apply_cli(&root, &kept));
    assert_eq!(fs::read_to_string(&claude).unwrap(), edited);
    // The answers changed too: both sides moved, so it is a conflict and nothing is written.
    request["native_roles"][0]["purpose"] = json!("Run `cargo test` and report every failure");
    let after_edit = ok_result(&generation_plan_cli(&root, &request));
    assert!(actions_of(&after_edit)
        .iter()
        .any(|(p, a)| p == ".claude/agents/prj-demo-tester.md" && a == "conflict"));
    let out = generation_apply_cli(&root, &after_edit);
    assert!(!out.status.success());
    let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(failure["error"]["kind"], json!("conflict"));
    assert_eq!(fs::read_to_string(&claude).unwrap(), edited);

    let fresh = generation_project();
    let mut base = generation_request("none");
    base["native_roles"] =
        json!([{"role_id": "moai-impl", "host": "claude", "purpose": "Shadow a base agent"}]);
    assert!(!generation_plan_cli(&fresh, &base).status.success());
    assert!(!fresh.path().join(".claude/agents").exists());
    // #2706: only `prj-<name>` is a project role name; the bare prefix and an unprefixed
    // name are refused, the shortest prefixed name is accepted.
    for (id, ok) in [
        ("prj-x", true),
        ("prj-", false),
        ("tester", false),
        ("xprj-x", false),
    ] {
        base["native_roles"] =
            json!([{"role_id": id, "host": "claude", "purpose": "A project role"}]);
        assert_eq!(
            generation_plan_cli(&fresh, &base).status.success(),
            ok,
            "{id}"
        );
    }
}

#[test]
fn generation_license_gate_and_attribution() {
    // 03C-AC-5: a projection citing an unreviewed source is not planned and warns;
    // a reviewed source renders an attribution line (no content copying).
    let root = generation_project();
    let sources = json!({
        "schema_version": 1,
        "sources": [
            {"source_url": "fixtures/skill-a", "license": {"status": "unknown"},
             "resolved_revision": null},
            {"source_url": "fixtures/skill-b", "license": {"status": "reviewed", "spdx": "MIT"},
             "resolved_revision": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
        ]
    });
    fs::create_dir_all(root.path().join(".moai-cli/manifests")).unwrap();
    fs::write(
        root.path().join(".moai-cli/manifests/sources.json"),
        serde_json::to_vec_pretty(&sources).unwrap(),
    )
    .unwrap();
    let mut request = generation_request("none");
    request["skill_projections"] = json!([
        {"id": "ref-a", "source_url": "fixtures/skill-a"},
        {"id": "ref-b", "source_url": "fixtures/skill-b",
         "resolved_revision": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
    ]);
    let plan = ok_result(&generation_plan_cli(&root, &request));
    let planned: Vec<&str> = plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["path"].as_str().unwrap().contains("ref-"))
        .map(|f| f["path"].as_str().unwrap())
        .collect();
    assert!(!planned.iter().any(|p| p.contains("ref-a")), "{planned:?}");
    assert!(planned.iter().any(|p| p.contains("ref-b")));
    assert!(plan.to_string().as_str().contains("not license-reviewed"));
    ok_result(&generation_apply_cli(&root, &plan));
    let projection_bytes = fs::read(root.path().join(".claude/skills/ref-b/SKILL.md")).unwrap();
    let projection = String::from_utf8_lossy(&projection_bytes).into_owned();
    assert!(projection.contains("source: fixtures/skill-b"));
    assert!(projection.contains("license: MIT"));
    assert!(projection.contains("name: ref-b"));
}

#[test]
fn generation_partial_apply_recovers() {
    // 03C-AC-6: after a forced interruption (the manifest loses its generation
    // baselines), re-applying absorbs the already-rendered files as baselines and
    // reaches the same final state as a clean apply.
    let reference = generation_project();
    let request = generation_request("ignored");
    let plan = ok_result(&generation_plan_cli(&reference, &request));
    ok_result(&generation_apply_cli(&reference, &plan));

    let root = generation_project();
    let plan = ok_result(&generation_plan_cli(&root, &request));
    ok_result(&generation_apply_cli(&root, &plan));
    // The interruption: the manifest forgets the generation baselines.
    let manifest_path = root.path().join(".moai-cli/manifests/generation.json");
    let mut value: Value = serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
    if let Some(list) = value["owned_files"].as_array_mut() {
        list.retain(|f| {
            !f["section_id"]
                .as_str()
                .unwrap_or("")
                .starts_with("generation:")
        });
    }
    fs::write(&manifest_path, serde_json::to_vec_pretty(&value).unwrap()).unwrap();
    // Re-apply: already-rendered files are absorbed as baselines, nothing is
    // re-published, and the manifest returns to the clean-apply state.
    let plan2 = ok_result(&generation_plan_cli(&root, &request));
    ok_result(&generation_apply_cli(&root, &plan2));
    for rel in [
        "AGENTS.md",
        ".moai-cli/workflows/review-flow/SKILL.md",
        ".moai-cli/workflows/review-flow/references/workflow.md",
        ".gitignore",
    ] {
        assert_eq!(
            fs::read(root.path().join(rel)).unwrap(),
            fs::read(reference.path().join(rel)).unwrap(),
            "{rel}"
        );
    }
    assert_eq!(
        fs::read(&manifest_path).unwrap(),
        fs::read(reference.path().join(".moai-cli/manifests/generation.json")).unwrap()
    );
}

#[test]
fn generation_rejects_frontmatter_delimiter_injection() {
    // REQ-03C-07: managed-marker tokens and frontmatter delimiter lines in any
    // rendered string are refused without echoing the value.
    for purpose in [
        "x <!-- moai:managed:end id=y --> z",
        "line one\n---\nline two",
    ] {
        let root = generation_project();
        let mut request = generation_request("none");
        request["native_roles"] = json!([
            {"role_id": "prj-injected", "host": "claude", "purpose": purpose}
        ]);
        let out = generation_plan_cli(&root, &request);
        assert!(!out.status.success(), "{purpose}");
        let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(failure["error"]["kind"], json!("invalid_input"));
        assert!(!String::from_utf8_lossy(&out.stdout).contains(purpose));
        assert!(!String::from_utf8_lossy(&out.stderr).contains(purpose));
    }
}

#[test]
fn generation_skill_projection_name_matches_directory() {
    // REQ-03C-09: projection ids follow the Agent Skills name rules and the rendered
    // SKILL.md name equals the directory name; an existing unclaimed directory is a
    // conflict.
    let root = generation_project();
    let mut request = generation_request("none");
    request["skill_projections"] = json!([
        {"id": "Bad_Name", "source_url": "fixtures/skill-b"}
    ]);
    let out = generation_plan_cli(&root, &request);
    assert!(!out.status.success());
    let failure: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(failure["error"]["kind"], json!("invalid_input"));
    // A valid projection renders name == directory.
    let sources = json!({
        "schema_version": 1,
        "sources": [{"source_url": "fixtures/skill-b",
                     "license": {"status": "reviewed", "spdx": "MIT"},
                     "resolved_revision": null}]
    });
    fs::create_dir_all(root.path().join(".moai-cli/manifests")).unwrap();
    fs::write(
        root.path().join(".moai-cli/manifests/sources.json"),
        serde_json::to_vec_pretty(&sources).unwrap(),
    )
    .unwrap();
    let mut request = generation_request("none");
    request["skill_projections"] = json!([
        {"id": "ref-b", "source_url": "fixtures/skill-b"}
    ]);
    let plan = ok_result(&generation_plan_cli(&root, &request));
    ok_result(&generation_apply_cli(&root, &plan));
    let projection_bytes = fs::read(root.path().join(".claude/skills/ref-b/SKILL.md")).unwrap();
    let projection = String::from_utf8_lossy(&projection_bytes).into_owned();
    assert!(projection.contains("name: ref-b"));
    // The same directory, never projected before, is a conflict for a fresh manifest.
    let root2 = generation_project();
    fs::create_dir_all(root2.path().join(".moai-cli/manifests")).unwrap();
    fs::write(
        root2.path().join(".moai-cli/manifests/sources.json"),
        serde_json::to_vec_pretty(&sources).unwrap(),
    )
    .unwrap();
    fs::create_dir_all(root2.path().join(".claude/skills/ref-b")).unwrap();
    fs::write(
        root2.path().join(".claude/skills/ref-b/SKILL.md"),
        b"user's own",
    )
    .unwrap();
    let plan = ok_result(&generation_plan_cli(&root2, &request));
    assert!(actions_of(&plan)
        .iter()
        .any(|(p, a)| p.contains("ref-b") && a == "conflict"));
    assert_eq!(
        fs::read(root2.path().join(".claude/skills/ref-b/SKILL.md")).unwrap(),
        b"user's own"
    );
}

#[test]
fn generation_manifest_entries_carry_kind() {
    // #2109 AC-1: newly written generation.json entries carry the additive `kind`
    // field, and the kind alone discriminates the entry type — without consulting
    // the `generation:…` section_id tag kept for readers of older manifests.
    let root = generation_project();
    let mut request = generation_request("none");
    request["native_roles"] = json!([
        {"role_id": "prj-fixer", "host": "claude", "purpose": "Fix lint and test failures"}
    ]);
    let plan = ok_result(&generation_plan_cli(&root, &request));
    ok_result(&generation_apply_cli(&root, &plan));
    let manifest_bytes = fs::read(root.path().join(".moai-cli/manifests/generation.json")).unwrap();
    let manifest: Value = serde_json::from_slice(&manifest_bytes).unwrap();
    let by_kind = |kind: &str| -> Vec<&Value> {
        manifest["owned_files"]
            .as_array()
            .unwrap()
            .iter()
            .filter(|f| f["kind"] == json!(kind))
            .collect()
    };
    let workflows = by_kind("workflow");
    assert_eq!(workflows.len(), 2, "{manifest}");
    assert!(workflows.iter().all(|f| f["path"]
        .as_str()
        .unwrap()
        .starts_with(".moai-cli/workflows/review-flow/")));
    assert_eq!(by_kind("native-role").len(), 1);
    assert_eq!(
        by_kind("native-role")[0]["path"],
        json!(".claude/agents/prj-fixer.md")
    );
    // The agents-owned entries stay kind-less: their section_id already carries the
    // managed section they belong to.
    let agents_entries = by_kind(""); // no direct filter — count kind-less below
    let _ = agents_entries;
    let kindless: Vec<&Value> = manifest["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f.get("kind").is_none())
        .collect();
    assert!(
        kindless
            .iter()
            .all(|f| f["section_id"] == json!("moai-workflow")),
        "{manifest}"
    );
    // Discriminating by kind alone (no section_id) partitions every generation
    // entry exactly as the tag does.
    let tag_partition: Vec<String> = manifest["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(|f| {
            f["section_id"]
                .as_str()
                .unwrap_or("")
                .strip_prefix("generation:")
                .and_then(|rest| rest.split(':').next())
                .map(str::to_string)
        })
        .collect();
    let kind_partition: Vec<String> = manifest["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(|f| f["kind"].as_str().map(str::to_string))
        .collect();
    assert_eq!(tag_partition, kind_partition);
}
