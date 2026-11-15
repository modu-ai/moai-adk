//! What a user may type into `moai card add` and how a card is named (#3070 K1).
//!
//! Everything here is a pure function of its arguments: no ledger, no file, no clock. The
//! rules are the ones the card tables enforce with `CHECK` (schema contract 6896940106), checked
//! first so the user gets a plain message that names the field instead of a constraint failure.
//! The ledger stays the last line of defence; nothing here replaces its checks.
//!
//! Three choices are this module's own, not the schema's, and are easy to change:
//! - a title and an acceptance criterion are one line (a control character is refused), because
//!   `list` and the export's human view show them on one line;
//! - a body holds no NUL character and is at most [`BODY_MAX_BYTES`] bytes;
//! - a write scope entry is a relative path with `/` separators, and the same entry twice is
//!   refused instead of being dropped without a word.
//!
//! What a write scope entry *means* (overlap between cards, directories against files) belongs
//! to the reservation code of #3068; this module only refuses what cannot be a repository path.

use std::fmt;

/// The longest title, in characters (`CHECK (length(title) BETWEEN 1 AND 300)`).
pub const TITLE_MAX_CHARS: usize = 300;
/// The longest body, in bytes.
pub const BODY_MAX_BYTES: usize = 1_048_576;
/// The kinds a card may have; the first is the default.
pub const KINDS: [&str; 7] = [
    "task", "feature", "fix", "docs", "research", "audit", "chore",
];
/// The priorities a card may have; `p2` is the default.
pub const PRIORITIES: [&str; 4] = ["p0", "p1", "p2", "p3"];
/// The states a card may be in, in the order the schema lists them.
pub const STATES: [&str; 6] = ["draft", "ready", "active", "review", "done", "hold"];

const DEFAULT_KIND: &str = "task";
const DEFAULT_PRIORITY: &str = "p2";

/// What the user typed for `card add`, before any check.
#[derive(Debug, Clone, Default)]
pub struct AddInput<'a> {
    pub title: &'a str,
    pub body: &'a str,
    pub kind: Option<&'a str>,
    pub priority: Option<&'a str>,
    pub scope: &'a [String],
    pub criteria: &'a [String],
}

/// An acceptance criterion as it will be stored. Its number (`seq`) is its place in the list,
/// counted from 1.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct NewAc {
    pub seq: i64,
    pub text: String,
}

/// A card that passed every check, in the form the ledger stores it. `add` always makes a
/// `draft`; no input chooses another state.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct NewCard {
    pub title: String,
    pub body: String,
    pub kind: &'static str,
    pub priority: &'static str,
    /// The stored JSON array text (compact, in the order given).
    pub write_scope: String,
    pub criteria: Vec<NewAc>,
}

/// The first thing wrong with an input. The message names the field and, for a list, which
/// entry (counted from 1, the way the user typed them).
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum InputError {
    TitleEmpty,
    TitleTooLong { chars: usize },
    TitleNotOneLine,
    BodyHasNul,
    BodyTooLong { bytes: usize },
    UnknownKind { given: String },
    UnknownPriority { given: String },
    UnknownState { given: String },
    ScopeEmpty { entry: usize },
    ScopeNotRelative { entry: usize },
    ScopeEscapes { entry: usize },
    ScopeBadCharacter { entry: usize },
    ScopeDuplicate { entry: usize, first: usize },
    CriterionEmpty { entry: usize },
    CriterionNotOneLine { entry: usize },
    BadReference { given: String },
}

impl fmt::Display for InputError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        use InputError::*;
        match self {
            TitleEmpty => f.write_str("the title is empty"),
            TitleTooLong { chars } => write!(
                f,
                "the title is {chars} characters; the longest is {TITLE_MAX_CHARS}"
            ),
            TitleNotOneLine => f.write_str("the title must be one line"),
            BodyHasNul => f.write_str("the body holds a NUL character"),
            BodyTooLong { bytes } => write!(
                f,
                "the body is {bytes} bytes; the longest is {BODY_MAX_BYTES}"
            ),
            UnknownKind { given } => {
                write!(f, "unknown kind `{given}`; use {}", KINDS.join(", "))
            }
            UnknownPriority { given } => write!(
                f,
                "unknown priority `{given}`; use {}",
                PRIORITIES.join(", ")
            ),
            UnknownState { given } => {
                write!(f, "unknown state `{given}`; use {}", STATES.join(", "))
            }
            ScopeEmpty { entry } => write!(f, "scope entry {entry} is empty"),
            ScopeNotRelative { entry } => write!(
                f,
                "scope entry {entry} is not a relative path; write it from the repository root"
            ),
            ScopeEscapes { entry } => write!(
                f,
                "scope entry {entry} leaves the repository (`..`) or names no place (`.`)"
            ),
            ScopeBadCharacter { entry } => write!(
                f,
                "scope entry {entry} holds a backslash or a control character; use `/` between folders"
            ),
            ScopeDuplicate { entry, first } => {
                write!(f, "scope entry {entry} repeats entry {first}")
            }
            CriterionEmpty { entry } => write!(f, "criterion {entry} is empty"),
            CriterionNotOneLine { entry } => write!(f, "criterion {entry} must be one line"),
            BadReference { given } => write!(
                f,
                "`{given}` is not a card; write C-<number> or #<source number>"
            ),
        }
    }
}

