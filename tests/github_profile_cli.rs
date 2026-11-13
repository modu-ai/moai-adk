//! #2106: the GitHub template opt-in read from the project profile (`github_issues`).
//! Each test works in its own temporary project; no network and no `gh`.
mod common;

use serde_json::{json, Value};
use std::{
    fs,
    path::{Path, PathBuf},
    process::{Command, Stdio},
};

/// The files `github-apply` writes: four Issue forms (follow-up included, #2434 decision 43)
/// and `labels.yml`.
const TEMPLATE_FILES: usize = 5;

fn cli(args: &[&str], input: &Value) -> Value {
    let mut child = Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    let out = child.wait_with_output().unwrap();
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| {
        panic!(
            "stdout is not JSON: {e}\n{}",
            String::from_utf8_lossy(&out.stdout)
        )
    })
}

fn ok(env: Value) -> Value {
    assert_eq!(env["ok"], true, "{env}");
    env["result"].clone()
}

fn project() -> (tempfile::TempDir, PathBuf) {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    fs::create_dir_all(&root).unwrap();
    (tmp, root)
}

/// A schema-v2 profile as written before #2106: no `github_issues` key.
fn old_profile() -> Value {
    json!({
        "schema_version": 2,
        "project_id": "demo",
        "purpose": "Build a small editor",
        "tools": ["Rust"],
        "development_methods": ["TDD"],
        "hosts": ["claude"]
    })
}

fn write_profile(root: &Path, profile: &Value) {
    let dir = root.join(".moai-cli");
    fs::create_dir_all(&dir).unwrap();
    fs::write(
        dir.join("project.yaml"),
        serde_json::to_vec_pretty(profile).unwrap(),
    )
    .unwrap();
}

fn github_plan(root: &Path, request: Value) -> Value {
    ok(cli(&["github-plan", root.to_str().unwrap()], &request))
}

fn planned_paths(plan: &Value) -> Vec<String> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .map(|f| f["path"].as_str().unwrap().to_string())
        .collect()
}

fn notes(plan: &Value) -> String {
    plan["notes"].to_string()
}

#[test]
fn github_plan_follows_profile_opt_in_without_request_flag() {
    // AC-1: the profile written through project-plan/apply carries the key, and
    // github-plan without a request flag plans the templates from it.
    let (_tmp, root) = project();
    let mut profile = old_profile();
    profile["github_issues"] = json!(true);
    let plan = ok(cli(&["project-plan", root.to_str().unwrap()], &profile));
    assert_eq!(plan["action"], "create");
    ok(cli(
        &[
            "project-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "test:2106",
        ],
        &plan,
    ));
    let written: Value =
        serde_json::from_slice(&fs::read(root.join(".moai-cli/project.yaml")).unwrap()).unwrap();
    assert_eq!(written["github_issues"], json!(true));

    let p = github_plan(&root, json!({"protocol_version": 1}));
    assert_eq!(planned_paths(&p).len(), TEMPLATE_FILES, "{p}");
    assert!(p["files"]
        .as_array()
        .unwrap()
        .iter()
        .all(|f| f["action"] == "create"));
    assert!(notes(&p).contains("opt-in from profile github_issues = true"));
    assert!(p["request"].get("github_issues").is_none());

    // The plan applies, and a second plan is unchanged.
    let applied = ok(cli(
        &[
            "github-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "test:2106",
        ],
        &p,
    ));
    assert_eq!(applied["written"].as_array().unwrap().len(), TEMPLATE_FILES);
    let again = github_plan(&root, json!({"protocol_version": 1}));
    assert!(again["files"]
        .as_array()
        .unwrap()
        .iter()
        .all(|f| f["action"] == "unchanged"));
}

#[test]
fn github_plan_without_profile_opt_in_generates_nothing() {
    // AC-1 negative: an absent key, a false key and an absent profile all plan nothing.
    let (_tmp, root) = project();
    let p = github_plan(&root, json!({"protocol_version": 1}));
    assert_eq!(p["files"], json!([]));
    assert!(notes(&p).contains("no project profile"));

    write_profile(&root, &old_profile());
    let p = github_plan(&root, json!({"protocol_version": 1}));
    assert_eq!(p["files"], json!([]));
    assert!(notes(&p).contains("both absent"));

    let mut off = old_profile();
    off["github_issues"] = json!(false);
    write_profile(&root, &off);
    let p = github_plan(&root, json!({"protocol_version": 1}));
    assert_eq!(p["files"], json!([]));
    assert!(notes(&p).contains("opt-in from profile github_issues = false"));
    assert!(!root.join(".github").exists());
}

#[test]
fn request_flag_wins_over_profile() {
    let (_tmp, root) = project();
    let mut on = old_profile();
    on["github_issues"] = json!(true);
    write_profile(&root, &on);
    let p = github_plan(
        &root,
        json!({"protocol_version": 1, "github_issues": false}),
    );
    assert_eq!(p["files"], json!([]));
    assert!(notes(&p).contains("opt-in from request.github_issues = false"));

    write_profile(&root, &old_profile());
    let p = github_plan(&root, json!({"protocol_version": 1, "github_issues": true}));
    assert_eq!(planned_paths(&p).len(), TEMPLATE_FILES);
    assert!(notes(&p).contains("opt-in from request.github_issues = true"));
}

