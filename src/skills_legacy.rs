//! Legacy `moai-*` skills in the user's skill folders (Issue #2592, Codex parity).
//!
//! Codex sessions measured on 2026-11-13 rarely opened the `moai` skill: the list held 387
//! skills with the Go generation's `moai-*` skills beside it (#2592 comment 6848426992).
//! `moai-cli update` installs and owns only `moai`, so the older folders stay and compete for
//! the same requests. This read-only check lists them per host folder so the person can back
//! them up or turn them off; it never moves, deletes or edits anything.
//!
//! Codex reads user skills from `$HOME/.agents/skills` and can turn one off without deleting
//! it: `[[skills.config]]` with `path = "<…>/SKILL.md"` and `enabled = false` in
//! `~/.codex/config.toml` (https://learn.chatgpt.com/docs/build-skills, read 2026-11-13). The
//! check prints that entry for each folder Codex reads; it does not write the config.
use serde_json::{json, Value};
use std::io::Read;
use std::path::Path;

/// The largest `SKILL.md` the check reads for its `name:` (#2592 pre-audit 6852075845 B1).
pub const SKILL_MD_MAX: u64 = 64 * 1024;

/// The user-level skill folders of the supported hosts, relative to HOME, with who reads them.
pub const ROOTS: [(&str, &str); 5] = [
    (".agents/skills", "codex, cursor, omp"),
    (".claude/skills", "claude, cursor"),
    (".codex/skills", "cursor"),
    (".omp/agent/skills", "omp"),
    (".cursor/skills", "cursor"),
];

/// The skill the installer owns; never reported.
pub const OWNED: &str = "moai";

/// `name:` of the first frontmatter block, when there is one.
fn frontmatter_name(text: &str) -> Option<String> {
    let rest = text.strip_prefix("---\n")?;
    let end = rest.find("\n---")?;
    rest[..end].lines().find_map(|l| {
        l.strip_prefix("name:")
            .map(|v| v.trim().trim_matches(['"', '\'']).to_string())
    })
}

/// Opens a file for reading without following a symlink and without blocking on a FIFO.
fn open_nofollow(path: &Path) -> std::io::Result<std::fs::File> {
    let mut options = std::fs::OpenOptions::new();
    options.read(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK);
    }
    options.open(path)
}

/// What a folder's `SKILL.md` is, and its text when it is a regular file of at most
/// [`SKILL_MD_MAX`] bytes (#2592 pre-audit 6852075845 B1, B2). A link is reported and never
/// followed; a FIFO or any other kind is never read, so the check cannot block on it. The kind
/// is read again from the opened handle, so a swap after the first look changes nothing.
fn skill_md(path: &Path) -> (&'static str, Option<String>) {
    match std::fs::symlink_metadata(path) {
        Err(_) => return ("absent", None),
        Ok(m) if m.file_type().is_symlink() => return ("link", None),
        Ok(m) if !m.is_file() => return ("not_a_file", None),
        Ok(_) => {}
    }
    let Ok(file) = open_nofollow(path) else {
        return ("unreadable", None);
    };
    match file.metadata() {
        Ok(m) if !m.is_file() => return ("not_a_file", None),
        Ok(m) if m.len() > SKILL_MD_MAX => return ("too_large", None),
        Ok(_) => {}
        Err(_) => return ("unreadable", None),
    }
    let mut text = String::new();
    match file.take(SKILL_MD_MAX + 1).read_to_string(&mut text) {
        Ok(n) if n as u64 <= SKILL_MD_MAX => ("file", Some(text)),
        Ok(_) => ("too_large", None),
        Err(_) => ("unreadable", None),
    }
}

/// `s` as a TOML basic string, escaped (#2592 pre-audit N1: a `\` or `"` in a path, such as
/// a Windows path, must not break or redirect the pasted entry).
fn toml_string(s: &str) -> String {
    let mut out = String::from("\"");
    for c in s.chars() {
        match c {
            '\\' => out.push_str("\\\\"),
            '"' => out.push_str("\\\""),
            c if c.is_control() => out.push_str(&format!("\\u{:04X}", c as u32)),
            c => out.push(c),
        }
    }
    out.push('"');
    out
}

