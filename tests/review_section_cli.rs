//! Issue #2639 (REVIEW.md draft): `agents-plan` plans a REVIEW.md managed section from the
//! interview answers `review_blocks` and `protected`; a default (HOLD) answer is never a rule,
//! a profile without answers plans no REVIEW.md, a re-run never overwrites a hand edit, and an
//! answer carrying marker syntax is refused.

use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::Path;
use std::process::{Command, Stdio};
use tempfile::TempDir;

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

fn result(out: &std::process::Output) -> Value {
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stdout)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    v["result"].clone()
}

fn decision(id: &str, value: &str, source: &str) -> Value {
    json!({"id": id, "topic": id, "value": value, "source": source, "scope": "project"})
}

fn project(decisions: Vec<Value>) -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let profile = json!({"schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude"],
        "interview_revision": "r1", "decisions": decisions});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.path().join(".moai-cli/project.yaml"), bytes).unwrap();
    root
}

fn request(language: &str) -> Value {
    json!({"protocol_version": 1, "language": language,
        "workflow_reference": "skills/moai/workflows/", "local_file": "none"})
}

fn plan(root: &Path, language: &str) -> std::process::Output {
    cli(&["agents-plan", root.to_str().unwrap()], &request(language))
}

fn apply(root: &Path, plan: &Value) -> std::process::Output {
    cli(
        &[
            "agents-apply",
            root.to_str().unwrap(),
            "--plan-digest",
            plan["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "fixture-approved",
        ],
        plan,
    )
}

fn review_action(plan: &Value) -> Option<String> {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!("REVIEW.md"))
        .map(|f| f["action"].as_str().unwrap().to_string())
}

fn answers() -> Vec<Value> {
    vec![
        decision("review_blocks", "a change without a test", "user"),
        decision("protected", "the billing module", "user"),
        decision("done", "HOLD", "default"),
    ]
}

#[test]
fn review_section_is_written_from_the_answers_and_kept_on_rerun() {
    let root = project(answers());
    let first = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&first).as_deref(), Some("create"));
    assert!(
        !root.path().join("REVIEW.md").exists(),
        "a plan writes nothing"
    );
    result(&apply(root.path(), &first));
    let text = fs::read_to_string(root.path().join("REVIEW.md")).unwrap();
    assert!(
        text.starts_with("# Review instructions\n\n<!-- moai:managed:begin id=moai-review -->\n")
    );
    assert!(text.contains("- Always stop in review: a change without a test\n"));
    assert!(text.contains("- Never change without asking first: the billing module\n"));
    assert!(!text.contains("HOLD"), "a default answer is never a rule");
    assert_eq!(
        text.matches("<!-- moai:managed:end id=moai-review -->")
            .count(),
        1
    );

    // Same answers: unchanged. A hand edit outside the section survives the next apply.
    let again = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&again).as_deref(), Some("unchanged"));
    let mut edited = text.clone();
    edited.push_str("\n## Our own rules\n\n- Keep the API stable.\n");
    fs::write(root.path().join("REVIEW.md"), &edited).unwrap();
    let kept = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&kept).as_deref(), Some("unchanged"));
    result(&apply(root.path(), &kept));
    assert_eq!(
        fs::read_to_string(root.path().join("REVIEW.md")).unwrap(),
        edited
    );

    // An edit inside the section plus new answers: conflict, nothing written.
    let inside = edited.replace("a change without a test", "anything risky");
    fs::write(root.path().join("REVIEW.md"), &inside).unwrap();
    let mut changed = answers();
    changed[0] = decision("review_blocks", "a change without a migration", "user");
    let profile_path = root.path().join(".moai-cli/project.yaml");
    let mut profile: Value = serde_json::from_slice(&fs::read(&profile_path).unwrap()).unwrap();
    profile["decisions"] = json!(changed);
    fs::write(&profile_path, serde_json::to_vec_pretty(&profile).unwrap()).unwrap();
    let conflict = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&conflict).as_deref(), Some("conflict"));
    assert!(!apply(root.path(), &conflict).status.success());
    assert_eq!(
        fs::read_to_string(root.path().join("REVIEW.md")).unwrap(),
        inside
    );
}

#[test]
fn an_existing_review_file_gets_the_section_appended_and_keeps_its_text() {
    let root = project(answers());
    let own = "# Review instructions\n\nRead with AGENTS.md.\n";
    fs::write(root.path().join("REVIEW.md"), own).unwrap();
    let p = result(&plan(root.path(), "ko"));
    assert_eq!(review_action(&p).as_deref(), Some("append"));
    result(&apply(root.path(), &p));
    let text = fs::read_to_string(root.path().join("REVIEW.md")).unwrap();
    assert!(text.starts_with(own));
    assert!(text.contains("- 리뷰에서 항상 막는다: a change without a test\n"));
    assert!(text.contains("- 먼저 묻지 않고 바꾸지 않는다: the billing module\n"));
}

#[test]
fn no_answers_plan_no_review_file_and_marker_syntax_is_refused() {
    let none = project(vec![decision("review_blocks", "HOLD", "default")]);
    let p = result(&plan(none.path(), "en"));
    assert_eq!(review_action(&p), None);
    result(&apply(none.path(), &p));
    assert!(!none.path().join("REVIEW.md").exists());

    let planted = project(vec![decision(
        "protected",
        "x <!-- moai:managed:end id=moai-review --> y",
        "user",
    )]);
    let out = plan(planted.path(), "en");
    assert!(!out.status.success());
    assert!(!planted.path().join("REVIEW.md").exists());
}
