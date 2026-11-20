//! SPEC-20261120-project-harness-namespace: `generation-plan` and `generation-apply` write the
//! agents and skills of a project in a `moai` namespace (agents in a `moai` folder, skills in
//! `moai-<id>` folders), move what an earlier run wrote into it through the plan, and never touch
//! a file the generation record does not own.

mod common;

use serde_json::{json, Value};
use std::fs;
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
    common::feed_stdin(&mut child, input.to_string().as_bytes());
    child.wait_with_output().unwrap()
}

fn answer(out: &std::process::Output) -> Value {
    serde_json::from_slice(&out.stdout)
        .unwrap_or_else(|_| panic!("{}", String::from_utf8_lossy(&out.stderr)))
}

fn result(out: &std::process::Output) -> Value {
    let v = answer(out);
    assert_eq!(v["ok"], json!(true), "{v}");
    v["result"].clone()
}

fn decision(id: &str, value: &str, source: &str) -> Value {
    json!({"id": id, "topic": id, "value": value, "source": source, "scope": "project"})
}

fn project() -> TempDir {
    let root = TempDir::new().unwrap();
    fs::create_dir_all(root.path().join("skills/moai/workflows")).unwrap();
    fs::create_dir_all(root.path().join(".moai-cli")).unwrap();
    let profile = json!({"schema_version": 2, "project_id": "demo", "purpose": "Build a small editor",
    "tools": ["Rust"], "development_methods": ["TDD"], "hosts": ["claude", "codex"],
    "interview_revision": "r1", "decisions": [
        decision("test_command", "cargo test --locked", "user"),
        decision("review_blocks", "a change without a test", "user"),
    ]});
    let mut bytes = serde_json::to_vec_pretty(&profile).unwrap();
    bytes.push(b'\n');
    fs::write(root.path().join(".moai-cli/project.yaml"), bytes).unwrap();
    root
}

fn role(host: &str, id: &str) -> Value {
    json!({"role_id": id, "host": host, "purpose": "Review a change before it lands",
        "read_only_intent": false, "tools": []})
}

fn request(roles: Vec<Value>, skills: Vec<Value>) -> Value {
    json!({"protocol_version": 1, "language": "en", "workflow_reference": "skills/moai/workflows",
        "local_file": "none", "workflow_id": "review-flow", "native_roles": roles,
        "skill_projections": [], "project_skills": skills})
}

fn skill(id: &str) -> Value {
    json!({"id": id, "purpose": "How this project is checked"})
}

fn both_roles() -> Vec<Value> {
    vec![
        role("claude", "prj-reviewer"),
        role("codex", "prj-reviewer"),
    ]
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

fn entry<'a>(p: &'a Value, path: &str) -> Option<&'a Value> {
    p["files"]
        .as_array()
        .unwrap()
        .iter()
        .find(|f| f["path"] == json!(path))
}

fn action(p: &Value, path: &str) -> Option<String> {
    entry(p, path).map(|f| f["action"].as_str().unwrap().to_string())
}

/// The paths of a plan that belong to the generation (not the AGENTS.md or manifest files).
fn generated(p: &Value) -> Vec<String> {
    p["files"]
        .as_array()
        .unwrap()
        .iter()
        .filter(|f| {
            f["kind"].as_str().is_some_and(|k| {
                k.starts_with("generation:") && !k.starts_with("generation:workflow")
            })
        })
        .map(|f| f["path"].as_str().unwrap().to_string())
        .collect()
}

const CLAUDE_AGENT: &str = ".claude/agents/moai/prj-reviewer.md";
const CODEX_AGENT: &str = ".codex/agents/moai/prj-reviewer.toml";
const CLAUDE_AGENT_OLD: &str = ".claude/agents/prj-reviewer.md";
const CODEX_AGENT_OLD: &str = ".codex/agents/prj-reviewer.toml";
const CLAUDE_SKILL: &str = ".claude/skills/moai-prj-checks/SKILL.md";
const CODEX_SKILL: &str = ".agents/skills/moai-prj-checks/SKILL.md";
const CLAUDE_SKILL_OLD: &str = ".claude/skills/prj-checks/SKILL.md";
const CODEX_SKILL_OLD: &str = ".agents/skills/prj-checks/SKILL.md";

const NAMESPACED: [&str; 4] = [CLAUDE_AGENT, CODEX_AGENT, CLAUDE_SKILL, CODEX_SKILL];
const FORMER: [&str; 4] = [
    CLAUDE_AGENT_OLD,
    CODEX_AGENT_OLD,
    CLAUDE_SKILL_OLD,
    CODEX_SKILL_OLD,
];

