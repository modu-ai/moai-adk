//! Worktree scopes (Issue #3010, item 1): the `scopes` key of `.moai-cli/project.yaml`.
//!
//! A scope names the folders one kind of work needs, so a worktree can be opened with only
//! those folders (a later piece) instead of the whole repository. The key is an ALLOW list:
//! a folder not named in a scope is not part of it, and the paths the tool itself reads at
//! the repository root are added by the worktree command, never typed here.
//!
//! This module only states and checks the declaration. It runs no git command and writes
//! nothing; the profile it lives in goes through `project-plan`/`project-apply` like any other
//! change. Limits are stated values chosen for the product, not measured ones.

use crate::SECRET_SHAPES;
use serde::{Deserialize, Serialize};

type Result<T> = std::result::Result<T, String>;

/// Most scopes one profile declares.
pub const SCOPE_LIMIT: usize = 16;
/// Most folders one scope names.
pub const PATH_LIMIT: usize = 64;
/// Longest folder path, in bytes.
pub const PATH_BYTES: usize = 256;
/// Longest scope name, in bytes.
pub const NAME_BYTES: usize = 64;

/// One scope: a name and the repository folders it needs, relative to the repository root.
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq, Eq)]
#[serde(deny_unknown_fields)]
pub struct ScopeEntry {
    pub name: String,
    pub paths: Vec<String>,
}

/// Whether `name` is a scope name: lowercase letters, digits, `-` and `_`, starting with a
/// letter or digit, at most [`NAME_BYTES`].
fn name_ok(name: &str) -> bool {
    !name.is_empty()
        && name.len() <= NAME_BYTES
        && name
            .bytes()
            .next()
            .is_some_and(|b| b.is_ascii_lowercase() || b.is_ascii_digit())
        && name
            .bytes()
            .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'-' || b == b'_')
}

/// Whether `path` is a plain repository-relative folder path: `/`-separated, no empty, `.` or
/// `..` part, nothing inside `.git`, no absolute or drive form, no glob or control character, no
/// trailing slash (one spelling per folder), no leading or trailing space in any part, no
/// secret-looking text.
pub fn path_ok(path: &str) -> bool {
    // A trailing slash leaves an empty last part, which the part check below refuses.
    if path.is_empty() || path.len() > PATH_BYTES {
        return false;
    }
    let lower = path.to_ascii_lowercase();
    if SECRET_SHAPES.iter().any(|s| lower.contains(s)) {
        return false;
    }
    if path.chars().any(|c| {
        c.is_control()
            || matches!(
                c,
                '\\' | ':' | '*' | '?' | '[' | ']' | '!' | '~' | '"' | '\''
            )
    }) {
        return false;
    }
    path.split('/').all(|part| {
        !part.is_empty()
            && part == part.trim()
            && part != "."
            && part != ".."
            && !part.eq_ignore_ascii_case(".git")
    })
}

/// Checks a `scopes` value: at most [`SCOPE_LIMIT`] scopes with distinct names, each naming
/// between one and [`PATH_LIMIT`] distinct valid folders. The list may be empty only by being
/// absent from the profile, so an empty `scopes` is refused.
pub fn validate(scopes: &[ScopeEntry]) -> Result<()> {
    if scopes.is_empty() || scopes.len() > SCOPE_LIMIT {
        return Err("invalid scopes: between 1 and 16 scopes".into());
    }
    let mut names = Vec::new();
    for s in scopes {
        if !name_ok(&s.name) || names.contains(&&s.name) {
            return Err("invalid scopes: a scope name is malformed or repeated".into());
        }
        names.push(&s.name);
        if s.paths.is_empty() || s.paths.len() > PATH_LIMIT {
            return Err("invalid scopes: a scope names between 1 and 64 folders".into());
        }
        // Folders that differ only in case are one folder on a case-insensitive disk.
        let mut seen: Vec<String> = Vec::new();
        for p in &s.paths {
            let lower = p.to_lowercase();
            if !path_ok(p) || seen.contains(&lower) {
                return Err("invalid scopes: a folder path is malformed or repeated".into());
            }
            seen.push(lower);
        }
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn scope(name: &str, paths: &[&str]) -> ScopeEntry {
        ScopeEntry {
            name: name.into(),
            paths: paths.iter().map(|p| p.to_string()).collect(),
        }
    }

    #[test]
    fn plain_folders_pass_and_everything_else_is_refused() {
        for ok in ["src", "apps/web", "packages/ui-kit", "a.b/c_d", ".github"] {
            assert!(path_ok(ok), "{ok}");
        }
        for no in [
            "",
            "/abs",
            "apps/",
            "apps//web",
            "./apps",
            "apps/./web",
            "../up",
            "apps/../x",
            ".git",
            ".GIT/hooks",
            "apps/.git/x",
            "C:/win",
            "a\\b",
            "apps/*",
            "apps/[ab]",
            "apps/web?",
            "!apps",
            "~/x",
            "a\"b",
            "line\nbreak",
            " apps",
            "apps ",
            "apps/ web",
            "apps/web /x",
            "password=fixture-only",
        ] {
            assert!(!path_ok(no), "{no:?}");
        }
    }

    #[test]
    fn the_path_length_limit_holds_at_its_edge() {
        assert!(path_ok(&"a".repeat(PATH_BYTES)));
        assert!(!path_ok(&"a".repeat(PATH_BYTES + 1)));
    }

    #[test]
    fn scope_names_are_lowercase_ids() {
        for ok in ["web", "api-2", "a_b", "9lives"] {
            assert!(name_ok(ok), "{ok}");
        }
        for no in ["", "Web", "-x", "_x", "a b", "a/b", "é"] {
            assert!(!name_ok(no), "{no:?}");
        }
        assert!(name_ok(&"a".repeat(NAME_BYTES)));
        assert!(!name_ok(&"a".repeat(NAME_BYTES + 1)));
    }

    #[test]
    fn counts_are_bounded_at_their_edges() {
        let many = |n: usize| -> Vec<ScopeEntry> {
            (0..n).map(|i| scope(&format!("s{i}"), &["src"])).collect()
        };
        assert!(validate(&many(SCOPE_LIMIT)).is_ok());
        assert!(validate(&many(SCOPE_LIMIT + 1)).is_err());
        assert!(validate(&[]).is_err());
        let paths = |n: usize| -> Vec<ScopeEntry> {
            let list: Vec<String> = (0..n).map(|i| format!("d{i}")).collect();
            vec![ScopeEntry {
                name: "web".into(),
                paths: list,
            }]
        };
        assert!(validate(&paths(PATH_LIMIT)).is_ok());
        assert!(validate(&paths(PATH_LIMIT + 1)).is_err());
        assert!(validate(&paths(0)).is_err());
    }

    #[test]
    fn repeated_names_and_repeated_folders_are_refused() {
        assert!(validate(&[scope("web", &["a"]), scope("api", &["b"])]).is_ok());
        assert!(validate(&[scope("web", &["a"]), scope("web", &["b"])]).is_err());
        assert!(validate(&[scope("web", &["a", "a"])]).is_err());
        assert!(validate(&[scope("web", &["apps", "Apps"])]).is_err());
        assert!(validate(&[scope("web", &["apps/Web", "APPS/web"])]).is_err());
        assert!(validate(&[scope("web", &["a", "b"]), scope("api", &["a"])]).is_ok());
        assert!(validate(&[scope("web", &["bad/"])]).is_err());
        assert!(validate(&[scope("Web", &["a"])]).is_err());
    }
}
