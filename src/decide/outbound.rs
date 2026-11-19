//! The text of a `moai decide` request that would leave the machine (card C-37 S2): the
//! question, the options and the evidence, and nothing else. The card, the default and the
//! deadline are never sent.
//!
//! Design v2, slice V2 (the secret and privacy screen). Each field is screened on its own:
//! `what`, every option value and description, `undo`, the card, and every evidence field's
//! name and value apart, so a field named `token` with an innocent value passes and a value
//! holding a key does not ([`build`]). [`screen`] reads one field in this order, and the
//! first hit refuses the whole request with its own code ([`SCREEN`]); nothing that hits is
//! masked and sent, and no reason quotes the text:
//! 1. the normaliser ([`super::normalise::normalise`]): a bidi or control character refuses
//!    (`bidi_character`, `control_character`); default-ignorables are taken out and
//!    full-width forms, `ſ` and the Kelvin sign read as ASCII; then a URL with a password in
//!    its user part, `://<user>:<password>@` (`url_credentials`, V2 audit N1), read before
//!    any mask could hide part of it;
//! 2. a word (letters, digits, `_`, `-`) that mixes ASCII letters with Cyrillic or Greek
//!    letters (`mixed_script`; a Greek mu right after a digit, the micro sign of `10μs`, is
//!    not one);
//! 3. the key shapes below (`secret_shape`), also with each character without width read
//!    as a break instead of taken out;
//! 4. a token of [`ENTROPY_LEN_MIN`] or more characters of `A-Z a-z 0-9 _ - + / =` with a
//!    letter and a digit and a Shannon entropy of [`ENTROPY_BITS_MIN`] bits per character or
//!    more (`high_entropy`), unless every part of it between `/ _ - + =` is an ordinary word
//!    ([`is_ordinary_part`]): so a commit id, a sha256 digest, a UUID, a path and a snake-case
//!    name pass;
//! 5. one layer of encoding read ([`super::normalise::decode_once`]: percent, `\u`, `\U`,
//!    `\x`, `\u{…}`, HTML numeric entities, base64 and hex runs) and the URL rule and rules 2
//!    to 4 again on the decoded text (`secret_encoded`).
//!
//! The key shapes are read in time linear in a field's length (V2 audit N2; a 64 KiB field
//! of a repeated prefix is pinned by a test).
//!
//! Known limits, stated: a secret split across two fields (each field is read alone); an
//! encoding inside an encoding (one layer is read); a short password in prose (`hunter2`
//! has no shape, and a field named `password` is not itself a hit); a random token whose
//! parts each read as a word ([`is_ordinary_part`]: `a8f3k2-x9q1m4-…` or
//! `AbCdEfGhIjKlMnOpQrSt_1` read as names); look-alike letters of scripts other than
//! Cyrillic and Greek; and any key format not listed here. A false hit costs one fallback to
//! the leader.
//!
//! Rules kept from slice S2:
//! 1. A field of more than 64 KiB refuses the request ([`FIELD_MAX`]) before any screen reads
//!    it; the whole body has its own bound in [`super::BODY_MAX`]. Refused, never cut: a cut
//!    text is another question.
//! 2. A credential shape anywhere in that text, or in the card, refuses the whole request
//!    (user decision 2026-11-16, D-secret-hit: refuse and fall back, zero calls; never mask
//!    and send). [`holds_secret_shape`] reads the shared shapes of
//!    [`crate::redact::redact_outgoing`]; the same key prefixes right after any character
//!    that is not an ASCII letter or digit (`**sk-…**`, `KEY_sk-…`, a key after Korean text,
//!    after a `%20`, or with a zero-width character in it); `Bearer <token>` and `glpat-…`;
//!    a pair whose key is token, secret, password, api_key, apikey or authorization with a
//!    value; and any run of 48 or more base64 characters that mixes upper case, lower case
//!    and digits (leader decision 2026-11-17, audit N7). A false hit costs one fallback to
//!    the leader. Commit ids and sha256 digests (lower-case hex) are not hits.
//!    Second audit (2026-11-17), N7: also the Stripe (`sk_live_`, `sk_test_`, `rk_live_`),
//!    Hugging Face (`hf_`) and npm (`npm_`) prefixes. N6: `sk-` before a lower-case
//!    hyphenated name (sk-learn-pipeline) and a macOS temporary folder are not hits.
//!    Slice V2: also `rk_test_`, `whsec_` and `sk-proj-` (a lower-case tail too), and any of
//!    the key prefixes glued after a letter or digit (`mysk_live_…`) when at least 16 letters
//!    and digits with a digit among them follow it.
//! 3. The home folder, the user name, an e-mail address (`[email]`, user decision U2) and a
//!    macOS temporary folder (`T` or `C`) are masked ([`mask`]): they say who and where, and
//!    a decision needs none of them. An option value that would change under the mask is
//!    refused instead, because the answer names an option by its value. After masking, each
//!    sent field is held to the request's own bound again (`what` 500 bytes, a description
//!    300, `undo` 200, an evidence value 4 KiB, all values 16 KiB): too long is refused,
//!    never cut.
//!
//! All of it is a conservative filter over known shapes, not a secret detector.
//!
//! The evidence is an object of named fields (design v2, slice V1). It is sent as one text,
//! one `name: value` line per field in name order ([`render_evidence`]). Each name and value
//! is masked on its own and then escaped (slice V2, V1 audit N8): `\` as `\\`, a line feed
//! as `\n`, a carriage return as `\r`, and the other line terminators as `\uXXXX`, so a value
//! cannot start a line of its own and forge another field. `undo` is recorded and never
//! sent; like the card it is screened and masked. The provider slice (V3) may send the
//! fields as an object instead.

use super::gate::Cleared;
use super::normalise;
use super::request::{DESCRIPTION_MAX, EVIDENCE_FIELD_MAX, EVIDENCE_MAX, UNDO_MAX, WHAT_MAX};
use super::Refusal;
use std::collections::BTreeMap;

/// What stands in for the user's name, and for the home folder.
pub const USER_MASK: &str = "[user]";
pub const HOME_MASK: &str = "~";

/// Whose home and name to mask. Given by the caller, so nothing here reads the environment.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct Who {
    pub home: Option<String>,
    pub user: Option<String>,
}

/// The masked texts, made only by [`build`]: the one thing the provider code reads.
#[derive(Debug, Clone, PartialEq)]
pub struct Outbound {
    what: String,
    options: Vec<(String, String)>,
    evidence: String,
    evidence_fields: Vec<String>,
    undo: String,
    card: Option<String>,
}

/// The evidence fields as the one text that is sent: `name: value`, one field after the
/// other in the order given (name order for a map), joined by a line break. Name and value
/// are escaped ([`escaped`]), so each field is exactly one line.
pub fn render_evidence<'a>(fields: impl IntoIterator<Item = (&'a String, &'a String)>) -> String {
    fields
        .into_iter()
        .map(|(name, value)| format!("{}: {}", escaped(name), escaped(value)))
        .collect::<Vec<_>>()
        .join("\n")
}

/// `text` with `\` as `\\`, a line feed as `\n`, a carriage return as `\r` and every other
/// line terminator (vertical tab, form feed, U+0085, U+2028, U+2029) as `\uXXXX`.
fn escaped(text: &str) -> String {
    let mut out = String::with_capacity(text.len());
    for c in text.chars() {
        match c {
            '\\' => out.push_str("\\\\"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\u{b}' | '\u{c}' | '\u{85}' | '\u{2028}' | '\u{2029}' => {
                out.push_str(&format!("\\u{:04x}", c as u32))
            }
            c => out.push(c),
        }
    }
    out
}

/// Largest single text field, in bytes (64 KiB), before any screen reads it. After masking,
/// the request's own smaller bounds apply again ([`build`]).
pub const FIELD_MAX: usize = 64 * 1024;

impl Outbound {
    /// The card, masked like the rest. Kept for the record; no provider code reads it.
    pub fn card(&self) -> Option<&str> {
        self.card.as_deref()
    }

    pub fn what(&self) -> &str {
        &self.what
    }

    /// `(value, description)` in the request's order.
    pub fn options(&self) -> &[(String, String)] {
        &self.options
    }

    /// The rendered evidence text ([`render_evidence`]), masked.
    pub fn evidence(&self) -> &str {
        &self.evidence
    }

    /// The names of the evidence fields, masked like the text; never their values. Kept for
    /// the record.
    pub fn evidence_fields(&self) -> &[String] {
        &self.evidence_fields
    }

    /// How the choice is undone, masked. Kept for the record; no provider code reads it.
    pub fn undo(&self) -> &str {
        &self.undo
    }
}

/// Shortest token the entropy rule reads, and the entropy, in bits per character, at which
/// it refuses.
pub const ENTROPY_LEN_MIN: usize = 20;
pub const ENTROPY_BITS_MIN: f64 = 4.0;

/// What stands in for an e-mail address (user decision U2, 2026-11-16).
pub const EMAIL_MASK: &str = "[email]";

/// The codes of [`screen`] and what each names in a refusal's reason. The text that hit is
/// never quoted.
pub const SCREEN: [(&str, &str); 7] = [
    (
        normalise::BIDI,
        "a bidirectional control character (U+202A to U+202E, U+2066 to U+2069)",
    ),
    (
        normalise::CONTROL,
        "a control character other than tab and line feed",
    ),
    ("url_credentials", "a URL with a password in its user part"),
    (
        "mixed_script",
        "a word that mixes ASCII letters with Cyrillic or Greek letters",
    ),
    ("secret_shape", "a credential shape"),
    (
        "high_entropy",
        "a random-looking token (20 or more characters with letters and digits, 4.0 or more \
         bits of entropy per character)",
    ),
    (
        "secret_encoded",
        "a credential shape or a random-looking token once one layer of encoding is read",
    ),
];

/// Shannon entropy of `token`, in bits per character (0 for an empty token).
pub fn entropy(token: &str) -> f64 {
    let mut counts = BTreeMap::<char, usize>::new();
    for c in token.chars() {
        *counts.entry(c).or_default() += 1;
    }
    let n = counts.values().sum::<usize>() as f64;
    counts
        .values()
        .map(|&count| {
            let p = count as f64 / n;
            -p * p.log2()
        })
        .sum()
}

/// The code of the first rule of the module text that refuses `text`, or `None`. Read for
/// the refusal only: text that passes is sent as it was written (masked).
pub fn screen(text: &str) -> Option<&'static str> {
    let normal = match normalise::normalise(text) {
        Ok(normal) => mask_temp_folders(&normal),
        Err(code) => return Some(code),
    };
    if holds_url_credentials(&normal) {
        return Some("url_credentials");
    }
    if holds_mixed_script(&normal) {
        return Some("mixed_script");
    }
    let spaced = mask_temp_folders(&normalise::spaced(text));
    if holds_pattern(&normal) || holds_pattern(&spaced) {
        return Some("secret_shape");
    }
    if holds_high_entropy(&normal) {
        return Some("high_entropy");
    }
    let decoded = mask_temp_folders(&normalise::fold(&normalise::decode_once(&normal)?));
    (holds_url_credentials(&decoded)
        || holds_mixed_script(&decoded)
        || holds_pattern(&decoded)
        || holds_high_entropy(&decoded))
    .then_some("secret_encoded")
}

