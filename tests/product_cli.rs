use serde_json::{json, Value};
use std::{
    fs,
    io::Write,
    process::{Command, Stdio},
};
use tempfile::TempDir;
fn profile(pin: Option<String>) -> Value {
    let mut v = json!({"schema_version":1,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude","codex"]});
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
    child
        .stdin
        .take()
        .unwrap()
        .write_all(input.to_string().as_bytes())
        .unwrap();
    child.wait_with_output().unwrap()
}
fn create(root: &TempDir, input: &Value) -> Value {
    let p = root.path().to_str().unwrap();
    let preview = cli(&["project-plan", p], input);
    assert!(preview.status.success(), "{:?}", preview.stderr);
    let plan: Value = serde_json::from_slice(&preview.stdout).unwrap();
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
    let bytes = serde_json::to_vec(&value).unwrap();
    fs::write(&file, &bytes).unwrap();
    create(&root, &input);
    assert_eq!(fs::read(&file).unwrap(), bytes);
    let mut changed = input;
    changed["purpose"] = json!("Different purpose");
    assert!(
        !cli(&["project-plan", root.path().to_str().unwrap()], &changed)
            .status
            .success()
    );
    assert_eq!(fs::read(&file).unwrap(), bytes);
}
#[test]
fn stale_and_tampered_plans_do_not_overwrite() {
    let root = TempDir::new().unwrap();
    let p = root.path().to_str().unwrap();
    let out = cli(&["project-plan", p], &profile(None));
    let mut plan: Value = serde_json::from_slice(&out.stdout).unwrap();
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
    let value: Value = serde_json::from_slice(&result.stdout).unwrap();
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
    assert_eq!(
        serde_json::from_slice::<Value>(&result.stdout).unwrap()["source"],
        "user"
    );
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
        assert!(out.stdout.is_empty());
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
    let plan: Value = serde_json::from_slice(&result.stdout).unwrap();
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
    let plan: Value = serde_json::from_slice(&out.stdout).unwrap();
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
    let plan: Value = serde_json::from_slice(&out.stdout).unwrap();
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
    let plan: Value = serde_json::from_slice(&out.stdout).unwrap();
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
    let value: Value = serde_json::from_slice(&result.stdout).unwrap();
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
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("en", "none")).stdout).unwrap();
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
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("en", "none")).stdout).unwrap();
    assert_eq!(actions(&plan)[0].1, "unchanged");

    // Template changes (language) after a user edit inside the block: conflict, nothing applied.
    let out = agents_plan(&root, &agents_request("ko", "none"));
    assert!(out.status.success());
    let plan: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(actions(&plan)[0].1, "conflict");
    let result = agents_apply(&root, &plan);
    assert!(!result.status.success());
    assert!(String::from_utf8_lossy(&result.stderr).contains("conflict"));
    assert_eq!(fs::read_to_string(&file).unwrap(), edited);

    // Unedited block + template change: merge replaces only the block.
    fs::write(&file, &outside).unwrap();
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("ko", "none")).stdout).unwrap();
    assert_eq!(actions(&plan)[0].1, "merge");
    assert!(agents_apply(&root, &plan).status.success());
    let merged = fs::read_to_string(&file).unwrap();
    assert!(merged.contains("MoAI 작업 흐름"));
    assert!(merged.ends_with("\n## User appendix\nkeep me\n"));
    assert!(merged.starts_with("# demo\n\n"));

    // User removes the managed block while the template is unchanged: the deletion is
    // respected (never re-appended silently). Once the template changes, it is a conflict.
    fs::write(&file, "# demo\n\nno block\n").unwrap();
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("ko", "none")).stdout).unwrap();
    assert_eq!(actions(&plan)[0].1, "unchanged");
    assert!(agents_apply(&root, &plan).status.success());
    assert_eq!(fs::read_to_string(&file).unwrap(), "# demo\n\nno block\n");
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("en", "none")).stdout).unwrap();
    assert_eq!(actions(&plan)[0].1, "conflict");

    // Malformed markers: conflict.
    fs::write(
        &file,
        "<!-- moai:managed:begin id=moai-workflow -->\nno end\n",
    )
    .unwrap();
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("ko", "none")).stdout).unwrap();
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
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("ko", "tracked")).stdout)
            .unwrap();
    assert!(agents_apply(&root, &plan).status.success());
    assert_eq!(
        fs::read(root.path().join("AGENTS.local.md")).unwrap(),
        b"personal\n"
    );
    // Deleted by the user: not recreated.
    fs::remove_file(root.path().join("AGENTS.local.md")).unwrap();
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("ko", "tracked")).stdout)
            .unwrap();
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
    let plan: Value = serde_json::from_slice(&out.stdout).unwrap();
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
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("en", "tracked")).stdout)
            .unwrap();
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
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("en", "tracked")).stdout)
            .unwrap();
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
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("ko", "none")).stdout).unwrap();
    assert_eq!(actions(&plan)[0].1, "merge");
    assert!(agents_apply(&root, &plan).status.success());
    // Simulate the crash between the AGENTS.md write and the manifest write.
    fs::write(&manifest, &v1_manifest).unwrap();
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("ko", "none")).stdout).unwrap();
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
    let plan: Value = serde_json::from_slice(&agents_plan(&root, &request).stdout).unwrap();
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
    let plan: Value =
        serde_json::from_slice(&agents_plan(&root, &agents_request("en", "tracked")).stdout)
            .unwrap();
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
