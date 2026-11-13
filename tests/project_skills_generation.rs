//! Issue #2706 P2: `generation-plan` builds project skills `prj-<name>` from the profile's
//! interview answers into `.agents/skills/` (Codex CLI) and `.claude/skills/` (Claude Code),
//! records them in `.moai-cli/manifests/generation.json`, never overwrites a hand edit and never
//! claims a folder it did not write. A request without project skills plans as before.

mod common;

use serde_json::{json, Value};
use std::fs;
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
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    child.wait_with_output().unwrap()
}

fn result(out: &std::process::Output) -> Value {
    let v: Value = serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)));
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn decision(id: &str, value: &str, source: &str) -> Value {
    json!({"id": id, "topic": id, "value": value, "source": source, "scope": "project"})
}

fn write_profile(root: &Path, decisions: Vec<Value>) {
    let profile = json!({"schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
        "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"],
        "interview_revision": "r1", "decisions": decisions});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.join(".moai-cli/project.yaml"), bytes).unwrap();
}

fn answers() -> Vec<Value> {
    vec![
        decision("test_command", "cargo test --locked", "user"),
        decision("review_blocks", "a change without a test", "user"),
        decision("protected", "the billing module", "user"),
        decision("build_command", "HOLD", "default"),
    ]
}

fn project() -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    write_profile(root.path(), answers());
    root
}

