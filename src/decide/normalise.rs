//! The one reading of a request text that every screen of `moai decide` takes (design v2,
//! slice V2): the secret screen ([`super::outbound::screen`]), the deny list
//! ([`super::gate::deny`]) and the reserved-value check of the parser all read a text only
//! after [`normalise`] has read it. Hand-rolled, no new crate (leader decision 2026-11-17).
//!
//! [`normalise`], in this order:
//! 1. refuses a bidirectional control character, U+202A to U+202E and U+2066 to U+2069
//!    (`bidi_character`), and a C0 or C1 control character other than tab and line feed
//!    (`control_character`): text that reads differently from how it is stored;
//! 2. takes out the default-ignorable code points ([`is_default_ignorable`]): characters
//!    without width that split a word for a pattern but not for a reader;
//! 3. folds the full-width forms U+FF01 to U+FF5E and the ideographic space U+3000 to ASCII,
//!    the long s `ſ` (U+017F) to `s` and the Kelvin sign (U+212A) to `K`.
//!
//! It is a reading for the screens only: a text that passes is sent as it was written.
//!
//! [`decode_once`] reads one layer of encoding (percent escapes, `\uXXXX`, `\U00XXXXXX`,
//! `\xXX`, `\u{…}`, HTML numeric entities, base64 runs and hex runs), so a screen can read the
//! text a second time. One layer only: an encoding inside an encoding is a stated limit of
//! the screen.

/// The code of a bidirectional control character.
pub const BIDI: &str = "bidi_character";
/// The code of a C0 or C1 control character other than tab and line feed.
pub const CONTROL: &str = "control_character";

/// True for a bidirectional embedding, override or isolate character.
fn is_bidi(c: char) -> bool {
    false
}

/// True for a default-ignorable code point (Unicode `Default_Ignorable_Code_Point`, the
/// ranges the design names).
pub fn is_default_ignorable(c: char) -> bool {
    false
}

/// `text` as the screens read it, or the code of the character that refuses it.
pub fn normalise(text: &str) -> Result<String, &'static str> {
    Ok(text.to_string())
}

/// `text` with each default-ignorable code point as a space instead of taken out: the reading
/// "as written", where such a character breaks a word like any other break.
pub fn spaced(text: &str) -> String {
    text.to_string()
}

/// The lenient reading of decoded text: like [`normalise`], but a character it would refuse
/// becomes a space (a decoded `%0A` is a line break, not a forged one).
pub fn fold(text: &str) -> String {
    text.to_string()
}

/// Shortest base64 run that is read, in characters (padding not counted).
pub const BASE64_MIN: usize = 16;
/// Shortest hex run that is read, in characters.
pub const HEX_MIN: usize = 32;

