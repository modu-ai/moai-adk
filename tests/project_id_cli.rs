//! #3009 S1 (decision A, comment 6889383420): the project's fixed id is the profile's
//! `project_id`. A request that leaves it out gets the existing profile's id, or a random 26
//! character one for a new profile; an existing id such as `demo` is never replaced. Real
//! binary, temporary projects only; no HOME is read or written.

mod common;

use common::cli::{cli, ok_result, profile};
use serde_json::{json, Value};
use std::fs;
use tempfile::TempDir;

fn without_id() -> Value {
    let mut v = profile(None);
    v.as_object_mut().unwrap().remove("project_id");
    v
}

fn plan_of(root: &TempDir, input: &Value) -> Value {
    let out = cli(&["project-plan", root.path().to_str().unwrap()], input);
    assert!(
        out.status.success(),
        "{:?}",
        String::from_utf8_lossy(&out.stdout)
    );
    ok_result(&out)
}

fn apply(root: &TempDir, plan: &Value) {
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
    assert!(
        out.status.success(),
        "{:?}",
        String::from_utf8_lossy(&out.stdout)
    );
}

fn random_form(id: &str) -> bool {
    id.len() == 26
        && id
            .bytes()
            .all(|b| b.is_ascii_lowercase() || (b'2'..=b'7').contains(&b))
}

#[test]
fn a_new_profile_without_an_id_gets_a_random_one_that_plan_and_apply_share() {
    let root = TempDir::new().unwrap();
    let plan = plan_of(&root, &without_id());
    assert_eq!(plan["action"], "create");
    let id = plan["profile"]["project_id"].as_str().unwrap().to_string();
    assert!(random_form(&id), "{id}");
    apply(&root, &plan);
    // A project with no state starts on `.moai/` (#2696 C2b); the profile is the id's home.
    let file = root.path().join(".moai/project.yaml");
    let written: Value = serde_json::from_slice(&fs::read(&file).unwrap()).unwrap();
    assert_eq!(written["project_id"], json!(id));
    // The same request again keeps the id and changes nothing.
    let again = plan_of(&root, &without_id());
    assert_eq!(again["profile"]["project_id"], json!(id));
    assert_eq!(again["action"], "unchanged");
    let bytes = fs::read(&file).unwrap();
    apply(&root, &again);
    assert_eq!(fs::read(&file).unwrap(), bytes);
}

#[test]
fn an_empty_id_is_the_same_as_a_missing_one() {
    let root = TempDir::new().unwrap();
    let mut input = profile(None);
    input["project_id"] = json!("");
    let id = plan_of(&root, &input)["profile"]["project_id"]
        .as_str()
        .unwrap()
        .to_string();
    assert!(random_form(&id), "{id}");
}

#[test]
fn two_new_projects_get_different_ids() {
    let (a, b) = (TempDir::new().unwrap(), TempDir::new().unwrap());
    let ia = plan_of(&a, &without_id())["profile"]["project_id"].clone();
    let ib = plan_of(&b, &without_id())["profile"]["project_id"].clone();
    assert_ne!(ia, ib);
}

#[test]
fn an_id_given_in_the_request_is_used_as_it_is() {
    let root = TempDir::new().unwrap();
    let plan = plan_of(&root, &profile(None));
    assert_eq!(plan["profile"]["project_id"], json!("demo"));
}

#[test]
fn an_existing_id_is_never_replaced_by_a_request_without_one() {
    let root = TempDir::new().unwrap();
    // The profile an earlier build wrote: a person-chosen id, in the earlier `.moai-cli/`.
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let first = plan_of(&root, &profile(None));
    apply(&root, &first);
    let file = root.path().join(".moai-cli/project.yaml");
    assert!(file.is_file());
    let bytes = fs::read(&file).unwrap();
    let plan = plan_of(&root, &without_id());
    assert_eq!(plan["profile"]["project_id"], json!("demo"));
    assert_eq!(plan["action"], "unchanged");
    apply(&root, &plan);
    assert_eq!(fs::read(&file).unwrap(), bytes);
}

#[test]
fn a_profile_that_is_not_json_is_not_given_an_id_and_stays_as_it_is() {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let file = root.path().join(".moai-cli/project.yaml");
    fs::write(&file, "project_id: demo\npurpose: by hand\n").unwrap();
    let out = cli(
        &["project-plan", root.path().to_str().unwrap()],
        &without_id(),
    );
    assert!(!out.status.success());
    let text = String::from_utf8_lossy(&out.stdout);
    assert!(text.contains("preservation review"), "{text}");
    assert_eq!(
        fs::read_to_string(&file).unwrap(),
        "project_id: demo\npurpose: by hand\n"
    );
}
