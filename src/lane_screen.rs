//! What a lane's rendered screen says about the lane (SPEC-20261119-lane-screen-guard, card
//! C-120): a state, the one row it was read from, and for a usage limit the reset time in UTC.
//!
//! Two accidents of 2026-11-19 gave this its rules. A text with Enter was sent into a lane
//! whose screen showed a Yes/No confirmation: the highlighted `Yes` ran `git reset --hard`.
//! And eight lanes stopped at a usage limit; the reset time on a glm lane's screen,
//! `Your limit will reset at 2026-11-19 17:30:55`, has no time zone and was UTC+8, but was read
//! as UTC, so the renew request went out an hour early.
//!
//! So a state that lets a text be sent (`working`, `idle`) needs enough evidence: at least
//! [`MIN_ROWS`] non-blank rows and its own marker; a state that refuses (`permission_prompt`,
//! `input_pending`, `usage_limit`) needs only its own marker. Anything else is `unknown`, never
//! `idle`. A time without a zone is never taken as local time or as UTC.
//!
//! Pure functions over text: no file, process or terminal is touched here except the optional
//! read of the lane's usage record by the caller through [`crate::usage_limit::evaluate`].

use crate::lane_status::PaneState;

/// Fewer non-blank rows than this never give `working` or `idle`.
pub const MIN_ROWS: usize = 6;

/// The state of one lane screen. `PaneState` names five of these; `Idle` and `Unknown` have no
/// name there.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum State {
    Working,
    Idle,
    PermissionPrompt,
    InputPending,
    UsageLimit,
    Unreachable,
    Unknown,
}

impl State {
    pub fn as_str(self) -> &'static str {
        match self {
            State::Working => "working",
            State::Idle => "idle",
            State::PermissionPrompt => "permission_prompt",
            State::InputPending => "input_pending",
            State::UsageLimit => "usage_limit",
            State::Unreachable => "unreachable",
            State::Unknown => "unknown",
        }
    }

    /// The `lane-status` name of this state, `None` for `Idle` and `Unknown`.
    pub fn pane_state(self) -> Option<PaneState> {
        match self {
            State::Working => Some(PaneState::Working),
            State::PermissionPrompt => Some(PaneState::PermissionPrompt),
            State::InputPending => Some(PaneState::InputPending),
            State::UsageLimit => Some(PaneState::UsageLimit),
            State::Unreachable => Some(PaneState::Unreachable),
            State::Idle | State::Unknown => None,
        }
    }
}

/// A classification: the state, the one screen row it was read from (cleaned and cut), and the
/// reason in a fixed phrase.
#[derive(Debug, PartialEq, Eq)]
pub struct Seen {
    pub state: State,
    pub evidence: Option<String>,
    pub reason: &'static str,
    /// Non-blank rows on the screen.
    pub rows: usize,
}

/// Longest evidence row, in characters; a longer row is cut and ends with `…`.
const EVIDENCE_CHARS: usize = 160;

/// `row` without ANSI escape sequences and control characters, a tab as a space, and trimmed.
fn clean(row: &str) -> String {
    let mut out = String::with_capacity(row.len());
    let mut chars = row.chars().peekable();
    while let Some(c) = chars.next() {
        if c == '\u{1b}' {
            if chars.peek() == Some(&'[') {
                chars.next();
                // A CSI sequence ends at a byte in 0x40..=0x7e.
                for n in chars.by_ref() {
                    if ('\u{40}'..='\u{7e}').contains(&n) {
                        break;
                    }
                }
            }
            continue;
        }
        if c == '\t' {
            out.push(' ');
        } else if !c.is_control() {
            out.push(c);
        }
    }
    out.trim().to_string()
}

/// Text that came from a terminal this program does not own, for the answer: no character
/// [`crate::lane_duty::unsafe_char`] names (a bidirectional or zero-width one included, removed
/// first so it cannot split a secret shape), and every credential-shaped word replaced by
/// `[redacted]` (re-audit C-B1 and C-N1, note 3497: an evidence row, a reset text, a title or an
/// error message of another terminal may carry a key, and the answer goes into a session record).
/// Two shape filters run, the word filter for shapes such as `AKIA…` and the substring filter for
/// `sk-`, `ghp_`, `password=` and the like; a word such as `task-list` is shortened too, which a
/// list of terminals can bear and a leaked key cannot.
fn safe_text(text: &str) -> String {
    let plain: String = text
        .chars()
        .filter(|c| !crate::lane_duty::unsafe_char(*c))
        .collect();
    let (words, _) = crate::redact::redact_outgoing(&plain);
    crate::redact::redact_secret_shapes(&words)
}

/// `row` made safe and cut to [`EVIDENCE_CHARS`] characters.
fn cut(row: &str) -> String {
    cut_to(&safe_text(row), EVIDENCE_CHARS)
}

/// `row` cut to `max` characters; a longer row ends with `…`.
fn cut_to(row: &str, max: usize) -> String {
    if row.chars().count() <= max {
        return row.to_string();
    }
    let mut s: String = row.chars().take(max).collect();
    s.push('…');
    s
}

/// How many of the last non-blank rows may hold a confirmation question (R-N1).
const BOTTOM_ROWS: usize = 8;

/// How many of the last non-blank rows may hold the entries of a choice list that has no
/// question above it (`[1] Yes`, `[2] No`).
const LIST_ROWS: usize = 6;

/// Words of a confirmation question, lower case.
const QUESTIONS: [&str; 6] = [
    "do you want",
    "would you like",
    "are you sure",
    "proceed?",
    "[y/n]",
    "(y/n)",
];

/// A choice row is short: at most this many characters once its marks are cut off.
const SHORT_ROW: usize = 40;

/// The longest word of a choice (`Approve`, `Continue`, `Decline`).
const CHOICE_WORD_CHARS: usize = 12;

/// The longest token of a bracket question group (`y`, `N`, `a`, `yes`, `no`).
const BRACKET_TOKEN_CHARS: usize = 3;

/// A row without the marks in front of it (frame, cursor, bullet): it starts at the first letter,
/// digit or bracket, and its closing frame is cut off.
fn body(row: &str) -> &str {
    row.trim_start_matches(|c: char| !c.is_alphanumeric() && c != '[' && c != '(')
        .trim_end_matches(|c: char| c.is_whitespace() || matches!(c, '│' | '┃' | '║'))
}

/// Whether `word` is one alphabetic word of at most `max` characters, a closing `?`, `.` or `:`
/// aside.
fn alpha_word(word: &str, max: usize) -> bool {
    let word = word.trim().trim_end_matches(['?', '.', ':']);
    !word.is_empty() && word.chars().count() <= max && word.chars().all(char::is_alphabetic)
}

/// Whether `text` is one to three short alphabetic words.
fn few_words(text: &str) -> bool {
    let words = text.split_whitespace().count();
    (1..=3).contains(&words)
        && text
            .split_whitespace()
            .all(|w| alpha_word(w, CHOICE_WORD_CHARS))
}

/// A bracket question group: a `(…)` or `[…]` that holds two or more tokens of at most three
/// letters split by `/` or `|`, as in `(y/N/a)`, `[y/n]`, `(yes/no)`, `(yes/no/[fingerprint])`.
fn bracket_question(row: &str) -> bool {
    row.match_indices(['(', '[']).any(|(at, _)| {
        let inside = row[at + 1..]
            .split([')', ']', '(', '['])
            .next()
            .unwrap_or("");
        let tokens: Vec<&str> = inside
            .split(['/', '|'])
            .map(str::trim)
            .filter(|t| !t.is_empty())
            .collect();
        tokens.len() >= 2 && tokens.iter().all(|t| alpha_word(t, BRACKET_TOKEN_CHARS))
    })
}

/// A short row that opens with a bracket the way a button or a list entry does: `[1] Yes`,
/// `[ Proceed ] [ Cancel ]`, `[Y]es [N]o`, `(a) Keep`.
fn bracketed(body: &str) -> bool {
    let close = match body.chars().next() {
        Some('[') => (']', CHOICE_WORD_CHARS),
        Some('(') => (')', BRACKET_TOKEN_CHARS),
        _ => return false,
    };
    body.chars().count() <= SHORT_ROW
        && body[1..]
            .split_once(close.0)
            .is_some_and(|(inside, _)| inside.trim().chars().count() <= close.1)
}

/// A short row that opens with a list marker: `1: Yes`, `a. Yes`, `a) Yes` and, with
/// `numbered_too`, the `1. Yes` and `2) No` rows [`numbered`] reads.
fn marked(body: &str, numbered_too: bool) -> bool {
    if body.chars().count() > SHORT_ROW {
        return false;
    }
    let digits = body.chars().take_while(char::is_ascii_digit).count();
    let (marker, rest) = if (1..=2).contains(&digits) {
        (true, &body[digits..])
    } else if body.chars().next().is_some_and(|c| c.is_ascii_alphabetic()) {
        (false, &body[1..])
    } else {
        return false;
    };
    let Some(after) = rest.strip_prefix(['.', ')', ':']) else {
        return false;
    };
    let colon = rest.starts_with(':');
    (after.is_empty() || after.starts_with(' ')) && (!marker || colon || numbered_too)
}

/// A short row of two or more words split by `/`, `|` or `or`: `Approve / Reject`, `Allow /
/// Deny`, `Yes/No`, `choose: yes or no`.
fn split_choice(body: &str) -> bool {
    if body.chars().count() > SHORT_ROW {
        return false;
    }
    if body.contains(['/', '|']) {
        let parts: Vec<&str> = body.split(['/', '|']).collect();
        return parts.len() >= 2 && parts.iter().all(|p| few_words(p));
    }
    body.split_once(" or ").is_some_and(|(left, right)| {
        let last = left.split_whitespace().next_back().unwrap_or("");
        let first = right.split_whitespace().next().unwrap_or("");
        alpha_word(last, CHOICE_WORD_CHARS) && alpha_word(first, CHOICE_WORD_CHARS)
    })
}