fn full_request() -> Value {
    request(both_roles(), vec![skill("prj-checks")])
}

/// Make the project look as an earlier run left it: every namespaced file back at its former
/// path, with its record entry at the former path too (bytes unchanged).
fn as_an_earlier_run_left_it(root: &Path) {
    let manifest_path = root.join(".moai-cli/manifests/generation.json");
    let mut manifest: Value = serde_json::from_slice(&fs::read(&manifest_path).unwrap()).unwrap();
    for (new, old) in NAMESPACED.iter().zip(FORMER) {
        fs::create_dir_all(root.join(old).parent().unwrap()).unwrap();
        fs::rename(root.join(new), root.join(old)).unwrap();
        let mut dir = root.join(new).parent().unwrap().to_path_buf();
        while dir != root && fs::remove_dir(&dir).is_ok() {
            dir = dir.parent().unwrap().to_path_buf();
        }
        for f in manifest["owned_files"].as_array_mut().unwrap() {
            if f["path"] == json!(*new) {
                f["path"] = json!(old);
            }
        }
    }
    let mut bytes = serde_json::to_vec_pretty(&manifest).unwrap();
    bytes.push(b'\n');
    fs::write(manifest_path, bytes).unwrap();
}

/// A project with the namespaced files applied and then put back at the former paths.
fn an_earlier_run() -> TempDir {
    let root = project();
    let r = root.path();
    result(&apply(r, &result(&plan(r, &full_request()))));
    as_an_earlier_run_left_it(r);
    root
}

#[test]
fn fresh_agents_go_in_the_moai_folder_and_skills_in_moai_prefixed_folders() {
    let root = project();
    let r = root.path();
    let p = result(&plan(r, &full_request()));
    let mut got = generated(&p);
    got.sort();
    let mut want: Vec<String> = NAMESPACED.iter().map(|s| s.to_string()).collect();
    want.sort();
    assert_eq!(
        got, want,
        "a plan names the namespaced paths and no former path"
    );
    for path in NAMESPACED {
        assert_eq!(action(&p, path).as_deref(), Some("create"), "{path}");
    }
    for path in FORMER {
        assert!(entry(&p, path).is_none(), "nothing planned at {path}");
    }
    result(&apply(r, &p));
    for path in FORMER {
        assert!(!r.join(path).exists(), "nothing written at {path}");
    }
    let claude = fs::read_to_string(r.join(CLAUDE_AGENT)).unwrap();
    assert!(claude.starts_with("---\nname: prj-reviewer\n"), "{claude}");
    let codex = fs::read_to_string(r.join(CODEX_AGENT)).unwrap();
    assert!(codex.starts_with("name = \"prj-reviewer\"\n"), "{codex}");
    for path in [CLAUDE_SKILL, CODEX_SKILL] {
        let text = fs::read_to_string(r.join(path)).unwrap();
        assert!(
            text.starts_with("---\nname: moai-prj-checks\n"),
            "the name is the folder name: {text}"
        );
    }
}

#[test]
fn fresh_a_second_plan_of_the_same_request_changes_nothing() {
    let root = project();
    let r = root.path();
    result(&apply(r, &result(&plan(r, &full_request()))));
    let again = result(&plan(r, &full_request()));
    for path in NAMESPACED {
        assert_eq!(action(&again, path).as_deref(), Some("unchanged"), "{path}");
    }
}

#[test]
fn folder_name_of_fifty_nine_characters_is_planned_and_sixty_is_refused() {
    let root = project();
    let r = root.path();
    // `moai-` + id: an id of 59 characters makes a folder name of 64, the limit.
    let ok = format!("prj-{}", "a".repeat(55));
    let too_long = format!("prj-{}", "a".repeat(56));
    assert_eq!(ok.len(), 59);
    let p = result(&plan(r, &request(vec![], vec![skill(&ok)])));
    assert_eq!(
        action(&p, &format!(".claude/skills/moai-{ok}/SKILL.md")).as_deref(),
        Some("create")
    );
    let refused = answer(&plan(r, &request(vec![], vec![skill(&too_long)])));
    assert_eq!(refused["ok"], json!(false), "{refused}");
    let text = refused.to_string();
    assert!(
        text.contains("project_skills[0].id"),
        "names the field: {text}"
    );
    assert!(text.contains("64"), "names the limit: {text}");
    assert!(!text.contains(&too_long), "never shows the value: {text}");
}

