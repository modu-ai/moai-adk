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

    /// The word that carries the meaning without color or symbol, in English.
    pub fn word(self) -> &'static str {
        self.word_in(crate::ui_text::Lang::En)
    }

    /// The word in `lang` (S6b): the same meaning as [`Token::word`], from the sentence table.
    pub fn word_in(self, lang: crate::ui_text::Lang) -> &'static str {
        use crate::ui_text::{text, Msg};
        match self {
            Token::Ok => text(Msg::WordOk, lang),
            Token::Warn => text(Msg::WordWarning, lang),
            Token::Error => text(Msg::WordError, lang),
            Token::Info => text(Msg::WordNext, lang),
            Token::Accent | Token::Muted | Token::Code | Token::Strong => "",
        }
    }
}

/// `text` safe to show on a terminal (Issue #2818 audit B1, CWE-150; leader decision 6867586011):
/// every control character — C0 including tab, CR and LF, DEL, and C1 (`\u{80}`–`\u{9f}`, which
/// some terminals read as escape introducers) — becomes a visible `\xNN`. A file or folder name
/// can then never move the cursor, retitle the window, write the clipboard or start a link.
/// Text without control characters is returned as it is, and the result is stable when applied
/// twice. Human output only: the stdout JSON envelope escapes these itself and never comes here.
pub fn visible(text: &str) -> std::borrow::Cow<'_, str> {
    if !text.chars().any(char::is_control) {
        return std::borrow::Cow::Borrowed(text);
    }
    let mut out = String::with_capacity(text.len() + 8);
    for c in text.chars() {
        if c.is_control() {
            out.push_str(&format!("\\x{:02x}", c as u32));
        } else {
            out.push(c);
        }
    }
    std::borrow::Cow::Owned(out)
}

/// Paints text for one stream: color and symbols as decided, plain text otherwise. Every text it
/// paints goes through [`visible`] first, so only the painter's own style codes reach the terminal.
/// `lang` is the language of the stream's human sentences (S6b): a screen takes its words from
/// [`crate::ui_text::text`] in this language, so one stream never mixes two.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Painter {
    pub color: bool,
    pub unicode: bool,
    pub lang: crate::ui_text::Lang,
}

impl Painter {
    /// The painter for `stream` under the current environment and `flag`; its language is
    /// [`ui_lang`] of the same stream.
    pub fn for_stream(stream: Stream, flag: ColorFlag) -> Painter {
        let env = Env::from_process();
        Painter {
            color: color_enabled(flag, &env, stream.is_terminal()),
            unicode: unicode_enabled(&env),
            lang: ui_lang(stream),
        }
    }

    /// No color, ASCII symbols, English: the pipe / `TERM=dumb` shape.
    pub fn plain() -> Painter {
        Painter {
            color: false,
            unicode: false,
            lang: crate::ui_text::Lang::En,
        }
    }

    /// The same painter speaking `lang` (`MOAI_OUTPUT=plain` drops color and UTF-8 symbols, not
    /// the language of a terminal).
    pub fn with_lang(self, lang: crate::ui_text::Lang) -> Painter {
        Painter { lang, ..self }
    }