/// Whether a (cleaned) row is choice-shaped: a bracket question, or a short row that is a button
/// or a list entry or a split of short words. The shape is the criterion, not a list of words
/// (REQ-6 of SPEC-20261120-lane-screen-followups).
fn choice_shaped(row: &str) -> bool {
    let body = body(row);
    bracket_question(row) || bracketed(body) || marked(body, true) || split_choice(body)
}

/// Whether a row is a list entry the shape reader newly knows (no question needed): `[1] Yes`,
/// `1: Yes`, `a. Yes`. The `1.` and `1)` rows are the reader of [`numbered`].
fn list_entry(row: &str) -> bool {
    let body = body(row);
    bracketed(body) || marked(body, false)
}

/// Whether two or more rows in a row are one bare word each (`Allow`, `Deny`): a list of choices
/// drawn one to a line.
fn word_rows(rows: &[String]) -> bool {
    let bare: Vec<bool> = rows
        .iter()
        .filter(|r| !r.is_empty())
        .map(|r| alpha_word(body(r), CHOICE_WORD_CHARS))
        .collect();
    bare.windows(2).any(|pair| pair[0] && pair[1])
}

/// The cursor a menu draws before the highlighted choice: `❯` is Claude Code's own, `>` is its
/// plain-text form, the others are the arrows other menus use. Only a dialog with two or more
/// choice rows and one of these in front of a row is `input_pending`; every other mark in front
/// of a number is skipped without being a cursor.
const CURSORS: [char; 8] = ['❯', '>', '›', '▶', '▸', '→', '➜', '➤'];

/// A choice row `❯ 2. text`: whether a cursor from [`CURSORS`] is on it, its number and its
/// text. Everything before the first letter or digit is a mark and is skipped, whatever it is
/// (a frame, a cursor, a bullet, a glyph nobody listed): audit note 3212 B1 (`> 1. Yes` and
/// `│ ❯ 1. Yes │`) and audit note 3362 B1 (a cursor outside the list). A letter or a digit
/// before the number makes the row a sentence, not a choice.
fn numbered(row: &str) -> Option<(bool, u32, &str)> {
    let start = row.find(char::is_alphanumeric)?;
    let (marks, rest) = row.split_at(start);
    let cursor = marks.chars().any(|c| CURSORS.contains(&c));
    let digits = rest.chars().take_while(char::is_ascii_digit).count();
    if digits == 0 || digits > 2 {
        return None;
    }
    let (n, after) = rest.split_at(digits);
    let after = after
        .strip_prefix('.')
        .or_else(|| after.strip_prefix(')'))?;
    let text = after.strip_prefix(' ')?.trim_start();
    Some((cursor, n.parse().ok()?, text))
}

/// Whether `text` starts with `word` as a whole word, ignoring case: `Yes, and …` and `No (esc)`
/// do, `Yesterday` and `Not now` do not.
fn starts_with_word(text: &str, word: &str) -> bool {
    text.to_ascii_lowercase()
        .strip_prefix(word)
        .is_some_and(|after| !after.starts_with(char::is_alphanumeric))
}

/// Whether a (cleaned) row is a usage-limit message.
fn is_limit_row(row: &str) -> bool {
    let low = row.to_ascii_lowercase();
    (low.contains("limit") && low.contains("reset")) || low.contains("limit reached")
}

/// The state of `rows` (one string per screen row).
///
/// Order: a Yes/No confirmation, a usage limit, another choice dialog (all three refuse a text),
/// then, only with [`MIN_ROWS`] non-blank rows, working and idle; anything else is unknown.
pub fn classify(rows: &[String]) -> Seen {
    let rows: Vec<String> = rows.iter().map(|r| clean(r)).collect();
    let non_blank = rows.iter().filter(|r| !r.is_empty()).count();
    let seen = |state: State, row: Option<&String>, reason: &'static str| Seen {
        state,
        evidence: row.map(|r| cut(r)),
        reason,
        rows: non_blank,
    };
    let lower: Vec<String> = rows.iter().map(|r| r.to_ascii_lowercase()).collect();
    let find = |needle: &str| lower.iter().position(|r| r.contains(needle));

    let choices: Vec<(usize, bool, &str)> = rows
        .iter()
        .enumerate()
        .filter_map(|(i, r)| numbered(r).map(|(cursor, _, text)| (i, cursor, text)))
        .collect();
    let choice_row = |word: &str| {
        choices
            .iter()
            .find(|(_, _, t)| starts_with_word(t, word))
            .map(|(i, _, _)| *i)
    };
    // A numbered Yes row together with a numbered No row is a confirmation, with or without the
    // `Do you want` question and whatever draws its cursor or frame: Enter would pick Yes.
    if let (Some(yes), Some(_)) = (choice_row("yes"), choice_row("no")) {
        return match find("do you want") {
            Some(q) => seen(
                State::PermissionPrompt,
                rows.get(q),
                "a Yes/No confirmation is on the screen",
            ),
            None => seen(
                State::PermissionPrompt,
                rows.get(yes),
                "numbered Yes and No rows are on the screen",
            ),
        };
    }
    if let Some(i) = rows.iter().position(|r| is_limit_row(r)) {
        return seen(
            State::UsageLimit,
            rows.get(i),
            "a usage-limit message is on the screen",
        );
    }
    if choices.len() >= 2 {
        if let Some((i, _, _)) = choices.iter().find(|(_, cursor, _)| *cursor) {
            return seen(
                State::InputPending,
                rows.get(*i),
                "a choice dialog is waiting for a key",
            );
        }
    }
    for footer in [
        "enter to confirm",
        "enter to select",
        "esc to cancel",
        "esc to exit",
    ] {
        if let Some(i) = find(footer) {
            return seen(
                State::InputPending,
                rows.get(i),
                "a choice dialog is waiting for a key",
            );
        }
    }
    if non_blank < MIN_ROWS {
        return seen(
            State::Unknown,
            None,
            "fewer than six non-blank rows; working and idle need six",
        );
    }
    // A confirmation question near the bottom with its choices on the question row itself or
    // below it, and no numbered Yes and No rows the reader could use (the choices are drawn in a
    // way it does not know, or in other words): Enter may pick the first choice, so the lane is
    // not `working` or `idle` whatever its footer says (leader R-N1, note 3256). The choices
    // are there when a choice-shaped row (`choice_shaped`: a bracket question, a button or list
    // entry, a split of short words) is on the question row or below it (R-B1, note 3433:
    // `proceed? (yes/no)` is one row), when two or more numbered rows are below it whatever they
    // say (R-N1, note 3433: `1. Allow once`, `3. Deny`), or when two bare words stand one to a
    // row (`Allow`, `Deny`). A bracket question such as `Overwrite? (y/N/a)` is its own question.
    // A question with only the input line under it is a lane that ended its turn and waits: that
    // is the screen a send is for (leader S-N2, note 3362). Without a question, two list entries
    // of the newer shapes (`[1] Yes`, `1: Yes`, `a. Yes`) in the window are a choice list too.
    let near_bottom: Vec<(usize, &String)> = rows
        .iter()
        .enumerate()
        .filter(|(_, r)| !r.is_empty())
        .rev()
        .take(BOTTOM_ROWS)
        .collect();
    if let Some((i, _)) = near_bottom.iter().find(|(i, row)| {
        let known = QUESTIONS.iter().any(|q| lower[*i].contains(q)) || bracket_question(row);
        known
            && (rows[*i..].iter().any(|r| choice_shaped(r))
                || word_rows(&rows[*i + 1..])
                || rows[*i + 1..]
                    .iter()
                    .filter(|r| numbered(r).is_some())
                    .count()
                    >= 2)
    }) {
        return seen(
            State::Unknown,
            rows.get(*i),
            "a question is near the bottom of the screen and its choices could not be read as Yes and No rows",
        );
    }
    // An unanswered dialog is the last thing drawn: entries with more than `LIST_ROWS` rows of
    // output under them are not the live prompt.
    let entries: Vec<usize> = near_bottom
        .iter()
        .take(LIST_ROWS)
        .rev()
        .filter(|(_, r)| list_entry(r))
        .map(|(i, _)| *i)
        .collect();
    if let [first, _, ..] = entries[..] {
        return seen(
            State::Unknown,
            rows.get(first),
            "a list of choices is near the bottom of the screen and could not be read as Yes and No rows",
        );
    }
    if let Some(i) = find("esc to interrupt") {
        return seen(
            State::Working,
            rows.get(i),
            "an esc-to-interrupt row is on the screen",
        );
    }
    let prompt = rows.iter().any(|r| r.starts_with('❯'));
    let footer = find("shift+tab to cycle").or_else(|| find("? for shortcuts"));
    if let (true, Some(i)) = (prompt, footer) {
        return seen(
            State::Idle,
            rows.get(i),
            "an input prompt and its footer are on the screen",
        );
    }
    seen(State::Unknown, None, "no state marker is on the screen")
}

