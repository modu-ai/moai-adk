//! #2696 C4b: which folder under HOME holds this generation's state — the HOME twin of
//! [`crate::state_root::dir`]. `~/.moai` once [`crate::state_root::classify_with`] with the
//! HOME's Go markers says `rust` (a valid marker `~/.moai/moai-cli.json`), otherwise the earlier
//! `~/.moai-cli` ([`crate::HOME_STATE_DIR`]). Read both, one writes: a HOME without the marker
//! keeps exactly the paths it had, and the Go generation's `~/.moai` is never written here.
//! Reads only.

use std::path::{Path, PathBuf};

/// This generation's HOME state folder name once the marker is there (#2696 C4).
pub const MOAI_DIR: &str = crate::state_root::MOAI_DIR;

/// `.moai` when `home/.moai` is this generation's (verdict `rust`), else `.moai-cli`.
pub fn dir(home: &Path) -> &'static str {
    if crate::state_root::classify_with(home, crate::home_state::home_go_markers).generation
        == "rust"
    {
        MOAI_DIR
    } else {
        crate::HOME_STATE_DIR
    }
}

/// `home/<dir(home)>`.
pub fn path(home: &Path) -> PathBuf {
    home.join(dir(home))
}

/// A folder name that is a HOME state root of either name (walks that look for the root).
pub fn is_state_dir_name(name: &std::ffi::OsStr) -> bool {
    name == crate::HOME_STATE_DIR || name == MOAI_DIR
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    #[test]
    fn the_marker_alone_moves_home_state_to_moai() {
        let home = tempfile::tempdir().unwrap();
        let h = home.path();
        assert_eq!(
            dir(h),
            crate::HOME_STATE_DIR,
            "empty HOME keeps the earlier root"
        );
        // The Go generation's `~/.moai` is not this generation's.
        fs::create_dir_all(h.join(".moai/state")).unwrap();
        fs::write(h.join(".moai/state/active-sessions.json"), "{}").unwrap();
        assert_eq!(dir(h), crate::HOME_STATE_DIR);
        let fresh = tempfile::tempdir().unwrap();
        fs::create_dir(fresh.path().join(".moai")).unwrap();
        fs::write(
            fresh.path().join(".moai/moai-cli.json"),
            r#"{"schema_version":1,"generation":"moai-cli"}"#,
        )
        .unwrap();
        assert_eq!(dir(fresh.path()), ".moai");
        assert_eq!(path(fresh.path()), fresh.path().join(".moai"));
        assert!(is_state_dir_name(std::ffi::OsStr::new(".moai")));
        assert!(is_state_dir_name(std::ffi::OsStr::new(
            crate::HOME_STATE_DIR
        )));
        assert!(!is_state_dir_name(std::ffi::OsStr::new(".moai-backup")));
    }

    /// Production code asks [`dir`]/[`path`]; the earlier name is spelled only where both
    /// names are meant, plus unit-test fixtures of the earlier layout.
    #[test]
    fn home_state_paths_are_resolved_per_home() {
        let allowed: &[(&str, usize, &str)] = &[
            (
                "hooks.rs",
                1,
                "hooks are refused inside either HOME state root",
            ),
            (
                "statusline/render.rs",
                1,
                "unit-test fixture of the earlier layout",
            ),
            (
                "usage_limit.rs",
                2,
                "unit-test fixture of the earlier layout",
            ),
        ];
        let root = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("src");
        let mut stack = vec![root.clone()];
        let mut found = std::collections::BTreeMap::new();
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                if p.is_dir() {
                    stack.push(p);
                    continue;
                }
                if p.extension().is_none_or(|x| x != "rs") {
                    continue;
                }
                let rel = p
                    .strip_prefix(&root)
                    .unwrap()
                    .to_string_lossy()
                    .replace('\\', "/");
                if rel == "lib.rs" || rel == "home_root.rs" {
                    continue;
                }
                let n = fs::read_to_string(&p)
                    .unwrap()
                    .lines()
                    .filter(|l| !l.trim_start().starts_with("//"))
                    .filter(|l| l.contains("HOME_STATE_DIR"))
                    .count();
                if n > 0 {
                    found.insert(rel, n);
                }
            }
        }
        let want: std::collections::BTreeMap<String, usize> = allowed
            .iter()
            .map(|(f, n, _)| (f.to_string(), *n))
            .collect();
        assert_eq!(found, want, "ask crate::home_root::{{dir, path}} instead");
    }
}
