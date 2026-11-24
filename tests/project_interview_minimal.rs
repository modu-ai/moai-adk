//! SPEC-20261124-moai-project-interview: the `/moai project` interview asks only what detection
//! cannot settle (the purpose and four languages), shows every detected value for confirmation,
//! and records the four operating rules as generic defaults without asking. The tests read the
//! question bank, the PROJECT workflow, the UPDATE workflow and the question rules as text; no
//! command, host or HOME runs here.

use std::fs;

fn read(rel: &str) -> String {
    fs::read_to_string(format!("{}/{rel}", env!("CARGO_MANIFEST_DIR")))
        .unwrap_or_else(|e| panic!("{rel}: {e}"))
}

const BANK: &str = "skills/moai/references/internal/interview.md";
const PROJECT: &str = "skills/moai/workflows/project.md";
const UPDATE: &str = "skills/moai/workflows/update.md";
const QUESTIONS: &str = "skills/moai/references/internal/questions.md";

/// The text of the `## <title>` section of `text`, up to the next `## ` heading.
fn section(text: &str, title: &str) -> String {
    let head = format!("## {title}");
    let start = text
        .lines()
        .position(|l| l.starts_with(&head))
        .unwrap_or_else(|| panic!("no `{head}` section"));
    let mut body = Vec::new();
    for line in text.lines().skip(start + 1) {
        if line.starts_with("## ") {
            break;
        }
        body.push(line);
    }
    body.join("\n")
}

/// The ids of the table rows `| `id` | …` of a section text, in order.
fn row_ids(section: &str) -> Vec<String> {
    section
        .lines()
        .filter_map(|l| l.strip_prefix("| `"))
        .filter_map(|rest| rest.split('`').next())
        .map(str::to_string)
        .collect()
}

/// The whole line of the table row of `id` in a section text.
fn row_line(section: &str, id: &str) -> String {
    section
        .lines()
        .find(|l| l.starts_with(&format!("| `{id}` |")))
        .unwrap_or_else(|| panic!("no `{id}` row"))
        .to_string()
}

const OPERATING_RULES: [&str; 4] = ["protected", "review_blocks", "autonomy", "done"];
const LANGUAGE_ITEMS: [&str; 4] = [
    "user_chat",
    "documents",
    "commit_messages",
    "instruction_files",
];

/// AC-1 (REQ-1): the asked set is the purpose, and the four languages are named as the other
/// asked items; every other row is confirmed or defaulted, and no operating rule is asked.
#[test]
fn the_bank_asks_only_the_purpose_and_four_languages() {
    let bank = read(BANK);
    let asked = section(&bank, "Asked");
    assert_eq!(row_ids(&asked), vec!["purpose"], "the asked table: {asked}");
    for item in LANGUAGE_ITEMS {
        assert!(
            asked.contains(&format!("`{item}`")),
            "the asked section must name the language item `{item}`"
        );
    }
    assert!(
        asked.contains("moai language-setup"),
        "the languages are asked through `moai language-setup`"
    );
    let confirmed = section(&bank, "Confirmed");
    let ids = row_ids(&confirmed);
    for id in [
        "stack",
        "test_command",
        "build_command",
        "method",
        "hosts",
        "versioning",
        "work_flow",
    ] {
        assert!(
            ids.contains(&id.to_string()),
            "confirmed rows {ids:?} lack `{id}`"
        );
    }
    for rule in OPERATING_RULES {
        assert!(
            !row_ids(&asked).contains(&rule.to_string()) && !ids.contains(&rule.to_string()),
            "`{rule}` must not be an asked or confirmed row"
        );
    }
    let project = read(PROJECT);
    assert!(
        project.contains("at most two calls"),
        "the PROJECT text must cap the interview at two calls"
    );
}

/// AC-2 (REQ-2): the purpose is typed by the user as plain text, with no draft.
#[test]
fn the_purpose_is_typed_with_no_draft() {
    let bank = read(BANK);
    let line = row_line(&section(&bank, "Asked"), "purpose");
    assert!(
        line.contains("no draft"),
        "the purpose row must say `no draft`: {line}"
    );
    assert!(
        !line.contains("drafts written from"),
        "the purpose row must not offer drafts: {line}"
    );
    let project = read(PROJECT);
    assert!(
        project.contains("no draft"),
        "the PROJECT text must say the purpose has no draft"
    );
    let questions = read(QUESTIONS);
    assert!(
        !questions.contains("offer drafts written from what was read"),
        "the question rules must not tell the session to offer purpose drafts"
    );
    assert!(
        questions.contains("the purpose") && questions.contains("plain text"),
        "the question rules must name the purpose as the one question asked as plain text"
    );
}

