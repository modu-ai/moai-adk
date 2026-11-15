//! The project id that does not move (#3009 S1, design comment 6887461482 §3, as decided in
//! comment 6889383420: option A).
//!
//! The project's fixed id is the `project_id` key the project profile already has: the JSON
//! object in `<state folder>/project.yaml` (`profile.rs`), tracked with the project, so every
//! clone, every worktree and a renamed folder read the same id from it. This module reads that
//! key and makes the random id a profile gets when its request has none; it writes no file of
//! its own. A folder path or a remote URL names a place; the id names the project.
//!
//! An existing id is never replaced, whatever its form: `demo` stays `demo`. Moving a project
//! that has such an id to a random one would move the HOME folders keyed by it, which is the
//! data move D4 and waits for the user. A later `check-tree` can warn about ids that are not
//! random.
use crate::profile::portable_id;
use chacha20poly1305::aead::{rand_core::RngCore, OsRng};
use serde_json::Value;
use std::path::Path;

type Result<T> = std::result::Result<T, String>;

/// Characters of a random id: 128 random bits as lower-case base32 (RFC 4648 alphabet).
pub const ID_LEN: usize = 26;

const ALPHABET: &[u8; 32] = b"abcdefghijklmnopqrstuvwxyz234567";

/// Whether `id` has the form [`new_id`] makes: exactly [`ID_LEN`] characters of `a-z` and
/// `2-7`. Older ids chosen by a person (`demo`) are valid ids too, just not of this form.
pub fn is_random_form(id: &str) -> bool {
    id.len() == ID_LEN && id.bytes().all(|b| ALPHABET.contains(&b))
}

/// A fresh random id from the operating system's random source (no new dependency: the key
/// vault already uses it). Safe as a folder name on a case-insensitive filesystem, and a
/// valid profile `project_id`.
pub fn new_id() -> Result<String> {
    let mut random = [0u8; 16];
    OsRng
        .try_fill_bytes(&mut random)
        .map_err(|_| "the system random source is unavailable".to_string())?;
    let mut id = String::with_capacity(ID_LEN);
    let (mut buffer, mut bits) = (0u32, 0u32);
    for byte in random {
        buffer = (buffer << 8) | u32::from(byte);
        bits += 8;
        while bits >= 5 {
            bits -= 5;
            id.push(char::from(ALPHABET[((buffer >> bits) & 31) as usize]));
        }
    }
    if bits > 0 {
        id.push(char::from(ALPHABET[((buffer << (5 - bits)) & 31) as usize]));
    }
    Ok(id)
}

/// The `project_id` in the bytes of a profile file: a JSON object whose `project_id` is a
/// string a profile accepts. The error never echoes the value.
pub fn of_profile_bytes(bytes: &[u8]) -> Result<String> {
    let value: Value = serde_json::from_slice(bytes)
        .map_err(|_| "existing YAML requires preservation review".to_string())?;
    let map = value.as_object().ok_or("invalid profile object")?;
    match map.get("project_id").and_then(Value::as_str) {
        Some(id) if portable_id(id) => Ok(id.to_string()),
        _ => Err("the profile has no usable project_id".into()),
    }
}

