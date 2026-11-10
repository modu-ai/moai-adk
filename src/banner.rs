//! The startup banner for the `moai-cli` human surfaces (Issue #2363 addition).
//!
//! Composition mirrors the Go CLI banner (go internal/cli/root.go,
//! internal/cli/uikit/banner.go, internal/tui/logo.go): the 6-line block-letter
//! MoAI-ADK logo plus the identity line `◆ MoAI-ADK Modu-AI's Agentic
//! Development Kit w/ SuperAgent MoAI`. The Go pill row (version / go / claude)
//! is not part of the requirement and is omitted.
//!
//! Print rules (Issue #2363 design comment 6760944791):
//! - nothing is printed unless stdout is a terminal. This is stricter than Go,
//!   which keeps printing a degraded banner into a pipe; the operator dispatch
//!   requires no banner in piped output.
//! - a non-empty `NO_COLOR` (an empty value counts as unset, per the NO_COLOR
//!   convention and the Go golden-test intent) downgrades to the plain variant.
//! - JSON envelope output never carries the banner: the only print sites are
//!   the bare no-argument run and the root-help shape; subcommands and
//!   subcommand help are untouched, so the output contract (envelope,
//!   operation names) stays byte-identical.
//!
//! Colours are the Go dark-theme coral tokens, pinned because the Rust RC has
//! no theme subsystem: Accent `#d97757`, AccentDeep `#b85e3f`, Body `#d8dedb`.
//! The logo rows get a 6-stop linear-RGB gradient between Accent and
//! AccentDeep in 24-bit ANSI (not lipgloss's CIELAB blend — a documented
//! approximation), bold Accent for `◆ MoAI-ADK`, and Body for the rest of the
//! identity line.

use std::ffi::OsStr;
use std::io::{self, IsTerminal, Write};

/// The 6 logo rows, byte-identical to the Go `moaiLogoArt` const
/// (internal/tui/logo.go lines 17-22). `tests/banner_golden.rs` pins these
/// bytes against an extracted copy
/// (`tests/fixtures/banner/go-logo-rows.golden`), so a transcription slip
/// cannot pass silently.
pub const LOGO_ROWS: &str = concat!(
    "███╗   ███╗          █████╗ ██╗       █████╗ ██████╗ ██╗  ██╗\n",
    "████╗ ████║ ██████╗ ██╔══██╗██║      ██╔══██╗██╔══██╗██║ ██╔╝\n",
    "██╔████╔██║██║   ██║███████║██║█████╗███████║██║  ██║█████╔╝\n",
    "██║╚██╔╝██║██║   ██║██╔══██║██║╚════╝██╔══██║██║  ██║██╔═██╗\n",
    "██║ ╚═╝ ██║╚██████╔╝██║  ██║██║      ██║  ██║██████╔╝██║  ██╗\n",
    "╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝      ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝\n",
);

const IDENTITY_MARK: &str = "◆ MoAI-ADK";
const IDENTITY_REST: &str = "Modu-AI's Agentic Development Kit w/ SuperAgent MoAI";

const ACCENT: (u8, u8, u8) = (0xd9, 0x77, 0x57);
const ACCENT_DEEP: (u8, u8, u8) = (0xb8, 0x5e, 0x3f);
const BODY: (u8, u8, u8) = (0xd8, 0xde, 0xdb);

const RESET: &str = "\x1b[0m";

fn lerp(a: u8, b: u8, t: f64) -> u8 {
    (a as f64 + (b as f64 - a as f64) * t).round() as u8
}

fn foreground(rgb: (u8, u8, u8)) -> String {
    format!("\x1b[38;2;{};{};{}m", rgb.0, rgb.1, rgb.2)
}

fn gradient_stop(t: f64) -> (u8, u8, u8) {
    (
        lerp(ACCENT.0, ACCENT_DEEP.0, t),
        lerp(ACCENT.1, ACCENT_DEEP.1, t),
        lerp(ACCENT.2, ACCENT_DEEP.2, t),
    )
}

/// What a print site should emit: nothing (non-TTY), the plain variant
/// (`NO_COLOR`), or the coloured variant.
#[derive(Debug, PartialEq, Eq)]
pub enum Banner {
    None,
    Plain,
    Colored,
}

/// The print gate, factored out for tests: terminal presence decides whether
/// anything is printed at all; `NO_COLOR` only decides which bytes.
pub fn decide(is_tty: bool, no_color: Option<&OsStr>) -> Banner {
    if !is_tty {
        return Banner::None;
    }
    match no_color {
        Some(value) if !value.is_empty() => Banner::Plain,
        _ => Banner::Colored,
    }
}

fn decide_now() -> Banner {
    let no_color = std::env::var_os("NO_COLOR");
    decide(io::stdout().is_terminal(), no_color.as_deref())
}

fn emit(text: &str) {
    let _ = io::stdout().write_all(text.as_bytes());
}

/// The logo alone: 6 rows plus a trailing newline.
pub fn render_logo(colored: bool) -> String {
    if !colored {
        return LOGO_ROWS.to_string();
    }
    let rows: Vec<&str> = LOGO_ROWS.lines().collect();
    let last = rows.len() - 1;
    rows.iter()
        .enumerate()
        .map(|(i, row)| {
            let t = i as f64 / last as f64;
            format!("{}{}{RESET}", foreground(gradient_stop(t)), row)
        })
        .collect::<Vec<_>>()
        .join("\n")
        + "\n"
}

/// The banner: the logo plus the identity line below it.
pub fn render_banner(colored: bool) -> String {
    let identity = if colored {
        format!(
            "\x1b[1m{}{IDENTITY_MARK}{RESET} {}{IDENTITY_REST}{RESET}",
            foreground(ACCENT),
            foreground(BODY)
        )
    } else {
        format!("{IDENTITY_MARK} {IDENTITY_REST}")
    };
    format!("{}{}\n", render_logo(colored), identity)
}

/// Root-help shape check, mirroring Go internal/cli/fang.go `isRootHelpArgs`:
/// only a bare `--help`/`-h`/`help` carries the logo; the no-argument run (the
/// banner surface) and subcommand help stay logo-free, so the logo never
/// double-prints on one invocation.
pub fn is_root_help_args(args: &[String]) -> bool {
    match args.first().map(String::as_str) {
        Some("--help") | Some("-h") => true,
        Some("help") => args.len() == 1,
        _ => false,
    }
}

/// Bare no-argument run: the banner goes to stdout when the gate allows.
pub fn maybe_print_banner() {
    match decide_now() {
        Banner::None => {}
        Banner::Plain => emit(&render_banner(false)),
        Banner::Colored => emit(&render_banner(true)),
    }
}

/// Root help (`--help`/`-h`/`help`): the logo only, above clap's help body.
pub fn maybe_print_logo_for(args: &[String]) {
    if !is_root_help_args(args) {
        return;
    }
    match decide_now() {
        Banner::None => {}
        Banner::Plain => emit(&render_logo(false)),
        Banner::Colored => emit(&render_logo(true)),
    }
}
