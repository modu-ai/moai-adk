//! The imperative reply-language sentence (SPEC-20261120-reply-language-rule).
//!
//! The AGENTS.md managed section used to list which language each kind of text is written in
//! and never said to answer in the conversation language, so a model could answer in English
//! in a Korean project. This module builds the sentence that says it, in the instruction
//! language (`en` or `ko`), and the one-line form the Codex block carries. It only builds text;
//! the callers decide where it goes.

use crate::language::{english_name, is_code, CODES};

/// The language's name in the sentence's own language; `None` for a code MoAI has no name for.
fn name(code: &str, ui: &str) -> Option<&'static str> {
    if !is_code(code) {
        return None;
    }
    CODES
        .iter()
        .find(|(c, _, _)| *c == code)
        .map(|(_, en, ko)| if ui == "ko" { *ko } else { *en })
}

/// The sentence that opens the writing-language part of the managed section: replies, reports
/// and questions in the conversation language, commit messages and documents in theirs,
/// everything else in English. `None` when a code has no name, so a caller never writes a
/// sentence with a hole in it. `ui` is the instruction language: `ko` gives Korean, any other
/// value English.
pub fn sentence(ui: &str, conversation: &str, commit: &str, documents: &str) -> Option<String> {
    let c = name(conversation, ui)?;
    let m = name(commit, ui)?;
    let d = name(documents, ui)?;
    Some(if ui == "ko" {
        // Every Korean language name ends in a vowel, so the particle is always `로`.
        format!(
            "사용자에게 답하고 보고서를 쓰고 질문할 때는 {c}로, 커밋 메시지는 {m}로, 문서는 {d}로 쓰고, 그 밖의 모든 글은 영어로 써요."
        )
    } else {
        format!(
            "Reply to the user, write reports and ask questions in {c}; write commit messages in {m}; write documents in {d}; write everything else in English."
        )
    })
}

/// The one line the Codex block carries: reply in the saved conversation language. `None` for a
/// code MoAI has no name for, and for English, which needs no instruction.
pub fn codex_line(conversation: &str) -> Option<String> {
    if conversation == "en" {
        return None;
    }
    let en = english_name(conversation)?;
    Some(format!(
        "Reply to the user, write reports and ask questions in {en}."
    ))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_english_sentence_names_every_language_and_the_english_rest() {
        let s = sentence("en", "ko", "en", "ko").unwrap();
        assert_eq!(
            s,
            "Reply to the user, write reports and ask questions in Korean; write commit messages in English; write documents in Korean; write everything else in English."
        );
    }

    #[test]
    fn the_korean_sentence_is_korean_and_keeps_the_same_meaning() {
        let s = sentence("ko", "ko", "en", "ja").unwrap();
        assert_eq!(
            s,
            "사용자에게 답하고 보고서를 쓰고 질문할 때는 한국어로, 커밋 메시지는 영어로, 문서는 일본어로 쓰고, 그 밖의 모든 글은 영어로 써요."
        );
        assert!(s.contains("영어"), "{s}");
    }

    #[test]
    fn any_other_instruction_language_gives_the_english_sentence() {
        assert_eq!(
            sentence("fr", "ko", "ko", "ko"),
            sentence("en", "ko", "ko", "ko")
        );
    }

    #[test]
    fn a_code_without_a_name_gives_no_sentence() {
        assert!(sentence("en", "fr", "en", "en").is_none());
        assert!(sentence("en", "ko", "xx", "en").is_none());
        assert!(sentence("ko", "ko", "en", "").is_none());
    }

    #[test]
    fn every_known_code_has_a_name_in_both_languages() {
        for (code, _, _) in CODES {
            for ui in ["en", "ko"] {
                assert!(sentence(ui, code, code, code).is_some(), "{code} {ui}");
            }
        }
    }

    #[test]
    fn the_codex_line_names_the_language_and_skips_english_and_unknown_codes() {
        assert_eq!(
            codex_line("ko").unwrap(),
            "Reply to the user, write reports and ask questions in Korean."
        );
        assert!(codex_line("en").is_none());
        assert!(codex_line("fr").is_none());
    }
}
