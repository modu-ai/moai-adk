//! Where a project's machine-written data lives in the HOME (#3009 S2, design comment
//! 6887461482 §2 (나)): `<HOME state root>/projects/<project key>/`. Reads only; nothing is
//! created, moved or written, and nothing calls it yet (the Go-state backup and receipt move
//! onto it in the next slice, #2936 G5-1, leader decision 6887465160).
//!
//! The project key is the project's fixed id ([`crate::project_id`]: the profile's
//! `project_id`) when the project has a profile. Before it has one, the key is a hash of the
//! project folder's real path, the same place the data will move from, so a project in the
//! middle of its first update still has a stable folder. The hash key changes if the folder is
//! renamed; that is what the id is for, and [`ProjectHome::key_source`] says which one this is.
use crate::{contract::PROJECT_PATH_UNAVAILABLE, profile::portable_id};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// Where the key came from.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum KeySource {
    /// The profile's `project_id`.
    Profile,
    /// A hash of the project folder's real path (no profile yet).
    PathHash,
}

/// The HOME folder of one project and the key that names it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ProjectHome {
    /// `<HOME state root>/projects/<key>`. Not created by this module.
    pub dir: PathBuf,
    pub key: String,
    pub key_source: KeySource,
}

/// The key a project folder without a profile gets: `path-` and 32 hex digits of the SHA-256
/// of its real path.
fn path_hash_key(real: &Path) -> Result<String> {
    let text = real
        .to_str()
        .ok_or("the project folder path is not UTF-8")?;
    let digest = crate::digest(text.as_bytes());
    Ok(format!("path-{}", &digest["sha256:".len()..][..32]))
}

/// The HOME folder of `project` under `home` (the user's HOME folder, not its state root: the
/// state root is chosen the way every HOME reader chooses it, [`crate::home_root::path`]).
/// A profile that is there but cannot give an id is an error; the key is never guessed from
/// the path in that case, since the path would name a different folder than the id will.
pub fn resolve(home: &Path, project: &Path) -> Result<ProjectHome> {
    let real = std::fs::canonicalize(project).map_err(|_| PROJECT_PATH_UNAVAILABLE.to_string())?;
    let (key, key_source) = match crate::project_id::read(&real)? {
        Some(id) => (id, KeySource::Profile),
        None => (path_hash_key(&real)?, KeySource::PathHash),
    };
    Ok(ProjectHome {
        dir: crate::home_root::path(home).join("projects").join(&key),
        key,
        key_source,
    })
}

