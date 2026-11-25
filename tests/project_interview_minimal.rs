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
const LANGUAGE_TOOL_TOKENS: [&str; 8] = [
    "cargo",
    "rustup",
    "npm",
    "pipx",
    "go install",
    "pytest",
    "go test",
    "cargo-nextest",
];

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
/// The exception spares the command cell of the install table from the language tools only: that
/// cell is still read for the MoAI-internal tokens (REQ-23).
fn tool_lines(text: &str) -> Vec<(usize, &'static str)> {
    let mut found = Vec::new();
    let mut in_exception_table = false;
    for (at, line) in text.lines().enumerate() {
        if line.trim() == EXCEPTION_HEADER {
            in_exception_table = true;
        } else if !is_row(line) {
            in_exception_table = false;
        }
        let spared = scanned_part(line, in_exception_table).to_lowercase();
        let whole = line.to_lowercase();
        let lists: [(&[&'static str], bool, &str); 2] = [
            (&LANGUAGE_TOOL_TOKENS, false, &spared),
            (&INTERNAL_TOKENS, true, &whole),
        ];
        for (tokens, internal, lower) in lists {
            for token in tokens {
                let mut from = 0;
                while let Some(pos) = lower[from..].find(token) {
                    let start = from + pos;
                    let end = start + token.len();
                    if is_command(lower, start, end, internal) {
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

// Revision 3 of the SPEC (REQ-13 to REQ-25): the text repair of the project-side files. The
// expected sentences are pinned here; the texts are repaired to say them.

const NEXT_FLOW: &str = "skills/moai/references/flows/next-flow.md";
const WORK_FLOW_PLAN: &str = "src/ops/work_flow_plan.rs";
const SPEC: &str = ".moai/specs/SPEC-20261124-moai-project-interview/spec.md";

/// The hosts that no longer exist for this product; no project-side text names them.
const WITHDRAWN_HOSTS: [&str; 2] = ["Codex", "OMP"];

/// True when `word` stands in `text` as a whole word.
fn whole_word(text: &str, word: &str) -> bool {
    let alnum = |c: Option<char>| c.is_some_and(char::is_alphanumeric);
    let mut from = 0;
    while let Some(pos) = text[from..].find(word) {
        let start = from + pos;
        let end = start + word.len();
        if !alnum(text[..start].chars().next_back()) && !alnum(text[end..].chars().next()) {
            return true;
        }
        from = end;
    }
    false
}

/// The last cell of a table row, trimmed.
fn last_cell(row: &str) -> &str {
    row.trim_end()
        .trim_end_matches('|')
        .rsplit('|')
        .next()
        .unwrap_or("")
        .trim()
}

/// The text from `at` to the end of its sentence: the first full stop followed by a space or the
/// end of the text (a full stop inside a name such as `a.b` does not end it).
fn sentence_from(text: &str, at: usize) -> &str {
    let rest = &text[at..];
    let mut chars = rest.char_indices().peekable();
    while let Some((i, c)) = chars.next() {
        if c == '.' && chars.peek().is_none_or(|(_, next)| next.is_whitespace()) {
            return &rest[..=i];
        }
    }
    rest
}

/// A line is a numbered step of the PROJECT procedure: `2. **…` or `6a. **…`.
fn is_step_line(line: &str) -> bool {
    let digits = line.chars().take_while(char::is_ascii_digit).count();
    digits > 0
        && line[digits..]
            .trim_start_matches(|c: char| c.is_ascii_lowercase())
            .starts_with(". **")
}

/// The one boundary sentence of `project.md` (REQ-13): the route does not move earlier-generation
/// state, `/moai update` does, and when updates remain the route ends with one line naming it.
/// The full sentence the text carries is
/// "Moving state of an earlier MoAI generation is `/moai update`'s job: this route does not move
/// it, and when updates remain the route ends with one line that names `/moai update`."
const BOUNDARY: &str = "Moving state of an earlier MoAI generation is `/moai update`'s job";

/// AC-10 (REQ-13, REQ-14): `project.md` carries the boundary sentence once, before the first
/// numbered step that handles earlier-generation state, and the sentence also says the route ends
/// with one line naming `/moai update` when updates remain.
#[test]
fn boundary_sentence_is_pinned() {
    let project = read(PROJECT);
    assert_eq!(
        project.matches(BOUNDARY).count(),
        1,
        "`{PROJECT}` must carry exactly one boundary sentence starting `{BOUNDARY}`"
    );
    let at = project.find(BOUNDARY).unwrap();
    let sentence = sentence_from(&project, at);
    for needle in ["when updates remain", "one line"] {
        assert!(
            sentence.contains(needle),
            "the boundary sentence must say `{needle}`: {sentence}"
        );
    }
    assert!(
        sentence.matches("`/moai update`").count() >= 2,
        "the boundary sentence must name `/moai update` as the owner and in the closing line: {sentence}"
    );
    // The first numbered step that handles earlier-generation state: the sentence stands before it.
    let mut offset = 0;
    for line in project.split_inclusive('\n') {
        if is_step_line(line) && line.to_lowercase().contains("earlier-generation") {
            assert!(
                at < offset,
                "the boundary sentence must stand before the step `{}`",
                line.trim()
            );
            break;
        }
        offset += line.len();
    }
}

/// Lines of `text` that carry a card number (`C-59`), an Issue number (`#2639`, three to five
/// digits) or a dated user decision (`user decision 2026-11-16`, `user decision 16`), numbered
/// from 1 (REQ-15).
fn citation_lines(text: &str) -> Vec<(usize, String)> {
    let digits = |s: &str| s.chars().take_while(char::is_ascii_digit).count();
    let alnum = |c: Option<char>| c.is_some_and(char::is_alphanumeric);
    let card = |line: &str| {
        line.match_indices("C-").any(|(at, _)| {
            let rest = &line[at + 2..];
            let n = digits(rest);
            n > 0 && !alnum(line[..at].chars().next_back()) && !alnum(rest[n..].chars().next())
        })
    };
    let issue = |line: &str| {
        line.match_indices('#')
            .any(|(at, _)| (3..=5).contains(&digits(&line[at + 1..])))
    };
    let decision = |line: &str| {
        let lower = line.to_lowercase();
        lower
            .match_indices("user decision ")
            .any(|(at, found)| digits(&lower[at + found.len()..]) > 0)
    };
    text.lines()
        .enumerate()
        .filter(|(_, l)| card(l) || issue(l) || decision(l))
        .map(|(at, l)| (at + 1, l.to_string()))
        .collect()
}

/// AC-11 (REQ-15): the three project-side files carry no card number, no Issue number and no dated
/// user decision. `update.md` is another card's file and is not read here.
#[test]
fn no_citation_in_project_side_files() {
    let planted =
        "a card C-59 closes\nsee Issue #2639\n(user decision 2026-11-16)\nuser decision 16 = B";
    let flagged: Vec<usize> = citation_lines(planted).iter().map(|(n, _)| *n).collect();
    assert_eq!(
        flagged,
        vec![1, 2, 3, 4],
        "the guard must flag each planted shape"
    );
    for fine in [
        "SPEC-20261121-x and ABC-12 and C-x",
        "issue #12 and #123456",
        "the user decision, noted",
    ] {
        assert!(citation_lines(fine).is_empty(), "{fine} is not a citation");
    }
    for rel in [PROJECT, BANK, QUESTIONS] {
        if let Some((line, text)) = citation_lines(&read(rel)).first() {
            panic!("{rel}:{line} carries a card, Issue or dated decision: {text}");
        }
    }
}

/// True when `text` says that a lane pushes or opens a pull request.
fn says_lane_publishes(text: &str) -> bool {
    let lower = text.to_lowercase();
    [
        "pushes",
        "pushing",
        "opens a pull request",
        "open a pull request",
        "opens a pr",
    ]
    .iter()
    .any(|phrase| lower.contains(phrase))
}

/// AC-12 (REQ-16): the flow text names no withdrawn host, does not send a lane to push or open a
/// pull request, and says what `flows/next-flow` says: the lane only commits and the leader lands.
#[test]
fn flow_text_matches_next_flow() {
    let truth = read(NEXT_FLOW);
    assert!(
        truth.contains("A lane only commits"),
        "the flow description this test follows changed: {NEXT_FLOW}"
    );
    assert!(says_lane_publishes(
        "the lane pushes its branch and opens a pull request"
    ));
    assert!(!says_lane_publishes(
        "the lane only commits: no push, no pull request per change"
    ));
    for rel in [BANK, PROJECT] {
        let text = read(rel);
        for host in WITHDRAWN_HOSTS {
            assert!(
                !whole_word(&text, host),
                "{rel} names the withdrawn host {host}"
            );
        }
        for (at, line) in text.lines().enumerate() {
            assert!(
                !(line.to_lowercase().contains("next-flow") && says_lane_publishes(line)),
                "{rel}:{} says a lane pushes or opens a pull request: {line}",
                at + 1
            );
        }
    }
    let bank = read(BANK);
    let flow = bank
        .lines()
        .find(|l| l.starts_with("- **next-flow"))
        .expect("the next-flow bullet");
    for needle in ["the lane only commits", "the leader lands"] {
        assert!(
            flow.contains(needle),
            "the next-flow bullet must say `{needle}`: {flow}"
        );
    }
}

/// The order the interview gives the integration branch: the order `moai work-flow-plan` uses.
const BRANCH_ORDER: &str =
    "the profile value first, then the remote default branch, then the current branch";

/// AC-12 (REQ-17, REQ-24): the bank names the integration branch in the order of
/// `moai work-flow-plan`, and `project.md` shows the `work_flow_remote` warning of `project-plan`.
#[test]
fn integration_branch_order_matches_work_flow_plan() {
    let source = read(WORK_FLOW_PLAN);
    assert!(
        source.contains("else the remote's default branch (`origin/HEAD`), else the current"),
        "the order this test follows changed in {WORK_FLOW_PLAN}"
    );
    let bank = read(BANK);
    let flow = section(&bank, "Git work flow");
    assert!(
        flow.contains(BRANCH_ORDER),
        "the Git work flow section must say `{BRANCH_ORDER}`: {flow}"
    );
    let project = read(PROJECT);
    assert!(
        project.lines().any(|l| {
            l.contains("`work_flow_remote`")
                && l.contains("`project-plan`")
                && l.contains("warning")
                && l.to_lowercase().contains("show")
        }),
        "project.md must say to show the `work_flow_remote` warning of `project-plan`"
    );
}

/// The profile fields step 4 puts into the plan (REQ-18).
const PROFILE_FIELDS: [&str; 8] = [
    "purpose",
    "tools",
    "development methods",
    "hosts",
    "versioning",
    "decisions",
    "work_flow",
    "language",
];

/// AC-12 (REQ-18): the field list of step 4 holds every field the interview fills.
#[test]
fn step_four_lists_all_profile_fields() {
    let project = read(PROJECT);
    let step = project
        .lines()
        .find(|l| l.starts_with("4. **Plan and apply.**"))
        .expect("step 4");
    let rest = &step[step
        .find("`result.profile`")
        .expect("`result.profile` in step 4")..];
    let open = rest.find('(').expect("the field list of step 4");
    let close = open + rest[open..].find(')').expect("the end of the field list");
    let list = &rest[open + 1..close];
    for field in PROFILE_FIELDS {
        assert!(
            list.contains(field),
            "the step 4 field list must name `{field}`: ({list})"
        );
    }
}

/// AC-12 (REQ-19, REQ-25): the `method` row says the words the profile records (`TDD`,
/// `characterization`), and the `hosts` row offers `claude` and no withdrawn host.
#[test]
fn rows_use_profile_words() {
    let bank = read(BANK);
    let confirmed = section(&bank, "Confirmed");
    let method = row_line(&confirmed, "method");
    let goes_to = last_cell(&method);
    for word in ["TDD", "characterization"] {
        assert!(
            goes_to.contains(word),
            "the method row's Goes-to cell must name the recorded word `{word}`: {goes_to}"
        );
    }
    let hosts = row_line(&confirmed, "hosts");
    assert!(
        hosts.contains("claude"),
        "the hosts row must name `claude`: {hosts}"
    );
    for host in WITHDRAWN_HOSTS {
        assert!(
            !whole_word(&hosts, host),
            "the hosts row must not name the withdrawn host {host}: {hosts}"
        );
    }
}

/// Lines of `text` that hold the CalVer format `YYYY.0M.MICRO` with no plain example (a date such
/// as `2026.11.0`) within the next 60 characters, numbered from 1.
fn bare_calver_format_lines(text: &str) -> Vec<usize> {
    let example = |s: &str| {
        let b = s.as_bytes();
        (0..b.len()).any(|i| {
            let digit = |k: usize| b.get(i + k).is_some_and(u8::is_ascii_digit);
            let dot = |k: usize| b.get(i + k) == Some(&b'.');
            (0..4).all(digit) && dot(4) && digit(5) && digit(6) && dot(7) && digit(8)
        })
    };
    text.lines()
        .enumerate()
        .filter(|(_, line)| {
            line.match_indices("YYYY.0M.MICRO").any(|(at, found)| {
                let after: String = line[at + found.len()..].chars().take(60).collect();
                !example(&after)
            })
        })
        .map(|(at, _)| at + 1)
        .collect()
}

/// The sentences of the development phase that no user should read (REQ-20).
const LEFTOVER_SENTENCES: [&str; 3] = [
    "not_run until then",
    "is planned and not implemented in 2026.11.0-rc.1",
    "Host use of `/moai project` is verified only after",
];

/// AC-12 (REQ-20): the texts the user reads carry no head number, no bare version-format token, no
/// phrase about the session's own host and no development-leftover sentence.
#[test]
fn no_internal_word_in_user_shown_text() {
    assert_eq!(
        bare_calver_format_lines("| `YYYY.0M.MICRO` (recommended) |"),
        vec![1]
    );
    assert!(bare_calver_format_lines("`YYYY.0M.MICRO` (for example 2026.11.0)").is_empty());
    for rel in [BANK, PROJECT] {
        let text = read(rel);
        for phrase in ["HeadVer head", "this session's host"] {
            assert!(!text.contains(phrase), "{rel} says `{phrase}`");
        }
        let bare = bare_calver_format_lines(&text);
        assert!(
            bare.is_empty(),
            "{rel} shows `YYYY.0M.MICRO` with no plain example (lines): {bare:?}"
        );
        for sentence in LEFTOVER_SENTENCES {
            assert!(!text.contains(sentence), "{rel} says `{sentence}`");
        }
    }
}

/// AC-12 (REQ-21): the Defaults text does not send the user to edit the profile by hand, it says to
/// run `/moai project` again, and `project.md` no longer calls the receipt step unimplemented.
#[test]
fn contradictions_are_gone() {
    let bank = read(BANK);
    let defaults = section(&bank, "Defaults").to_lowercase();
    assert!(
        !defaults.contains("editing the profile") && !defaults.contains("edit the profile"),
        "the Defaults text must not tell the user to edit the profile by hand: {defaults}"
    );
    assert!(
        defaults.contains("`/moai project` again"),
        "the Defaults text must say to run `/moai project` again: {defaults}"
    );
    let project = read(PROJECT);
    assert!(
        !project.contains("planned and not implemented"),
        "project.md must not call the receipt step planned and not implemented"
    );
}

/// AC-13 (REQ-22): the scan lists `pytest`, `go test` and `cargo-nextest`, in a command position
/// and inside backticks, and does not flag a longer word.
#[test]
fn scan_tokens_cover_pytest_go_test_nextest() {
    for (planted, token) in [
        ("run pytest here", "pytest"),
        ("then `pytest` passes", "pytest"),
        ("run go test ./... here", "go test"),
        ("`go test` passes", "go test"),
        ("run cargo-nextest run", "cargo-nextest"),
        ("| `done` | `cargo-nextest` passes |", "cargo-nextest"),
    ] {
        assert_eq!(tool_lines(planted), vec![(1, token)], "{planted}");
    }
    assert!(
        tool_lines("a pytester or the gopher test is not the token").is_empty(),
        "a token inside a longer word is not flagged"
    );
}

/// AC-13 (REQ-23): the exempt install-table cell is still read for the MoAI-internal tokens; the
/// real cells hold none.
#[test]
fn internal_tokens_scanned_in_exempt_cell() {
    let table =
        |cell: &str| format!("{EXCEPTION_HEADER}\n| --- | --- |\n| `rust-analyzer` | `{cell}` |");
    for (cell, token) in [
        ("moai check affected", "moai check"),
        ("tools/local_check.sh", "local_check"),
        ("check-guard", "check-guard"),
    ] {
        assert_eq!(tool_lines(&table(cell)), vec![(3, token)], "{cell}");
    }
    assert!(
        tool_lines(&table("rustup component add rust-analyzer")).is_empty(),
        "a language-tool command in the exempt cell stays exempt"
    );
    assert!(
        tool_lines(&read(PROJECT)).is_empty(),
        "the real install table must pass the scan"
    );
}

/// AC-12 (SPEC D-13): the `## Result` section of the SPEC gives AC-1, AC-2, AC-4 and AC-5 each a
/// line of its own.
#[test]
fn result_lists_each_ac_on_its_own_line() {
    let result = section(&read(SPEC), "Result");
    for ac in ["AC-1", "AC-2", "AC-4", "AC-5"] {
        assert!(
            result.lines().any(|l| l.starts_with(&format!("- {ac}:"))),
            "the Result section must have a line starting `- {ac}:`"
        );
    }
}