/// V2 audit N1: a URL whose user part holds a password, `://<user>:<password>@` (the user may
/// be empty; the password is one or more characters without white space, `/` or `@`). Read
/// before anything is masked: the e-mail mask would otherwise take `<password>@<host>` for an
/// address and send the user name beside it. `://<user>@`, `git@host:…` and a port
/// (`host:8443/`) are not hits.
fn holds_url_credentials(text: &str) -> bool {
    text.match_indices("://").any(|(at, _)| {
        // Each scan stops at the next `/`, so no part of the text is read twice.
        let rest = &text[at + 3..];
        let user = rest
            .find(|c: char| c.is_whitespace() || "/@:".contains(c))
            .unwrap_or(rest.len());
        let Some(after) = rest[user..].strip_prefix(':') else {
            return false;
        };
        let password = after
            .find(|c: char| c.is_whitespace() || "/@".contains(c))
            .unwrap_or(after.len());
        password > 0 && after[password..].starts_with('@')
    })
}

/// True when [`screen`] refuses `text` for any reason.
pub fn holds_secret_shape(text: &str) -> bool {
    screen(text).is_some()
}

/// The key shapes of rule 3, on text the normaliser has read. Two shapes are plain text and
/// not hits (second audit, N6): a macOS temporary folder (the caller has masked it) and a
/// lower-case hyphenated name after `sk-` ([`without_sk_names`]).
fn holds_pattern(text: &str) -> bool {
    let text = without_sk_names(text);
    crate::redact::redact_outgoing(&text).1 > 0
        || holds_glued_shape(&text)
        || holds_secret_pair(&text)
        || holds_long_mixed_run(&text)
}

/// True for a Cyrillic or Greek letter (the Greek mu too; [`holds_mixed_script`] lets it
/// stand for the micro sign right after a digit).
fn is_cyrillic_or_greek(c: char) -> bool {
    c.is_alphabetic()
        && matches!(
            c,
            '\u{370}'..='\u{3ff}'
                | '\u{1f00}'..='\u{1fff}'
                | '\u{400}'..='\u{52f}'
                | '\u{1c80}'..='\u{1c8f}'
                | '\u{2de0}'..='\u{2dff}'
                | '\u{a640}'..='\u{a69f}'
        )
}

/// Rule 2: a word (letters and digits of any script, `_`, `-`) with an ASCII letter and a
/// Cyrillic or Greek letter in it. A Greek mu right after a digit is the micro sign
/// (`10μs`) and does not count; anywhere else it does (`tokμn`).
fn holds_mixed_script(text: &str) -> bool {
    text.split(|c: char| !(c.is_alphanumeric() || c == '_' || c == '-'))
        .any(|word| {
            let mut before = None;
            let other = word.chars().any(|c| {
                let micro = c == '\u{3bc}' && before.is_some_and(|b: char| b.is_ascii_digit());
                before = Some(c);
                is_cyrillic_or_greek(c) && !micro
            });
            other && word.chars().any(|c| c.is_ascii_alphabetic())
        })
}

/// A part of a token, between `/ _ - + =`, that reads as an ordinary word rather than as
/// random: empty, only digits, no digit at all (`ProviderResponse`, `CARGO`), at most 8
/// characters without an upper-case letter (`x86`, `v2026`, `agent9`), a longer one without
/// an upper-case letter whose letters and digits change places at most twice (`followup3`,
/// `ubuntu2404`; found by the false-hit measurement of slice V2), or hex digits of one
/// letter case (a commit id, a digest, a UUID part).
pub fn is_ordinary_part(part: &str) -> bool {
    let has = |f: fn(&u8) -> bool| part.bytes().any(|b| f(&b));
    let changes = part
        .as_bytes()
        .windows(2)
        .filter(|pair| pair[0].is_ascii_digit() != pair[1].is_ascii_digit())
        .count();
    part.bytes().all(|b| b.is_ascii_digit())
        || !has(u8::is_ascii_digit)
        || (part.len() <= 8 && !has(u8::is_ascii_uppercase))
        || (changes <= 2 && !has(u8::is_ascii_uppercase))
        || (part.bytes().all(|b| b.is_ascii_hexdigit())
            && !(has(u8::is_ascii_uppercase) && has(u8::is_ascii_lowercase)))
}

/// Rule 4: a token of [`ENTROPY_LEN_MIN`] or more characters of `A-Z a-z 0-9 _ - + / =`,
/// with a letter and a digit, not made of ordinary parts only ([`is_ordinary_part`]), whose
/// entropy is [`ENTROPY_BITS_MIN`] or more.
fn holds_high_entropy(text: &str) -> bool {
    text.split(|c: char| !(c.is_ascii_alphanumeric() || "_-+/=".contains(c)))
        .any(|token| {
            token.len() >= ENTROPY_LEN_MIN
                && token.bytes().any(|b| b.is_ascii_digit())
                && token.bytes().any(|b| b.is_ascii_alphabetic())
                && !token.split(['/', '_', '-', '+', '=']).all(is_ordinary_part)
                && entropy(token) >= ENTROPY_BITS_MIN
        })
}

/// `text` with each name of the kind `sk-learn-pipeline` blanked out (the same number of
/// `x`): `sk-` where a key could start, then lower-case words joined by at least one more
/// hyphen, with no digit, upper-case letter or underscore anywhere in the run. A key is
/// random and does not look like that; a pair around the name (`token: sk-…`) still hits.
/// `sk-proj-` starts a key whatever follows it (slice V2: a lower-case tail too).
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
        let len = bytes[at..].iter().take_while(|b| key(**b)).count();
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
    // Only ASCII bytes were replaced, each by an ASCII byte.
    String::from_utf8(bytes).unwrap_or_else(|_| text.to_string())
}

fn starts_with_ignoring_case(text: &str, prefix: &str) -> bool {
    text.len() >= prefix.len()
        && text.as_bytes()[..prefix.len()].eq_ignore_ascii_case(prefix.as_bytes())
}

/// How many leading bytes of `text` satisfy `part`.
fn run(text: &str, part: impl Fn(u8) -> bool) -> usize {
    text.bytes().take_while(|b| part(*b)).count()
}

/// True when the first `least` bytes of `text` all satisfy `part`. Reads at most `least`
/// bytes, so a check made at every position of a text stays linear in it (V2 audit N2).
fn reaches(text: &str, part: impl Fn(u8) -> bool, least: usize) -> bool {
    text.bytes().take(least).take_while(|b| part(*b)).count() == least
}

/// A key character: an ASCII letter or digit, `_` or `-`.
fn is_key(b: u8) -> bool {
    b.is_ascii_alphanumeric() || b == b'_' || b == b'-'
}

/// For each byte position of a text (and its end): where the run of ASCII letters and digits
/// that starts there ends and whether it holds a digit, and where the run of key characters
/// and dots ([`is_key`], `.`) that starts there ends and how many dots it holds. Built in one
/// pass from the end, so a rule asked at every position never reads a run again (V2 audit
/// N2: a 64 KiB field of repeated prefixes took 41 to 67 s).
struct Runs {
    alnum_end: Vec<usize>,
    alnum_digit: Vec<bool>,
    token_end: Vec<usize>,
    token_dots: Vec<usize>,
}

impl Runs {
    fn of(bytes: &[u8]) -> Runs {
        let n = bytes.len();
        let mut runs = Runs {
            alnum_end: vec![n; n + 1],
            alnum_digit: vec![false; n + 1],
            token_end: vec![n; n + 1],
            token_dots: vec![0; n + 1],
        };
        for at in (0..n).rev() {
            let b = bytes[at];
            if b.is_ascii_alphanumeric() {
                runs.alnum_end[at] = runs.alnum_end[at + 1];
                runs.alnum_digit[at] = b.is_ascii_digit() || runs.alnum_digit[at + 1];
            } else {
                runs.alnum_end[at] = at;
            }
            if is_key(b) || b == b'.' {
                runs.token_end[at] = runs.token_end[at + 1];
                runs.token_dots[at] = usize::from(b == b'.') + runs.token_dots[at + 1];
            } else {
                runs.token_end[at] = at;
            }
        }
        runs
    }
}

/// More key prefixes (second audit, N7; slice V2) and the fewest letters and digits right
/// after one that make a hit: Stripe secret, restricted and webhook keys, Hugging Face and
/// npm tokens. Their tails hold no `_` or `-`, which keeps names like `npm_config_cache` out.
const MORE_PREFIXES: [(&str, usize); 7] = [
    ("sk_live_", 16),
    ("sk_test_", 16),
    ("rk_live_", 16),
    ("rk_test_", 16),
    ("whsec_", 16),
    ("hf_", 20),
    ("npm_", 20),
];

/// The key prefixes that also count glued after a letter or digit (`mysk_live_…`, slice V2),
/// matched in any letter case but the last three, and the fewest letters and digits right
/// after one, with a digit among them. The digit keeps words out (`work_test_suite…`,
/// `task-level-configuration`).
const GLUED_PREFIXES: [&str; 17] = [
    "sk-", "sk_live_", "sk_test_", "rk_live_", "rk_test_", "whsec_", "hf_", "npm_", "glpat-",
    "xoxa-", "xoxb-", "xoxp-", "xoxr-", "xoxs-", "AKIA", "ASIA", "AIza",
];
pub const GLUED_TAIL_MIN: usize = 16;

/// True when the text at `at`, which follows a letter or digit, starts with one of
/// [`GLUED_PREFIXES`] and [`GLUED_TAIL_MIN`] or more letters and digits with a digit among
/// them. The run after the prefix is looked up in `runs`, never measured again.
fn holds_glued_after_letter(text: &str, at: usize, runs: &Runs) -> bool {
    let here = &text[at..];
    GLUED_PREFIXES.iter().any(|prefix| {
        let case_matters = prefix.bytes().any(|b| b.is_ascii_uppercase());
        let starts = if case_matters {
            here.starts_with(prefix)
        } else {
            starts_with_ignoring_case(here, prefix)
        };
        let tail = at + prefix.len();
        starts && runs.alnum_end[tail] - tail >= GLUED_TAIL_MIN && runs.alnum_digit[tail]
    })
}

