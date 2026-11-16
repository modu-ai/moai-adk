//! The text of a `moai decide` request that would leave the machine (card C-37 S2): the
//! question, the options and the evidence, and nothing else. The card, the default and the
//! deadline are never sent.
//!
//! Three rules, in this order:
//! 1. A field of more than 64 KiB refuses the request ([`FIELD_MAX`]); the whole body has its
//!    own bound in [`super::BODY_MAX`]. Refused, never cut: a cut text is another question.
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
//! 3. The home folder and the user name are masked ([`mask`]): they say who and where, and a
//!    decision needs neither. An option value that would change under the mask is refused
//!    instead, because the answer names an option by its value.
//!
//! All of it is a conservative filter over known shapes, not a secret detector.

use super::gate::Cleared;
use super::Refusal;

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
    card: Option<String>,
}

/// Largest single text field, in bytes (64 KiB), before and after masking.
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

    pub fn evidence(&self) -> &str {
        &self.evidence
    }
}

/// True when `text` holds a credential shape (see the module rules). Read twice: as written,
/// where a character without width is a break like any other, and with those characters
/// taken out, so one placed inside a key does not split it.
pub fn holds_secret_shape(text: &str) -> bool {
    let reads = |t: &str| {
        crate::redact::redact_outgoing(t).1 > 0
            || holds_glued_shape(t)
            || holds_secret_pair(t)
            || holds_long_mixed_run(t)
    };
    if reads(text) {
        return true;
    }
    let joined: String = text.chars().filter(|c| !ZERO_WIDTH.contains(c)).collect();
    joined.len() != text.len() && reads(&joined)
}

/// Characters that take no room on the screen: zero-width space, non-joiner and joiner, word
/// joiner, byte-order mark, soft hyphen.
const ZERO_WIDTH: [char; 6] = [
    '\u{200b}', '\u{200c}', '\u{200d}', '\u{2060}', '\u{feff}', '\u{00ad}',
];

fn starts_with_ignoring_case(text: &str, prefix: &str) -> bool {
    text.len() >= prefix.len()
        && text.as_bytes()[..prefix.len()].eq_ignore_ascii_case(prefix.as_bytes())
}

/// How many leading bytes of `text` satisfy `part`.
fn run(text: &str, part: impl Fn(u8) -> bool) -> usize {
    text.bytes().take_while(|b| part(*b)).count()
}

