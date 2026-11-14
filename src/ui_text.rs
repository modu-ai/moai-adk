//! Human-output text in English and Korean (CLI design S6a, Issue #2825; design
//! `docs/design-review/cli-s6-korean-2026-11-14.md`, leader decisions S6-1 to S6-5,
//! 6870716572).
//!
//! - [`Lang`] is the language of one human stream. It is chosen once by [`choose`] in this
//!   order (design 2.1): `MOAI_LANG` (`en` or `ko`; any other value is skipped), the project
//!   profile's `language.user_chat` (S6-1; a profile code other than `ko` means English),
//!   the first non-empty of `LC_ALL`, `LC_MESSAGES`, `LANG` starting with `ko`, else English.
//!   A stream that is not a terminal is always English (S6-2): lanes, tests and scripts read
//!   English bytes. [`crate::ui::ui_lang`] applies that to a real stream.
//! - [`Msg`] names one sentence, and [`text`] holds its English and Korean side by side in one
//!   `match`, so a missing translation is a compile error. A sentence may carry `{name}`
//!   placeholders; [`fill`] puts data into them through [`crate::ui::visible`], so a file name
//!   can never bring a control character to the terminal (leader decision 6867586011).
//! - The stdout JSON envelope, `--version`, hook stderr and the status line never come here
//!   (D3 6865936070): they stay English bytes.
//!
//! S6a adds only the mechanism and the first sentences; the screens move to [`text`] in S6b–S6d.

use std::path::Path;

/// The language of one human stream.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Lang {
    En,
    Ko,
}

impl Lang {
    /// `en` or `ko`; anything else is not a choice.
    pub fn from_code(code: &str) -> Option<Lang> {
        match code {
            "en" => Some(Lang::En),
            "ko" => Some(Lang::Ko),
            _ => None,
        }
    }
}

/// The inputs of the choice, read once so tests can pass their own.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Inputs {
    /// Whether the stream is a terminal.
    pub terminal: bool,
    pub moai_lang: Option<String>,
    /// The profile's `language.user_chat`, when a project profile was read.
    pub profile_chat: Option<String>,
    pub lc_all: Option<String>,
    pub lc_messages: Option<String>,
    pub lang: Option<String>,
}

/// The language for these inputs (design 2.1, S6-2).
pub fn choose(i: &Inputs) -> Lang {
    if !i.terminal {
        return Lang::En;
    }
    if let Some(l) = i.moai_lang.as_deref().and_then(Lang::from_code) {
        return l;
    }
    if let Some(code) = i.profile_chat.as_deref() {
        // The profile made a choice: Korean when it is Korean, English for every other code
        // (ja, zh, …), without falling through to the locale.
        return if code == "ko" { Lang::Ko } else { Lang::En };
    }
    let locale = [&i.lc_all, &i.lc_messages, &i.lang]
        .into_iter()
        .flatten()
        .find(|v| !v.is_empty());
    match locale {
        Some(v) if v.starts_with("ko") => Lang::Ko,
        _ => Lang::En,
    }
}

/// `language.user_chat` of the project that holds `cwd`: walking up from `cwd`, the first
/// folder whose profile reads; the walk stops at the first folder holding `.git` (the project
/// boundary), and HOME and the file-system root are never read as projects. Every read
/// failure is silent — the language of a screen must never make a command fail.
pub fn profile_chat(cwd: &Path, home: Option<&Path>) -> Option<String> {
    for dir in cwd.ancestors() {
        if dir.parent().is_none() || Some(dir) == home {
            return None;
        }
        if let Ok((profile, _)) = crate::profile::read_profile(dir) {
            return profile.language.and_then(|l| l.get("user_chat").cloned());
        }
        if dir.join(".git").exists() {
            return None;
        }
    }
    None
}

/// One human sentence. Every variant has an English and a Korean text in [`text`].
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Msg {
    /// The word of the `ok` token (S2 `Token::word`).
    WordOk,
    /// The word of the `warning` token.
    WordWarning,
    /// The word of the `error` token.
    WordError,
    /// The word of the `next` token.
    WordNext,
    /// A plan that changes nothing (S4 summary).
    WritesNothing,
    /// An apply that finished (S4 summary).
    Done,
    /// Rows left out of a long summary table; `{count}` is a number.
    MoreRows,
    /// The note while a wait goes on; `{what}` names what is awaited (S5 progress).
    StillWaiting,
}