/// A known key prefix right after any character that is not an ASCII letter or digit (or at
/// the start, or after a `%XX` escape): `**sk-…**`, `KEY_sk-…`, a key after Korean text or a
/// curly quote. The prefixes and lengths are those of [`crate::redact::redact_outgoing`],
/// plus `glpat-` and `Bearer <token>` (audit N7). After a letter or digit, the narrower
/// [`holds_glued_after_letter`] reads instead. Each length is read only up to the number it
/// is compared with ([`reaches`]), or looked up in [`Runs`], so the whole check is linear in
/// the text (V2 audit N2).
fn holds_glued_shape(text: &str) -> bool {
    let bytes = text.as_bytes();
    let runs = Runs::of(bytes);
    let key = is_key;
    (0..bytes.len()).any(|at| {
        if !bytes[at].is_ascii_alphanumeric() {
            return false;
        }
        let escaped = at >= 3
            && bytes[at - 3] == b'%'
            && bytes[at - 2].is_ascii_hexdigit()
            && bytes[at - 1].is_ascii_hexdigit();
        if at > 0 && bytes[at - 1].is_ascii_alphanumeric() && !escaped {
            // `at` follows an ASCII byte, so the slice is valid.
            return holds_glued_after_letter(text, at, &runs);
        }
        // `at` follows an ASCII byte or starts a character, so the slice is valid.
        let here = &text[at..];
        let starts = |prefix: &str| starts_with_ignoring_case(here, prefix);
        if starts("sk-") && reaches(here, key, 13) {
            return true;
        }
        if (here.starts_with("AKIA") || here.starts_with("ASIA"))
            && reaches(
                &here[4..],
                |b| b.is_ascii_uppercase() || b.is_ascii_digit(),
                16,
            )
        {
            return true;
        }
        if starts("xox")
            && matches!(bytes.get(at + 3), Some(b'a' | b'b' | b'p' | b'r' | b's'))
            && bytes.get(at + 4) == Some(&b'-')
            && reaches(here, key, 15)
        {
            return true;
        }
        if here.starts_with("AIza") && reaches(here, key, 30) {
            return true;
        }
        if here.starts_with("eyJ") && runs.token_end[at] - at >= 30 && runs.token_dots[at] >= 2 {
            return true;
        }
        if starts("glpat-") && reaches(here, key, 14) {
            return true;
        }
        if MORE_PREFIXES.iter().any(|(prefix, least)| {
            starts(prefix) && reaches(&here[prefix.len()..], |b| b.is_ascii_alphanumeric(), *least)
        }) {
            return true;
        }
        if starts("bearer") {
            let after = &here["bearer".len()..];
            // The gap ends where a word starts, and a `bearer` word starts after it, so no
            // gap is read twice.
            let gap = run(after, |b| b == b' ' || b == b'\t');
            let token = |b: u8| b.is_ascii_alphanumeric() || b"._~+/=-".contains(&b);
            if gap > 0 && reaches(&after[gap..], token, 16) {
                return true;
            }
        }
        false
    })
}

/// Key names whose value is a secret wherever they stand, also as the end of a longer name
/// (`client_secret`, `ACCESS_TOKEN`).
const SECRET_KEYS: [&str; 6] = [
    "token",
    "secret",
    "password",
    "api_key",
    "apikey",
    "authorization",
];

/// A JSON, YAML or assignment pair whose key is one of [`SECRET_KEYS`] and whose value, on
/// the same line, is not empty: `"token": "x"`, `secret: x`, `PASSWORD = x`. A quote may be
/// escaped with `\`, as in a pair written inside a JSON string (`{\"password\": \"x\"}`, V2
/// audit N5); an empty escaped value (`\"\"`) is empty.
fn holds_secret_pair(text: &str) -> bool {
    let lower = text.to_ascii_lowercase();
    let blank = |b: u8| b == b' ' || b == b'\t';
    // `s` without a leading quote, escaped or not, and the quote that was there.
    let unquote = |s: &str| -> (Option<&'static str>, usize) {
        for quote in ["\\\"", "\\'", "\"", "'"] {
            if s.starts_with(quote) {
                return (Some(quote), quote.len());
            }
        }
        (None, 0)
    };
    SECRET_KEYS.iter().any(|key| {
        lower.match_indices(key).any(|(at, _)| {
            let mut rest = &lower[at + key.len()..];
            rest = &rest[unquote(rest).1..];
            rest = &rest[run(rest, blank)..];
            let Some(value) = rest.strip_prefix([':', '=']) else {
                return false;
            };
            let value = &value[run(value, blank)..];
            match unquote(value) {
                (Some(quote), len) => {
                    let inside = &value[len..];
                    !inside.is_empty() && !inside.starts_with(quote)
                }
                (None, _) => !value.is_empty() && !value.starts_with(['\n', '\r']),
            }
        })
    })
}

/// Shortest run of base64 characters that counts as a secret when it mixes upper case, lower
/// case and digits (audit N7).
pub const MIXED_RUN_MIN: usize = 48;

/// A run of [`MIXED_RUN_MIN`] or more characters of the base64 alphabets (`+ / =` and the
/// URL-safe `- _` included) with upper case, lower case and digits all in it. A commit id or
/// a sha256 digest (lower-case hex) has no upper case and is not one.
fn holds_long_mixed_run(text: &str) -> bool {
    let (mut len, mut upper, mut lower, mut digit) = (0, false, false, false);
    for b in text.bytes().chain([b' ']) {
        if b.is_ascii_alphanumeric() || b"+/=-_".contains(&b) {
            len += 1;
            upper |= b.is_ascii_uppercase();
            lower |= b.is_ascii_lowercase();
            digit |= b.is_ascii_digit();
            continue;
        }
        if len >= MIXED_RUN_MIN && upper && lower && digit {
            return true;
        }
        (len, upper, lower, digit) = (0, false, false, false);
    }
    false
}

/// Home-shaped folders, lower case: the path part after one of these is a person's name.
const HOME_PARENTS: [&str; 4] = [
    "/users/",
    "/home/",
    "\\users\\",
    "\\documents and settings\\",
];

/// Where a name after a [`HOME_PARENTS`] prefix ends when no path separator closes it (or at
/// white space).
const NAME_ENDS: &str = "/\\\"'`)]}>,;:";

/// Longest name, in bytes, and most spaces in it, for a name closed by a path separator
/// (`/Users/Test Person/`): past either, the name ends at its first break instead.
const NAME_MAX: usize = 64;
const NAME_SPACES_MAX: usize = 2;

/// Account names that say nothing about a person and are ordinary words in a question
/// (`the repository root`): never masked as a name.
const PLAIN_NAMES: [&str; 1] = ["root"];

/// `text` with the home folder as `~`, the user name as a whole word as `[user]` (also in a
/// host name made from it, `<name>s-MacBook-Pro`), and the name in any other home-shaped
/// path (`/Users/<name>`, `/home/<name>`, `\Users\<name>`, `\Documents and Settings\<name>`)
/// as `[user]`, a macOS temporary folder as `[tmp]` ([`mask_temp_folders`]), and an e-mail
/// address as `[email]` ([`mask_emails`], first, so an address made from the name goes
/// whole). The name is `who.user` and the last part of `who.home`. Letter case is
/// ignored when matching (ASCII), since the systems that use these folders ignore it too.
pub fn mask(text: &str, who: &Who) -> String {
    let mut out = mask_temp_folders(&mask_emails(text));
    let home = who
        .home
        .as_deref()
        .map(|h| h.trim_end_matches(['/', '\\']))
        .filter(|h| h.len() > 1);
    if let Some(home) = home {
        // The folder itself, not a sibling whose name only starts the same way.
        out = replace_outside_words(&out, home, HOME_MASK, |_, rest| {
            rest.chars()
                .next()
                .is_some_and(|c| c.is_alphanumeric() || c == '_' || c == '-')
        });
    }
    let from_home = home.and_then(|h| h.rsplit(['/', '\\']).next());
    for name in [who.user.as_deref(), from_home].into_iter().flatten() {
        if name.is_empty() || PLAIN_NAMES.contains(&name) {
            continue;
        }
        let in_word = |c: char| c.is_alphanumeric() || c == '_';
        out = replace_outside_words(&out, name, USER_MASK, |before, rest| {
            // `<name>s-…` is a host name made from the name, not another word.
            before.is_some_and(in_word)
                || (rest.chars().next().is_some_and(in_word)
                    && !starts_with_ignoring_case(rest, "s-"))
        });
    }
    for parent in HOME_PARENTS {
        out = mask_names_after(&out, parent);
    }
    out
}

/// `text` with each e-mail address as `[email]` (user decision U2). An address is a local
/// part of `A-Z a-z 0-9 . _ % + -`, an `@`, and a host of two or more labels of
/// `A-Z a-z 0-9 -` joined by `.`, the last of two or more letters; a `.` or `-` that ends
/// the host ends the sentence instead. `git@host:…` is an SSH remote, not a person's
/// address, and stays.
fn mask_emails(text: &str) -> String {
    let bytes = text.as_bytes();
    let local = |b: &u8| b.is_ascii_alphanumeric() || b"._%+-".contains(b);
    let host = |b: &u8| b.is_ascii_alphanumeric() || *b == b'.' || *b == b'-';
    let mut out = String::with_capacity(text.len());
    let mut from = 0;
    for (at, _) in text.match_indices('@') {
        if at < from {
            continue;
        }
        let start = at
            - bytes[from..at]
                .iter()
                .rev()
                .take_while(|b| local(b))
                .count();
        let mut end = at + 1 + bytes[at + 1..].iter().take_while(|b| host(b)).count();
        while end > at + 1 && matches!(bytes[end - 1], b'.' | b'-') {
            end -= 1;
        }
        // Only ASCII bytes were taken on either side, so both ends are character boundaries.
        let labels: Vec<&str> = text[at + 1..end].split('.').collect();
        let shaped = start < at
            && labels.len() >= 2
            && labels.iter().all(|label| !label.is_empty())
            && labels
                .last()
                .is_some_and(|tld| tld.len() >= 2 && tld.bytes().all(|b| b.is_ascii_alphabetic()));
        if !shaped || text[start..=at].eq_ignore_ascii_case("git@") {
            continue;
        }
        out.push_str(&text[from..start]);
        out.push_str(EMAIL_MASK);
        from = end;
    }
    out.push_str(&text[from..]);
    out
}

/// What stands in for a macOS per-user temporary folder.
pub const TEMP_MASK: &str = "[tmp]";

/// Where macOS keeps the per-user temporary folders, and the longest folder id read.
const TEMP_PARENT: &str = "/var/folders/";
const TEMP_ID_MAX: usize = 64;

/// `text` with each macOS temporary folder, `/var/folders/<2 characters>/<id>/T` or the
/// cache folder `…/C` beside it (slice V2; also under `/private`), as `[tmp]`. The id is
/// made from the user's account and says nothing a decision needs; with its `/` and the `T`
/// it would also read as a long mixed run (second audit, N6). `<2 characters>` and `<id>`
/// are lower-case letters, digits and `_`; the `T` or `C` ends the path or is followed by
/// something that is not part of a name.
fn mask_temp_folders(text: &str) -> String {
    let part = |b: u8| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'_';
    let mut out = String::with_capacity(text.len());
    let mut from = 0;
    while let Some(found) = text[from..].find(TEMP_PARENT) {
        let at = from + found;
        let rest = &text.as_bytes()[at + TEMP_PARENT.len()..];
        let id = rest
            .get(3..)
            .map_or(0, |r| r.iter().take_while(|b| part(**b)).count());
        let shaped = rest.len() >= 3
            && part(rest[0])
            && part(rest[1])
            && rest[2] == b'/'
            && (1..=TEMP_ID_MAX).contains(&id)
            && (rest[3 + id..].starts_with(b"/T") || rest[3 + id..].starts_with(b"/C"))
            && !rest
                .get(3 + id + 2)
                .is_some_and(|b| b.is_ascii_alphanumeric() || *b == b'_' || *b == b'-');
        if !shaped {
            out.push_str(&text[from..at + TEMP_PARENT.len()]);
            from = at + TEMP_PARENT.len();
            continue;
        }
        let start = if text[from..at].ends_with("/private") {
            at - "/private".len()
        } else {
            at
        };
        out.push_str(&text[from..start]);
        out.push_str(TEMP_MASK);
        // All of the path read so far is ASCII, so this is a character boundary.
        from = at + TEMP_PARENT.len() + 3 + id + 2;
    }
    out.push_str(&text[from..]);
    out
}

