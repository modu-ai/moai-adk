//! Human-message language (Issue #2825 S6a; design PR #2826 sections 2–3, leader decision
//! 6870716572): which language a person-facing line uses, and the in-code table of those lines.
//!
//! - **Which language**: `MOAI_LANG` (`en`/`ko`; other values are skipped) → the project
//!   profile's `language.user_chat` → `LC_ALL` → `LC_MESSAGES` → `LANG` (empty values skipped,
//!   the first two letters `ko`) → English. A profile asking for another allowed code (`ja`,
//!   `zh`) gets English: the catalog holds `en` and `ko` only (S6-1, S6-3).
//! - **Only on a terminal**: a stream that is not a terminal always gets English, so pipes,
//!   files and every test that greps the English lines see the same bytes (S6-2).
//! - **The table**: [`Msg`] lists every translated line and [`text`] answers it per language in
//!   one `match`; a message without a Korean line does not compile. A line may hold `{name}`
//!   placeholders; the value put there is data and reaches the terminal through
//!   `ui::Painter`, which makes control characters visible (#2818, decision 6867586011).
//!
//! Nothing here touches the stdout JSON envelope, `--version`, hook stderr or the status line.

use std::ffi::OsString;
use std::path::Path;

/// The two catalog languages.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Lang {
    En,
    Ko,
}

impl Lang {
    /// `en` → English, `ko` → Korean; any other code (an allowed `ja`/`zh` or garbage) has no
    /// catalog and is `None`.
    pub fn from_code(code: &str) -> Option<Lang> {
        match code.trim() {
            "en" => Some(Lang::En),
            "ko" => Some(Lang::Ko),
            _ => None,
        }
    }
}

/// The environment the choice reads, so the order can be tested without touching the process
/// environment.
#[derive(Clone, Debug, Default)]
pub struct LangEnv {
    pub moai_lang: Option<OsString>,
    pub lc_all: Option<OsString>,
    pub lc_messages: Option<OsString>,
    pub lang: Option<OsString>,
}

impl LangEnv {
    /// The current process environment.
    pub fn from_process() -> LangEnv {
        LangEnv {
            moai_lang: std::env::var_os("MOAI_LANG"),
            lc_all: std::env::var_os("LC_ALL"),
            lc_messages: std::env::var_os("LC_MESSAGES"),
            lang: std::env::var_os("LANG"),
        }
    }
}

/// The language a stream gets. `terminal` is whether the stream is a terminal; `profile_chat`
/// is the project profile's `language.user_chat`, if one was read.
pub fn choose(terminal: bool, env: &LangEnv, profile_chat: Option<&str>) -> Lang {
    if !terminal {
        return Lang::En;
    }
    if let Some(lang) = env
        .moai_lang
        .as_ref()
        .and_then(|v| v.to_str())
        .and_then(Lang::from_code)
    {
        return lang;
    }
    if let Some(code) = profile_chat {
        // An allowed code without a catalog (`ja`, `zh`) is English, not the locale's choice:
        // the profile is the person's explicit pick for this project.
        return Lang::from_code(code).unwrap_or(Lang::En);
    }
    // POSIX message category order: the first non-empty of LC_ALL, LC_MESSAGES, LANG decides.
    let locale = [&env.lc_all, &env.lc_messages, &env.lang]
        .into_iter()
        .flatten()
        .find(|v| !v.is_empty());
    match locale.and_then(|v| v.to_str()) {
        Some(v) if v.starts_with("ko") => Lang::Ko,
        _ => Lang::En,
    }
}

/// Folders above the start that the profile search looks at, at most.
const PROFILE_SEARCH_DEPTH: usize = 40;

/// The project profile's `language.user_chat`, searched from `start` up to the first folder
/// holding `.git` (that folder included). Any read failure is quiet and gives `None`: a
/// person's screen language never makes a command fail.
pub fn profile_chat_from(start: &Path) -> Option<String> {
    let mut dir = Some(start);
    for _ in 0..PROFILE_SEARCH_DEPTH {
        let here = dir?;
        if let Ok((profile, _)) = crate::profile::read_profile(here) {
            return profile
                .language
                .as_ref()
                .and_then(|l| l.get("user_chat"))
                .cloned();
        }
        if std::fs::symlink_metadata(here.join(".git")).is_ok() {
            return None;
        }
        dir = here.parent();
    }
    None
}

/// The language for standard error, from the process environment and the profile found from
/// the current folder. The profile is read only when the environment does not already decide.
pub fn stderr_lang() -> Lang {
    use std::io::IsTerminal;
    stream_lang(std::io::stderr().is_terminal())
}

/// As [`stderr_lang`], for a stream whose terminal state the caller knows.
pub fn stream_lang(terminal: bool) -> Lang {
    if !terminal {
        return Lang::En;
    }
    let env = LangEnv::from_process();
    let explicit = env
        .moai_lang
        .as_ref()
        .and_then(|v| v.to_str())
        .and_then(Lang::from_code);
    if explicit.is_some() {
        return choose(true, &env, None);
    }
    let chat = std::env::current_dir()
        .ok()
        .and_then(|d| profile_chat_from(&d));
    choose(true, &env, chat.as_deref())
}

