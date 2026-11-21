//! The statusline repair after the audit (SPEC-20261121-statusline-detached): the renderer
//! sources carry no bar-graph leftovers and no card numbers in `render.rs` comments, and the
//! product document describes the numbers-only look. The files are read with `include_str!`.

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
    for wanted in ["numbers only", "resets in", "n/a", "no `/clear` hint"] {
        assert!(section.contains(wanted), "the doc does not say `{wanted}`");
    }
}
