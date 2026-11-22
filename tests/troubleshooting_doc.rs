//! SPEC-20261122-troubleshooting-doc: `docs/troubleshooting.md` holds the statements a user
//! needs when a session misbehaves (REQ-1 to REQ-7), and a missing one is named (REQ-8). The
//! page is read from the tree; nothing is written.

use std::path::PathBuf;

const PAGE: &str = "docs/troubleshooting.md";

/// Each statement of the page, by the words every spelling of it must hold.
const STATEMENTS: [(&str, &[&str]); 7] = [
    (
        "safe mode first (REQ-1)",
        &["claude --safe-mode", "skills", "agents", "rules", "hooks"],
    ),
    (
        "a clean configuration folder, never HOME (REQ-2)",
        &["CLAUDE_CONFIG_DIR", "--setting-sources user", "`HOME`"],
    ),
    (
        "a debug file and the verbose level (REQ-3)",
        &[
            "--debug-file",
            "CLAUDE_CODE_DEBUG_LOG_LEVEL=verbose",
            "finished with status",
        ],
    ),
    (
        "clean a debug file before sharing it (REQ-4)",
        &["before sharing", "keys", "tokens", "paths"],
    ),
    (
        "the classifier notice behind a gateway (REQ-5)",
        &[
            "auto mode",
            "gateway",
            "CLAUDE_CODE_AUTO_MODE_SERVER=0",
            "temporary",
        ],
    ),
    (
        "the commands that show what loaded (REQ-6)",
        &[
            "/context",
            "/memory",
            "/skills",
            "/hooks",
            "/permissions",
            "/status",
            "/doctor",
            "claude doctor",
        ],
    ),
    (
        "where moai does not reach (REQ-7)",
        &["cloud threads", "self-hosted runners", "Linux desktop"],
    ),
];

fn page() -> String {
    let path = PathBuf::from(env!("CARGO_MANIFEST_DIR")).join(PAGE);
    std::fs::read_to_string(&path).unwrap_or_else(|e| panic!("{PAGE} cannot be read: {e}"))
}

/// The names of the statements whose words are not all on the page.
fn missing(text: &str) -> Vec<&'static str> {
    STATEMENTS
        .iter()
        .filter(|(_, words)| !words.iter().all(|w| text.contains(w)))
        .map(|(name, _)| *name)
        .collect()
}

/// A card number (`C-` and digits) or an issue number (`#` and three or more digits).
fn card_or_issue_numbers(text: &str) -> Vec<String> {
    let b = text.as_bytes();
    let mut found = Vec::new();
    for i in 0..b.len() {
        let starts_word = i == 0 || !b[i - 1].is_ascii_alphanumeric();
        let digits = |from: usize| b[from..].iter().take_while(|c| c.is_ascii_digit()).count();
        if starts_word && b[i] == b'C' && b.get(i + 1) == Some(&b'-') && digits(i + 2) > 0 {
            found.push(text[i..i + 2 + digits(i + 2)].to_string());
        }
        if b[i] == b'#' && digits(i + 1) >= 3 {
            found.push(text[i..i + 1 + digits(i + 1)].to_string());
        }
    }
    found
}

/// AC-1 (REQ-1 to REQ-7): the page holds every statement and names no card or issue.
#[test]
fn the_page_holds_every_statement() {
    let text = page();
    assert_eq!(
        missing(&text),
        Vec::<&str>::new(),
        "{PAGE} lacks statements"
    );
    assert_eq!(
        card_or_issue_numbers(&text),
        Vec::<String>::new(),
        "{PAGE} names a card or an issue"
    );
}

/// AC-2 (REQ-8): with one statement's words taken out, the check names that statement.
#[test]
fn a_missing_statement_is_named() {
    let text = page();
    for (name, words) in STATEMENTS {
        let mut cut = text.clone();
        for w in words {
            cut = cut.replace(w, "");
        }
        assert!(
            missing(&cut).contains(&name),
            "removing {name:?} was not noticed"
        );
    }
    // The samples are built at run time so this file holds no card-shaped text itself.
    let (card, issue) = (format!("{}-12", 'C'), format!("{}2936", '#'));
    assert_eq!(
        card_or_issue_numbers(&format!("see {card} and {issue}")),
        [card, issue]
    );
    assert!(card_or_issue_numbers(&format!("AB{}-12 or {}12", 'C', '#')).is_empty());
}