impl std::error::Error for InputError {}

fn has_control(text: &str) -> bool {
    text.chars().any(char::is_control)
}

fn title(text: &str) -> Result<String, InputError> {
    let text = text.trim();
    if text.is_empty() {
        return Err(InputError::TitleEmpty);
    }
    if has_control(text) {
        return Err(InputError::TitleNotOneLine);
    }
    let chars = text.chars().count();
    if chars > TITLE_MAX_CHARS {
        return Err(InputError::TitleTooLong { chars });
    }
    Ok(text.to_string())
}

fn body(text: &str) -> Result<String, InputError> {
    if text.contains('\0') {
        return Err(InputError::BodyHasNul);
    }
    if text.len() > BODY_MAX_BYTES {
        return Err(InputError::BodyTooLong { bytes: text.len() });
    }
    Ok(text.to_string())
}

fn pick(
    given: Option<&str>,
    default: &'static str,
    list: &'static [&'static str],
) -> Option<&'static str> {
    match given {
        None => Some(default),
        Some(text) => list.iter().copied().find(|v| *v == text),
    }
}

/// The kind a card gets: `task` when none was typed.
pub fn kind(given: Option<&str>) -> Result<&'static str, InputError> {
    pick(given, DEFAULT_KIND, &KINDS).ok_or_else(|| InputError::UnknownKind {
        given: given.unwrap_or_default().to_string(),
    })
}

/// The priority a card gets: `p2` when none was typed.
pub fn priority(given: Option<&str>) -> Result<&'static str, InputError> {
    pick(given, DEFAULT_PRIORITY, &PRIORITIES).ok_or_else(|| InputError::UnknownPriority {
        given: given.unwrap_or_default().to_string(),
    })
}

/// A state typed for `list --state`.
pub fn state(given: &str) -> Result<&'static str, InputError> {
    STATES
        .iter()
        .copied()
        .find(|s| *s == given)
        .ok_or_else(|| InputError::UnknownState {
            given: given.to_string(),
        })
}

fn scope_entry(text: &str, entry: usize) -> Result<String, InputError> {
    let text = text.trim();
    if text.is_empty() {
        return Err(InputError::ScopeEmpty { entry });
    }
    if text.contains('\\') || has_control(text) {
        return Err(InputError::ScopeBadCharacter { entry });
    }
    if text.starts_with('/') {
        return Err(InputError::ScopeNotRelative { entry });
    }
    // One trailing `/` is how a folder is written; it is kept as typed.
    let walk = text.strip_suffix('/').unwrap_or(text);
    if walk.is_empty() || walk.split('/').any(|part| matches!(part, "" | "." | "..")) {
        return Err(InputError::ScopeEscapes { entry });
    }
    Ok(text.to_string())
}

fn scope(entries: &[String]) -> Result<String, InputError> {
    let mut seen: Vec<String> = Vec::with_capacity(entries.len());
    for (i, raw) in entries.iter().enumerate() {
        let clean = scope_entry(raw, i + 1)?;
        if let Some(first) = seen.iter().position(|s| *s == clean) {
            return Err(InputError::ScopeDuplicate {
                entry: i + 1,
                first: first + 1,
            });
        }
        seen.push(clean);
    }
    // A list of strings always serializes; there is no failure to report.
    Ok(serde_json::to_string(&seen).unwrap_or_else(|_| "[]".to_string()))
}