/// A usage limit's reset time.
#[derive(Debug, PartialEq, Eq)]
pub enum Reset {
    /// Unix seconds, shown in UTC.
    Utc(i64),
    /// The time cannot be told in UTC; `raw` is the message row as shown.
    Unknown { raw: String, why: &'static str },
}

/// The reset time of the usage-limit message in `rows`, with `now` (Unix seconds) to place a
/// time that shows no date or year. `None` when no row is a usage-limit message.
pub fn reset_of(rows: &[String], now: i64) -> Option<Reset> {
    rows.iter()
        .map(|r| clean(r))
        .find(|r| is_limit_row(r))
        .map(|r| parse_reset(&r, now))
}

const MONTHS: [&str; 12] = [
    "jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec",
];

/// Days since 1970-01-01 of a civil date (proleptic Gregorian).
fn days_from_civil(y: i64, m: i64, d: i64) -> i64 {
    let y = if m <= 2 { y - 1 } else { y };
    let era = y.div_euclid(400);
    let yoe = y - era * 400;
    let doy = (153 * (if m > 2 { m - 3 } else { m + 9 }) + 2) / 5 + d - 1;
    let doe = yoe * 365 + yoe / 4 - yoe / 100 + doy;
    era * 146_097 + doe - 719_468
}

/// The civil year of a day count since 1970-01-01.
fn year_of_days(days: i64) -> i64 {
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z - era * 146_097;
    let yoe = (doe - doe / 1_460 + doe / 36_524 - doe / 146_096) / 365;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    yoe + era * 400 + i64::from(m <= 2)
}

fn days_in_month(y: i64, m: i64) -> i64 {
    match m {
        1 | 3 | 5 | 7 | 8 | 10 | 12 => 31,
        4 | 6 | 9 | 11 => 30,
        _ if (y % 4 == 0 && y % 100 != 0) || y % 400 == 0 => 29,
        _ => 28,
    }
}

/// A time of day in seconds, from `H:MM[:SS]` (24 hours) or `H[:MM]` with `am`/`pm`.
/// `meridiem` is `Some(true)` for pm. `None` when it is not a time of day.
fn time_of_day(text: &str, meridiem: Option<bool>) -> Option<i64> {
    let mut parts = text.split(':');
    let h: i64 = parts.next()?.parse().ok()?;
    let m: i64 = parts.next().map_or(Some(0), |p| p.parse().ok())?;
    let s: i64 = parts.next().map_or(Some(0), |p| p.parse().ok())?;
    if parts.next().is_some() || !(0..60).contains(&m) || !(0..60).contains(&s) {
        return None;
    }
    let h = match meridiem {
        Some(pm) => {
            if !(1..=12).contains(&h) {
                return None;
            }
            h % 12 + if pm { 12 } else { 0 }
        }
        None if (0..24).contains(&h) => h,
        None => return None,
    };
    Some(h * 3600 + m * 60 + s)
}

/// A token as a time of day: `7pm`, `7:30am`, `19:00`, `17:30:55`; the rest of the token after
/// the time (a zone glued on, as in `17:30:55Z`) comes back too.
fn time_token(token: &str) -> Option<(Option<i64>, &str)> {
    let end = token
        .find(|c: char| !(c.is_ascii_digit() || c == ':'))
        .unwrap_or(token.len());
    let (num, rest) = token.split_at(end);
    if num.is_empty() || !num.chars().next()?.is_ascii_digit() {
        return None;
    }
    let low = rest.to_ascii_lowercase();
    for (mark, pm) in [("am", false), ("pm", true)] {
        if let Some(after) = low.strip_prefix(mark) {
            let tail = &rest[rest.len() - after.len()..];
            return Some((time_of_day(num, Some(pm)), tail));
        }
    }
    // A bare number needs a colon to be a time: `7` alone is not.
    if !num.contains(':') {
        return None;
    }
    Some((time_of_day(num, None), rest))
}

/// The offset of a zone in seconds east of UTC, or `None` for an unknown zone.
fn zone_offset(text: &str) -> Option<i64> {
    let z = text
        .trim()
        .trim_matches(|c| c == '(' || c == ')' || c == '[' || c == ']');
    let z = z.trim().to_ascii_lowercase();
    match z.as_str() {
        "utc" | "gmt" | "z" => return Some(0),
        "asia/seoul" | "asia/tokyo" => return Some(9 * 3600),
        "asia/shanghai" | "asia/hong_kong" | "asia/singapore" => return Some(8 * 3600),
        "asia/kolkata" | "asia/calcutta" => return Some(5 * 3600 + 1800),
        _ => {}
    }
    let signed = z
        .strip_prefix("utc")
        .or_else(|| z.strip_prefix("gmt"))
        .unwrap_or(&z);
    if !signed.is_ascii() {
        return None;
    }
    let (sign, digits) = if let Some(d) = signed.strip_prefix('+') {
        (1, d)
    } else if let Some(d) = signed.strip_prefix('-') {
        (-1, d)
    } else {
        return None;
    };
    let (h, m): (i64, i64) = if let Some((h, m)) = digits.split_once(':') {
        (h.parse().ok()?, m.parse().ok()?)
    } else if digits.len() == 4 {
        (digits[..2].parse().ok()?, digits[2..].parse().ok()?)
    } else {
        (digits.parse().ok()?, 0)
    };
    if !digits.chars().all(|c| c.is_ascii_digit() || c == ':') || h > 14 || !(0..60).contains(&m) {
        return None;
    }
    Some(sign * (h * 3600 + m * 60))
}

/// The reset time one message row states.
pub fn parse_reset(row: &str, now: i64) -> Reset {
    let raw = clean(row);
    let unknown = |why: &'static str| Reset::Unknown {
        raw: raw.clone(),
        why,
    };
    let Some(at) = raw.to_ascii_lowercase().find("reset") else {
        return unknown("the row has no reset time");
    };
    let mut rest = raw[at + "reset".len()..]
        .trim_start_matches('s')
        .trim_start();
    if rest.get(..3).is_some_and(|w| w.eq_ignore_ascii_case("at ")) {
        rest = rest[3..].trim_start();
    }
    let tokens: Vec<&str> = rest
        .split_whitespace()
        .map(|t| t.trim_matches(|c: char| "(),;.∙·".contains(c)))
        .filter(|t| !t.is_empty())
        .collect();

    // The date and the time: (year, month, day), seconds of the day, tokens used.
    let mut date: Option<(Option<i64>, i64, i64)> = None;
    let mut tod: Option<Option<i64>> = None;
    let mut glued = "";
    let mut i = 0;
    let first = tokens.first().copied().unwrap_or("");
    let (iso_date, iso_time) = first.split_once('T').unwrap_or((first, ""));
    let iso: Vec<&str> = iso_date.split('-').collect();
    if iso.len() == 3
        && iso
            .iter()
            .all(|p| !p.is_empty() && p.bytes().all(|b| b.is_ascii_digit()))
    {
        let (y, m, d) = (
            iso[0].parse().unwrap_or(0),
            iso[1].parse().unwrap_or(0),
            iso[2].parse().unwrap_or(0),
        );
        date = Some((Some(y), m, d));
        i = 1;
        if !iso_time.is_empty() {
            if let Some((t, tail)) = time_token(iso_time) {
                tod = Some(t);
                glued = tail;
            }
        } else if let Some((t, tail)) = tokens.get(1).and_then(|t| time_token(t)) {
            tod = Some(t);
            glued = tail;
            i = 2;
        }
    } else if let Some(month) = MONTHS
        .iter()
        .position(|m| first.get(..3).is_some_and(|w| w.eq_ignore_ascii_case(m)))
        .filter(|_| first.chars().all(|c| c.is_ascii_alphabetic()))
    {
        let day: Option<i64> = tokens.get(1).and_then(|t| {
            t.trim_end_matches(|c: char| c.is_ascii_alphabetic())
                .parse()
                .ok()
        });
        if let Some(day) = day {
            i = 2;
            let mut year = None;
            if let Some(y) = tokens
                .get(i)
                .filter(|t| t.len() == 4 && t.bytes().all(|b| b.is_ascii_digit()))
            {
                year = y.parse().ok();
                i += 1;
            }
            date = Some((year, month as i64 + 1, day));
            if tokens.get(i).is_some_and(|t| t.eq_ignore_ascii_case("at")) {
                i += 1;
            }
            i = read_time(&tokens, i, &mut tod, &mut glued);
        }
    } else {
        i = read_time(&tokens, 0, &mut tod, &mut glued);
    }
    let zone_from = i;
    let Some(tod) = tod else {
        return unknown("no time of day after the reset words");
    };
    let Some(tod) = tod else {
        return unknown("the time of day is impossible");
    };
    let zone_text = format!(
        "{glued} {}",
        tokens[zone_from.min(tokens.len())..].join(" ")
    );
    let zone_text = zone_text.trim();
    if zone_text.is_empty() {
        return unknown("the time shows no time zone, so it is neither UTC nor local time");
    }
    let Some(offset) = zone_offset(zone_text) else {
        return unknown("unknown zone name; only UTC, UTC offsets and a few Asia zones are read");
    };
    let at_zone = |y: i64, m: i64, d: i64| -> Option<i64> {
        // A year outside 1970..=9999 is no reset time (audit note 3212 N1: a huge year made the
        // day count overflow). It also keeps `days_from_civil` far from the `i64` limits.
        if !(1970..=9999).contains(&y) || !(1..=12).contains(&m) || d < 1 || d > days_in_month(y, m)
        {
            return None;
        }
        // The range holds for the time in UTC too (card C-128 R-N4): `9999-12-31 23:59 -0800` is
        // the year 10000 in UTC, and `1970-01-01 07:59 +0800` is 1969.
        let t = days_from_civil(y, m, d) * 86_400 + tod - offset;
        (1970..=9999)
            .contains(&year_of_days(t.div_euclid(86_400)))
            .then_some(t)
    };
    let local_now = now + offset;
    let found = match date {
        Some((Some(y), m, d)) => at_zone(y, m, d),
        Some((None, m, d)) => {
            let this_year = year_of_days(local_now.div_euclid(86_400));
            (this_year..this_year + 9)
                .filter_map(|y| at_zone(y, m, d))
                .find(|t| *t >= now)
        }
        None => {
            let t = local_now.div_euclid(86_400) * 86_400 + tod - offset;
            Some(if t < now { t + 86_400 } else { t })
        }
    };
    match found {
        Some(t) => Reset::Utc(t),
        None => unknown("the date is impossible"),
    }
}