/// A known key prefix right after any character that is not an ASCII letter or digit (or at
/// the start, or after a `%XX` escape): `**sk-…**`, `KEY_sk-…`, a key after Korean text or a
/// curly quote. The prefixes and lengths are those of [`crate::redact::redact_outgoing`],
/// plus `glpat-` and `Bearer <token>` (audit N7).
fn holds_glued_shape(text: &str) -> bool {
    let bytes = text.as_bytes();
    let key = |b: u8| b.is_ascii_alphanumeric() || b == b'_' || b == b'-';
    (0..bytes.len()).any(|at| {
        if !bytes[at].is_ascii_alphanumeric() {
            return false;
        }
        let escaped = at >= 3
            && bytes[at - 3] == b'%'
            && bytes[at - 2].is_ascii_hexdigit()
            && bytes[at - 1].is_ascii_hexdigit();
        if at > 0 && bytes[at - 1].is_ascii_alphanumeric() && !escaped {
            return false;
        }
        // `at` follows an ASCII byte or starts a character, so the slice is valid.
        let here = &text[at..];
        let starts = |prefix: &str| starts_with_ignoring_case(here, prefix);
        if starts("sk-") && run(here, key) >= 13 {
            return true;
        }
        if (here.starts_with("AKIA") || here.starts_with("ASIA"))
            && run(&here[4..], |b| b.is_ascii_uppercase() || b.is_ascii_digit()) >= 16
        {
            return true;
        }
        if starts("xox")
            && matches!(bytes.get(at + 3), Some(b'a' | b'b' | b'p' | b'r' | b's'))
            && bytes.get(at + 4) == Some(&b'-')
            && run(here, key) >= 15
        {
            return true;
        }
        if here.starts_with("AIza") && run(here, key) >= 30 {
            return true;
        }
        if here.starts_with("eyJ") {
            let token = &here[..run(here, |b| key(b) || b == b'.')];
            if token.len() >= 30 && token.matches('.').count() >= 2 {
                return true;
            }
        }
        if starts("glpat-") && run(here, key) >= 14 {
            return true;
        }
        if starts("bearer") {
            let after = &here["bearer".len()..];
            let gap = run(after, |b| b == b' ' || b == b'\t');
            let token = |b: u8| b.is_ascii_alphanumeric() || b"._~+/=-".contains(&b);
            if gap > 0 && run(&after[gap..], token) >= 16 {
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
/// the same line, is not empty: `"token": "x"`, `secret: x`, `PASSWORD = x`.
fn holds_secret_pair(text: &str) -> bool {
    let lower = text.to_ascii_lowercase();
    let blank = |b: u8| b == b' ' || b == b'\t';
    SECRET_KEYS.iter().any(|key| {
        lower.match_indices(key).any(|(at, _)| {
            let mut rest = &lower[at + key.len()..];
            rest = rest.strip_prefix(['"', '\'']).unwrap_or(rest);
            rest = &rest[run(rest, blank)..];
            let Some(value) = rest.strip_prefix([':', '=']) else {
                return false;
            };
            let value = &value[run(value, blank)..];
            let mut chars = value.chars();
            match chars.next() {
                None | Some('\n' | '\r') => false,
                Some(quote @ ('"' | '\'')) => chars.next().is_some_and(|c| c != quote),
                Some(_) => true,
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
/// as `[user]`. The name is `who.user` and the last part of `who.home`. Letter case is
/// ignored when matching (ASCII), since the systems that use these folders ignore it too.
pub fn mask(text: &str, who: &Who) -> String {
    let mut out = text.to_string();
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
/// rule 1, or of an option value the mask would change. The card is not sent, but it is
/// recorded: it is screened and masked like the fields that are.
pub fn build(cleared: &Cleared<'_>, who: &Who) -> Result<Outbound, Refusal> {
    let request = cleared.request();
    let texts = || {
        [request.what.as_str(), request.evidence.as_str()]
            .into_iter()
            .chain(request.card.as_deref())
            .chain(
                request
                    .options
                    .iter()
                    .flat_map(|o| [o.value.as_str(), o.description.as_str()]),
            )
    };
    // Size first: nothing below reads a field longer than the bound.
    if texts().any(|text| text.len() > FIELD_MAX) {
        return Err(too_large());
    }
    if texts().any(holds_secret_shape) {
        return Err(Refusal::new(
            "secret_shape",
            "the text to send holds a credential shape: nothing is sent, the leader decides",
        ));
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
    let outbound = Outbound {
        what: mask(&request.what, who),
        options: request
            .options
            .iter()
            .map(|o| (o.value.clone(), mask(&o.description, who)))
            .collect(),
        evidence: mask(&request.evidence, who),
        card: request.card.as_deref().map(|card| mask(card, who)),
    };
    // `[user]` can be longer than the name it replaces.
    let grown = [outbound.what.len(), outbound.evidence.len()]
        .into_iter()
        .chain(
            outbound
                .options
                .iter()
                .map(|(_, description)| description.len()),
        )
        .any(|len| len > FIELD_MAX);
    if grown {
        return Err(too_large());
    }
    Ok(outbound)
}

#[cfg(test)]
mod tests {
    use super::super::request::parse;
    use super::super::testing::{request_json, who};
    use super::*;
    use serde_json::json;

    /// Past the gate without asking it: these cases are about this step alone.
    fn built(edit: impl FnOnce(&mut serde_json::Value)) -> Result<Outbound, Refusal> {
        let request = parse(&request_json(edit)).unwrap();
        build(&Cleared::unchecked(&request), &who())
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
            for field in ["what", "evidence"] {
                let r =
                    built(|v| v[field] = json!(format!("it printed {fake} twice"))).unwrap_err();
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
            for field in ["what", "evidence", "card"] {
                let r = built(|v| v[field] = json!(text)).unwrap_err();
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
        let r = built(|v| v["evidence"] = json!(format!("config has {long}"))).unwrap_err();
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
            "abcdefghijkl0123456789".repeat(3),
            "ABCDEFGHIJKLabcdefghijkl".repeat(3),
            format!("at commit {commit}"),
            format!("sha256:{digest}"),
            format!("{commit} {digest}"),
        ] {
            assert!(!holds_secret_shape(&text), "{text}");
        }
        assert!(built(|v| v["evidence"] = json!(format!("built from {commit}"))).is_ok());
    }

    /// Audit N10: one field at its edge, 65,536 bytes kept and 65,537 refused, raw and after
    /// the mask made it longer.
    #[test]
    fn a_field_over_64_kib_is_refused() {
        assert_eq!(FIELD_MAX, 65_536);
        for field in ["what", "evidence"] {
            assert!(
                built(|v| v[field] = json!("a".repeat(65_536))).is_ok(),
                "{field}"
            );
            let r = built(|v| v[field] = json!("a".repeat(65_537))).unwrap_err();
            assert_eq!(r.code, "too_large", "{field}");
        }
        assert!(built(|v| v["options"][1]["description"] = json!("a".repeat(65_536))).is_ok());
        let r = built(|v| v["options"][1]["description"] = json!("a".repeat(65_537))).unwrap_err();
        assert_eq!(r.code, "too_large");
        // `al` becomes `[user]`: four bytes longer, so a field that fit no longer does.
        let al = Who {
            home: None,
            user: Some("al".into()),
        };
        let grown = |n: usize| {
            let request = parse(&request_json(|v| {
                v["evidence"] = json!(format!("al {}", "a".repeat(n - 3)))
            }))
            .unwrap();
            build(&Cleared::unchecked(&request), &al)
        };
        assert_eq!(grown(65_532).unwrap().evidence().len(), 65_536);
        assert_eq!(grown(65_533).unwrap_err().code, "too_large");
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
        assert!(built(|v| v["evidence"] = json!("a risk-based, task-level choice")).is_ok());
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
            v["evidence"] = json!("/Users/testperson/proj/ci.yml runs nextest");
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
        assert_eq!(out.evidence(), "~/proj/ci.yml runs nextest");
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
}
