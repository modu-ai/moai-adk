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