/// One root: every `moai-*` entry directly under it. Links are reported, never followed —
/// the root, every folder between HOME and it (#2592 pre-audit N2), the entries and each
/// entry's `SKILL.md`.
fn scan_root(home: &Path, rel: &str) -> Value {
    let root = home.join(rel);
    let mut walked = home.to_path_buf();
    for part in Path::new(rel).components() {
        walked.push(part);
        match std::fs::symlink_metadata(&walked) {
            Err(_) => return json!({"root": rel, "state": "absent", "legacy": []}),
            Ok(m) if m.file_type().is_symlink() => {
                return json!({"root": rel, "state": "link_not_followed", "legacy": []});
            }
            Ok(m) if !m.is_dir() => {
                return json!({"root": rel, "state": "not_a_folder", "legacy": []});
            }
            Ok(_) => {}
        }
    }
    let Ok(entries) = std::fs::read_dir(&root) else {
        return json!({"root": rel, "state": "unreadable", "legacy": []});
    };
    let mut legacy: Vec<Value> = entries
        .flatten()
        .filter_map(|e| {
            let name = e.file_name().to_string_lossy().to_string();
            if name == OWNED || !name.starts_with("moai-") {
                return None;
            }
            let path = e.path();
            let kind = match std::fs::symlink_metadata(&path) {
                Ok(m) if m.file_type().is_symlink() => "link",
                Ok(m) if m.is_dir() => "folder",
                Ok(_) => "file",
                Err(_) => "unreadable",
            };
            let (skill_md, text) = if kind == "folder" {
                skill_md(&path.join("SKILL.md"))
            } else {
                ("not_checked", None)
            };
            Some(json!({
                "folder": name,
                "path": path.to_string_lossy(),
                "kind": kind,
                "skill_md": skill_md,
                "name": text.as_deref().and_then(frontmatter_name),
            }))
        })
        .collect();
    legacy.sort_by(|a, b| a["folder"].as_str().cmp(&b["folder"].as_str()));
    json!({"root": rel, "state": "present", "legacy": legacy})
}

/// `skills-legacy`: the legacy `moai-*` skill folders under HOME's host skill folders, with
/// the Codex `[[skills.config]]` entry that turns each Codex-read one off. Reads only.
pub fn check(home: &Path) -> Value {
    let roots: Vec<Value> = ROOTS
        .iter()
        .map(|(rel, readers)| {
            let mut r = scan_root(home, rel);
            r["read_by"] = json!(readers);
            r
        })
        .collect();
    let count: usize = roots
        .iter()
        .map(|r| r["legacy"].as_array().map_or(0, Vec::len))
        .sum();
    let codex_off: Vec<String> = roots
        .iter()
        .filter(|r| r["root"] == ".agents/skills")
        .flat_map(|r| r["legacy"].as_array().cloned().unwrap_or_default())
        .filter(|l| l["skill_md"] == "file")
        .map(|l| {
            format!(
                "[[skills.config]]\npath = {}\nenabled = false",
                toml_string(&format!(
                    "{}/SKILL.md",
                    l["path"].as_str().unwrap_or_default()
                ))
            )
        })
        .collect();
    let summary = if count == 0 {
        "0건: no legacy moai-* skill folder under the host skill folders".to_string()
    } else {
        format!(
            "{count}건: legacy moai-* skill folders found; back them up or turn them off — nothing was changed"
        )
    };
    json!({
        "home": home.to_string_lossy(),
        "count": count,
        "summary": summary,
        "roots": roots,
        "codex_disable_entries": codex_off,
        "writes": false,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn frontmatter_name_reads_the_first_block_only() {
        assert_eq!(
            frontmatter_name("---\nname: moai-plan\ndescription: x\n---\nname: other\n"),
            Some("moai-plan".into())
        );
        assert_eq!(
            frontmatter_name("---\nname: \"q\"\n---\n"),
            Some("q".into())
        );
        assert_eq!(frontmatter_name("no frontmatter\nname: x\n"), None);
    }
}
