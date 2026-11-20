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
        "One SPEC = one branch `spec/<slug>-<yyyymmdd>` = one pull request",
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
    // The session trailer exists since SDD-6: the line names the commands that print and check
    // it and the reference that holds its rules, and no longer calls it planned.
    let session = text
        .lines()
        .find(|l| l.contains("Moai-Session"))
        .expect("sdd.md names the session trailer");
    for phrase in [
        "`moai trailers --spec <SPEC-ID>`",
        "`moai trailer-check <base>..HEAD`",
        "`internal/commit-trailers`",
        // SPEC-20261118-skill-audit-followups REQ-1 (SDD-5 audit N-a): the session line is
        // conditional on the host giving an id.
        "when the host gives a session id",
        "otherwise `Moai-Spec` only",
    ] {
        assert!(session.contains(phrase), "{session} lacks {phrase}");
    }
    assert!(!session.contains("planned"), "{session}");
}

/// SPEC-20261118-skill-audit-followups REQ-3 (SDD-6 audit N1): the trailer reference tells CI
/// and a detached checkout to name the branch, in the `trailer-check` item itself.
#[test]
fn the_trailer_reference_names_the_branch_flag_for_ci() {
    let text = read("skills/moai/references/internal/commit-trailers.md");
    let item = text
        .lines()
        .find(|l| l.contains("`moai trailer-check <base>..HEAD`"))
        .expect("commit-trailers.md names trailer-check");
    for phrase in ["`--branch <name>`", "detached", "CI"] {
        assert!(item.contains(phrase), "{item} lacks {phrase}");
    }
}

/// Every `moai spec <action>` the reference names is a subcommand of this build. The card
/// link (`moai card spec`, `--require-approved-spec`) comes with a separate change: while this
/// build lacks it, every line that names it says it is planned; once the build has it, no such
/// line may still say so (audit N1: the reference must then be updated).
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
    assert_eq!(actions, ["check", "close", "list", "new", "show", "take"]);
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
    // The trailer commands the session line names (SDD-6) are subcommands of this build.
    for command in ["trailers", "trailer-check"] {
        assert!(text.contains(&format!("`moai {command} ")), "{command}");
        let out = Command::new(common::MOAI_BIN)
            .clean_env()
            .env("HOME", home.path())
            .args([command, "--help"])
            .output()
            .unwrap();
        assert!(out.status.success(), "moai {command} --help");
    }
    let card = Command::new(common::MOAI_BIN)
        .clean_env()
        .env("HOME", home.path())
        .args(["card", "spec", "--help"])
        .output()
        .unwrap();
    let lines: Vec<&str> = text
        .lines()
        .filter(|l| l.contains("moai card spec") || l.contains("--require-approved-spec"))
        .collect();
    assert!(!lines.is_empty(), "sdd.md does not mention the card link");
    let available = card.status.success();
    for line in lines {
        let planned = line.contains("planned") || line.contains("once that change lands");
        if available {
            assert!(
                !planned,
                "`moai card spec` exists: say it is available: {line}"
            );
        } else {
            assert!(
                line.contains("planned") && line.contains("once that change lands"),
                "the card link is named as available: {line}"
            );
        }
    }
}

/// Audit B2: the two user rules no test held — existing code changes start from committed
/// characterization tests, and agents never add `@cc:` comments on their own.
#[test]
fn preserve_then_change_and_the_cc_rule_are_stated() {
    for (rel, phrase) in [
        (
            "skills/moai/workflows/run.md",
            "For a change to existing code (preserve-then-change), first commit characterization \
             tests that pass on the base and pin the behaviour the change must keep, then make the \
             change and show they still pass.",
        ),
        (
            "skills/moai/references/internal/sdd.md",
            "preserve-then-change for existing code",
        ),
        (
            "skills/moai/references/internal/sdd.md",
            "Agents never add them on their own; SYNC keeps the existing ones true to the code.",
        ),
        (
            "skills/moai/workflows/sync.md",
            "keeps the existing `@cc:` comments true to the code (it never adds new ones)",
        ),
        (
            "agents/docs.md",
            "change only existing `@cc:` comment lines, never add one.",
        ),
    ] {
        assert!(read(rel).contains(phrase), "{rel} lacks {phrase:?}");
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
/// Verify wave 2: a new check that blocks starts in observe — the rule sentence in the SDD
/// guidance, the one-line reminder in the full SPEC template, and the manifest entry reason
/// naming this rule (the digest itself stays with the source-manifest check).
#[test]
fn new_blocking_gates_start_in_observe() {
    let sdd = sdd();
    for phrase in [
        "starts in observe",
        "would have blocked",
        "false positives",
        "the switch to enforce is a recorded decision",
    ] {
        assert!(sdd.contains(phrase), "sdd.md lacks {phrase:?}");
    }
    let template = read("templates/spec/full.md");
    assert!(
        template.contains("starts in observe"),
        "templates/spec/full.md lacks the observe reminder"
    );
    let manifest = read("skills/moai/references/source-manifest.json");
    let at = manifest
        .find("\"references/internal/sdd.md\"")
        .expect("the manifest lists the guidance file");
    // The entry runs to the next top-level key, not to a fixed window: an entry before this
    // one that grows long must not push the reason out of the searched span.
    let entry_end = manifest[at + 1..]
        .find("\n    \"")
        .map(|at| at + 1)
        .unwrap_or(manifest.len());
    let entry = &manifest[at..at + entry_end];
    assert!(
        entry.contains("gates-start-in-observe"),
        "the manifest entry reason does not name the rule change"
    );
}
