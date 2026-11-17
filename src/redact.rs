//! Shared secret-shape detection and message redaction (Issue #1996, 03-B owns the
//! module; scope decision 2026-11-08).
//!
//! One shared list, one predicate and one redacting formatter so callers refuse or
//! mask the same shapes instead of restating the list. The list is the union of the
//! conservative profile shapes (`crate::SECRET_SHAPES`, lib.rs) and the merged gh
//! token prefixes (`crate::forge::transport::TOKEN_PREFIXES`) — the prefixes stay
//! spelled only where they were already merged (#2351). forge::report consumes
//! this module's predicate instead of carrying its own copy. It is a conservative shape filter, not a secret detector; the common
//! redaction policy (what may be logged at all) belongs to the caller.

use crate::forge::transport::TOKEN_PREFIXES;

/// The shared shape list as one sequence: the conservative profile shapes plus the
/// merged gh token prefixes (#2351). Tests build their synthetic tokens from this
/// at runtime — no caller restates the prefixes.
pub fn secret_shapes() -> Vec<&'static str> {
    crate::SECRET_SHAPES
        .iter()
        .chain(TOKEN_PREFIXES.iter())
        .copied()
        .collect()
}

/// True when `s` carries a credential-shaped substring (lower-cased match): the profile shapes by
/// [`crate::secret_shaped`] (`sk-` only at the start of a word, card C-90) or a gh token prefix
/// anywhere.
pub fn contains_secret_shape(s: &str) -> bool {
    let lower = s.to_ascii_lowercase();
    crate::secret_shaped(s) || TOKEN_PREFIXES.iter().any(|shape| lower.contains(shape))
}

/// Replaces every credential-shaped key and the value that follows it (up to the
/// next whitespace or the end of the string) with `[redacted]`, so diagnostics can
/// quote the surrounding context without carrying the secret value.
pub fn redact_secret_shapes(s: &str) -> String {
    let mut out = s.to_string();
    for shape in crate::SECRET_SHAPES.iter().chain(TOKEN_PREFIXES.iter()) {
        loop {
            let lower = out.to_ascii_lowercase();
            let Some(start) = lower.find(shape) else {
                break;
            };
            let value_end = out[start..]
                .find(char::is_whitespace)
                .map(|i| start + i)
                .unwrap_or(out.len());
            out.replace_range(start..value_end, "[redacted]");
        }
    }
    out
}

/// Text that is about to be posted where anyone can read it (#3071 H4, audit of #3119 B1).
///
/// Works word by word (a word ends at whitespace) and replaces a whole word that has the shape
/// of a credential with `[redacted]`. Returns the text and how many words were replaced, so the
/// caller can warn before anything is posted. It adds to [`redact_secret_shapes`] on purpose:
/// that function matches its short shapes anywhere inside a word (it would turn `risk-based`
/// into `ri[redacted]`), which is right for a diagnostic and wrong for a title a person reads.
/// Here a prefix shape must start the word and be long enough to be a key:
///
/// - the merged gh token prefixes anywhere in a word (they are specific);
/// - `key=value` shapes (`password=`, `token=`, `api_key=`) and a PEM header anywhere in a word;
/// - `sk-` (13 or more characters), `AKIA`/`ASIA` plus 16 upper-case letters or digits,
///   `xox[abprs]-` (15 or more), `AIza` (30 or more) and a three-part `eyJ…` token (30 or more),
///   each at the start of the word after opening punctuation.
///
/// It is a conservative shape filter, not a secret detector.
pub fn redact_outgoing(s: &str) -> (String, usize) {
    let mut count = 0;
    let out: String = s
        .split_inclusive(char::is_whitespace)
        .map(|chunk| {
            let word = chunk.trim_end();
            if word.is_empty() || !word_has_secret_shape(word) {
                return chunk.to_string();
            }
            count += 1;
            format!("[redacted]{}", &chunk[word.len()..])
        })
        .collect();
    (out, count)
}