/// `text` with each `find` (any ASCII letter case) replaced by `with`, except where
/// `inside(character before, text after)` says it is part of something longer.
fn replace_outside_words(
    text: &str,
    find: &str,
    with: &str,
    inside: impl Fn(Option<char>, &str) -> bool,
) -> String {
    // ASCII lower-casing keeps every byte offset.
    let lower = text.to_ascii_lowercase();
    let mut out = String::with_capacity(text.len());
    let mut from = 0;
    for (at, _) in lower.match_indices(&find.to_ascii_lowercase()) {
        let end = at + find.len();
        if inside(text[..at].chars().next_back(), &text[end..]) {
            continue;
        }
        out.push_str(&text[from..at]);
        out.push_str(with);
        from = end;
    }
    out.push_str(&text[from..]);
    out
}

/// `text` with the name after each `parent` (lower case; matched in any ASCII letter case)
/// replaced by `[user]`. The name runs to the next path separator of the parent's kind when
/// that makes a plausible name (same line, at most [`NAME_MAX`] bytes and
/// [`NAME_SPACES_MAX`] spaces, no other break in it); otherwise to the first break.
fn mask_names_after(text: &str, parent: &str) -> String {
    let separator = if parent.ends_with('\\') { '\\' } else { '/' };
    let lower = text.to_ascii_lowercase();
    let mut out = String::with_capacity(text.len());
    let mut from = 0;
    while let Some(found) = lower[from..].find(parent) {
        let start = from + found + parent.len();
        out.push_str(&text[from..start]);
        from = start;
        let tail = &text[start..];
        if tail.starts_with(USER_MASK) {
            continue;
        }
        let line = &tail[..tail.find('\n').unwrap_or(tail.len())];
        let closed = line.find(separator).filter(|&end| {
            let name = &line[..end];
            (1..=NAME_MAX).contains(&end)
                && name.trim() == name
                && name.matches(' ').count() <= NAME_SPACES_MAX
                && !name.contains(|c: char| c == '\t' || NAME_ENDS.contains(c))
        });
        let end = closed.unwrap_or_else(|| {
            let end = tail
                .find(|c: char| c.is_whitespace() || NAME_ENDS.contains(c))
                .unwrap_or(tail.len());
            // A full stop that ends the sentence is not part of the name.
            tail[..end].trim_end_matches('.').len()
        });
        if end == 0 {
            continue;
        }
        out.push_str(USER_MASK);
        from = start + end;
    }
    out.push_str(&text[from..]);
    out
}

pub(super) fn too_large() -> Refusal {
    Refusal::new(
        "too_large",
        "a text field is larger than 64 KiB or the request body larger than 256 KiB: nothing \
         is sent, the leader decides",
    )
}

/// The outbound text of a cleared request, or the refusal of a field over [`FIELD_MAX`], of
/// the first [`screen`] hit in any field, of an option value the mask would change, or of a
/// field the mask made longer than its bound. The card and `undo` are not sent, but they are
/// recorded: they are screened and masked like the fields that are.
pub fn build(cleared: &Cleared<'_>, who: &Who) -> Result<Outbound, Refusal> {
    let request = cleared.request();
    // Each field on its own, with the fixed name of its place, so a reason never quotes the
    // request (an evidence field's name is the request's own text too).
    let mut fields = vec![("`what`", request.what.as_str())];
    for option in &request.options {
        fields.push(("an option value", &option.value));
        fields.push(("an option description", &option.description));
    }
    fields.extend(request.undo.as_deref().map(|undo| ("`undo`", undo)));
    fields.extend(request.card.as_deref().map(|card| ("the card", card)));
    for (name, value) in &request.evidence {
        fields.push(("an evidence field name", name));
        fields.push(("an evidence value", value));
    }
    // Size first: no screen reads a field longer than the bound.
    if fields.iter().any(|(_, text)| text.len() > FIELD_MAX) {
        return Err(too_large());
    }
    for (place, text) in &fields {
        if let Some(code) = screen(text) {
            let names = SCREEN
                .iter()
                .find(|(c, _)| *c == code)
                .map(|(_, names)| *names)
                .unwrap_or_default();
            return Err(Refusal::new(
                code,
                format!("{place} holds {names}: nothing is sent, the leader decides"),
            ));
        }
    }
    if request
        .options
        .iter()
        .any(|o| mask(&o.value, who) != o.value)
    {
        return Err(Refusal::new(
            "option_value_private",
            "an option value holds the home folder or the user name: give the option a short \
             name instead",
        ));
    }
    // Each name and value masked on its own, then rendered (escaped) as one text.
    let evidence: Vec<(String, String)> = request
        .evidence
        .iter()
        .map(|(name, value)| (mask(name, who), mask(value, who)))
        .collect();
    let outbound = Outbound {
        what: mask(&request.what, who),
        options: request
            .options
            .iter()
            .map(|o| (o.value.clone(), mask(&o.description, who)))
            .collect(),
        evidence: render_evidence(evidence.iter().map(|(name, value)| (name, value))),
        evidence_fields: evidence.iter().map(|(name, _)| name.clone()).collect(),
        undo: mask(request.undo.as_deref().unwrap_or_default(), who),
        card: request.card.as_deref().map(|card| mask(card, who)),
    };
    // `[user]` and `[email]` can be longer than what they replace: each sent field is held to
    // the request's bound again. The card is recorded, never sent, and keeps no bound here.
    let values = || evidence.iter().map(|(_, value)| value.len());
    let grown = if outbound.what.len() > WHAT_MAX {
        Some(("what_too_long", "`what` is longer than 500 bytes"))
    } else if outbound
        .options
        .iter()
        .any(|(_, description)| description.len() > DESCRIPTION_MAX)
    {
        Some((
            "description_too_long",
            "an option's description is longer than 300 bytes",
        ))
    } else if outbound.undo.len() > UNDO_MAX {
        Some(("undo_too_long", "`undo` is longer than 200 bytes"))
    } else if values().any(|len| len > EVIDENCE_FIELD_MAX) {
        Some((
            "evidence_field_too_large",
            "an evidence value is larger than 4 KiB",
        ))
    } else if values().sum::<usize>() > EVIDENCE_MAX {
        Some((
            "evidence_too_large",
            "the evidence values together are larger than 16 KiB",
        ))
    } else {
        None
    };
    if let Some((code, rule)) = grown {
        return Err(Refusal::new(
            code,
            format!(
                "{rule} once the home folder, the user name and e-mail addresses are masked: \
                 refused, never cut; the leader decides"
            ),
        ));
    }
    Ok(outbound)
}

#[cfg(test)]
mod tests {
    use super::super::request::{parse, Request};
    use super::super::testing::{evidence, request_json, who};
    use super::*;
    use serde_json::{json, Value};

    /// Past the gate without asking it: these cases are about this step alone.
    fn built(edit: impl FnOnce(&mut serde_json::Value)) -> Result<Outbound, Refusal> {
        let request = parse(&request_json(edit)).unwrap();
        build(&Cleared::unchecked(&request), &who())
    }

    /// Past the parser too: for the bounds of this step, which the parser's smaller bounds
    /// would otherwise hide.
    fn built_raw(edit: impl FnOnce(&mut serde_json::Value)) -> Result<Outbound, Refusal> {
        let request: Request = serde_json::from_str(&request_json(edit)).unwrap();
        build(&Cleared::unchecked(&request), &who())
    }

    /// Puts `text` in `field`; the evidence takes it as its one field `log`.
    fn put(v: &mut Value, field: &str, text: impl serde::Serialize) {
        if field == "evidence" {
            v["evidence"] = evidence(text);
        } else {
            v[field] = json!(text);
        }
    }

    /// The evidence is sent as one text, `name: value` per field in name order; its names
    /// are kept for the record, and `undo` is kept, masked and screened, but not sent.
    #[test]
    fn the_evidence_fields_are_rendered_in_name_order_and_undo_is_kept() {
        let out = built(|v| {
            v["evidence"] = json!({
                "zz_last": "the /Users/testperson/ci.yml line",
                "ci": "CI runs nextest",
                "testperson": "a field named like the user"
            });
            v["undo"] = json!("git revert the commit in /Users/testperson/proj");
        })
        .unwrap();
        assert_eq!(
            out.evidence(),
            "ci: CI runs nextest\n[user]: a field named like the user\nzz_last: the ~/ci.yml line"
        );
        assert_eq!(out.evidence_fields(), ["ci", "[user]", "zz_last"]);
        assert_eq!(out.undo(), "git revert the commit in ~/proj");
        let one = BTreeMap::from([("log".to_string(), "one\ntwo".to_string())]);
        assert_eq!(render_evidence(&one), "log: one\\ntwo");
        let key = format!("sk-{}", "a1".repeat(12));
        let r = built(|v| v["undo"] = json!(format!("put {key} back"))).unwrap_err();
        assert_eq!(r.code, "secret_shape");
    }

    /// V1 audit N8: a field's name and its value are screened apart, so a field named like a
    /// secret key with an innocent value is sent, and a secret in the value still refuses.
    /// A short password alone in a value is a stated limit (module text).
    #[test]
    fn field_names_and_values_are_screened_apart() {
        for name in [
            "token",
            "password",
            "secret",
            "api_key",
            "apikey",
            "authorization",
        ] {
            for value in [
                "the rotation policy says ninety days",
                "see the vault entry of the deploy step",
                "hunter2",
            ] {
                let out = built(|v| v["evidence"] = json!({ name: value }));
                assert!(out.is_ok(), "{name}: {value}");
                assert_eq!(out.unwrap().evidence(), format!("{name}: {value}"));
            }
        }
        let key = format!("sk-{}", "a1".repeat(12));
        for value in [
            key.clone(),
            format!("token: {key}"),
            "password: x".to_string(),
        ] {
            let r =
                built(|v| v["evidence"] = json!({"notes": "plain", "token": value})).unwrap_err();
            assert_eq!(r.code, "secret_shape", "{value}");
            assert!(r.reason.starts_with("an evidence value"), "{}", r.reason);
        }
        // A name is screened on its own too.
        let r = built(|v| v["evidence"] = json!({ format!("hf_{}", "a1".repeat(10)): "plain" }))
            .unwrap_err();
        assert_eq!(r.code, "secret_shape");
        assert!(
            r.reason.starts_with("an evidence field name"),
            "{}",
            r.reason
        );
    }

    /// V1 audit N8: a value with a line break and `name:` after it cannot add a field line.
    /// The rendering escapes (it is not a JSON object): `\` as `\\`, a line feed as `\n`, a
    /// carriage return as `\r`, U+2028 and U+2029 as ` ` and ` `; one line per field.
    #[test]
    fn a_value_cannot_forge_another_field_line() {
        let out = built(|v| {
            v["evidence"] = json!({
                "a": "first\nb: forged\nc: also\u{2028}d: and\u{2029}e: more",
                "b": "real \\n backslash",
            })
        })
        .unwrap();
        let lines: Vec<&str> = out.evidence().lines().collect();
        assert_eq!(
            lines,
            [
                "a: first\\nb: forged\\nc: also\\u2028d: and\\u2029e: more",
                "b: real \\\\n backslash",
            ]
        );
        assert!(!out.evidence().contains(['\u{2028}', '\u{2029}', '\r']));
        // A carriage return never passes the parser; the rendering escapes it all the same.
        let raw = BTreeMap::from([("a".to_string(), "x\r\ny: z".to_string())]);
        assert_eq!(render_evidence(&raw), "a: x\\r\\ny: z");
    }

