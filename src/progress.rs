//! Static progress lines and the confirm prompt of the human layer (Issue #2806, design
//! #2799 D2 §5·§7, leader decisions D3 6865936070 and 6866110172).
//!
//! Progress: a long-running command calls [`note`] at each stage; nothing happens unless the
//! command installed a sink. [`install_stderr`] installs one only when stderr is a terminal,
//! and it prints one plain line per stage (`→ scanning tracked files (1/3)`): no spinner, no
//! carriage return, a new line when the stage changes. With stderr on a pipe, a hook or a
//! test harness, no byte is written; the stdout envelope is never touched.
//!
//! Confirm: [`ask`] shows a preview and a `[y/N]` question on stderr and reads one answer
//! line; the default is no. [`ask_terminal`] asks only when stdin and stderr are both
//! terminals and otherwise answers no without writing anything, so the caller keeps its
//! non-interactive refusal and its `next:` hint.
//!
//! Every text this module writes to the terminal goes through [`crate::ui::visible`] first
//! (leader decision 6867586011): a path or a state line carrying a control character can never
//! move the cursor or retitle the window. `MOAI_OUTPUT=quiet` installs no progress sink and
//! `MOAI_OUTPUT=plain` writes without color or UTF-8 symbols, as the S4 summary does; the
//! confirm question is still asked under `quiet`, because without it the answer is no.

use crate::ui::{ColorFlag, Painter, Stream, Token};
use crate::ui_summary::OutputMode;
use crate::ui_text::{Lang, Msg};
use std::cell::{Cell, RefCell};
use std::io::{self, BufRead, IsTerminal, Write};

type Sink = Box<dyn FnMut(&str)>;

thread_local! {
    static SINK: RefCell<Option<Sink>> = const { RefCell::new(None) };
    /// The language of the installed sink's lines (S6b); English without one.
    static LANG: Cell<Lang> = const { Cell::new(Lang::En) };
}

/// Reports one stage of a long-running command; a no-op unless a sink is installed.
pub fn note(text: &str) {
    SINK.with(|s| {
        if let Some(sink) = s.borrow_mut().as_mut() {
            sink(text);
        }
    });
}

/// [`note`] of one sentence in the sink's language, with `values` filled in (S6b). Nothing is
/// formatted when no sink listens.
pub fn note_msg(msg: Msg, values: &[(&str, &str)]) {
    if SINK.with(|s| s.borrow().is_some()) {
        note(&say(msg, values));
    }
}

/// `msg` in the sink's language with `values` filled in: the text a progress line would show.
pub fn say(msg: Msg, values: &[(&str, &str)]) -> String {
    crate::ui_text::fill(msg, LANG.with(Cell::get), values)
}

/// Installs `sink` for this thread's later [`note`] calls (tests pass a recorder), in English.
pub fn install(sink: Sink) {
    install_in(sink, Lang::En);
}

/// Installs `sink` with the language its sentences are written in.
pub fn install_in(sink: Sink, lang: Lang) {
    SINK.with(|s| *s.borrow_mut() = Some(sink));
    LANG.with(|l| l.set(lang));
}

/// Removes the sink; later [`note`] calls write nothing and sentences are English again.
pub fn uninstall() {
    SINK.with(|s| *s.borrow_mut() = None);
    LANG.with(|l| l.set(Lang::En));
}

/// One progress line: the info mark (`→`, or `->` without UTF-8) and the text, with any
/// control character shown as `\xNN`.
pub fn line(painter: &Painter, text: &str) -> String {
    format!(
        "{} {}\n",
        painter.mark(Token::Info),
        crate::ui::visible(text)
    )
}

/// `text` with each line made [`crate::ui::visible`]; the line breaks themselves stay, so a
/// multi-line preview keeps its shape while a control character inside a line is shown.
pub fn visible_lines(text: &str) -> String {
    text.split('\n')
        .map(|l| crate::ui::visible(l).into_owned())
        .collect::<Vec<_>>()
        .join("\n")
}

