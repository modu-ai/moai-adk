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

fn apply(project: &Path, source: &Path, hosts: &[&str]) -> String {
    project_skill_apply(&ProjectSkillRequest {
        protocol_version: 1,
        opt_in: true,
        project_root: project.to_path_buf(),
        home: None,
        hosts: hosts.iter().map(|h| (*h).to_string()).collect(),
        source_dir: source.to_path_buf(),
    })
    .unwrap()
    .request_id
}

fn revert(project: &Path, request_id: &str) -> Result<serde_json::Value, String> {
    project_skill_revert(&ProjectSkillRevertRequest {
        protocol_version: 1,
        project_root: project.to_path_buf(),
        request_id: request_id.to_string(),
    })
}

/// D3-A "user dirty loss": reverting a project skill copy the apply created must not delete a
/// file the user added to that folder afterwards. The revert removes a created folder only
/// while its digest is still the one the apply recorded; a changed copy refuses the whole
/// revert before anything moves, and the revert works once the user's file is moved out.
#[test]
fn d3a_revert_of_a_created_skill_keeps_a_user_file() {
    let project = tempfile::TempDir::new().unwrap();
    let source = tempfile::TempDir::new().unwrap();
    write_source(source.path());
    let request_id = apply(project.path(), source.path(), &["claude"]);
    let skill = project.path().join(".claude/skills/moai");
    assert!(
        skill.join("SKILL.md").exists(),
        "the apply created the copy"
    );

    let user_file = skill.join("my-notes.md");
    fs::write(&user_file, b"the user's own notes\n").unwrap();

    let reverted = revert(project.path(), &request_id);
    assert!(
        user_file.exists(),
        "revert deleted a file the user added after apply (result: {reverted:?})"
    );
    let err = reverted.expect_err("a changed copy refuses the revert");
    assert!(
        err.contains("claude") && err.contains("nothing reverted"),
        "{err}"
    );
    assert!(skill.join("SKILL.md").exists(), "nothing was removed");

    // With the user's file moved out the copy is the applied one again and the revert runs.
    fs::remove_file(&user_file).unwrap();
    revert(project.path(), &request_id).expect("the unchanged copy is reverted");
    assert!(!skill.exists(), "the created copy is removed");
}

/// D3-A: an edited file counts as a change too, and one changed host refuses the whole revert,
/// so the unchanged host's copy is not removed either (nothing half reverted).
#[test]
fn d3a_one_changed_host_refuses_the_whole_revert() {
    let project = tempfile::TempDir::new().unwrap();
    let source = tempfile::TempDir::new().unwrap();
    write_source(source.path());
    let request_id = apply(project.path(), source.path(), &["claude", "codex"]);
    let claude = project.path().join(".claude/skills/moai");
    let codex = project.path().join(".agents/skills/moai");
    fs::write(codex.join("references/a.md"), b"edited by the user").unwrap();

    let err = revert(project.path(), &request_id).expect_err("an edited copy refuses");
    assert!(err.contains("codex") && !err.contains("claude"), "{err}");
    assert!(
        claude.join("SKILL.md").exists(),
        "the unchanged host was not removed"
    );
    assert_eq!(
        fs::read(codex.join("references/a.md")).unwrap(),
        b"edited by the user"
    );
}

/// D3-A: a copy the digest cannot read (here the user added a symlink, which the staged-package
/// rules refuse) counts as changed, so the revert refuses and removes nothing.
#[cfg(unix)]
#[test]
fn d3a_a_copy_the_digest_cannot_read_is_not_removed() {
    let project = tempfile::TempDir::new().unwrap();
    let source = tempfile::TempDir::new().unwrap();
    write_source(source.path());
    let request_id = apply(project.path(), source.path(), &["claude"]);
    let skill = project.path().join(".claude/skills/moai");
    let link = skill.join("my-link");
    std::os::unix::fs::symlink(project.path(), &link).unwrap();

    let err = revert(project.path(), &request_id).expect_err("an unreadable copy refuses");
    assert!(
        err.contains("claude") && err.contains("nothing reverted"),
        "{err}"
    );
    assert!(link.symlink_metadata().is_ok(), "the user's link is kept");
    assert!(skill.join("SKILL.md").exists(), "nothing was removed");
}