    /// Test data: shaped like credentials, built at run time, none real.
    fn fakes() -> Vec<String> {
        let mut all: Vec<String> = crate::redact::secret_shapes()
            .into_iter()
            .filter(|shape| *shape != "sk-")
            .map(|shape| format!("{shape}{}", "0a".repeat(20)))
            .collect();
        all.push(format!("sk-{}", "a1".repeat(12)));
        all.push(format!("AKIA{}", "A1".repeat(8)));
        all.push(format!("xoxb-{}", "1".repeat(12)));
        all
    }

    #[test]
    fn a_secret_shape_in_any_sent_field_refuses_the_request() {
        for fake in fakes() {
            assert!(holds_secret_shape(&format!("see {fake} here")), "{fake}");
            for field in ["what", "evidence", "undo"] {
                let r = built(|v| put(v, field, format!("it printed {fake} twice"))).unwrap_err();
                assert_eq!(r.code, "secret_shape", "{field}");
                assert!(!r.reason.contains(&fake));
            }
            let r = built(|v| v["options"][1]["description"] = json!(fake.clone())).unwrap_err();
            assert_eq!(r.code, "secret_shape");
        }
    }

    /// Audit B3: a key glued to other characters, each literal, in each field the gate lets
    /// through (and the card). Refused, never masked and sent.
    #[test]
    fn a_key_glued_to_other_characters_refuses_the_request_in_every_field() {
        let key = format!("sk-{}", "a1".repeat(12));
        let glued = [
            format!("**{key}**"),
            format!("*{key}"),
            format!("\u{201c}{key}\u{201d}"),
            format!("\u{2018}{key}\u{2019}"),
            format!("key\u{ff1a}{key}"),
            format!("\u{d0a4}\u{b294}{key}\u{c785}\u{b2c8}\u{b2e4}"),
            format!("x\u{200b}{key}"),
            format!("s\u{200b}k-{}", "a1".repeat(12)),
            format!("sk\u{feff}-{}", "a1".repeat(12)),
            format!("+{key}"),
            format!("-{key}"),
            format!("KEY_{key}"),
            format!("cfg.{key}"),
            format!("${key}"),
            format!("~{key}"),
            format!("x%20{key}"),
        ];
        for text in &glued {
            assert!(holds_secret_shape(text), "{text}");
            for field in ["what", "evidence", "card", "undo"] {
                let r = built(|v| put(v, field, text)).unwrap_err();
                assert_eq!(r.code, "secret_shape", "{field}: {text}");
            }
            let r = built(|v| v["options"][1]["description"] = json!(text)).unwrap_err();
            assert_eq!(r.code, "secret_shape", "description: {text}");
            let r = built(|v| {
                v["options"][0]["value"] = json!(text);
                v["default"] = json!(text);
            })
            .unwrap_err();
            assert_eq!(r.code, "secret_shape", "value: {text}");
        }
    }

    /// Audit N6: every shared shape as an option value (the answer would name it).
    #[test]
    fn a_secret_shape_as_an_option_value_refuses_the_request() {
        for fake in fakes() {
            for at in [0, 1] {
                let r = built(|v| {
                    v["options"][at]["value"] = json!(fake.clone());
                    v["default"] = json!(fake.clone());
                })
                .unwrap_err();
                assert_eq!(r.code, "secret_shape", "{fake}");
                assert!(!r.reason.contains(&fake));
            }
        }
    }

    /// Audit N7 (leader decision): more shapes that refuse. A false hit only sends the choice
    /// back to the leader.
    #[test]
    fn bearer_tokens_key_value_pairs_and_long_mixed_runs_are_secret_shapes() {
        let long = "A1b2".repeat(12); // 48 characters, upper and lower case and digits
        for text in [
            "Authorization: Bearer abcdefghijklmnop".to_string(),
            "curl -H 'bearer  AbCdEfGh.0123456789' x".to_string(),
            "the deploy key is glpat-abcdefghij123456".to_string(),
            r#"{"token": "abc"}"#.to_string(),
            r#"{"Authorization":"Basic xyz"}"#.to_string(),
            "secret: hunter2".to_string(),
            "password = x".to_string(),
            "api_key: v".to_string(),
            "apikey=v".to_string(),
            "client_secret: 'v'".to_string(),
            "ACCESS_TOKEN=v".to_string(),
            format!("blob {long} end"),
            format!("blob={long}"),
        ] {
            assert!(holds_secret_shape(&text), "{text}");
        }
        let r = built(|v| v["evidence"] = evidence(format!("config has {long}"))).unwrap_err();
        assert_eq!(r.code, "secret_shape");
    }

    /// The miss direction of N7. A pair with no value, the words in prose, a run one short of
    /// 48, a run without all three kinds, and commit ids and sha256 digests (pure lower-case
    /// hex of 40 and 64 characters) are ordinary evidence.
    #[test]
    fn empty_pairs_prose_short_runs_and_hex_digests_are_not_secret_shapes() {
        let short = format!("{}A1b", "A1b2".repeat(11)); // 47 characters
        assert_eq!(short.len(), 47);
        let commit = "0123456789abcdef0123456789abcdef01234567";
        let digest = "0123456789abcdef".repeat(4);
        assert_eq!((commit.len(), digest.len()), (40, 64));
        for text in [
            r#"{"token": ""}"#.to_string(),
            "password:".to_string(),
            "secret: ''".to_string(),
            "token:\n  next line".to_string(),
            "a bearer token is sent by the client".to_string(),
            "the secret santa list and the password rule".to_string(),
            format!("blob {short} end"),
            "a".repeat(100),
            "ABCDEFGHIJKLabcdefghijkl".repeat(3),
            format!("at commit {commit}"),
            format!("sha256:{digest}"),
            format!("{commit} {digest}"),
        ] {
            assert!(!holds_secret_shape(&text), "{text}");
        }
        assert!(built(|v| v["evidence"] = evidence(format!("built from {commit}"))).is_ok());
    }

    /// Second audit, N7: the key prefixes of more services, alone, glued and in a sentence.
    /// Test data built at run time; the shortest tail that counts and one short of it.
    #[test]
    fn stripe_hugging_face_and_npm_keys_are_secret_shapes() {
        for (prefix, least) in [
            ("sk_live_", 16),
            ("sk_test_", 16),
            ("rk_live_", 16),
            ("rk_test_", 16),
            ("whsec_", 16),
            ("hf_", 20),
            ("npm_", 20),
        ] {
            let key = format!("{prefix}{}", "a1".repeat(least / 2));
            for text in [
                key.clone(),
                format!("**{key}**"),
                format!("it printed {key}."),
                format!("KEY_{key}"),
                key.to_ascii_uppercase(),
            ] {
                assert!(holds_secret_shape(&text), "{text}");
            }
            let r = built(|v| v["evidence"] = evidence(format!("log: {key}"))).unwrap_err();
            assert_eq!(r.code, "secret_shape", "{prefix}");
            let short = &key[..key.len() - 1];
            assert!(!holds_secret_shape(&format!("see {short} here")), "{short}");
        }
        for slack in ["xoxb-", "xoxp-"] {
            let key = format!("{slack}{}", "1".repeat(12));
            assert!(holds_secret_shape(&format!("**{key}**")), "{key}");
        }
        // Names that only start the same way: the tail of a key is letters and digits.
        for text in [
            "hf_hub_download(repo_id, filename)",
            "npm_config_cache and npm_package_version",
            "the sk_test_mode flag and sk_live_mode",
            "a shelf_abcdefghijklmnopqrstuvwxyz name",
        ] {
            assert!(!holds_secret_shape(text), "{text}");
        }
    }

    /// Second audit, N7: a key written in full-width letters, or with `\uXXXX` escapes in it,
    /// is the same key. Read for the refusal only: text that passes is sent as it was written.
    #[test]
    fn full_width_letters_and_unicode_escapes_do_not_hide_a_secret_shape() {
        let key = format!("sk-{}", "a1".repeat(12));
        let wide = |text: &str| -> String {
            text.chars()
                .map(|c| char::from_u32(c as u32 + 0xfee0).unwrap())
                .collect()
        };
        for text in [
            wide(&key),
            format!("{}{}", wide("sk-"), &key[3..]),
            format!("s{}", wide(&key[1..])),
            format!("\\u0073k-{}", &key[3..]),
            format!("sk\\u002d{}", &key[3..]),
            format!("sk\\u002D{}", &key[3..]),
            format!("sk-a1\\u200b{}", &key[5..]),
            format!("\\uff53\\uff4b\\uff0d{}", &key[3..]),
            "token\\u003a v".to_string(),
            format!("{}: v", wide("password")),
        ] {
            assert!(holds_secret_shape(&text), "{text}");
            let r = built(|v| v["evidence"] = evidence(format!("see {text}"))).unwrap_err();
            // Full-width letters fold before the patterns read; an escape is read by the
            // one decoding pass (design v2, slice V2).
            assert!(
                ["secret_shape", "secret_encoded"].contains(&r.code),
                "{text}: {}",
                r.code
            );
        }
        for text in [
            wide("Which runner?"),
            "caf\\u00e9 and C:\\users\\public and \\u12 and \\uzzzz".to_string(),
            "an escaped \\ud83d half of a pair".to_string(),
        ] {
            assert!(!holds_secret_shape(&text), "{text}");
        }
        let prose = format!("{} \\u0041", wide("Which runner?"));
        let out = built(|v| v["evidence"] = evidence(prose.clone())).unwrap();
        // Sent as written, with its backslash escaped by the rendering (slice V2).
        assert_eq!(
            out.evidence(),
            format!("log: {} \\\\u0041", wide("Which runner?"))
        );
    }

    /// Second audit, N6: `sk-` before lower-case words joined by further hyphens is a name
    /// (sk-learn-pipeline), as long as it holds no digit and no upper-case letter.
    #[test]
    fn a_hyphenated_lower_case_name_after_sk_is_not_a_secret_shape() {
        for text in [
            "sk-learn-pipeline",
            "the sk-learn-pipeline step, then (sk-learn-model-selection).",
            "`sk-learn-pipeline`",
        ] {
            assert!(!holds_secret_shape(text), "{text}");
            assert!(built(|v| v["evidence"] = evidence(text)).is_ok(), "{text}");
        }
        for text in [
            "sk-learn-pipeline2",
            "sk-learn-Pipeline",
            "sk-learnpipelinestep",
            "sk-learn_pipeline-step",
            "sk-learn--pipeline",
            "sk-learn-pipeline-",
            "token: sk-learn-pipeline",
            "%20sk-learn-pipeline",
        ] {
            assert!(holds_secret_shape(text), "{text}");
        }
    }

