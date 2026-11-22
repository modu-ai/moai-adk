//! The statusline repair after the audit (SPEC-20261121-statusline-detached): the renderer
//! sources carry no bar-graph leftovers and no card numbers in `render.rs` comments, and the
//! product document describes the numbers-only look. The files are read with `include_str!`.

mod common;

use common::home::TempHome;
use std::io::Write;
use std::process::Stdio;

/// The other statusline test files, read as text (this file is left out: it must name the
/// words it forbids).
const SIBLING_TESTS: [(&str, &str); 6] = [
    ("statusline_cards.rs", include_str!("statusline_cards.rs")),
    (
        "statusline_counts_refresh.rs",
        include_str!("statusline_counts_refresh.rs"),
    ),
    (
        "statusline_git_env.rs",
        include_str!("statusline_git_env.rs"),
    ),
    (
        "statusline_home_state.rs",
        include_str!("statusline_home_state.rs"),
    ),
    (
        "statusline_tag_order.rs",
        include_str!("statusline_tag_order.rs"),
    ),
    (
        "statusline_worktree_label.rs",
        include_str!("statusline_worktree_label.rs"),
    ),
];

const RENDER: &str = include_str!("../src/statusline/render.rs");
const SEGMENTS: &str = include_str!("../src/statusline/segments.rs");
const DOC: &str = include_str!("../docs/rust-product-path.md");

/// Whole-word occurrences of `word` in `text`, case-insensitive.
fn words(text: &str, word: &str) -> usize {
    text.split(|c: char| !c.is_alphanumeric())
        .filter(|w| w.eq_ignore_ascii_case(word))
        .count()
}

/// A card number in the shape `C-<digits>` or `card <digits>` / `card #<digits>`.
fn card_numbers(text: &str) -> Vec<String> {
    let mut found = Vec::new();
    let bytes: Vec<char> = text.chars().collect();
    for i in 0..bytes.len() {
        let before_ok = i == 0 || !bytes[i - 1].is_alphanumeric();
        if bytes[i] == 'C' && before_ok && bytes.get(i + 1) == Some(&'-') {
            let digits: String = bytes[i + 2..]
                .iter()
                .take_while(|c| c.is_ascii_digit())
                .collect();
            if !digits.is_empty() {
                found.push(format!("C-{digits}"));
            }
        }
    }
    for (line_no, line) in text.lines().enumerate() {
        let lower = line.to_ascii_lowercase();
        for (at, _) in lower.match_indices("card ") {
            let rest = lower[at + 5..].trim_start_matches('#');
            if rest.chars().next().is_some_and(|c| c.is_ascii_digit()) {
                found.push(format!("card <n> on line {}", line_no + 1));
            }
        }
    }
    found
}

#[test]
fn the_renderer_sources_carry_no_bar_leftovers_and_no_card_numbers() {
    for (name, text) in [("render.rs", RENDER), ("segments.rs", SEGMENTS)] {
        for mark in ['█', '░'] {
            assert!(!text.contains(mark), "{name} still holds `{mark}`");
        }
        for phrase in ["gradient", "10-cell", "CTX bar", "context bar", "stops"] {
            assert!(!text.contains(phrase), "{name} still holds `{phrase}`");
        }
        assert_eq!(
            words(text, "bar") + words(text, "bars"),
            0,
            "{name} still names a bar"
        );
    }
    let numbers = card_numbers(RENDER);
    assert!(
        numbers.is_empty(),
        "render.rs names a card number: {numbers:?}"
    );
}

/// The part of the product document that describes line 3.
fn line_three_section() -> &'static str {
    let from = DOC
        .find("3. Usage windows and card counts:")
        .expect("the line 3 item");
    let rest = &DOC[from..];
    &rest[..rest.find("shortens nothing").expect("the end of the item")]
}

#[test]
fn the_statusline_doc_describes_the_numbers_only_look() {
    let section = line_three_section();
    for mark in ['█', '░'] {
        assert!(!section.contains(mark), "the doc still draws `{mark}`");
    }
    for stale in [
        "CTX:",
        "Week:",
        "Week ",
        "CTX",
        "(⚠️ /clear)",
        "(🛑/clear!)",
        "미확인",
        "five-hour bar",
    ] {
        assert!(!section.contains(stale), "the doc still holds `{stale}`");
    }
    assert_eq!(
        words(section, "bar") + words(section, "bars"),
        0,
        "the doc still names a bar"
    );
    for wanted in [
        "numbers only",
        "resets in",
        "n/a",
        "no `/clear` hint",
        "7D",
        "CW",
    ] {
        assert!(section.contains(wanted), "the doc does not say `{wanted}`");
    }
}

#[test]
fn the_statusline_test_files_carry_no_bar_wording() {
    for (name, text) in SIBLING_TESTS {
        assert_eq!(
            words(text, "bar") + words(text, "bars"),
            0,
            "{name} still names a bar"
        );
    }
}

/// One plain render as Claude Code runs it, from an empty HOME: the session's JSON on stdin
/// with the given `context_window`. The exit code must be 0. Returns stdout.
fn render_plain(context_window: serde_json::Value) -> String {
    let home = TempHome::new();
    let project = home.path().join("project");
    std::fs::create_dir_all(&project).unwrap();
    let stdin = serde_json::json!({
        "cwd": project,
        "model": {"display_name": "Opus"},
        "workspace": {"current_dir": project, "project_dir": project},
        "context_window": context_window,
    })
    .to_string();
    let mut cmd = home.command(&["statusline", "--render-budget-ms", "3600000"]);
    cmd.current_dir(&project)
        .env("PATH", std::env::var_os("PATH").unwrap_or_default())
        .env("NO_COLOR", "1")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());
    let mut child = cmd.spawn().unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(stdin.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    assert_eq!(
        out.status.code(),
        Some(0),
        "{}",
        String::from_utf8_lossy(&out.stderr)
    );
    String::from_utf8(out.stdout).unwrap()
}

#[test]
fn a_context_window_with_a_null_percentage_says_ctx_n_a() {
    let unknown = [
        serde_json::json!({"context_window_size": 200000, "used_percentage": null}),
        serde_json::json!({"context_window_size": 200000}),
        serde_json::json!({}),
    ];
    for window in unknown {
        let out = render_plain(window.clone());
        let line = out
            .lines()
            .find(|l| l.contains("CW"))
            .unwrap_or_else(|| panic!("no CW in the render of {window}: {out}"));
        assert!(line.contains("CW n/a"), "{window} gave `{line}`");
        assert!(!line.contains("0%"), "{window} gave `{line}`");
        assert!(
            !line.contains("🔋") && !line.contains("🪫"),
            "{window} gave an icon: `{line}`"
        );
    }
    // The edge that stays a number: a percentage of zero is a real reading.
    let out = render_plain(serde_json::json!({
        "context_window_size": 200000,
        "used_percentage": 0
    }));
    assert!(out.contains("🔋 CW 0%"), "a real zero gave: {out}");
}
