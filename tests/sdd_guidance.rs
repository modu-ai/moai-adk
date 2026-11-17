//! SPEC-20261117-sdd-guidance: the agent-facing SDD rules live in one reference
//! (`skills/moai/references/internal/sdd.md`), name only commands this build has, and the
//! workflows and the chain table agree on one order: PLAN → RUN → VERIFY → SYNC → AUDIT → merge.
mod common;

use common::CleanEnv;
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
};

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
            rest.split(|c: char| !c.is_ascii_lowercase()).next().unwrap()
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
    let mut files = Vec::new();
    for dir in ["skills/moai/workflows", "skills/moai/references/internal", "agents"] {
        for e in fs::read_dir(root().join(dir)).unwrap() {
            files.push(e.unwrap().path());
        }
    }
    for path in files {
        let text = fs::read_to_string(&path).unwrap();
        let shown = path.strip_prefix(root()).unwrap_or(Path::new("?")).display();
        assert!(!text.contains("DDD"), "{shown} still says DDD");
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
        assert!(read(rel).contains("`internal/sdd`"), "{rel} does not point to internal/sdd");
    }
}
