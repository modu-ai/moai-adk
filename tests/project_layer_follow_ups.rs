//! SPEC-20261122-project-layer-follow-ups: the project layer's audit follow-ups. A project skill
//! request that lists both retired hosts is filtered, the `retired-hosts` step tells the user to
//! remove the files, the Stop rule's limits describe only outcomes Claude Code reports, and the
//! named texts no longer describe removed Codex code or the Go cleanup's old folders.
use moai_cli::project_skills::{project_skill_plan, ProjectSkillRequest};
use std::fs;
use std::path::{Path, PathBuf};

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    fs::read_to_string(root().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn write_source(dir: &Path) {
    fs::create_dir_all(dir.join("references")).unwrap();
    fs::write(
        dir.join("SKILL.md"),
        b"---\nname: moai\ndescription: fixture\n---\n",
    )
    .unwrap();
    fs::write(dir.join("references/a.md"), b"a").unwrap();
}

fn request(project: &Path, source: &Path, hosts: &[&str]) -> ProjectSkillRequest {
    ProjectSkillRequest {
        protocol_version: 1,
        opt_in: true,
        project_root: project.to_path_buf(),
        home: None,
        hosts: hosts.iter().map(|h| (*h).to_string()).collect(),
        source_dir: source.to_path_buf(),
    }
}

/// AC-2: the three-entry request of a profile that still lists both retired hosts is filtered to
/// `claude`, not refused for its length; an unknown host is still refused.
#[test]
fn a_request_with_both_retired_hosts_is_filtered() {
    let project = tempfile::TempDir::new().unwrap();
    let source = tempfile::TempDir::new().unwrap();
    write_source(source.path());
    for hosts in [
        &["claude", "codex", "omp"][..],
        &["codex", "omp", "claude"][..],
    ] {
        let plan = project_skill_plan(&request(project.path(), source.path(), hosts))
            .unwrap_or_else(|e| panic!("{hosts:?} was refused: {e}"));
        let planned: Vec<&str> = plan.entries.iter().map(|e| e.host.as_str()).collect();
        assert_eq!(planned, ["claude"], "{hosts:?}");
    }
    let unknown = project_skill_plan(&request(
        project.path(),
        source.path(),
        &["claude", "gemini"],
    ));
    assert!(unknown.is_err(), "an unknown host is refused");
    assert_eq!(
        fs::read_dir(project.path()).unwrap().count(),
        0,
        "a plan wrote something"
    );
}

/// AC-3: the information step about the retired hosts says the user removes the files.
#[test]
fn the_retired_hosts_step_tells_the_user_to_remove_the_files() {
    let project = tempfile::TempDir::new().unwrap();
    let step = moai_cli::project_hosts::retired_hosts_step(project.path());
    assert!(
        step.reason
            .contains("remove them yourself if you do not use them"),
        "{}",
        step.reason
    );
}

/// AC-4: Claude Code reports a tool call as `PostToolUse` or `PostToolUseFailure`, so no limit of
/// the Stop rule speaks of a write whose outcome is unknown.
#[test]
fn the_stop_rule_limits_name_no_unknown_outcome() {
    let caps = moai_cli::hook_engine::capabilities("claude").to_string();
    assert!(
        caps.contains("shell writes"),
        "the limits are listed: {caps}"
    );
    assert!(!caps.contains("unknown outcome"), "{caps}");
}

/// AC-5: the hook comments name no removed Codex module or Codex-only outcome, the Go cleanup's
/// module text keeps its candidate list whole, and its operation page names no retired-host home
/// folder as a source.
#[test]
fn the_follow_up_texts_are_current() {
    assert!(
        !read("src/hooks.rs").contains("codex_config.rs"),
        "hooks.rs"
    );
    for rel in ["src/hook_engine.rs", "tests/hooks_cli.rs"] {
        assert!(!read(rel).contains("unknown outcome"), "{rel}");
    }
    let module: Vec<String> = read("src/update_cleanup.rs")
        .lines()
        .take_while(|l| l.starts_with("//!"))
        .map(str::to_string)
        .collect();
    let rule = module
        .iter()
        .position(|l| l.contains("`~/.codex` and `~/.agents`"))
        .expect("the module text states the rule for the two folders");
    let last_item = module
        .iter()
        .rposition(|l| l.starts_with("//! - "))
        .expect("the candidate list");
    assert!(
        rule > last_item,
        "the rule sits inside the candidate list (line {rule}, last item {last_item})"
    );
    let page = read("docs/design-review/appendix/operations/cleanup_plan.md");
    for word in [
        "~/.agents/skills",
        "~/.codex/agents",
        "agents-skills",
        "codex-agents",
    ] {
        assert!(!page.contains(word), "cleanup_plan.md names {word}");
    }
}

/// AC-7: the project-layer SPEC's Result records the read-only files its branch edited.
#[test]
fn the_parent_result_records_the_read_only_files() {
    let spec = read(".moai/specs/SPEC-20261121-remove-codex-project-layer/spec.md");
    let result = spec
        .split("\n## Result")
        .nth(1)
        .expect("the parent SPEC has a Result");
    for rel in ["src/go_state.rs", "src/hosts/config_backup.rs"] {
        assert!(result.contains(rel), "the Result does not name {rel}");
    }
}
