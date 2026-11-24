//! This repository's own `CLAUDE.md` is tracked and holds only the facts a team sharing the
//! repository needs, within the 200 lines the official memory guidance recommends
//! (`SPEC-20261123-claude-md-rewrite`). Operating detail lives in `docs/lane-rules.md` and
//! `docs/leader-rules.md`, the origin of each rule in `docs/claude-md-rule-provenance.md`, and
//! machine facts in the ignored `CLAUDE.local.md`. Claude Code loads `CLAUDE.md` from the
//! session's own working tree, so a lane worktree gets the project rules only when git carries
//! the file. `AGENTS.md` and `AGENTS.local.md` are removed (user decision 2026-11-22,
//! `SPEC-20261122-claude-md-layout` Plan 5, AC-6).
mod common;

use std::{fs, path::Path};

fn root() -> &'static Path {
    Path::new(env!("CARGO_MANIFEST_DIR"))
}

fn read(path: &str) -> String {
    fs::read_to_string(root().join(path)).unwrap_or_else(|e| panic!("{path}: {e}"))
}

fn claude_md() -> String {
    read("CLAUDE.md")
}

const SPEC_DIR: &str = ".moai/specs/SPEC-20261123-claude-md-rewrite";
const BEGIN: &str = "<!-- moai:managed:begin id=moai-workflow -->";
const END: &str = "<!-- moai:managed:end id=moai-workflow -->";

/// Card numbers (`C-` and digits, not inside a longer word) in `text`.
fn card_numbers(text: &str) -> Vec<String> {
    let bytes = text.as_bytes();
    let mut hits = Vec::new();
    for (i, _) in text.match_indices("C-") {
        let digits = bytes[i + 2..]
            .iter()
            .take_while(|b| b.is_ascii_digit())
            .count();
        let before_ok = i == 0 || !bytes[i - 1].is_ascii_alphanumeric();
        if digits > 0 && before_ok {
            hits.push(text[i..i + 2 + digits].to_string());
        }
    }
    hits
}

fn assert_has(text: &str, what: &str, phrases: &[&str]) {
    for phrase in phrases {
        assert!(text.contains(phrase), "{what} must say `{phrase}`");
    }
}

fn assert_lacks(text: &str, what: &str, phrases: &[&str]) {
    for phrase in phrases {
        assert!(!text.contains(phrase), "{what} still says `{phrase}`");
    }
}

/// #2667 AC-1, AC-6: no ignore rule hides `CLAUDE.md` — a re-added `/CLAUDE.md` line in
/// `.gitignore` would bring back worktrees without the project rules.
#[test]
fn claude_md_is_not_git_ignored() {
    let out = common::git::git_raw(
        root(),
        &["check-ignore", "--no-index", "-v", "--", "CLAUDE.md"],
    );
    assert!(
        !out.success && out.stdout.is_empty(),
        "CLAUDE.md is git-ignored: {}{}",
        out.stdout,
        out.stderr
    );
    assert!(out.stderr.is_empty(), "check-ignore failed: {}", out.stderr);
}

/// AC-6: the personal file stays ignored, and the two originals are no longer tracked.
#[test]
fn claude_local_md_is_ignored_and_the_agents_files_are_gone() {
    let out = common::git::git_raw(
        root(),
        &["check-ignore", "--no-index", "-v", "--", "CLAUDE.local.md"],
    );
    assert!(
        out.success && out.stdout.contains("CLAUDE.local.md"),
        "CLAUDE.local.md is not git-ignored: {}{}",
        out.stdout,
        out.stderr
    );
    // An index entry is looked up by name (`:<path>`), so the test reads only the named paths;
    // `Cargo.toml` shows the lookup itself works.
    let in_index =
        |path: &str| common::git::git_raw(root(), &["cat-file", "-e", &format!(":{path}")]).success;
    assert!(
        in_index("Cargo.toml"),
        "the index lookup does not work here"
    );
    for gone in ["AGENTS.md", "AGENTS.local.md"] {
        assert!(!in_index(gone), "{gone} is still tracked");
    }
}

