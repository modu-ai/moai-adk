//! #2876 (I4 of #2871): the rules every route, host and base agent follows live inside the
//! `moai` skill package as `references/internal/core-rules.md` — only what changes behaviour,
//! in host-neutral words, under a size cap so loading the skill stays cheap.
use std::{fs, path::Path};

/// The cap the author proposed for the leader's record (load cost of the skill), raised from
/// 4096 to 4608 bytes by SPEC-20261120-guidance-sweep-core for the Korean developer-words
/// sentence of rule 7 and the recurring-problem rule 22, and to 5120 bytes by
/// SPEC-20261121-report-rules-in-core-rules for the two report rules that moved out of the
/// skill entry (the same bytes leave `SKILL.md`, so the two files load as many bytes as before).
const CORE_RULES_MAX_BYTES: usize = 5120;

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

/// Card C-87 S1: rule 7 says where the conversation language comes from, that it holds after a
/// compaction, and that a subagent's report reaches the user in it.
#[test]
fn rule_seven_keeps_the_conversation_language_for_relayed_reports() {
    let rules = package("references/internal/core-rules.md");
    let seven = rules
        .lines()
        .find(|l| l.starts_with("7. "))
        .expect("rule 7");
    for phrase in [
        "Answer in the user's conversation language",
        "the project's language setting, which the session-start note repeats after a compaction",
        "relay a subagent's report to the user in that language, never as it came",
    ] {
        assert!(
            seven.contains(phrase),
            "rule 7 must say `{phrase}`: {seven}"
        );
    }
}

/// #3096 S1: rules 19-21 and the report-first sentence of rule 18 keep their scope and reason,
/// and the lenses file reports without changing code.
#[test]
fn prior_art_scope_and_code_wins_rules_keep_their_words() {
    let rules = package("references/internal/core-rules.md");
    for phrase in [
        "19. Before building an unfamiliar feature from scratch, find at least two public implementations of something similar, compare their approaches and write down why you chose yours.",
        "20. Do not design for a rare case or a future situation nobody asked for; note it as a follow-up.",
        "Input checks at trust boundaries, data-loss prevention and security stay in scope, as in rule 3.",
        "21. When an instruction file and the code or a command's real output disagree, the code and the output win; report the disagreement.",
        "is read by someone who did not watch the work: it gives the conclusion first, in plain words, and ends with two lists",
    ] {
        assert!(rules.contains(phrase), "core-rules.md must say `{phrase}`");
    }
    let lenses = package("references/internal/lenses.md");
    assert!(
        lenses.contains("Each lens produces a report and changes no code"),
        "{lenses}"
    );
    for lens in [
        "**State machine.**",
        "**CI in three stretches.**",
        "**What would we keep if we built it again today.**",
        "**Refactor risk.**",
    ] {
        assert!(lenses.contains(lens), "lenses.md lost {lens}");
    }
    for (file, needle) in [
        ("workflows/plan.md", "`internal/lenses`"),
        ("workflows/read.md", "`internal/lenses`"),
        ("references/chains/index.md", "`internal/lenses`"),
    ] {
        assert!(
            package(file).contains(needle),
            "{file} must point to {needle}"
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

/// Card C-81 (narrowed by the leader after audit B1): pushing a tag or publishing a release waits
/// for the user at every autonomy level (case (2), no deadline default); the internal references
/// that restate the decision order say so too. Lifting a HOLD is not part of it: the 2026-11-08
/// user decision lets the leader settle a pending HOLD, and the user decides any change later.
#[test]
fn release_tags_wait_for_the_user_at_every_level() {
    for (file, phrase) in [
        (
            "SKILL.md",
            "pushing a tag or publishing a release is case (2) at every autonomy level, `full` included",
        ),
        (
            "references/internal/decide.md",
            "goes to the user and waits, and so does pushing a tag or publishing a release at every autonomy level;",
        ),
        (
            "references/internal/versioning.md",
            "pushing a tag or publishing a release is `SKILL.md` case (2) at every autonomy level, `full` included: the user picks it and it waits, with no deadline default",
        ),
    ] {
        assert!(package(file).contains(phrase), "{file} must say `{phrase}`");
    }
    for file in ["SKILL.md", "references/internal/decide.md"] {
        assert!(
            !package(file).contains("lifting a HOLD and pushing a tag"),
            "{file} moves lifting a HOLD to case (2) without a user decision"
        );
    }
}