    /// `msg` in this painter's language (no data filled in).
    pub fn text(&self, msg: crate::ui_text::Msg) -> &'static str {
        crate::ui_text::text(msg, self.lang)
    }

    /// `msg` in this painter's language with `values` filled in through [`visible`].
    pub fn fill(&self, msg: crate::ui_text::Msg, values: &[(&str, &str)]) -> String {
        crate::ui_text::fill(msg, self.lang, values)
    }

    /// `text` in the token's style, or unchanged without color.
    pub fn paint(&self, token: Token, text: &str) -> String {
        let text = visible(text);
        if !self.color || text.is_empty() {
            return text.into_owned();
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

/// The language of human output on `stream`: English, on every stream and under every setting
/// (Issue #2825 R1, user decision 6878265192 — the CLI's human output is English only; the
/// user's language belongs to the LLM session). `MOAI_LANG`, the profile's `language.user_chat`
/// and the locale are not read. The Korean texts that S6a–S6e added stay unused until R2 removes
/// them.
pub fn ui_lang(_stream: Stream) -> crate::ui_text::Lang {
    crate::ui_text::Lang::En
}

/// Whether stderr is a terminal: the error frame (D2 §4) is drawn only there, so a pipe — a
/// lane, a script, a test — keeps today's one-line message byte for byte (D3 6865936070).
pub fn stderr_is_terminal() -> bool {
    Stream::Stderr.is_terminal()
}

/// The stderr error frame (D2 §4, card #2804): what failed, then the indented `detail` lines
/// as given (a usage error's missing argument names), then why and what to do next. The last
/// line is always the next step when there is one (clig.dev "Errors"). Every label is a word,
/// so the frame reads the same without color. Every text goes through [`visible`] first: an
/// operation message can carry a file or folder name, and only the frame's own newlines and
/// style codes may reach the terminal (leader decision 6867586011, #2818 audit B1).
pub fn error_frame(
    p: &Painter,
    what: &str,
    detail: &[&str],
    cause: Option<&str>,
    next: Option<&str>,
) -> String {
    // S6c (#2825): the labels are words in the stream's language — the `error` and `next`
    // token words (S6b) and the `cause` sentence; the texts inside stay as given.
    let label = |word: &str| format!("{word}:");
    let mut out = format!(
        "{} {}\n",
        p.paint(Token::Error, &label(Token::Error.word_in(p.lang))),
        visible(what)
    );
    for line in detail {
        out.push_str(&visible(line));
        out.push('\n');
    }
    if let Some(cause) = cause {
        out.push_str(&format!(
            "  {} {}\n",
            p.paint(
                Token::Muted,
                &label(p.text(crate::ui_text::Msg::LabelCause))
            ),
            visible(cause)
        ));
    }
    if let Some(next) = next {
        out.push_str(&format!(
            "  {} {}\n",
            p.paint(Token::Info, &label(Token::Info.word_in(p.lang))),
            visible(next)
        ));
    }
    out
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
            lang: crate::ui_text::Lang::En,
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

    #[test]
    fn control_characters_become_visible_text() {
        // C0 (NUL, BEL, TAB, LF, CR, ESC), DEL and C1 (CSI U+009B, the last C1 U+009F).
        let raw = "a\u{0}b\u{7}c\td\ne\rf\u{1b}]0;T\u{7}\u{1b}[2Jg\u{7f}h\u{9b}31mi\u{9f}j";
        let shown = visible(raw);
        assert_eq!(
            shown,
            "a\\x00b\\x07c\\x09d\\x0ae\\x0df\\x1b]0;T\\x07\\x1b[2Jg\\x7fh\\x9b31mi\\x9fj"
        );
        assert!(!shown.chars().any(char::is_control), "{shown:?}");
        // Plain text, Hangul and wide symbols pass unchanged and borrowed; applying twice is stable.
        assert!(matches!(
            visible("한글 path/✓"),
            std::borrow::Cow::Borrowed(_)
        ));
        assert_eq!(visible(&shown), shown);
        // U+00A0 (after C1) and U+0080 (first C1) are the boundaries.
        assert_eq!(visible("\u{a0}"), "\u{a0}");
        assert_eq!(visible("\u{80}"), "\\x80");
    }

    #[test]
    fn painting_never_passes_a_control_character_from_the_text() {
        let raw = "dir\u{1b}]52;c;Zm9v\u{7}\u{9b}2J";
        for p in [
            Painter::plain(),
            Painter {
                color: true,
                unicode: true,
                lang: crate::ui_text::Lang::En,
            },
        ] {
            for t in [Token::Ok, Token::Muted, Token::Strong] {
                let painted = p.paint(t, raw);
                let escapes = painted.matches('\u{1b}').count();
                // Only the painter's own SGR start and reset remain.
                assert_eq!(escapes, if p.color { 2 } else { 0 }, "{painted:?}");
                assert!(
                    !painted.contains('\u{7}') && !painted.contains('\u{9b}'),
                    "{painted:?}"
                );
                assert!(
                    painted.contains("\\x1b]52;c;Zm9v\\x07\\x9b2J"),
                    "{painted:?}"
                );
            }
            assert!(!p
                .status(Token::Error, "con\u{1b}[2Jflict")
                .contains("\u{1b}[2J"));
        }
    }

    /// Card #2804 follow-up (leader decision 6867586011): a control character in any part of the
    /// frame — a path in an operation message — reaches the terminal as visible `\xNN`; the only
    /// newlines are the frame's own and, in color, the only escapes are the painter's.
    #[test]
    fn the_error_frame_shows_control_characters_as_text() {
        let evil = "/tmp/a\u{1b}]0;pwn\u{7}b\nc\u{9b}d\u{7f}";
        let shown = "/tmp/a\\x1b]0;pwn\\x07b\\x0ac\\x9bd\\x7f";
        let plain = Painter::plain();
        assert_eq!(
            error_frame(&plain, evil, &[evil], Some(evil), Some(evil)),
            format!("error: {shown}\n{shown}\n  cause: {shown}\n  next: {shown}\n")
        );
        let color = Painter {
            color: true,
            unicode: true,
            lang: crate::ui_text::Lang::En,
        };
        let painted = error_frame(&color, evil, &[evil], Some(evil), Some(evil));
        assert_eq!(painted.lines().count(), 4, "{painted:?}");
        assert!(
            !painted.contains('\u{7}')
                && !painted.contains('\u{9b}')
                && !painted.contains('\u{7f}')
        );
        // Every ESC left is the start of one of the painter's own style codes (`ESC [ … m`).
        for (i, _) in painted.match_indices('\u{1b}') {
            assert!(painted[i..].starts_with("\u{1b}["), "{painted:?}");
            assert!(!painted[i..].starts_with("\u{1b}]"), "{painted:?}");
        }
        // Text without control characters is unchanged.
        assert_eq!(
            error_frame(&plain, "plain text", &[], None, None),
            "error: plain text\n"
        );
    }

    /// S6c (#2825): a Korean stream labels the frame `오류:`, `원인:`, `다음:` — the token words
    /// of S6b and one new sentence — while the text inside stays as given. Control characters
    /// in the given text are still shown as `\xNN` (agent-22 #2887 note).
    #[test]
    fn a_korean_frame_has_korean_labels() {
        let ko = Painter::plain().with_lang(crate::ui_text::Lang::Ko);
        assert_eq!(
            error_frame(
                &ko,
                "invalid typed JSON input",
                &["  <PROJECT>"],
                Some("c"),
                Some("n")
            ),
            "오류: invalid typed JSON input\n  <PROJECT>\n  원인: c\n  다음: n\n"
        );
        let color = Painter {
            color: true,
            unicode: true,
            lang: crate::ui_text::Lang::Ko,
        };
        assert_eq!(
            error_frame(&color, "x", &[], None, Some("n")),
            "\u{1b}[31m오류:\u{1b}[0m x\n  \u{1b}[36m다음:\u{1b}[0m n\n"
        );
        let evil = "a\u{1b}]0;t\u{7}b";
        let shown = error_frame(&ko, evil, &[evil], Some(evil), Some(evil));
        assert!(
            !shown.chars().any(|c| c.is_control() && c != '\n'),
            "{shown:?}"
        );
        assert_eq!(shown.matches("a\\x1b]0;t\\x07b").count(), 4, "{shown}");
    }

    /// Card #2804: the frame's golden shapes. Plain (NO_COLOR, TERM=dumb, ASCII) is the same
    /// text with no escape; color paints only the labels; `next` is always the last line.
    #[test]
    fn the_error_frame_reads_what_detail_cause_then_next() {
        let plain = Painter::plain();
        assert_eq!(
            error_frame(
                &plain,
                "the following required arguments were not provided:",
                &["  <PROJECT>"],
                Some("the plan is older than the files"),
                Some("run `moai project-plan --help` to see its arguments"),
            ),
            "error: the following required arguments were not provided:\n  <PROJECT>\n  cause: the plan is older than the files\n  next: run `moai project-plan --help` to see its arguments\n"
        );
        // Nothing optional: one line, the same words a pipe would show after `error: `.
        assert_eq!(
            error_frame(&plain, "invalid typed JSON input", &[], None, None),
            "error: invalid typed JSON input\n"
        );
        let ascii = Painter {
            color: false,
            unicode: false,
            lang: crate::ui_text::Lang::En,
        };
        let unicode = Painter {
            color: false,
            unicode: true,
            lang: crate::ui_text::Lang::En,
        };
        let f = |p: &Painter| error_frame(p, "x", &[], None, Some("y"));
        assert_eq!(f(&ascii), f(&unicode), "no symbol rides in the frame");

        let color = Painter {
            color: true,
            unicode: true,
            lang: crate::ui_text::Lang::En,
        };
        let painted = error_frame(&color, "x", &["  d"], Some("c"), Some("n"));
        assert_eq!(
            painted,
            "\u{1b}[31merror:\u{1b}[0m x\n  d\n  \u{1b}[90mcause:\u{1b}[0m c\n  \u{1b}[36mnext:\u{1b}[0m n\n"
        );
        // Stripping the escapes gives the plain frame back: color carries no meaning alone.
        let stripped: String = {
            let mut out = String::new();
            let mut it = painted.chars().peekable();
            while let Some(ch) = it.next() {
                if ch == '\u{1b}' {
                    for c in it.by_ref() {
                        if c == 'm' {
                            break;
                        }
                    }
                } else {
                    out.push(ch);
                }
            }
            out
        };
        assert_eq!(
            stripped,
            error_frame(&plain, "x", &["  d"], Some("c"), Some("n"))
        );
    }
}