/// REQ-1, REQ-11: the tracked file keeps the shape `agents-apply` maintains — exactly one
/// managed section, at the end — and stays within 200 lines with no machine-specific path.
#[test]
fn tracked_claude_md_keeps_one_managed_section_within_200_lines() {
    let text = claude_md();
    assert_eq!(text.matches(BEGIN).count(), 1);
    assert_eq!(text.matches(END).count(), 1);
    assert!(
        text.trim_end().ends_with(END),
        "the managed section is not the last part"
    );
    let lines = text.lines().count();
    assert!(lines <= 200, "CLAUDE.md has {lines} lines");
    for machine in ["/Users/", "/home/", "C:\\Users\\"] {
        assert!(!text.contains(machine), "CLAUDE.md names {machine}");
    }
}

/// REQ-11: the managed section names the hosts of the project profile, and the profile names
/// Claude Code only (the section is edited by hand, `SPEC-20261123-claude-md-rewrite` D-3).
#[test]
fn managed_section_matches_the_profile_hosts() {
    let profile: serde_json::Value =
        serde_json::from_str(&read(".moai/project.yaml")).expect("the profile is JSON");
    let hosts: Vec<&str> = profile["hosts"]
        .as_array()
        .expect("hosts")
        .iter()
        .map(|h| h.as_str().expect("host name"))
        .collect();
    assert_eq!(hosts, ["claude"], "the profile still names another host");
    let purpose = profile["purpose"].as_str().expect("purpose");
    assert!(!purpose.contains("Codex"), "{purpose}");
    let text = claude_md();
    let managed = &text[text.find(BEGIN).expect("managed section")..];
    assert!(managed.contains("hosts: claude."), "{managed}");
    assert!(
        managed.contains(purpose),
        "the purpose line differs: {managed}"
    );
    assert!(!managed.contains("codex"), "{managed}");
}

/// The gate, HOLD, secret and audit sentences stay worded as before: the file got shorter by
/// moving text out, never by weakening one of them. Their words are pinned.
#[test]
fn shortened_claude_md_keeps_its_gate_and_hold_sentences() {
    let text = claude_md();
    assert_has(
        &text,
        "CLAUDE.md",
        &[
            "Anything not recorded there is still HOLD.",
            "Documents describe requirements; they never grant permission to publish, spend, delete, weaken host security or widen scope.",
            "HOLD items are reported blocked, never decided by a lane.",
            "Secrets never enter the repo, cards, PRs, reports or messages",
            "Never redo a hook-denied action through another tool or path",
            "A session other than the author audits the exact head",
            "`AUDIT <card> (<scope>) head <40hex> base <40hex> — verdict: PASS|FAIL`, with no qualifiers; conditions make it FAIL.",
            "The author never audits.",
            "A lane only commits: no push, no pull request, no merge of a moved tip.",
            "`moai card land <card> --head <40hex> --by leader`",
            "Nobody moves the integration branch by hand, by a reset or by a push",
            "Remote checks never block a landing",
            "The author, the auditor and the lander of one change are three different sessions.",
            "`moai/audit` is posted only when the `audit-gate` run succeeded on the same head",
            "An audit dispatch names its caps: no whole `cargo test`, named test modules only.",
            "No branch switch, reset or stash in the shared checkout",
            "Stage by explicit pathspec after re-reading `git status --short`",
            "Never route around an unapproved permission, spend, HOLD or security gate; stop after three failures of the same step.",
        ],
    );
}

/// S1: the work ledger is the local card ledger, not GitHub Issues.
#[test]
fn stale_ledger_is_the_local_card_ledger() {
    let text = claude_md();
    assert_has(
        &text,
        "CLAUDE.md",
        &["local cards (`moai card`, user decision 2026-11-16) are the work ledger"],
    );
    assert_lacks(
        &text,
        "CLAUDE.md",
        &[
            "GitHub Issues are the work ledger",
            "PLAN #<issue>",
            "the lowest live comment id wins",
            "Moai-Claim:",
        ],
    );
}