    /// Second audit, N6: a macOS temporary folder is masked like a home folder instead of
    /// being read as a long mixed run; what follows it is still screened.
    #[test]
    fn a_macos_temp_folder_is_masked_and_not_refused() {
        let folder = "/var/folders/zz/zyxvpxvq6csfxvn_n0000000000000/T";
        let who = who();
        for (text, want) in [
            (
                format!("{folder}/tmp.x1Yz/out.log failed"),
                "[tmp]/tmp.x1Yz/out.log failed",
            ),
            (format!("/private{folder}/a.txt"), "[tmp]/a.txt"),
            (format!("in {folder}."), "in [tmp]."),
            (format!("TMPDIR={folder}/"), "TMPDIR=[tmp]/"),
            // The per-user cache folder, `C`, beside `T` (design v2, slice V2).
            (
                "/var/folders/zz/zyxvpxvq6csfxvn_n0000000000000/C/com.apple.x".to_string(),
                "[tmp]/com.apple.x",
            ),
        ] {
            assert!(!holds_secret_shape(&text), "{text}");
            assert_eq!(mask(&text, &who), want);
            let out = built(|v| v["evidence"] = evidence(text.clone())).unwrap();
            assert_eq!(out.evidence(), format!("log: {want}"));
        }
        // Not that folder: another last part, a longer first part, a name that goes on.
        for text in [
            "/var/folders/zz/zyxvpxvq6csfxvn_n0000000000000/X/com.apple.x".to_string(),
            "/var/folders/zzz/zyxvpxvq6csfxvn_n0000000000000/T/a".to_string(),
            format!("{folder}mp/a"),
        ] {
            assert_eq!(mask(&text, &who), text);
        }
        let key = format!("sk-{}", "a1".repeat(12));
        for text in [
            format!("{folder}/{key}"),
            format!("{folder}/{}", "A1b2".repeat(12)),
        ] {
            assert!(holds_secret_shape(&text), "{text}");
        }
        // An option value is never rewritten.
        let r = built(|v| {
            v["options"][0]["value"] = json!(format!("{folder}/a"));
            v["default"] = json!(format!("{folder}/a"));
        })
        .unwrap_err();
        assert_eq!(r.code, "option_value_private");
    }

    /// Audit N10: one field at its edge, 65,536 bytes kept and 65,537 refused, raw and after
    /// the mask made it longer. The parser's bounds are smaller (design v2), so this step's
    /// own bound is reached past the parser; the rendered evidence counts its `log: ` too.
    #[test]
    fn a_field_over_64_kib_is_refused() {
        assert_eq!(FIELD_MAX, 65_536);
        // At 65,536 bytes the screen reads the field, and the request's own bound (checked
        // again after masking) refuses it; at 65,537 the screen never reads it.
        for (field, at_bound) in [
            ("what", "what_too_long"),
            ("evidence", "evidence_field_too_large"),
        ] {
            let r = built_raw(|v| put(v, field, "a".repeat(65_536))).unwrap_err();
            assert_eq!(r.code, at_bound, "{field}");
            let r = built_raw(|v| put(v, field, "a".repeat(65_537))).unwrap_err();
            assert_eq!(r.code, "too_large", "{field}");
        }
        let r =
            built_raw(|v| v["options"][1]["description"] = json!("a".repeat(65_536))).unwrap_err();
        assert_eq!(r.code, "description_too_long");
        let r =
            built_raw(|v| v["options"][1]["description"] = json!("a".repeat(65_537))).unwrap_err();
        assert_eq!(r.code, "too_large");
    }

    /// Item 6 of slice V2: after masking, each sent field is checked against the request's
    /// bound again; a field the mask made too long is refused, never cut. `al` becomes
    /// `[user]`, four bytes longer. N and N+1 for each bound.
    #[test]
    fn a_field_the_mask_made_too_long_is_refused() {
        use super::super::request::{
            DESCRIPTION_MAX, EVIDENCE_FIELD_MAX, EVIDENCE_MAX, UNDO_MAX, WHAT_MAX,
        };
        let al = Who {
            home: None,
            user: Some("al".into()),
        };
        // A text of `n` bytes, starting with the name, that is `n + 4` bytes once masked.
        let grows = |n: usize| format!("al {}", "a".repeat(n - 3));
        let with = |edit: &dyn Fn(&mut Value)| {
            let request = parse(&request_json(edit)).unwrap();
            build(&Cleared::unchecked(&request), &al)
        };
        for (bound, code, edit) in [
            (
                WHAT_MAX,
                "what_too_long",
                &(|v: &mut Value, t: String| v["what"] = json!(t)) as &dyn Fn(&mut Value, String),
            ),
            (DESCRIPTION_MAX, "description_too_long", &|v, t| {
                v["options"][1]["description"] = json!(t)
            }),
            (UNDO_MAX, "undo_too_long", &|v, t| v["undo"] = json!(t)),
            (EVIDENCE_FIELD_MAX, "evidence_field_too_large", &|v, t| {
                v["evidence"] = evidence(t)
            }),
        ] {
            assert!(with(&|v| edit(v, grows(bound - 4))).is_ok(), "{code}");
            let r = with(&|v| edit(v, grows(bound - 3))).unwrap_err();
            assert_eq!(r.code, code);
            assert!(r.reason.contains("mask"), "{}", r.reason);
        }
        // The sum of the values: 16,380 bytes before the mask and 16,384 after it fit; one
        // more byte in a fifth field is 16,381 before (the parser lets it through) and
        // 16,385 after.
        assert_eq!(EVIDENCE_MAX, 16_384);
        let fields = |fifth: bool| {
            let mut v = json!({
                "a": "e".repeat(4_096),
                "b": "e".repeat(4_096),
                "c": "e".repeat(4_096),
                "d": grows(4_092),
            });
            if fifth {
                v["e"] = json!("e");
            }
            v
        };
        let fits = with(&|v| v["evidence"] = fields(false)).unwrap();
        assert_eq!(fits.evidence_fields().len(), 4);
        let r = with(&|v| v["evidence"] = fields(true)).unwrap_err();
        assert_eq!(r.code, "evidence_too_large");
    }

    /// Audit N9: names with a space, other letter cases, a host name made from the name, and
    /// the old Windows home folder.
    #[test]
    fn names_with_spaces_other_cases_and_host_names_are_masked() {
        let w = who();
        let nobody = Who::default();
        for (text, want) in [
            ("/Users/Test Person/x", "/Users/[user]/x"),
            ("C:\\Users\\Bob Smith\\src", "C:\\Users\\[user]\\src"),
            (
                "/USERS/alice/x and /Home/bob/y",
                "/USERS/[user]/x and /Home/[user]/y",
            ),
            (
                "C:\\Documents and Settings\\carol\\Desktop",
                "C:\\Documents and Settings\\[user]\\Desktop",
            ),
            // No separator closes the name: it ends at the first break, as before.
            ("see /home/alice.", "see /home/[user]."),
            (
                "/home/alice then four more words /tmp",
                "/home/[user] then four more words /tmp",
            ),
        ] {
            assert_eq!(mask(text, &nobody), want, "{text}");
        }
        for (text, want) in [
            ("/users/TESTPERSON/x by TestPerson", "~/x by [user]"),
            (
                "host testpersons-MacBook-Pro.local",
                "host [user]s-MacBook-Pro.local",
            ),
            ("host Testpersons-iMac", "host [user]s-iMac"),
        ] {
            assert_eq!(mask(text, &w), want, "{text}");
        }
        let spaced = Who {
            home: Some("/Users/Test Person".into()),
            user: None,
        };
        assert_eq!(
            mask("/Users/Test Person/proj by Test Person", &spaced),
            "~/proj by [user]"
        );
    }

    #[test]
    fn the_card_is_kept_masked_for_the_record() {
        let out = built(|v| v["card"] = json!("testperson-C-37")).unwrap();
        assert_eq!(out.card(), Some("[user]-C-37"));
        let none = built(|v| {
            v.as_object_mut().unwrap().remove("card");
        })
        .unwrap();
        assert_eq!(none.card(), None);
    }

    /// A key glued to a label or wrapped in punctuation is still a key.
    #[test]
    fn punctuation_does_not_hide_a_secret_shape() {
        let key = format!("sk-{}", "a1".repeat(12));
        for text in [
            format!("Authorization:{key}"),
            format!("header=Bearer,{key};"),
            format!("{{\"k\":\"{key}\"}}"),
            format!("see <{key}>"),
        ] {
            assert!(holds_secret_shape(&text), "{text}");
        }
    }

    /// The miss direction: ordinary words that only contain a shape's letters are sent.
    #[test]
    fn ordinary_text_is_not_a_secret_shape() {
        for text in [
            "a risk-based, task-level choice",
            "disk-full handling: ask-first or retry",
            "the token limit and the password rule (no value here)",
            "sk-short",
            "CI runs nextest; the local docs name both runners.",
        ] {
            assert!(!holds_secret_shape(text), "{text}");
        }
        assert!(built(|v| v["evidence"] = evidence("a risk-based, task-level choice")).is_ok());
    }

    #[test]
    fn the_home_folder_and_the_user_name_are_masked() {
        let w = who();
        for (text, want) in [
            ("/Users/testperson/proj/Cargo.toml", "~/proj/Cargo.toml"),
            ("see /Users/testperson.", "see ~."),
            ("cd /Users/testperson && ls", "cd ~ && ls"),
            // Another folder that only starts like the home: its name is masked as a name.
            ("/Users/testperson2/x", "/Users/[user]/x"),
            // Someone else's home, on each system's layout.
            (
                "/Users/other/x and /home/alice/y",
                "/Users/[user]/x and /home/[user]/y",
            ),
            ("C:\\Users\\bob\\src", "C:\\Users\\[user]\\src"),
            // The name alone, as a whole word only.
            (
                "owner: testperson, branch testperson/fix",
                "owner: [user], branch [user]/fix",
            ),
            (
                "testpersons and mytestperson stay",
                "testpersons and mytestperson stay",
            ),
            ("no path here", "no path here"),
        ] {
            assert_eq!(mask(text, &w), want, "{text}");
        }
    }

    #[test]
    fn masking_uses_only_what_it_is_given() {
        // Nobody to mask: only the home-shaped paths change.
        let nobody = Who::default();
        assert_eq!(
            mask("/home/alice/x testperson", &nobody),
            "/home/[user]/x testperson"
        );
        // A home without a name still gives one: its last part.
        let home_only = Who {
            home: Some("/home/carol".into()),
            user: None,
        };
        assert_eq!(mask("/home/carol/a by carol", &home_only), "~/a by [user]");
        // A home of `/` or an empty name masks nothing by itself.
        let root = Who {
            home: Some("/".into()),
            user: Some(String::new()),
        };
        assert_eq!(mask("/etc/hosts and more", &root), "/etc/hosts and more");
        // The account `root` names nobody: its home is masked, the word stays a word.
        let superuser = Who {
            home: Some("/root".into()),
            user: Some("root".into()),
        };
        assert_eq!(
            mask("/root/proj is the repository root", &superuser),
            "~/proj is the repository root"
        );
    }

    #[test]
    fn the_sent_fields_are_masked_and_nothing_else_is_kept() {
        let out = built(|v| {
            v["what"] = json!("Which runner for /Users/testperson/proj?");
            v["options"][0]["description"] = json!("as testperson set it up");
            v["evidence"] = evidence("/Users/testperson/proj/ci.yml runs nextest");
        })
        .unwrap();
        assert_eq!(out.what(), "Which runner for ~/proj?");
        assert_eq!(
            out.options(),
            [
                ("nextest".to_string(), "as [user] set it up".to_string()),
                ("cargo-test".to_string(), "plain cargo test".to_string()),
            ]
        );
        assert_eq!(out.evidence(), "log: ~/proj/ci.yml runs nextest");
    }

    /// An option value is what the answer names, so it is never rewritten: a value the mask
    /// would change refuses the request.
    #[test]
    fn an_option_value_the_mask_would_change_is_refused() {
        let r = built(|v| {
            v["options"][0]["value"] = json!("testperson");
            v["default"] = json!("testperson");
        })
        .unwrap_err();
        assert_eq!(r.code, "option_value_private");
        assert!(!r.reason.contains("testperson"));
    }

    /// Test data for slice V2, built at run time: shaped like keys and not keys.
    fn key() -> String {
        format!("sk-{}", "a1".repeat(12))
    }

