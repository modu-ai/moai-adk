//! Which folder under HOME holds this generation's state — the HOME twin of
//! [`crate::state_root::dir`]. Always `~/.moai`: the earlier `~/.moai-cli` is never looked at,
//! whatever it holds, and is never written (user decision 6504,
//! SPEC-20261123-home-state-retire). A HOME that still has one is cleaned up by the confirmed
//! step `home-state-retire-*` ([`crate::home_state_retire`]), the only code that names it.
//! The Go generation's files in `~/.moai` and this generation's names share the folder; the HOME
//! name table ([`crate::home_layout`]) keeps them apart. A schema-2 marker (#2871 I2 R1,
//! [`crate::state_root::classify_home`]) makes the HOME `rust-first`, laid out by
//! [`crate::home_layout`]. Reads only.

use std::path::{Path, PathBuf};

/// This generation's HOME state folder name (#2696 C4).
pub const MOAI_DIR: &str = crate::state_root::MOAI_DIR;

/// `.moai`, for every HOME.
pub fn dir(_home: &Path) -> &'static str {
    MOAI_DIR
}

/// `home/.moai`; a `rust-first` HOME resolves to the layout's root
/// ([`crate::home_layout::rust_path`], user decision D-M1), which is the same folder.
pub fn path(home: &Path) -> PathBuf {
    if crate::state_root::classify_home(home).generation == "rust-first" {
        crate::home_layout::rust_path(home, "")
    } else {
        home.join(MOAI_DIR)
    }
}

/// #2696 C4b2: a HOME with no MoAI state at all — verdict `none` (no `~/.moai` of any kind,
/// Go's included) — starts on `~/.moai`. An earlier `~/.moai-cli` does not count.
pub fn is_new_home(home: &Path) -> bool {
    crate::state_root::classify_home(home).generation == "none"
}

/// Make a new HOME's `~/.moai/` and its marker, so every later read and write resolves to
/// `~/.moai`. Does nothing for any other HOME (a Go `~/.moai`, any `~/.moai`) and says
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
    fn every_home_resolves_to_moai_whatever_the_earlier_folder_holds() {
        let home = tempfile::tempdir().unwrap();
        let h = home.path();
        assert_eq!(dir(h), MOAI_DIR, "an empty HOME starts on .moai");
        fs::create_dir(h.join(crate::HOME_STATE_DIR)).unwrap();
        assert_eq!(dir(h), MOAI_DIR, "the earlier folder is not looked at");
        assert_eq!(path(h), h.join(".moai"));
        fs::remove_dir(h.join(crate::HOME_STATE_DIR)).unwrap();
        // The Go generation's `~/.moai` shares the folder; the name table keeps the names apart.
        fs::create_dir_all(h.join(".moai/state")).unwrap();
        fs::write(h.join(".moai/state/active-sessions.json"), "{}").unwrap();
        assert_eq!(dir(h), MOAI_DIR);
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
    fn a_home_with_no_state_is_adopted_the_earlier_folder_does_not_count() {
        // New HOME: `.moai` and a marker that `dir` then reads.
        let home = tempfile::tempdir().unwrap();
        assert!(is_new_home(home.path()));
        assert_eq!(adopt_new(home.path(), "test"), Ok(true));
        assert_eq!(dir(home.path()), MOAI_DIR);
        // #2871 marker rename: the new name with generation "moai", and no earlier name.
        let written = fs::read_to_string(home.path().join(".moai/moai.json")).unwrap();
        assert!(written.contains(r#""generation": "moai""#), "{written}");
        assert!(!home.path().join(".moai/moai-cli.json").exists());
        assert_eq!(adopt_new(home.path(), "test"), Ok(false), "only once");
        // An earlier `~/.moai-cli` is ignored: the HOME is new and the folder stays untouched.
        let earlier = tempfile::tempdir().unwrap();
        fs::create_dir(earlier.path().join(crate::HOME_STATE_DIR)).unwrap();
        assert!(is_new_home(earlier.path()));
        assert_eq!(adopt_new(earlier.path(), "test"), Ok(true));
        assert!(earlier.path().join(MOAI_DIR).join("moai.json").exists());
        assert!(fs::read_dir(earlier.path().join(crate::HOME_STATE_DIR))
            .unwrap()
            .next()
            .is_none());
        // A Go `~/.moai`: not ours, nothing written into it.
        let go = tempfile::tempdir().unwrap();
        fs::create_dir_all(go.path().join(".moai/state")).unwrap();
        fs::write(go.path().join(".moai/state/active-sessions.json"), "{}").unwrap();
        assert!(!is_new_home(go.path()));
        assert_eq!(adopt_new(go.path(), "test"), Ok(false));
        assert!(!go.path().join(".moai/moai-cli.json").exists());
        assert!(!go.path().join(".moai/moai.json").exists());
    }

    /// #2871 I2 R1 (D-M1): a schema-2 marker resolves HOME paths to the layout's root even
    /// beside Go markers; a schema-1 marker beside Go markers answers `.moai` too.
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
        assert_eq!(dir(mixed.path()), MOAI_DIR);
        assert_eq!(path(mixed.path()), mixed.path().join(".moai"));
    }

    /// Production code asks [`dir`]/[`path`]; the earlier name is spelled only where both
    /// names are meant, plus unit-test fixtures of the earlier layout.
    #[test]
    fn home_state_paths_are_resolved_per_home() {
        let allowed: &[(&str, usize, &str)] = &[
            (
                "star_ask.rs",
                1,
                "the unit test creates the retired folder for the retired-folder refusal; the code asks home_root::dir/path",
            ),
            (
                "home_install.rs",
                3,
                "the install reads the earlier folder whatever generation the HOME is in (one use in `old_root`, which the plan, apply and revert ask, one in the install-journal lookup that tells the readers an earlier install id (R5), one in the unit-test fixture)",
            ),
            (
                "home_state_retire.rs",
                2,
                "the one place that names the earlier folder on purpose: the archive step's source path and one unit-test fixture",
            ),
            (
                "hooks.rs",
                1,
                "hooks are refused inside either HOME state root",
            ),
            (
                "project_home.rs",
                2,
                "unit-test fixtures of the earlier layout (the expected folder of a HOME with no marker); the code asks home_root::path",
            ),
            (
                "statusline/render.rs",
                1,
                "unit-test fixture of the earlier layout",
            ),
            (
                "update_keys.rs",
                1,
                "the keys layer lists legacy key files under either HOME state root",
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
