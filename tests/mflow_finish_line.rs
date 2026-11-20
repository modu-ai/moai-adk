//! SPEC-20261120-mflow-finish-line: the M-Flow finish line is the verified local integration
//! branch, and a remote pull request is opened only for a formal release. The tests read the
//! shipped guidance text only; no host, no HOME and no command runs here.

use std::fs;
use std::path::{Path, PathBuf};

const MARKER: &str = "M-Flow finish line (user decision 2026-11-20)";
const MERGE_ORDER: &str = "a lane works on its own card branch, gets its audit, pushes, and posts LANE-READY-INTEGRATION with the exact 40-hex head; the leader alone merges each READY head serially into the LOCAL integration branch (the one saved at first setup; v2026.11.0 here), verifies the merged tree, and fast-forwards the local branch.";
const WORK_ENDS: &str = "That verified local branch is where normal work ends: nobody opens a pull request per card, per bundle or per integration round.";
const RELEASE_ONLY: &str = "A pull request from the local integration branch to the remote one (head branch release/v{head}.{yearweek}.{build}, the full CI on Linux, macOS and Windows, an independent audit by a lane with no commits in the range) is opened only when the user calls a formal release or a production deploy.";

/// The files that tell a lane or the leader how a change ends.
const NAMED_FILES: [&str; 6] = [
    "AGENTS.md",
    "AGENTS.local.md",
    "skills/moai/workflows/mode-agent.md",
    "skills/moai/workflows/mode-leader.md",
    "skills/moai/workflows/sync.md",
    "skills/moai/workflows/verify.md",
];

/// Phrases that tell someone to open a pull request per card, per bundle or per integration.
const OLD_PHRASES: [&str; 9] = [
    "bundle PR",
    "bundle pull request",
    "one release PR",
    "one release pull request",
    "gh pr create --base v2026.11.0",
    "Pull request to the integration branch with `Refs #<n>`",
    "Before opening a pull request, count the open pull requests",
    "a pull request merges when its current head has a successful `ci` run",
    "when the work is done, push the branch and open a pull request (recommended)",
];

fn root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
}

fn read(rel: &str) -> String {
    fs::read_to_string(root().join(rel)).unwrap_or_else(|e| panic!("{rel}: {e}"))
}

#[test]
fn agents_md_states_the_finish_line_once() {
    let text = read("AGENTS.md");
    assert_eq!(text.matches(MARKER).count(), 1, "the marker is stated once");
    for (name, sentence) in [
        ("merge order", MERGE_ORDER),
        ("where work ends", WORK_ENDS),
        ("release only", RELEASE_ONLY),
    ] {
        assert!(text.contains(sentence), "AGENTS.md lacks the {name} sentence");
    }
    let line = text
        .lines()
        .find(|l| l.contains(MARKER))
        .expect("a line holds the marker");
    assert!(
        line.contains(&format!("{MARKER}: {MERGE_ORDER}")),
        "the marker opens the merge-order sentence: {line}"
    );
}

#[test]
fn agents_md_has_no_pull_request_per_issue_step() {
    let text = read("AGENTS.md");
    for old in [
        "6. Pull request to the integration branch with `Refs #<n>`",
        "a pull request merges when its current head has a successful `ci` run",
    ] {
        assert!(!text.contains(old), "AGENTS.md still says `{old}`");
    }
    // The merge rule that stays is the formal release's.
    assert!(
        text.contains("release pull request merges when its current head has a successful `ci` run"),
        "AGENTS.md ties the pull request merge rule to the release pull request"
    );
}

#[test]
fn agents_md_stays_inside_its_budget() {
    let text = read("AGENTS.md");
    assert!(text.len() < 9 * 1024, "{} bytes", text.len());
    let longest = text.lines().map(|l| l.chars().count()).max().unwrap();
    assert!(longest <= 400, "a line is {longest} characters");
}

#[test]
fn every_named_file_states_the_finish_line_once() {
    for rel in NAMED_FILES {
        let count = read(rel).matches(MARKER).count();
        assert_eq!(count, 1, "{rel} states the finish line {count} times");
    }
}

/// Every guidance file, without the rule history and the source manifest, which quote old text.
fn guidance_files() -> Vec<PathBuf> {
    fn walk(dir: &Path, out: &mut Vec<PathBuf>) {
        let Ok(entries) = fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.is_dir() {
                walk(&path, out);
            } else if matches!(
                path.extension().and_then(|e| e.to_str()),
                Some("md") | Some("json")
            ) {
                out.push(path);
            }
        }
    }
    let mut files = vec![root().join("AGENTS.md"), root().join("AGENTS.local.md")];
    walk(&root().join("agents"), &mut files);
    walk(&root().join("skills"), &mut files);
    files.retain(|p| {
        let name = p.file_name().and_then(|n| n.to_str()).unwrap_or("");
        name != "rule-history.md" && name != "source-manifest.json"
    });
    files
}

#[test]
fn no_guidance_opens_a_pull_request_per_bundle() {
    let files = guidance_files();
    assert!(files.len() > 30, "the scan reads {} files", files.len());
    let mut hits = Vec::new();
    for path in &files {
        let Ok(text) = fs::read_to_string(path) else {
            continue;
        };
        for old in OLD_PHRASES {
            if text.contains(old) {
                hits.push(format!("{}: {old}", path.strip_prefix(root()).unwrap().display()));
            }
        }
    }
    assert!(hits.is_empty(), "old finish-line wording remains:\n{}", hits.join("\n"));
}
