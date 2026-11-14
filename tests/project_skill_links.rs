//! #2936 R9 S4: the project `moai` skill copy never writes or removes through a linked folder.
//! `project_skill_plan` calls a host folder (`.agents`, `.agents/skills`, `.claude`,
//! `.claude/skills`) or the state folder that is a symbolic link, or not a folder, `blocked`;
//! `project_skill_apply` then refuses before any write, and `project_skill_revert` refuses
//! before anything moves. Temporary directories only; the real HOME is never touched.
#![cfg(unix)]

use moai_cli::project_skills::{
    project_skill_apply, project_skill_plan, project_skill_revert, ProjectSkillRequest,
    ProjectSkillRevertRequest, PROJECT_SKILL_SCHEMA_VERSION,
};
use std::{
    fs,
    os::unix::fs::symlink,
    path::{Path, PathBuf},
};
use tempfile::TempDir;

fn write(path: &Path, body: &str) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, body).unwrap();
}

struct Fixture {
    _tmp: TempDir,
    project: PathBuf,
    /// A folder outside the project: nothing may ever land in it or leave it.
    outside: PathBuf,
    source: PathBuf,
}

fn fixture() -> Fixture {
    let tmp = TempDir::new().unwrap();
    let root = fs::canonicalize(tmp.path()).unwrap();
    let project = root.join("project");
    let outside = root.join("outside");
    let source = root.join("source");
    for d in [&project, &outside] {
        fs::create_dir_all(d).unwrap();
    }
    write(
        &source.join("SKILL.md"),
        "---\nname: moai\ndescription: fixture\n---\n# MoAI\n",
    );
    write(&source.join("references/routes.json"), "{\"routes\":1}\n");
    Fixture {
        _tmp: tmp,
        project,
        outside,
        source,
    }
}

fn request(f: &Fixture, hosts: &[&str]) -> ProjectSkillRequest {
    ProjectSkillRequest {
        protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
        opt_in: true,
        project_root: f.project.clone(),
        home: None,
        hosts: hosts.iter().map(|h| h.to_string()).collect(),
        source_dir: f.source.clone(),
    }
}

fn nothing_outside(f: &Fixture) {
    assert_eq!(
        fs::read_dir(&f.outside).unwrap().count(),
        0,
        "something landed outside the project"
    );
}

fn entry_action(plan: &moai_cli::project_skills::ProjectSkillPlan, host: &str) -> String {
    plan.entries
        .iter()
        .find(|e| e.host == host)
        .unwrap()
        .action
        .clone()
}

#[test]
fn a_linked_agents_folder_blocks_the_codex_copy_and_nothing_is_written() {
    let f = fixture();
    symlink(&f.outside, f.project.join(".agents")).unwrap();
    let req = request(&f, &["codex"]);
    let plan = project_skill_plan(&req).unwrap();
    assert_eq!(entry_action(&plan, "codex"), "blocked");
    let reason = plan.entries[0].blocked_reason.clone().unwrap();
    assert!(reason.contains(".agents"), "{reason}");
    assert!(reason.contains("symbolic link"), "{reason}");
    let err = project_skill_apply(&req).unwrap_err();
    assert!(err.contains("nothing written"), "{err}");
    nothing_outside(&f);
    assert!(
        !f.project.join(".moai-cli").exists(),
        "the rollback folder must not be made"
    );
}

#[test]
fn a_linked_skills_folder_blocks_the_codex_copy() {
    let f = fixture();
    fs::create_dir_all(f.project.join(".agents")).unwrap();
    symlink(&f.outside, f.project.join(".agents/skills")).unwrap();
    let req = request(&f, &["codex"]);
    let plan = project_skill_plan(&req).unwrap();
    assert_eq!(entry_action(&plan, "codex"), "blocked");
    assert!(plan.entries[0]
        .blocked_reason
        .as_deref()
        .unwrap()
        .contains(".agents/skills"));
    assert!(project_skill_apply(&req).is_err());
    nothing_outside(&f);
}

#[test]
fn a_linked_claude_folder_blocks_the_claude_copy() {
    let f = fixture();
    symlink(&f.outside, f.project.join(".claude")).unwrap();
    let req = request(&f, &["claude"]);
    let plan = project_skill_plan(&req).unwrap();
    assert_eq!(entry_action(&plan, "claude"), "blocked");
    assert!(project_skill_apply(&req).is_err());
    nothing_outside(&f);
}

