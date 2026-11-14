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
//! S6a added the mechanism and the first sentences. S6b moves the S2 token words, the S4
//! summary (heads, counts, `next:`), the S5 progress lines and the `key-remove`/`update`
//! confirm screens here; each screen takes its language from its stream's
//! [`crate::ui::Painter`]. Data inside a sentence (paths, action words from the result, the
//! `update` summary body) stays as the envelope has it. S6c–S6e move the error frame and help.

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

/// The language of one stream from its real inputs (what [`crate::ui::ui_lang`] passes): the
/// stream's terminal flag, an environment reader, the current folder and HOME. The terminal
/// flag goes into [`choose`], so the tested rule — a pipe is English (S6-2) — is the one that
/// runs; for a pipe nothing else is read (audit #2862 B1).
pub fn for_stream(
    terminal: bool,
    var: &dyn Fn(&str) -> Option<String>,
    cwd: Option<&Path>,
    home: Option<&Path>,
) -> Lang {
    let read = |f: &dyn Fn() -> Option<String>| if terminal { f() } else { None };
    choose(&Inputs {
        terminal,
        moai_lang: read(&|| var("MOAI_LANG")),
        profile_chat: read(&|| cwd.and_then(|c| profile_chat(c, home))),
        lc_all: read(&|| var("LC_ALL")),
        lc_messages: read(&|| var("LC_MESSAGES")),
        lang: read(&|| var("LANG")),
    })
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
    /// Rows left out of a long summary table; `{count}` is a number, `{ellipsis}` the
    /// painter's `…` or `...`.
    MoreRows,
    /// The note while a wait goes on; `{what}` names what is awaited (S5 progress).
    StillWaiting,
    /// S4 summary table heads (S6b): the action column, the note column and the name column
    /// under each name key (`path`, `host`, `target`, `selection`, anything else).
    HeadAction,
    HeadNote,
    HeadPath,
    HeadHost,
    HeadTarget,
    HeadSelection,
    HeadName,
    /// A note cell that holds the first of several reasons; `{first}` is data, `{count}` a
    /// number.
    MoreReasons,
    /// Stands for the arguments of the preview in the `next:` line of a summary.
    SameArguments,
    /// `index-build` stages (S5 progress).
    ScanningTrackedFiles,
    BuildingViews,
    PublishingGeneration,
    /// `gate-loop` progress; `{pr}`, `{at}` (1-based) and `{total}` are numbers.
    CheckingPullRequest,
    /// Paging through an API list; `{n}` is a number, `{path}` the API path.
    ReadingPage,
    /// `forge wait` states; `{head}` is a short commit id, the rest numbers.
    CiNoRunYet,
    CiRunsCompleted,
    PullRequestStillAt,
    NoNewComment,
    /// `key-remove` confirm preview: the first line, then `{provider}`, `{scope}`, `{path}`.
    KeyRemoveDeletes,
    KeyRemoveWhich,
    /// `key-remove` confirm question.
    RemoveIt,
    /// After a refused `key-remove` question: what to run instead.
    RemoveWithYes,
    /// `update` confirm preview title and question.
    UpdatePreview,
    ApplyUpdate,
    /// S6c: the error frame's `cause:` label (the `error:` and `next:` labels are token words).
    LabelCause,
    /// S6c: a usage error's next step for one command; `{program}` and `{word}` as typed.
    NextUsageWord,
    /// S6c: a usage error's next step without a command word; `{program}` as typed.
    NextUsageCommands,
    /// S6c: the next step of a missing permission reference.
    NextPermission,
    /// S6c: the next step of an unsupported protocol version.
    NextProtocol,
    /// S6c: the next step of a stale plan.
    NextStalePlan,
    /// S6c: the next step of an input over its size limit.
    NextInputTooLarge,
    /// S6c: the first line of a usage error on a Korean terminal, by kind (design 5). The
    /// English texts are clap's own words, which an English terminal keeps showing as clap
    /// renders them; the argument a user typed is never put in.
    UsageMissingArguments,
    UsageUnknownArgument,
    UsageInvalidValue,
    UsageOther,
    /// `install-inspect`'s terminal note when the first `moai` on PATH is not this generation's
    /// (#2867; S6 follow-up of #2825). The English text is `install::PATH_FIRST_NOTE`, which the
    /// envelope keeps in every language.
    PathFirstNote,
}