#[test]
fn profile_change_between_plan_and_apply_makes_the_plan_stale() {
    let (_tmp, root) = project();
    let mut on = old_profile();
    on["github_issues"] = json!(true);
    write_profile(&root, &on);
    let p = github_plan(&root, json!({"protocol_version": 1}));
    assert_eq!(planned_paths(&p).len(), TEMPLATE_FILES);
    write_profile(&root, &old_profile());
    let env = cli(
        &[
            "github-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "test:2106",
        ],
        &p,
    );
    assert_eq!(env["ok"], false, "{env}");
    assert!(env["error"].to_string().contains("stale or altered plan"));
    assert!(!root.join(".github").exists());
}

#[test]
fn existing_profile_without_the_key_reads_unchanged() {
    // AC-2: a pre-#2106 profile is read as it is and re-plans as unchanged; the key is
    // never written when absent.
    let (_tmp, root) = project();
    write_profile(&root, &old_profile());
    let before = fs::read(root.join(".moai-cli/project.yaml")).unwrap();
    let plan = ok(cli(
        &["project-plan", root.to_str().unwrap()],
        &old_profile(),
    ));
    assert_eq!(plan["action"], "unchanged", "{plan}");
    // Absent everywhere, the key stays out of the delta, so the plan is the pre-#2106 one.
    for list in ["added", "changed", "removed", "unchanged", "conflict"] {
        assert!(
            !plan["delta"][list]
                .as_array()
                .unwrap()
                .contains(&json!("github_issues")),
            "{plan}"
        );
    }
    assert!(plan["profile"].get("github_issues").is_none());
    assert_eq!(
        fs::read(root.join(".moai-cli/project.yaml")).unwrap(),
        before
    );
}

#[test]
fn reconfigure_adds_the_key_to_an_existing_profile() {
    // A project that applied a pre-#2106 profile opts in later through project-plan.
    let (_tmp, root) = project();
    let apply = |plan: &Value| {
        ok(cli(
            &[
                "project-apply",
                root.to_str().unwrap(),
                "--plan-digest",
                plan["plan_digest"].as_str().unwrap(),
                "--permission-reference",
                "test:2106",
            ],
            plan,
        ))
    };
    apply(&ok(cli(
        &["project-plan", root.to_str().unwrap()],
        &old_profile(),
    )));
    let mut on = old_profile();
    on["github_issues"] = json!(true);
    let plan = ok(cli(&["project-plan", root.to_str().unwrap()], &on));
    assert_eq!(plan["action"], "update", "{plan}");
    assert_eq!(plan["delta"]["added"], json!(["github_issues"]));
    apply(&plan);
    let p = github_plan(&root, json!({"protocol_version": 1}));
    assert_eq!(planned_paths(&p).len(), TEMPLATE_FILES, "{p}");
}

/// Decision 41 (#2434): the profile's `language.issue_pr_records` picks the form language,
/// falling back to that item's default (`user_chat`).
#[test]
fn profile_issue_language_picks_the_forms() {
    for (language, expected) in [
        (json!({"issue_pr_records": "en"}), "en"),
        (json!({"user_chat": "en"}), "en"),
        (json!({"user_chat": "en", "issue_pr_records": "ko"}), "ko"),
    ] {
        let (_tmp, root) = project();
        let mut profile = old_profile();
        profile["github_issues"] = json!(true);
        profile["language"] = language.clone();
        write_profile(&root, &profile);
        let p = github_plan(&root, json!({"protocol_version": 1}));
        assert!(
            notes(&p).contains(&format!("form language {expected}")),
            "{language}: {p}"
        );
    }
}

/// The form language is part of what the plan renders: a profile language change between
/// plan and apply makes the plan stale instead of writing the other variant.
#[test]
fn profile_language_change_between_plan_and_apply_makes_the_plan_stale() {
    let (_tmp, root) = project();
    let mut profile = old_profile();
    profile["github_issues"] = json!(true);
    write_profile(&root, &profile);
    let p = github_plan(&root, json!({"protocol_version": 1}));
    assert!(notes(&p).contains("form language ko"), "{p}");
    profile["language"] = json!({"issue_pr_records": "en"});
    write_profile(&root, &profile);
    let env = cli(
        &[
            "github-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "test:2434",
        ],
        &p,
    );
    assert_eq!(env["ok"], false, "{env}");
    assert!(env.to_string().contains("stale"), "{env}");
    assert!(!root.join(".github").exists());
}

/// The request language wins over the profile for that run, like `github_issues`.
#[test]
fn request_language_wins_over_profile_language() {
    let (_tmp, root) = project();
    let mut profile = old_profile();
    profile["github_issues"] = json!(true);
    profile["language"] = json!({"issue_pr_records": "en"});
    write_profile(&root, &profile);
    let p = github_plan(&root, json!({"protocol_version": 1, "language": "ko"}));
    assert!(
        notes(&p).contains("form language ko from request.language"),
        "{p}"
    );
}
