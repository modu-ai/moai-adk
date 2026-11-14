//! #2876 (I4 of #2871): the rules every route, host and base agent follows live inside the
//! `moai` skill package as `references/internal/core-rules.md` — only what changes behaviour,
//! in host-neutral words, under a size cap so loading the skill stays cheap.
use std::{fs, path::Path};

/// The cap the author proposed for the leader's record (load cost of the skill).
const CORE_RULES_MAX_BYTES: usize = 4096;

fn package(rel: &str) -> String {
    fs::read_to_string(
        Path::new(env!("CARGO_MANIFEST_DIR"))
            .join("skills/moai")
            .join(rel),
    )
    .unwrap()
}

#[test]
fn core_rules_stay_small_numbered_and_host_neutral() {
    let text = package("references/internal/core-rules.md");
    assert!(
        text.len() <= CORE_RULES_MAX_BYTES,
        "core-rules.md is {} bytes, cap {CORE_RULES_MAX_BYTES}",
        text.len()
    );
    // Rules are numbered 1..=n without gaps, so a reference "core rule 7" stays stable.
    let numbers: Vec<usize> = text
        .lines()
        .filter_map(|l| l.split_once(". ").and_then(|(n, _)| n.parse().ok()))
        .collect();
    assert_eq!(numbers, (1..=numbers.len()).collect::<Vec<_>>());
    assert!(numbers.len() >= 10, "{numbers:?}");
    // Every host follows them the same way: no host-only tool or feature names.
    for host_only in [
        "AskUserQuestion",
        "ToolSearch",
        "TeamCreate",
        "Agent tool",
        "spawn_agent",
        "settings.json",
        "statusline",
        "/clear",
        "hook",
        "MCP",
        "Claude",
        "Codex",
        ".claude/",
        "moai hook",
    ] {
        assert!(
            !text.contains(host_only),
            "core-rules.md names `{host_only}`"
        );
    }
}

/// SKILL.md sends every session to the file, and every base agent reads it first.
#[test]
fn the_skill_and_every_base_agent_read_core_rules() {
    let skill = package("SKILL.md");
    assert!(skill.contains(
        "read the core rules the route table names under `core_rules`, once per session"
    ));
    // SKILL.md keeps naming only the route table (one hop); the table names the file.
    let routes: serde_json::Value =
        serde_json::from_str(&package("references/routes.json")).unwrap();
    assert_eq!(routes["core_rules"], "internal/core-rules.md");
    let agents = Path::new(env!("CARGO_MANIFEST_DIR")).join("agents");
    let mut seen = 0;
    for entry in fs::read_dir(&agents).unwrap() {
        let path = entry.unwrap().path();
        if path.extension().and_then(|e| e.to_str()) != Some("md") {
            continue;
        }
        let text = fs::read_to_string(&path).unwrap();
        let reads = text
            .lines()
            .find_map(|l| l.strip_prefix("reads: "))
            .unwrap_or_else(|| panic!("{} has no reads line", path.display()));
        assert!(
            reads.starts_with("internal/core-rules"),
            "{}: {reads}",
            path.display()
        );
        seen += 1;
    }
    assert_eq!(seen, 5, "the five base agents");
}

/// The merged rules sit where the work they guard is read.
#[test]
fn merged_rules_live_in_their_workflow_files() {
    for (file, phrase) in [
        (
            "references/internal/git-safety.md",
            "stage by explicit path — never `git add -A`, `git add .` or `git commit -a`",
        ),
        (
            "references/internal/git-safety.md",
            "never switch, create, reset, stash, rebase or merge branches in the shared primary checkout",
        ),
        (
            "references/internal/review.md",
            "A Critical or High security finding blocks",
        ),
        (
            "references/internal/review.md",
            "a missing or stale scan is unverified, never a pass",
        ),
        (
            "workflows/run.md",
            "look its exact name and version up in its registry",
        ),
        (
            "workflows/run.md",
            "An automatic CI repair runs at most three rounds",
        ),
    ] {
        assert!(package(file).contains(phrase), "{file} must say `{phrase}`");
    }
}
