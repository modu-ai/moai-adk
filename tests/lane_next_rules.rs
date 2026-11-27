//! #2831 AC-4 (D3): the agent workflow sends a lane to `lane-next` when a task ends, says it
//! reads GitHub afresh, lets a reversible decision go on with the safe default, and names
//! only flags the command really has.

use std::{fs, path::Path};

fn work() -> String {
    fs::read_to_string(Path::new(env!("CARGO_MANIFEST_DIR")).join("skills/moai/workflows/work.md"))
        .unwrap()
}

/// Re-pointed to WORK with the retired MODE-AGENT route (SPEC-20261122-next-flow-skills-agents
/// REQ-3): the next card comes from the ledger, read afresh each time; the name stays.
#[test]
fn the_agent_loop_runs_lane_next_and_goes_on_with_reversible_defaults() {
    let text = work();
    // Card C-53 S5: the next task comes from the ledger (`moai card next`); `lane-next` reads
    // Issue claims, so the loop no longer names it. WORK opens no pull request, so the pull
    // request, draft, decision-applied and tip-repair phrases left with MODE-AGENT.
    for phrase in [
        "Read the state afresh each time (`moai card next`) — never what you remember",
        "a restart, a usage-limit stop or a compacted conversation misses nothing",
    ] {
        assert!(text.contains(phrase), "work.md must say `{phrase}`");
    }
    assert!(!text.contains("lane-next"), "work.md names `lane-next`");
}

/// #2880 W4 (AC-3), retired by the local card ledger (user decision 2026-11-16) and
/// `SPEC-20261123-claude-md-rewrite` (rule-table row R-03): the project rules no longer say that
/// the lane's PLAN comment on an Issue is part of the design; the SPEC the card names is. The
/// provenance table keeps the origin of the retired rule.
#[test]
fn the_retired_plan_comment_rule_keeps_its_provenance_row() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let rules = fs::read_to_string(root.join("CLAUDE.md")).unwrap();
    assert!(!rules.contains("PLAN #<issue>"));
    let provenance = fs::read_to_string(root.join("docs/claude-md-rule-provenance.md")).unwrap();
    assert!(provenance.contains(
        "| Retired: the lane's `PLAN #<issue>` comment is part of the design (Authority 2 of the earlier file); the card and its SPEC replace it | Issue #2880;"
    ));
}

/// The leader condition of the next-card rule, as `docs/lane-rules.md` says it.
const LEADER_CLAUSE: &str = ", and only when no leader is registered; with a registered leader the lane waits for that leader's dispatch.";

/// What the next-card bullet of the lane rules lacks of its leader condition: with a registered
/// leader a lane waits for that leader's dispatch, and only without one does it take the next
/// card itself.
fn next_card_leader_problems(rules: &str) -> Vec<String> {
    let Some(bullet) = rules
        .lines()
        .find(|l| l.starts_with("- A READY note or a `LANE-DONE` line hands a card back"))
    else {
        return vec!["docs/lane-rules.md has no next-card bullet".to_string()];
    };
    [
        "only when no leader is registered",
        "with a registered leader the lane waits for that leader's dispatch",
    ]
    .iter()
    .filter(|part| !bullet.contains(**part))
    .map(|part| format!("the next-card bullet lacks {part:?}: {bullet}"))
    .collect()
}

/// SPEC-20261123-doc-sentence-pins AC-1 (REQ-1): the lane rules keep both halves of the leader
/// condition of the next-card rule.
#[test]
fn next_card_leader_clause_is_in_the_lane_rules() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let rules = fs::read_to_string(root.join("docs/lane-rules.md")).unwrap();
    assert!(
        rules.contains(LEADER_CLAUSE),
        "the leader clause moved or changed"
    );
    let problems = next_card_leader_problems(&rules);
    assert!(problems.is_empty(), "{problems:?}");
}

/// SPEC-20261123-doc-sentence-pins AC-1 (REQ-2): the same text without the clause, or with one
/// half of it, or without the bullet, is a problem; each mutant really changes the text.
#[test]
fn next_card_leader_clause_mutants_are_caught() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let rules = fs::read_to_string(root.join("docs/lane-rules.md")).unwrap();
    let wait = "; with a registered leader the lane waits for that leader's dispatch";
    let mutants = [
        ("clause removed", rules.replace(LEADER_CLAUSE, ".")),
        ("wait half removed", rules.replace(wait, "")),
        (
            "condition half removed",
            rules.replace(", and only when no leader is registered", ""),
        ),
        (
            "bullet removed",
            rules
                .lines()
                .filter(|l| {
                    !l.starts_with("- A READY note or a `LANE-DONE` line hands a card back")
                })
                .collect::<Vec<_>>()
                .join("\n"),
        ),
    ];
    for (name, mutant) in mutants {
        assert!(
            mutant != rules,
            "{name}: the mutant did not change the text"
        );
        assert!(
            !next_card_leader_problems(&mutant).is_empty(),
            "{name}: the pin passes without the clause"
        );
    }
}

