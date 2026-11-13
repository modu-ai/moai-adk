//! Human output tokens and the per-stream color decision (Issue #2803, design #2799 D2 §2·§3·§8,
//! leader decision D3 6865936070).
//!
//! Only the human layer uses this module: `--help`, the banner, and (later cards S3–S5) the
//! stderr error frame, summaries, prompts and progress lines. The stdout JSON envelope, hook
//! stderr and the status line never go through it, so their bytes cannot change here.
//!
//! Color decision, checked for each stream on its own (clig.dev "Output": "It's best to
//! individually check"):
//!
//! 1. an explicit flag (`--color=never|always`) — a per-instance argument wins over `NO_COLOR`
//!    (no-color.org FAQ 2); the flag itself is wired by the help card, callers pass it here;
//! 2. `NO_COLOR` present and not empty → off (no-color.org);
//! 3. `TERM=dumb` → off;
//! 4. `FORCE_COLOR` or `CLICOLOR_FORCE` set, not empty and not `0` → on;
//! 5. otherwise on only when that stream is a terminal.
//!
//! Colors are the 4-bit ANSI palette, so the user's terminal theme decides the actual shade
//! (GitHub CLI accessibility work, 2025). Every token also has a symbol with an ASCII fallback
//! and a word, so no meaning rides on color alone.
//!
//! Status line correspondence (names only; the status line keeps its own theme bytes):
//! `ok` ↔ success/green segments, `warn` ↔ near-limit, `error` ↔ failure, `accent` ↔ the
//! MoAI coral, `muted` ↔ dim text. Banner depth (truecolor / 256) uses the status line's
//! [`crate::statusline::color::Profile`] arithmetic, so one place converts colors.

use anstyle::{AnsiColor, Style};
use std::ffi::OsString;
use std::io::IsTerminal;

/// Which standard stream a piece of human output goes to.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Stream {
    Stdout,
    Stderr,
}

impl Stream {
    fn is_terminal(self) -> bool {
        match self {
            Stream::Stdout => std::io::stdout().is_terminal(),
            Stream::Stderr => std::io::stderr().is_terminal(),
        }
    }
}

/// An explicit color choice from the command line.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ColorFlag {
    Auto,
    Always,
    Never,
}

/// The environment values the decisions read, captured once so tests can pass their own.
#[derive(Clone, Debug, Default)]
pub struct Env {
    pub no_color: Option<OsString>,
    pub term: Option<OsString>,
    pub force_color: Option<OsString>,
    pub clicolor_force: Option<OsString>,
    pub colorterm: Option<OsString>,
    pub lc_all: Option<OsString>,
    pub lc_ctype: Option<OsString>,
    pub lang: Option<OsString>,
    pub columns: Option<OsString>,
}

impl Env {
    /// The current process environment.
    pub fn from_process() -> Env {
        let v = std::env::var_os;
        Env {
            no_color: v("NO_COLOR"),
            term: v("TERM"),
            force_color: v("FORCE_COLOR"),
            clicolor_force: v("CLICOLOR_FORCE"),
            colorterm: v("COLORTERM"),
            lc_all: v("LC_ALL"),
            lc_ctype: v("LC_CTYPE"),
            lang: v("LANG"),
            columns: v("COLUMNS"),
        }
    }

    fn set(value: &Option<OsString>) -> bool {
        value.as_ref().is_some_and(|v| !v.is_empty())
    }

    fn forced(value: &Option<OsString>) -> bool {
        value.as_ref().is_some_and(|v| !v.is_empty() && v != "0")
    }

    /// `TERM=dumb`: a terminal that understands no escape sequences.
    pub fn term_is_dumb(&self) -> bool {
        self.term.as_ref().is_some_and(|t| t == "dumb")
    }
}

/// Whether a stream gets ANSI color, in the order the module comment lists.
pub fn color_enabled(flag: ColorFlag, env: &Env, is_terminal: bool) -> bool {
    match flag {
        ColorFlag::Always => return true,
        ColorFlag::Never => return false,
        ColorFlag::Auto => {}
    }
    if Env::set(&env.no_color) || env.term_is_dumb() {
        return false;
    }
    if Env::forced(&env.force_color) || Env::forced(&env.clicolor_force) {
        return true;
    }
    is_terminal
}