/// S2: the launcher flags `-l|-a` are retired (`SPEC-20261124-launcher-retire-role-flags`, REQ-9);
/// a session starts with `moai cc|glm` and an optional `--name`.
#[test]
fn stale_launcher_flags_are_retired() {
    let text = claude_md();
    assert_has(
        &text,
        "CLAUDE.md",
        &[
            "Start a session with `moai cc|glm`: a plain host session with no role (model `opus`, effort `medium`); `--name <name>` names it.",
            "The launcher flags `-l|-a` are retired (user decision D-1 2026-11-09, `SPEC-20261124-launcher-retire-role-flags`) and the launcher refuses them.",
        ],
    );
    assert_lacks(
        &text,
        "CLAUDE.md",
        &[
            "`-l|-a` only name the session and set its access",
            "deprecated since",
            "#2164",
        ],
    );
}

/// S3: a mode flag is request text for the skill, not a route.
#[test]
fn stale_mode_flags_are_request_text() {
    let text = claude_md();
    assert_has(
        &text,
        "CLAUDE.md",
        &["`/moai --mode-leader` or `/moai --mode-agent`, which the skill reads as request text"],
    );
    assert_lacks(
        &text,
        "CLAUDE.md",
        &["only mode spellings", "two mode workflows", "--autonomy"],
    );
}

/// S4: Claude Code and GLM only; Codex and OMP are withdrawn.
#[test]
fn stale_hosts_are_claude_only() {
    let text = claude_md();
    assert_has(&text, "CLAUDE.md", &["Codex and OMP are withdrawn"]);
    assert_lacks(
        &text,
        "CLAUDE.md",
        &[
            "Hosts: Claude Code and Codex",
            "hosts: claude, codex",
            "`moai cc|glm|codex`",
            "pastes into a Codex window",
        ],
    );
}

/// S5: new worktrees under `.claude/worktrees/`, the state folder is `.moai/`.
#[test]
fn stale_paths_and_worktrees_follow_the_decisions() {
    let text = claude_md();
    assert_has(
        &text,
        "CLAUDE.md",
        &[
            "New worktrees go to `<project>/.claude/worktrees/<name>`",
            "project state is `<project>/.moai/`",
            "`.moai/specs/`, `.moai/project.yaml` and `.moai/evals/` are tracked",
        ],
    );
    assert_lacks(
        &text,
        "CLAUDE.md",
        &[
            "New worktrees go to `~/.moai/worktrees/<project>-<id>`",
            "project state `<project>/.moai-cli/` (tracked)",
            "tags `moai-cli-v",
            "`~/.moai-cli/bin/moai-cli`",
        ],
    );
}

/// The integration branch is `NeXT` (user decision 2026-11-23); `v2026.11.0` is only a recovery
/// ref that stopped at `b3ba2b30d5`.
#[test]
fn stale_integration_branch_is_next() {
    let text = claude_md();
    assert_has(
        &text,
        "CLAUDE.md",
        &[
            "Integration target `NeXT` (local branch, user decision 2026-11-23).",
            "`work_flow.integration_branch` of `.moai/project.yaml` (`NeXT`)",
        ],
    );
    assert_lacks(&text, "CLAUDE.md", &["Integration target `v2026.11.0`"]);
}

/// N2 of the audit: the local branch `v2026.11.0` and the remote one are two different things.
/// The local one is a recovery ref; the remote one is ruleset-protected and is the base of the
/// release pull request (`.github/workflows/audit-gate.yml` runs on pull requests to it).
#[test]
fn the_local_and_the_remote_v2026_branch_are_told_apart() {
    let text = claude_md();
    assert_has(
        &text,
        "CLAUDE.md",
        &[
            "The local branch `v2026.11.0` is a recovery ref that stopped at `b3ba2b30d5`",
            "the remote `v2026.11.0`",
            "ruleset 26791615",
            "base of the release pull request",
        ],
    );
    let gate = read(".github/workflows/audit-gate.yml");
    assert!(
        gate.contains("branches: [v2026.11.0]"),
        "the release pull request base named in CLAUDE.md is no longer the gate's branch"
    );
}

