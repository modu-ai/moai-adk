//! Go-generation settings offered as setup defaults (#2817).
//!
//! After the Go `.moai/` was kept whole as `<project>/.moai-backup` (#2696 S2c), its
//! `config/sections/language.yaml` and `project.yaml` still say which languages and which
//! name the project used. Their format differs from this generation's profile, so nothing is
//! copied: the values are read, checked against this generation's own rules and offered as
//! defaults in the setup input step, which the user confirms like any other answer.
//!
//! The backup is untrusted input and is only read (leader decisions D1–D3 on #2817): only
//! those two files, every path element refused when it is a link, at most `FILE_LIMIT` bytes,
//! only the plain two-level `section:` / `  key: value` shape, unknown keys ignored, values
//! accepted only from `crate::language::CODES` (languages) or as a portable id (name).
//! Anything else gives no default — never an error.

use crate::language::LanguageSettings;
use serde_json::{json, Value};
use std::fs;
use std::io::Read as _;
use std::path::{Path, PathBuf};

/// The largest Go settings file read (bytes); a larger one gives no defaults.
pub const FILE_LIMIT: u64 = 64 * 1024;

/// Go `language.yaml` keys and the item each one supplies; `conversation_language` is the
/// conversation language instead (see [`GoDefaults::conversation`]).
const LANGUAGE_KEYS: &[(&str, &str)] = &[
    ("git_commit_messages", "commit_messages"),
    ("code_comments", "code_comments"),
    ("agent_prompt_language", "agent_messages"),
];

/// Go keys that have no counterpart here; listed so the user sees they were not carried over.
const NOT_MOVED: &[&str] = &[
    "language.conversation_language_name",
    "language.documentation",
    "language.error_messages",
    "project.description",
    "project.created_at",
    "project.initialized",
    "project.optimized",
    "project.template_version",
    "project.mode",
];

/// What the backup offers. Empty when there is no backup or nothing in it passed the checks.
#[derive(Debug, Default, Clone, PartialEq)]
pub struct GoDefaults {
    /// Whether `<project>/.moai-backup/config/sections` is a real folder (not a link).
    pub found: bool,
    /// The Go conversation language, when it is an accepted code.
    pub conversation: Option<String>,
    /// Item defaults taken from the other Go language keys.
    pub items: LanguageSettings,
    /// The Go project name, when it already is a portable id.
    pub project_id: Option<String>,
}

impl GoDefaults {
    /// The `go_backup` field `language-setup` adds when the backup exists.
    pub fn to_json(&self) -> Value {
        json!({
            "source": format!("{}/config/sections", crate::state_root::GO_BACKUP),
            "conversation_language": self.conversation,
            "items": self.items,
            "project_id": self.project_id,
            "not_moved": NOT_MOVED,
            "copied": false,
        })
    }
}

/// `<project>/.moai-backup/config/sections` when every element below the project is a real
/// folder; `None` when one is missing or a link.
fn sections_dir(project: &Path) -> Option<PathBuf> {
    let mut dir = project.to_path_buf();
    for part in [crate::state_root::GO_BACKUP, "config", "sections"] {
        dir.push(part);
        if !fs::symlink_metadata(&dir).ok()?.file_type().is_dir() {
            return None;
        }
    }
    Some(dir)
}

/// A regular file of at most `FILE_LIMIT` bytes, read without following a link.
fn read_bounded(path: &Path) -> Option<String> {
    // The size is judged on the bytes read (at most one past the limit), so a file that
    // grows between the check and the read cannot slip through.
    if !fs::symlink_metadata(path).ok()?.file_type().is_file() {
        return None;
    }
    let mut bytes = Vec::new();
    fs::File::open(path)
        .ok()?
        .take(FILE_LIMIT + 1)
        .read_to_end(&mut bytes)
        .ok()?;
    if bytes.len() as u64 > FILE_LIMIT {
        return None;
    }
    String::from_utf8(bytes).ok()
}

