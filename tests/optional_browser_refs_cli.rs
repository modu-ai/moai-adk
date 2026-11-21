//! The bundled Aside and ego(lite) references follow the installed official guides and never
//! need the tool (SPEC-20261121-optional-browser-refs): the shipped files and pack data are read
//! from the repository, and the install runs through the binary in a temporary HOME whose PATH
//! holds no browser tool.

mod common;
use common::home::TempHome;
use serde_json::Value;
use std::{fs, path::Path};

const ASIDE: &str = "templates/bundle-packs/.claude/skills/moai-ref-aside-browser/SKILL.md";
const EGO: &str = "templates/bundle-packs/.claude/skills/moai-ref-ego-browser/SKILL.md";

fn repo_file(rel: &str) -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join(rel))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn packs() -> Value {
    serde_json::from_str(&repo_file("templates/bundle-packs/packs.json")).unwrap()
}

/// The `name:` of a file's front matter.
fn skill_name(text: &str) -> String {
    text.lines()
        .find_map(|l| l.strip_prefix("name:"))
        .map(|n| n.trim().to_string())
        .unwrap_or_default()
}

/// REQ-1, REQ-2.
#[test]
fn the_aside_reference_is_cli_only_and_points_to_the_installed_guide() {
    let text = repo_file(ASIDE);
    assert!(
        text.contains("aside guide"),
        "no pointer to the installed guide"
    );
    assert!(
        text.contains("CLI only"),
        "it must say the command line is the only way in"
    );
    for stale in ["moai mcp add", "Optional registration", "--args mcp"] {
        assert!(
            !text.contains(stale),
            "old MCP advice is still there: {stale}"
        );
    }
}

/// REQ-3, REQ-4: a pointer and a boundary, not a copy of the official interface.
#[test]
fn the_ego_reference_points_to_the_installed_guide_without_copying_it() {
    let text = repo_file(EGO);
    assert_eq!(skill_name(&text), "moai-ref-ego-browser");
    assert!(
        text.contains("ego-browser skill"),
        "no pointer to the installed guide"
    );
    assert!(
        text.contains("## When to use"),
        "no statement of when to use it"
    );
    assert!(
        text.contains("## How far it may go"),
        "no statement of how far it may go"
    );
    for copied in ["await page.", "task.page(", "taskSpace(", "task.finish("] {
        assert!(
            !text.contains(copied),
            "official interface copied: {copied}"
        );
    }
    assert!(
        text.lines().count() <= 120,
        "{} lines",
        text.lines().count()
    );
}

/// REQ-5, REQ-6, REQ-7: the same safe lines in both.
#[test]
fn both_references_state_the_safe_stop_the_limits_and_the_badge_rule() {
    for (rel, tool) in [(ASIDE, "aside"), (EGO, "ego-browser")] {
        let text = repo_file(rel);
        assert!(text.contains("not_run"), "{rel}: no not_run fallback");
        assert!(
            text.contains("as absent"),
            "{rel}: no rule for a failing probe"
        );
        assert!(
            text.contains(&format!("never install or upgrade {tool}"))
                || text.contains("Never install or upgrade"),
            "{rel}: no ban on installing or upgrading"
        );
        assert!(
            text.contains("never change the user's tool settings or MCP settings")
                || text.contains("Never change the user's tool settings or MCP settings"),
            "{rel}: no promise to leave the user's settings alone"
        );
        assert!(
            text.contains("badge") && text.contains("not completion"),
            "{rel}: no badge rule"
        );
    }
    assert!(
        repo_file(EGO).contains("ego-browser import"),
        "the ego(lite) reference must name the import it forbids"
    );
}

/// REQ-8.
#[test]
fn the_ops_tools_pack_lists_both_references_with_matching_names() {
    let data = packs();
    let entries = data["packs"]["ops-tools"]["entries"].as_array().unwrap();
    for (name, rel) in [
        ("moai-ref-aside-browser", ASIDE),
        ("moai-ref-ego-browser", EGO),
    ] {
        let entry = entries
            .iter()
            .find(|e| e["name"] == name)
            .unwrap_or_else(|| panic!("{name} is not in ops-tools"));
        assert_eq!(entry["kind"], "skill");
        assert_eq!(entry["root"], "claude-skills");
        assert_eq!(entry["path"], format!(".claude/skills/{name}"));
        assert_eq!(entry["depends_skills"], serde_json::json!([]));
        assert_eq!(entry["sha256"].as_str().map(str::len), Some(64), "{name}");
        assert_eq!(
            skill_name(&repo_file(rel)),
            name,
            "{rel}: front matter name"
        );
    }
}

/// REQ-9: the install needs neither tool (the clean environment has no browser tool on PATH).
#[test]
fn adding_ops_tools_installs_both_references_with_no_browser_tool_present() {
    let home = TempHome::new();
    let out = home
        .command(&["bundle", "add", "ops-tools"])
        .output()
        .unwrap();
    assert!(
        out.status.success(),
        "exit {:?}: {}{}",
        out.status.code(),
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    for (name, rel) in [
        ("moai-ref-aside-browser", ASIDE),
        ("moai-ref-ego-browser", EGO),
    ] {
        let target = home
            .path()
            .join(".claude/skills")
            .join(name)
            .join("SKILL.md");
        assert_eq!(
            fs::read_to_string(&target).unwrap_or_else(|e| panic!("{}: {e}", target.display())),
            repo_file(rel),
            "{name}: the shipped bytes"
        );
    }
}
