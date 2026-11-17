//! SPEC-20261117-sdd-guidance: the agent-facing SDD rules live in one reference
//! (`skills/moai/references/internal/sdd.md`), name only commands this build has, and the
//! workflows and the chain table agree on one order: PLAN → RUN → VERIFY → SYNC → AUDIT → merge.
mod common;

use common::CleanEnv;
use std::{fs, path::PathBuf, process::Command};

const ORDER: &str = "PLAN → RUN → VERIFY → SYNC → AUDIT → merge";

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    fs::read_to_string(root().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

fn sdd() -> String {
    read("skills/moai/references/internal/sdd.md")
}

#[test]
fn the_reference_states_the_rules_the_user_decided() {
    let text = sdd();
    assert!(text.starts_with("# Internal: sdd\n\nWhen to read: "));
    for phrase in [
        "Every change that modifies the repository has a SPEC",
        "One SPEC = one branch `spec/<yyyymmdd>-<slug>` = one pull request",
        "--minimal",
        "exactly one `shall`",
        "IF <condition>, THEN",
        "(REQ-<n>)",
        "· failure: none",
        "[NEEDS CLARIFICATION",
        "RUN does not start while one is open",
        "Only the user's in-session selection approves a SPEC",
        "digest",
        "`Moai-Spec: <SPEC-ID>`",
        "`@cc:later`",
        "preserve-then-change",
        ORDER,
    ] {
        assert!(text.contains(phrase), "sdd.md lacks {phrase:?}");
    }
    // The session trailer is planned, never presented as available.
    let session = text
        .lines()
        .find(|l| l.contains("Moai-Session"))
        .expect("sdd.md names the session trailer");
    assert!(session.contains("planned"), "{session}");
}

/// Every `moai spec <action>` the reference names is a subcommand of this build. The card
/// link (`moai card spec`, `--require-approved-spec`) comes with a separate change: while this
/// build lacks it, every line that names it says it is planned; once the build has it, that
/// wording may go.
#[test]
fn the_reference_names_only_commands_this_build_has() {
    let text = sdd();
    let mut actions: Vec<&str> = text
        .match_indices("`moai spec ")
        .map(|(i, m)| {
            let rest = &text[i + m.len()..];
            rest.split(|c: char| !c.is_ascii_lowercase())
                .next()
                .unwrap()
        })
        .collect();
    actions.sort_unstable();
    actions.dedup();
    assert_eq!(actions, ["check", "list", "new", "show"]);
    let home = tempfile::tempdir().unwrap();
    for action in &actions {
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("HOME", home.path())
            .args(["spec", action, "--help"])
            .output()
            .unwrap();
        assert!(out.status.success(), "moai spec {action} --help");
    }
    let card = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", home.path())
        .args(["card", "spec", "--help"])
        .output()
        .unwrap();
    if !card.status.success() {
        let lines: Vec<&str> = text
            .lines()
            .filter(|l| l.contains("moai card spec") || l.contains("--require-approved-spec"))
            .collect();
        assert!(!lines.is_empty(), "sdd.md does not mention the card link");
        for line in lines {
            assert!(
                line.contains("planned") && line.contains("once that change lands"),
                "the card link is named as available: {line}"
            );
        }
    }
}

#[test]
fn no_card_number_or_ddd_wording_in_the_package_and_agents() {
    let card = |s: &str| {
        s.match_indices("C-")
            .any(|(i, _)| s[i + 2..].starts_with(|c: char| c.is_ascii_digit()))
    };
    assert!(!card(&sdd()), "sdd.md names a card number");
    // Named files only (no folder walk): every workflow the route table names, the internal
    // references the method question and the order live in, the skill entry and the agents.
    let routes: serde_json::Value =
        serde_json::from_str(&read("skills/moai/references/routes.json")).unwrap();
    let mut files: Vec<String> = routes["routes"]
        .as_object()
        .unwrap()
        .values()
        .map(|r| {
            let rel = r["reference"].as_str().unwrap();
            format!(
                "skills/moai/{}",
                rel.strip_prefix("../").expect("route reference")
            )
        })
        .collect();
    for name in ["sdd", "interview", "questions", "test-design", "core-rules"] {
        files.push(format!("skills/moai/references/internal/{name}.md"));
    }
    files.push("skills/moai/references/chains/index.md".into());
    files.push("skills/moai/SKILL.md".into());
    for agent in ["plan", "research", "impl", "audit", "docs"] {
        files.push(format!("agents/{agent}.md"));
    }
    assert!(files.len() >= 22, "{files:?}");
    for rel in files {
        assert!(!read(&rel).contains("DDD"), "{rel} still says DDD");
    }
}

/// The order is stated in the chain table and in every workflow it orders, and the old
/// build order (audit before docs) is gone.
#[test]
fn every_place_states_one_order() {
    for rel in [
        "skills/moai/references/chains/index.md",
        "skills/moai/workflows/run.md",
        "skills/moai/workflows/verify.md",
        "skills/moai/workflows/sync.md",
        "skills/moai/workflows/audit.md",
    ] {
        assert!(read(rel).contains(ORDER), "{rel} lacks the order");
    }
    for rel in [
        "skills/moai/references/chains/index.md",
        "skills/moai/SKILL.md",
        "skills/moai/workflows/run.md",
        "skills/moai/workflows/sync.md",
        "skills/moai/workflows/audit.md",
    ] {
        assert!(
            !read(rel).contains("moai-impl, moai-audit, moai-docs"),
            "{rel} keeps the old order"
        );
    }
    for rel in [
        "skills/moai/workflows/plan.md",
        "skills/moai/workflows/run.md",
        "skills/moai/workflows/sync.md",
        "skills/moai/workflows/audit.md",
    ] {
        assert!(
            read(rel).contains("`internal/sdd`"),
            "{rel} does not point to internal/sdd"
        );
    }
}