impl Msg {
    /// Every variant, for the tests that walk the table.
    pub const ALL: [Msg; 44] = [
        Msg::WordOk,
        Msg::WordWarning,
        Msg::WordError,
        Msg::WordNext,
        Msg::WritesNothing,
        Msg::Done,
        Msg::MoreRows,
        Msg::StillWaiting,
        Msg::HeadAction,
        Msg::HeadNote,
        Msg::HeadPath,
        Msg::HeadHost,
        Msg::HeadTarget,
        Msg::HeadSelection,
        Msg::HeadName,
        Msg::MoreReasons,
        Msg::SameArguments,
        Msg::ScanningTrackedFiles,
        Msg::BuildingViews,
        Msg::PublishingGeneration,
        Msg::CheckingPullRequest,
        Msg::ReadingPage,
        Msg::CiNoRunYet,
        Msg::CiRunsCompleted,
        Msg::PullRequestStillAt,
        Msg::NoNewComment,
        Msg::KeyRemoveDeletes,
        Msg::KeyRemoveWhich,
        Msg::RemoveIt,
        Msg::RemoveWithYes,
        Msg::UpdatePreview,
        Msg::ApplyUpdate,
        Msg::LabelCause,
        Msg::NextUsageWord,
        Msg::NextUsageCommands,
        Msg::NextPermission,
        Msg::NextProtocol,
        Msg::NextStalePlan,
        Msg::NextInputTooLarge,
        Msg::UsageMissingArguments,
        Msg::UsageUnknownArgument,
        Msg::UsageInvalidValue,
        Msg::UsageOther,
        Msg::PathFirstNote,
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
        (MoreRows, En) => "{ellipsis} {count} more",
        (MoreRows, Ko) => "{ellipsis} 외 {count}개",
        (StillWaiting, En) => "still waiting: {what}",
        (StillWaiting, Ko) => "아직 기다리는 중: {what}",
        (HeadAction, En) => "ACTION",
        (HeadAction, Ko) => "동작",
        (HeadNote, En) => "NOTE",
        (HeadNote, Ko) => "메모",
        (HeadPath, En) => "PATH",
        (HeadPath, Ko) => "경로",
        (HeadHost, En) => "HOST",
        (HeadHost, Ko) => "호스트",
        (HeadTarget, En) => "TARGET",
        (HeadTarget, Ko) => "대상",
        (HeadSelection, En) => "SELECTION",
        (HeadSelection, Ko) => "선택",
        (HeadName, En) => "NAME",
        (HeadName, Ko) => "이름",
        (MoreReasons, En) => "{first} (+{count} more)",
        (MoreReasons, Ko) => "{first} (외 {count}개)",
        (SameArguments, En) => "<same arguments>",
        (SameArguments, Ko) => "<같은 인자>",
        (ScanningTrackedFiles, En) => "scanning tracked files (1/3)",
        (ScanningTrackedFiles, Ko) => "추적 중인 파일 훑는 중 (1/3)",
        (BuildingViews, En) => "building the views (2/3)",
        (BuildingViews, Ko) => "보기 만드는 중 (2/3)",
        (PublishingGeneration, En) => "publishing the generation (3/3)",
        (PublishingGeneration, Ko) => "새 세대 내보내는 중 (3/3)",
        (CheckingPullRequest, En) => "checking pull request #{pr} ({at}/{total})",
        (CheckingPullRequest, Ko) => "풀 리퀘스트 #{pr} 확인 중 ({at}/{total})",
        (ReadingPage, En) => "reading page {n} of {path}",
        (ReadingPage, Ko) => "{path} 의 {n}쪽 읽는 중",
        (CiNoRunYet, En) => "ci on {head}: no workflow run yet",
        (CiNoRunYet, Ko) => "{head} 의 ci: 아직 워크플로 실행 없음",
        (CiRunsCompleted, En) => "ci on {head}: {done} of {total} runs completed",
        (CiRunsCompleted, Ko) => "{head} 의 ci: 실행 {total}개 중 {done}개 끝남",
        (PullRequestStillAt, En) => "pull request #{pr} still at {head}",
        (PullRequestStillAt, Ko) => "풀 리퀘스트 #{pr} 의 head 가 아직 {head}",
        (NoNewComment, En) => "no new comment on #{issue}",
        (NoNewComment, Ko) => "#{issue} 에 새 댓글 없음",
        (KeyRemoveDeletes, En) => "key-remove deletes 1 file",
        (KeyRemoveDeletes, Ko) => "key-remove 가 파일 1개를 지워요",
        (KeyRemoveWhich, En) => "{provider} key, {scope} scope: {path}",
        (KeyRemoveWhich, Ko) => "{provider} 키, {scope} 범위: {path}",
        (RemoveIt, En) => "Remove it?",
        (RemoveIt, Ko) => "지울까요?",
        (RemoveWithYes, En) => "run with --yes to remove it without asking",
        (RemoveWithYes, Ko) => "묻지 않고 지우려면 --yes 를 붙여 실행하세요",
        (UpdatePreview, En) => "moai update preview:",
        (UpdatePreview, Ko) => "moai update 미리보기:",
        (ApplyUpdate, En) => "Apply this update?",
        (ApplyUpdate, Ko) => "이 업데이트를 적용할까요?",
        (LabelCause, En) => "cause",
        (LabelCause, Ko) => "원인",
        (NextUsageWord, En) => "run `{program} {word} --help` to see its arguments",
        (NextUsageWord, Ko) => "`{program} {word} --help` 로 인자를 확인하세요",
        (NextUsageCommands, En) => "run `{program} --help` to see the commands",
        (NextUsageCommands, Ko) => "`{program} --help` 로 명령 목록을 확인하세요",
        (NextPermission, En) => {
            "pass `--permission-reference <ref>`, the approval reference the host gives for this plan"
        }
        (NextPermission, Ko) => {
            "이 plan 에 호스트가 준 승인 참조를 `--permission-reference <ref>` 로 넘기세요"
        }
        (NextProtocol, En) => {
            "send version 1 (`protocol_version: 1`, or `schema_version: 1` in a request file)"
        }
        (NextProtocol, Ko) => {
            "버전 1 로 보내세요(`protocol_version: 1`, 요청 파일이면 `schema_version: 1`)"
        }
        (NextStalePlan, En) => "make the plan again and apply it with the new digest",
        (NextStalePlan, Ko) => "plan 을 다시 만들고 새 digest 로 적용하세요",
        (NextInputTooLarge, En) => "send a smaller input",
        (NextInputTooLarge, Ko) => "더 작은 입력을 보내세요",
        (UsageMissingArguments, En) => "the following required arguments were not provided:",
        (UsageMissingArguments, Ko) => "꼭 필요한 인자가 빠졌어요:",
        (UsageUnknownArgument, En) => "unexpected argument",
        (UsageUnknownArgument, Ko) => "알 수 없는 인자예요",
        (UsageInvalidValue, En) => "invalid argument value",
        (UsageInvalidValue, Ko) => "인자 값이 올바르지 않아요",
        (UsageOther, En) => "invalid command line",
        (UsageOther, Ko) => "명령줄이 올바르지 않아요",
        (PathFirstNote, En) => crate::install::PATH_FIRST_NOTE,
        (PathFirstNote, Ko) => {
            "R5 전까지 hook 은 moai-cli 로 만들어져요. hook 이 moai 를 쓰게 하려면 Rust moai 를 \
             PATH 맨 앞에 두세요(바꾼 것 없음)"
        }
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
    fill_text(text(msg, lang), values)
}

/// One pass over `template`: each `{name}` is replaced once, so a value that itself contains
/// `{other}` is never expanded again (audit #2862 N1).
fn fill_text(template: &str, values: &[(&str, &str)]) -> String {
    let mut out = String::with_capacity(template.len());
    let mut rest = template;
    while let Some(open) = rest.find('{') {
        out.push_str(&rest[..open]);
        let after = &rest[open + 1..];
        match after.find('}') {
            Some(close) => {
                let name = &after[..close];
                match values.iter().find(|(n, _)| *n == name) {
                    Some((_, value)) => out.push_str(&crate::ui::visible(value)),
                    None => {
                        out.push('{');
                        out.push_str(name);
                        out.push('}');
                    }
                }
                rest = &after[close + 1..];
            }
            None => {
                out.push_str(&rest[open..]);
                rest = "";
            }
        }
    }
    out.push_str(rest);
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
            // LC_ALL comes before LC_MESSAGES, and LC_MESSAGES before LANG (audit #2862 B1).
            (
                Inputs {
                    lc_all: some("C"),
                    lc_messages: some("ko_KR.UTF-8"),
                    ..tty()
                },
                Lang::En,
            ),
            (
                Inputs {
                    lc_all: some("ko_KR.UTF-8"),
                    lc_messages: some("C"),
                    ..tty()
                },
                Lang::Ko,
            ),
            (
                Inputs {
                    lc_messages: some("ko_KR.UTF-8"),
                    lang: some("C"),
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

    /// `ALL` lists each sentence once, so the table walks above see every one.
    #[test]
    fn all_lists_each_sentence_once() {
        for (i, a) in Msg::ALL.iter().enumerate() {
            assert!(!Msg::ALL[i + 1..].contains(a), "{a:?} twice");
        }
    }

    /// The English texts are today's strings, so the English screens keep their bytes; the
    /// token words now come from this table (S6b), so they are pinned here as literals.
    #[test]
    fn english_words_match_the_current_tokens() {
        use crate::ui::Token;
        for (m, t, word) in [
            (Msg::WordOk, Token::Ok, "ok"),
            (Msg::WordWarning, Token::Warn, "warning"),
            (Msg::WordError, Token::Error, "error"),
            (Msg::WordNext, Token::Info, "next"),
        ] {
            assert_eq!(text(m, Lang::En), word, "{m:?}");
            assert_eq!(t.word(), word, "{t:?}");
            assert_eq!(t.word_in(Lang::Ko), text(m, Lang::Ko), "{t:?}");
        }
        for t in [Token::Accent, Token::Muted, Token::Code, Token::Strong] {
            assert_eq!(t.word_in(Lang::Ko), "", "{t:?} has no word");
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
        assert_eq!(
            fill(
                Msg::MoreRows,
                Lang::En,
                &[("ellipsis", "…"), ("count", "3")]
            ),
            "… 3 more"
        );
        assert_eq!(
            fill(Msg::MoreRows, Lang::Ko, &[("ellipsis", "...")]),
            "... 외 {count}개"
        );
    }

    /// The function `ui::ui_lang` calls: a pipe is English whatever the environment says, and
    /// nothing is read for it; a terminal follows the environment (audit #2862 B1).
    #[test]
    fn a_real_stream_on_a_pipe_is_english() {
        let asked = std::cell::RefCell::new(Vec::new());
        let env = |k: &str| {
            asked.borrow_mut().push(k.to_string());
            match k {
                "MOAI_LANG" | "LANG" => Some("ko".to_string()),
                _ => None,
            }
        };
        let cwd = Path::new("/nonexistent/moai-ui-text");
        assert_eq!(for_stream(false, &env, Some(cwd), None), Lang::En);
        assert!(
            asked.borrow().is_empty(),
            "a pipe reads nothing: {:?}",
            asked.borrow()
        );
        assert_eq!(for_stream(true, &env, Some(cwd), None), Lang::Ko);
        assert!(asked.borrow().contains(&"MOAI_LANG".to_string()));
    }

    /// One pass: a value holding another placeholder's text is not expanded again, and an
    /// unmatched brace stays as written (audit #2862 N1).
    #[test]
    fn fill_replaces_each_placeholder_once() {
        assert_eq!(
            fill_text("{a} and {b}", &[("a", "{b}"), ("b", "B")]),
            "{b} and B"
        );
        assert_eq!(fill_text("x {missing} y", &[]), "x {missing} y");
        assert_eq!(fill_text("open { tail", &[("tail", "t")]), "open { tail");
        assert_eq!(fill_text("", &[]), "");
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