/// Reads a time of day at `tokens[i]` (`7:30pm`, `7 pm`, `19:00`); returns the next index.
fn read_time<'a>(
    tokens: &[&'a str],
    i: usize,
    tod: &mut Option<Option<i64>>,
    glued: &mut &'a str,
) -> usize {
    let Some(token) = tokens.get(i) else {
        return i;
    };
    if let Some((t, tail)) = time_token(token) {
        *tod = Some(t);
        *glued = tail;
        return i + 1;
    }
    // `7 pm`: a bare number, then am or pm.
    let meridiem = tokens.get(i + 1).map(|m| m.to_ascii_lowercase());
    if let Some(pm) = meridiem.as_deref().and_then(|m| match m {
        "am" => Some(false),
        "pm" => Some(true),
        _ => None,
    }) {
        if token.bytes().all(|b| b.is_ascii_digit() || b == b':') && !token.is_empty() {
            *tod = Some(time_of_day(token, Some(pm)));
            return i + 2;
        }
    }
    i
}

/// The rows of a screen read, or the verdict the input itself settles.
#[derive(Debug, PartialEq, Eq)]
pub struct Screen {
    pub rows: Vec<String>,
    /// `orca:screen` for an Orca read, `stdin:text` for plain text.
    pub source: &'static str,
    /// Set when the input alone decides the state (a terminal that is gone, a read that is not
    /// the rendered screen).
    pub verdict: Option<(State, &'static str)>,
}

/// Reads `input`: an Orca `terminal read --screen --json` answer, or plain screen text.
pub fn read_input(input: &str) -> Screen {
    let text = || Screen {
        rows: input.lines().map(String::from).collect(),
        source: "stdin:text",
        verdict: None,
    };
    let Ok(serde_json::Value::Object(read)) = serde_json::from_str::<serde_json::Value>(input)
    else {
        return text();
    };
    // A JSON object that is not an Orca answer (no `ok`, no `result`) is just text.
    if !read.contains_key("ok") && !read.contains_key("result") {
        return text();
    }
    let gone = |reason: &'static str| Screen {
        rows: Vec::new(),
        source: "orca:screen",
        verdict: Some((State::Unreachable, reason)),
    };
    let not_screen = |reason: &'static str| Screen {
        rows: Vec::new(),
        source: "orca:screen",
        verdict: Some((State::Unknown, reason)),
    };
    if read.get("ok") == Some(&serde_json::Value::Bool(false)) {
        return gone("the terminal read failed");
    }
    let Some(terminal) = read
        .get("result")
        .and_then(|r| r.get("terminal"))
        .and_then(serde_json::Value::as_object)
    else {
        return not_screen("the read holds no terminal, so no screen was seen");
    };
    if terminal.get("status").and_then(serde_json::Value::as_str) == Some("exited") {
        return gone("the terminal has exited");
    }
    if terminal.get("source").and_then(serde_json::Value::as_str) != Some("screen") {
        return not_screen("the read is not the rendered screen (its source is not `screen`)");
    }
    let rows = terminal
        .get("tail")
        .and_then(serde_json::Value::as_array)
        .map(|rows| {
            rows.iter()
                .filter_map(serde_json::Value::as_str)
                .map(String::from)
                .collect()
        })
        .unwrap_or_default();
    Screen {
        rows,
        source: "orca:screen",
        verdict: None,
    }
}

/// Reads `input` and classifies it: the verdict the input itself settles, else the state of its
/// rows. Shared by `lane-screen` and `lane-send` so both read a screen the same way.
pub fn observe(input: &str) -> (Screen, Seen) {
    let screen = read_input(input);
    let non_blank = screen.rows.iter().filter(|r| !clean(r).is_empty()).count();
    let seen = match screen.verdict {
        Some((state, reason)) => Seen {
            state,
            evidence: None,
            reason,
            rows: non_blank,
        },
        None => classify(&screen.rows),
    };
    (screen, seen)
}

/// The `reset` object of an answer for a reset time read from the screen text.
fn reset_json(found: &Reset) -> serde_json::Value {
    use crate::forge::policy::format_utc;
    match found {
        Reset::Utc(t) => serde_json::json!({"utc": format_utc(*t), "source": "screen"}),
        Reset::Unknown { raw, why } => {
            serde_json::json!({"utc": null, "unknown": true, "why": why, "raw": cut(raw)})
        }
    }
}

/// The answer of `moai lane-screen`: the state, its evidence row and reason, the reset time of a
/// usage limit (from the record of the lane's `home` in the cache folder for `tmpdir`, the
/// `TMPDIR` value, when `home` is given), the `pane` object `lane-status` takes, and the count of
/// non-blank rows. `now` is Unix seconds.
pub fn report(
    input: &str,
    home: Option<&std::path::Path>,
    tmpdir: Option<&str>,
    now: i64,
) -> serde_json::Value {
    use crate::forge::policy::{format_utc, parse_utc};
    use serde_json::{json, Value};

    let (screen, seen) = observe(input);

    // The reset time of a usage limit: the lane's own record when a window is at its limit,
    // else the screen's text.
    let mut reset = Value::Null;
    let mut until: Option<String> = None;
    if seen.state == State::UsageLimit {
        let from_home = home
            .map(|h| crate::usage_limit::evaluate(h, tmpdir, now))
            .filter(|v| v["state"] == "reached")
            .and_then(|v| v["resets_at"].as_str().map(String::from))
            .filter(|t| parse_utc(t).is_some());
        if let Some(utc) = from_home {
            until = Some(utc.clone());
            reset = json!({"utc": utc, "source": "home"});
        } else {
            if let Some(found) = reset_of(&screen.rows, now) {
                if let Reset::Utc(t) = &found {
                    until = Some(format_utc(*t));
                }
                reset = reset_json(&found);
            }
        }
    }
    let pane = match seen.state.pane_state() {
        Some(_) => json!({
            "state": seen.state.as_str(),
            "source": screen.source,
            "observed_at": format_utc(now),
            "until": until,
        }),
        None => Value::Null,
    };
    json!({
        "state": seen.state.as_str(),
        "evidence": seen.evidence,
        "reason": seen.reason,
        "rows": seen.rows,
        "reset": reset,
        "pane": pane,
    })
}

/// Most terminals one scan reads; a longer list is read up to here and reported as cut short.
pub const MAX_SCAN: usize = 200;
/// Longest title, branch or handle shown, in characters.
const SHOWN_CHARS: usize = 80;

/// The limits of one scan (`--max-seconds`, `--max-terminals`).
#[derive(Clone, Copy, Debug, Default)]
pub struct ScanLimits {
    /// The instant after which no screen is read any more, and the seconds it was set from (the
    /// reasons name them).
    pub deadline: Option<(std::time::Instant, u64)>,
    /// The most terminals a list may hold; a longer list fails with no screen read.
    pub max_terminals: Option<usize>,
}

impl ScanLimits {
    /// The reason of a terminal the time cap left unread, or `None` while the time remains.
    fn expired(&self) -> Option<String> {
        let (at, secs) = self.deadline?;
        (std::time::Instant::now() >= at).then(|| time_cap_text(secs))
    }
}

/// How the reasons name the total time cap.
fn time_cap_text(secs: u64) -> String {
    format!("the scan's time cap of --max-seconds {secs} passed")
}

/// The error text for a terminal list that cannot be read (the program is absent, it failed, or
/// its answer is not a list).
pub fn list_error(detail: &str) -> String {
    format!(
        "the terminal list could not be read: {}",
        cut(&clean(detail))
    )
}

/// A text somebody else chose (a terminal title or branch), for display: no escape or control
/// characters, no line-separator or bidirectional character, no credential-shaped word (see
/// [`safe_text`]), cut to [`SHOWN_CHARS`].
fn shown(text: &str) -> String {
    cut_to(&safe_text(&clean(text)), SHOWN_CHARS)
}

