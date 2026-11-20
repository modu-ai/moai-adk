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
/// entry (counted from 1, the way the user typed them). The `*SecretShaped` variants carry the
/// count of words the card text screen ([`secret_words`]) refuses.
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
    TitleSecretShaped { words: usize },
    BodySecretShaped { words: usize },
    CriterionSecretShaped { entry: usize, words: usize },
    BadReference { given: String },
}

/// The longest piece of a user's text an error message repeats, in characters.
pub const ECHO_MAX_CHARS: usize = 40;

/// Whether [`echo`] shows `c` as `?`: a control character (C0, DEL, C1), a line or paragraph
/// separator, a direction mark, override or isolate, or a character that takes no room (zero-width
/// space, word joiner, byte order mark). The zero-width joiners U+200C/U+200D stay: they join
/// letters of real words.
fn hidden_in_echo(c: char) -> bool {
    c.is_control()
        || matches!(
            c,
            '\u{061C}'
                | '\u{200B}'
                | '\u{200E}'..='\u{200F}'
                | '\u{2028}'..='\u{202E}'
                | '\u{2060}'
                | '\u{2066}'..='\u{2069}'
                | '\u{FEFF}'
        )
}

/// What an error message may repeat of text the user typed: the first [`ECHO_MAX_CHARS`]
/// characters with every control character, line separator, direction character and invisible
/// character shown as `?` (see [`hidden_in_echo`]), so an escape sequence typed as a card name does
/// not reach a terminal and a name cannot reorder or hide the text around it, and `…` when the
/// text was cut.
pub fn echo(text: &str) -> String {
    let mut out: String = text
        .chars()
        .take(ECHO_MAX_CHARS)
        .map(|c| if hidden_in_echo(c) { '?' } else { c })
        .collect();
    if text.chars().count() > ECHO_MAX_CHARS {
        out.push('…');
    }
    out
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
                write!(f, "unknown kind `{}`; use {}", echo(given), KINDS.join(", "))
            }
            UnknownPriority { given } => write!(
                f,
                "unknown priority `{}`; use {}",
                echo(given),
                PRIORITIES.join(", ")
            ),
            UnknownState { given } => {
                write!(f, "unknown state `{}`; use {}", echo(given), STATES.join(", "))
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
            TitleSecretShaped { words } => write!(f, "the title has {}", secret_refusal(*words)),
            BodySecretShaped { words } => write!(f, "the body has {}", secret_refusal(*words)),
            CriterionSecretShaped { entry, words } => {
                write!(f, "criterion {entry} has {}", secret_refusal(*words))
            }
            BadReference { given } => write!(
                f,
                "`{}` is not a card; write C-<number> or #<source number>",
                echo(given)
            ),
        }
    }
}

impl std::error::Error for InputError {}

fn has_control(text: &str) -> bool {
    text.chars().any(char::is_control)
}

/// The name pieces (cut at `-` and `_`, any letter case) that make an assignment secret-named;
/// `api` followed by `key` is the fourth (SPEC-20261119-card-secret-screen D-1, D-2).
const SECRET_PIECES: [&str; 3] = ["password", "token", "secret"];

/// What a refusal says after the field it names: the count and the reason, never the text.
pub fn secret_refusal(words: usize) -> String {
    format!(
        "{words} word(s) with the shape of a secret (a token, a key, a secret-named \
         `name=value`); a card text is kept for good and written out by `card export`, so take \
         them out"
    )
}

