//! Opt-in hook rules from `~/.moai-cli/config` (Issue #2391 stage 7; user decision, Issue #2391
//! comment 6779829826, design 6777574227). Both rules are off unless the user turns them on:
//!
//! ```toml
//! [hooks.outside_project_writes]   # B9
//! enabled = true
//! allow = ["~/.claude/", "~/.codex/", "~/.moai-cli/", "/tmp/", "/private/tmp/"]
//!
//! [hooks.primary_checkout]         # B3
//! enabled = true
//! repos = ["/Users/me/src/moai-adk"]
//! allow_commands = ["git switch v2026.11.0"]
//! ```
//!
//! A config that is absent, not a regular file (a symlink too), over 64 KiB or not valid TOML
//! turns both rules off: a broken config must not block every write (the engine's fail-open
//! rule, D1). A rule whose own table is malformed (a non-boolean `enabled`, a non-string list
//! entry) is off as well.

use std::fs;
use std::path::{Component, Path, PathBuf};

/// Largest config file read.
pub const CONFIG_MAX: u64 = 65_536;

/// B9: writes outside the project are refused unless under one of `allow`.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct OutsideRule {
    /// Absolute, normalized folder prefixes (`~` expanded).
    pub allow: Vec<PathBuf>,
}

/// B3: switching, resetting, stashing or rebasing a shared primary checkout is refused.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PrimaryRule {
    /// Primary worktree folders the rule covers (absolute, normalized, `~` expanded).
    pub repos: Vec<PathBuf>,
    /// Exact command lines that are let through.
    pub allow_commands: Vec<String>,
}

#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct HookConfig {
    pub outside: Option<OutsideRule>,
    pub primary: Option<PrimaryRule>,
}

/// The config of `HOME` (both rules off when `HOME` is unset or the file is unusable).
pub fn load() -> HookConfig {
    let Some(home) = std::env::var_os("HOME").filter(|h| !h.is_empty()) else {
        return HookConfig::default();
    };
    let home = PathBuf::from(home);
    let path = home.join(crate::HOME_STATE_DIR).join("config");
    match fs::symlink_metadata(&path) {
        Ok(m) if m.is_file() && m.len() <= CONFIG_MAX => {}
        _ => return HookConfig::default(),
    }
    match fs::read_to_string(&path) {
        Ok(text) => parse(&text, &home),
        Err(_) => HookConfig::default(),
    }
}

/// Parses the config text; `home` expands a leading `~`.
pub fn parse(text: &str, home: &Path) -> HookConfig {
    let Ok(doc) = text.parse::<toml_edit::DocumentMut>() else {
        return HookConfig::default();
    };
    let hooks = doc.get("hooks").and_then(|h| h.as_table_like());
    let table = |name: &str| {
        hooks
            .and_then(|h| h.get(name))
            .and_then(|t| t.as_table_like())
    };
    let enabled =
        |t: &dyn toml_edit::TableLike| t.get("enabled").and_then(|e| e.as_bool()) == Some(true);
    let strings = |t: &dyn toml_edit::TableLike, key: &str| -> Option<Vec<String>> {
        match t.get(key) {
            None => Some(Vec::new()),
            Some(item) => item
                .as_array()?
                .iter()
                .map(|v| v.as_str().map(str::to_string))
                .collect(),
        }
    };
    let outside = table("outside_project_writes").and_then(|t| {
        if !enabled(t) {
            return None;
        }
        let allow = strings(t, "allow")?;
        Some(OutsideRule {
            allow: allow.iter().filter_map(|a| absolute(a, home)).collect(),
        })
    });
    let primary = table("primary_checkout").and_then(|t| {
        if !enabled(t) {
            return None;
        }
        Some(PrimaryRule {
            repos: strings(t, "repos")?
                .iter()
                .filter_map(|r| absolute(r, home))
                .map(|p| canonical(&p))
                .collect(),
            allow_commands: strings(t, "allow_commands")?
                .into_iter()
                .map(|c| c.trim().to_string())
                .collect(),
        })
    });
    HookConfig { outside, primary }
}

