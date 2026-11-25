//! SPEC-20261125-base-agent-contract: the base agents hand implementation over inside the lane
//! worktree, `moai-audit` holds one tier table, `moai-docs` takes a `head`, no `reads:` entry
//! goes unused and the repair chain lists `moai-docs`. Each test pins sentences of the agent
//! files; the files are compiled in because the repository scan table counts a run-time read.

const IMPL: &str = include_str!("../agents/impl.md");
const AUDIT: &str = include_str!("../agents/audit.md");
const DOCS: &str = include_str!("../agents/docs.md");
const PLAN: &str = include_str!("../agents/plan.md");
const RESEARCH: &str = include_str!("../agents/research.md");
const CHAINS: &str = include_str!("../skills/moai/references/chains/index.md");
const CONTRACT: &str = include_str!("../skills/moai/references/audit.md");

/// The six texts the contract names, by file.
const FILES: [(&str, &str); 6] = [
    ("agents/impl.md", IMPL),
    ("agents/audit.md", AUDIT),
    ("agents/docs.md", DOCS),
    ("agents/plan.md", PLAN),
    ("agents/research.md", RESEARCH),
    ("skills/moai/references/chains/index.md", CHAINS),
];

/// The text of `## <name>` up to the next `## ` heading (empty when the section is missing).
fn section<'a>(text: &'a str, name: &str) -> &'a str {
    let head = format!("## {name}\n");
    let Some(start) = text.find(&head) else {
        return "";
    };
    let rest = &text[start + head.len()..];
    rest.split("\n## ").next().unwrap_or("")
}

/// The phrases of `phrases` that `text` does not hold.
fn lacking<'a>(text: &str, phrases: &[&'a str]) -> Vec<&'a str> {
    phrases
        .iter()
        .copied()
        .filter(|p| !text.contains(p))
        .collect()
}

/// The `reads:` entries of an agent file's front matter, in order.
fn reads(text: &str) -> Vec<String> {
    let front = text
        .strip_prefix("---\n")
        .and_then(|rest| rest.split_once("\n---\n"))
        .map(|(front, _)| front)
        .expect("an agent file starts with front matter");
    front
        .lines()
        .find_map(|l| l.strip_prefix("reads: "))
        .map(|v| v.split(", ").map(str::to_string).collect())
        .unwrap_or_default()
}

/// The body of an agent file, after its front matter.
fn body(text: &str) -> &str {
    text.split_once("\n---\n").expect("front matter").1
}

/// AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-13): the implementation agent takes the worktree, the
/// card and the generation, checks the folder and the base, asks `moai card check` before the
/// first write of a path, and leaves its change uncommitted in the parent's worktree.
#[test]
fn ac1_impl_takes_the_lane_inputs_checks_them_and_returns_an_uncommitted_change() {
    let input = section(IMPL, "Input");
    let missing = lacking(
        input,
        &[
            "- `worktree`: the absolute path of the folder your parent works in",
            "- `card`: ",
            "- `generation`: ",
            "- `base`: ",
            "the parent gives you only paths that `moai card check` allowed",
        ],
    );
    assert!(missing.is_empty(), "impl Input lacks {missing:?}");
    let steps = section(IMPL, "Steps");
    let missing = lacking(
        steps,
        &[
            "`git rev-parse --show-toplevel`",
            "`git rev-parse HEAD`",
            "end with `RESULT impl blocked` before you write",
            "`moai card check <card> --generation <generation> --path <path>`",
            "exit 0 is your permission to write that path",
            "end with `RESULT impl blocked` naming the answer as it is",
        ],
    );
    assert!(missing.is_empty(), "impl Steps lack {missing:?}");
    let missing = lacking(
        IMPL,
        &[
            "the change stays as uncommitted files in your parent's worktree",
            "`RESULT impl` names the HEAD the change sits on",
            "`git status --short`",
            "the parent compares them with `Files written` before it runs its own `moai card check` and commits",
        ],
    );
    assert!(missing.is_empty(), "agents/impl.md lacks {missing:?}");
}