/// Whether symbols may use UTF-8 (`✓ ✗ → ◆`); otherwise the ASCII fallbacks. The first set
/// value of `LC_ALL`, `LC_CTYPE`, `LANG` decides (POSIX order); none set counts as UTF-8,
/// which is what current terminals use. `TERM=dumb` always gets ASCII.
pub fn unicode_enabled(env: &Env) -> bool {
    if env.term_is_dumb() {
        return false;
    }
    let locale = [&env.lc_all, &env.lc_ctype, &env.lang]
        .into_iter()
        .find_map(|v| v.as_ref().filter(|v| !v.is_empty()));
    match locale {
        None => true,
        Some(value) => {
            let value = value.to_string_lossy().to_ascii_lowercase();
            value.contains("utf-8") || value.contains("utf8")
        }
    }
}

/// How many colors the terminal shows, for surfaces that want more than the 4-bit tokens
/// (the banner gradient).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Depth {
    Ansi16,
    Ansi256,
    TrueColor,
}

impl Depth {
    /// The status line's color arithmetic for this depth; `None` for 16 colors.
    pub fn profile(self) -> Option<crate::statusline::color::Profile> {
        use crate::statusline::color::Profile;
        match self {
            Depth::TrueColor => Some(Profile::TrueColor),
            Depth::Ansi256 => Some(Profile::Ansi256),
            Depth::Ansi16 => None,
        }
    }
}

/// `COLORTERM=truecolor|24bit` → 24-bit; a `TERM` naming 256 colors → 256; else 16.
pub fn depth(env: &Env) -> Depth {
    let colorterm = env
        .colorterm
        .as_ref()
        .map(|c| c.to_string_lossy().trim().to_ascii_lowercase())
        .unwrap_or_default();
    if colorterm == "truecolor" || colorterm == "24bit" {
        return Depth::TrueColor;
    }
    if env
        .term
        .as_ref()
        .is_some_and(|t| t.to_string_lossy().contains("256"))
    {
        return Depth::Ansi256;
    }
    Depth::Ansi16
}

/// The width to lay human output out in: `COLUMNS` when it is a positive number (an explicit
/// choice wins), else the stream's terminal size, else `None` (callers pick their default).
pub fn width(stream: Stream, env: &Env) -> Option<usize> {
    if let Some(n) = env
        .columns
        .as_ref()
        .and_then(|c| c.to_str())
        .and_then(|c| c.trim().parse::<usize>().ok())
        .filter(|n| *n > 0)
    {
        return Some(n);
    }
    terminal_columns(stream)
}

#[cfg(unix)]
fn terminal_columns(stream: Stream) -> Option<usize> {
    let fd = match stream {
        Stream::Stdout => libc::STDOUT_FILENO,
        Stream::Stderr => libc::STDERR_FILENO,
    };
    let mut size = libc::winsize {
        ws_row: 0,
        ws_col: 0,
        ws_xpixel: 0,
        ws_ypixel: 0,
    };
    // SAFETY: TIOCGWINSZ writes one `winsize` into the pointer we pass; a non-terminal fd
    // answers an error and leaves it untouched.
    let rc = unsafe { libc::ioctl(fd, libc::TIOCGWINSZ, &mut size) };
    (rc == 0 && size.ws_col > 0).then_some(size.ws_col as usize)
}

#[cfg(not(unix))]
fn terminal_columns(_stream: Stream) -> Option<usize> {
    None
}

/// The semantic tokens of the human layer (D2 §2).
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Token {
    /// Done, passed, nothing to write.
    Ok,
    /// Needs attention or a user decision.
    Warn,
    /// Failed or refused.
    Error,
    /// Progress and guidance.
    Info,
    /// Titles and the MoAI mark.
    Accent,
    /// Secondary detail: paths, ids.
    Muted,
    /// Commands, flags, file names.
    Code,
    /// Table heads and key values.
    Strong,
}

impl Token {
    /// The 4-bit style; the terminal theme picks the shade.
    pub fn style(self) -> Style {
        let fg = |c: AnsiColor| Style::new().fg_color(Some(c.into()));
        match self {
            Token::Ok => fg(AnsiColor::Green),
            Token::Warn => fg(AnsiColor::Yellow),
            Token::Error => fg(AnsiColor::Red),
            Token::Info => fg(AnsiColor::Cyan),
            Token::Accent => fg(AnsiColor::Magenta),
            Token::Muted => fg(AnsiColor::BrightBlack),
            Token::Code | Token::Strong => Style::new().bold(),
        }
    }