fn criteria(entries: &[String]) -> Result<Vec<NewAc>, InputError> {
    let mut out = Vec::with_capacity(entries.len());
    for (i, raw) in entries.iter().enumerate() {
        let entry = i + 1;
        let text = raw.trim();
        if text.is_empty() {
            return Err(InputError::CriterionEmpty { entry });
        }
        if has_control(text) {
            return Err(InputError::CriterionNotOneLine { entry });
        }
        out.push(NewAc {
            seq: entry as i64,
            text: text.to_string(),
        });
    }
    Ok(out)
}

/// Checks everything `card add` was given and returns the card to store. The first problem
/// found is returned; the order is title, body, kind, priority, scope, criteria.
pub fn validate_add(input: &AddInput<'_>) -> Result<NewCard, InputError> {
    Ok(NewCard {
        title: title(input.title)?,
        body: body(input.body)?,
        kind: kind(input.kind)?,
        priority: priority(input.priority)?,
        write_scope: scope(input.scope)?,
        criteria: criteria(input.criteria)?,
    })
}

/// How a command names one card.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CardRef {
    /// `C-<card_id>`.
    Id(i64),
    /// `#<number>`: the number the card came from on its code host. Which host and project is
    /// the command's business; the same number can exist in more than one.
    SourceNumber(i64),
}

fn positive(digits: &str) -> Option<i64> {
    // One spelling per number: no sign, no leading zero, ASCII digits only.
    let ok = !digits.is_empty()
        && digits.bytes().all(|b| b.is_ascii_digit())
        && !digits.starts_with('0');
    if ok {
        digits.parse::<i64>().ok()
    } else {
        None
    }
}

/// Reads `C-12` (the `C` may be lower case) or `#3066`.
pub fn parse_ref(text: &str) -> Result<CardRef, InputError> {
    let found = if let Some(rest) = text.strip_prefix("C-").or_else(|| text.strip_prefix("c-")) {
        positive(rest).map(CardRef::Id)
    } else {
        text.strip_prefix('#')
            .and_then(positive)
            .map(CardRef::SourceNumber)
    };
    found.ok_or_else(|| InputError::BadReference {
        given: text.to_string(),
    })
}

/// The name a card is shown under.
pub fn card_name(card_id: i64) -> String {
    format!("C-{card_id}")
}

#[cfg(test)]
mod tests {
    use super::*;