/// The id of `project`: `Ok(None)` when it has no profile yet, `Ok(Some(id))` when the profile
/// carries one, and an error when the profile is there but is a link, is not a regular file,
/// is too large, is not JSON, or has no usable `project_id`. Reads only.
pub fn read(project: &Path) -> Result<Option<String>> {
    crate::profile::existing_project_id(project)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::{collections::HashSet, fs};

    const GOOD: &str = "abcdefghijklmnopqrstuvwxyz";

    fn put_profile(project: &Path, dir: &str, text: &str) {
        fs::create_dir_all(project.join(dir)).unwrap();
        fs::write(project.join(dir).join("project.yaml"), text).unwrap();
    }

    #[test]
    fn a_new_id_has_the_random_form_and_is_a_valid_profile_id() {
        let mut seen = HashSet::new();
        for _ in 0..200 {
            let id = new_id().unwrap();
            assert!(is_random_form(&id), "{id}");
            assert!(portable_id(&id), "{id}");
            assert!(seen.insert(id), "two ids were the same");
        }
    }

    #[test]
    fn the_random_form_edges_are_25_no_26_yes_27_no() {
        assert!(!is_random_form(&GOOD[..25]));
        assert!(is_random_form(GOOD));
        assert!(!is_random_form(&format!("{GOOD}a")));
        assert!(!is_random_form(&GOOD.to_uppercase()));
        assert!(!is_random_form(&format!("{}1", &GOOD[..25])));
        assert!(!is_random_form("demo"));
        assert!(!is_random_form(""));
    }

    #[test]
    fn a_project_without_a_profile_has_no_id_and_reading_writes_nothing() {
        let p = tempfile::tempdir().unwrap();
        assert_eq!(read(p.path()).unwrap(), None);
        assert_eq!(fs::read_dir(p.path()).unwrap().count(), 0);
    }

    #[test]
    fn the_id_is_read_from_the_profile_whatever_its_form() {
        for id in ["demo", GOOD, "My_proj-2"] {
            let p = tempfile::tempdir().unwrap();
            put_profile(
                p.path(),
                crate::PROJECT_STATE_DIR,
                &format!("{{\n  \"schema_version\": 2,\n  \"project_id\": \"{id}\"\n}}\n"),
            );
            assert_eq!(read(p.path()).unwrap().as_deref(), Some(id));
        }
    }

    #[test]
    fn a_profile_without_one_usable_id_is_an_error_and_is_left_alone() {
        let long = "a".repeat(81);
        let cases: Vec<(&str, String)> = vec![
            ("empty file", String::new()),
            ("yaml text", "project_id: demo\n".into()),
            ("not an object", "[\"demo\"]".into()),
            ("no key", "{\"schema_version\":2}".into()),
            ("not a string", "{\"project_id\":7}".into()),
            ("empty id", "{\"project_id\":\"\"}".into()),
            ("a slash", "{\"project_id\":\"a/b\"}".into()),
            ("a space", "{\"project_id\":\"a b\"}".into()),
            ("81 characters", format!("{{\"project_id\":\"{long}\"}}")),
        ];
        for (name, text) in cases {
            let p = tempfile::tempdir().unwrap();
            put_profile(p.path(), crate::PROJECT_STATE_DIR, &text);
            assert!(read(p.path()).is_err(), "{name}");
            assert_eq!(
                fs::read_to_string(p.path().join(crate::PROJECT_STATE_DIR).join("project.yaml"))
                    .unwrap(),
                text,
                "{name}: the file must be left as it was"
            );
        }
        // 80 characters is the profile's own limit and is accepted.
        let p = tempfile::tempdir().unwrap();
        let eighty = "a".repeat(80);
        put_profile(
            p.path(),
            crate::PROJECT_STATE_DIR,
            &format!("{{\"project_id\":\"{eighty}\"}}"),
        );
        assert_eq!(read(p.path()).unwrap(), Some(eighty));
    }

    #[test]
    fn an_error_does_not_echo_the_value() {
        // Built at run time so no token-shaped literal sits in the source.
        let secret = format!("{}{}/", "ghp", "_abcdefghijklmnopqrstuvwxyz0123456789");
        let err =
            of_profile_bytes(format!("{{\"project_id\":\"{secret}\"}}").as_bytes()).unwrap_err();
        assert!(
            !err.contains(&secret) && !err.contains("abcdefghijklmnopqrstuv"),
            "{err}"
        );
    }

    #[test]
    fn a_renamed_folder_keeps_its_id() {
        let parent = tempfile::tempdir().unwrap();
        let before = parent.path().join("before");
        put_profile(
            &before,
            crate::PROJECT_STATE_DIR,
            &format!("{{\"project_id\":\"{GOOD}\"}}"),
        );
        let after = parent.path().join("after");
        fs::rename(&before, &after).unwrap();
        assert_eq!(read(&after).unwrap().as_deref(), Some(GOOD));
    }

    #[cfg(unix)]
    #[test]
    fn a_profile_that_is_a_link_is_refused_and_not_followed() {
        use std::os::unix::fs::symlink;
        let p = tempfile::tempdir().unwrap();
        let other = tempfile::tempdir().unwrap();
        let real = other.path().join("real.yaml");
        fs::write(&real, format!("{{\"project_id\":\"{GOOD}\"}}")).unwrap();
        fs::create_dir(p.path().join(crate::PROJECT_STATE_DIR)).unwrap();
        symlink(
            &real,
            p.path().join(crate::PROJECT_STATE_DIR).join("project.yaml"),
        )
        .unwrap();
        assert!(read(p.path()).is_err());
    }
}