/// Every person-facing line that has a translation. S6a seeds the table with lines S6b moves
/// here; each English line is the exact text the code prints today.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Msg {
    /// `ui::Token::Ok` word.
    WordOk,
    /// `ui::Token::Warn` word.
    WordWarning,
    /// `ui::Token::Error` word.
    WordError,
    /// `ui::Token::Info` word (the next-step label).
    WordNext,
    /// A polling command that stays in one state (`progress::Watch`).
    StillWaiting,
    /// The `key-remove` question on a terminal.
    ConfirmRemove,
    /// What `key-remove` says after a "no".
    RunWithYes,
}

impl Msg {
    /// Every message, for tests that walk the table.
    pub const ALL: [Msg; 7] = [
        Msg::WordOk,
        Msg::WordWarning,
        Msg::WordError,
        Msg::WordNext,
        Msg::StillWaiting,
        Msg::ConfirmRemove,
        Msg::RunWithYes,
    ];
}

/// The line for `msg` in `lang`. `{name}` marks a value the caller puts in.
pub fn text(msg: Msg, lang: Lang) -> &'static str {
    use Lang::{En, Ko};
    use Msg::*;
    match (msg, lang) {
        (WordOk, En) => "ok",
        (WordOk, Ko) => "완료",
        (WordWarning, En) => "warning",
        (WordWarning, Ko) => "주의",
        (WordError, En) => "error",
        (WordError, Ko) => "오류",
        (WordNext, En) => "next",
        (WordNext, Ko) => "다음",
        (StillWaiting, En) => "still waiting: {state}",
        (StillWaiting, Ko) => "아직 기다리는 중: {state}",
        (ConfirmRemove, En) => "Remove it?",
        (ConfirmRemove, Ko) => "지울까요?",
        (RunWithYes, En) => "run with --yes to remove it without asking",
        (RunWithYes, Ko) => "묻지 않고 지우려면 --yes 를 붙여 실행하세요",
    }
}