/// The one text screen of the card commands (SPEC-20261119-card-secret-screen REQ-1..4,
/// SPEC-20261120-card-secret-followups REQ-1..7): how many words of `text` it refuses. A word is
/// refused once, whatever is wrong with it:
/// - it holds a `name=value` whose name has a secret piece ([`SECRET_PIECES`], or `api` and `key`
///   side by side), whatever the value is, an empty one included;
/// - or the equals sign starts its own word and the word before it is such a name, spaces and
///   tabs between the two notwithstanding (follow-ups REQ-1);
/// - or its name sits in quotes right before the equals sign, which are read stripped
///   (follow-ups REQ-2), or holds a secret word behind a camel hump (follow-ups REQ-3);
/// - or its value has a credential shape (`redact::redact_outgoing`), after the name of a
///   `--option=value` whose value is plain was blanked (a plain option value is prose);
/// - or a value with a credential shape follows a separator (`= : , ; / | @ ? & . + # \ (`)
///   anywhere inside the word, after a plain name or none (audit N1, follow-ups REQ-6);
/// - a run of lower-case hyphenated words after `sk-` is a name, not a key, and goes through
///   (the rule `decide`'s own `without_sk_names` states) — but only as a name: when `=` follows
///   it, the run is the name of an assignment and its pieces are checked like any other name
///   (audit B1), so the names are read from the word as written, before the run is blanked.
pub fn secret_words(text: &str) -> usize {
    let blanked = without_sk_names(text);
    // Blanking changes only letters, digits, `-` and `_`, so both texts hold the same words.
    let written: Vec<&str> = text.split_whitespace().collect();
    let mut count = 0;
    for (i, (word, blanked_word)) in written.iter().zip(blanked.split_whitespace()).enumerate() {
        if word_is_refused(word, blanked_word) {
            count += 1;
            continue;
        }
        // Follow-ups REQ-1: `name = value`, the equals sign starting its own word — the word
        // before it is the name (read as written, quotes stripped like any assignment name).
        if written.get(i + 1).is_some_and(|next| next.starts_with('='))
            && name_is_secret(strip_quotes(word))
        {
            count += 1;
        }
    }
    count
}

/// The quote characters a person may wrap an assignment name in (follow-ups REQ-2).
fn strip_quotes(word: &str) -> &str {
    word.trim_matches(|c| matches!(c, '"' | '\'' | '`'))
}

/// `written` is the word as the user typed it, `blanked` the same word with the plain `sk` names
/// blanked.
fn word_is_refused(written: &str, blanked: &str) -> bool {
    if secret_named_assignment(written) {
        return true;
    }
    let (plain_options_blanked, value_shaped) = blank_option_names(blanked);
    if value_shaped > 0 || crate::redact::redact_outgoing(&plain_options_blanked).1 > 0 {
        return true;
    }
    // The shared screen sees a prefix shape only at the start of a word: make each piece between
    // two separators a word of its own by turning the separators into spaces, and screen the
    // text once (audit B2: screening everything behind every separator took time in the square
    // of the word's length). A shape holds none of the separators, so it lies inside one piece —
    // except the three-part dotted token (`eyJ…`), whose dots are separators too: it is seen in
    // a second pass that keeps dot-joined pieces whole (follow-ups REQ-6). The separator set is
    // `= : ,` (audit N1) plus the eleven follow-ups REQ-6 characters.
    let split = |also_dots: bool| -> String {
        blanked
            .chars()
            .map(|c| {
                if matches!(
                    c,
                    '=' | ':' | ',' | ';' | '/' | '|' | '@' | '?' | '&' | '+' | '#' | '\\' | '('
                ) || (also_dots && c == '.')
                {
                    ' '
                } else {
                    c
                }
            })
            .collect()
    };
    crate::redact::redact_outgoing(&split(true)).1 > 0
        || crate::redact::redact_outgoing(&split(false)).1 > 0
}

fn name_is_secret(name: &str) -> bool {
    let lower = name.to_ascii_lowercase();
    let pieces: Vec<&str> = lower.split(['-', '_']).filter(|p| !p.is_empty()).collect();
    pieces.iter().any(|p| SECRET_PIECES.contains(p))
        || pieces.windows(2).any(|w| w[0] == "api" && w[1] == "key")
        || camel_hump_secret(name)
}

/// Follow-ups REQ-3: the name holds a secret word, or a joined `api…key`, right behind a camel
/// hump — an upper-case letter that follows a lower-case one — as `myTokenCount` or `myApiKey`
/// do. An all-lower-case joined or plural name (`tokens`, `mytoken`) keeps passing: the pieces
/// rule of SPEC-20261119-card-secret-screen REQ-2 stands.
fn camel_hump_secret(name: &str) -> bool {
    // One lower-cased copy of the whole name: ASCII lower-casing keeps every byte offset, so a
    // hump's offset in `name` splits `lower` too, and each hump compares two short slices, no new
    // string (audit B1: building the rest of the name per hump took time in the square of its
    // length).
    let lower = name.to_ascii_lowercase();
    let mut previous: Option<char> = None;
    for (at, c) in name.char_indices() {
        let hump = previous.is_some_and(|p| p.is_ascii_lowercase()) && c.is_uppercase();
        previous = Some(c);
        if !hump {
            continue;
        }
        let (before, tail) = lower.split_at(at);
        if SECRET_PIECES.iter().any(|w| tail.starts_with(w))
            || tail.starts_with("apikey")
            // `apiKey`: the hump starts the `key` half of a joined `api…key`.
            || (tail.starts_with("key") && before.ends_with("api"))
        {
            return true;
        }
    }
    false
}