/// One scalar value: unquoted, or in matching single or double quotes (escapes are not decoded).
fn scalar(raw: &str) -> Option<String> {
    let v = raw.trim();
    for q in ['"', '\''] {
        if let Some(rest) = v.strip_prefix(q) {
            // The closing quote, then nothing or a ` #` comment.
            let (inner, after) = rest.split_once(q)?;
            let after = after.trim_start();
            // No escape handling: a value holding `\` is not an accepted code or id anyway.
            if !(after.is_empty() || after.starts_with('#')) {
                return None;
            }
            return Some(inner.to_string());
        }
    }
    // An unquoted value ends at a ` #` comment; anything that starts a YAML structure is refused.
    let v = v.split(" #").next().unwrap_or("").trim();
    if v.starts_with(['[', '{', '&', '*', '!', '|', '>', '-', '"', '\'']) {
        return None;
    }
    Some(v.to_string())
}

/// The `key: value` pairs under the top-level `section:` of a plain two-level YAML file.
/// Any line outside that shape (a list, a nested map, a document marker, tabs, a second
/// top-level key holding the section again) refuses the whole file: `None`.
fn section_pairs(text: &str, section: &str) -> Option<Vec<(String, String)>> {
    let mut pairs = Vec::new();
    let mut inside = false;
    let mut seen = false;
    for line in text.lines() {
        if line.contains('\t') {
            return None;
        }
        let trimmed = line.trim();
        if trimmed.is_empty() || trimmed.starts_with('#') {
            continue;
        }
        if !line.starts_with(' ') {
            // A top-level key: only `name:` with nothing after it opens a section.
            let (key, rest) = trimmed.split_once(':')?;
            if !rest.trim().is_empty() && !rest.trim().starts_with('#') {
                return None;
            }
            inside = key == section;
            if inside {
                if seen {
                    return None;
                }
                seen = true;
            }
            continue;
        }
        if !inside {
            continue;
        }
        let indent = line.len() - line.trim_start().len();
        if indent != 2 {
            return None;
        }
        let (key, raw) = trimmed.split_once(':')?;
        if key.is_empty() || !key.bytes().all(|b| b.is_ascii_lowercase() || b == b'_') {
            return None;
        }
        pairs.push((key.to_string(), scalar(raw)?));
    }
    seen.then_some(pairs)
}

fn pairs_of(dir: &Path, file: &str, section: &str) -> Vec<(String, String)> {
    read_bounded(&dir.join(file))
        .and_then(|t| section_pairs(&t, section))
        .unwrap_or_default()
}

/// Reads the Go defaults of `project`. Never writes; never fails — a missing, linked,
/// oversized or malformed file, an unknown key or an unaccepted value simply offers nothing.
pub fn read(project: &Path) -> GoDefaults {
    let Some(dir) = sections_dir(project) else {
        return GoDefaults::default();
    };
    let mut out = GoDefaults {
        found: true,
        ..GoDefaults::default()
    };
    for (key, value) in pairs_of(&dir, "language.yaml", "language") {
        if !crate::language::is_code(&value) {
            continue;
        }
        if key == "conversation_language" {
            out.conversation = Some(value);
        } else if let Some((_, item)) = LANGUAGE_KEYS.iter().find(|(k, _)| *k == key) {
            out.items.insert((*item).to_string(), value);
        }
    }
    out.project_id = pairs_of(&dir, "project.yaml", "project")
        .into_iter()
        .find(|(k, _)| k == "name")
        .map(|(_, v)| v)
        .filter(|v| crate::profile::portable_id(v));
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_go_template_shape_is_read() {
        let text = "# Language\nlanguage:\n  # comment\n  conversation_language: \"ko\"\n  git_commit_messages: en\n  code_comments: 'ja' # note\n";
        assert_eq!(
            section_pairs(text, "language").unwrap(),
            vec![
                ("conversation_language".into(), "ko".into()),
                ("git_commit_messages".into(), "en".into()),
                ("code_comments".into(), "ja".into()),
            ]
        );
    }

    #[test]
    fn other_shapes_refuse_the_whole_file() {
        for text in [
            "language:\n  code_comments:\n    - en\n",
            "language:\n    code_comments: en\n",
            "language:\n\tcode_comments: en\n",
            "language: {code_comments: en}\n",
            "language:\n  code_comments: [en]\n",
            "language:\n  code_comments: &a en\n",
            "language:\n  code_comments: \"en\\\"\"\n",
            "language:\n  a: en\nlanguage:\n  b: ko\n",
            "other:\n  code_comments: en\n",
            "language:\n  Code_Comments: en\n",
        ] {
            assert_eq!(section_pairs(text, "language"), None, "{text:?}");
        }
    }
}