/// The principles of the command forms the host worktree guard, the boundary hook and the
/// SPEC-branch preflight accept, in the order `work.md` states them: one paragraph, each a
/// sentence of reasons and rules, never a list of refused words.
const COMMAND_FORM_SENTENCES: [&str; 7] = [
    "Write each shell call in a form the host and the hooks can read before it runs.",
    "One plain command per call: one `git` command with its arguments spelled out, several in separate calls in order, never `git` inside a loop, a chain of commands, a shell function, a substitution or `eval`.",
    "Write every path and branch name as literal text, not as a variable, a glob or a brace pattern.",
    "Put SQL and any long or multi-line text in a file made with the host's file tool and give the command that file's path.",
    "On a SPEC branch write the SPEC file before any other file, because the preflight refuses the rest until the SPEC exists.",
    "Search file contents with the shell's `grep`, not with a tool the host does not list.",
    "When a guard or hook refuses a call, read `internal/shell-guards` before the next try: it separates a refusal of the form, which one plain split command answers, from a refusal of the effect, which is reported as it stands and never redone through another command, tool or path.",
];

/// The bullets of the `Command forms` section of `docs/lane-rules.md`: the repository's reason
/// behind each principle.
const COMMAND_FORM_BULLETS: [&str; 6] = [
    "- One plain command per call.",
    "- Literal paths and branch names.",
    "- Files for long text.",
    "- SPEC file first.",
    "- Shell search.",
    "- A refusal: `skills/moai/references/internal/shell-guards.md` separates",
];

/// The text from the heading `heading` up to the next `## ` heading.
fn section_of<'a>(text: &'a str, heading: &str) -> &'a str {
    let start = text
        .find(&format!("\n{heading}\n"))
        .unwrap_or_else(|| panic!("no section {heading}"));
    let body = &text[start + 1..];
    let end = body[heading.len()..]
        .find("\n## ")
        .map_or(body.len(), |at| at + heading.len());
    &body[..end]
}

/// C-606: the lane reads, in the skill text it is given, the command forms that the host
/// worktree guard, the boundary hook and the SPEC-branch preflight accept, as principles with
/// the reason for each; the sentences stand in one paragraph of `work.md`, in this order, after
/// the paragraph about hook settings and before the closing paragraph about lane commands.
#[test]
fn work_md_teaches_the_command_forms_the_guards_accept() {
    let text = work();
    let mut from = 0;
    for sentence in COMMAND_FORM_SENTENCES {
        let at = text[from..]
            .find(sentence)
            .unwrap_or_else(|| panic!("work.md lacks, in order, `{sentence}`"));
        from += at + sentence.len();
    }
    let first = text.find(COMMAND_FORM_SENTENCES[0]).unwrap();
    let paragraph_start = text[..first].rfind("\n\n").map_or(0, |at| at + 2);
    let paragraph_end = text[first..]
        .find("\n\n")
        .map_or(text.len(), |at| first + at);
    let paragraph = &text[paragraph_start..paragraph_end];
    assert!(
        paragraph.starts_with(COMMAND_FORM_SENTENCES[0])
            && paragraph.ends_with(COMMAND_FORM_SENTENCES[6]),
        "the command forms are one paragraph that opens and ends with its sentences: {paragraph}"
    );
    let hooks = text
        .find(
            "Never edit the project's or the user's `.claude/settings.json` to quiet the warning.",
        )
        .expect("work.md has the hook settings paragraph");
    let tail = text
        .find("\nWhere the project offers lane commands")
        .expect("work.md has the lane commands paragraph");
    assert!(
        hooks < first && first < tail,
        "the paragraph is out of place: {hooks} {first} {tail}"
    );
}

/// C-606: `docs/lane-rules.md` keeps the repository's reason behind each principle in a short
/// section of its own: the six bullets in order, and no more lines than the principles need.
#[test]
fn lane_rules_name_the_reason_behind_each_command_form() {
    let root = Path::new(env!("CARGO_MANIFEST_DIR"));
    let rules = fs::read_to_string(root.join("docs/lane-rules.md")).unwrap();
    let section = section_of(&rules, "## Command forms");
    let mut from = 0;
    for bullet in COMMAND_FORM_BULLETS {
        let at = section[from..]
            .find(bullet)
            .unwrap_or_else(|| panic!("the Command forms section lacks, in order, `{bullet}`"));
        from += at + bullet.len();
    }
    assert!(
        section.lines().filter(|l| !l.is_empty()).count() <= 10,
        "the Command forms section grew past ten lines:\n{section}"
    );
    // The refusal bullet and the last sentence of `work.md` point at the reference that holds the
    // rule; the pointer must not dangle, and the rule is the one that never redoes a refused
    // effect through another path.
    let guards = fs::read_to_string(root.join("skills/moai/references/internal/shell-guards.md"))
        .expect("the reference the refusal bullet names");
    assert!(
        guards.contains("never reach the same effect by another path"),
        "shell-guards.md lost the rule the lane texts point at"
    );
    for guard in ["worktree guard", "boundary hook", "SPEC-branch preflight"] {
        assert!(
            section.contains(guard),
            "the section does not name the {guard}"
        );
    }
}

// Retired with the MODE-AGENT route (SPEC-20261122-next-flow-skills-agents REQ-2): `the_flags_the_workflow_names_are_the_commands_own` pinned the `lane-next` and `moai wait` flags MODE-AGENT showed, and WORK names neither.
