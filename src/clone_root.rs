//! #2696 C4c: which folder under the git common dir holds the clone-shared state (claims,
//! goal, sessions, run receipts) — the clone twin of [`crate::state_root::dir`]. `<common>/.moai`
//! once [`crate::state_root::classify`] says `rust` there (a valid marker
//! `<common>/.moai/moai-cli.json`), otherwise the earlier `<common>/.moai-cli`
//! ([`crate::CLONE_STATE_DIR`]). Read both, one writes.
//!
//! Nothing moves this state and nothing writes that marker yet (decision D-C4-1 default,
//! #2696 6867006943): running lanes keep their claims and sessions in `.moai-cli` until the
//! leader picks the time for an explicit move. Reads only.

use std::path::{Path, PathBuf};

/// `.moai` when `common/.moai` is this generation's (verdict `rust`), else `.moai-cli`.
pub fn dir(common: &Path) -> &'static str {
    if crate::state_root::classify(common).generation == "rust" {
        crate::state_root::MOAI_DIR
    } else {
        crate::CLONE_STATE_DIR
    }
}

/// `common/<dir(common)>`.
pub fn path(common: &Path) -> PathBuf {
    common.join(dir(common))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    #[test]
    fn the_marker_alone_moves_clone_state_to_moai() {
        let common = tempfile::tempdir().unwrap();
        let c = common.path();
        assert_eq!(
            dir(c),
            crate::CLONE_STATE_DIR,
            "no marker: the earlier root"
        );
        fs::create_dir_all(c.join(crate::CLONE_STATE_DIR).join("claims")).unwrap();
        fs::create_dir(c.join(".moai")).unwrap();
        assert_eq!(dir(c), crate::CLONE_STATE_DIR, "an empty .moai is not ours");
        fs::write(
            c.join(".moai/moai-cli.json"),
            r#"{"schema_version":1,"generation":"moai-cli"}"#,
        )
        .unwrap();
        assert_eq!(dir(c), ".moai");
        assert_eq!(path(c), c.join(".moai"));
    }

    /// Production code asks [`dir`]/[`path`]; the earlier name lives only in src/lib.rs.
    #[test]
    fn clone_state_paths_are_resolved_per_clone() {
        let root = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("src");
        let mut stack = vec![root.clone()];
        let mut found = Vec::new();
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
                if rel == "lib.rs" || rel == "clone_root.rs" {
                    continue;
                }
                if fs::read_to_string(&p)
                    .unwrap()
                    .lines()
                    .any(|l| !l.trim_start().starts_with("//") && l.contains("CLONE_STATE_DIR"))
                {
                    found.push(rel);
                }
            }
        }
        assert!(
            found.is_empty(),
            "ask crate::clone_root::{{dir, path}}: {found:?}"
        );
    }
}