    /// The symbol for `unicode` terminals, or its ASCII fallback; empty for text-only tokens.
    pub fn symbol(self, unicode: bool) -> &'static str {
        match (self, unicode) {
            (Token::Ok, true) => "✓",
            (Token::Ok, false) => "ok",
            (Token::Warn, _) => "!",
            (Token::Error, true) => "✗",
            (Token::Error, false) => "x",
            (Token::Info, true) => "→",
            (Token::Info, false) => "->",
            (Token::Accent, true) => "◆",
            (Token::Accent, false) => "*",
            (Token::Muted | Token::Code | Token::Strong, _) => "",
        }
    }

    /// The word that carries the meaning without color or symbol.
    pub fn word(self) -> &'static str {
        match self {
            Token::Ok => "ok",
            Token::Warn => "warning",
            Token::Error => "error",
            Token::Info => "next",
            Token::Accent | Token::Muted | Token::Code | Token::Strong => "",
        }
    }
}

/// Paints text for one stream: color and symbols as decided, plain text otherwise.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Painter {
    pub color: bool,
    pub unicode: bool,
}

impl Painter {
    /// The painter for `stream` under the current environment and `flag`.
    pub fn for_stream(stream: Stream, flag: ColorFlag) -> Painter {
        let env = Env::from_process();
        Painter {
            color: color_enabled(flag, &env, stream.is_terminal()),
            unicode: unicode_enabled(&env),
        }
    }

    /// No color, ASCII symbols: the pipe / `TERM=dumb` shape.
    pub fn plain() -> Painter {
        Painter {
            color: false,
            unicode: false,
        }
    }

    /// `text` in the token's style, or unchanged without color.
    pub fn paint(&self, token: Token, text: &str) -> String {
        if !self.color || text.is_empty() {
            return text.to_string();
        }
        let style = token.style();
        format!("{}{text}{}", style.render(), style.render_reset())
    }

    /// The token's symbol (UTF-8 or ASCII), painted.
    pub fn mark(&self, token: Token) -> String {
        self.paint(token, token.symbol(self.unicode))
    }

    /// A status cell: symbol and word together, painted (`✓ create`, `x conflict`), so the
    /// meaning survives without color.
    pub fn status(&self, token: Token, word: &str) -> String {
        let symbol = token.symbol(self.unicode);
        let text = if symbol.is_empty() {
            word.to_string()
        } else {
            format!("{symbol} {word}")
        };
        self.paint(token, &text)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn env(pairs: &[(&str, &str)]) -> Env {
        let mut e = Env::default();
        for (k, v) in pairs {
            let v = Some(OsString::from(v));
            match *k {
                "NO_COLOR" => e.no_color = v,
                "TERM" => e.term = v,
                "FORCE_COLOR" => e.force_color = v,
                "CLICOLOR_FORCE" => e.clicolor_force = v,
                "COLORTERM" => e.colorterm = v,
                "LC_ALL" => e.lc_all = v,
                "LC_CTYPE" => e.lc_ctype = v,
                "LANG" => e.lang = v,
                "COLUMNS" => e.columns = v,
                other => panic!("unknown key {other}"),
            }
        }
        e
    }

    #[test]
    fn the_color_decision_follows_the_documented_order() {
        use ColorFlag::*;
        // (flag, environment, stream is a terminal, color expected)
        type Case<'a> = (ColorFlag, &'a [(&'a str, &'a str)], bool, bool);
        let cases: &[Case] = &[
            (Auto, &[], true, true),
            (Auto, &[], false, false),
            // NO_COLOR: present and not empty turns color off; empty counts as unset.
            (Auto, &[("NO_COLOR", "1")], true, false),
            (Auto, &[("NO_COLOR", "0")], true, false),
            (Auto, &[("NO_COLOR", "")], true, true),
            (Auto, &[("TERM", "dumb")], true, false),
            (Auto, &[("TERM", "xterm-256color")], true, true),
            // FORCE_COLOR / CLICOLOR_FORCE: on even without a terminal, but not over NO_COLOR.
            (Auto, &[("FORCE_COLOR", "1")], false, true),
            (Auto, &[("FORCE_COLOR", "0")], false, false),
            (Auto, &[("FORCE_COLOR", "")], false, false),
            (Auto, &[("CLICOLOR_FORCE", "1")], false, true),
            (
                Auto,
                &[("FORCE_COLOR", "1"), ("NO_COLOR", "1")],
                false,
                false,
            ),
            (Auto, &[("FORCE_COLOR", "1"), ("TERM", "dumb")], true, false),
            // An explicit flag wins over everything.
            (Always, &[("NO_COLOR", "1"), ("TERM", "dumb")], false, true),
            (Never, &[("FORCE_COLOR", "1")], true, false),
        ];
        for (flag, pairs, tty, want) in cases {
            assert_eq!(
                color_enabled(*flag, &env(pairs), *tty),
                *want,
                "{flag:?} {pairs:?} tty={tty}"
            );
        }
    }

