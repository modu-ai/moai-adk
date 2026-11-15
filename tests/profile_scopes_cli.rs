//! #3010 item 1: the `scopes` key of the project profile goes through `project-plan` and
//! `project-apply` like any other key, is refused when a scope is malformed, and leaves a
//! profile that does not carry it unchanged. Real binary, temporary projects only.

mod common;

use common::cli::{cli, ok_result, profile};
use serde_json::{json, Value};
use std::fs;
use tempfile::TempDir;

fn with_scopes(scopes: Value) -> Value {
    let mut v = profile(None);
    v["scopes"] = scopes;
    v
}

fn plan(root: &TempDir, input: &Value) -> std::process::Output {
    cli(&["project-plan", root.path().to_str().unwrap()], input)
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

#[test]
fn declared_scopes_are_planned_written_and_read_back_unchanged() {
    let root = TempDir::new().unwrap();
    let scopes = json!([
        {"name": "web", "paths": ["apps/web", "packages/ui"]},
        {"name": "api", "paths": ["apps/api"]}
    ]);
    let out = plan(&root, &with_scopes(scopes.clone()));
    assert!(
        out.status.success(),
        "{:?}",
        String::from_utf8_lossy(&out.stdout)
    );
    let first = ok_result(&out);
    assert_eq!(first["action"], "create");
    assert_eq!(first["profile"]["scopes"], scopes);
    apply(&root, &first);
    let file = root.path().join(".moai/project.yaml");
    let written: Value = serde_json::from_slice(&fs::read(&file).unwrap()).unwrap();
    assert_eq!(written["scopes"], scopes);

    // the same request again changes nothing
    let again = ok_result(&plan(&root, &with_scopes(scopes)));
    assert_eq!(again["action"], "unchanged");
}

#[test]
fn a_profile_without_scopes_does_not_gain_the_key() {
    let root = TempDir::new().unwrap();
    let p = ok_result(&plan(&root, &profile(None)));
    assert!(p["profile"].get("scopes").is_none(), "{p}");
    apply(&root, &p);
    let written: Value =
        serde_json::from_slice(&fs::read(root.path().join(".moai/project.yaml")).unwrap()).unwrap();
    assert!(written.get("scopes").is_none(), "{written}");
}

#[test]
fn a_malformed_scope_is_refused_and_nothing_is_written() {
    for (bad, why) in [
        (
            json!([{"name": "web", "paths": ["../up"]}]),
            "parent folder",
        ),
        (json!([{"name": "web", "paths": ["/abs"]}]), "absolute path"),
        (json!([{"name": "web", "paths": ["apps/*"]}]), "glob"),
        (
            json!([{"name": "Web", "paths": ["apps/web"]}]),
            "upper-case name",
        ),
        (json!([{"name": "web", "paths": []}]), "no folders"),
        (json!([]), "empty list"),
        (
            json!([{"name": "web", "paths": ["a"]}, {"name": "web", "paths": ["b"]}]),
            "repeated name",
        ),
    ] {
        let root = TempDir::new().unwrap();
        let out = plan(&root, &with_scopes(bad));
        assert!(!out.status.success(), "{why}");
        assert!(
            !root.path().join(".moai/project.yaml").exists(),
            "{why}: a refused plan wrote a profile"
        );
    }
}
