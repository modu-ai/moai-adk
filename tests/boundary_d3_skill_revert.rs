//! Issue #2392 D3 (decision 6774899518 row 15 = A): the design's integration list names cases
//! outside run/audit/verify/scope. This card only pins them with tests; repairs are follow-up
//! cards. An `#[ignore]` test records a RED (the current code fails it) and states the
//! behaviour a repair must establish; run it with `-- --ignored`.
use moai_cli::project_skills::{
    project_skill_apply, project_skill_revert, ProjectSkillRequest, ProjectSkillRevertRequest,
};
use std::{fs, path::Path};

fn write_source(dir: &Path) {
    fs::create_dir_all(dir.join("references")).unwrap();
    fs::write(
        dir.join("SKILL.md"),
        b"---\nname: moai\ndescription: fixture\n---\n",
    )
    .unwrap();
    fs::write(dir.join("references/a.md"), b"a").unwrap();
}

/// D3-A "user dirty loss": reverting a project skill copy the apply created must not delete a
/// file the user added to that folder afterwards. Today `project_skill_revert` removes the
/// created folder with `remove_dir_all` without comparing it to what the apply wrote, so the
/// user's file is lost (RED).
#[test]
#[ignore = "#2392 D3-A: RED until project-skill-revert keeps a file the user added after apply"]
fn d3a_revert_of_a_created_skill_keeps_a_user_file() {
    let project = tempfile::TempDir::new().unwrap();
    let source = tempfile::TempDir::new().unwrap();
    write_source(source.path());
    let receipt = project_skill_apply(&ProjectSkillRequest {
        protocol_version: 1,
        opt_in: true,
        project_root: project.path().to_path_buf(),
        home: None,
        hosts: vec!["claude".into()],
        source_dir: source.path().to_path_buf(),
    })
    .unwrap();
    let skill = project.path().join(".claude/skills/moai");
    assert!(
        skill.join("SKILL.md").exists(),
        "the apply created the copy"
    );

    let user_file = skill.join("my-notes.md");
    fs::write(&user_file, b"the user's own notes\n").unwrap();

    let reverted = project_skill_revert(&ProjectSkillRevertRequest {
        protocol_version: 1,
        project_root: project.path().to_path_buf(),
        request_id: receipt.request_id.clone(),
    });
    assert!(
        user_file.exists(),
        "revert deleted a file the user added after apply (result: {reverted:?})"
    );
}