/// The stderr painter `MOAI_OUTPUT` asks for; `None` for progress under `quiet`. `plain` drops
/// color and UTF-8 symbols but keeps the normal painter's language (S6b).
pub fn progress_painter(mode: OutputMode, normal: impl FnOnce() -> Painter) -> Option<Painter> {
    match mode {
        OutputMode::Quiet => None,
        OutputMode::Plain => Some(Painter::plain().with_lang(normal().lang)),
        OutputMode::Normal => Some(normal()),
    }
}

fn output_mode() -> OutputMode {
    crate::ui_summary::output_mode(std::env::var_os("MOAI_OUTPUT").as_deref())
}

/// Installs a stderr sink when stderr is a terminal and `MOAI_OUTPUT` is not `quiet`;
/// otherwise leaves [`note`] silent.
pub fn install_stderr() {
    if !io::stderr().is_terminal() {
        return;
    }
    let Some(painter) = progress_painter(output_mode(), || {
        Painter::for_stream(Stream::Stderr, ColorFlag::Auto)
    }) else {
        return;
    };
    install_in(
        Box::new(move |text| {
            let mut err = io::stderr();
            let _ = err.write_all(line(&painter, text).as_bytes());
            let _ = err.flush();
        }),
        painter.lang,
    );
}

/// After this long in one unchanged state, a waiting command says it is still waiting
/// (leader decision 6866110172, Q3).
pub const STILL_WAITING_SECS: i64 = 300;

/// The lines a polling command prints: one when its state changes, and `still waiting: …`
/// once every [`STILL_WAITING_SECS`] the state stays the same. Times come from the caller's
/// clock, so tests drive it without waiting.
#[derive(Debug, Default)]
pub struct StateNotes {
    last: Option<String>,
    since: i64,
}

impl StateNotes {
    /// The line to print for `state` observed at `now` (seconds), if any.
    pub fn observe(&mut self, state: &str, now: i64) -> Option<String> {
        if self.last.as_deref() != Some(state) {
            self.last = Some(state.to_string());
            self.since = now;
            return Some(state.to_string());
        }
        if now - self.since >= STILL_WAITING_SECS {
            self.since = now;
            return Some(say(Msg::StillWaiting, &[("what", state)]));
        }
        None
    }
}

/// Shows `preview` and `question [y/N] ` on `err`, reads one line from `input`, and answers
/// whether it was a yes (`y`, `Y`, `yes`, `Yes`, `YES`). An empty line or end of input is no.
pub fn ask<R: BufRead, W: Write>(
    input: &mut R,
    err: &mut W,
    painter: &Painter,
    preview: &str,
    question: &str,
) -> io::Result<bool> {
    write!(
        err,
        "{}{} {} [y/N] ",
        visible_lines(preview),
        painter.mark(Token::Warn),
        crate::ui::visible(question)
    )?;
    err.flush()?;
    let mut answer = String::new();
    input.read_line(&mut answer)?;
    Ok(matches!(answer.trim(), "y" | "Y" | "yes" | "Yes" | "YES"))
}