/// The four rows of the tier table and the words each must hold, taken from the audit
/// contract's "Audit tiers" lines (the phrases are asserted against the contract as well).
const TIER_ROWS: [(&str, &[&str]); 4] = [
    (
        "| docs-only (tier 0) |",
        &[
            "only documents or SPEC text change",
            "nothing is built",
            "the auditor records a PASS without a receipt",
        ],
    ),
    (
        "| test-only (tier 0, tier 1 when a tool changes) |",
        &[
            "only tests or test tools change",
            "the auditor gives it a light audit",
            "reads the change and the lane's receipt and runs nothing",
        ],
    ),
    (
        "| code (tier 1) |",
        &[
            "product code or skill text changes",
            "a session that wrote none of it gives a light audit",
            "cites the lane's receipt instead of running the head again",
            "runs no mutant",
        ],
    ),
    (
        "| high-risk (tier 2) |",
        &[
            "a changed path starts with a prefix of the floor list",
            "a session that wrote none of it gives the full audit",
            "the code steps plus the boundary, security and data-loss checks of this contract",
            "the lane's guard-mutant lines read",
            "at most two named test modules run",
        ],
    ),
];

/// The tier table of `agents/audit.md`, line by line and exactly: the header, the separator and
/// the four rows. A row holds the contract's words and nothing else, so a clause added to a row
/// (an exception that lets the author audit, say) is not a row of the table.
const TIER_TABLE: [&str; 6] = [
    "| Tier | Changed paths | Runs | Cites | Who audits, and what the verdict then needs |",
    "| --- | --- | --- | --- | --- |",
    "| docs-only (tier 0) | only documents or SPEC text change | nothing is built | no receipt | the auditor records a PASS without a receipt |",
    "| test-only (tier 0, tier 1 when a tool changes) | only tests or test tools change | nothing | the lane's receipt | the auditor gives it a light audit: it reads the change and the lane's receipt and runs nothing |",
    "| code (tier 1) | product code or skill text changes | runs no mutant | cites the lane's receipt instead of running the head again | a session that wrote none of it gives a light audit |",
    "| high-risk (tier 2) | a changed path starts with a prefix of the floor list | the code steps plus the boundary, security and data-loss checks of this contract, and at most two named test modules run | the lane's guard-mutant lines read | a session that wrote none of it gives the full audit |",
];

/// The table row of `text` that starts with `start`.
fn row<'a>(text: &'a str, start: &str) -> Option<&'a str> {
    text.lines().find(|l| l.starts_with(start))
}

/// What is wrong with the tier table of `text`: a line that is not exactly the pinned one, a row
/// after the table, a row missing, or a row without a word the contract's line holds.
fn tier_table_problems(text: &str) -> Vec<String> {
    let mut problems = Vec::new();
    let lines: Vec<&str> = text.lines().collect();
    match lines.iter().position(|l| l.starts_with("| Tier |")) {
        None => problems.push("no table starts with the header".to_string()),
        Some(first) => {
            for (k, want) in TIER_TABLE.iter().enumerate() {
                if lines.get(first + k) != Some(want) {
                    problems.push(format!(
                        "table line {k} is not exactly {want:?}: {:?}",
                        lines.get(first + k)
                    ));
                }
            }
            if lines
                .get(first + TIER_TABLE.len())
                .is_some_and(|l| l.starts_with('|'))
            {
                problems.push("a row follows the table".to_string());
            }
        }
    }
    for (start, words) in TIER_ROWS {
        match row(text, start) {
            None => problems.push(format!("no row starts with {start:?}")),
            Some(line) => {
                for word in lacking(line, words) {
                    problems.push(format!("{start:?} lacks {word:?}"));
                }
                if line.matches(" | ").count() + 1 != 5 {
                    problems.push(format!("{start:?} does not hold five cells"));
                }
            }
        }
    }
    problems
}