#[test]
fn one_linked_host_folder_stops_both_hosts_before_any_write() {
    let f = fixture();
    symlink(&f.outside, f.project.join(".agents")).unwrap();
    let req = request(&f, &["claude", "codex"]);
    let plan = project_skill_plan(&req).unwrap();
    assert_eq!(entry_action(&plan, "claude"), "create");
    assert_eq!(entry_action(&plan, "codex"), "blocked");
    assert!(project_skill_apply(&req).is_err());
    assert!(
        !f.project.join(".claude").exists(),
        "the Claude copy must not be made while the Codex copy is blocked"
    );
    nothing_outside(&f);
}

#[test]
fn a_dangling_link_and_a_file_in_the_place_of_a_folder_are_blocked_too() {
    let f = fixture();
    symlink(f.outside.join("missing"), f.project.join(".agents")).unwrap();
    let plan = project_skill_plan(&request(&f, &["codex"])).unwrap();
    assert_eq!(entry_action(&plan, "codex"), "blocked");

    let g = fixture();
    write(&g.project.join(".claude"), "a file, not a folder\n");
    let plan = project_skill_plan(&request(&g, &["claude"])).unwrap();
    assert_eq!(entry_action(&plan, "claude"), "blocked");
    let reason = plan.entries[0].blocked_reason.clone().unwrap();
    assert!(reason.contains("not a folder"), "{reason}");
    assert!(project_skill_apply(&request(&g, &["claude"])).is_err());
    assert_eq!(
        fs::read_to_string(g.project.join(".claude")).unwrap(),
        "a file, not a folder\n"
    );
}

#[test]
fn a_linked_state_folder_blocks_every_copy() {
    let f = fixture();
    symlink(&f.outside, f.project.join(".moai-cli")).unwrap();
    let req = request(&f, &["claude", "codex"]);
    let plan = project_skill_plan(&req).unwrap();
    assert!(plan.entries.iter().all(|e| e.action == "blocked"));
    assert!(plan.entries[0]
        .blocked_reason
        .as_deref()
        .unwrap()
        .contains(".moai-cli"));
    assert!(project_skill_apply(&req).is_err());
    nothing_outside(&f);
    assert!(!f.project.join(".claude").exists());
    assert!(!f.project.join(".agents").exists());
}

#[test]
fn a_linked_rollback_folder_blocks_every_copy() {
    let f = fixture();
    fs::create_dir_all(f.project.join(".moai-cli")).unwrap();
    symlink(
        &f.outside,
        f.project.join(".moai-cli/project-skill-rollback"),
    )
    .unwrap();
    let req = request(&f, &["codex"]);
    let plan = project_skill_plan(&req).unwrap();
    assert_eq!(entry_action(&plan, "codex"), "blocked");
    assert!(project_skill_apply(&req).is_err());
    nothing_outside(&f);
    assert!(!f.project.join(".agents").exists());
}

#[test]
fn real_folders_still_copy_and_revert_exactly_as_before() {
    let f = fixture();
    let req = request(&f, &["claude", "codex"]);
    let plan = project_skill_plan(&req).unwrap();
    assert!(plan.entries.iter().all(|e| e.action == "create"));
    let receipt = project_skill_apply(&req).unwrap();
    assert_eq!(receipt.status, "complete");
    assert!(f.project.join(".claude/skills/moai/SKILL.md").is_file());
    assert!(f.project.join(".agents/skills/moai/SKILL.md").is_file());
    let reverted = project_skill_revert(&ProjectSkillRevertRequest {
        protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
        project_root: f.project.clone(),
        request_id: receipt.request_id,
    })
    .unwrap();
    assert_eq!(reverted["state"], "reverted");
    assert!(!f.project.join(".agents/skills/moai").exists());
}

#[test]
fn a_revert_never_removes_through_a_folder_that_became_a_link() {
    let f = fixture();
    let req = request(&f, &["codex"]);
    let receipt = project_skill_apply(&req).unwrap();
    // The user moves the real `.agents` away and puts a link in its place: the copy now
    // resolves outside the project, and `remove_dir_all` would follow the link.
    let moved = f.outside.join("moved-agents");
    fs::rename(f.project.join(".agents"), &moved).unwrap();
    symlink(&moved, f.project.join(".agents")).unwrap();
    let err = project_skill_revert(&ProjectSkillRevertRequest {
        protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
        project_root: f.project.clone(),
        request_id: receipt.request_id,
    })
    .unwrap_err();
    assert!(err.contains("symbolic link"), "{err}");
    assert!(err.contains("nothing reverted"), "{err}");
    assert!(
        moved.join("skills/moai/SKILL.md").is_file(),
        "the copy outside the project was removed"
    );
}