    #[test]
    fn utf8_symbols_follow_the_first_set_locale_variable() {
        assert!(unicode_enabled(&env(&[])));
        assert!(unicode_enabled(&env(&[("LANG", "ko_KR.UTF-8")])));
        assert!(unicode_enabled(&env(&[("LANG", "en_US.utf8")])));
        assert!(!unicode_enabled(&env(&[("LANG", "C")])));
        assert!(!unicode_enabled(&env(&[("LANG", "POSIX")])));
        // LC_ALL wins over LANG; LC_CTYPE over LANG; empty values are skipped.
        assert!(!unicode_enabled(&env(&[
            ("LC_ALL", "C"),
            ("LANG", "en_US.UTF-8")
        ])));
        assert!(unicode_enabled(&env(&[
            ("LC_CTYPE", "en_US.UTF-8"),
            ("LANG", "C")
        ])));
        assert!(unicode_enabled(&env(&[
            ("LC_ALL", ""),
            ("LANG", "en_US.UTF-8")
        ])));
        assert!(!unicode_enabled(&env(&[
            ("TERM", "dumb"),
            ("LANG", "en_US.UTF-8")
        ])));
    }

    #[test]
    fn depth_reads_colorterm_then_term() {
        assert_eq!(depth(&env(&[("COLORTERM", "truecolor")])), Depth::TrueColor);
        assert_eq!(depth(&env(&[("COLORTERM", " 24BIT ")])), Depth::TrueColor);
        assert_eq!(depth(&env(&[("TERM", "xterm-256color")])), Depth::Ansi256);
        assert_eq!(depth(&env(&[("TERM", "xterm")])), Depth::Ansi16);
        assert_eq!(depth(&env(&[])), Depth::Ansi16);
        assert_eq!(Depth::Ansi16.profile(), None);
    }

    #[test]
    fn columns_wins_when_it_is_a_positive_number() {
        assert_eq!(width(Stream::Stdout, &env(&[("COLUMNS", "60")])), Some(60));
        assert_eq!(
            width(Stream::Stdout, &env(&[("COLUMNS", " 61 ")])),
            Some(61)
        );
        // Not a positive number: falls through to the terminal (none under `cargo test`).
        let fallback = terminal_columns(Stream::Stdout);
        assert_eq!(width(Stream::Stdout, &env(&[("COLUMNS", "0")])), fallback);
        assert_eq!(
            width(Stream::Stdout, &env(&[("COLUMNS", "wide")])),
            fallback
        );
    }

    #[test]
    fn plain_painting_carries_no_escape_and_keeps_the_word() {
        let p = Painter::plain();
        for t in [
            Token::Ok,
            Token::Warn,
            Token::Error,
            Token::Info,
            Token::Accent,
            Token::Muted,
            Token::Code,
            Token::Strong,
        ] {
            assert!(!p.paint(t, "text").contains('\x1b'));
            assert!(!p.status(t, "create").contains('\x1b'));
            assert!(p.status(t, "create").ends_with("create"));
            assert!(p.mark(t).is_ascii(), "{t:?}");
        }
        assert_eq!(p.status(Token::Error, "conflict"), "x conflict");
        assert_eq!(p.status(Token::Ok, "create"), "ok create");
    }

    #[test]
    fn colored_painting_uses_the_four_bit_palette() {
        let p = Painter {
            color: true,
            unicode: true,
        };
        assert_eq!(p.status(Token::Ok, "create"), "\x1b[32m✓ create\x1b[0m");
        assert_eq!(p.paint(Token::Error, "x"), "\x1b[31mx\x1b[0m");
        assert_eq!(p.paint(Token::Warn, "x"), "\x1b[33mx\x1b[0m");
        assert_eq!(p.paint(Token::Info, "x"), "\x1b[36mx\x1b[0m");
        assert_eq!(p.paint(Token::Accent, "x"), "\x1b[35mx\x1b[0m");
        assert_eq!(p.paint(Token::Muted, "x"), "\x1b[90mx\x1b[0m");
        assert_eq!(p.paint(Token::Strong, "x"), "\x1b[1mx\x1b[0m");
        // Only 4-bit codes: no 256-color or 24-bit sequences in the token layer.
        for t in [
            Token::Ok,
            Token::Warn,
            Token::Error,
            Token::Info,
            Token::Accent,
            Token::Muted,
        ] {
            let s = p.paint(t, "x");
            assert!(!s.contains("38;5;") && !s.contains("38;2;"), "{t:?}: {s:?}");
        }
        assert_eq!(p.paint(Token::Ok, ""), "");
    }
}