fn word_has_secret_shape(word: &str) -> bool {
    let lower = word.to_ascii_lowercase();
    let anywhere =
        TOKEN_PREFIXES
            .iter()
            .copied()
            .chain(["password=", "token=", "api_key=", "-----begin"]);
    if anywhere.into_iter().any(|shape| lower.contains(shape)) {
        return true;
    }
    let core = word.trim_start_matches(|c: char| "\"'`([{<".contains(c));
    let core_lower = core.to_ascii_lowercase();
    if core_lower.starts_with("sk-") && core.chars().count() >= 13 {
        return true;
    }
    for prefix in ["AKIA", "ASIA"] {
        if let Some(rest) = core.strip_prefix(prefix) {
            let key: String = rest
                .chars()
                .take_while(|c| c.is_ascii_uppercase() || c.is_ascii_digit())
                .collect();
            if key.len() >= 16 {
                return true;
            }
        }
    }
    if core_lower.starts_with("xox")
        && core.len() >= 15
        && matches!(
            core.as_bytes().get(3),
            Some(b'a' | b'b' | b'p' | b'r' | b's')
        )
        && core.as_bytes().get(4) == Some(&b'-')
    {
        return true;
    }
    if core.starts_with("AIza") && core.len() >= 30 {
        return true;
    }
    core.starts_with("eyJ") && core.len() >= 30 && core.matches('.').count() >= 2
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn outgoing_text_loses_key_shaped_words_and_keeps_ordinary_ones() {
        let (out, n) = redact_outgoing(
            "ping sk-live-abcdefghijklmnop AKIAABCDEFGHIJKLMNOP token=abc risk-based task-list \"sk-short\"",
        );
        assert_eq!(n, 3, "{out}");
        assert!(!out.contains("abcdefghijklmnop"), "{out}");
        assert!(!out.contains("AKIA"), "{out}");
        assert!(!out.contains("token=abc"), "{out}");
        assert!(out.contains("risk-based task-list"), "{out}");
        assert!(out.contains("\"sk-short\""), "{out}");
        let (same, none) = redact_outgoing("Add login for the admin page\nsecond line");
        assert_eq!(
            (same.as_str(), none),
            ("Add login for the admin page\nsecond line", 0)
        );
    }

    #[test]
    fn outgoing_text_covers_each_prefix_shape_at_its_length_boundary() {
        // sk- needs 13 characters: 12 stays, 13 goes.
        assert_eq!(redact_outgoing("sk-123456789").1, 0);
        assert_eq!(redact_outgoing("sk-1234567890").1, 1);
        // AKIA/ASIA need 16 more upper-case letters or digits.
        assert_eq!(redact_outgoing("AKIA123456789012345").1, 0);
        assert_eq!(redact_outgoing("AKIA1234567890123456").1, 1);
        assert_eq!(redact_outgoing("ASIA1234567890123456").1, 1);
        assert_eq!(redact_outgoing("AKIAlowercaselowercase").1, 0);
        assert_eq!(redact_outgoing("xoxb-1234567890").1, 1);
        assert_eq!(redact_outgoing("xoxz-1234567890").1, 0);
        assert_eq!(
            redact_outgoing("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.c2lnbmF0dXJl").1,
            1
        );
        assert_eq!(redact_outgoing("eyJhbGciOiJIUzI1NiJ9").1, 0);
        assert_eq!(redact_outgoing("-----BEGIN").1, 1);
    }

    #[test]
    fn detects_the_shared_shapes_without_echoing() {
        assert!(contains_secret_shape("token=abc123"));
        assert!(contains_secret_shape("see PASSWORD=topsecret"));
        assert!(!contains_secret_shape("plain value"));
        let line = "failed at token=abc123 near the end";
        let redacted = redact_secret_shapes(line);
        assert!(!redacted.contains("abc123"), "{redacted}");
        assert!(redacted.contains("[redacted]"));
        assert!(redacted.contains("near the end"));
    }

    /// Card C-90: `sk-` counts only at the start of a word, so ordinary words with `sk-` inside
    /// pass; every other shape still counts anywhere, inside a word too.
    #[test]
    fn sk_counts_only_at_the_start_of_a_word() {
        for ordinary in [
            "task-based",
            "risk-free",
            "desk-top",
            "Kiosk-Mode",
            "ask-first then RISK-FREE",
        ] {
            assert!(!crate::secret_shaped(ordinary), "{ordinary}");
            assert!(!contains_secret_shape(ordinary), "{ordinary}");
        }
        // The edge of the rule is the character before `sk-`: an ASCII letter or digit keeps it
        // inside a word; its neighbours in ASCII order (`/` `:` `@` `[` `` ` `` `{`), a space,
        // `_`, `-`, a non-ASCII letter or nothing at all start a word.
        for before in ['0', '9', 'a', 'z', 'A', 'Z'] {
            let s = format!("{before}sk-abc");
            assert!(!crate::secret_shaped(&s), "{s}");
        }
        for before in [
            "", " ", "/", ":", "@", "[", "`", "{", "_", "-", "é", "(", "\"",
        ] {
            let s = format!("{before}sk-abc");
            assert!(crate::secret_shaped(&s), "{s:?}");
            assert!(contains_secret_shape(&s), "{s:?}");
            assert!(crate::secret_shaped(&s.to_uppercase()), "{s:?}");
        }
        // A later word start still counts after an ordinary word.
        assert!(crate::secret_shaped("task-based sk-abc"));
        assert!(crate::secret_shaped("sk-"));
        for shape in crate::SECRET_SHAPES.iter().filter(|s| **s != "sk-") {
            let s = format!("x{shape}y");
            assert!(crate::secret_shaped(&s), "{s}");
        }
    }

    #[test]
    fn the_merged_gh_prefixes_are_part_of_the_shared_list() {
        // The four gh-shaped prefixes (start "gh", end "_", four chars) from the
        // merged transport list must be known to the shared predicate — spelled
        // here only through the list itself, never as literals.
        let gh: Vec<&str> = secret_shapes()
            .iter()
            .filter(|s| s.starts_with("gh") && s.ends_with('_') && s.len() == 4)
            .copied()
            .collect();
        assert!(
            gh.len() >= 4,
            "the merged gh token prefixes must be part of the shared shapes: {gh:?}"
        );
        for prefix in gh {
            assert!(contains_secret_shape(&format!("{prefix}synthetic-value")));
        }
    }
}