/// The scan of `orca terminal list --json` (`list_json`): one row per listed terminal with the
/// state of its screen, read one terminal at a time through `read` (the handle in, the answer of
/// `orca terminal read --screen --json` out). A terminal that cannot be read, has no handle or is
/// not answered with an Orca answer is `unreachable`; the scan goes on. A usage limit carries its
/// reset time in UTC from the screen text (no usage record: a terminal does not say whose
/// HOME it is). `now` is Unix seconds. `Err` only when the list itself cannot be read or holds
/// more terminals than `limits.max_terminals` (nothing is read then). Once `limits.deadline` has
/// passed, the terminals not read yet are `unreachable` and say so.
pub fn scan(
    list_json: &str,
    read: &mut dyn FnMut(&str) -> Result<String, String>,
    now: i64,
    limits: &ScanLimits,
) -> Result<serde_json::Value, String> {
    use serde_json::{json, Map, Value};

    let Ok(Value::Object(answer)) = serde_json::from_str::<Value>(list_json) else {
        return Err(list_error(
            "the answer of `orca terminal list` is not an Orca answer",
        ));
    };
    if answer.get("ok") == Some(&Value::Bool(false)) {
        return Err(list_error("Orca answered with an error"));
    }
    let Some(listed) = answer
        .get("result")
        .and_then(|r| r.get("terminals"))
        .and_then(Value::as_array)
    else {
        return Err(list_error("the answer holds no `terminals` list"));
    };
    let reported = answer
        .get("result")
        .and_then(|r| r.get("totalCount"))
        .and_then(Value::as_u64)
        .map_or(0, |n| usize::try_from(n).unwrap_or(usize::MAX));
    let total = reported.max(listed.len());
    if let Some(max) = limits.max_terminals.filter(|max| total > *max) {
        return Err(format!(
            "the terminal list holds {total} terminals, more than --max-terminals {max}; no screen was read"
        ));
    }
    let cut_short = total > listed.len()
        || listed.len() > MAX_SCAN
        || answer
            .get("result")
            .and_then(|r| r.get("truncated"))
            .and_then(Value::as_bool)
            == Some(true);

    let mut terminals = Vec::new();
    let mut by_state: Map<String, Value> = Map::new();
    let mut table = vec![format!(
        "{:<16} {:<18} {:<20} {}",
        "handle", "state", "reset (UTC)", "branch  title"
    )];
    for term in listed.iter().take(MAX_SCAN) {
        let text = |key: &str| term.get(key).and_then(Value::as_str).map(shown);
        let handle = text("handle");
        let (title, branch) = (
            text("title").unwrap_or_default(),
            text("branch").unwrap_or_default(),
        );
        // (state, reason, evidence, reset) of this terminal.
        let (state, reason, evidence, reset) = match term.get("handle").and_then(Value::as_str) {
            None => unreachable_row("the list gives this terminal no handle"),
            Some(_) if limits.expired().is_some() => unreachable_row(&format!(
                "the terminal was not read: {}",
                limits.expired().unwrap_or_default()
            )),
            Some(raw) => match read(raw) {
                Err(d) => unreachable_row(&match limits.expired() {
                    // The read ran out of the time that was left, or ended just after the cap.
                    Some(cap) => {
                        format!("the terminal could not be read: {cap}: {}", cut(&clean(&d)))
                    }
                    None => format!("the terminal could not be read: {}", cut(&clean(&d))),
                }),
                Ok(out) => {
                    let input = read_input(&out);
                    if input.source != "orca:screen" {
                        unreachable_row("the answer of `orca terminal read` is not an Orca answer")
                    } else {
                        let seen = report(&out, None, None, now);
                        let state = seen["state"].as_str().unwrap_or("unknown").to_string();
                        let mut reason = seen["reason"].as_str().unwrap_or("").to_string();
                        let mut reset = seen["reset"].clone();
                        // A Yes/No menu wins over a usage limit (R-N3): the lane is still
                        // stopped by the limit, so its reset time stays in the list.
                        if state == State::PermissionPrompt.as_str() && reset.is_null() {
                            if let Some(found) = reset_of(&input.rows, now) {
                                reset = reset_json(&found);
                                reason.push_str("; a usage-limit message is also on the screen");
                            }
                        }
                        (state, reason, seen["evidence"].clone(), reset)
                    }
                }
            },
        };
        let reset_text = match reset["utc"].as_str() {
            Some(utc) => utc,
            None if reset["unknown"] == Value::Bool(true) => "unknown",
            None => "-",
        };
        table.push(format!(
            "{:<16} {:<18} {:<20} {}  {}",
            handle.as_deref().unwrap_or("-"),
            state,
            reset_text,
            branch,
            title
        ));
        let n = by_state.get(&state).and_then(Value::as_u64).unwrap_or(0);
        by_state.insert(state.clone(), json!(n + 1));
        terminals.push(json!({
            "handle": handle,
            "title": title,
            "branch": branch,
            "state": state,
            "reason": reason,
            "evidence": evidence,
            "reset": reset,
        }));
    }
    let count = terminals.len();
    if cut_short {
        table.push(format!(
            "{count} of {total} terminals shown: the list was cut short"
        ));
    }
    Ok(json!({
        "terminals": terminals,
        "count": count,
        "total": total,
        "truncated": cut_short,
        "by_state": by_state,
        "table": table,
    }))
}

