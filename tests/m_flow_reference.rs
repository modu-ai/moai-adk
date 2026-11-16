//! M-Flow (#3096): the collaboration rules live in one reference file, and SKILL.md, MODE-LEADER
//! and MODE-AGENT only point to it. These tests pin that the rules exist, that each of the seven
//! is written in that one file and nowhere else in the skill package, that the leader's and the
//! lane's workflow read the same file, and that no sentence weakens a gate the package keeps.

use std::{fs, path::Path};

const RULES: &str = "skills/moai/references/internal/m-flow.md";
/// The router and the workflows name no package path (one hop from the route table), so they
/// point to the rules through this route-table entry.
const POINTER: &str = "the route table names under `flow_rules`";

fn root() -> &'static Path {
    Path::new(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    fs::read_to_string(root().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

/// Every file of the skill package as `(path from the repository root, text)`. The file list
/// is the source manifest's, which `product_cli` pins to cover every package file, so this test
/// names its files and does not walk the tree.
fn manifest_files() -> Vec<(String, String)> {
    let manifest: serde_json::Value =
        serde_json::from_str(&read("skills/moai/references/source-manifest.json")).unwrap();
    let mut out = Vec::new();
    for name in manifest["files"].as_object().unwrap().keys() {
        let rel = format!("skills/moai/{name}");
        if let Ok(text) = fs::read_to_string(root().join(&rel)) {
            out.push((rel, text));
        }
    }
    assert!(
        out.len() > 20,
        "the manifest lists {} readable files",
        out.len()
    );
    out
}

/// One distinctive phrase per rule, in rule order.
const RULE_PHRASES: [&str; 7] = [
    "it sends the conflict back to the lane that wrote the change",
    "never writes to the integration reference itself",
    "The leader is its only writer, and the tests run on the real combined tree",
    "never push straight to the Go `origin/main`",
    "proceeds on local Git alone, and only the remote features wait",
    "a public release waits for the Linux, macOS and Windows verification of one fixed SHA",
    "No sentence of this flow weakens one of them",
];

#[test]
fn the_seven_rules_are_written_in_the_reference_file() {
    let text = read(RULES);
    assert!(text.starts_with("# Internal: m-flow"));
    for phrase in RULE_PHRASES {
        assert!(text.contains(phrase), "m-flow.md lacks {phrase:?}");
    }
    for needle in [
        "Only the leader orders the integration",
        "`origin/v2026.11.0`",
        "Keep three references apart: the Rust upstream, the local integration reference and the promotion reference",
        "When two of them do not share a history, stop with an error",
        "never force an integration",
        "a protected remote target takes a final pull request or merge request with the required security checks and an independent audit",
        "A local macOS candidate and a public multi-OS release are separate",
        "ownership by claim comment, independent audit, the pull request and CI gate of a shared branch, and the protection rules all stay",
    ] {
        assert!(text.contains(needle), "m-flow.md lacks {needle:?}");
    }
}

#[test]
fn each_rule_is_written_once_in_the_whole_package() {
    for phrase in RULE_PHRASES {
        let holders: Vec<String> = manifest_files()
            .into_iter()
            .filter(|(_, text)| text.contains(phrase))
            .map(|(rel, _)| rel)
            .collect();
        assert_eq!(
            holders,
            vec![RULES.to_string()],
            "{phrase:?} is held by {holders:?}"
        );
    }
}

#[test]
fn skill_and_both_mode_workflows_point_to_the_one_file() {
    let routes: serde_json::Value =
        serde_json::from_str(&read("skills/moai/references/routes.json")).unwrap();
    assert_eq!(routes["flow_rules"], "internal/m-flow.md");
    assert!(root()
        .join("skills/moai/references")
        .join("internal/m-flow.md")
        .is_file());
    for rel in [
        "skills/moai/SKILL.md",
        "skills/moai/workflows/mode-leader.md",
        "skills/moai/workflows/mode-agent.md",
    ] {
        let text = read(rel);
        assert!(text.contains(POINTER), "{rel} does not point to {POINTER}");
        assert!(text.contains("M-Flow"), "{rel} does not name M-Flow");
    }
    // Leader and lane read the same rules: both name the file and both say they repeat none.
    for rel in [
        "skills/moai/workflows/mode-leader.md",
        "skills/moai/workflows/mode-agent.md",
    ] {
        let text = read(rel);
        assert!(
            text.contains("workflow repeats none of them, and none of them weakens a gate, a claim or the independent audit named here."),
            "{rel} lacks the no-repeat, no-weakening sentence"
        );
    }
}

#[test]
fn the_flow_is_named_m_flow_and_weakens_no_gate() {
    for (rel, text) in manifest_files() {
        assert!(
            !text.contains("N Flow"),
            "{rel} uses an old spelling of the name"
        );
    }
    let rules = read(RULES);
    for kept in [
        "GEARS",
        "claim comment",
        "independent audit",
        "CI gate",
        "protection rules",
    ] {
        assert!(rules.contains(kept), "m-flow.md no longer keeps {kept:?}");
    }
    // A rule that weakens a gate would say so with one of these.
    for weaker in [
        "skip the audit",
        "without an audit",
        "bypass",
        "force-push",
        "force push to",
    ] {
        assert!(!rules.contains(weaker), "m-flow.md contains {weaker:?}");
    }
}