    fn wide(text: &str) -> String {
        text.chars()
            .map(|c| char::from_u32(c as u32 + 0xfee0).unwrap())
            .collect()
    }

    fn base64(bytes: &[u8]) -> String {
        const A: &[u8] = b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
        let mut out = String::new();
        for chunk in bytes.chunks(3) {
            let n = chunk.iter().fold(0u32, |n, b| n << 8 | *b as u32) << (8 * (3 - chunk.len()));
            for i in 0..=chunk.len() {
                out.push(A[(n >> (18 - 6 * i) & 63) as usize] as char);
            }
            for _ in chunk.len()..3 {
                out.push('=');
            }
        }
        out
    }

    /// Every bypass the audits of slices S2 and V1 listed, with the code it is refused with.
    fn bypasses() -> Vec<(&'static str, String, &'static str)> {
        let key = key();
        let tail = &key[3..];
        let hex: String = key.bytes().map(|b| format!("{b:02x}")).collect();
        vec![
            ("bold markers", format!("**{key}**"), "secret_shape"),
            (
                "Hangul adjacency",
                format!("\u{d0a4}\u{b294}{key}\u{c785}\u{b2c8}\u{b2e4}"),
                "secret_shape",
            ),
            ("KEY_ prefix", format!("KEY_{key}"), "secret_shape"),
            ("%20 before", format!("x%20{key}"), "secret_shape"),
            (
                "zero-width before",
                format!("x\u{200b}{key}"),
                "secret_shape",
            ),
            (
                "zero-width inside",
                format!("s\u{200b}k-{tail}"),
                "secret_shape",
            ),
            (
                "soft hyphen inside",
                format!("sk\u{ad}-{tail}"),
                "secret_shape",
            ),
            (
                "tag character inside",
                format!("sk-a1\u{e0041}{}", &key[5..]),
                "secret_shape",
            ),
            ("full-width", wide(&key), "secret_shape"),
            (
                "full-width prefix",
                format!("{}{tail}", wide("sk-")),
                "secret_shape",
            ),
            ("long s", format!("\u{17f}k-{tail}"), "secret_shape"),
            ("plain sk-", key.clone(), "secret_shape"),
            (
                "\\U00000073",
                format!("\\U00000073k-{tail}"),
                "secret_encoded",
            ),
            ("\\x73", format!("\\x73k-{tail}"), "secret_encoded"),
            ("\\u{73}", format!("\\u{{73}}k-{tail}"), "secret_encoded"),
            ("\\u0073", format!("\\u0073k-{tail}"), "secret_encoded"),
            ("%73", format!("%73k-{tail}"), "secret_encoded"),
            ("&#115;", format!("&#115;k-{tail}"), "secret_encoded"),
            ("&#x73;", format!("&#x73;k-{tail}"), "secret_encoded"),
            ("Cyrillic s", format!("\u{455}k-{tail}"), "mixed_script"),
            (
                "Greek omicron",
                format!("t\u{3bf}ken-{tail}"),
                "mixed_script",
            ),
            ("base64 of a key", base64(key.as_bytes()), "secret_encoded"),
            ("hex of a key", hex, "secret_encoded"),
            (
                "sk_live_",
                format!("sk_live_{}", "a1".repeat(8)),
                "secret_shape",
            ),
            (
                "sk_test_",
                format!("sk_test_{}", "a1".repeat(8)),
                "secret_shape",
            ),
            (
                "rk_live_",
                format!("rk_live_{}", "a1".repeat(8)),
                "secret_shape",
            ),
            (
                "rk_test_",
                format!("rk_test_{}", "a1".repeat(8)),
                "secret_shape",
            ),
            (
                "whsec_",
                format!("whsec_{}", "a1".repeat(8)),
                "secret_shape",
            ),
            (
                "whsec_ lower case",
                format!("whsec_{}", "abcdefgh".repeat(2)),
                "secret_shape",
            ),
            (
                "sk-proj- lower case",
                format!("sk-proj-{}", "abcdefgh".repeat(3)),
                "secret_shape",
            ),
            ("hf_", format!("hf_{}", "a1".repeat(10)), "secret_shape"),
            ("npm_", format!("npm_{}", "a1".repeat(10)), "secret_shape"),
            ("xoxb-", format!("xoxb-{}", "1".repeat(12)), "secret_shape"),
            (
                "glued after letters",
                format!("mysk_live_{}", "a1".repeat(8)),
                "secret_shape",
            ),
            (
                "sk- glued after letters",
                format!("my{key}"),
                "secret_shape",
            ),
            (
                "random token",
                "q7Zr2LkX9pWm4TfV8sYb".to_string(),
                "high_entropy",
            ),
        ]
    }

    /// Each bypass, refused with its code in every field the request carries: `what`, an
    /// option description, `undo`, an evidence value, and, where it fits their shorter
    /// bounds, the card and an option value. The reason never quotes it.
    #[test]
    fn every_listed_bypass_is_refused_in_every_field() {
        for (name, text, code) in bypasses() {
            assert_eq!(screen(&text), Some(code), "{name}: {text}");
            for field in ["what", "evidence", "undo"] {
                let r = built(|v| put(v, field, format!("it printed {text} twice"))).unwrap_err();
                assert_eq!(r.code, code, "{name} in {field}");
                assert!(!r.reason.contains(&text), "{name}: {}", r.reason);
            }
            let r = built(|v| v["options"][1]["description"] = json!(text)).unwrap_err();
            assert_eq!(r.code, code, "{name} in a description");
            if text.len() <= 64 {
                let r = built(|v| v["card"] = json!(text)).unwrap_err();
                assert_eq!(r.code, code, "{name} in the card");
                let r = built(|v| {
                    v["options"][0]["value"] = json!(text);
                    v["default"] = json!(text);
                })
                .unwrap_err();
                assert_eq!(r.code, code, "{name} in an option value");
            }
            // Through the whole chain too: the gate does not take another code first.
            let r = super::super::prepare(
                &request_json(|v| v["evidence"] = json!({"ci": "plain", "log": text})),
                super::super::provider::Provider::Typesafe,
                &who(),
            )
            .unwrap_err();
            assert_eq!(r.code, code, "{name} through prepare");
        }
        // The key prefixes that fit an evidence field name are refused there too.
        for name in [
            format!("sk_live_{}", "a1".repeat(8)),
            format!("rk_test_{}", "a1".repeat(8)),
            format!("whsec_{}", "a1".repeat(8)),
            format!("hf_{}", "a1".repeat(10)),
            format!("npm_{}", "a1".repeat(10)),
            format!("mysk_live_{}", "a1".repeat(8)),
        ] {
            let r = built(|v| v["evidence"] = json!({ name.clone(): "plain" })).unwrap_err();
            assert_eq!(r.code, "secret_shape", "{name}");
            assert!(!r.reason.contains(&name));
        }
    }

    /// The miss direction of slice V2: ordinary engineering text passes every screen and is
    /// sent (masked where it names a person).
    #[test]
    fn ordinary_engineering_text_passes_every_screen() {
        let run_url = format!(
            "{}://github.com/modu-ai/moai-adk/actions/runs/19234567890/job/54321098765",
            "https"
        );
        let ordinary = [
            "at commit c7ef9b1f8c2365fe08963269af718f3d48de7d51".to_string(),
            format!(
                "sha256:{}",
                "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
            ),
            "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08".to_string(),
            "run cf998b8a-4360-419c-b982-edacba5e85da failed".to_string(),
            "max_tokens: 4096".to_string(),
            run_url,
            "CARGO_TARGET_DIR=target/v2".to_string(),
            "ProviderResponseValidationError".to_string(),
            "ProviderResponseValidationErrorKindUnknown".to_string(),
            "src/decide/outbound.rs:540 and tests/home_isolation_scan.rs".to_string(),
            "/Users/testperson/MoAI/moai-adk/.claude/worktrees/agent-ad8f65b06b51e1d65".to_string(),
            "refs/heads/card/37-v2-screen".to_string(),
            "x86_64-unknown-linux-gnu and linux-x86_64-musl-static-v2026".to_string(),
            "the_largest_v2_request_fits_the_body_bound".to_string(),
            "a_key_glued_to_other_characters_refuses_the_request_in_every_field".to_string(),
            "feedback_new_worktrees_only_in_moai_home".to_string(),
            "github-actions-runner-ubuntu-22-04-x64".to_string(),
            "v2026.11.0-rc.2+local.4".to_string(),
            "cargo clippy --locked --all-targets -- -D warnings".to_string(),
            "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>".to_string(),
            "the risk-based-v2-rollout-plan and task-level-configuration".to_string(),
            "work_test_suiteforintegration".to_string(),
            "/var/folders/zz/2b_k5x8n1m7q9_3h8r0z9y4w0000gn/T/tmp.x1Yz".to_string(),
            "키는 값입니다: CI runs nextest".to_string(),
            "10\u{3bc}s per call, \u{3b1} = 0.5".to_string(),
            "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz".to_string(),
            "12345678901234567890123456789".to_string(),
            // Found by the false-hit measurement of slice V2 (a branch name in a merge commit).
            "Merge pull request #2332 from modu-ai/issue/2330-a9-followup3".to_string(),
            "images/ubuntu/ubuntu2404-readme and sha256sum-checker2".to_string(),
        ];
        for text in &ordinary {
            assert_eq!(screen(text), None, "{text}");
            assert!(built(|v| v["evidence"] = evidence(text)).is_ok(), "{text}");
        }
    }