/// The `{name}` placeholders of a line, in order.
pub fn placeholders(line: &str) -> Vec<&str> {
    let mut out = Vec::new();
    let mut rest = line;
    while let Some(open) = rest.find('{') {
        let after = &rest[open + 1..];
        let Some(close) = after.find('}') else { break };
        out.push(&after[..close]);
        rest = &after[close + 1..];
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    fn env(pairs: &[(&str, &str)]) -> LangEnv {
        let mut e = LangEnv::default();
        for (k, v) in pairs {
            let v = Some(OsString::from(v));
            match *k {
                "MOAI_LANG" => e.moai_lang = v,
                "LC_ALL" => e.lc_all = v,
                "LC_MESSAGES" => e.lc_messages = v,
                "LANG" => e.lang = v,
                other => panic!("unknown key {other}"),
            }
        }
        e
    }

    /// Design 2.1 as a table: (terminal, environment, profile user_chat) → language.
    #[test]
    fn the_language_follows_the_documented_order() {
        use Lang::{En, Ko};
        let rows: &[(bool, &[(&str, &str)], Option<&str>, Lang, &str)] = &[
            (true, &[], None, En, "nothing set"),
            (
                false,
                &[("MOAI_LANG", "ko")],
                Some("ko"),
                En,
                "a pipe is always English",
            ),
            (
                true,
                &[("MOAI_LANG", "ko")],
                Some("en"),
                Ko,
                "MOAI_LANG beats the profile",
            ),
            (
                true,
                &[("MOAI_LANG", "en"), ("LANG", "ko_KR.UTF-8")],
                Some("ko"),
                En,
                "MOAI_LANG=en",
            ),
            (
                true,
                &[("MOAI_LANG", "fr")],
                Some("ko"),
                Ko,
                "an unknown MOAI_LANG is skipped",
            ),
            (
                true,
                &[("MOAI_LANG", "")],
                None,
                En,
                "an empty MOAI_LANG is skipped",
            ),
            (
                true,
                &[("LANG", "ko_KR.UTF-8")],
                Some("en"),
                En,
                "the profile beats the locale",
            ),
            (
                true,
                &[("LANG", "ko_KR.UTF-8")],
                Some("ja"),
                En,
                "ja has no catalog",
            ),
            (
                true,
                &[("LANG", "ko_KR.UTF-8")],
                Some("zh"),
                En,
                "zh has no catalog",
            ),
            (true, &[("LANG", "ko_KR.UTF-8")], None, Ko, "LANG ko"),
            (
                true,
                &[("LC_ALL", ""), ("LANG", "ko_KR.UTF-8")],
                None,
                Ko,
                "an empty LC_ALL is skipped",
            ),
            (
                true,
                &[("LC_ALL", "C"), ("LANG", "ko_KR.UTF-8")],
                None,
                En,
                "LC_ALL decides first",
            ),
            (
                true,
                &[("LC_MESSAGES", "ko_KR"), ("LANG", "en_US")],
                None,
                Ko,
                "LC_MESSAGES before LANG",
            ),
            (
                true,
                &[("LC_ALL", "ko"), ("LC_MESSAGES", "en_US")],
                None,
                Ko,
                "LC_ALL before LC_MESSAGES",
            ),
            (true, &[("LANG", "kok_IN")], None, Ko, "two letters only"),
        ];
        for (terminal, pairs, chat, want, why) in rows {
            assert_eq!(choose(*terminal, &env(pairs), *chat), *want, "{why}");
        }
    }

    /// Every message has a non-empty line in both languages, the same placeholders, and no
    /// control character (the constants are safe; only inserted values need the Painter).
    #[test]
    fn every_message_has_both_lines_with_the_same_placeholders() {
        for m in Msg::ALL {
            let en = text(m, Lang::En);
            let ko = text(m, Lang::Ko);
            assert!(!en.trim().is_empty() && !ko.trim().is_empty(), "{m:?}");
            let mut a = placeholders(en);
            let mut b = placeholders(ko);
            a.sort_unstable();
            b.sort_unstable();
            assert_eq!(a, b, "{m:?}: placeholders differ");
            assert!(
                !en.chars().chain(ko.chars()).any(char::is_control),
                "{m:?}: control character in a constant"
            );
            assert_ne!(
                en, ko,
                "{m:?}: the Korean line is a copy of the English one"
            );
        }
    }

    /// `ALL` names every variant once (a new variant added to the enum but not to `ALL` would
    /// escape the table test above).
    #[test]
    fn all_lists_every_message_once() {
        let mut seen = std::collections::BTreeSet::new();
        for m in Msg::ALL {
            assert!(seen.insert(format!("{m:?}")), "{m:?} twice");
            // The match below fails to compile when a variant is added and not listed here.
            match m {
                Msg::WordOk
                | Msg::WordWarning
                | Msg::WordError
                | Msg::WordNext
                | Msg::StillWaiting
                | Msg::ConfirmRemove
                | Msg::RunWithYes => {}
            }
        }
        assert_eq!(seen.len(), Msg::ALL.len());
    }

    /// The English lines are the exact text the code prints today (S6b swaps them in without
    /// changing English bytes).
    #[test]
    fn english_lines_are_todays_text() {
        assert_eq!(text(Msg::WordOk, Lang::En), crate::ui::Token::Ok.word());
        assert_eq!(
            text(Msg::WordWarning, Lang::En),
            crate::ui::Token::Warn.word()
        );
        assert_eq!(
            text(Msg::WordError, Lang::En),
            crate::ui::Token::Error.word()
        );
        assert_eq!(text(Msg::WordNext, Lang::En), crate::ui::Token::Info.word());
    }

    #[test]
    fn placeholders_are_read_in_order() {
        assert_eq!(placeholders("a {x} b {y}"), vec!["x", "y"]);
        assert_eq!(placeholders("none"), Vec::<&str>::new());
        assert_eq!(placeholders("open {x"), Vec::<&str>::new());
    }

    /// The profile search stops at the first `.git` and is quiet on a missing or broken
    /// profile.
    #[test]
    fn the_profile_search_is_bounded_and_quiet() {
        let root = tempfile::tempdir().unwrap();
        let repo = root.path().join("repo");
        let sub = repo.join("a/b");
        std::fs::create_dir_all(&sub).unwrap();
        std::fs::create_dir(repo.join(".git")).unwrap();
        assert_eq!(profile_chat_from(&sub), None, "no profile");
        // A broken profile file is quiet too.
        let profile = repo.join(".moai-cli/project.yaml");
        std::fs::create_dir_all(profile.parent().unwrap()).unwrap();
        std::fs::write(&profile, "{not json").unwrap();
        assert_eq!(profile_chat_from(&sub), None, "broken profile");
        // A valid profile with `language.user_chat` is found from a subfolder.
        std::fs::write(
            &profile,
            r#"{"schema_version":2,"project_id":"demo","purpose":"Build a small editor","tools":["Rust"],"development_methods":["TDD"],"hosts":["claude"],"interview_revision":"rc1-fixture-1","language":{"user_chat":"ko"}}"#,
        )
        .unwrap();
        assert_eq!(profile_chat_from(&sub).as_deref(), Some("ko"));
        // Above the first `.git` nothing is read.
        let outer = root.path().join("outer");
        std::fs::create_dir_all(outer.join("inner/.git")).unwrap();
        std::fs::create_dir_all(outer.join(".moai-cli")).unwrap();
        std::fs::copy(&profile, outer.join(".moai-cli/project.yaml")).unwrap();
        assert_eq!(
            profile_chat_from(&outer.join("inner")),
            None,
            "stops at .git"
        );
    }
}