    fn add(title: &str) -> AddInput<'_> {
        AddInput {
            title,
            ..AddInput::default()
        }
    }

    fn strings(items: &[&str]) -> Vec<String> {
        items.iter().map(|s| s.to_string()).collect()
    }

    #[test]
    fn a_bare_title_gets_the_defaults() {
        let card = validate_add(&add("  Fix the claim lease  ")).unwrap();
        assert_eq!(card.title, "Fix the claim lease");
        assert_eq!(card.body, "");
        assert_eq!((card.kind, card.priority), ("task", "p2"));
        assert_eq!(card.write_scope, "[]");
        assert!(card.criteria.is_empty());
    }

    #[test]
    fn the_title_is_one_to_three_hundred_characters() {
        assert_eq!(validate_add(&add("")), Err(InputError::TitleEmpty));
        assert_eq!(validate_add(&add(" \t ")), Err(InputError::TitleEmpty));
        assert!(validate_add(&add("a")).is_ok());
        assert!(validate_add(&add(&"a".repeat(300))).is_ok());
        assert_eq!(
            validate_add(&add(&"a".repeat(301))),
            Err(InputError::TitleTooLong { chars: 301 })
        );
    }

    #[test]
    fn the_title_length_counts_characters_not_bytes() {
        // 300 Hangul syllables are 900 bytes and must pass; 301 must not.
        assert!(validate_add(&add(&"가".repeat(300))).is_ok());
        assert_eq!(
            validate_add(&add(&"가".repeat(301))),
            Err(InputError::TitleTooLong { chars: 301 })
        );
    }

    #[test]
    fn a_title_with_a_line_break_or_control_character_is_refused() {
        for bad in ["a\nb", "a\tb", "a\u{1b}[31mb", "a\rb", "a\0b"] {
            assert_eq!(
                validate_add(&add(bad)),
                Err(InputError::TitleNotOneLine),
                "{bad:?}"
            );
        }
        // A control character that trim removes never reaches the check.
        assert!(validate_add(&add("title\n")).is_ok());
    }

    #[test]
    fn the_body_has_no_nul_and_at_most_one_mebibyte() {
        let mut input = add("t");
        let edge = "x".repeat(1_048_576);
        input.body = &edge;
        assert!(validate_add(&input).is_ok());
        let over = "x".repeat(1_048_577);
        input.body = &over;
        assert_eq!(
            validate_add(&input),
            Err(InputError::BodyTooLong { bytes: 1_048_577 })
        );
        input.body = "a\0b";
        assert_eq!(validate_add(&input), Err(InputError::BodyHasNul));
        // Line breaks, tabs and Markdown are the body's business.
        input.body = "# Heading\n\n- a\tb\r\n";
        assert_eq!(
            validate_add(&input).unwrap().body,
            "# Heading\n\n- a\tb\r\n"
        );
    }

    #[test]
    fn kind_and_priority_come_from_the_schema_lists_and_say_what_is_allowed() {
        for k in [
            "task", "feature", "fix", "docs", "research", "audit", "chore",
        ] {
            assert_eq!(kind(Some(k)).unwrap(), k);
        }
        for p in ["p0", "p1", "p2", "p3"] {
            assert_eq!(priority(Some(p)).unwrap(), p);
        }
        assert_eq!(kind(None).unwrap(), "task");
        assert_eq!(priority(None).unwrap(), "p2");
        for bad in ["", "Task", "bug", "task "] {
            assert!(kind(Some(bad)).is_err(), "{bad:?}");
        }
        for bad in ["", "P1", "p4", "p1 "] {
            assert!(priority(Some(bad)).is_err(), "{bad:?}");
        }
        let e = kind(Some("bug")).unwrap_err();
        assert_eq!(
            e,
            InputError::UnknownKind {
                given: "bug".into()
            }
        );
        assert!(e
            .to_string()
            .contains("task, feature, fix, docs, research, audit, chore"));
        let e = priority(Some("p4")).unwrap_err();
        assert!(e.to_string().contains("p0, p1, p2, p3"), "{e}");
    }

    #[test]
    fn a_state_filter_accepts_the_six_states_only() {
        for s in ["draft", "ready", "active", "review", "done", "hold"] {
            assert_eq!(state(s).unwrap(), s);
        }
        for bad in ["", "Done", "open", "closed"] {
            assert_eq!(
                state(bad),
                Err(InputError::UnknownState { given: bad.into() })
            );
        }
        assert!(state("open")
            .unwrap_err()
            .to_string()
            .contains("draft, ready, active, review, done, hold"));
    }

    #[test]
    fn scope_entries_are_stored_as_a_compact_json_array_in_the_order_given() {
        let entries = strings(&["src/card/", " tests/card_add.rs ", "docs/a b.md"]);
        let mut input = add("t");
        input.scope = &entries;
        assert_eq!(
            validate_add(&input).unwrap().write_scope,
            r#"["src/card/","tests/card_add.rs","docs/a b.md"]"#
        );
        // JSON escaping keeps a quote or a non-ASCII name intact.
        let entries = strings(&["docs/\"quoted\".md", "문서/가.md"]);
        input.scope = &entries;
        assert_eq!(
            validate_add(&input).unwrap().write_scope,
            r#"["docs/\"quoted\".md","문서/가.md"]"#
        );
    }

    #[test]
    fn a_scope_entry_that_cannot_be_a_repository_path_is_refused_by_number() {
        let cases: [(&str, InputError); 11] = [
            ("", InputError::ScopeEmpty { entry: 2 }),
            ("   ", InputError::ScopeEmpty { entry: 2 }),
            ("/etc/passwd", InputError::ScopeNotRelative { entry: 2 }),
            ("..", InputError::ScopeEscapes { entry: 2 }),
            ("../x", InputError::ScopeEscapes { entry: 2 }),
            ("a/../b", InputError::ScopeEscapes { entry: 2 }),
            (".", InputError::ScopeEscapes { entry: 2 }),
            ("a//b", InputError::ScopeEscapes { entry: 2 }),
            ("/", InputError::ScopeNotRelative { entry: 2 }),
            ("a\\b", InputError::ScopeBadCharacter { entry: 2 }),
            ("a\nb", InputError::ScopeBadCharacter { entry: 2 }),
        ];
        for (bad, expected) in cases {
            let entries = strings(&["ok/", bad]);
            let mut input = add("t");
            input.scope = &entries;
            assert_eq!(validate_add(&input), Err(expected), "{bad:?}");
        }
    }

    #[test]
    fn a_dot_inside_a_name_is_not_a_dot_segment() {
        let entries = strings(&["a..b", ".github/", "x/.hidden", "..a", "a..", "v1.0/"]);
        let mut input = add("t");
        input.scope = &entries;
        assert!(validate_add(&input).is_ok());
    }

    #[test]
    fn the_same_scope_entry_twice_names_both_places() {
        let entries = strings(&["src/", "tests/", "src/"]);
        let mut input = add("t");
        input.scope = &entries;
        assert_eq!(
            validate_add(&input),
            Err(InputError::ScopeDuplicate { entry: 3, first: 1 })
        );
        // `src` and `src/` are different spellings; telling whether they overlap is #3068's job.
        let entries = strings(&["src", "src/"]);
        input.scope = &entries;
        assert!(validate_add(&input).is_ok());
    }

    #[test]
    fn criteria_are_numbered_from_one_in_the_order_typed() {
        let entries = strings(&["  first  ", "second"]);
        let mut input = add("t");
        input.criteria = &entries;
        assert_eq!(
            validate_add(&input).unwrap().criteria,
            vec![
                NewAc {
                    seq: 1,
                    text: "first".into()
                },
                NewAc {
                    seq: 2,
                    text: "second".into()
                },
            ]
        );
    }

    #[test]
    fn an_empty_or_multi_line_criterion_is_refused_by_number() {
        for (bad, expected) in [
            ("", InputError::CriterionEmpty { entry: 2 }),
            ("  \t", InputError::CriterionEmpty { entry: 2 }),
            ("a\nb", InputError::CriterionNotOneLine { entry: 2 }),
            ("a\u{7}b", InputError::CriterionNotOneLine { entry: 2 }),
        ] {
            let entries = strings(&["fine", bad]);
            let mut input = add("t");
            input.criteria = &entries;
            assert_eq!(validate_add(&input), Err(expected), "{bad:?}");
        }
    }

    #[test]
    fn the_first_problem_in_field_order_is_the_one_reported() {
        let scope = strings(&["/abs"]);
        let criteria = strings(&[""]);
        let input = AddInput {
            title: "",
            body: "a\0",
            kind: Some("bug"),
            priority: Some("p9"),
            scope: &scope,
            criteria: &criteria,
        };
        assert_eq!(validate_add(&input), Err(InputError::TitleEmpty));
        let input = AddInput {
            title: "t",
            ..input
        };
        assert_eq!(validate_add(&input), Err(InputError::BodyHasNul));
        let input = AddInput { body: "", ..input };
        assert!(matches!(
            validate_add(&input),
            Err(InputError::UnknownKind { .. })
        ));
        let input = AddInput {
            kind: None,
            ..input
        };
        assert!(matches!(
            validate_add(&input),
            Err(InputError::UnknownPriority { .. })
        ));
        let input = AddInput {
            priority: None,
            ..input
        };
        assert!(matches!(
            validate_add(&input),
            Err(InputError::ScopeNotRelative { entry: 1 })
        ));
        let input = AddInput {
            scope: &[],
            ..input
        };
        assert_eq!(
            validate_add(&input),
            Err(InputError::CriterionEmpty { entry: 1 })
        );
    }

    #[test]
    fn a_card_is_named_by_c_dash_number_or_by_hash_source_number() {
        assert_eq!(parse_ref("C-1"), Ok(CardRef::Id(1)));
        assert_eq!(parse_ref("c-42"), Ok(CardRef::Id(42)));
        assert_eq!(parse_ref("#3066"), Ok(CardRef::SourceNumber(3066)));
        assert_eq!(
            parse_ref("C-9223372036854775807"),
            Ok(CardRef::Id(i64::MAX))
        );
        assert_eq!(card_name(7), "C-7");
        assert_eq!(parse_ref(&card_name(123)), Ok(CardRef::Id(123)));
    }

    #[test]
    fn anything_else_is_not_a_card_name() {
        for bad in [
            "",
            "C",
            "C-",
            "C-0",
            "C-01",
            "C-+1",
            "C--1",
            "C-1 ",
            " C-1",
            "C-１",
            "C-1a",
            "C-9223372036854775808",
            "c_1",
            "#",
            "#0",
            "#07",
            "#-1",
            "# 1",
            "3066",
            "C-1\n",
            "#1.5",
            "cc-1",
        ] {
            assert_eq!(
                parse_ref(bad),
                Err(InputError::BadReference {
                    given: bad.to_string()
                }),
                "{bad:?}"
            );
        }
    }
}
