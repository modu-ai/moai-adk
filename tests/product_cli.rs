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

/// Runs the launcher syntax (00-F #2028). Tests use `--dry-run` only and never start a host.
fn launcher(args: &[&str], token: Option<&str>) -> std::process::Output {
    let mut c = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    c.args(args).env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL");
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
        "--advisor",
        "fable",
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
    // `-a` without a host is the Claude Code agent with a fixed default name.
    let v = launched(&launcher(&["-a", "--dry-run"], None));
    assert_eq!(v["role"], json!("agent"));
    assert_eq!(v["argv"][2], json!("agent-1"));
    assert!(v["notes"].to_string().contains("HOLD"));
    // Ambiguous or empty invocations fail instead of guessing.
    assert!(!launcher(&["-l", "-a", "--dry-run"], None).status.success());
    assert!(!launcher(&["-l", "cc", "--dry-run"], None).status.success());
    assert!(!launcher(&["cc", "--dry-run"], None).status.success());
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
    let argv = v["argv"].to_string();
    assert!(argv.contains(r#""--model","sonnet","--effort","high","--advisor","fable""#));
    assert!(!launcher(&["-l", "--effort", "turbo", "--dry-run"], None)
        .status
        .success());
    assert!(!launcher(&["-l", "--name", "two words", "--dry-run"], None)
        .status
        .success());
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
            "ANTHROPIC_AUTH_TOKEN"
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
    assert!(v["notes"].to_string().contains("advisor"));
}
#[test]
fn launcher_codex_argv() {
    let v = launched(&launcher(&["codex", "-l", "--dry-run"], None));
    assert_eq!(
        v["argv"],
        json!([
            "codex",
            "-s",
            "workspace-write",
            "-a",
            "on-request",
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
            "workspace-write",
            "-a",
            "on-request",
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
/// CRLF bytes, a symlink, one modified and one untracked file.
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
    let v = serde_json::from_slice(&out.stdout).unwrap_or(Value::Null);
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
    assert_eq!(file("link")["kind"], json!("symlink"));
    assert_eq!(file("link")["sha256"], json!(moai_cli::digest(b"crlf.txt")));
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
    fs::remove_file(r.join("link")).unwrap();
    std::os::unix::fs::symlink("dir/b.txt", r.join("link")).unwrap();
    assert_ne!(d1, moai_cli::tree_digest(r, &["skip/"]).unwrap());
    fs::remove_file(r.join("link")).unwrap();
    std::os::unix::fs::symlink("a.txt", r.join("link")).unwrap();
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
const MASKED_KEYS: [&str; 3] = ["target_binding", "plan_digest", "collected_at"];

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
    );
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
    );
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
    // success output carries exactly the required keys and nothing outside properties.
    for (schema, fixture) in [
        ("plan", "06-project-plan-create"),
        ("plan", "10-project-plan-unchanged"),
        ("agents-plan", "16-agents-plan-create"),
        ("package-result", "11-package-inspect-ok"),
        ("package-result", "13-package-resolve-user"),
    ] {
        let s = repo_json(&format!("schemas/protocol/v1/{schema}.schema.json"));
        let out = &repo_json(&format!("{PROTOCOL_FIXTURES}/{fixture}.json"))["stdout"];
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
        // REQ-02A-05: no fixed agent definitions are shipped.
        assert!(package_files()
            .iter()
            .all(|p| !p.starts_with("agents/") && !p.ends_with(".toml")));
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
        let mut want = INTENTS.to_vec();
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
