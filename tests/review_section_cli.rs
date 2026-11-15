//! Issue #2639 (REVIEW.md draft): `agents-plan` plans a REVIEW.md managed section from the
//! interview answers `review_blocks` and `protected`; a default (HOLD) answer is never a rule,
//! a profile without answers plans no REVIEW.md, a re-run never overwrites a hand edit, and an
//! answer carrying marker syntax is refused.

mod common;
use serde_json::{json, Value};
use std::fs;
use std::io::Write;
use std::path::Path;
use std::process::{Command, Stdio};
use tempfile::TempDir;

fn cli(args: &[&str], input: &Value) -> std::process::Output {
    let mut child = Command::new(crate::common::MOAI_BIN)
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

// #2936 (from #2722): the first apply records the REVIEW.md managed block as an owned
// baseline, so a later change of the instruction language or of an answer merges instead of
// reporting a conflict nobody caused.

const REVIEW_BEGIN: &str = "<!-- moai:managed:begin id=moai-review -->";
const REVIEW_END: &str = "<!-- moai:managed:end id=moai-review -->\n";
const MANIFEST: &str = ".moai-cli/manifests/generation.json";

fn edit_profile(root: &Path, change: impl FnOnce(&mut Value)) {
    let path = root.join(".moai-cli/project.yaml");
    let mut profile: Value = serde_json::from_slice(&fs::read(&path).unwrap()).unwrap();
    change(&mut profile);
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(path, bytes).unwrap();
}

fn use_korean_instruction_files(root: &Path) {
    edit_profile(root, |p| {
        p["language"] = json!({"instruction_files": "ko"});
    });
}

fn review_text(root: &Path) -> String {
    fs::read_to_string(root.join("REVIEW.md")).unwrap()
}

/// The managed block as the engine sees it: begin marker through the end marker's newline.
fn review_block(text: &str) -> &str {
    let start = text.find(REVIEW_BEGIN).unwrap();
    let end = text.find(REVIEW_END).unwrap() + REVIEW_END.len();
    &text[start..end]
}

fn review_row(plan: &Value) -> Value {
    plan["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!("REVIEW.md"))
        .cloned()
        .unwrap()
}

fn manifest(root: &Path) -> Value {
    serde_json::from_slice(&fs::read(root.join(MANIFEST)).unwrap()).unwrap()
}

/// Plans and applies once; returns the plan.
fn plan_apply(root: &Path) -> Value {
    let p = result(&plan(root, "en"));
    result(&apply(root, &p));
    p
}

#[test]
fn a_language_change_merges_the_review_section_and_keeps_text_outside_it() {
    let root = project(answers());
    let first = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&first).as_deref(), Some("create"));
    result(&apply(root.path(), &first));
    let before = review_text(root.path());
    assert!(!before.contains("프로젝트 리뷰 규칙"));

    // T4: text the user wrote before and after the managed block.
    let own_head = "# Review instructions\n\nHand-written intro.\n\n";
    let own_tail = "\n## Our own rules\n\n- Keep the API stable.\n";
    let edited = format!("{own_head}{}{own_tail}", review_block(&before));
    fs::write(root.path().join("REVIEW.md"), &edited).unwrap();

    // T1: the language change merges, applies and settles.
    use_korean_instruction_files(root.path());
    let second = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&second).as_deref(), Some("merge"));
    result(&apply(root.path(), &second));
    let after = review_text(root.path());
    assert!(
        review_block(&after).contains("프로젝트 리뷰 규칙"),
        "{after}"
    );
    assert!(review_block(&after).contains("리뷰에서 항상 막는다: a change without a test\n"));
    assert!(after.starts_with(own_head), "{after}");
    assert!(after.ends_with(own_tail), "{after}");
    assert_eq!(
        after.len() - review_block(&after).len(),
        own_head.len() + own_tail.len(),
        "nothing outside the block changed"
    );
    let third = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&third).as_deref(), Some("unchanged"));
}

#[test]
fn a_changed_answer_merges_the_review_section() {
    // T2: the answer value changes, the language does not.
    let root = project(answers());
    plan_apply(root.path());
    let mut changed = answers();
    changed[0] = decision("review_blocks", "a change without a migration", "user");
    edit_profile(root.path(), |p| p["decisions"] = json!(changed));
    let second = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&second).as_deref(), Some("merge"));
    result(&apply(root.path(), &second));
    assert!(review_text(root.path())
        .contains("- Always stop in review: a change without a migration\n"));
    let third = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&third).as_deref(), Some("unchanged"));
}

#[test]
fn every_apply_action_moves_the_review_baseline_forward() {
    // A merge records the new block, so a second change merges again (create -> merge -> merge).
    let root = project(answers());
    plan_apply(root.path());
    use_korean_instruction_files(root.path());
    let merged = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&merged).as_deref(), Some("merge"));
    result(&apply(root.path(), &merged));
    let mut changed = answers();
    changed[1] = decision("protected", "the payments module", "user");
    edit_profile(root.path(), |p| p["decisions"] = json!(changed));
    let again = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&again).as_deref(), Some("merge"));
    result(&apply(root.path(), &again));
    assert!(
        review_text(root.path()).contains("먼저 묻지 않고 바꾸지 않는다: the payments module\n")
    );

    // An append records the block too (append -> merge).
    let own = project(answers());
    fs::write(
        own.path().join("REVIEW.md"),
        "# Review instructions\n\nRead with AGENTS.md.\n",
    )
    .unwrap();
    let appended = result(&plan(own.path(), "en"));
    assert_eq!(review_action(&appended).as_deref(), Some("append"));
    result(&apply(own.path(), &appended));
    use_korean_instruction_files(own.path());
    let next = result(&plan(own.path(), "en"));
    assert_eq!(review_action(&next).as_deref(), Some("merge"));
    result(&apply(own.path(), &next));
    assert!(review_text(own.path()).starts_with("# Review instructions\n\nRead with AGENTS.md.\n"));
}

#[test]
fn an_edit_inside_the_review_section_still_conflicts_after_a_language_change() {
    // T3: the baseline is a real digest, not "always merge".
    let root = project(answers());
    plan_apply(root.path());
    let inside = review_text(root.path()).replace("a change without a test", "anything risky");
    fs::write(root.path().join("REVIEW.md"), &inside).unwrap();
    use_korean_instruction_files(root.path());
    let second = result(&plan(root.path(), "en"));
    assert_eq!(review_action(&second).as_deref(), Some("conflict"));
    assert!(!apply(root.path(), &second).status.success());
    assert_eq!(review_text(root.path()), inside);
}

#[test]
fn the_first_apply_records_the_review_block_as_an_owned_baseline() {
    // T5: an owned section with the block digest; REVIEW.md is not a created-once file.
    let root = project(answers());
    let first = plan_apply(root.path());
    let m = manifest(root.path());
    let rows: Vec<&Value> = m["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["path"] == json!("REVIEW.md"))
        .collect();
    assert_eq!(rows.len(), 1, "{m}");
    assert_eq!(rows[0]["section_id"], json!("moai-review"));
    let text = review_text(root.path());
    let block_digest = moai_cli::digest(review_block(&text).as_bytes());
    assert_eq!(rows[0]["baseline_sha256"], json!(block_digest));
    assert_eq!(
        review_row(&first)["rendered_sha256"],
        json!(block_digest),
        "the rendered text is the block the engine finds"
    );
    let created: Vec<&str> = m["created_once"]
        .as_array()
        .unwrap()
        .iter()
        .filter_map(Value::as_str)
        .collect();
    assert!(!created.contains(&"REVIEW.md"), "{m}");
}