impl ProjectHome {
    /// `<dir>/backups/<stamp>/<kind>`: where one backup of `kind` (for example `go-state`)
    /// taken at `stamp` (`YYYYMMDDTHHMMSSZ`) goes. Both parts are checked, so neither can name
    /// a parent folder or a second level.
    pub fn backup_area(&self, stamp: &str, kind: &str) -> Result<PathBuf> {
        if !crate::state_root::utc_stamp(stamp) {
            return Err(format!("stamp `{stamp}` is not YYYYMMDDTHHMMSSZ"));
        }
        if !portable_id(kind) {
            return Err("backup kind is not a plain name".into());
        }
        Ok(self.dir.join("backups").join(stamp).join(kind))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;
    use std::fs;

    const STAMP: &str = "20261115T050000Z";

    fn profile(project: &Path, id: &str) {
        let dir = project.join(crate::PROJECT_STATE_DIR);
        fs::create_dir_all(&dir).unwrap();
        fs::write(
            dir.join("project.yaml"),
            format!("{{\"project_id\":\"{id}\"}}"),
        )
        .unwrap();
    }

    /// Every path and size under `dir`, so a test can show nothing was created.
    fn listing(dir: &Path) -> Vec<String> {
        let mut out = Vec::new();
        let mut stack = vec![dir.to_path_buf()];
        while let Some(d) = stack.pop() {
            for e in fs::read_dir(&d).unwrap().flatten() {
                out.push(e.path().strip_prefix(dir).unwrap().display().to_string());
                if e.path().is_dir() {
                    stack.push(e.path());
                }
            }
        }
        out.sort();
        out
    }

    #[test]
    fn a_project_with_a_profile_is_keyed_by_its_id_and_nothing_is_created() {
        let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
        profile(project.path(), "abcdefghijklmnopqrstuvwxyz");
        let (h0, p0) = (listing(home.path()), listing(project.path()));
        let ph = resolve(home.path(), project.path()).unwrap();
        assert_eq!(ph.key, "abcdefghijklmnopqrstuvwxyz");
        assert_eq!(ph.key_source, KeySource::Profile);
        assert_eq!(
            ph.dir,
            home.path()
                .join(crate::HOME_STATE_DIR)
                .join("projects")
                .join("abcdefghijklmnopqrstuvwxyz")
        );
        assert_eq!((listing(home.path()), listing(project.path())), (h0, p0));
    }

    #[test]
    fn a_person_chosen_id_is_a_key_too() {
        let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
        profile(project.path(), "demo");
        let ph = resolve(home.path(), project.path()).unwrap();
        assert_eq!(
            (ph.key.as_str(), ph.key_source),
            ("demo", KeySource::Profile)
        );
    }

    #[test]
    fn a_project_without_a_profile_is_keyed_by_a_hash_of_its_real_path() {
        let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
        let other = tempfile::tempdir().unwrap();
        let a = resolve(home.path(), project.path()).unwrap();
        let b = resolve(home.path(), project.path()).unwrap();
        let c = resolve(home.path(), other.path()).unwrap();
        assert_eq!(a.key_source, KeySource::PathHash);
        assert!(a.key.starts_with("path-") && a.key.len() == 37, "{}", a.key);
        assert!(portable_id(&a.key));
        assert_eq!(a, b, "the same folder gets the same key");
        assert_ne!(a.key, c.key, "two folders get two keys");
        assert!(!a.key.contains(&*project.path().to_string_lossy()));
    }

    #[test]
    fn a_renamed_folder_keeps_its_profile_key_and_loses_its_hash_key() {
        let (home, parent) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
        let (with, without) = (parent.path().join("with"), parent.path().join("without"));
        fs::create_dir(&without).unwrap();
        profile(&with, "abcdefghijklmnopqrstuvwxyz");
        let (k1, h1) = (
            resolve(home.path(), &with).unwrap(),
            resolve(home.path(), &without).unwrap(),
        );
        let (with2, without2) = (parent.path().join("with2"), parent.path().join("without2"));
        fs::rename(&with, &with2).unwrap();
        fs::rename(&without, &without2).unwrap();
        assert_eq!(resolve(home.path(), &with2).unwrap(), k1);
        assert_ne!(resolve(home.path(), &without2).unwrap().key, h1.key);
    }

    #[test]
    fn a_profile_that_gives_no_id_is_an_error_and_the_path_is_not_used_instead() {
        let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
        let dir = project.path().join(crate::PROJECT_STATE_DIR);
        fs::create_dir_all(&dir).unwrap();
        fs::write(dir.join("project.yaml"), "project_id: by hand\n").unwrap();
        assert!(resolve(home.path(), project.path()).is_err());
    }

    #[test]
    fn a_missing_project_folder_says_so() {
        let home = tempfile::tempdir().unwrap();
        let err = resolve(home.path(), &home.path().join("nowhere")).unwrap_err();
        assert_eq!(err, PROJECT_PATH_UNAVAILABLE);
    }

    #[test]
    fn the_state_root_follows_the_home_the_way_every_home_reader_does() {
        let project = tempfile::tempdir().unwrap();
        profile(project.path(), "abcdefghijklmnopqrstuvwxyz");
        // A HOME with nothing: the earlier folder.
        let plain = tempfile::tempdir().unwrap();
        let ph = resolve(plain.path(), project.path()).unwrap();
        assert!(ph.dir.starts_with(plain.path().join(crate::HOME_STATE_DIR)));
        // A HOME with this generation's marker: `.moai`, and the same as `home_root::path`.
        for schema in [1u64, crate::state_root::HOME_MARKER_SCHEMA] {
            let home = tempfile::tempdir().unwrap();
            let moai = home.path().join(crate::state_root::MOAI_DIR);
            fs::create_dir(&moai).unwrap();
            fs::write(
                moai.join(crate::state_root::RUST_MARKER),
                json!({"schema_version": schema, "generation": "moai-cli"}).to_string(),
            )
            .unwrap();
            let ph = resolve(home.path(), project.path()).unwrap();
            assert_eq!(
                ph.dir,
                crate::home_root::path(home.path())
                    .join("projects")
                    .join("abcdefghijklmnopqrstuvwxyz"),
                "schema {schema}"
            );
            assert!(
                ph.dir.starts_with(&moai),
                "schema {schema}: {}",
                ph.dir.display()
            );
        }
    }

    #[test]
    fn a_backup_area_is_projects_key_backups_stamp_kind() {
        let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
        profile(project.path(), "demo");
        let ph = resolve(home.path(), project.path()).unwrap();
        assert_eq!(
            ph.backup_area(STAMP, "go-state").unwrap(),
            ph.dir.join("backups").join(STAMP).join("go-state")
        );
    }

    #[test]
    fn a_stamp_or_kind_that_could_name_another_folder_is_refused() {
        let (home, project) = (tempfile::tempdir().unwrap(), tempfile::tempdir().unwrap());
        profile(project.path(), "demo");
        let ph = resolve(home.path(), project.path()).unwrap();
        for stamp in [
            "",
            "..",
            "20261115T050000",
            "20261115T050000Zx",
            "../20261115T050000Z",
        ] {
            assert!(
                ph.backup_area(stamp, "go-state").is_err(),
                "stamp {stamp:?}"
            );
        }
        for kind in ["", "..", ".", "a/b", "a\\b", "a b", &"k".repeat(81)] {
            assert!(ph.backup_area(STAMP, kind).is_err(), "kind {kind:?}");
        }
        // The edges: a kind of 80 characters is the longest plain name.
        assert!(ph.backup_area(STAMP, &"k".repeat(80)).is_ok());
    }
}