/// `text` with one layer of encoding read, or `None` when it holds no encoding this reads.
pub fn decode_once(text: &str) -> Option<String> {
    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn bidi_and_control_characters_are_refused_with_their_own_codes() {
        for c in [
            '\u{202a}', '\u{202b}', '\u{202c}', '\u{202d}', '\u{202e}', '\u{2066}', '\u{2067}',
            '\u{2068}', '\u{2069}',
        ] {
            assert_eq!(normalise(&format!("a{c}b")), Err(BIDI), "{:x}", c as u32);
        }
        for c in [
            '\u{0}', '\u{7}', '\u{b}', '\u{c}', '\r', '\u{1b}', '\u{1f}', '\u{7f}', '\u{80}',
            '\u{85}', '\u{9f}',
        ] {
            assert_eq!(normalise(&format!("a{c}b")), Err(CONTROL), "{:x}", c as u32);
        }
        // Tab and line feed are text; the neighbours of the bidi ranges are not bidi.
        assert_eq!(normalise("a\tb\nc").unwrap(), "a\tb\nc");
        for c in ['\u{2029}', '\u{202f}', '\u{2065}', '\u{206a}', '\u{a0}'] {
            assert!(
                !normalise(&format!("a{c}b")).is_err_and(|code| code == BIDI),
                "{:x}",
                c as u32
            );
        }
    }

    /// The first and the last code point of every range the design names is taken out; the
    /// code point just outside a range is kept.
    #[test]
    fn default_ignorable_code_points_are_taken_out() {
        let stripped: [(u32, u32); 18] = [
            (0x00ad, 0x00ad),
            (0x034f, 0x034f),
            (0x061c, 0x061c),
            (0x115f, 0x1160),
            (0x17b4, 0x17b5),
            (0x180b, 0x180f),
            (0x200b, 0x200f),
            (0x2060, 0x2065),
            (0x206a, 0x206f),
            (0x3164, 0x3164),
            (0xfe00, 0xfe0f),
            (0xfeff, 0xfeff),
            (0xffa0, 0xffa0),
            (0x1bca0, 0x1bca3),
            (0x1d173, 0x1d17a),
            (0xe0000, 0xe0000),
            (0xe0001, 0xe007f),
            (0xe0100, 0xe0fff),
        ];
        for (first, last) in stripped {
            for code in [first, last] {
                let c = char::from_u32(code).unwrap();
                assert!(is_default_ignorable(c), "{code:x}");
                assert_eq!(normalise(&format!("s{c}k")).unwrap(), "sk", "{code:x}");
                assert_eq!(spaced(&format!("s{c}k")), "s k", "{code:x}");
            }
        }
        for code in [
            0x00ac, 0x00ae, 0x034e, 0x0350, 0x115e, 0x1161, 0x17b3, 0x17b6, 0x180a, 0x200a,
            0x2010, 0x3163, 0x3165, 0xfdff, 0xfe10, 0xff9f, 0xffa1, 0x1d172, 0x1d17b, 0xe1000,
        ] {
            let c = char::from_u32(code).unwrap();
            assert!(!is_default_ignorable(c), "{code:x}");
        }
        // Korean text is not touched: syllables are not fillers.
        assert_eq!(normalise("키는 값입니다").unwrap(), "키는 값입니다");
    }

    #[test]
    fn full_width_forms_the_long_s_and_the_kelvin_sign_fold_to_ascii() {
        assert_eq!(normalise("\u{ff01}\u{ff5e}").unwrap(), "!~");
        assert_eq!(normalise("\u{ff53}\u{ff4b}\u{ff0d}").unwrap(), "sk-");
        assert_eq!(normalise("a\u{3000}b").unwrap(), "a b");
        assert_eq!(normalise("\u{17f}k-x \u{212a}EY").unwrap(), "sk-x KEY");
        // Just outside the full-width range: kept as written.
        for c in ['\u{ff00}', '\u{ff5f}', '\u{ff61}'] {
            assert_eq!(normalise(&c.to_string()).unwrap(), c.to_string());
        }
    }

    #[test]
    fn fold_reads_decoded_text_without_refusing() {
        assert_eq!(fold("a\u{202e}b\u{1b}c\u{200b}d\u{ff53}"), "a b cds");
        assert_eq!(fold("one\ntwo\tx"), "one\ntwo\tx");
    }

    #[test]
    fn each_escape_kind_is_read_once() {
        for (text, want) in [
            ("%73k-x", "sk-x"),
            ("%D1%95", "\u{455}"),
            ("\\u0073k", "sk"),
            ("\\U00000073k", "sk"),
            ("\\x73k", "sk"),
            ("\\u{73}k", "sk"),
            ("\\u{1F600}", "\u{1f600}"),
            ("&#115;k", "sk"),
            ("&#x73;k", "sk"),
            ("&#X73;k", "sk"),
            ("a%2520b", "a%20b"),
        ] {
            assert_eq!(decode_once(text).as_deref(), Some(want), "{text}");
        }
        // Nothing this reads, or an escape that names nothing: no decoded reading.
        for text in [
            "plain text",
            "100% sure",
            "%zz and %4",
            "\\uzzzz \\u12 \\ud83d \\U0011ffff \\u{110000} \\u{}",
            "&#; &#x; &amp; &#1114112;",
            "C:\\users\\public",
        ] {
            assert_eq!(decode_once(text), None, "{text}");
        }
    }

    /// Test data: built at run time, shaped like a key and not one.
    fn fake_key() -> String {
        format!("sk-{}", "a1".repeat(12))
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

    fn hex(bytes: &[u8]) -> String {
        bytes.iter().map(|b| format!("{b:02x}")).collect()
    }

    #[test]
    fn a_base64_run_of_sixteen_or_more_that_reads_as_text_is_decoded() {
        assert_eq!(BASE64_MIN, 16);
        let key = fake_key();
        let coded = base64(key.as_bytes());
        assert!(decode_once(&format!("blob {coded} end")).unwrap().contains(&key));
        assert!(decode_once(&format!("KEY_{coded}")).unwrap().contains(&key));
        // 12 bytes are 16 characters (decoded); 11 bytes are 15 without padding (not read).
        let twelve = base64(b"hello world!");
        assert_eq!(twelve.len(), 16);
        assert!(decode_once(&twelve).unwrap().contains("hello world!"));
        let eleven = base64(b"hello world");
        assert_eq!(eleven.trim_end_matches('=').len(), 15);
        assert_eq!(decode_once(eleven.trim_end_matches('=')), None);
        // "Mostly printable": 18 of 20 bytes (90 %) is read, 17 of 20 is not.
        let mut bytes = *b"abcdefghijklmnopqrst";
        bytes[0] = 0x01;
        bytes[1] = 0xff;
        assert!(decode_once(&base64(&bytes)).is_some());
        bytes[2] = 0x02;
        assert_eq!(decode_once(&base64(&bytes)), None);
        // Ordinary words of the base64 alphabet do not read as text.
        for text in [
            "CARGO_TARGET_DIR",
            "ProviderResponseValidationError",
            "src/decide/outbound",
            "the_largest_v2_request_fits_the_body_bound",
        ] {
            assert_eq!(decode_once(text), None, "{text}");
        }
    }

    #[test]
    fn a_hex_run_of_thirty_two_or_more_that_reads_as_text_is_decoded() {
        assert_eq!(HEX_MIN, 32);
        let key = fake_key();
        assert!(decode_once(&hex(key.as_bytes())).unwrap().contains(&key));
        // 16 bytes are 32 hex characters (decoded); 31 characters are not read.
        let sixteen = hex(b"0123456789abcdef");
        assert_eq!(sixteen.len(), 32);
        assert!(decode_once(&sixteen).is_some());
        assert_eq!(decode_once(&sixteen[..31]), None);
        // One byte that is not printable: not text, not read.
        let mut bytes = *b"0123456789abcdef";
        bytes[3] = 0x01;
        assert_eq!(decode_once(&hex(&bytes)), None);
        // A commit id and a digest are random bytes, not text.
        let commit = "c7ef9b1f8c2365fe08963269af718f3d48de7d51";
        assert_eq!(decode_once(commit), None);
        assert_eq!(decode_once(&format!("sha256:{}", "0123456789abcdef".repeat(4))), None);
    }
}
