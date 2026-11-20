//! SPEC-20261120-mflow-finish-line: the M-Flow finish line is the verified local integration
//! branch, and a remote pull request is opened only for a formal release. The tests read the
//! shipped guidance text only; no host, no HOME and no command runs here.

use std::fs;
use std::path::{Path, PathBuf};

const MARKER: &str = "M-Flow (user decisions 2026-11-20; one canonical text, put once per file)";
/// `AGENTS.local.md` has no room for the full text and points to `AGENTS.md` step 6.
const LOCAL_MARKER: &str =
    "M-Flow finish line (user decisions 2026-11-20; the full six-item text is step 6 of `AGENTS.md`)";
// The leader's final wording (user instruction 2026-11-20 ~10:45Z), item by item. AGENTS.md puts
// the items that run past 400 characters on continuation lines, so each piece is one line.
const PIECES: [(&str, &str); 10] = [
    ("2 FAIL is rerun, not claimed flaky", "A `FAIL` in the checks line is not merged on a flaky claim alone: the lane reruns the full check on the same merge and posts the `PASS` line."),
    ("1 lane commits, pushes, audit", "Lane: commits on its own card branch and pushes it; gets the independent audit (tier 2: the leader starts it)."),
    ("2 lane merges, checks, MERGE-READY", "Lane: after PASS, makes the merge of its head onto the current LOCAL integration branch tip (the primary branch saved at first setup; v2026.11.0 here) in its own worktree, runs `moai check affected <tip>..<merge>` with the binary built in that worktree, and posts `MERGE-READY <card> merge=<40-hex> base=<40-hex> checks=<one line>` on the card."),
    ("3 fast-forward under a lock", "Fast-forward: one at a time under a lock (`~/.moai/locks/<project>-<branch>.lock`), the primary checkout of the integration branch is fast-forwarded with `git merge --ff-only` to that merge."),
    ("3 isolation and the allowed command", "Where the host isolates a lane session to its worktree (Claude Code), the leader's supervision round does this step and writes `MERGED-LOCAL`; where a moai primary-checkout guard is on, the single fixed command `git -C <primary> merge --ff-only refs/moai/lane-merge` is the one allowed (hook-config allow_commands)."),
    ("3 base moved", "If the base moved, the lane redoes step 2 on the new tip (a check result measured on an earlier tree is never copied to the new one)."),
    ("4 local test and install", "Local test and install: builds for local testing and installation are made from the verified LOCAL integration branch tip (not from the remote branch); the leader checks the local tip as a whole from time to time and repairs a broken tip with a revert commit."),
    ("4 statusline", "The statusline shows the branch with commits ahead (↑) and behind (↓) its upstream so local merges are visible."),
    ("5 release only", "Release: only for a real deployment the leader creates `release/v{head}.{yearweek}.{build}` from the local integration branch and opens one pull request to the remote integration branch (full CI on Linux, macOS and Windows, independent audit by a lane with no commits in the range). No pull request per card, per bundle or per integration."),
    ("6 card done", "A card that landed locally is closed with `card done` on the local landing evidence (the merge on the local integration branch), not a pull request."),
];

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
const OLD_PHRASES: [&str; 16] = [
    // The second wording (the lane merges with a fixed ref), replaced by the final one.
    "M-Flow (user decision 2026-11-20, replaces the earlier wording)",
    "under a lock (`mkdir ~/.moai/locks/<project>-<integration>.lock`)",
    "the ref moved with a compare-and-swap update-ref",
    "the lane fast-forwards the primary checkout of the integration branch",
    // The first wording of 2026-11-20 (the leader alone merges), replaced the same morning.
    "M-Flow finish line (user decision 2026-11-20)",
    "the leader alone merges each READY head serially",
    "posts LANE-READY-INTEGRATION with the exact 40-hex head; the leader alone",
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
    for (name, sentence) in PIECES {
        assert!(
            text.contains(sentence),
            "AGENTS.md lacks the {name} sentence"
        );
    }
    let line = text
        .lines()
        .find(|l| l.contains(MARKER))
        .expect("a line holds the marker");
    assert!(
        line.starts_with(&format!("6. {MARKER}:")),
        "the marker opens step 6: {line}"
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
        text.contains(
            "release pull request merges when its current head has a successful `ci` run"
        ),
        "AGENTS.md ties the pull request merge rule to the release pull request"
    );
}

#[test]
fn agents_md_stays_inside_its_budget() {
    let text = read("AGENTS.md");
    // Codex stops adding AGENTS.md files once the chain reaches `project_doc_max_bytes`, 32 KiB by
    // default (https://developers.openai.com/codex/guides/agents-md, UTF-8 bytes). Claude Code's
    // memory page also recommends fewer than 200 lines per file (https://code.claude.com/docs/en/memory,
    // advice, not a limit); this file's line count is in the message below.
    assert!(
        text.len() < 32 * 1024,
        "{} bytes in {} lines",
        text.len(),
        text.lines().count()
    );
    let longest = text.lines().map(|l| l.chars().count()).max().unwrap();
    assert!(longest <= 400, "a line is {longest} characters");
}

#[test]
fn every_named_file_states_the_finish_line_once() {
    for rel in NAMED_FILES {
        let text = read(rel);
        if rel == "AGENTS.local.md" {
            // This file is at its own 32 KiB limit, so it points to the full text in AGENTS.md.
            let count = text.matches(LOCAL_MARKER).count();
            assert_eq!(count, 1, "{rel} points to the finish line {count} times");
            continue;
        }
        let count = text.matches(MARKER).count();
        assert_eq!(count, 1, "{rel} states the finish line {count} times");
        for (name, sentence) in PIECES {
            assert!(text.contains(sentence), "{rel} lacks the {name} sentence");
        }
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
                hits.push(format!(
                    "{}: {old}",
                    path.strip_prefix(root()).unwrap().display()
                ));
            }
        }
    }
    assert!(
        hits.is_empty(),
        "old finish-line wording remains:\n{}",
        hits.join("\n")
    );
}
