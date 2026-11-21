//! Reading what a command wrote to a terminal: the text a person sees, without its colors.

#![allow(dead_code)]

/// `text` without its `ESC [ <digits and ;> m` color sequences. Any other escape stays, so a
/// test that forbids escapes still sees them.
pub fn strip_colors(text: &str) -> String {
    let bytes = text.as_bytes();
    let mut out = Vec::with_capacity(bytes.len());
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] == 0x1b && bytes.get(i + 1) == Some(&b'[') {
            let mut j = i + 2;
            while j < bytes.len() && (bytes[j].is_ascii_digit() || bytes[j] == b';') {
                j += 1;
            }
            if bytes.get(j) == Some(&b'm') {
                i = j + 1;
                continue;
            }
        }
        out.push(bytes[i]);
        i += 1;
    }
    String::from_utf8(out).expect("only whole ASCII sequences were removed")
}

/// Whether some line of `text` has `words` as consecutive whitespace-separated words, so a table
/// row is found whatever the column padding is.
pub fn has_row(text: &str, words: &[&str]) -> bool {
    text.lines().any(|line| {
        let cells: Vec<&str> = line.split_whitespace().collect();
        cells.windows(words.len()).any(|w| w == words)
    })
}