/// `~` or `~/…` expanded against `home`; a relative path is ignored (None).
fn absolute(text: &str, home: &Path) -> Option<PathBuf> {
    let text = text.trim();
    let path = if text == "~" {
        home.to_path_buf()
    } else if let Some(rest) = text.strip_prefix("~/") {
        home.join(rest)
    } else {
        PathBuf::from(text)
    };
    path.is_absolute().then(|| normalize(&path))
}

/// `.` dropped and `..` applied lexically (never above the root).
pub fn normalize(path: &Path) -> PathBuf {
    let mut out = PathBuf::new();
    for c in path.components() {
        match c {
            Component::CurDir => {}
            Component::ParentDir => {
                out.pop();
            }
            other => out.push(other.as_os_str()),
        }
    }
    out
}

/// The path with its longest existing prefix resolved through symlinks, then normalized, so a
/// symlink out of a folder and a `..` in a part that does not exist yet are both seen.
pub fn canonical(path: &Path) -> PathBuf {
    let path = normalize(path);
    let mut existing = path.clone();
    let mut tail = Vec::new();
    while fs::symlink_metadata(&existing).is_err() {
        match (
            existing.file_name().map(|n| n.to_os_string()),
            existing.parent(),
        ) {
            (Some(name), Some(parent)) => {
                tail.push(name);
                existing = parent.to_path_buf();
            }
            _ => return path,
        }
    }
    let mut out = fs::canonicalize(&existing).unwrap_or(existing);
    for name in tail.into_iter().rev() {
        out.push(name);
    }
    // The tail was normalized above and canonicalize leaves no `..`.
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rules_are_off_unless_enabled_and_well_formed() {
        let home = Path::new("/home/u");
        for text in [
            "",
            "not toml [",
            "[hooks]\n",
            "[hooks.outside_project_writes]\nallow = [\"/tmp/\"]\n",
            "[hooks.outside_project_writes]\nenabled = \"yes\"\n",
            "[hooks.outside_project_writes]\nenabled = true\nallow = [1]\n",
            "[hooks.primary_checkout]\nenabled = false\nrepos = [\"/r\"]\n",
            "[hooks.primary_checkout]\nenabled = true\nrepos = \"/r\"\n",
        ] {
            let c = parse(text, home);
            assert_eq!(c, HookConfig::default(), "{text:?}");
        }
    }

    #[test]
    fn lists_are_expanded_and_normalized() {
        let home = Path::new("/home/u");
        let c = parse(
            "[hooks.outside_project_writes]\nenabled = true\nallow = [\"~/.claude/\", \"~\", \"/tmp/x/../y\", \"relative/dir\"]\n\
             [hooks.primary_checkout]\nenabled = true\nrepos = [\"/no/such/repo/./a\"]\nallow_commands = [\"  git switch main  \"]\n",
            home,
        );
        assert_eq!(
            c.outside.unwrap().allow,
            vec![
                PathBuf::from("/home/u/.claude"),
                PathBuf::from("/home/u"),
                PathBuf::from("/tmp/y")
            ]
        );
        let p = c.primary.unwrap();
        assert_eq!(p.repos, vec![PathBuf::from("/no/such/repo/a")]);
        assert_eq!(p.allow_commands, vec!["git switch main".to_string()]);
    }

    #[test]
    fn normalize_applies_parent_steps_lexically() {
        for (input, want) in [
            ("/repo/../etc/x", "/etc/x"),
            ("/repo/new/../../etc", "/etc"),
            ("/../../x", "/x"),
            ("/a/./b/", "/a/b"),
        ] {
            assert_eq!(normalize(Path::new(input)), PathBuf::from(want), "{input}");
        }
    }
}
