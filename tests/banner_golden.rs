//! Golden tests for the startup banner (Issue #2363 addition).
//!
//! `MOAI_UPDATE_BANNER_GOLDEN=1` regenerates the fixture files (the Go
//! `UPDATE_GOLDEN=1` convention); without it the rendered bytes must equal the
//! committed fixtures exactly.

mod common;
use std::ffi::OsStr;
use std::path::Path;

use moai_cli::banner::{self, Banner};

const UPDATE_ENV: &str = "MOAI_UPDATE_BANNER_GOLDEN";
const GOLDEN_DIR: &str = "tests/fixtures/banner";

fn check(name: &str, got: &str) {
    let path = Path::new(GOLDEN_DIR).join(format!("{name}.golden"));
    if std::env::var_os(UPDATE_ENV).is_some() {
        std::fs::write(&path, got).expect("write golden fixture");
        return;
    }
    let want = std::fs::read_to_string(&path).unwrap_or_else(|e| {
        panic!(
            "read {}: {e} (set {UPDATE_ENV}=1 to generate the fixtures)",
            path.display()
        )
    });
    assert_eq!(
        got, want,
        "banner bytes changed for {name}; regenerate fixtures only on purpose"
    );
}

#[test]
fn logo_rows_are_byte_identical_to_the_go_const() {
    assert_eq!(
        banner::LOGO_ROWS,
        include_str!("fixtures/banner/go-logo-rows.golden"),
        "the logo rows must stay byte-identical to the Go moaiLogoArt const"
    );
}

#[test]
fn plain_banner_matches_its_golden() {
    check("banner-plain", &banner::render_banner(false));
}

#[test]
fn colored_banner_matches_its_golden() {
    check("banner-color", &banner::render_banner(true));
}

#[test]
fn plain_variants_carry_zero_ansi_bytes() {
    assert!(
        !banner::render_banner(false).contains('\x1b'),
        "the plain variant must not carry ANSI escapes"
    );
    assert!(
        !banner::render_logo(false).contains('\x1b'),
        "the plain logo must not carry ANSI escapes"
    );
}

#[test]
fn gate_suppresses_non_tty_and_respects_no_color() {
    assert_eq!(banner::decide(false, None), Banner::None);
    assert_eq!(banner::decide(false, Some(OsStr::new("1"))), Banner::None);
    assert_eq!(banner::decide(true, None), Banner::Colored);
    // An empty NO_COLOR counts as unset (NO_COLOR convention; Go test intent).
    assert_eq!(banner::decide(true, Some(OsStr::new(""))), Banner::Colored);
    assert_eq!(banner::decide(true, Some(OsStr::new("1"))), Banner::Plain);
    assert_eq!(banner::decide(true, Some(OsStr::new("0"))), Banner::Plain);
}

#[test]
fn root_help_shape_mirrors_go() {
    let args: Vec<String> = vec!["--help".into()];
    assert!(banner::is_root_help_args(&args));
    assert!(banner::is_root_help_args(&["-h".into()]));
    assert!(banner::is_root_help_args(&["help".into()]));
    // No-args is the banner surface; subcommand help stays logo-free.
    assert!(!banner::is_root_help_args(&[]));
    assert!(!banner::is_root_help_args(&[
        "help".into(),
        "claim-verify".into()
    ]));
    assert!(!banner::is_root_help_args(&["cc".into(), "--help".into()]));
}

#[test]
fn piped_bare_run_prints_no_banner_and_keeps_the_usage_contract() {
    let child = std::process::Command::new(crate::common::MOAI_BIN)
        .stdin(std::process::Stdio::null())
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .expect("spawn moai-cli");
    let out = child.wait_with_output().expect("wait for moai-cli");
    assert!(
        out.stdout.is_empty(),
        "piped bare run must not print the banner, got {:?}",
        String::from_utf8_lossy(&out.stdout)
    );
    assert!(
        String::from_utf8_lossy(&out.stderr).starts_with("usage:"),
        "the bare-run usage error contract must stay on stderr"
    );
    assert_eq!(out.status.code(), Some(1), "bare run keeps exit 1");
}

// ---- #2803 (design #2799 D2 §3): the terminal decides the banner's shape and depth ----

fn ui_env(pairs: &[(&str, &str)]) -> moai_cli::ui::Env {
    let mut e = moai_cli::ui::Env::default();
    for (k, v) in pairs {
        let v = Some(std::ffi::OsString::from(v));
        match *k {
            "NO_COLOR" => e.no_color = v,
            "TERM" => e.term = v,
            "COLORTERM" => e.colorterm = v,
            "LANG" => e.lang = v,
            other => panic!("unknown key {other}"),
        }
    }
    e
}