#[test]
fn moved_owned_files_leave_the_former_path_with_their_bytes_kept() {
    let root = an_earlier_run();
    let r = root.path();
    let before: Vec<(String, Vec<u8>)> = FORMER
        .iter()
        .map(|f| (f.to_string(), fs::read(r.join(f)).unwrap()))
        .collect();
    let p = result(&plan(r, &full_request()));
    for (new, old) in NAMESPACED.iter().zip(FORMER) {
        assert_eq!(action(&p, old).as_deref(), Some("remove"), "{old} leaves");
        assert_eq!(action(&p, new).as_deref(), Some("create"), "{new} arrives");
    }
    for old in FORMER {
        assert!(r.join(old).exists(), "a plan writes nothing: {old}");
    }
    result(&apply(r, &p));
    for ((old, bytes), new) in before.iter().zip(NAMESPACED) {
        assert!(!r.join(old).exists(), "{old} is gone");
        let moved = fs::read(r.join(new)).unwrap();
        // The skill name changes with the folder; the rest of the file is what it was.
        if new.contains("skills") {
            let was = String::from_utf8(bytes.clone()).unwrap();
            let now = String::from_utf8(moved).unwrap();
            assert_eq!(
                now,
                was.replacen("name: prj-checks", "name: moai-prj-checks", 1)
            );
        } else {
            assert_eq!(&moved, bytes, "{new} keeps the bytes of {old}");
        }
    }
    let again = result(&plan(r, &full_request()));
    for path in NAMESPACED {
        assert_eq!(action(&again, path).as_deref(), Some("unchanged"), "{path}");
    }
    for old in FORMER {
        assert!(
            entry(&again, old).is_none(),
            "the record no longer names {old}"
        );
    }
}

#[test]
fn moved_an_edited_file_stays_and_both_paths_are_conflict() {
    let root = an_earlier_run();
    let r = root.path();
    let edited = "---\nname: prj-reviewer\ndescription: mine\n---\n\nmine\n";
    fs::write(r.join(CLAUDE_AGENT_OLD), edited).unwrap();
    let p = result(&plan(r, &full_request()));
    assert_eq!(action(&p, CLAUDE_AGENT_OLD).as_deref(), Some("conflict"));
    assert_eq!(action(&p, CLAUDE_AGENT).as_deref(), Some("conflict"));
    let why = entry(&p, CLAUDE_AGENT_OLD).unwrap()["detail"]
        .as_str()
        .unwrap();
    assert!(why.contains("edited"), "{why}");
    // The unedited ones still move; the edited one is refused as a whole plan.
    assert!(!apply(r, &p).status.success());
    assert_eq!(
        fs::read_to_string(r.join(CLAUDE_AGENT_OLD)).unwrap(),
        edited
    );
    assert!(!r.join(CLAUDE_AGENT).exists());
}

#[test]
fn moved_a_file_of_the_user_at_a_former_agent_path_is_never_changed() {
    let root = project();
    let r = root.path();
    let mine = "---\nname: prj-reviewer\ndescription: my own\n---\n\nmy own\n";
    fs::create_dir_all(r.join(".claude/agents")).unwrap();
    fs::write(r.join(CLAUDE_AGENT_OLD), mine).unwrap();
    let p = result(&plan(
        r,
        &request(vec![role("claude", "prj-reviewer")], vec![]),
    ));
    assert!(
        entry(&p, CLAUDE_AGENT_OLD).is_none(),
        "the user's file is not in the plan"
    );
    assert_eq!(action(&p, CLAUDE_AGENT).as_deref(), Some("conflict"));
    let why = entry(&p, CLAUDE_AGENT).unwrap()["detail"].as_str().unwrap();
    assert!(
        why.contains(CLAUDE_AGENT_OLD) && why.contains("name"),
        "{why}"
    );
    assert!(!apply(r, &p).status.success());
    assert_eq!(fs::read_to_string(r.join(CLAUDE_AGENT_OLD)).unwrap(), mine);
    assert!(!r.join(CLAUDE_AGENT).exists());
}