/// B1 of the audit: the staging rule keeps the raw `references/` folder out of git. The folder
/// exists untracked in the checkout root and no ignore rule covers it, so the sentence and its
/// row of the rule table must both say it.
#[test]
fn staging_rule_keeps_raw_references_unstaged() {
    let text = claude_md();
    assert_has(
        &text,
        "CLAUDE.md",
        &["`.mcp.json` and raw `references/` are never staged"],
    );
    let rows = table_rows();
    let row = rows.iter().find(|r| r[0] == "R-103").expect("R-103");
    assert!(
        row[2].contains("raw `references/`"),
        "R-103 rule: {}",
        row[2]
    );
    assert_eq!(row[4], "CLAUDE.md");
    assert!(
        row[5].contains("raw `references/` are never staged"),
        "R-103 anchor: {}",
        row[5]
    );
}

/// N1 of the audit: three rules the first rewrite dropped stay in the file, each with a row of
/// the rule table that holds its words as the anchor.
#[test]
fn the_three_dropped_rules_are_back_with_a_row_each() {
    let phrases = [
        "Commits already pushed are not rewritten.",
        "the conversation with the user and reports to the user are written in Korean",
        "This rule moves no tag and no package version in `Cargo.toml`",
    ];
    assert_has(&claude_md(), "CLAUDE.md", &phrases);
    let rows = table_rows();
    for phrase in phrases {
        let row = rows
            .iter()
            .find(|r| r[5].contains(phrase))
            .unwrap_or_else(|| panic!("no row of the rule table is anchored on `{phrase}`"));
        assert_eq!(row[4], "CLAUDE.md", "{}: the destination", row[0]);
    }
}

/// N3 of the audit: the tracked SPEC files and docs carry no personal path. The folder of the
/// session's profile and the encoded project path of the memory folder are written as
/// placeholders.
#[test]
fn tracked_rewrite_files_hold_no_personal_path() {
    for path in [
        "CLAUDE.md".to_string(),
        "docs/lane-rules.md".to_string(),
        "docs/leader-rules.md".to_string(),
        "docs/claude-md-rule-provenance.md".to_string(),
        format!("{SPEC_DIR}/spec.md"),
        format!("{SPEC_DIR}/rule-table.md"),
        format!("{SPEC_DIR}/claude-local-draft.md"),
    ] {
        assert_lacks(
            &read(&path),
            &path,
            &["/Users/", "-Users-", "claude-profiles/", "goos"],
        );
    }
    assert_has(
        &read(&format!("{SPEC_DIR}/claude-local-draft.md")),
        "the local draft",
        &["<encoded project path>", "<Orca profile>"],
    );
}

/// N4 of the audit: `~/.claude` is the one canonical home; the profile folder of Orca reads it
/// through links and an `@` import. The draft must not read the other way round.
#[test]
fn local_draft_names_the_home_claude_folder_as_canonical() {
    let draft = read(&format!("{SPEC_DIR}/claude-local-draft.md"));
    assert_has(
        &draft,
        "the local draft",
        &[
            "`~/.claude` is the canonical home",
            "links and an `@` import",
        ],
    );
    assert_lacks(
        &draft,
        "the local draft",
        &[
            "not `~/.claude/settings.json`",
            "a separate copy that no session reads",
        ],
    );
}

/// N6 of the audit: a lane does not widen a user rule. Only the operator opens a new session.
#[test]
fn lane_rules_keep_the_operator_as_the_only_one_who_opens_sessions() {
    let rules = read("docs/lane-rules.md");
    assert_has(
        &rules,
        "docs/lane-rules.md",
        &["New lane sessions are opened only by the operator; no session spawns another."],
    );
    assert_lacks(&rules, "docs/lane-rules.md", &["or by the leader"]);
}

