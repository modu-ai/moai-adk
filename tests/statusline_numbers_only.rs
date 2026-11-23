//! The statusline limits line (SPEC-20261121-statusline-detached, whose "numbers only, no bar"
//! decision SPEC-20261123-statusline-order-bars REQ-3 replaces): the renderer draws the ten-cell
//! bar, carries no card numbers in `render.rs` comments and no color-gradient leftovers, and the
//! product document describes the bars and the line order. The files are read with
//! `include_str!`.

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
fn the_renderer_draws_the_bar_and_carries_no_color_gradient_and_no_card_numbers() {
    // The bar is drawn by the limits line of the renderer: a ten-cell width and the two glyphs.
    for needle in ["const BAR_WIDTH: i64 = 10;", "'█'", "'░'"] {
        assert!(
            RENDER.contains(needle),
            "render.rs does not hold `{needle}`"
        );
    }
    // The segment readers draw nothing, and neither file brings the color gradient back: a bar
    // has the base color for its filled cells and the muted color for the empty ones (D-3).
    for mark in ['█', '░'] {
        assert!(!SEGMENTS.contains(mark), "segments.rs holds `{mark}`");
    }
    for (name, text) in [("render.rs", RENDER), ("segments.rs", SEGMENTS)] {
        for phrase in ["gradient", "CTX bar", "context bar", "stops"] {
            assert!(!text.contains(phrase), "{name} still holds `{phrase}`");
        }
    }
    let numbers = card_numbers(RENDER);
    assert!(
        numbers.is_empty(),
        "render.rs names a card number: {numbers:?}"
    );
}

/// The part of the product document that describes line 3.
fn line_three_section() -> &'static str {
    let from = DOC.find("3. Usage windows:").expect("the line 3 item");
    let rest = &DOC[from..];
    &rest[..rest.find("shortens nothing").expect("the end of the item")]
}

#[test]
fn the_statusline_doc_describes_the_bars_and_the_line_order() {
    let section = line_three_section();
    for stale in [
        "CTX:",
        "Week:",
        "Week ",
        "CTX",
        "(⚠️ /clear)",
        "(🛑/clear!)",
        "미확인",
        "five-hour bar",
        "numbers only",
        "no graph",
    ] {
        assert!(!section.contains(stale), "the doc still holds `{stale}`");
    }
    for wanted in [
        "ten-cell bar",
        "`█`",
        "`░`",
        "resets in",
        "n/a",
        "no `/clear` hint",
        "7D",
        "CW",
    ] {
        assert!(section.contains(wanted), "the doc does not say `{wanted}`");
    }
    // The second line opens with the two versions and closes with my card.
    let statusline = statusline_section();
    for wanted in [
        "`🤖 <Claude Code version> | 🗿 <MoAI version> | ",
        "🎯 C-<n> <stage icon>`",
    ] {
        assert!(
            statusline.contains(wanted),
            "the doc does not say `{wanted}`"
        );
    }
    // The first line (SPEC-20261123-statusline-order-bars REQ-8 to REQ-12): the integration item
    // with the tip of its local ref, the checked-out branch, no distance and no marks.
    for wanted in [
        "`🅱️ <integration>(🏷️ <tip>)`",
        "never from `HEAD`",
        "`(🏷️ n/a)`",
        "`🌳 <branch>`",
        "no `🧭` item",
        "no ahead, behind or uncommitted mark",
    ] {
        assert!(
            statusline.contains(wanted),
            "the doc does not say `{wanted}`"
        );
    }
    for stale in [
        "`🌳 <worktree name>`",
        "branch with its ahead, behind and dirty counts",
        "The branch shows `↑N`",
        "shows the short SHA in the branch slot",
        "The branch's `+N` is the only change",
    ] {
        assert!(!statusline.contains(stale), "the doc still holds `{stale}`");
    }
}

/// The sibling test files pin the bar: the limits-line forms of the cards tests hold the glyphs of
/// the bar, and the renderer's own tests carry the named acceptance tests of
/// SPEC-20261123-statusline-order-bars.
#[test]
fn the_statusline_tests_pin_the_bar() {
    let cards = SIBLING_TESTS
        .iter()
        .find(|(name, _)| *name == "statusline_cards.rs")
        .map(|(_, text)| *text)
        .unwrap();
    for mark in ['█', '░'] {
        assert!(cards.contains(mark), "statusline_cards.rs lacks `{mark}`");
    }
    for name in [
        "fn the_bar_has_ten_cells_and_changes_at_each_tenth",
        "fn the_row_shows_what_is_left_of_each_window",
        "fn unknown_windows_and_the_hide_list_stay_as_they_were",
        "fn the_colored_bar_has_the_plain_text",
        "fn the_versions_open_the_flow_line_and_my_card_closes_it",
        "fn every_form_is_exactly_its_string_and_changes_at_its_boundary",
    ] {
        assert!(RENDER.contains(name), "render.rs lacks `{name}`");
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
    assert!(
        out.contains("🔋 CW ██████████ 100%"),
        "a real zero gave: {out}"
    );
}

/// The part of the product document about the status line: from its heading to the next one.
fn statusline_section() -> &'static str {
    let from = DOC
        .find("\n## Claude Code status line")
        .expect("the status line section")
        + 1;
    let rest = &DOC[from..];
    let end = rest[3..].find("\n## ").map_or(rest.len(), |at| at + 3);
    &rest[..end]
}

/// SPEC-20261122-statusline-where AC-7 (REQ-12): the document says that the status line starts
/// git and a detached `gh` refresh and writes a cache under `$TMPDIR`, and no longer says that
/// rendering is read-only and free of the network.
#[test]
fn the_statusline_doc_describes_the_cache() {
    let section = statusline_section();
    for stale in [
        "Rendering remains read-only and network-free",
        "read-only and network-free",
        "network-free",
    ] {
        assert!(!section.contains(stale), "the doc still says `{stale}`");
    }
    for wanted in [
        "starts `git`",
        "detached `gh` refresh",
        "cache under `$TMPDIR`",
        "moai-statusline/<session_id>.json",
        "ten seconds",
    ] {
        assert!(section.contains(wanted), "the doc does not say `{wanted}`");
    }
}