/// Whether any `=` of the word follows a secret name: the name is the run of letters, digits,
/// `-` and `_` right before the `=`, so a limit flag with a token piece and a number is named,
/// and so is a name with a password piece behind other dashes; a plain flag whose value is an
/// assignment with a secret name is named by its second assignment. Quotes around the name are
/// read stripped (follow-ups REQ-2).
fn secret_named_assignment(word: &str) -> bool {
    let word = word.replace(['"', '\'', '`'], "");
    let bytes = word.as_bytes();
    let name_byte = |b: &u8| b.is_ascii_alphanumeric() || *b == b'_' || *b == b'-';
    (0..bytes.len()).filter(|&at| bytes[at] == b'=').any(|eq| {
        // Every byte of a multi-byte character fails `name_byte`, so `start` is a boundary.
        let start = bytes[..eq]
            .iter()
            .rposition(|b| !name_byte(b))
            .map_or(0, |i| i + 1);
        name_is_secret(&word[start..eq])
    })
}

/// The name of an option assignment whose value is not itself secret-shaped is blanked, and a
/// value that is secret-shaped is counted in the second return: the shared screen sees a word
/// like `--flag=sk-…` as one word that starts with the option name, so the value alone decides.
fn blank_option_names(text: &str) -> (String, usize) {
    let mut bytes = text.as_bytes().to_vec();
    let mut shaped_values = 0;
    let mut at = 0;
    while at < bytes.len() {
        let starts =
            bytes[at..].starts_with(b"--") && (at == 0 || !bytes[at - 1].is_ascii_alphanumeric());
        if !starts {
            at += 1;
            continue;
        }
        let name_end = (at + 2..bytes.len())
            .find(|&i| !(bytes[i].is_ascii_alphanumeric() || bytes[i] == b'-' || bytes[i] == b'_'))
            .unwrap_or(bytes.len());
        if name_end < bytes.len() && bytes[name_end] == b'=' {
            let value_end = (name_end + 1..bytes.len())
                .find(|&i| bytes[i].is_ascii_whitespace())
                .unwrap_or(bytes.len());
            let value = std::str::from_utf8(&bytes[name_end + 1..value_end]).unwrap_or("");
            let (_, shaped) = crate::redact::redact_outgoing(value);
            if shaped == 0 {
                bytes[at + 2..name_end].fill(b'x');
            } else {
                shaped_values += shaped;
            }
            at = value_end;
        } else {
            at = name_end;
        }
    }
    // Only ASCII bytes were replaced, each by one ASCII byte.
    (
        String::from_utf8(bytes).unwrap_or_else(|_| text.to_string()),
        shaped_values,
    )
}