/// AC-4 (REQ-5, REQ-6): detected and profile values are shown in one confirmation; with none, the
/// generic defaults GitHub flow, SemVer and tests first are shown as defaults.
#[test]
fn detected_values_are_confirmed_and_defaults_are_generic() {
    let bank = read(BANK);
    let confirmed = section(&bank, "Confirmed");
    for needle in [
        "one confirmation",
        "a value the profile already holds",
        "GitHub flow",
        "SemVer",
        "tests first",
    ] {
        assert!(
            confirmed.contains(needle),
            "the confirmed section must say `{needle}`"
        );
    }
    let hosts = row_line(&confirmed, "hosts");
    for retired in ["Codex", "OMP"] {
        assert!(
            !hosts.contains(retired),
            "the hosts row must not name the retired host {retired}: {hosts}"
        );
    }
    let project = read(PROJECT);
    assert!(
        project.contains("one confirmation"),
        "the PROJECT text must carry the one confirmation"
    );
}

/// AC-5 (REQ-7): the four operating rules are never asked and each has one generic default.
#[test]
fn operating_rules_are_never_asked_and_have_defaults() {
    let bank = read(BANK);
    let defaults = section(&bank, "Defaults");
    assert_eq!(
        row_ids(&defaults),
        OPERATING_RULES,
        "the defaults table lists the four operating rules in order"
    );
    for (id, needles) in [
        ("protected", vec!["secret and key files"]),
        (
            "review_blocks",
            vec![
                "a secret in code or logs",
                "a change without a test",
                "a change outside the task's scope",
            ],
        ),
        (
            "autonomy",
            vec!["hard to undo", "money, secrets or deletion"],
        ),
        ("done", vec!["the project's own tests pass"]),
    ] {
        let line = row_line(&defaults, id);
        for needle in needles {
            assert!(
                line.contains(needle),
                "the `{id}` default must say `{needle}`: {line}"
            );
        }
    }
    assert!(
        defaults.contains("\"source\": \"default\""),
        "the defaults section must say they are recorded with source default"
    );
    let project = read(PROJECT);
    assert!(
        project.contains("never asked"),
        "the PROJECT text must say the operating rules are never asked"
    );
}

/// Tokens that tie a question, an option or an example to one language's tooling (REQ-9).
const LANGUAGE_TOOL_TOKENS: [&str; 5] = ["cargo", "rustup", "npm", "pipx", "go install"];

/// Tokens that name this repository's own commands and scripts (REQ-9). They are file or command
/// names, so any boundary after them counts (`tools/local_check.sh`, `moai check affected`).
const INTERNAL_TOKENS: [&str; 3] = ["moai check", "local_check", "check-guard"];

/// The header row of the one named exception (SPEC D-6): the language-server install table of
/// step 6a. Only the last cell of its rows, the install command, is exempt; the other cells and
/// every other line are scanned.
const EXCEPTION_HEADER: &str = "| Language server | Install route |";

/// A line `text` is a table row when its first non-space character is a bar.
fn is_row(line: &str) -> bool {
    line.trim_start().starts_with('|')
}

/// The part of `line` that the scan reads: the whole line, or for a row of the exception table
/// every cell but the last one.
fn scanned_part(line: &str, in_exception_table: bool) -> String {
    if !(in_exception_table && is_row(line)) {
        return line.to_string();
    }
    let trimmed = line.trim_end().trim_end_matches('|');
    match trimmed.rfind('|') {
        Some(at) => trimmed[..at].to_string(),
        None => String::new(),
    }
}

/// True when `token` stands in `lower` at `start..end` as a command or a name, not inside a longer
/// word and not as the stem of a file name such as `Cargo.toml`.
fn is_command(lower: &str, start: usize, end: usize, internal: bool) -> bool {
    let word = |c: Option<char>| c.is_some_and(|c| c.is_alphanumeric());
    let before = lower[..start].chars().next_back();
    let mut after = lower[end..].chars();
    let next = after.next();
    if word(before) {
        return false;
    }
    if internal {
        // `local_check.sh`, `moai check affected`, `check-guard`: any boundary ends the name.
        return !word(next) && next != Some('_');
    }
    // A language tool ends at a space, a quote, a closing mark, the end of the line, or a full
    // stop that closes a sentence (`the user runs cargo.`); `Cargo.toml` (a full stop followed by
    // a letter) is a file name, not a command.
    match next {
        None => true,
        Some(c) if c.is_whitespace() || "`,;:)\"'!?/".contains(c) => true,
        Some('.') => !word(after.next()),
        Some(_) => false,
    }
}

