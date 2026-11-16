//! Where a project's ledger file is (Issue #3068 P1, user decision: `~/.moai/projects/<id>/
//! ledger.db`). [`ledger_path`] is the only function that builds the path; the repository scan
//! `the_ledger_file_name_is_spelled_once` keeps the file name from being spelled anywhere else
//! in `src/`. HOME is an argument, never read from the environment here.

use super::LedgerError;
use std::path::{Path, PathBuf};

/// The ledger's file name.
const LEDGER_FILE: &str = "ledger.db";

/// `<home>/.moai/projects/<project_id>`: the folder `register` creates and nothing else does.
/// `project_id` must be a portable id (the profile's `project_id` form), so it cannot name a
/// parent folder or a second level.
pub fn project_dir(home: &Path, project_id: &str) -> Result<PathBuf, LedgerError> {
    if !crate::profile::portable_id(project_id) {
        return Err(LedgerError::InvalidProjectId);
    }
    Ok(crate::home_workdir::projects_root(home).join(project_id))
}

/// `<home>/.moai/projects/<project_id>/ledger.db`. Builds a path; touches nothing.
pub fn ledger_path(home: &Path, project_id: &str) -> Result<PathBuf, LedgerError> {
    Ok(project_dir(home, project_id)?.join(LEDGER_FILE))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_path_is_home_dot_moai_projects_id_ledger() {
        let home = Path::new("/h");
        assert_eq!(
            ledger_path(home, "demo").unwrap(),
            Path::new("/h/.moai/projects/demo/ledger.db")
        );
        assert_eq!(
            project_dir(home, "demo").unwrap(),
            Path::new("/h/.moai/projects/demo")
        );
    }

    #[test]
    fn a_project_id_that_could_leave_its_folder_is_refused() {
        let home = Path::new("/h");
        let longest = "a".repeat(80);
        assert!(ledger_path(home, &longest).is_ok());
        let too_long = "a".repeat(81);
        for bad in [
            "", "..", ".", "a/b", "a\\b", "a b", "a.b", "/abs", "a\0b", &too_long,
        ] {
            assert_eq!(
                ledger_path(home, bad),
                Err(LedgerError::InvalidProjectId),
                "{bad:?}"
            );
        }
    }

    /// The file name is spelled once: here. Every other source file asks [`ledger_path`].
    #[test]
    fn the_ledger_file_name_is_spelled_once() {
        let own = Path::new(env!("CARGO_MANIFEST_DIR")).join("src/ledger/location.rs");
        let needle = ["ledger", ".db"].concat();
        let mut stack = vec![Path::new(env!("CARGO_MANIFEST_DIR")).join("src")];
        let (mut files, mut found) = (0, Vec::new());
        while let Some(dir) = stack.pop() {
            for entry in std::fs::read_dir(&dir).unwrap() {
                let path = entry.unwrap().path();
                if path.is_dir() {
                    stack.push(path);
                } else if path.extension().is_some_and(|e| e == "rs") && path != own {
                    files += 1;
                    let text = std::fs::read_to_string(&path).unwrap();
                    for (n, line) in text.lines().enumerate() {
                        if line.contains(&needle) && !line.trim_start().starts_with("//") {
                            found.push(format!("{}:{}", path.display(), n + 1));
                        }
                    }
                }
            }
        }
        assert!(files > 100, "the walk saw {files} files; src moved?");
        assert!(
            found.is_empty(),
            "build the ledger path through ledger::location::ledger_path:\n{}",
            found.join("\n")
        );
    }

    #[test]
    fn the_spelling_scan_sees_a_literal() {
        // The needle is assembled so this file's own scan line never matches itself.
        let needle = ["ledger", ".db"].concat();
        assert!(format!("let p = dir.join(\"{needle}\");").contains(&needle));
        assert_eq!(needle, "ledger.db");
        assert_eq!(LEDGER_FILE, needle);
    }
}
