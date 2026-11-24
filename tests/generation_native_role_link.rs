//! SPEC-20261124-native-role-link-refusal: `generation-plan` and `generation-apply` never write a
//! native-role agent file or a skill projection through a symbolic link, as they already never do
//! for a project skill. A link on any part of the path is a `conflict` in the plan, and a link that
//! appears after the plan stops the apply before it writes anything more. A path without a link
//! is planned and written as before.

mod common;

use serde_json::{json, Value};
use std::fs;
use std::path::Path;
use std::process::Stdio;
use tempfile::TempDir;

const ROLE: &str = ".claude/agents/moai/prj-fixer.md";
const PROJECTION: &str = ".claude/skills/ref-b/SKILL.md";

/// Runs `moai` with a clean environment and a HOME of its own, so the run never reads the HOME of
/// the test harness (the decoy files of `tools/test_home_guard.sh`).
fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let home = common::home::TempHome::new();
    let mut child = home
        .command(args)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    child.wait_with_output().unwrap()
}

fn result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

/// A project with a profile and a license-reviewed source for the projection `ref-b`.
fn project() -> TempDir {
    let root = TempDir::new().unwrap();
    let r = root.path();
    fs::create_dir_all(r.join("skills/moai/workflows")).unwrap();
    let state = r.join(moai_cli::state_root::dir(r));
    fs::create_dir_all(&state).unwrap();
    let profile = json!({"schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "interview_revision": "r1", "decisions": []});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(state.join("project.yaml"), bytes).unwrap();
    let sources = json!({"schema_version": 1, "sources": [
        {"source_url": "fixtures/skill-b", "license": {"status": "reviewed", "spdx": "MIT"},
         "resolved_revision": null}]});
    let at = moai_cli::state_root::path(r, "manifests/sources.json");
    fs::create_dir_all(at.parent().unwrap()).unwrap();
    fs::write(at, serde_json::to_vec_pretty(&sources).unwrap()).unwrap();
    root
}

fn request(roles: bool, projections: bool) -> Value {
    json!({"protocol_version": 1, "language": "en", "workflow_reference": "skills/moai/workflows",
        "local_file": "none", "workflow_id": "review-flow",
        "native_roles": if roles {
            json!([{"role_id": "prj-fixer", "host": "claude", "purpose": "Fix lint failures"}])
        } else { json!([]) },
        "skill_projections": if projections {
            json!([{"id": "ref-b", "source_url": "fixtures/skill-b", "host_targets": ["claude"]}])
        } else { json!([]) }})
}

fn plan(root: &Path, req: &Value) -> std::process::Output {
    cli(&["generation-plan", root.to_str().unwrap()], req)
}

fn apply(root: &Path, p: &Value) -> std::process::Output {
    cli(
        &[
            "generation-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        p,
    )
}

fn file<'a>(p: &'a Value, path: &str) -> &'a Value {
    p["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!(path))
        .unwrap_or_else(|| panic!("{path} is not in the plan: {p}"))
}

fn said(out: &std::process::Output) -> String {
    format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    )
}

/// SPEC revision 3 REQ-2, REQ-5: `generation-apply` plans again before it writes, so a link made
/// after the plan is refused as a stale plan (or, when it appears later, as a link conflict);
/// either way nothing is written through it.
fn assert_refused_for_the_link(out: &std::process::Output) {
    let text = said(out);
    assert!(
        text.contains("symbolic link") || text.contains("stale or altered plan"),
        "{text}"
    );
}

/// Every way a link can sit on `path`: each folder above the file linked to an empty folder
/// outside the project, and the file itself linked to an outside file with bytes of its own.
#[cfg(unix)]
fn assert_every_link_is_a_conflict(path: &str, roles: bool, projections: bool, folders: &[&str]) {
    for link in folders {
        let root = project();
        let r = root.path();
        let outside = TempDir::new().unwrap();
        let at = r.join(link);
        fs::create_dir_all(at.parent().unwrap()).unwrap();
        std::os::unix::fs::symlink(outside.path(), &at).unwrap();
        let p = result(&plan(r, &request(roles, projections)));
        let f = file(&p, path);
        assert_eq!(f["action"], json!("conflict"), "{link}: {f}");
        assert!(
            f["detail"].as_str().unwrap_or("").contains("symbolic link"),
            "{link}: {f}"
        );
        let out = apply(r, &p);
        assert!(!out.status.success(), "{link}: apply must refuse");
        assert_eq!(
            fs::read_dir(outside.path()).unwrap().count(),
            0,
            "{link}: nothing written outside"
        );
    }

    let root = project();
    let r = root.path();
    let outside = TempDir::new().unwrap();
    let target = outside.path().join("target.md");
    fs::write(&target, "outside bytes\n").unwrap();
    fs::create_dir_all(r.join(path).parent().unwrap()).unwrap();
    std::os::unix::fs::symlink(&target, r.join(path)).unwrap();
    let p = result(&plan(r, &request(roles, projections)));
    let f = file(&p, path);
    assert_eq!(f["action"], json!("conflict"), "file link: {f}");
    assert!(
        f["detail"].as_str().unwrap_or("").contains("symbolic link"),
        "file link: {f}"
    );
    assert!(
        !apply(r, &p).status.success(),
        "file link: apply must refuse"
    );
    assert_eq!(fs::read(&target).unwrap(), b"outside bytes\n");
}

/// AC-1 (REQ-1).
#[cfg(unix)]
#[test]
fn a_linked_agents_path_is_a_conflict_and_nothing_is_written_outside() {
    assert_every_link_is_a_conflict(
        ROLE,
        true,
        false,
        &[".claude", ".claude/agents", ".claude/agents/moai"],
    );
}

/// AC-2 (REQ-2): the plan saw no link; one appears before the apply.
#[cfg(unix)]
#[test]
fn a_link_made_after_the_plan_stops_the_apply() {
    let root = project();
    let r = root.path();
    let p = result(&plan(r, &request(true, false)));
    assert_eq!(file(&p, ROLE)["action"], json!("create"), "{p}");
    let outside = TempDir::new().unwrap();
    fs::create_dir_all(r.join(".claude/agents")).unwrap();
    std::os::unix::fs::symlink(outside.path(), r.join(".claude/agents/moai")).unwrap();
    let out = apply(r, &p);
    assert!(!out.status.success(), "{}", said(&out));
    assert_refused_for_the_link(&out);
    assert_eq!(fs::read_dir(outside.path()).unwrap().count(), 0);
}

/// AC-3 (REQ-3): plain folders are planned `create` and written, as before.
#[test]
fn a_plain_agents_and_skills_folder_is_planned_and_written_as_before() {
    let root = project();
    let r = root.path();
    fs::create_dir_all(r.join(".claude/agents")).unwrap();
    fs::create_dir_all(r.join(".claude/skills")).unwrap();
    let p = result(&plan(r, &request(true, true)));
    assert_eq!(file(&p, ROLE)["action"], json!("create"), "{p}");
    assert_eq!(file(&p, PROJECTION)["action"], json!("create"), "{p}");
    let out = apply(r, &p);
    assert!(out.status.success(), "{}", said(&out));
    assert!(fs::read_to_string(r.join(ROLE))
        .unwrap()
        .contains("prj-fixer"));
    assert!(fs::read_to_string(r.join(PROJECTION))
        .unwrap()
        .contains("ref-b"));
}

/// AC-4 (REQ-4): the same refusal for a skill projection, including a link above a skills folder
/// that does not exist yet (the case the "not claimed" check let through).
#[cfg(unix)]
#[test]
fn a_linked_skills_path_is_a_conflict_for_a_projection() {
    assert_every_link_is_a_conflict(
        PROJECTION,
        false,
        true,
        &[".claude", ".claude/skills", ".claude/skills/ref-b"],
    );
}

/// AC-5 (REQ-5): the plan saw no link; one appears before the apply.
#[cfg(unix)]
#[test]
fn a_link_made_after_the_plan_stops_the_projection_apply() {
    let root = project();
    let r = root.path();
    let p = result(&plan(r, &request(false, true)));
    assert_eq!(file(&p, PROJECTION)["action"], json!("create"), "{p}");
    let outside = TempDir::new().unwrap();
    fs::create_dir_all(r.join(".claude")).unwrap();
    std::os::unix::fs::symlink(outside.path(), r.join(".claude/skills")).unwrap();
    let out = apply(r, &p);
    assert!(!out.status.success(), "{}", said(&out));
    assert_refused_for_the_link(&out);
    assert_eq!(fs::read_dir(outside.path()).unwrap().count(), 0);
}
