//! #2286 decision row 22: the skill package is named `moai`, the folder name an
//! earlier-generation (Go-era) MoAI skill already uses. A `moai` folder this tool did
//! not write is blocked with guidance and never replaced automatically; the
//! backup-then-replace path is a separate confirmed procedure and is not implemented.
//!
//! (i) update-plan names the Go-era possibility and the user's options in the reason.
//! (ii) a foreign project copy plans `blocked`; apply refuses before any write.
//! (iii) a copy this tool made earlier still plans `replace` (boundary).
//! Every fixture runs in temporary directories; the real HOME is never touched.

use moai_cli::project_skills::{
    project_skill_apply, project_skill_plan, ProjectSkillRequest, PROJECT_SKILL_SCHEMA_VERSION,
};
use moai_cli::tree_digest;
use std::{fs, path::Path};
use tempfile::TempDir;

fn write(path: &Path, body: &str) {
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(path, body).unwrap();
}

fn write_source(dir: &Path, description: &str) {
    write(
        &dir.join("SKILL.md"),
        &format!("---\nname: moai\ndescription: {description}\n---\n# MoAI\n"),
    );
    write(&dir.join("references/routes.json"), "{\"routes\":1}\n");
}

fn write_go_skill(dir: &Path) {
    write(
        &dir.join("SKILL.md"),
        "---\nname: moai\ndescription: go\n---\nGo-era MoAI skill body\n",
    );
    write(&dir.join("workflows/plan.md"), "# Go plan workflow\n");
}

fn request(project: &Path, source: &Path) -> ProjectSkillRequest {
    ProjectSkillRequest {
        protocol_version: PROJECT_SKILL_SCHEMA_VERSION,
        opt_in: true,
        project_root: project.to_path_buf(),
        home: None,
        hosts: vec!["claude".into(), "codex".into()],
        source_dir: source.to_path_buf(),
    }
}

/// (i) A Go-era `skills/moai` in the HOME blocks update-plan; the reason keeps the
/// #2268 wording and adds the earlier-generation guidance.
#[cfg(unix)]
#[test]
fn update_plan_reason_names_the_go_era_skill_and_the_user_options() {
    use serde_json::Value;
    let root = TempDir::new().unwrap();
    let home = root.path().join("home");
    let stage = root.path().join("stage");
    write_source(&stage, "fixture package");
    for host_dir in [".claude/skills", ".agents/skills"] {
        write_go_skill(&home.join(host_dir).join("moai"));
    }
    let out = std::process::Command::new(env!("CARGO_BIN_EXE_moai-cli"))
        .args([
            "update-plan",
            home.to_str().unwrap(),
            stage.to_str().unwrap(),
        ])
        .env("HOME", &home)
        .stdin(std::process::Stdio::null())
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    for h in v["result"]["hosts"].as_array().unwrap() {
        assert_eq!(h["preflight"], "blocked", "{h}");
        let reason = h["blocked_reason"].as_str().unwrap();
        // Existing #2268 wording is kept.
        assert!(reason.contains("unmanaged existing package"), "{reason}");
        assert!(reason.contains("backup-then-replace"), "{reason}");
        // Row 22 guidance is appended.
        assert!(
            reason.contains("earlier-generation (Go-era) MoAI `moai` skill"),
            "{reason}"
        );
        assert!(reason.contains("will not replace it"), "{reason}");
        assert!(reason.contains("move it aside"), "{reason}");
        assert!(
            reason.contains("confirmed backup-then-replace procedure"),
            "{reason}"
        );
    }
}