/// AC-2 (REQ-5, REQ-6, REQ-7, REQ-14, REQ-15): one table with the four tier rows, no second
/// statement of a tier as a bullet, no "list below", no internal SPEC name, and a Verification
/// and a Handoff that ask for the checks the tier names.
#[test]
fn ac2_audit_holds_one_tier_table_and_no_contradiction() {
    let problems = tier_table_problems(AUDIT);
    assert!(problems.is_empty(), "{problems:#?}");
    let tiers = section(AUDIT, "Audit tiers");
    assert_eq!(
        tiers.lines().filter(|l| l.starts_with("| Tier |")).count(),
        1,
        "one table header"
    );
    for start in [
        "- docs-only (tier",
        "- test-only (tier",
        "- code (tier",
        "- high-risk (tier",
    ] {
        assert!(
            !AUDIT.lines().any(|l| l.starts_with(start)),
            "{start:?} is stated again as a bullet"
        );
    }
    assert!(
        AUDIT.contains("The floor list of paths is in the contract's section \"Audit tiers\""),
        "the floor list is pointed to by the contract's section name"
    );
    assert!(!AUDIT.contains("list below"), "a list below is named");
    assert!(!AUDIT.contains("SPEC-"), "an internal SPEC is named");
    let verification = section(AUDIT, "Verification");
    let handoff = section(AUDIT, "Handoff");
    assert!(
        verification.contains(
            "Every check the tier names was run, or is listed under `Not run` with the reason."
        ),
        "Verification asks for the checks the tier names"
    );
    assert!(
        !verification.contains("Every acceptance criterion has a check you ran"),
        "Verification still asks for a check on every criterion"
    );
    assert!(
        handoff.contains(
            "Return `ok` when every check the tier names ran and the findings are complete"
        ),
        "Handoff asks for the checks the tier names"
    );
    assert!(
        handoff.contains("a light audit that runs nothing returns `ok` with `Not run` naming what its tier leaves out"),
        "Handoff says what a light audit returns"
    );
    assert!(
        !handoff.contains("when every check ran"),
        "Handoff still asks for every check"
    );
}

/// AC-3 (REQ-5): the table's rows hold the same facts as the contract's "Audit tiers" lines:
/// every word asserted of a row is in the contract too, and a row that loses the high-risk
/// row's guard-mutant or named-module words is found (the guard mutants below).
#[test]
fn ac3_tier_rows_carry_the_contract_facts() {
    let contract = {
        let start = CONTRACT.find("## Audit tiers\n").expect("contract section");
        CONTRACT[start..].split("\n## ").next().unwrap_or("")
    };
    for (_, words) in TIER_ROWS {
        for word in words {
            assert!(
                contract.contains(word),
                "the contract's Audit tiers section lacks {word:?}"
            );
        }
    }
    let problems = tier_table_problems(AUDIT);
    assert!(problems.is_empty(), "{problems:#?}");
    // Guard mutants: the table without the row's words is a problem, and each mutant changes
    // the text.
    for (start, word) in [
        (
            "| high-risk (tier 2) |",
            "the lane's guard-mutant lines read",
        ),
        (
            "| high-risk (tier 2) |",
            "at most two named test modules run",
        ),
        ("| high-risk (tier 2) |", "data-loss"),
        ("| code (tier 1) |", "runs no mutant"),
    ] {
        let mutant = AUDIT
            .lines()
            .map(|l| {
                if l.starts_with(start) {
                    l.replacen(word, "", 1)
                } else {
                    l.to_string()
                }
            })
            .collect::<Vec<_>>()
            .join("\n");
        assert!(mutant != AUDIT, "dropping {word:?} did not change the text");
        assert!(
            !tier_table_problems(&mutant).is_empty(),
            "dropping {word:?} from {start:?} passes"
        );
    }
}

/// AC-3 guard mutant (the repair of the audit's FAIL): a clause that reverses a rule, added to the
/// end of any one row, is found, and so is a row added after the table. The first mutant is the
/// one the audit planted: the high-risk row ends with "or the author when no other session is
/// free".
#[test]
fn ac3_a_clause_added_to_a_row_is_found() {
    const CLAUSE: &str = ", or the author when no other session is free";
    let planted = AUDIT.replacen(
        "gives the full audit |",
        &format!("gives the full audit{CLAUSE} |"),
        1,
    );
    assert!(
        planted != AUDIT,
        "the planted clause did not change the text"
    );
    assert!(
        !tier_table_problems(&planted).is_empty(),
        "the high-risk row with the author clause passes"
    );
    for want in &TIER_TABLE[2..] {
        let longer = format!("{}{CLAUSE} |", want.trim_end_matches(" |"));
        let mutant = AUDIT.replacen(want, &longer, 1);
        assert!(mutant != AUDIT, "the clause did not change {want:?}");
        assert!(
            !tier_table_problems(&mutant).is_empty(),
            "a row with an added clause passes: {want}"
        );
    }
    let extra = AUDIT.replacen(
        TIER_TABLE[5],
        &format!("{}\n| extra (tier 9) | a | b | c | d |", TIER_TABLE[5]),
        1,
    );
    assert!(extra != AUDIT, "the extra row did not change the text");
    assert!(
        !tier_table_problems(&extra).is_empty(),
        "a row added after the table passes"
    );
}

