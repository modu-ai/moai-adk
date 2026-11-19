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
//! read of `usage-limits.json` by the caller through [`crate::usage_limit::evaluate`].

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

/// `row` cut to [`EVIDENCE_CHARS`] characters.
fn cut(row: &str) -> String {
    cut_to(row, EVIDENCE_CHARS)
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

/// Words of a confirmation question, lower case.
const QUESTIONS: [&str; 6] = [
    "do you want",
    "would you like",
    "are you sure",
    "proceed?",
    "[y/n]",
    "(y/n)",
];

/// The questions of [`QUESTIONS`] that carry their own choice, lower case.
const BRACKET_QUESTIONS: [&str; 2] = ["[y/n]", "(y/n)"];

/// Whether a (lower case) row holds `yes` or `no` as a whole word: `Yes, go` and `(No)` do,
/// `Yesterday`, `Nobody` and `Not now` do not.
fn has_yes_or_no_word(row: &str) -> bool {
    row.split(|c: char| !c.is_alphanumeric())
        .any(|w| w == "yes" || w == "no")
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
    // A confirmation question near the bottom with a `yes` or a `no` word on a row below it, and
    // no numbered Yes and No rows the reader could use (the choices are drawn in a way it does not
    // know): Enter may pick Yes, so the lane is not `working` or `idle` whatever its footer says
    // (leader R-N1, note 3256). A question with only the input line under it is a lane that
    // ended its turn and waits: that is the screen a send is for (leader S-N2, note 3362). A
    // bracket question such as `[y/n]` is its own choice, so it counts with nothing below it.
    let near_bottom = rows
        .iter()
        .enumerate()
        .filter(|(_, r)| !r.is_empty())
        .rev()
        .take(BOTTOM_ROWS);
    if let Some((i, _)) = near_bottom.into_iter().find(|(i, _)| {
        let Some(row) = lower.get(*i) else {
            return false;
        };
        QUESTIONS.iter().any(|q| row.contains(q))
            && (BRACKET_QUESTIONS.iter().any(|q| row.contains(q))
                || lower[*i + 1..].iter().any(|r| has_yes_or_no_word(r)))
    }) {
        return seen(
            State::Unknown,
            rows.get(i),
            "a question is near the bottom of the screen and a yes or no row below it could not be read as a choice",
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
        Some(days_from_civil(y, m, d) * 86_400 + tod - offset)
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
/// usage limit (from the lane's `home` when given), the `pane` object `lane-status` takes, and
/// the count of non-blank rows. `now` is Unix seconds.
pub fn report(input: &str, home: Option<&std::path::Path>, now: i64) -> serde_json::Value {
    use crate::forge::policy::{format_utc, parse_utc};
    use serde_json::{json, Value};

    let (screen, seen) = observe(input);

    // The reset time of a usage limit: the lane's own record when a window is at its limit,
    // else the screen's text.
    let mut reset = Value::Null;
    let mut until: Option<String> = None;
    if seen.state == State::UsageLimit {
        let from_home = home
            .map(|h| crate::usage_limit::evaluate(h, now))
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

/// The error text for a terminal list that cannot be read (the program is absent, it failed, or
/// its answer is not a list).
pub fn list_error(detail: &str) -> String {
    format!(
        "the terminal list could not be read: {}",
        cut(&clean(detail))
    )
}

/// A text somebody else chose (a terminal title or branch), for display: no escape or control
/// characters, no line-separator or bidirectional character, cut to [`SHOWN_CHARS`].
fn shown(text: &str) -> String {
    let safe: String = clean(text)
        .chars()
        .filter(|c| !crate::lane_duty::unsafe_char(*c))
        .collect();
    cut_to(&safe, SHOWN_CHARS)
}

/// The scan of `orca terminal list --json` (`list_json`): one row per listed terminal with the
/// state of its screen, read one terminal at a time through `read` (the handle in, the answer of
/// `orca terminal read --screen --json` out). A terminal that cannot be read, has no handle or is
/// not answered with an Orca answer is `unreachable`; the scan goes on. A usage limit carries its
/// reset time in UTC from the screen text (no `usage-limits.json`: a terminal does not say whose
/// HOME it is). `now` is Unix seconds. `Err` only when the list itself cannot be read.
pub fn scan(
    list_json: &str,
    read: &mut dyn FnMut(&str) -> Result<String, String>,
    now: i64,
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
            Some(raw) => match read(raw) {
                Err(d) => unreachable_row(&format!(
                    "the terminal could not be read: {}",
                    cut(&clean(&d))
                )),
                Ok(out) => {
                    let input = read_input(&out);
                    if input.source != "orca:screen" {
                        unreachable_row("the answer of `orca terminal read` is not an Orca answer")
                    } else {
                        let seen = report(&out, None, now);
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
    fn a_question_above_yes_and_no_words_that_are_no_choice_rows_is_unknown() {
        // The rows below the question: `[1] Yes` and `[2] No` are no choice rows (the number
        // is closed by `]`), then filler rows with no yes or no word.
        let with = |question: &str, below: usize| {
            let mut v = filler(6, &["✳ Frosting… (esc to interrupt)", question]);
            v.extend(["[1] Yes".to_string(), "[2] No".to_string()]);
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
        assert_eq!(with("Do you want to proceed?", 5).state, State::Unknown);
        assert_eq!(with("Do you want to proceed?", 6).state, State::Working);
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
    fn a_question_that_ends_the_turn_is_not_unknown_without_a_yes_or_no_word_below() {
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
        for word in ["Yesterday", "Nobody", "Not now", "Yeses", "Noted"] {
            assert_eq!(ends(q, &[word]), State::Working, "{word}");
        }
        for word in ["yes", "No.", "(Yes)", "[no]", "yes/no", "answer: NO"] {
            assert_eq!(ends(q, &[word]), State::Unknown, "{word}");
        }
        // A yes or no ABOVE the question does not count.
        let mut v = filler(6, &["yes it ran", "✳ Frosting… (esc to interrupt)", q]);
        assert_eq!(classify(&v).state, State::Working);
        // A bracket question is its own choice: unknown with nothing below it.
        v.push("Overwrite the file? [y/n]".to_string());
        assert_eq!(classify(&v).state, State::Unknown);
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