    /// Shannon entropy, and the rule's two edges: 19 characters pass and 20 are refused (all
    /// distinct, 4.25 and 4.32 bits); at 32 characters 4.0 bits exactly are refused and 3.98
    /// pass. A token without a digit, or without a letter, is never refused by it.
    #[test]
    fn the_entropy_rule_refuses_at_twenty_characters_and_four_bits() {
        assert_eq!((ENTROPY_LEN_MIN, ENTROPY_BITS_MIN), (20, 4.0));
        assert_eq!(entropy("ab"), 1.0);
        assert_eq!(entropy("aaaa"), 0.0);
        // Letters and digits change places three times, so no part reads as a word.
        let nineteen = "abcdefgh1ijklmnopq2";
        let twenty = "abcdefgh1ijklmnopq23";
        assert_eq!((nineteen.len(), twenty.len()), (19, 20));
        assert!(entropy(nineteen) >= 4.0);
        assert_eq!(screen(nineteen), None);
        assert_eq!(screen(twenty), Some("high_entropy"));
        let at = "0123456789ghijkl".repeat(2);
        assert_eq!(entropy(&at), 4.0);
        assert_eq!(screen(&at), Some("high_entropy"));
        let below = format!("{}0", &at[..31]);
        assert!(
            entropy(&below) < 4.0 && entropy(&below) > 3.97,
            "{}",
            entropy(&below)
        );
        assert_eq!(screen(&below), None);
        // No digit, or no letter: passes at any entropy.
        assert_eq!(screen("abcdefghijklmnopqrstuvwxyz"), None);
        assert_eq!(screen("12345678901234567890"), None);
        // The same token, split by a space, is two short tokens.
        assert_eq!(screen("abcdefghij klmnopqr12"), None);
        // A path whose every part is an ordinary word passes; one random part refuses it.
        assert_eq!(
            screen("moai-adk/actions/runs/19234567890/job/54321098765"),
            None
        );
        assert_eq!(
            screen("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
            Some("high_entropy")
        );
        // A lower-case part is a word when its letters and digits change places at most
        // twice (`followup3`, `a9b`); three changes (`a9b8`) in a long part read as random.
        assert!(is_ordinary_part("followup3") && is_ordinary_part("ubuntu2404x"));
        assert!(is_ordinary_part("abcdefghi9z"));
        assert!(!is_ordinary_part("abcdefghi9z8"));
        assert_eq!(screen("x9q1m4k2p7z3w8r5t6y0"), Some("high_entropy"));
        assert_eq!(
            screen("modu-ai/issue/2330-x9q1m4k2p7"),
            Some("high_entropy")
        );
    }

    /// A word that mixes ASCII letters with Cyrillic or Greek letters is refused; a word in one
    /// script, or a Greek letter beside a number, is not.
    #[test]
    fn a_word_mixing_ascii_with_cyrillic_or_greek_letters_is_refused() {
        for text in [
            "\u{455}k-live",
            "p\u{430}ssword",
            "t\u{3bf}ken",
            "AKI\u{391}",
            "my_\u{3c3}ecret-x",
        ] {
            assert_eq!(screen(text), Some("mixed_script"), "{text}");
        }
        for text in [
            "\u{43f}\u{440}\u{438}\u{432}\u{435}\u{442} world",
            "\u{3b1}\u{3b2}\u{3b3} and abc",
            "10\u{3bc}s",
            "caf\u{e9}",
            "\u{d0a4}abc",
        ] {
            assert_eq!(screen(text), None, "{text}");
        }
        // The Greek mu is the micro sign only right after a digit (`10μs`); elsewhere it is a
        // look-alike of `u` like any other Greek letter.
        for text in ["tok\u{3bc}n", "\u{3bc}ser_x", "10\u{3bc}s_t\u{3bc}"] {
            assert_eq!(screen(text), Some("mixed_script"), "{text}");
        }
    }

    /// Slice V2: a key prefix glued after a letter or digit counts from 16 letters and digits
    /// with a digit among them: 15 pass and 16 refuse; 16 without a digit pass.
    #[test]
    fn a_prefix_glued_after_letters_needs_sixteen_letters_and_digits_with_a_digit() {
        assert_eq!(GLUED_TAIL_MIN, 16);
        for prefix in ["sk_live_", "rk_test_", "whsec_", "sk-", "AKIA"] {
            let tail = |n: usize| format!("{}{}", "1", "a".repeat(n - 1));
            let at = format!("my{prefix}{}", tail(16));
            let below = format!("my{prefix}{}", tail(15));
            let no_digit = format!("my{prefix}{}", "a".repeat(16));
            assert_eq!(screen(&at), Some("secret_shape"), "{at}");
            assert_eq!(screen(&below), None, "{below}");
            assert_eq!(screen(&no_digit), None, "{no_digit}");
        }
    }

    /// Slice V2: `sk-proj-` starts a key whatever letters follow it, from the 13 characters
    /// of the shared `sk-` rule: 12 pass and 13 refuse.
    #[test]
    fn sk_proj_counts_from_thirteen_characters() {
        assert_eq!(screen("sk-proj-abcd"), None);
        assert_eq!(screen("sk-proj-abcde"), Some("secret_shape"));
        assert_eq!(
            screen("see sk-proj-abcdefghijklmnopqrstuvwx."),
            Some("secret_shape")
        );
    }

    /// V2 audit N1: a password in a URL's user part is refused with its own code before
    /// anything is masked, in every field: with a dotted host the e-mail mask would otherwise
    /// hide only part of it and send the rest. Test data, none real.
    #[test]
    fn a_url_with_a_password_is_refused_in_every_field_before_masking() {
        // The web schemes are joined at run time: the scan of this folder in
        // tests/home_isolation_scan.rs allows the two endpoint constants only.
        let web = |scheme: &str, rest: &str| format!("{scheme}://{rest}");
        for text in [
            "postgres://admin:S3cretPass@localhost:5432/app".to_string(),
            "redis://:pa55w0rdX@cache:6379/0".to_string(),
            web("https", "user:hunter2@host.internal/x"),
            "postgres://admin:S3cretPass@db.example.com/app".to_string(),
            "://u:p@h".to_string(),
            "://:p@h".to_string(),
        ] {
            let text = text.as_str();
            assert_eq!(screen(text), Some("url_credentials"), "{text}");
            for field in ["what", "evidence", "undo"] {
                let r = built(|v| put(v, field, format!("it uses {text} now"))).unwrap_err();
                assert_eq!(r.code, "url_credentials", "{field}: {text}");
                assert!(!r.reason.contains(text), "{}", r.reason);
            }
            let r = built(|v| v["options"][1]["description"] = json!(text)).unwrap_err();
            assert_eq!(r.code, "url_credentials", "description: {text}");
            let r = built(|v| v["card"] = json!(text)).unwrap_err();
            assert_eq!(r.code, "url_credentials", "card: {text}");
            let r = built(|v| {
                v["options"][0]["value"] = json!(text);
                v["default"] = json!(text);
            })
            .unwrap_err();
            assert_eq!(r.code, "url_credentials", "value: {text}");
        }
        // A user without a password, an SSH remote, a port, a `:` after the path.
        for text in [
            "://u@h".to_string(),
            "git@host:org/repo".to_string(),
            "ssh://git@host/x".to_string(),
            "git@github.com:org/repo.git".to_string(),
            web("https", "host.internal:8443/x"),
            format!(
                "{} and {}",
                web("http", "localhost:3000"),
                web("https", "x.org/a:b@c")
            ),
        ] {
            let text = text.as_str();
            assert_eq!(screen(text), None, "{text}");
            assert!(built(|v| v["evidence"] = evidence(text)).is_ok(), "{text}");
        }
    }

    /// V2 audit N5: a pair written inside a JSON string, with its quotes escaped, is a pair.
    #[test]
    fn a_json_escaped_secret_pair_is_refused() {
        for text in [
            r#"{\"password\": \"Tr0ub4dor&3\"}"#,
            r#"{\"token\":\"abc\"}"#,
            r#"\'secret\': \'v\'"#,
        ] {
            assert_eq!(screen(text), Some("secret_shape"), "{text}");
        }
        for text in [r#"{\"password\": \"\"}"#, r#"{\"token\": \'\'}"#] {
            assert_eq!(screen(text), None, "{text}");
        }
    }

    /// V2 audit N2, reworked for card C-111: every key shape is read in time linear in the
    /// text, and the test pins that by a ratio against a baseline measured in the same
    /// process, never by wall-clock time — at load average 37 the C-93 audit saw this test
    /// take 6.3 s and fail the old 2 s bound while the scan itself stayed linear, and under
    /// load the rework reproduced it 6 of 6 times (2026-11-19). The baseline is one linear
    /// pass over the text's bytes in this crate, measured just before each `screen` call, so
    /// machine load scales both sides of the ratio. Measured in a debug build on 2026-11-19:
    /// a 64 KiB field screens in 6 to 1,100 baseline passes across the 14 shapes below, while
    /// a quadratic re-scan — reading the text again at each occurrence, the 2026-11-17 defect
    /// — costs at least about 160,000 of them at this text length; `COST_RATIO_LIMIT`
    /// separates the two with room on both sides (the mutant run is in the SPEC Result).
    #[test]
    fn a_64_kib_field_of_a_repeated_key_prefix_is_screened_quickly() {
        const COST_RATIO_LIMIT: f64 = 20_000.0;
        let fill = |unit: &str| unit.repeat(FIELD_MAX / unit.len());
        for text in [
            fill("AKIA"),
            fill("ASIA"),
            fill("AIza"),
            fill("eyJ-"),
            fill("eyJ."),
            fill("xAKIA1"),
            // Glued after a letter with no digit in the tail: the whole tail is read at each
            // prefix and none is a hit.
            fill("xAKIA"),
            fill("xASIA"),
            fill("xAIza"),
            fill("sk-"),
            fill("xoxb-"),
            fill("glpat-"),
            fill("bearer "),
            fill("://a"),
        ] {
            // One in-crate linear pass over the same bytes is the reference the ratio
            // divides by; `black_box` keeps a release build from folding the pass away.
            let text: &str = std::hint::black_box(text.as_str());
            let started = std::time::Instant::now();
            let mut probe: usize = 0;
            for byte in text.bytes() {
                probe = probe.wrapping_add(byte as usize);
            }
            std::hint::black_box(probe);
            let baseline = started.elapsed();
            let started = std::time::Instant::now();
            let _ = screen(text);
            let screened = started.elapsed();
            let ratio = screened.as_secs_f64() / baseline.as_secs_f64().max(1e-9);
            assert!(
                ratio < COST_RATIO_LIMIT,
                "{}: screened {screened:?} over {baseline:?} of baseline passes, ratio {ratio:.0} at limit {COST_RATIO_LIMIT}",
                &text[..8]
            );
        }
    }

    /// A bidirectional control character is refused in every screened field, by the screen
    /// itself (the gate refuses it earlier in the chain, with the same code).
    #[test]
    fn a_bidi_character_is_refused_by_the_screen() {
        for field in ["what", "evidence", "undo"] {
            let r = built(|v| put(v, field, "see \u{202e}txt.exe")).unwrap_err();
            assert_eq!(r.code, "bidi_character", "{field}");
        }
    }

    /// User decision U2: an e-mail address is masked as `[email]` and sent; an SSH remote
    /// (`git@host:…`), a package version and a decorator are not addresses.
    #[test]
    fn e_mail_addresses_are_masked() {
        let w = who();
        for (text, want) in [
            ("mail testperson@example.com now", "mail [email] now"),
            ("a.b+c@sub.example.co.kr.", "[email]."),
            (
                "Co-Authored-By: Claude <noreply@anthropic.com>",
                "Co-Authored-By: Claude <[email]>",
            ),
            ("TESTPERSON@EXAMPLE.ORG", "[email]"),
            ("x@y.io,z@w.dev", "[email],[email]"),
        ] {
            assert_eq!(mask(text, &w), want, "{text}");
        }
        for text in [
            "@types/node",
            "pkg@1.2.3",
            "user@localhost",
            "git@github.com:org/repo.git",
            "a@b",
            "@ alone",
            "x@.com",
            "x@y.c",
        ] {
            assert_eq!(mask(text, &Who::default()), text, "{text}");
        }
        let out = built(|v| v["evidence"] = evidence("reported by someone@example.com")).unwrap();
        assert_eq!(out.evidence(), "log: reported by [email]");
        let r = built(|v| {
            v["options"][0]["value"] = json!("someone@example.com");
            v["default"] = json!("someone@example.com");
        })
        .unwrap_err();
        assert_eq!(r.code, "option_value_private");
    }

    /// Each refusal code of the screen has its own reason, which names the field's place and
    /// the rule, never the text.
    #[test]
    fn each_screen_code_has_a_reason_that_quotes_nothing() {
        let codes: Vec<&str> = bypasses().iter().map(|(_, _, code)| *code).collect();
        for code in [
            "secret_shape",
            "secret_encoded",
            "mixed_script",
            "high_entropy",
        ] {
            assert!(codes.contains(&code), "{code}");
        }
        let marker = "q7Zr2LkX9pWm4TfV8sYb";
        let r = built(|v| v["what"] = json!(format!("Which {marker}?"))).unwrap_err();
        assert_eq!(r.code, "high_entropy");
        assert!(r.reason.starts_with("`what`"), "{}", r.reason);
        assert!(!r.reason.contains(marker));
    }
}