/// The row of a terminal that could not be looked at.
fn unreachable_row(reason: &str) -> (String, String, serde_json::Value, serde_json::Value) {
    (
        State::Unreachable.as_str().to_string(),
        reason.to_string(),
        serde_json::Value::Null,
        serde_json::Value::Null,
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    fn rows(text: &str) -> Vec<String> {
        text.lines().map(String::from).collect()
    }

    /// `n` non-blank filler rows, then `last`.
    fn filler(n: usize, last: &[&str]) -> Vec<String> {
        let mut v: Vec<String> = (0..n).map(|i| format!("row {i}")).collect();
        v.extend(last.iter().map(|s| s.to_string()));
        v
    }

    const NOW: i64 = 1_795_046_400; // 2026-11-19T00:00:00Z

    fn utc(row: &str, now: i64) -> Option<i64> {
        match parse_reset(row, now) {
            Reset::Utc(s) => Some(s),
            Reset::Unknown { .. } => None,
        }
    }

    #[test]
    fn a_choice_row_is_read_through_a_frame_and_any_cursor_glyph() {
        assert_eq!(numbered("❯ 1. Yes"), Some((true, 1, "Yes")));
        assert_eq!(numbered("> 1. Yes"), Some((true, 1, "Yes")));
        assert_eq!(numbered("› 2) No"), Some((true, 2, "No")));
        assert_eq!(numbered("│ ❯ 1. Yes │"), Some((true, 1, "Yes │")));
        assert_eq!(numbered("│   2. No   │"), Some((false, 2, "No   │")));
        assert_eq!(numbered("  3. Maybe"), Some((false, 3, "Maybe")));
        // Not choice rows: no number, three digits, no dot, no space after the dot.
        assert_eq!(numbered("> Yes"), None);
        assert_eq!(numbered("│ Do you want to proceed? │"), None);
        assert_eq!(numbered("123. x"), None);
        assert_eq!(numbered("1 Yes"), None);
        assert_eq!(numbered("1.Yes"), None);
    }

    #[test]
    fn yes_and_no_are_whole_words() {
        assert!(starts_with_word("Yes", "yes"));
        assert!(starts_with_word("Yes, and don't ask again", "yes"));
        assert!(starts_with_word("No (esc)", "no"));
        assert!(!starts_with_word("Yesterday", "yes"));
        assert!(!starts_with_word("Not now", "no"));
        assert!(!starts_with_word("", "no"));
    }

    #[test]
    fn a_question_above_a_choice_shaped_row_is_unknown_until_it_leaves_the_window() {
        // The row below the question is a split of short words (REQ-6), then filler rows that
        // are no choice. (Before the shape criterion this test used `[1] Yes` and `[2] No`; those
        // two rows are a choice list by themselves now, see `two_list_entries_are_a_choice_list`.)
        let with = |question: &str, below: usize| {
            let mut v = filler(6, &["✳ Frosting… (esc to interrupt)", question]);
            v.push("Allow / Deny".to_string());
            v.extend((0..below).map(|i| format!("below {i}")));
            classify(&v)
        };
        for q in [
            "Do you want to proceed?",
            "Would you like to continue?",
            "Are you sure?",
            "Proceed? [y/N]",
            "Overwrite the file (y/n)",
        ] {
            let seen = with(q, 0);
            assert_eq!(seen.state, State::Unknown, "{q}");
            assert_eq!(seen.evidence.as_deref(), Some(q), "{q}");
        }
        // The window is the last 8 non-blank rows: the 8th still counts, the 9th is transcript.
        assert_eq!(with("Do you want to proceed?", 6).state, State::Unknown);
        assert_eq!(with("Do you want to proceed?", 7).state, State::Working);
        // A blank row is not a row of the window.
        let mut blanks = filler(
            6,
            &[
                "✳ Frosting… (esc to interrupt)",
                "Do you want to proceed?",
                "[1] Yes",
            ],
        );
        blanks.extend(std::iter::repeat_n(String::new(), 20));
        assert_eq!(classify(&blanks).state, State::Unknown);
        // No question: still working.
        assert_eq!(with("All done.", 2).state, State::Working);
    }

    #[test]
    fn a_question_that_ends_the_turn_is_not_unknown_without_a_choice_shaped_row_below() {
        let ends = |question: &str, below: &[&str]| {
            let mut v = filler(6, &["✳ Frosting… (esc to interrupt)", question]);
            v.extend(below.iter().map(|s| s.to_string()));
            classify(&v).state
        };
        let q = "Would you like me to run the lint step next?";
        assert_eq!(ends(q, &[]), State::Working);
        assert_eq!(
            ends(q, &["❯ ", "⏵⏵ bypass (shift+tab to cycle)"]),
            State::Working
        );
        // REQ-7: a bare word or a sentence is plain prose, not a choice shape, so it no longer
        // refuses (`yes`, `No.` and `answer: NO` did before the word list was removed).
        for word in [
            "Yesterday",
            "Nobody",
            "Not now",
            "Yeses",
            "Noted",
            "yes",
            "No.",
            "answer: NO",
        ] {
            assert_eq!(ends(q, &[word]), State::Working, "{word}");
        }
        // A short bracketed word and a split of short words are choice shapes.
        for word in ["(Yes)", "[no]", "yes/no"] {
            assert_eq!(ends(q, &[word]), State::Unknown, "{word}");
        }
        // Two bare words one to a row are a list of choices drawn one to a line.
        assert_eq!(ends(q, &["Allow", "Deny"]), State::Unknown);
        // A yes or no ABOVE the question does not count.
        let mut v = filler(6, &["yes it ran", "✳ Frosting… (esc to interrupt)", q]);
        assert_eq!(classify(&v).state, State::Working);
        // A bracket question is its own choice: unknown with nothing below it.
        v.push("Overwrite the file? [y/n]".to_string());
        assert_eq!(classify(&v).state, State::Unknown);
    }

    #[test]
    fn a_question_with_a_choice_shape_on_its_row_or_below_is_unknown() {
        let screen = |rows: &[&str]| {
            let mut v = filler(6, &["✳ Frosting… (esc to interrupt)"]);
            v.extend(rows.iter().map(|s| s.to_string()));
            classify(&v)
        };
        // The choice is on the question row (note 3433, R-B1).
        for q in [
            "Do you want to proceed? (yes/no)",
            "Are you sure you want to continue connecting (yes/no/[fingerprint])?",
            "Do you want to proceed? (y/N/a)",
            "Do you want to proceed? Allow or deny",
        ] {
            let seen = screen(&[q]);
            assert_eq!(seen.state, State::Unknown, "{q}");
            assert_eq!(seen.evidence.as_deref(), Some(q), "{q}");
        }
        // The choices use other words (note 3433, R-N1): a choice word below, or two numbered rows.
        let q = "Do you want to proceed?";
        for below in [
            &["Allow / Deny"][..],
            &["Allow", "Deny"],
            &["1. Allow once", "2. Allow always", "3. Deny"],
            &["◆ 1. Y", "  2. N"],
            &["  1. Proceed", "  2. Cancel"],
            // REQ-6: one numbered row under a known question is a choice shape too.
            &["  1. maybe later"],
        ] {
            let mut rows = vec![q];
            rows.extend_from_slice(below);
            assert_eq!(screen(&rows).state, State::Unknown, "{below:?}");
        }
        // REQ-7: one bare word (the word list is gone), or words that only contain a choice
        // word, are plain prose.
        for word in [
            "allow",
            "Deny",
            "y",
            "N",
            "Allowance",
            "denied",
            "yard",
            "Nine",
            "yesterday",
            "nobody",
        ] {
            assert_eq!(screen(&[q, word]).state, State::Working, "{word}");
        }
        // A choice word ABOVE the question does not count, and neither does one in an unrelated row.
        assert_eq!(screen(&["yes", "All done.", q]).state, State::Working);
        // Numbered rows ABOVE the question do not count.
        assert_eq!(screen(&["1. a", "2. b", q]).state, State::Working);
    }

    #[test]
    fn two_list_entries_are_a_choice_list_without_any_question() {
        let screen = |rows: &[&str]| {
            let mut v = filler(6, &["✳ Frosting… (esc to interrupt)"]);
            v.extend(rows.iter().map(|s| s.to_string()));
            classify(&v)
        };
        for rows in [
            &["[1] Yes", "[2] No"][..],
            &["1: Yes", "2: No"],
            &["a. Yes", "b. No"],
            &["a) Keep", "b) Replace"],
            &["Overwrite? (y/N/a)"],
        ] {
            let seen = screen(rows);
            assert_eq!(seen.state, State::Unknown, "{rows:?}");
            assert!(seen.evidence.is_some(), "{rows:?}");
        }
        // One entry is not a list; numbered `1.` rows without a question are the reader of
        // `numbered`, not of the shape reader (a transcript lists steps that way).
        assert_eq!(screen(&["[1] Yes"]).state, State::Working);
        assert_eq!(
            screen(&["1. first step", "2. second step"]).state,
            State::Working
        );
        // The two entries must be among the last six rows: five rows below them still refuse,
        // a sixth makes them old output.
        let far = |below: usize| {
            let mut v = filler(6, &["✳ Frosting… (esc to interrupt)", "[1] Yes", "[2] No"]);
            v.extend((0..below).map(|i| format!("below {i}")));
            classify(&v).state
        };
        assert_eq!(far(4), State::Unknown);
        assert_eq!(far(5), State::Working);
        // The input prompt and the footer of an idle lane are no list entries.
        let mut idle = filler(6, &["Would you like me to run the lint step next?", "❯ "]);
        idle.push("? for shortcuts".to_string());
        assert_eq!(classify(&idle).state, State::Idle);
    }

    #[test]
    fn the_shapes_read_the_rows_they_name_and_leave_prose_alone() {
        // bracket_question: two or more tokens of at most three letters in one group.
        for row in [
            "Overwrite? (y/N/a)",
            "Proceed? [y/n]",
            "Save? (yes/no)",
            "continue connecting (yes/no/[fingerprint])?",
        ] {
            assert!(bracket_question(row), "{row}");
        }
        for row in [
            "(esc to interrupt)",
            "(shift+tab to cycle)",
            "(src/lane_screen.rs)",
            "(yes)",
            "(approve/reject)",
            "no brackets",
        ] {
            assert!(!bracket_question(row), "{row}");
        }
        // bracketed: a short row that opens with a bracket group.
        for row in [
            "[1] Yes",
            "[ Proceed ] [ Cancel ]",
            "[Y]es [N]o",
            "(a) Keep",
            "(Yes)",
        ] {
            assert!(bracketed(row), "{row}");
        }
        for row in [
            "(esc to interrupt)",
            "[a very long bracket group here] x",
            "text [1]",
            "[1] Yes and a tail that runs past forty characters",
        ] {
            assert!(!bracketed(row), "{row}");
        }
        // marked: `1: Yes`, `a. Yes`, `a) Yes`; the `1.` rows only with `numbered_too`.
        for row in ["1: Yes", "a. Yes", "b) No", "c:"] {
            assert!(marked(row, false), "{row}");
        }
        assert!(!marked("1. Yes", false));
        assert!(marked("1. Yes", true));
        assert!(marked("2) No", true));
        for row in [
            "1.Yes",
            "12:30 pm",
            "note: this",
            "ab. no",
            "No changes yet",
            "a",
        ] {
            assert!(!marked(row, true), "{row}");
        }
        // split_choice: short words split by `/`, `|` or `or`.
        for row in [
            "Approve / Reject",
            "OK / Cancel",
            "Yes/No",
            "Keep | Replace",
            "choose: yes or no",
            "Allow once / Allow always",
        ] {
            assert!(split_choice(row), "{row}");
        }
        for row in [
            "src/lane_screen.rs",
            "and then it ran / and stopped after a very long sentence here",
            "no changes yet",
            "a/1",
            "Approve",
        ] {
            assert!(!split_choice(row), "{row}");
        }
        // word_rows: two bare words, one to a row, with no other row between them.
        assert!(word_rows(&["Allow".into(), "Deny".into()]));
        assert!(word_rows(&[
            "x y".into(),
            "Allow".into(),
            String::new(),
            "Deny".into()
        ]));
        assert!(!word_rows(&["Allow".into()]));
        assert!(!word_rows(&[
            "Allow".into(),
            "two words".into(),
            "Deny".into()
        ]));
    }

    #[test]
    fn a_numbered_yes_and_no_after_any_mark_is_a_prompt_whatever_the_footer() {
        let dialog = |prefix: &str| {
            let mut v = filler(6, &["✳ Frosting… (esc to interrupt)"]);
            v.push(format!("{prefix}1. Yes"));
            v.push("   2. No".to_string());
            classify(&v)
        };
        for prefix in [
            "◆ ", "» ", "• ", "- ", "* ", "→→ ", ">>> ", "✓ ", "· ", "│ ◆ ", "[•] ", "☞ ", "◆",
            "  ", "",
        ] {
            let seen = dialog(prefix);
            assert_eq!(seen.state, State::PermissionPrompt, "prefix {prefix:?}");
            assert!(seen.evidence.is_some(), "prefix {prefix:?}");
        }
        // A letter or a digit before the number is a word of a sentence, not a mark.
        assert_eq!(dialog("ok ").state, State::Working);
        assert_eq!(dialog("2 ").state, State::Working);
        // A question the reader does not know, far from the known words, changes nothing.
        let mut v = filler(
            6,
            &[
                "✳ Frosting… (esc to interrupt)",
                "Continue with the reset of the branch?",
            ],
        );
        v.push("◆ 1. Yes".to_string());
        v.push("  2. No".to_string());
        assert_eq!(classify(&v).state, State::PermissionPrompt);
    }

    #[test]
    fn numbered_skips_any_mark_and_keeps_only_a_listed_cursor_as_a_cursor() {
        assert_eq!(numbered("◆ 1. Yes"), Some((false, 1, "Yes")));
        assert_eq!(numbered("» 2) No"), Some((false, 2, "No")));
        assert_eq!(numbered("❯ 1. Yes"), Some((true, 1, "Yes")));
        assert_eq!(numbered("│ ❯ 1. Yes │"), Some((true, 1, "Yes │")));
        assert_eq!(numbered("◆1. Yes"), Some((false, 1, "Yes")));
        assert_eq!(numbered("ok 1. Yes"), None);
        assert_eq!(numbered("1.Yes"), None);
        assert_eq!(numbered("[1] Yes"), None);
        assert_eq!(numbered("123. Yes"), None);
    }

    #[test]
    fn a_numbered_yes_and_no_is_a_prompt_in_any_dress_and_never_working_or_idle() {
        let dialog = |yes: &str, no: &str| {
            let mut v = filler(6, &["esc to interrupt"]);
            v.push(yes.to_string());
            v.push(no.to_string());
            classify(&v).state
        };
        assert_eq!(dialog("> 1. Yes", "  2. No"), State::PermissionPrompt);
        assert_eq!(
            dialog("│ ❯ 1. Yes │", "│   2. No │"),
            State::PermissionPrompt
        );
        assert_eq!(dialog("  1. Yes", "  2. No"), State::PermissionPrompt);
        assert_eq!(dialog("  1. Yesterday", "  2. Not now"), State::Working);
    }

    /// An `orca terminal list --json` answer with `n` terminals.
    fn list_of(n: usize) -> String {
        let terminals: Vec<serde_json::Value> = (1..=n)
            .map(|i| serde_json::json!({"handle": format!("t{i}"), "title": "lane", "branch": "main"}))
            .collect();
        serde_json::json!({
            "ok": true,
            "result": {"terminals": terminals, "totalCount": n, "truncated": false},
        })
        .to_string()
    }

    fn working_answer() -> String {
        let rows = ["a", "b", "c", "d", "esc to interrupt", "e"];
        serde_json::json!({"ok": true, "result": {"terminal": {
            "handle": "t", "status": "running", "tail": rows, "truncated": false,
            "limited": false, "returnedLineCount": 6, "source": "screen",
        }}})
        .to_string()
    }

    #[test]
    fn a_list_over_max_terminals_fails_naming_both_sizes_and_reads_nothing() {
        let mut reads = 0;
        let limits = ScanLimits {
            deadline: None,
            max_terminals: Some(2),
        };
        let err = scan(
            &list_of(3),
            &mut |_| {
                reads += 1;
                Ok(working_answer())
            },
            NOW,
            &limits,
        )
        .unwrap_err();
        assert!(err.contains("3 terminals"), "{err}");
        assert!(err.contains("--max-terminals 2"), "{err}");
        assert_eq!(reads, 0, "no screen is read from an over-large list");
        // A list exactly as long as the cap is scanned, and so is any list with no cap.
        for max in [Some(3), None] {
            let limits = ScanLimits {
                deadline: None,
                max_terminals: max,
            };
            let done = scan(&list_of(3), &mut |_| Ok(working_answer()), NOW, &limits).unwrap();
            assert_eq!(done["count"], serde_json::json!(3), "{max:?}");
        }
    }

    #[test]
    fn after_the_time_cap_the_rest_are_unreachable_and_name_the_cap() {
        use std::time::{Duration, Instant};
        // A cap that already passed reads nothing and every terminal says so.
        let past = Instant::now()
            .checked_sub(Duration::from_secs(1))
            .expect("the clock has run for a second");
        let limits = ScanLimits {
            deadline: Some((past, 7)),
            max_terminals: None,
        };
        let mut reads = 0;
        let done = scan(
            &list_of(2),
            &mut |_| {
                reads += 1;
                Ok(working_answer())
            },
            NOW,
            &limits,
        )
        .unwrap();
        assert_eq!(reads, 0);
        for row in done["terminals"].as_array().unwrap() {
            assert_eq!(row["state"], serde_json::json!("unreachable"));
            let reason = row["reason"].as_str().unwrap();
            assert!(reason.contains("--max-seconds 7"), "{reason}");
            assert!(reason.contains("not read"), "{reason}");
        }
        // Time left: every terminal is read.
        let limits = ScanLimits {
            deadline: Some((Instant::now() + Duration::from_secs(60), 60)),
            max_terminals: None,
        };
        let done = scan(&list_of(2), &mut |_| Ok(working_answer()), NOW, &limits).unwrap();
        assert_eq!(done["by_state"]["working"], serde_json::json!(2));
        // A read that ends after the cap keeps its own reason beside the cap.
        let limits = ScanLimits {
            deadline: Some((Instant::now() + Duration::from_millis(30), 1)),
            max_terminals: None,
        };
        let done = scan(
            &list_of(1),
            &mut |_| {
                std::thread::sleep(Duration::from_millis(80));
                Err("did not answer within 1 seconds: `orca`".to_string())
            },
            NOW,
            &limits,
        )
        .unwrap();
        let reason = done["terminals"][0]["reason"].as_str().unwrap();
        assert!(reason.contains("--max-seconds 1"), "{reason}");
        assert!(
            reason.contains("did not answer within 1 seconds"),
            "{reason}"
        );
    }

    #[test]
    fn the_state_names_are_the_lane_status_names() {
        assert_eq!(State::Working.as_str(), "working");
        assert_eq!(State::Idle.as_str(), "idle");
        assert_eq!(State::PermissionPrompt.as_str(), "permission_prompt");
        assert_eq!(State::InputPending.as_str(), "input_pending");
        assert_eq!(State::UsageLimit.as_str(), "usage_limit");
        assert_eq!(State::Unreachable.as_str(), "unreachable");
        assert_eq!(State::Unknown.as_str(), "unknown");
    }

    #[test]
    fn five_states_read_as_a_lane_status_pane_and_two_do_not() {
        for (state, name) in [
            (State::Working, "working"),
            (State::PermissionPrompt, "permission_prompt"),
            (State::InputPending, "input_pending"),
            (State::UsageLimit, "usage_limit"),
            (State::Unreachable, "unreachable"),
        ] {
            let pane = serde_json::json!({
                "state": state.as_str(),
                "source": "orca:screen",
                "observed_at": "2026-11-19T00:00:00Z",
                "until": null,
            });
            let read: crate::lane_status::PaneObservation =
                serde_json::from_value(pane).unwrap_or_else(|e| panic!("{name}: {e}"));
            assert_eq!(Some(read.state), state.pane_state(), "{name}");
        }
        assert!(State::Idle.pane_state().is_none());
        assert!(State::Unknown.pane_state().is_none());
    }

    #[test]
    fn a_yes_no_dialog_wins_over_every_other_marker() {
        let screen = rows(
            "✳ Working… (12s · esc to interrupt)\n❯ \n  ⏵⏵ bypass permissions on (shift+tab to cycle)\n\
             Do you want to proceed?\n ❯ 1. Yes\n   2. No",
        );
        let seen = classify(&screen);
        assert_eq!(seen.state, State::PermissionPrompt);
        assert!(seen.evidence.unwrap().contains("Do you want"));
    }

    #[test]
    fn a_dialog_needs_both_a_yes_row_and_a_no_row() {
        let only_yes = rows("Do you want to continue?\n ❯ 1. Yes\n   2. Maybe");
        assert_eq!(classify(&only_yes).state, State::InputPending);
        // Without the `Do you want` question a numbered Yes and No are still a confirmation
        // (audit note 3212 B1): Enter would pick Yes.
        let no_question = rows(" ❯ 1. Yes\n   2. No");
        let seen = classify(&no_question);
        assert_eq!(seen.state, State::PermissionPrompt);
        assert!(seen.evidence.unwrap().contains("Yes"));
    }

    #[test]
    fn a_numbered_choice_or_a_dialog_footer_is_input_pending() {
        assert_eq!(
            classify(&rows(" ❯ 1. Default\n   2. Sonnet")).state,
            State::InputPending
        );
        assert_eq!(
            classify(&rows("Pick one\nEnter to confirm · Esc to exit")).state,
            State::InputPending
        );
        assert_eq!(
            classify(&rows("Settings\nEsc to cancel")).state,
            State::InputPending
        );
    }

    #[test]
    fn a_usage_limit_message_wins_over_a_menu_and_a_prompt() {
        let screen = rows(
            "  5-hour limit reached ∙ resets 7pm (UTC)\n ❯ 1. Stop and wait for limit to reset\n   2. Upgrade",
        );
        let seen = classify(&screen);
        assert_eq!(seen.state, State::UsageLimit);
        assert!(seen.evidence.unwrap().contains("limit reached"));
    }

    #[test]
    fn working_and_idle_need_six_non_blank_rows() {
        let working = |n| classify(&filler(n, &["✳ Frosting… (esc to interrupt)"])).state;
        // Five non-blank rows (four fillers + the marker row): unknown; six: working.
        assert_eq!(working(4), State::Unknown);
        assert_eq!(working(5), State::Working);
        let idle = |n| {
            classify(&filler(
                n,
                &["❯ ", "  ⏵⏵ bypass permissions on (shift+tab to cycle)"],
            ))
            .state
        };
        assert_eq!(idle(3), State::Unknown); // 5 rows
        assert_eq!(idle(4), State::Idle); // 6 rows
    }

    #[test]
    fn a_refusal_marker_needs_no_minimum_of_rows() {
        let short = rows("Do you want to proceed?\n ❯ 1. Yes\n   3. No");
        assert_eq!(classify(&short).state, State::PermissionPrompt);
        assert_eq!(
            classify(&rows("Your limit will reset at 2026-11-19 17:30:55")).state,
            State::UsageLimit
        );
    }

    #[test]
    fn blank_rows_do_not_count_and_an_empty_screen_is_unknown() {
        let mut screen = filler(4, &["✳ Frosting… (esc to interrupt)"]);
        screen.extend(std::iter::repeat_n(String::new(), 30));
        screen.extend(std::iter::repeat_n("   ".to_string(), 5));
        let seen = classify(&screen);
        assert_eq!(seen.state, State::Unknown);
        assert_eq!(seen.rows, 5);
        assert_eq!(classify(&[]).state, State::Unknown);
        assert_eq!(classify(&rows("\n\n  \n")).state, State::Unknown);
    }

    #[test]
    fn a_shell_screen_is_unknown_never_idle() {
        let shell = rows("a\nb\nc\nd\ne\nf\ngoos@host project %");
        assert_eq!(classify(&shell).state, State::Unknown);
    }

    #[test]
    fn evidence_is_one_clean_short_row() {
        let long = format!("Do you want {}\n ❯ 1. Yes\n   2. No", "x".repeat(400));
        let seen = classify(&rows(&long));
        let evidence = seen.evidence.unwrap();
        assert!(
            evidence.chars().count() <= 161,
            "{}",
            evidence.chars().count()
        );
        assert!(evidence.ends_with('…'));
        let control = "Do you want\x1b[31m red\x07\n 1. Yes\n 3. No";
        let seen = classify(&rows(control));
        let evidence = seen.evidence.unwrap();
        assert!(!evidence.chars().any(|c| c.is_control()), "{evidence:?}");
    }

    #[test]
    fn text_from_a_terminal_hides_secret_shapes_and_unsafe_characters() {
        // The shapes are built from pieces so none stands in this file (re-audit C-B1, C-N1).
        let gh = ["gh", "p_"].concat() + "A1b2C3d4E5f6G7h8I9j0";
        let sk = ["s", "k-"].concat() + "abcdefghij0123456789";
        let pw = ["pass", "word="].concat() + "hunter2hunter2";
        for secret in [&gh, &sk, &pw] {
            let shown_text = shown(&format!("deploy {secret} now"));
            assert_eq!(shown_text, "deploy [redacted] now", "{secret}");
            let seen = classify(&rows(&format!(
                "Do you want to proceed? (yes/no) {secret}\nrow\nrow\nrow\nrow\nrow"
            )));
            let evidence = seen.evidence.unwrap();
            assert!(evidence.ends_with("[redacted]"), "{evidence}");
            assert!(!evidence.contains("A1b2") && !evidence.contains("hunter2"));
            assert!(
                list_error(&format!("no runtime {secret}")).ends_with("[redacted]"),
                "{secret}"
            );
        }
        // A zero-width or bidirectional character cannot split a shape and hide it.
        let split = format!("{}\u{200b}{}", &gh[..4], &gh[4..]);
        assert_eq!(shown(&split), "[redacted]");
        assert_eq!(shown("a\u{202e}b\u{200b}c\u{2066}d\u{2028}e"), "abcde");
        // An ordinary row is not changed.
        assert_eq!(shown("work lane"), "work lane");
        assert_eq!(cut("Do you want to proceed?"), "Do you want to proceed?");
        // The reset text of an unknown zone is safe too.
        let found = reset_of(
            &rows(&format!(
                "Your limit will reset at 2026-11-19 17:30:55 {sk}"
            )),
            NOW,
        );
        let found = found.expect("a limit row has a reset");
        assert!(matches!(found, Reset::Unknown { .. }), "{found:?}");
        let raw = reset_json(&found)["raw"].as_str().unwrap().to_string();
        assert!(raw.ends_with("[redacted]"), "{raw}");
    }

    #[test]
    fn a_time_with_no_zone_is_unknown_with_its_text() {
        let row = "  ⎿  Your limit will reset at 2026-11-19 17:30:55";
        match parse_reset(row, NOW) {
            Reset::Unknown { raw, why } => {
                assert!(raw.contains("2026-11-19 17:30:55"), "{raw}");
                assert!(why.contains("time zone"), "{why}");
            }
            other => panic!("{other:?}"),
        }
    }

    #[test]
    fn utc_offsets_and_named_zones_are_converted() {
        let base = "Your limit will reset at 2026-11-19 17:30:55";
        let at = |zone: &str| utc(&format!("{base} {zone}"), NOW);
        let utc_1730 = Some(NOW + 17 * 3600 + 30 * 60 + 55);
        assert_eq!(at("UTC"), utc_1730);
        assert_eq!(at("(UTC)"), utc_1730);
        assert_eq!(at("GMT"), utc_1730);
        assert_eq!(at("Z"), utc_1730);
        assert_eq!(at("UTC+8"), Some(NOW + 9 * 3600 + 30 * 60 + 55));
        assert_eq!(at("UTC+9"), Some(NOW + 8 * 3600 + 30 * 60 + 55));
        assert_eq!(at("UTC+08:00"), at("UTC+8"));
        assert_eq!(at("+0800"), at("UTC+8"));
        assert_eq!(at("GMT+8"), at("UTC+8"));
        assert_eq!(at("UTC-5"), Some(NOW + 22 * 3600 + 30 * 60 + 55));
        assert_eq!(at("UTC+5:30"), Some(NOW + 12 * 3600 + 55));
        assert_eq!(at("(Asia/Seoul)"), at("UTC+9"));
        assert_eq!(at("(Asia/Shanghai)"), at("UTC+8"));
        assert_eq!(at("(asia/seoul)"), at("UTC+9"));
    }

    #[test]
    fn eight_hours_ahead_is_one_hour_after_nine_hours_ahead() {
        // The accident: the same printed time is an hour LATER when the zone is UTC+8.
        let row = |z: &str| format!("Your limit will reset at 2026-11-19 17:30:55 {z}");
        assert_eq!(
            utc(&row("UTC+8"), NOW).unwrap() - utc(&row("UTC+9"), NOW).unwrap(),
            3600
        );
    }

    #[test]
    fn an_unknown_zone_is_unknown_with_its_text() {
        match parse_reset("limit reached ∙ resets 7pm (America/New_York)", NOW) {
            Reset::Unknown { raw, why } => {
                assert!(raw.contains("America/New_York"), "{raw}");
                assert!(why.contains("zone"), "{why}");
            }
            other => panic!("{other:?}"),
        }
        assert!(utc("resets 7pm (PST)", NOW).is_none());
    }

    #[test]
    fn a_date_without_a_year_takes_the_next_one_after_now() {
        // Nov 20 7pm Seoul = 10:00 UTC.
        assert_eq!(
            utc("limit reached ∙ resets Nov 20 at 7pm (Asia/Seoul)", NOW),
            Some(NOW + 86_400 + 10 * 3600)
        );
        // Past this year already: next year.
        let dec30 = NOW + 41 * 86_400; // 2026-12-30
        assert_eq!(
            utc("resets Jan 2 at 1am (UTC)", dec30),
            Some(dec30 + 3 * 86_400 + 3600)
        );
        assert_eq!(
            utc("resets November 20, 2026 at 7pm (UTC)", NOW),
            Some(NOW + 86_400 + 19 * 3600)
        );
    }

    #[test]
    fn a_time_whose_utc_year_leaves_1970_to_9999_is_unknown() {
        // 9999-12-31 23:59:00 UTC is the last readable minute; one minute later is year 10000.
        assert_eq!(
            utc("resets 9999-12-31 15:59 -0800", NOW),
            Some(253_402_300_740)
        );
        assert_eq!(utc("resets 9999-12-31 16:00 -0800", NOW), None);
        assert_eq!(utc("resets 9999-12-31 23:59 -0800", NOW), None);
        // 1970-01-01 00:00:00 UTC is the first readable second; a minute earlier is 1969.
        assert_eq!(utc("resets 1970-01-01 08:00 +0800", NOW), Some(0));
        assert_eq!(utc("resets 1970-01-01 07:59 +0800", NOW), None);
        // A month-name date with a year takes the same door.
        assert_eq!(utc("resets Dec 31, 9999 at 11:59pm (UTC-8)", NOW), None);
    }

    #[test]
    fn a_time_of_day_alone_is_the_next_one_after_now() {
        // 19:00 UTC, asked at 20:00 UTC: tomorrow; asked at 10:00 UTC: today.
        assert_eq!(
            utc("resets 7pm (UTC)", NOW + 20 * 3600),
            Some(NOW + 86_400 + 19 * 3600)
        );
        assert_eq!(
            utc("resets 7pm (UTC)", NOW + 10 * 3600),
            Some(NOW + 19 * 3600)
        );
        // 07:30 in Seoul is 22:30 UTC of the day before: already past at 00:00 UTC, so the next.
        assert_eq!(
            utc("resets 7:30am (Asia/Seoul)", NOW),
            Some(NOW + 22 * 3600 + 30 * 60)
        );
        assert_eq!(utc("resets 12am (UTC)", NOW + 3600), Some(NOW + 86_400));
        assert_eq!(utc("resets 12pm (UTC)", NOW), Some(NOW + 12 * 3600));
    }

    #[test]
    fn impossible_times_are_unknown() {
        for row in [
            "resets 13pm (UTC)",
            "resets Feb 30 at 1am (UTC)",
            "resets 2026-13-01 00:00:00 UTC",
            "resets 2026-11-19 25:00:00 UTC",
            "resets 2026-11-19 17:61:00 UTC",
            "resets 0am (UTC)",
        ] {
            assert!(utc(row, NOW).is_none(), "{row}");
        }
    }

    #[test]
    fn reset_of_finds_the_message_row_and_only_that() {
        let screen = rows("working\n  ⎿  Your limit will reset at 2026-11-19 17:30:55 UTC\nmore");
        assert_eq!(
            reset_of(&screen, NOW),
            Some(Reset::Utc(NOW + 17 * 3600 + 30 * 60 + 55))
        );
        assert_eq!(reset_of(&rows("working\nnothing"), NOW), None);
    }

    #[test]
    fn an_orca_screen_read_gives_its_rows() {
        let json = serde_json::json!({
            "id": "x", "ok": true,
            "result": {"terminal": {"handle": "term_x", "status": "running",
                "tail": ["a", "b"], "truncated": false, "returnedLineCount": 2, "source": "screen"}}
        });
        let screen = read_input(&json.to_string());
        assert_eq!(screen.rows, vec!["a".to_string(), "b".to_string()]);
        assert_eq!(screen.source, "orca:screen");
        assert_eq!(screen.verdict, None);
    }

    #[test]
    fn an_orca_read_that_is_not_a_screen_is_unknown() {
        for source in ["stream", "screen-unavailable"] {
            let json = serde_json::json!({
                "ok": true,
                "result": {"terminal": {"status": "running", "tail": ["a"], "source": source}}
            });
            let screen = read_input(&json.to_string());
            assert_eq!(
                screen.verdict.map(|v| v.0),
                Some(State::Unknown),
                "{source}"
            );
        }
        let absent = serde_json::json!({"ok": true, "result": {"terminal": {"tail": ["a"]}}});
        assert_eq!(
            read_input(&absent.to_string()).verdict.map(|v| v.0),
            Some(State::Unknown)
        );
    }

    #[test]
    fn a_gone_terminal_or_an_orca_error_is_unreachable() {
        let exited = serde_json::json!({
            "ok": true,
            "result": {"terminal": {"status": "exited", "tail": [], "source": "screen"}}
        });
        assert_eq!(
            read_input(&exited.to_string()).verdict.map(|v| v.0),
            Some(State::Unreachable)
        );
        let failed = serde_json::json!({"ok": false, "error": {"message": "no terminal"}});
        assert_eq!(
            read_input(&failed.to_string()).verdict.map(|v| v.0),
            Some(State::Unreachable)
        );
    }

    #[test]
    fn a_numbered_list_in_the_text_is_not_a_choice_dialog() {
        // No `❯` cursor on any numbered row: a list the lane wrote, not a dialog.
        let screen = filler(
            4,
            &["1. first", "2. second", "✳ Frosting… (esc to interrupt)"],
        );
        assert_eq!(classify(&screen).state, State::Working);
        let with_cursor = filler(4, &["❯ 1. first", "2. second"]);
        assert_eq!(classify(&with_cursor).state, State::InputPending);
    }

    #[test]
    fn plain_text_is_taken_as_the_screen_the_caller_says_it_is() {
        let screen = read_input("a\nb\r\nc\n");
        assert_eq!(screen.rows, vec!["a", "b", "c"]);
        assert_eq!(screen.source, "stdin:text");
        assert_eq!(screen.verdict, None);
        // JSON that does not parse is text, not an error.
        assert_eq!(read_input("{not json").source, "stdin:text");
    }
}
