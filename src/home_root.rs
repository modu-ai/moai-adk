//! #2696 C4b: which folder under HOME holds this generation's state — the HOME twin of
//! [`crate::state_root::dir`]. `~/.moai` once [`crate::state_root::classify_with`] with the
//! HOME's Go markers says `rust` (a valid marker `~/.moai/moai-cli.json`), otherwise the earlier
//! `~/.moai-cli` ([`crate::HOME_STATE_DIR`]). Read both, one writes: a HOME without the marker
//! keeps exactly the paths it had, and the Go generation's `~/.moai` is never written here.
//! A schema-2 marker (#2871 I2 R1, [`crate::state_root::classify_home`]) makes the HOME
//! `rust-first`: `~/.moai` is this generation's even beside Go markers, laid out by
//! [`crate::home_layout`]. Reads only.

use std::path::{Path, PathBuf};

/// This generation's HOME state folder name once the marker is there (#2696 C4).
pub const MOAI_DIR: &str = crate::state_root::MOAI_DIR;

/// `.moai` when `home/.moai` is this generation's (verdict `rust`, or `rust-first` for a
/// schema-2 marker, #2871 I2 R1), else `.moai-cli`.
pub fn dir(home: &Path) -> &'static str {
    match crate::state_root::classify_home(home).generation {
        "rust" | "rust-first" => MOAI_DIR,
        _ => crate::HOME_STATE_DIR,
    }
}

/// `home/<dir(home)>`; a `rust-first` HOME resolves to the layout's root
/// ([`crate::home_layout::rust_path`], user decision D-M1).
pub fn path(home: &Path) -> PathBuf {
    match crate::state_root::classify_home(home).generation {
        "rust-first" => crate::home_layout::rust_path(home, ""),
        "rust" => home.join(MOAI_DIR),
        _ => home.join(crate::HOME_STATE_DIR),
    }
}

/// #2696 C4b2: a HOME with no MoAI state at all — no earlier `~/.moai-cli` and verdict `none`
/// (no `~/.moai` of any kind, Go's included) — starts on `~/.moai`.
pub fn is_new_home(home: &Path) -> bool {
    std::fs::symlink_metadata(home.join(crate::HOME_STATE_DIR)).is_err()
        && crate::state_root::classify_home(home).generation == "none"
}

/// Make a new HOME's `~/.moai/` and its marker, so every later read and write resolves to
/// `~/.moai`. Does nothing for any other HOME (an earlier `~/.moai-cli`, a Go `~/.moai`) and says
/// whether it wrote. The folder is created exclusively and the marker published without
/// overwriting.
pub fn adopt_new(home: &Path, created_by: &str) -> Result<bool, String> {
    if !is_new_home(home) {
        return Ok(false);
    }
    std::fs::create_dir(home.join(MOAI_DIR))
        .map_err(|_| format!("~/{MOAI_DIR}: folder create failed"))?;
    crate::apply::publish_exclusive(
        &home.join(MOAI_DIR).join(crate::state_root::RUST_MARKER),
        &crate::state_root::marker_bytes(created_by, &[]),
    )?;
    Ok(true)
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

    #[test]
    fn only_a_home_with_no_state_is_adopted() {
        // New HOME: `.moai` and a marker that `dir` then reads.
        let home = tempfile::tempdir().unwrap();
        assert!(is_new_home(home.path()));
        assert_eq!(adopt_new(home.path(), "test"), Ok(true));
        assert_eq!(dir(home.path()), MOAI_DIR);
        assert_eq!(adopt_new(home.path(), "test"), Ok(false), "only once");
        // An earlier `~/.moai-cli`: kept, nothing written.
        let earlier = tempfile::tempdir().unwrap();
        fs::create_dir(earlier.path().join(crate::HOME_STATE_DIR)).unwrap();
        assert!(!is_new_home(earlier.path()));
        assert_eq!(adopt_new(earlier.path(), "test"), Ok(false));
        assert!(!earlier.path().join(MOAI_DIR).exists());
        // A Go `~/.moai`: not ours, nothing written into it.
        let go = tempfile::tempdir().unwrap();
        fs::create_dir_all(go.path().join(".moai/state")).unwrap();
        fs::write(go.path().join(".moai/state/active-sessions.json"), "{}").unwrap();
        assert!(!is_new_home(go.path()));
        assert_eq!(adopt_new(go.path(), "test"), Ok(false));
        assert!(!go.path().join(".moai/moai-cli.json").exists());
    }

    /// #2871 I2 R1 (D-M1): a schema-2 marker resolves HOME paths to the layout's root even
    /// beside Go markers; a schema-1 marker beside Go markers stays `mixed` → the earlier root.
    #[test]
    fn a_schema_2_home_resolves_to_the_layout_root() {
        let v2 = r#"{"schema_version":2,"generation":"moai-cli"}"#;
        let v1 = r#"{"schema_version":1,"generation":"moai-cli"}"#;
        let home = |marker: &str| {
            let t = tempfile::tempdir().unwrap();
            fs::create_dir_all(t.path().join(".moai/state")).unwrap();
            fs::write(t.path().join(".moai/state/active-sessions.json"), "[]").unwrap();
            fs::write(t.path().join(".moai/moai-cli.json"), marker).unwrap();
            t
        };
        let shared = home(v2);
        let h = shared.path();
        assert_eq!(dir(h), MOAI_DIR);
        assert_eq!(path(h), crate::home_layout::rust_path(h, ""));
        assert_eq!(path(h), h.join(".moai"));
        assert!(!is_new_home(h));
        assert_eq!(adopt_new(h, "test"), Ok(false));
        let mixed = home(v1);
        assert_eq!(dir(mixed.path()), crate::HOME_STATE_DIR);
        assert_eq!(path(mixed.path()), mixed.path().join(crate::HOME_STATE_DIR));
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
