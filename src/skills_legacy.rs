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
use std::path::Path;

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

/// One root: every `moai-*` entry directly under it. Links are reported, never followed.
fn scan_root(home: &Path, rel: &str) -> Value {
    let root = home.join(rel);
    let meta = match std::fs::symlink_metadata(&root) {
        Ok(m) => m,
        Err(_) => return json!({"root": rel, "state": "absent", "legacy": []}),
    };
    if meta.file_type().is_symlink() {
        return json!({"root": rel, "state": "link_not_followed", "legacy": []});
    }
    if !meta.is_dir() {
        return json!({"root": rel, "state": "not_a_folder", "legacy": []});
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
            let skill_md = path.join("SKILL.md");
            let skill_name = (kind == "folder")
                .then(|| std::fs::read_to_string(&skill_md).ok())
                .flatten()
                .and_then(|t| frontmatter_name(&t));
            Some(json!({
                "folder": name,
                "path": path.to_string_lossy(),
                "kind": kind,
                "skill_md": kind == "folder" && skill_md.is_file(),
                "name": skill_name,
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
        .filter(|l| l["skill_md"] == true)
        .map(|l| {
            format!(
                "[[skills.config]]\npath = \"{}/SKILL.md\"\nenabled = false",
                l["path"].as_str().unwrap_or_default()
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