/// Lines of `text` that hold a token of the two lists as a whole word, outside the exception.
fn tool_lines(text: &str) -> Vec<(usize, &'static str)> {
    let mut found = Vec::new();
    let mut in_exception_table = false;
    for (at, line) in text.lines().enumerate() {
        if line.trim() == EXCEPTION_HEADER {
            in_exception_table = true;
        } else if !is_row(line) {
            in_exception_table = false;
        }
        let lower = scanned_part(line, in_exception_table).to_lowercase();
        let lists: [(&[&'static str], bool); 2] =
            [(&LANGUAGE_TOOL_TOKENS, false), (&INTERNAL_TOKENS, true)];
        for (tokens, internal) in lists {
            for token in tokens {
                let mut from = 0;
                while let Some(pos) = lower[from..].find(token) {
                    let start = from + pos;
                    let end = start + token.len();
                    if is_command(&lower, start, end, internal) {
                        found.push((at + 1, *token));
                        break;
                    }
                    from = end;
                }
            }
        }
    }
    found
}

/// The files the scan reads: the question bank, the two workflows that ask, and the question
/// rules (N1: the rules carry examples too).
const SCANNED: [&str; 4] = [BANK, PROJECT, UPDATE, QUESTIONS];

/// AC-7 (REQ-10): the scan list names the question rules as well as the bank and the workflows.
#[test]
fn the_scan_reads_the_question_rules_too() {
    for rel in [BANK, PROJECT, UPDATE, QUESTIONS] {
        assert!(SCANNED.contains(&rel), "the scan must read {rel}");
    }
}

/// AC-7 (REQ-9, REQ-10): the shipped interview texts name no language tool and no MoAI-internal
/// command outside the one named exception.
#[test]
fn no_question_option_or_example_names_a_language_tool() {
    for rel in SCANNED {
        let found = tool_lines(&read(rel));
        assert!(
            found.is_empty(),
            "{rel} names a tool (line, token): {found:?}"
        );
    }
}

/// AC-7 (REQ-10): the exception is one table of `project.md`, named by its header row, and the
/// header stands once; a second table would widen the exception unseen.
#[test]
fn the_named_exception_is_one_install_table() {
    let project = read(PROJECT);
    let headers = project
        .lines()
        .filter(|l| l.trim() == EXCEPTION_HEADER)
        .count();
    assert_eq!(headers, 1, "the install table header must stand once");
    for rel in [BANK, UPDATE, QUESTIONS] {
        assert!(
            !read(rel).lines().any(|l| l.trim() == EXCEPTION_HEADER),
            "{rel} must not carry the exception table"
        );
    }
}

/// AC-7 (REQ-10, guard): the scan flags a planted tool in an option, in a sentence end, in a script
/// path and beside the exception table, and spares only the install cell of that table.
#[test]
fn the_tool_scan_flags_a_planted_tool_and_spares_only_the_named_exception() {
    let planted =
        "| `done` | always | Which check? | `cargo test` passes (recommended) | decisions |";
    assert_eq!(tool_lines(planted), vec![(1, "cargo")]);
    // The shapes the first scan missed: a token before a full stop, and a script path.
    assert_eq!(
        tool_lines("In that case the user runs cargo."),
        vec![(1, "cargo")]
    );
    assert_eq!(
        tool_lines(
            "| `done` | always | Which check? | `tools/local_check.sh` passes | decisions |"
        ),
        vec![(1, "local_check")]
    );
    assert_eq!(
        tool_lines("run tools/local_check.sh and read it"),
        vec![(1, "local_check")]
    );
    assert_eq!(
        tool_lines("then `moai check affected` runs"),
        vec![(1, "moai check")]
    );
    assert_eq!(tool_lines("an npm. install"), vec![(1, "npm")]);
    // The exception table: the install cell is spared, the other cells and the lines around it
    // are not.
    let table = format!(
        "{EXCEPTION_HEADER}\n| --- | --- |\n| `rust-analyzer` | `rustup component add rust-analyzer` |\n| `cargo` | `npm install -g x` |\nThe user runs cargo."
    );
    assert_eq!(tool_lines(&table), vec![(4, "cargo"), (5, "cargo")]);
    let beside = format!("{table}\n\n| other | `npm install -g y` |");
    assert_eq!(
        tool_lines(&beside).last(),
        Some(&(7, "npm")),
        "a table row after the exception table is scanned in full"
    );
    assert!(
        tool_lines("a word like pnpm or cargoes is not the token").is_empty(),
        "a token inside a longer word is not flagged"
    );
    assert!(
        tool_lines("the version in the root `Cargo.toml` and `package.json`").is_empty(),
        "a manifest file name is a detection fact, not a command"
    );
}