/// [`ask`] on the process's stdin and stderr, only when both are terminals; otherwise no,
/// with nothing written. The preview is written by `preview` for the painter that asks, and the
/// question is one sentence, so both come in the stream's language (S6b); the answer is still
/// `y`/`yes` and the prompt `[y/N]` (leader decision S6-4).
pub fn ask_terminal(
    preview: impl FnOnce(&Painter) -> String,
    question: Msg,
) -> Result<bool, String> {
    if !(io::stdin().is_terminal() && io::stderr().is_terminal()) {
        return Ok(false);
    }
    // The question is asked under `quiet` too; `plain` only drops color and UTF-8 symbols.
    let painter = progress_painter(output_mode(), || {
        Painter::for_stream(Stream::Stderr, ColorFlag::Auto)
    })
    .unwrap_or_else(|| Painter::plain().with_lang(crate::ui::ui_lang(Stream::Stderr)));
    let preview = preview(&painter);
    ask(
        &mut io::stdin().lock(),
        &mut io::stderr(),
        &painter,
        &preview,
        painter.text(question),
    )
    .map_err(|_| "could not read the answer".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::rc::Rc;

    #[test]
    fn note_is_silent_without_a_sink_and_reaches_an_installed_one() {
        uninstall();
        note("nothing listens");
        let seen: Rc<RefCell<Vec<String>>> = Rc::default();
        let keep = seen.clone();
        install(Box::new(move |t| keep.borrow_mut().push(t.to_string())));
        note("scanning tracked files (1/3)");
        uninstall();
        note("after uninstall");
        assert_eq!(*seen.borrow(), ["scanning tracked files (1/3)"]);
    }

    #[test]
    fn a_progress_line_is_static_and_carries_a_symbol() {
        let plain = line(&Painter::plain(), "reading page 2 of /repos/o/r/issues");
        assert_eq!(plain, "-> reading page 2 of /repos/o/r/issues\n");
        let colored = line(
            &Painter {
                color: true,
                unicode: true,
                lang: crate::ui_text::Lang::En,
            },
            "x",
        );
        assert_eq!(colored, "\x1b[36m→\x1b[0m x\n");
        for l in [plain, colored] {
            assert!(!l.contains('\r'), "no carriage return: {l:?}");
            assert_eq!(l.matches('\n').count(), 1);
        }
    }

    #[test]
    fn state_notes_print_on_change_and_after_five_unchanged_minutes() {
        let mut n = StateNotes::default();
        assert_eq!(
            n.observe("0 of 3 runs completed", 1000).as_deref(),
            Some("0 of 3 runs completed")
        );
        assert_eq!(n.observe("0 of 3 runs completed", 1030), None);
        // The boundary: 299 s unchanged is quiet, 300 s prints once and restarts the count.
        assert_eq!(n.observe("0 of 3 runs completed", 1000 + 299), None);
        assert_eq!(
            n.observe("0 of 3 runs completed", 1000 + 300).as_deref(),
            Some("still waiting: 0 of 3 runs completed")
        );
        assert_eq!(n.observe("0 of 3 runs completed", 1000 + 599), None);
        assert_eq!(
            n.observe("0 of 3 runs completed", 1000 + 600).as_deref(),
            Some("still waiting: 0 of 3 runs completed")
        );
        // A change prints at once and restarts the five minutes.
        assert_eq!(
            n.observe("2 of 3 runs completed", 1000 + 610).as_deref(),
            Some("2 of 3 runs completed")
        );
        assert_eq!(n.observe("2 of 3 runs completed", 1000 + 909), None);
    }

    #[test]
    fn ask_says_yes_only_for_a_yes_and_defaults_to_no() {
        let painter = Painter::plain();
        for (answer, want) in [
            ("y\n", true),
            ("Y\n", true),
            ("yes\n", true),
            ("YES\n", true),
            ("n\n", false),
            ("\n", false),
            ("", false),
            ("yep\n", false),
            ("no\n", false),
        ] {
            let mut err = Vec::new();
            let got = ask(
                &mut answer.as_bytes(),
                &mut err,
                &painter,
                "preview\n",
                "Remove it?",
            )
            .unwrap();
            assert_eq!(got, want, "{answer:?}");
            assert_eq!(
                String::from_utf8(err).unwrap(),
                "preview\n! Remove it? [y/N] "
            );
        }
    }

    #[test]
    fn terminal_text_shows_control_characters_and_keeps_preview_lines() {
        // Leader decision 6867586011: no raw control byte reaches the terminal from here.
        let plain = Painter::plain();
        assert_eq!(
            line(&plain, "ci on a\u{1b}]0;x\u{7}b"),
            "-> ci on a\\x1b]0;x\\x07b\n"
        );
        let mut err = Vec::new();
        ask(
            &mut "n\n".as_bytes(),
            &mut err,
            &plain,
            "remove /tmp/a\u{1b}[2Jb\n  shared key\n",
            "Remove\u{9b}it?",
        )
        .unwrap();
        let shown = String::from_utf8(err).unwrap();
        assert_eq!(
            shown,
            "remove /tmp/a\\x1b[2Jb\n  shared key\n! Remove\\x9bit? [y/N] "
        );
        assert!(!shown.chars().any(|c| c.is_control() && c != '\n'));
    }

    #[test]
    fn moai_output_quiet_drops_progress_and_plain_drops_color() {
        let colored = Painter {
            color: true,
            unicode: true,
            lang: crate::ui_text::Lang::En,
        };
        assert_eq!(progress_painter(OutputMode::Quiet, || colored), None);
        assert_eq!(
            progress_painter(OutputMode::Plain, || colored),
            Some(Painter::plain())
        );
        assert_eq!(
            progress_painter(OutputMode::Normal, || colored),
            Some(colored)
        );
        // `plain` drops color and symbols, never the language (S6b).
        let korean = colored.with_lang(Lang::Ko);
        assert_eq!(
            progress_painter(OutputMode::Plain, || korean),
            Some(Painter::plain().with_lang(Lang::Ko))
        );
    }

    /// S6b: a sentence is written in the language of the sink that listens; nothing is said
    /// without a sink, and uninstalling goes back to English.
    #[test]
    fn notes_follow_the_language_of_the_installed_sink() {
        let seen: Rc<RefCell<Vec<String>>> = Rc::default();
        let keep = seen.clone();
        install_in(
            Box::new(move |t| keep.borrow_mut().push(t.to_string())),
            Lang::Ko,
        );
        note_msg(Msg::ScanningTrackedFiles, &[]);
        note_msg(
            Msg::CheckingPullRequest,
            &[("pr", "12"), ("seen", "1"), ("ready", "3")],
        );
        assert_eq!(
            say(Msg::CiNoRunYet, &[("head", "0123abcd")]),
            "0123abcd 의 ci: 아직 워크플로 실행 없음"
        );
        let mut n = StateNotes::default();
        n.observe("상태", 0);
        assert_eq!(
            n.observe("상태", STILL_WAITING_SECS).as_deref(),
            Some("아직 기다리는 중: 상태")
        );
        uninstall();
        assert_eq!(
            *seen.borrow(),
            [
                "추적 중인 파일 훑는 중 (1/3)",
                "풀 리퀘스트 #12 확인 중 (1/3)"
            ]
        );
        assert_eq!(
            say(Msg::CiNoRunYet, &[("head", "0123abcd")]),
            "ci on 0123abcd: no workflow run yet"
        );
        // English sentences are today's strings, byte for byte.
        install(Box::new(|_| {}));
        assert_eq!(say(Msg::BuildingViews, &[]), "building the views (2/3)");
        assert_eq!(
            say(
                Msg::CiRunsCompleted,
                &[("head", "01234567"), ("done", "0"), ("total", "1")]
            ),
            "ci on 01234567: 0 of 1 runs completed"
        );
        assert_eq!(
            say(Msg::PullRequestStillAt, &[("pr", "7"), ("head", "abc")]),
            "pull request #7 still at abc"
        );
        assert_eq!(
            say(Msg::NoNewComment, &[("issue", "9")]),
            "no new comment on #9"
        );
        assert_eq!(
            say(
                Msg::ReadingPage,
                &[("n", "2"), ("path", "/repos/o/r/issues")]
            ),
            "reading page 2 of /repos/o/r/issues"
        );
        uninstall();
    }

    /// A Korean question keeps `[y/N]` and the `y`/`yes` answers (leader decision S6-4):
    /// `예`/`네` are not a yes.
    #[test]
    fn a_korean_question_keeps_the_y_n_answers() {
        let painter = Painter::plain().with_lang(Lang::Ko);
        for (answer, want) in [
            ("y\n", true),
            ("yes\n", true),
            ("예\n", false),
            ("네\n", false),
            ("\n", false),
        ] {
            let mut err = Vec::new();
            let got = ask(
                &mut answer.as_bytes(),
                &mut err,
                &painter,
                "미리보기\n",
                painter.text(Msg::RemoveIt),
            )
            .unwrap();
            assert_eq!(got, want, "{answer:?}");
            assert_eq!(
                String::from_utf8(err).unwrap(),
                "미리보기\n! 지울까요? [y/N] "
            );
        }
    }
}
