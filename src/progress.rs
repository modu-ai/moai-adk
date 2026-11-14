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

use crate::ui::{ColorFlag, Painter, Stream, Token};
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

/// One progress line: the info mark (`→`, or `->` without UTF-8) and the text.
pub fn line(painter: &Painter, text: &str) -> String {
    format!("{} {text}\n", painter.mark(Token::Info))
}

/// Installs a stderr sink when stderr is a terminal; otherwise leaves [`note`] silent.
pub fn install_stderr() {
    if !io::stderr().is_terminal() {
        return;
    }
    let painter = Painter::for_stream(Stream::Stderr, ColorFlag::Auto);
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
        "{preview}{} {question} [y/N] ",
        painter.mark(Token::Warn)
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
    let painter = Painter::for_stream(Stream::Stderr, ColorFlag::Auto);
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
}