/// AC-4 (REQ-8, REQ-16): the documentation agent takes `head` and ends `blocked` when HEAD
/// differs.
#[test]
fn ac4_docs_takes_a_head_and_blocks_on_a_different_head() {
    let missing = lacking(
        section(DOCS, "Input"),
        &[
            "- `base`: the 40-hex commit the change builds on, when given.",
            "- `head`: the 40-hex commit that holds the change.",
        ],
    );
    assert!(missing.is_empty(), "docs Input lacks {missing:?}");
    let missing = lacking(
        section(DOCS, "Steps"),
        &[
            "`git rev-parse HEAD`",
            "end with `RESULT docs blocked` before you write",
            "`base..head`",
        ],
    );
    assert!(missing.is_empty(), "docs Steps lack {missing:?}");
}

/// AC-5 (REQ-9): the `reads:` lists are the ones D-4 chose, and each entry is named by the
/// agent's body (an entry the body does not name is read for nothing).
#[test]
fn ac5_every_reads_entry_is_named_in_the_body() {
    let want: [(&str, &str, &[&str]); 5] = [
        (
            "agents/plan.md",
            PLAN,
            &[
                "internal/core-rules",
                "internal/decide",
                "internal/code-atlas",
            ],
        ),
        (
            "agents/research.md",
            RESEARCH,
            &["internal/core-rules", "internal/code-atlas"],
        ),
        (
            "agents/impl.md",
            IMPL,
            &[
                "internal/core-rules",
                "internal/test-design",
                "internal/git-safety",
                "internal/tdd-results",
                "internal/self-check",
            ],
        ),
        (
            "agents/audit.md",
            AUDIT,
            &[
                "internal/core-rules",
                "internal/test-design",
                "internal/check-completeness",
                "internal/boundary-checks",
                "internal/security-checklist",
            ],
        ),
        (
            "agents/docs.md",
            DOCS,
            &[
                "internal/core-rules",
                "internal/decide",
                "internal/doc-claims",
                "internal/sdd",
            ],
        ),
    ];
    for (name, text, expected) in want {
        assert_eq!(reads(text), expected, "{name}: reads list");
        for entry in reads(text) {
            assert!(
                body(text).contains(&format!("`{entry}`")),
                "{name}: the body does not name `{entry}`"
            );
        }
    }
}

/// AC-6 (REQ-10): the repair row lists the three agents and the three routes.
#[test]
fn ac6_the_repair_row_lists_impl_docs_and_audit() {
    let line = CHAINS
        .lines()
        .find(|l| l.starts_with("| Repair after a FAIL verdict |"))
        .expect("the repair row");
    let cells: Vec<&str> = line.trim_matches('|').split('|').map(str::trim).collect();
    assert_eq!(cells[1], "RUN, SYNC, AUDIT");
    assert_eq!(cells[2], "moai-impl, moai-docs, moai-audit");
}

/// The first place in `text` that holds `#` followed by four digits, or a card number.
fn number_leftover(text: &str) -> Option<String> {
    let bytes: Vec<char> = text.chars().collect();
    for (i, c) in bytes.iter().enumerate() {
        let after = |n: usize| bytes.get(i + n).copied().unwrap_or(' ');
        let before = if i == 0 { ' ' } else { bytes[i - 1] };
        if *c == '#' && (1..=4).all(|n| after(n).is_ascii_digit()) {
            return Some(bytes[i..bytes.len().min(i + 5)].iter().collect());
        }
        if *c == 'C' && after(1) == '-' && after(2).is_ascii_digit() && !before.is_alphanumeric() {
            return Some(bytes[i..bytes.len().min(i + 5)].iter().collect());
        }
    }
    None
}

/// AC-7 (REQ-11): no Issue, card or internal SPEC name and no retired host in the six files.
#[test]
fn ac7_no_issue_card_spec_or_retired_host_name() {
    for (name, text) in FILES {
        assert!(!text.contains("Issue #"), "{name}: an Issue number");
        if let Some(found) = number_leftover(text) {
            panic!("{name}: a number name {found:?}");
        }
        for word in ["SPEC-2", "Codex", "OMP", "Kiro", "Cursor"] {
            assert!(!text.contains(word), "{name}: {word:?}");
        }
    }
    // The scan reads what it says.
    assert!(number_leftover("see #2637 here").is_some());
    assert!(number_leftover("card C-526 here").is_some());
    assert!(number_leftover("the C-level and #name").is_none());
}