impl Msg {
    /// Every variant, for the tests that walk the table.
    pub const ALL: [Msg; 8] = [
        Msg::WordOk,
        Msg::WordWarning,
        Msg::WordError,
        Msg::WordNext,
        Msg::WritesNothing,
        Msg::Done,
        Msg::MoreRows,
        Msg::StillWaiting,
    ];
}

/// The text of `msg` in `lang`. The English texts are today's strings, byte for byte.
pub fn text(msg: Msg, lang: Lang) -> &'static str {
    use Lang::{En, Ko};
    use Msg::*;
    match (msg, lang) {
        (WordOk, En) => "ok",
        (WordOk, Ko) => "정상",
        (WordWarning, En) => "warning",
        (WordWarning, Ko) => "경고",
        (WordError, En) => "error",
        (WordError, Ko) => "오류",
        (WordNext, En) => "next",
        (WordNext, Ko) => "다음",
        (WritesNothing, En) => "writes nothing",
        (WritesNothing, Ko) => "쓰는 것 없음",
        (Done, En) => "done",
        (Done, Ko) => "완료",
        (MoreRows, En) => "… {count} more",
        (MoreRows, Ko) => "… 외 {count}개",
        (StillWaiting, En) => "still waiting: {what}",
        (StillWaiting, Ko) => "아직 기다리는 중: {what}",
    }
}

/// The `{name}` placeholders of a text, in order of appearance.
pub fn placeholders(text: &str) -> Vec<&str> {
    let mut out = Vec::new();
    let mut rest = text;
    while let Some(open) = rest.find('{') {
        let after = &rest[open + 1..];
        let Some(close) = after.find('}') else { break };
        out.push(&after[..close]);
        rest = &after[close + 1..];
    }
    out
}