/// N7 of the audit: a lane does not turn a decision into a record. The decisions log holds the
/// user's own decisions of this rewrite and no transcribed row, and the leader rules do not state
/// as decided what the log does not record.
#[test]
fn the_decisions_log_holds_no_transcribed_row() {
    let log = read("docs/decisions-2026-11-08.md");
    assert_lacks(&log, "the decisions log", &["흐름 재설계 결정 옮겨 적기"]);
    let rules = read("docs/leader-rules.md");
    assert_lacks(
        &rules,
        "docs/leader-rules.md",
        &["turned the fixed polling loops off"],
    );
    assert_has(&rules, "docs/leader-rules.md", &["card note 7366"]);
}

/// S6: tests are Rust only; the Python fixtures and the Python adapter left the file.
#[test]
fn stale_python_tests_are_not_a_rule_input() {
    let text = claude_md();
    assert_has(&text, "CLAUDE.md", &["Tests are Rust only."]);
    assert_lacks(
        &text,
        "CLAUDE.md",
        &[
            "tests/*_test.py",
            "adapters/claude/",
            "unittest discover` (about",
        ],
    );
}

/// REQ-6, REQ-7: the origin table and the long operating detail live in `docs/`; the file points
/// to all three and holds no provenance table.
#[test]
fn claude_md_points_to_the_docs_and_holds_no_provenance_table() {
    let text = claude_md();
    for doc in [
        "docs/lane-rules.md",
        "docs/leader-rules.md",
        "docs/claude-md-rule-provenance.md",
    ] {
        assert!(text.contains(doc), "CLAUDE.md does not point to {doc}");
        assert!(root().join(doc).is_file(), "{doc} is missing");
    }
    assert!(!text.contains("## Rule provenance"));
    assert!(!text.contains("| Rule in this file |"));
    assert!(!text.contains("## Leader supervision"));
    assert!(!text.contains("## Agent loop"));
}

/// D-6: no card number in the file, the three docs, the SPEC or the local draft.
#[test]
fn the_rewrite_names_no_card_number() {
    for path in [
        "CLAUDE.md".to_string(),
        "docs/lane-rules.md".to_string(),
        "docs/leader-rules.md".to_string(),
        "docs/claude-md-rule-provenance.md".to_string(),
        format!("{SPEC_DIR}/spec.md"),
        format!("{SPEC_DIR}/rule-table.md"),
        format!("{SPEC_DIR}/claude-local-draft.md"),
    ] {
        let hits = card_numbers(&read(&path));
        assert!(hits.is_empty(), "{path} names a card number: {hits:?}");
    }
}

/// AC-6 of the layout SPEC: the two personal lines of the old local file went to the ignored
/// `CLAUDE.local.md`, never into the tracked file, which every lane worktree and every clone gets.
#[test]
fn tracked_claude_md_holds_neither_personal_line() {
    let text = claude_md();
    for personal in [
        "- Memory: Claude auto-memory for this project",
        "- Dev Mac; GitHub auth through `gh` as the operator",
    ] {
        assert!(
            !text.contains(personal),
            "CLAUDE.md holds the personal line starting {personal:?}"
        );
    }
}

/// One row of `rule-table.md`: `(id, old lines, rule, disposition, destination or basis, anchor)`.
fn table_rows() -> Vec<[String; 6]> {
    let text = read(&format!("{SPEC_DIR}/rule-table.md"));
    let mut rows = Vec::new();
    for line in text.lines().filter(|l| l.starts_with("| R-")) {
        let inner = line
            .strip_prefix("| ")
            .and_then(|l| l.strip_suffix(" |"))
            .unwrap_or_else(|| panic!("a row without its outer bars: {line}"));
        let cells: Vec<String> = inner
            .replace("\\|", "\u{1}")
            .split(" | ")
            .map(|c| c.replace('\u{1}', "|"))
            .collect();
        let cells: [String; 6] = cells
            .try_into()
            .unwrap_or_else(|c: Vec<String>| panic!("{} cells in `{line}`", c.len()));
        rows.push(cells);
    }
    rows
}