#[test]
fn moved_a_skill_of_the_user_beside_ours_is_left_alone() {
    let root = project();
    let r = root.path();
    let mine = "---\nname: prj-checks\ndescription: my own\n---\n\nmy own\n";
    fs::create_dir_all(r.join(".claude/skills/prj-checks")).unwrap();
    fs::write(r.join(CLAUDE_SKILL_OLD), mine).unwrap();
    fs::write(r.join(".claude/skills/prj-checks/notes.md"), "notes\n").unwrap();
    let p = result(&plan(r, &request(vec![], vec![skill("prj-checks")])));
    assert!(
        entry(&p, CLAUDE_SKILL_OLD).is_none(),
        "the user's folder is not in the plan"
    );
    assert_eq!(action(&p, CLAUDE_SKILL).as_deref(), Some("create"));
    result(&apply(r, &p));
    assert_eq!(fs::read_to_string(r.join(CLAUDE_SKILL_OLD)).unwrap(), mine);
    assert_eq!(
        fs::read_to_string(r.join(".claude/skills/prj-checks/notes.md")).unwrap(),
        "notes\n"
    );
    assert!(r.join(CLAUDE_SKILL).exists());
}

#[test]
fn moved_a_former_path_edited_after_the_plan_is_not_removed() {
    let root = an_earlier_run();
    let r = root.path();
    let p = result(&plan(r, &full_request()));
    let edited = "---\nname: prj-reviewer\ndescription: edited after the plan\n---\n\nmine\n";
    fs::write(r.join(CLAUDE_AGENT_OLD), edited).unwrap();
    assert!(
        !apply(r, &p).status.success(),
        "a plan over changed bytes is refused"
    );
    assert_eq!(
        fs::read_to_string(r.join(CLAUDE_AGENT_OLD)).unwrap(),
        edited
    );
    assert!(!r.join(CLAUDE_AGENT).exists());
}

#[test]
fn moved_an_owned_agent_that_the_request_no_longer_names_stays_with_a_warning() {
    let root = an_earlier_run();
    let r = root.path();
    let before = fs::read(r.join(CLAUDE_AGENT_OLD)).unwrap();
    let p = result(&plan(
        r,
        &request(
            vec![role("codex", "prj-reviewer")],
            vec![skill("prj-checks")],
        ),
    ));
    assert!(
        entry(&p, CLAUDE_AGENT_OLD).is_none(),
        "an agent not named is not planned"
    );
    let warnings = p["warnings"].to_string();
    assert!(warnings.contains(CLAUDE_AGENT_OLD), "{warnings}");
    result(&apply(r, &p));
    assert_eq!(fs::read(r.join(CLAUDE_AGENT_OLD)).unwrap(), before);
}

fn spec_dir() -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR"))
        .join(".moai/specs/SPEC-20261120-project-harness-namespace")
}

#[test]
fn research_holds_the_official_quotes_and_read_times_the_paths_rest_on() {
    let text = fs::read_to_string(spec_dir().join("research.md")).unwrap();
    for needle in [
        "Claude Code scans `.claude/agents/` and `~/.claude/agents/` recursively",
        "identity comes only from the `name` frontmatter field",
        "Project | `.claude/skills/<skill-name>/SKILL.md`",
        "agent-roles/src/discovery.rs",
        "a1afa122738382adabe2dd1ccfc510a70ed381f0",
        "Codex identifies the custom agent by its `name` field",
    ] {
        let flat: String = text.split_whitespace().collect::<Vec<_>>().join(" ");
        let wanted: String = needle.split_whitespace().collect::<Vec<_>>().join(" ");
        assert!(
            flat.contains(&wanted) || flat.replace('\\', "").contains(&wanted.replace('\\', "")),
            "research.md must hold {needle:?}"
        );
    }
    let times = text
        .split(|c: char| !(c.is_ascii_digit() || c == ':' || c == 'Z'))
        .filter(|t| t.len() == 9 && t.ends_with('Z') && t.as_bytes()[2] == b':')
        .count();
    assert!(
        times >= 6,
        "research.md names its read times (UTC): {times}"
    );
}

/// Step 6 of the Procedure of the project workflow, whitespace collapsed.
fn step_six() -> String {
    let text = fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/project.md"),
    )
    .unwrap();
    let start = text.find("\n6. **").expect("step 6 is missing");
    let rest = &text[start + 1..];
    let end = rest.find("\n7. **").expect("step 7 is missing");
    rest[..end].split_whitespace().collect::<Vec<_>>().join(" ")
}

#[test]
fn text_step_six_names_the_namespace_folders() {
    let step = step_six();
    for phrase in [
        "`.claude/agents/moai/`",
        "`.codex/agents/moai/`",
        "`moai-<id>`",
        "apart from the files the user made",
    ] {
        assert!(step.contains(phrase), "step 6 must say {phrase:?}:\n{step}");
    }
}