#[test]
fn the_shape_follows_tty_no_color_term_locale_and_width() {
    use banner::Shape;
    use moai_cli::ui::Depth;
    let wide = Some(banner::LOGO_WIDTH + 1);
    assert_eq!(banner::shape(false, &ui_env(&[]), wide), Shape::None);
    assert_eq!(
        banner::shape(false, &ui_env(&[("TERM", "dumb")]), wide),
        Shape::None
    );
    assert_eq!(
        banner::shape(true, &ui_env(&[("TERM", "xterm")]), wide),
        Shape::Colored(Depth::Ansi16)
    );
    assert_eq!(
        banner::shape(true, &ui_env(&[("TERM", "xterm-256color")]), wide),
        Shape::Colored(Depth::Ansi256)
    );
    assert_eq!(
        banner::shape(
            true,
            &ui_env(&[("TERM", "xterm-256color"), ("COLORTERM", "truecolor")]),
            wide
        ),
        Shape::Colored(Depth::TrueColor)
    );
    assert_eq!(
        banner::shape(true, &ui_env(&[("NO_COLOR", "1")]), wide),
        Shape::Plain
    );
    // TERM=dumb and a non-UTF-8 locale get the one ASCII line, NO_COLOR or not.
    assert_eq!(
        banner::shape(true, &ui_env(&[("TERM", "dumb")]), wide),
        Shape::Line
    );
    assert_eq!(
        banner::shape(true, &ui_env(&[("TERM", "dumb"), ("NO_COLOR", "1")]), wide),
        Shape::Line
    );
    assert_eq!(
        banner::shape(true, &ui_env(&[("LANG", "C")]), wide),
        Shape::Line
    );
    // Width: the logo needs LOGO_WIDTH + 1 columns (61 → line, 62 → logo); unknown → logo.
    assert_eq!(
        banner::shape(true, &ui_env(&[]), Some(banner::LOGO_WIDTH)),
        Shape::Line
    );
    assert_eq!(
        banner::shape(true, &ui_env(&[]), Some(banner::LOGO_WIDTH + 1)),
        Shape::Colored(Depth::Ansi16)
    );
    assert_eq!(
        banner::shape(true, &ui_env(&[]), None),
        Shape::Colored(Depth::Ansi16)
    );
}

#[test]
fn logo_width_is_the_widest_row() {
    let widest = banner::LOGO_ROWS
        .lines()
        .map(|r| r.chars().count())
        .max()
        .unwrap();
    assert_eq!(widest, banner::LOGO_WIDTH);
}

#[test]
fn truecolor_depth_keeps_the_recorded_bytes() {
    use moai_cli::ui::Depth;
    assert_eq!(
        banner::render_logo_in(Depth::TrueColor),
        banner::render_logo(true)
    );
    assert_eq!(
        banner::render_banner_in(Depth::TrueColor),
        banner::render_banner(true)
    );
    check("banner-color", &banner::render_banner_in(Depth::TrueColor));
}

#[test]
fn lower_depths_never_send_24_bit_sequences() {
    use moai_cli::ui::Depth;
    let c256 = banner::render_banner_in(Depth::Ansi256);
    assert!(c256.contains("\x1b[38;5;"), "{c256:?}");
    assert!(!c256.contains("38;2;"), "{c256:?}");
    // The 256-color stops are the status line's conversion of the same coral colors.
    let accent = moai_cli::statusline::color::Profile::Ansi256.foreground("#d97757");
    assert!(c256.contains(&format!("\x1b[{accent}m")), "{c256:?}");
    let c16 = banner::render_banner_in(Depth::Ansi16);
    assert!(!c16.contains("38;"), "{c16:?}");
    assert!(c16.contains("\x1b[35m"), "{c16:?}");
    for text in [&c256, &c16] {
        let plain: String = strip_ansi(text);
        assert_eq!(plain, banner::render_banner(false));
    }
}

#[test]
fn the_line_form_is_ascii_without_escapes() {
    let line = banner::render_line();
    assert!(line.is_ascii() && !line.contains('\x1b'), "{line:?}");
    assert!(line.starts_with("MoAI-ADK ") && line.ends_with('\n'));
    assert_eq!(line.lines().count(), 1);
}

fn strip_ansi(s: &str) -> String {
    let mut out = String::new();
    let mut chars = s.chars().peekable();
    while let Some(c) = chars.next() {
        if c == '\x1b' {
            for d in chars.by_ref() {
                if d == 'm' {
                    break;
                }
            }
        } else {
            out.push(c);
        }
    }
    out
}