/// `sk-` before lower-case hyphenated words, blanked the way `decide` states it: no digit,
/// upper-case letter or underscore anywhere in the run, `sk-proj-` apart.
fn without_sk_names(text: &str) -> String {
    let mut bytes = text.as_bytes().to_vec();
    let key = |b: u8| b.is_ascii_alphanumeric() || b == b'_' || b == b'-';
    let mut at = 0;
    while at < bytes.len() {
        let starts = bytes[at..].starts_with(b"sk-")
            && !bytes[at..].starts_with(b"sk-proj-")
            && (at == 0 || !bytes[at - 1].is_ascii_alphanumeric());
        if !starts {
            at += 1;
            continue;
        }
        let mut len = 0;
        while at + len < bytes.len() && key(bytes[at + len]) {
            len += 1;
        }
        let words = &bytes[at + 3..at + len];
        let name = words.contains(&b'-')
            && words
                .split(|b| *b == b'-')
                .all(|word| !word.is_empty() && word.iter().all(u8::is_ascii_lowercase));
        if name {
            bytes[at..at + len].fill(b'x');
        }
        at += len;
    }
    String::from_utf8(bytes).unwrap_or_else(|_| text.to_string())
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
    let words = secret_words(text);
    if words > 0 {
        return Err(InputError::TitleSecretShaped { words });
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
    let words = secret_words(text);
    if words > 0 {
        return Err(InputError::BodySecretShaped { words });
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
    let seen = validate_scope(entries)?;
    // A list of strings always serializes; there is no failure to report.
    Ok(serde_json::to_string(&seen).unwrap_or_else(|_| "[]".to_string()))
}

/// A write scope as `add --scope` takes it, for every action that sets one (`add`, `scope`,
/// `take --scope`): each entry trimmed and checked by [`scope_entry`], no entry twice; the
/// entries in the order given. An empty list is not refused here.
pub fn validate_scope(entries: &[String]) -> Result<Vec<String>, InputError> {
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
    Ok(seen)
}

/// The criterion texts of `card add --ac` and `card criterion-add --ac`: trimmed, not empty and
/// one line each. A refusal names the entry by its place in the list, counted from 1.
pub fn criterion_texts(entries: &[String]) -> Result<Vec<String>, InputError> {
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
        let words = secret_words(text);
        if words > 0 {
            return Err(InputError::CriterionSecretShaped { entry, words });
        }
        out.push(text.to_string());
    }
    Ok(out)
}

fn criteria(entries: &[String]) -> Result<Vec<NewAc>, InputError> {
    Ok(criterion_texts(entries)?
        .into_iter()
        .enumerate()
        .map(|(i, text)| NewAc {
            seq: i as i64 + 1,
            text,
        })
        .collect())
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

    #[test]
    fn an_error_does_not_repeat_line_separators_direction_marks_or_invisible_characters() {
        // Each is shown as `?`, alone and inside a word.
        let hidden = [
            '\u{200B}', '\u{200E}', '\u{200F}', '\u{061C}', '\u{2028}', '\u{2029}', '\u{202A}',
            '\u{202B}', '\u{202C}', '\u{202D}', '\u{202E}', '\u{2060}', '\u{2066}', '\u{2067}',
            '\u{2068}', '\u{2069}', '\u{FEFF}',
        ];
        for c in hidden {
            assert_eq!(echo(&format!("a{c}b")), "a?b", "U+{:04X}", c as u32);
            assert_eq!(echo(&c.to_string()), "?", "U+{:04X}", c as u32);
            let text = InputError::UnknownKind {
                given: format!("x{c}y"),
            }
            .to_string();
            assert!(!text.contains(c), "U+{:04X} in {text:?}", c as u32);
        }
        // The characters next to each range are left alone: the zero-width joiners that join
        // letters, spaces, a hyphenation point and unassigned code points.
        for c in [
            '\u{061B}', '\u{061D}', '\u{200A}', '\u{200C}', '\u{200D}', '\u{2010}', '\u{2027}',
            '\u{202F}', '\u{205F}', '\u{2061}', '\u{2065}', '\u{206A}', '\u{FEFE}', '\u{FF00}',
        ] {
            let s = format!("a{c}b");
            assert_eq!(echo(&s), s, "U+{:04X}", c as u32);
        }
        // The 40-character cut counts a replaced character like any other.
        let long = format!("{}{}", "\u{202E}".repeat(40), "z");
        assert_eq!(echo(&long), format!("{}…", "?".repeat(40)));
    }

    #[test]
    fn an_error_repeats_at_most_forty_plain_characters_of_what_was_typed() {
        // Control characters never reach the message as they were typed.
        for given in ["\u{1b}[31mred", "a\nb", "a\u{7}b", "\u{0}"] {
            for err in [
                InputError::UnknownKind {
                    given: given.into(),
                },
                InputError::UnknownPriority {
                    given: given.into(),
                },
                InputError::UnknownState {
                    given: given.into(),
                },
                InputError::BadReference {
                    given: given.into(),
                },
            ] {
                let text = err.to_string();
                assert!(!text.chars().any(char::is_control), "{text:?}");
                assert!(text.contains('?'), "{text:?}");
            }
        }
        // 40 characters are repeated whole, 41 are cut with a mark.
        let (ok, cut) = ("x".repeat(40), "x".repeat(41));
        assert_eq!(echo(&ok), ok);
        assert_eq!(echo(&cut), format!("{ok}…"));
        // Counted in characters, not bytes.
        let wide = "한".repeat(40);
        assert_eq!(echo(&wide), wide);
        assert_eq!(echo(""), "");
        let msg = InputError::BadReference { given: cut }.to_string();
        assert!(
            msg.contains(&format!("{ok}…")) && !msg.contains(&"x".repeat(41)),
            "{msg}"
        );
    }

    // SPEC-20261119-card-secret-screen. Every word is joined from pieces so this file holds no
    // secret-shaped word.
    fn j(parts: &[&str]) -> String {
        parts.concat()
    }

    /// REQ-1: the name's `-`/`_` pieces decide, case does not, the value does not.
    #[test]
    fn secret_screen_name_pieces_decide() {
        let pw = j(&["pass", "word"]);
        let tk = j(&["to", "ken"]);
        for (text, n) in [
            (format!("--{pw}=x"), 1),
            (format!("{pw}="), 1),
            (format!("--{tk}=100"), 1),
            (format!("--max-{tk}=100"), 1),
            (format!("--skip_{}=true", pw.to_uppercase()), 1),
            (format!("x--{pw}=1"), 1),
            (j(&["--sec", "ret=1"]), 1),
            (j(&["--api", "-key=1"]), 1),
            (j(&["API", "_KEY=1"]), 1),
            (format!("--config={pw}=1"), 1),
            (format!("--{pw}=1 --{tk}=2"), 2),
        ] {
            assert_eq!(secret_words(&text), n, "{text}");
        }
    }

    /// REQ-2, REQ-4: no secret piece, a plain value, a space-separated value, a lower-case
    /// hyphenated `sk` name — all pass; a piece must equal the word, not contain it.
    #[test]
    fn secret_screen_passes_plain_assignments_and_names() {
        let pw = j(&["pass", "word"]);
        for text in [
            "login works after a restart".to_string(),
            "--foo=bar --config=prod.yaml --api-version=2".to_string(),
            j(&["--to", "kens=3 --sec", "retive=1"]),
            j(&["--pass", "wordless=1"]),
            format!("--{pw} hunter2"),
            j(&["s", "k-learn-pipeline tuned"]),
            "key=value monkey=1 api=1".to_string(),
        ] {
            assert_eq!(secret_words(&text), 0, "{text}");
        }
    }

    /// REQ-3: a credential-shaped value is refused whatever the name, and one word counts once.
    #[test]
    fn secret_screen_counts_a_word_once() {
        let key = j(&["AK", "IA", "ABCDEFGHIJKLMNOP"]);
        let tk = j(&["to", "ken"]);
        assert_eq!(secret_words(&format!("--out={key}")), 1);
        assert_eq!(secret_words(&format!("--{tk}={key}")), 1);
        assert_eq!(secret_words(&format!("--{tk}={key} and --{tk}={key}")), 2);
        assert_eq!(secret_words(&j(&["s", "k-", "Ab3", "cdefghijklmnop"])), 1);
        assert_eq!(secret_words(""), 0);
    }

    /// Audit B1: the names of an assignment are read as written, so an `sk` run with a secret
    /// piece in front of `=` is refused, while the bare run and a plain-named assignment pass.
    #[test]
    fn secret_screen_reads_sk_names_before_they_are_blanked() {
        let (sk, tk, pw) = (j(&["s", "k-"]), j(&["to", "ken"]), j(&["pass", "word"]));
        for text in [
            format!("{sk}learn-{pw}=1"),
            format!("--{sk}a-b-{tk}=1"),
            format!("{sk}a-{tk}-b="),
            format!("{sk}{}=1", j(&["api", "-", "key"])),
        ] {
            assert_eq!(secret_words(&text), 1, "{text}");
        }
        for text in [
            format!("{sk}learn-pipeline"),
            format!("{sk}learn-{pw}"),
            format!("{sk}learn-pipeline=on"),
        ] {
            assert_eq!(secret_words(&text), 0, "{text}");
        }
    }

    /// Audit N1: a credential-shaped value is seen after a separator inside the word.
    #[test]
    fn secret_screen_sees_values_after_a_separator() {
        let key = j(&["AK", "IA", "ABCDEFGHIJKLMNOP"]);
        for text in [
            format!("config={key}"),
            format!("key:{key}"),
            format!("a,{key}"),
            format!("x=y,{key}"),
            format!("(see:{key})"),
        ] {
            assert_eq!(secret_words(&text), 1, "{text}");
        }
        for text in [
            "see: the list, then run it",
            "mode=fast, retries=3",
            "time 12:30,x",
        ] {
            assert_eq!(secret_words(text), 0, "{text}");
        }
        // One word counts once even when two separators lead to shapes.
        assert_eq!(secret_words(&format!("a:{key},{key}")), 1);
    }

    /// Audit B2: one long word without a space is screened in time that grows with its length,
    /// not with its square. Sizes: 200 KB, one byte under the note limit and the limit itself; the
    /// bound is far above the linear time and far below the square-time one (40 KB took 30 s).
    #[test]
    fn secret_screen_time_on_one_long_word_has_a_bound() {
        use std::time::{Duration, Instant};
        // The screen runs on its own thread, so a square-time regression fails after the bound
        // instead of holding the test for minutes (the stuck thread dies with the test process).
        fn timed(text: String) -> Result<(usize, Duration), String> {
            let (tx, rx) = std::sync::mpsc::channel();
            let started = Instant::now();
            std::thread::spawn(move || {
                let _ = tx.send(secret_words(&text));
            });
            rx.recv_timeout(Duration::from_secs(15))
                .map(|words| (words, started.elapsed()))
                .map_err(|_| "no answer within 15 s".to_string())
        }
        let limit = crate::ledger::workflow::MAX_NOTE_BYTES;
        let key = j(&["AK", "IA", "ABCDEFGHIJKLMNOP"]);
        let units = ["a,", "a:", "a=", "{\"a\":1,\"b\":[2,3]},"];
        for size in [200_000, limit - 1, limit] {
            for unit in units {
                let plain = unit.repeat(size / unit.len() + 1)[..size].to_string();
                match timed(plain.clone()) {
                    Ok((words, _)) => assert_eq!(words, 0, "{unit} x {size}"),
                    Err(e) => panic!("{unit} x {size} bytes: {e}"),
                }
                // The shape at the very end of a long word is still found (two sizes are enough).
                if size != limit - 1 {
                    let tail = format!("{},{key}", &plain[..size - key.len() - 1]);
                    match timed(tail) {
                        Ok((words, _)) => assert_eq!(words, 1, "{unit} x {size} with a shape"),
                        Err(e) => panic!("{unit} x {size} bytes with a shape: {e}"),
                    }
                }
            }
        }
    }

    /// REQ-6: the refusal names the field and does not repeat the text.
    #[test]
    fn secret_screen_messages_name_the_field() {
        let m = InputError::CriterionSecretShaped { entry: 2, words: 3 }.to_string();
        assert!(m.starts_with("criterion 2 has 3 word(s)"), "{m}");
        assert!(InputError::TitleSecretShaped { words: 1 }
            .to_string()
            .starts_with("the title has 1 word(s)"));
        assert!(InputError::BodySecretShaped { words: 1 }
            .to_string()
            .starts_with("the body has 1 word(s)"));
    }

    /// REQ-5: the three add fields refuse before anything is stored.
    #[test]
    fn secret_screen_guards_title_body_and_criteria() {
        let bad = j(&["run --pass", "word=x"]);
        let criteria = [String::from("fine"), bad.clone()];
        let input = AddInput {
            criteria: &criteria,
            ..add("t")
        };
        assert_eq!(
            validate_add(&input),
            Err(InputError::CriterionSecretShaped { entry: 2, words: 1 })
        );
        assert_eq!(
            validate_add(&add(&bad)),
            Err(InputError::TitleSecretShaped { words: 1 })
        );
        let input = AddInput {
            body: &bad,
            ..add("t")
        };
        assert_eq!(
            validate_add(&input),
            Err(InputError::BodySecretShaped { words: 1 })
        );
    }
}
