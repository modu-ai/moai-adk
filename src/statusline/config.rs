//! A deliberately small reader for the block-style YAML mappings the statusline reads
//! from the configuration (`statusline.yaml` and `statusline.local.yaml` under `.moai-cli/`
//! first, else the Go files — #2350 scope 3; Go-only `system.yaml`,
//! `llm.yaml`). The crate carries no YAML dependency and an offline build cannot add one.
//!
//! Supported: nested `key: value` / `key:` mappings by indentation, `#` comments, quoted
//! scalars. Anything else (flow collections, sequences, tabs, anchors, multi-line
//! scalars) makes the whole file unreadable, which every caller treats the way Go treats
//! an unparseable file: as absent.

use std::collections::BTreeMap;
use std::path::Path;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Node {
    Scalar(String),
    Map(BTreeMap<String, Node>),
}

impl Node {
    /// The node at a dotted key path.
    pub fn get(&self, path: &[&str]) -> Option<&Node> {
        let mut node = self;
        for key in path {
            match node {
                Node::Map(map) => node = map.get(*key)?,
                Node::Scalar(_) => return None,
            }
        }
        Some(node)
    }

    pub fn scalar(&self, path: &[&str]) -> Option<&str> {
        match self.get(path)? {
            Node::Scalar(s) => Some(s),
            Node::Map(_) => None,
        }
    }
}

/// Reads and parses a file; `None` when it is absent, unreadable or outside the subset.
pub fn read(path: &Path) -> Option<Node> {
    parse(&std::fs::read_to_string(path).ok()?)
}

pub fn parse(text: &str) -> Option<Node> {
    let mut lines = Vec::new();
    for raw in text.lines() {
        let line = strip_comment(raw);
        if line.trim().is_empty() || line.trim() == "---" {
            continue;
        }
        if line.contains('\t') {
            return None;
        }
        let indent = line.len() - line.trim_start().len();
        lines.push((indent, line.trim().to_string()));
    }
    let mut pos = 0;
    let root = block(&lines, &mut pos, 0)?;
    (pos == lines.len()).then_some(Node::Map(root))
}

fn block(
    lines: &[(usize, String)],
    pos: &mut usize,
    indent: usize,
) -> Option<BTreeMap<String, Node>> {
    let mut map = BTreeMap::new();
    while let Some((line_indent, text)) = lines.get(*pos) {
        if *line_indent < indent {
            break;
        }
        if *line_indent > indent {
            return None;
        }
        let (key, value) = split_key(text)?;
        *pos += 1;
        let node = if value.is_empty() {
            match lines.get(*pos) {
                Some((next, _)) if *next > indent => Node::Map(block(lines, pos, *next)?),
                _ => Node::Scalar(String::new()),
            }
        } else {
            Node::Scalar(scalar(value)?)
        };
        // A duplicate key is an error in YAML.
        if map.insert(key, node).is_some() {
            return None;
        }
    }
    Some(map)
}

fn split_key(text: &str) -> Option<(String, &str)> {
    if text.starts_with(['-', '[', '{', '&', '*', '!', '|', '>', '?']) {
        return None;
    }
    let idx = text
        .find(": ")
        .or_else(|| text.ends_with(':').then(|| text.len() - 1))?;
    let key = scalar(text[..idx].trim())?;
    Some((key, text[idx + 1..].trim()))
}

fn scalar(value: &str) -> Option<String> {
    if let Some(inner) = value.strip_prefix('"') {
        return inner
            .strip_suffix('"')
            .map(|s| s.replace("\\\"", "\"").replace("\\\\", "\\"));
    }
    if let Some(inner) = value.strip_prefix('\'') {
        return inner.strip_suffix('\'').map(|s| s.replace("''", "'"));
    }
    if value.starts_with(['[', '{', '&', '*', '!', '|', '>']) {
        return None;
    }
    Some(value.to_string())
}

/// Drops a `#` comment that starts a line or follows whitespace outside quotes.
fn strip_comment(line: &str) -> &str {
    let mut quote: Option<char> = None;
    let mut prev_space = true;
    for (i, c) in line.char_indices() {
        match quote {
            Some(q) if c == q => quote = None,
            Some(_) => {}
            None if c == '"' || c == '\'' => quote = Some(c),
            None if c == '#' && prev_space => return &line[..i],
            None => {}
        }
        prev_space = c == ' ';
    }
    line
}

/// A YAML boolean as yaml.v3 decodes it into a Go `bool` (YAML 1.1 words included).
pub fn boolean(value: &str) -> Option<bool> {
    match value {
        "true" | "True" | "TRUE" | "yes" | "Yes" | "YES" | "on" | "On" | "ON" | "y" | "Y" => {
            Some(true)
        }
        "false" | "False" | "FALSE" | "no" | "No" | "NO" | "off" | "Off" | "OFF" | "n" | "N" => {
            Some(false)
        }
        _ => None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn nested_mapping_with_comments_and_quotes() {
        let n = parse(
            "# header\nstatusline:\n  theme: \"catppuccin-latte\"  # trailing\n  segments:\n    model: true\n    pr: false\n  forge: 'none'\n",
        )
        .unwrap();
        assert_eq!(n.scalar(&["statusline", "theme"]), Some("catppuccin-latte"));
        assert_eq!(n.scalar(&["statusline", "segments", "pr"]), Some("false"));
        assert_eq!(n.scalar(&["statusline", "forge"]), Some("none"));
        assert!(n.scalar(&["statusline", "segments"]).is_none());
    }

    #[test]
    fn unsupported_shapes_are_unreadable() {
        assert!(parse("statusline:\n  segments: {model: true}\n").is_none());
        assert!(parse("statusline:\n  - model\n").is_none());
        assert!(parse("a: 1\na: 2\n").is_none());
        assert!(parse("a:\n\tb: 1\n").is_none());
        assert!(parse("a:\n    b: 1\n  c: 2\n").is_none());
    }

    #[test]
    fn booleans_follow_yaml_v3_into_bool() {
        assert_eq!(boolean("TRUE"), Some(true));
        assert_eq!(boolean("off"), Some(false));
        assert_eq!(boolean("1"), None);
    }
}
