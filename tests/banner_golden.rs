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