/// (ii) A foreign `moai` folder inside the project plans `blocked` for both hosts;
/// apply refuses before any write — the foreign trees are byte-identical and no
/// rollback directory is created.
#[test]
fn foreign_project_moai_is_blocked_and_apply_writes_nothing() {
    let project = TempDir::new().unwrap();
    let source = TempDir::new().unwrap();
    write_source(source.path(), "rc");
    let mut before = Vec::new();
    for host_dir in [".claude/skills", ".agents/skills"] {
        let go = project.path().join(host_dir).join("moai");
        write_go_skill(&go);
        before.push((go.clone(), tree_digest(&go, &[]).unwrap()));
    }
    let req = request(project.path(), source.path());

    let plan = project_skill_plan(&req).unwrap();
    assert_eq!(plan.entries.len(), 2);
    for entry in &plan.entries {
        assert_eq!(entry.action, "blocked", "{}", entry.host);
        let reason = entry.blocked_reason.as_deref().unwrap();
        assert!(
            reason.contains("earlier-generation (Go-era) MoAI `moai` skill"),
            "{reason}"
        );
        assert!(reason.contains("will not replace it"), "{reason}");
    }

    let err = project_skill_apply(&req).unwrap_err();
    assert!(err.contains("blocked"), "{err}");
    for (go, digest) in before {
        assert_eq!(tree_digest(&go, &[]).unwrap(), digest, "{}", go.display());
    }
    assert!(
        !project.path().join(".moai-cli").exists(),
        "a refused apply must not create the rollback directory"
    );
}

/// (ii, one host) A single blocked host refuses the whole apply: the other, vacant
/// host is not created either.
#[test]
fn one_blocked_host_refuses_the_whole_apply() {
    let project = TempDir::new().unwrap();
    let source = TempDir::new().unwrap();
    write_source(source.path(), "rc");
    let go = project.path().join(".claude/skills/moai");
    write_go_skill(&go);
    let digest = tree_digest(&go, &[]).unwrap();
    let req = request(project.path(), source.path());
    let plan = project_skill_plan(&req).unwrap();
    let actions: Vec<(&str, &str)> = plan
        .entries
        .iter()
        .map(|e| (e.host.as_str(), e.action.as_str()))
        .collect();
    assert_eq!(actions, [("claude", "blocked"), ("codex", "create")]);
    assert!(project_skill_apply(&req).is_err());
    assert_eq!(tree_digest(&go, &[]).unwrap(), digest);
    assert!(!project.path().join(".agents/skills/moai").exists());
    assert!(!project.path().join(".moai-cli").exists());
}

/// (iii, boundary) A copy this tool made — recorded in its own rollback receipt —
/// still plans `unchanged` for the same source and `replace` for a changed one.
#[test]
fn own_earlier_copy_still_plans_replace() {
    let project = TempDir::new().unwrap();
    let source = TempDir::new().unwrap();
    write_source(source.path(), "v1");
    let req = request(project.path(), source.path());
    assert_eq!(project_skill_apply(&req).unwrap().status, "complete");
    let again = project_skill_plan(&req).unwrap();
    assert!(again.entries.iter().all(|e| e.action == "unchanged"));

    let source2 = TempDir::new().unwrap();
    write_source(source2.path(), "v2");
    let req2 = request(project.path(), source2.path());
    let plan = project_skill_plan(&req2).unwrap();
    for entry in &plan.entries {
        assert_eq!(entry.action, "replace", "{}", entry.host);
        assert!(entry.blocked_reason.is_none());
    }
    assert_eq!(project_skill_apply(&req2).unwrap().status, "complete");
    let skill = fs::read_to_string(project.path().join(".claude/skills/moai/SKILL.md")).unwrap();
    assert!(skill.contains("v2"));

    // A user edit after our copy makes the folder foreign again: blocked.
    write(
        &project.path().join(".claude/skills/moai/SKILL.md"),
        "hand edited\n",
    );
    let edited = project_skill_plan(&req).unwrap();
    assert_eq!(edited.entries[0].action, "blocked");
    assert_eq!(edited.entries[1].action, "replace");
}

/// A foreign folder the package digest cannot read (here: a symlink inside) is
/// blocked, not a plan error and never followed.
#[cfg(unix)]
#[test]
fn unreadable_foreign_folder_is_blocked() {
    let project = TempDir::new().unwrap();
    let source = TempDir::new().unwrap();
    write_source(source.path(), "rc");
    let go = project.path().join(".claude/skills/moai");
    write_go_skill(&go);
    std::os::unix::fs::symlink("/nonexistent-target", go.join("link")).unwrap();
    let mut req = request(project.path(), source.path());
    req.hosts = vec!["claude".into()];
    let plan = project_skill_plan(&req).unwrap();
    assert_eq!(plan.entries[0].action, "blocked");
    assert_eq!(plan.entries[0].current_sha256, None);
    assert!(project_skill_apply(&req).is_err());
    assert!(go.join("link").symlink_metadata().is_ok());
}
