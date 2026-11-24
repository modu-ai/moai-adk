//! #2936 R9 S4: the project `moai` skill copy never writes or removes through a linked folder.
//! `project_skill_plan` calls a host folder (`.claude`, `.claude/skills`) or the state folder
//! that is a symbolic link, or not a folder, `blocked`; `project_skill_apply` then refuses
//! before any write, and `project_skill_revert` refuses before anything moves. A request that
//! lists the retired host `codex` is filtered, not refused. Temporary directories only; the
//! real HOME is never touched.
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

/// A request that lists `codex` (or `omp`) is not refused: the retired host is filtered out
/// with a warning that names it, the rest is planned, and a list that is empty after the
/// filter reads as `["claude"]`. Nothing is ever planned or written for the retired host.
#[test]
fn a_request_that_lists_codex_is_filtered_not_refused() {
    let f = fixture();
    for (hosts, retired) in [
        (&["claude", "codex"][..], "codex"),
        (&["codex"][..], "codex"),
        (&["omp", "claude"][..], "omp"),
    ] {
        let req = request(&f, hosts);
        let plan = project_skill_plan(&req).expect("a retired host is filtered, not refused");
        let planned: Vec<&str> = plan.entries.iter().map(|e| e.host.as_str()).collect();
        assert_eq!(planned, ["claude"], "{hosts:?}");
        assert!(
            plan.warnings
                .iter()
                .any(|w| w.contains(retired) && w.contains("not managed")),
            "{:?}",
            plan.warnings
        );
    }
    assert_eq!(
        fs::read_dir(&f.project).unwrap().count(),
        0,
        "a plan wrote something"
    );
    // An apply of the same request writes the Claude copy only.
    let receipt = project_skill_apply(&request(&f, &["codex", "claude"])).unwrap();
    assert_eq!(receipt.status, "complete");
    assert_eq!(receipt.applied.len(), 1, "{:?}", receipt.applied);
    assert!(f.project.join(".claude/skills/moai/SKILL.md").is_file());
    assert!(!f.project.join(".agents").exists());
    // A host that never existed is still an error.
    assert!(project_skill_plan(&request(&f, &["zcode"])).is_err());
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
fn a_dangling_link_and_a_file_in_the_place_of_a_folder_are_blocked_too() {
    let f = fixture();
    symlink(f.outside.join("missing"), f.project.join(".claude")).unwrap();
    let plan = project_skill_plan(&request(&f, &["claude"])).unwrap();
    assert_eq!(entry_action(&plan, "claude"), "blocked");

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
    let req = request(&f, &["claude"]);
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
    let req = request(&f, &["claude"]);
    let plan = project_skill_plan(&req).unwrap();
    assert_eq!(entry_action(&plan, "claude"), "blocked");
    assert!(project_skill_apply(&req).is_err());
    nothing_outside(&f);
    assert!(!f.project.join(".claude").exists());
}

#[test]
fn real_folders_still_copy_and_revert_exactly_as_before() {
    let f = fixture();
    let req = request(&f, &["claude"]);
    let plan = project_skill_plan(&req).unwrap();
    assert!(plan.entries.iter().all(|e| e.action == "create"));
    let receipt = project_skill_apply(&req).unwrap();
    assert_eq!(receipt.status, "complete");
    assert!(f.project.join(".claude/skills/moai/SKILL.md").is_file());
    let reverted = project_skill_revert(&ProjectSkillRevertRequest {
        protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
        project_root: f.project.clone(),
        request_id: receipt.request_id,
    })
    .unwrap();
    assert_eq!(reverted["state"], "reverted");
    assert!(!f.project.join(".claude/skills/moai").exists());
}

#[test]
fn a_revert_never_removes_through_a_folder_that_became_a_link() {
    let f = fixture();
    let request_id = project_skill_apply(&request(&f, &["claude"]))
        .unwrap()
        .request_id;
    // The user moves the real `.claude` away and puts a link in its place: the copy now
    // resolves outside the project, and `remove_dir_all` would follow the link.
    let moved = f.outside.join("moved-claude");
    fs::rename(f.project.join(".claude"), &moved).unwrap();
    symlink(&moved, f.project.join(".claude")).unwrap();
    let err = project_skill_revert(&ProjectSkillRevertRequest {
        protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
        project_root: f.project.clone(),
        request_id,
    })
    .unwrap_err();
    assert!(err.contains("symbolic link"), "{err}");
    assert!(err.contains("nothing reverted"), "{err}");
    assert!(
        moved.join("skills/moai/SKILL.md").is_file(),
        "the copy outside the project was removed"
    );
}
