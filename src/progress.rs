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
use std::cell::RefCell;
use std::io::{self, BufRead, IsTerminal, Write};

type Sink = Box<dyn FnMut(&str)>;

thread_local! {
    static SINK: RefCell<Option<Sink>> = const { RefCell::new(None) };
}

/// Reports one stage of a long-running command; a no-op unless a sink is installed.
pub fn note(text: &str) {
    SINK.with(|s| {
        if let Some(sink) = s.borrow_mut().as_mut() {
            sink(text);
        }
    });
}

/// Installs `sink` for this thread's later [`note`] calls (tests pass a recorder).
pub fn install(sink: Sink) {
    SINK.with(|s| *s.borrow_mut() = Some(sink));
}

/// Removes the sink; later [`note`] calls write nothing.
pub fn uninstall() {
    SINK.with(|s| *s.borrow_mut() = None);
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

/// The stderr painter `MOAI_OUTPUT` asks for; `None` for progress under `quiet`.
pub fn progress_painter(mode: OutputMode, normal: impl FnOnce() -> Painter) -> Option<Painter> {
    match mode {
        OutputMode::Quiet => None,
        OutputMode::Plain => Some(Painter::plain()),
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
    install(Box::new(move |text| {
        let mut err = io::stderr();
        let _ = err.write_all(line(&painter, text).as_bytes());
        let _ = err.flush();
    }));
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
            return Some(format!("still waiting: {state}"));
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
/// with nothing written.
pub fn ask_terminal(preview: &str, question: &str) -> Result<bool, String> {
    if !(io::stdin().is_terminal() && io::stderr().is_terminal()) {
        return Ok(false);
    }
    // The question is asked under `quiet` too; `plain` only drops color and UTF-8 symbols.
    let painter = progress_painter(output_mode(), || {
        Painter::for_stream(Stream::Stderr, ColorFlag::Auto)
    })
    .unwrap_or_else(Painter::plain);
    ask(
        &mut io::stdin().lock(),
        &mut io::stderr(),
        &painter,
        preview,
        question,
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
    }
}