/// `text(msg, lang)` with each `{name}` replaced by its value, made safe for a terminal by
/// [`crate::ui::visible`]. A placeholder without a value is left as written.
pub fn fill(msg: Msg, lang: Lang, values: &[(&str, &str)]) -> String {
    let mut out = text(msg, lang).to_string();
    for (name, value) in values {
        out = out.replace(&format!("{{{name}}}"), &crate::ui::visible(value));
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    fn some(s: &str) -> Option<String> {
        Some(s.to_string())
    }

    fn tty() -> Inputs {
        Inputs {
            terminal: true,
            ..Inputs::default()
        }
    }

    /// Design 2.1 as a table: each row is the inputs and the language they give.
    #[test]
    fn the_choice_follows_the_documented_order() {
        let rows: Vec<(Inputs, Lang)> = vec![
            (tty(), Lang::En),
            // MOAI_LANG beats the profile and the locale.
            (
                Inputs {
                    moai_lang: some("ko"),
                    profile_chat: some("en"),
                    lang: some("en_US.UTF-8"),
                    ..tty()
                },
                Lang::Ko,
            ),
            (
                Inputs {
                    moai_lang: some("en"),
                    profile_chat: some("ko"),
                    ..tty()
                },
                Lang::En,
            ),
            // An unknown MOAI_LANG is skipped, not taken as English.
            (
                Inputs {
                    moai_lang: some("fr"),
                    profile_chat: some("ko"),
                    ..tty()
                },
                Lang::Ko,
            ),
            // The profile beats the locale; a non-Korean profile code is English.
            (
                Inputs {
                    profile_chat: some("ja"),
                    lang: some("ko_KR.UTF-8"),
                    ..tty()
                },
                Lang::En,
            ),
            // Locale: the first non-empty of LC_ALL, LC_MESSAGES, LANG.
            (
                Inputs {
                    lc_all: some(""),
                    lc_messages: some("ko_KR.UTF-8"),
                    lang: some("en_US.UTF-8"),
                    ..tty()
                },
                Lang::Ko,
            ),
            (
                Inputs {
                    lc_all: some("C"),
                    lang: some("ko_KR.UTF-8"),
                    ..tty()
                },
                Lang::En,
            ),
            (
                Inputs {
                    lang: some("ko"),
                    ..tty()
                },
                Lang::Ko,
            ),
        ];
        for (inputs, want) in rows {
            assert_eq!(choose(&inputs), want, "{inputs:?}");
            // S6-2: the same inputs on a pipe are always English.
            let piped = Inputs {
                terminal: false,
                ..inputs.clone()
            };
            assert_eq!(choose(&piped), Lang::En, "{piped:?}");
        }
    }

    /// Every sentence has a non-empty Korean text with the same placeholders as the English
    /// one, and no text constant carries a control character.
    #[test]
    fn every_sentence_has_a_matching_korean_text() {
        for m in Msg::ALL {
            let (en, ko) = (text(m, Lang::En), text(m, Lang::Ko));
            assert!(!en.is_empty() && !ko.is_empty(), "{m:?}");
            assert_ne!(en, ko, "{m:?} is not translated");
            let (mut pe, mut pk) = (placeholders(en), placeholders(ko));
            pe.sort_unstable();
            pk.sort_unstable();
            assert_eq!(pe, pk, "{m:?} placeholders");
            for t in [en, ko] {
                assert!(!t.chars().any(char::is_control), "{m:?}: {t:?}");
            }
        }
    }

    /// The English texts are today's strings, so the English screens keep their bytes.
    #[test]
    fn english_words_match_the_current_tokens() {
        use crate::ui::Token;
        for (m, t) in [
            (Msg::WordOk, Token::Ok),
            (Msg::WordWarning, Token::Warn),
            (Msg::WordError, Token::Error),
            (Msg::WordNext, Token::Info),
        ] {
            assert_eq!(text(m, Lang::En), t.word(), "{m:?}");
        }
    }

    /// Data put into a sentence goes through `ui::visible`: no control byte reaches the screen.
    #[test]
    fn filled_values_are_made_visible() {
        let s = fill(
            Msg::StillWaiting,
            Lang::Ko,
            &[("what", "ci \u{1b}]0;x\u{7}\u{9b}31m")],
        );
        assert!(!s.chars().any(char::is_control), "{s:?}");
        assert!(
            s.starts_with("아직 기다리는 중: ci \\x1b]0;x\\x07\\x9b31m"),
            "{s}"
        );
        assert_eq!(fill(Msg::MoreRows, Lang::En, &[("count", "3")]), "… 3 more");
        assert_eq!(fill(Msg::MoreRows, Lang::Ko, &[]), "… 외 {count}개");
    }

    #[test]
    fn placeholders_are_read_in_order() {
        assert_eq!(placeholders("a {x} b {y}"), ["x", "y"]);
        assert!(placeholders("no braces").is_empty());
        assert_eq!(placeholders("open { only"), Vec::<&str>::new());
    }

    /// The profile is found from a subfolder of the project, stops at `.git`, skips HOME, and
    /// a missing or unreadable profile is silently nothing.
    #[test]
    fn the_profile_is_read_from_the_project_that_holds_the_folder() {
        let d = tempfile::TempDir::new().unwrap();
        let base = fs::canonicalize(d.path()).unwrap();
        let project = base.join("proj");
        fs::create_dir(&project).unwrap();
        crate::launcher_init::prepare(&project, "claude", false).unwrap();
        let file = crate::state_root::path(&project, "project.yaml");
        let mut v: serde_json::Value = serde_json::from_slice(&fs::read(&file).unwrap()).unwrap();
        v["language"] = serde_json::json!({"user_chat": "ko"});
        fs::write(&file, serde_json::to_vec(&v).unwrap()).unwrap();

        let sub = project.join("src/deep");
        fs::create_dir_all(&sub).unwrap();
        assert_eq!(profile_chat(&sub, None).as_deref(), Some("ko"));
        assert_eq!(profile_chat(&project, None).as_deref(), Some("ko"));
        // HOME is never read as a project, even when it holds one.
        assert_eq!(profile_chat(&project, Some(&project)), None);

        // A nested repository without a profile is its own project: the walk stops there.
        let nested = project.join("vendor/other");
        fs::create_dir_all(nested.join(".git")).unwrap();
        assert_eq!(profile_chat(&nested, None), None);

        // An unreadable profile is nothing, not an error.
        fs::write(&file, b"not json").unwrap();
        assert_eq!(profile_chat(&sub, None), None);
    }
}