/// REQ-2 to REQ-4 (AC-2): the table lists every rule with a disposition, the ids run 1..n, a kept
/// or moved row has its anchor in its destination, and a discarded row names its basis.
#[test]
fn rule_table_has_a_disposition_for_every_rule_and_every_anchor_is_found() {
    let rows = table_rows();
    assert!(rows.len() >= 100, "only {} rows", rows.len());
    let (mut kept, mut moved, mut discarded) = (0, 0, 0);
    for (i, [id, old, rule, disp, place, anchor]) in rows.iter().enumerate() {
        assert_eq!(id, &format!("R-{:02}", i + 1), "ids run 1..n");
        assert!(!old.is_empty() && !rule.is_empty(), "{id} lacks a field");
        match disp.trim_end_matches("-corrected") {
            "kept" | "moved" => {
                if disp.starts_with("kept") {
                    kept += 1;
                } else {
                    moved += 1;
                }
                let anchor = anchor
                    .strip_prefix('«')
                    .and_then(|a| a.strip_suffix('»'))
                    .unwrap_or_else(|| panic!("{id}: the anchor is not between « and »"));
                assert!(!anchor.is_empty(), "{id}: an empty anchor");
                let text = fs::read_to_string(root().join(place))
                    .unwrap_or_else(|e| panic!("{id}: destination {place}: {e}"));
                assert!(
                    text.contains(anchor),
                    "{id}: {place} does not contain «{anchor}»"
                );
            }
            "discarded" => {
                discarded += 1;
                assert!(anchor.is_empty(), "{id}: a discarded row has an anchor");
                assert!(
                    place.contains("decision")
                        || place.contains("SPEC-")
                        || place.contains('`')
                        || place.contains("history"),
                    "{id}: the basis names no decision or fact: {place}"
                );
            }
            other => panic!("{id}: disposition `{other}`"),
        }
    }
    assert!(
        kept > 0 && moved > 0 && discarded > 0,
        "{kept}/{moved}/{discarded}"
    );
}

/// REQ-8 (AC-5): the draft of the personal file holds this machine's facts and none of the team
/// rules; its anchors are the facts a lane cannot read from the repository.
#[test]
fn local_draft_holds_machine_facts_and_no_team_rule() {
    let draft = read(&format!("{SPEC_DIR}/claude-local-draft.md"));
    assert_has(
        &draft,
        "the local draft",
        &[
            "Claude Code 2.1.330",
            "`CLAUDE_CONFIG_DIR`",
            "`sccache`",
            "/tmp/moai-rc-cargo-<lane>",
            "`aside` and `ego-browser`",
            "MEMORY.md",
        ],
    );
    for [id, _, _, disp, place, anchor] in table_rows() {
        // A rule kept in `CLAUDE.md` must not also sit in the personal file.
        if disp.starts_with("kept") && place == "CLAUDE.md" {
            let anchor = anchor.trim_start_matches('«').trim_end_matches('»');
            assert!(
                !draft.contains(anchor),
                "{id}: the local draft repeats a team rule: {anchor}"
            );
        }
    }
    for team in ["## Authority", "## Standing principles", "## Flow:"] {
        assert!(!draft.contains(team), "the local draft holds `{team}`");
    }
}

/// REQ-10 (AC-6): the decisions log has the two decisions with their time.
#[test]
fn decisions_log_records_the_agents_md_removal_and_this_rewrite() {
    let log = read("docs/decisions-2026-11-08.md");
    for row in [
        "(11-22) | AGENTS.md 제거",
        "(11-23) | CLAUDE.md 다시 쓰기",
        "(11-23) | 통합 브랜치 NeXT",
    ] {
        assert!(
            log.contains(row),
            "the decisions log lacks a row with `{row}`"
        );
    }
}

/// REQ-12 (AC-7): the SPEC records what was measured about the load of the two files for a lane
/// under `.claude/worktrees/`, with the date, the version and the limit of the measurement.
#[test]
fn load_measurement_is_recorded_in_the_spec() {
    let spec = read(&format!("{SPEC_DIR}/spec.md"));
    assert_has(
        &spec,
        "the SPEC",
        &[
            "2026-11-22T01:18Z",
            "2.1.330",
            "겹쳐 읽지 않음",
            "`CLAUDE.local.md` 는 보였어요",
            "세션 시작 시점뿐",
        ],
    );
}