fn request(skills: Value) -> Value {
    let mut r = json!({"protocol_version": 1, "language": "en", "workflow_reference": "skills/moai/workflows",
        "local_file": "none", "workflow_id": "review-flow", "native_roles": [], "skill_projections": []});
    if !skills.is_null() {
        r["project_skills"] = skills;
    }
    r
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

fn action(p: &Value, path: &str) -> Option<String> {
    p["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!(path))
        .map(|f| f["action"].as_str().unwrap().to_string())
}

const CLAUDE: &str = ".claude/skills/prj-checks/SKILL.md";
const CODEX: &str = ".agents/skills/prj-checks/SKILL.md";

fn checks() -> Value {
    json!([{"id": "prj-checks", "purpose": "How this project is checked"}])
}

#[test]
fn a_project_skill_is_written_for_both_hosts_from_the_answers() {
    let root = project();
    let r = root.path();
    let p = result(&plan(r, &request(checks())));
    assert_eq!(action(&p, CLAUDE).as_deref(), Some("create"));
    assert_eq!(action(&p, CODEX).as_deref(), Some("create"));
    assert!(!r.join(CLAUDE).exists(), "a plan writes nothing");
    result(&apply(r, &p));
    for path in [CLAUDE, CODEX] {
        let text = fs::read_to_string(r.join(path)).unwrap();
        assert!(
            text.starts_with(
                "---\nname: prj-checks\ndescription: How this project is checked\n---\n"
            ),
            "{text}"
        );
        assert!(text.contains("- Run the tests: cargo test --locked"));
        assert!(text.contains("- Review always stops: a change without a test"));
        assert!(text.contains("- Never change without asking first: the billing module"));
        assert!(
            !text.contains("HOLD"),
            "an unanswered axis is never written"
        );
    }
    let manifest: Value =
        serde_json::from_slice(&fs::read(r.join(".moai-cli/manifests/generation.json")).unwrap())
            .unwrap();
    let kinds: Vec<&str> = manifest["owned_files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| f["path"].as_str().unwrap().contains("prj-checks"))
        .map(|f| f["kind"].as_str().unwrap())
        .collect();
    assert_eq!(kinds, vec!["project-skill", "project-skill"]);
}

#[test]
fn a_rerun_keeps_hand_edits_and_conflicts_when_both_sides_move() {
    let root = project();
    let r = root.path();
    result(&apply(r, &result(&plan(r, &request(checks())))));
    let again = result(&plan(r, &request(checks())));
    assert_eq!(action(&again, CLAUDE).as_deref(), Some("unchanged"));

    let edited = "---\nname: prj-checks\ndescription: mine\n---\n\nmine\n";
    fs::write(r.join(CLAUDE), edited).unwrap();
    let kept = result(&plan(r, &request(checks())));
    assert_eq!(action(&kept, CLAUDE).as_deref(), Some("unchanged"));
    result(&apply(r, &kept));
    assert_eq!(fs::read_to_string(r.join(CLAUDE)).unwrap(), edited);

    let mut changed = answers();
    changed[0] = decision("test_command", "cargo nextest run", "user");
    write_profile(r, changed);
    let moved = result(&plan(r, &request(checks())));
    assert_eq!(action(&moved, CLAUDE).as_deref(), Some("conflict"));
    assert_eq!(
        action(&moved, CODEX).as_deref(),
        Some("merge"),
        "the unedited copy follows the answers"
    );
    assert!(!apply(r, &moved).status.success());
    assert_eq!(fs::read_to_string(r.join(CLAUDE)).unwrap(), edited);
}

#[test]
fn a_folder_it_did_not_write_is_never_claimed() {
    let root = project();
    let r = root.path();
    fs::create_dir_all(r.join(".claude/skills/prj-checks")).unwrap();
    fs::write(r.join(CLAUDE), "someone else's skill\n").unwrap();
    let p = result(&plan(r, &request(checks())));
    assert_eq!(action(&p, CLAUDE).as_deref(), Some("conflict"));
    assert_eq!(action(&p, CODEX).as_deref(), Some("create"));
    assert!(!apply(r, &p).status.success());
    assert_eq!(
        fs::read_to_string(r.join(CLAUDE)).unwrap(),
        "someone else's skill\n"
    );
    // A folder holding other files but no SKILL.md is someone else's too (mutation K4).
    let root = project();
    let r = root.path();
    fs::create_dir_all(r.join(".claude/skills/prj-checks")).unwrap();
    fs::write(r.join(".claude/skills/prj-checks/notes.md"), "notes\n").unwrap();
    let p = result(&plan(r, &request(checks())));
    assert_eq!(action(&p, CLAUDE).as_deref(), Some("conflict"));
    assert!(!r.join(CLAUDE).exists());
}

#[test]
fn names_hosts_and_hosts_list_are_checked() {
    let root = project();
    let r = root.path();
    for (skills, ok) in [
        (json!([{"id": "prj-x", "purpose": "p"}]), true),
        (json!([{"id": "checks", "purpose": "p"}]), false),
        (json!([{"id": "moai-checks", "purpose": "p"}]), false),
        (
            json!([{"id": "prj-x", "purpose": "p", "hosts": ["kiro"]}]),
            false,
        ),
        (json!([{"id": "prj-x", "purpose": ""}]), false),
    ] {
        assert_eq!(
            plan(r, &request(skills.clone())).status.success(),
            ok,
            "{skills}"
        );
    }
    let only_codex = result(&plan(
        r,
        &request(json!([{"id": "prj-x", "purpose": "p", "hosts": ["codex"]}])),
    ));
    assert_eq!(
        action(&only_codex, ".agents/skills/prj-x/SKILL.md").as_deref(),
        Some("create")
    );
    assert_eq!(action(&only_codex, ".claude/skills/prj-x/SKILL.md"), None);
}

#[test]
fn a_request_without_project_skills_plans_as_before() {
    let root = project();
    let r = root.path();
    let without = result(&plan(r, &request(Value::Null)));
    let empty = result(&plan(r, &request(json!([]))));
    assert_eq!(without["plan_digest"], empty["plan_digest"]);
    assert!(without["request"].get("project_skills").is_none());
}

fn recover(root: &Path, receipt: &str) -> std::process::Output {
    cli(
        &[
            "apply-recover",
            root.to_str().unwrap(),
            "--receipt",
            receipt,
            "--permission-reference",
            "fixture-approved",
        ],
        &json!({}),
    )
}

const EXTRA_CLAUDE: &str = ".claude/skills/prj-extra/SKILL.md";
const EXTRA_CODEX: &str = ".agents/skills/prj-extra/SKILL.md";

fn two() -> Value {
    json!([{"id": "prj-checks", "purpose": "How this project is checked"},
           {"id": "prj-extra", "purpose": "Another skill"}])
}

/// #2706 P2b: a project skill no longer requested is removed — only while nobody edited it,
/// with its bytes kept in the receipt so `apply-recover` restores it — and leaves the list.
#[test]
fn a_skill_no_longer_requested_is_removed_and_can_be_recovered() {
    let root = project();
    let r = root.path();
    result(&apply(r, &result(&plan(r, &request(two())))));
    let original = fs::read_to_string(r.join(EXTRA_CLAUDE)).unwrap();

    let p = result(&plan(r, &request(checks())));
    assert_eq!(action(&p, EXTRA_CLAUDE).as_deref(), Some("remove"));
    assert_eq!(action(&p, EXTRA_CODEX).as_deref(), Some("remove"));
    assert_eq!(action(&p, CLAUDE).as_deref(), Some("unchanged"));
    assert!(r.join(EXTRA_CLAUDE).exists(), "a plan removes nothing");
    let applied = result(&apply(r, &p));
    assert!(!r.join(EXTRA_CLAUDE).exists() && !r.join(EXTRA_CODEX).exists());
    assert!(
        !r.join(".claude/skills/prj-extra").exists(),
        "the emptied folder goes too"
    );
    assert!(r.join(CLAUDE).exists());
    let manifest: Value =
        serde_json::from_slice(&fs::read(r.join(".moai-cli/manifests/generation.json")).unwrap())
            .unwrap();
    assert!(
        !manifest.to_string().contains("prj-extra"),
        "the list forgets it"
    );
    // Nothing more to remove on the next plan.
    assert_eq!(
        action(&result(&plan(r, &request(checks()))), EXTRA_CLAUDE),
        None
    );

    let receipt = applied["receipt"].as_str().unwrap().to_string();
    result(&recover(r, &receipt));
    assert_eq!(fs::read_to_string(r.join(EXTRA_CLAUDE)).unwrap(), original);
    assert!(r.join(EXTRA_CODEX).exists());
}

/// An edited skill that is no longer requested is a conflict and is never removed.
#[test]
fn an_edited_skill_no_longer_requested_is_kept() {
    let root = project();
    let r = root.path();
    result(&apply(r, &result(&plan(r, &request(two())))));
    fs::write(r.join(EXTRA_CLAUDE), "my own notes\n").unwrap();
    let p = result(&plan(r, &request(checks())));
    assert_eq!(action(&p, EXTRA_CLAUDE).as_deref(), Some("conflict"));
    assert_eq!(action(&p, EXTRA_CODEX).as_deref(), Some("remove"));
    assert!(!apply(r, &p).status.success());
    assert_eq!(
        fs::read_to_string(r.join(EXTRA_CLAUDE)).unwrap(),
        "my own notes\n"
    );
    assert!(
        r.join(EXTRA_CODEX).exists(),
        "a refused apply removes nothing"
    );
}
