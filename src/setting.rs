//! `moai setting list|get` (#2871 follow-up S1) and the selected hosts (S3): the user's CLI
//! defaults in `<HOME state folder>/config/launcher.yaml`, read through [`crate::user_config`].
//!
//! The keys form a closed table. A key outside it is refused, a value outside its set is shown
//! as invalid without being quoted, and the answer lists no other content of the file (only how
//! many keys it holds that the table does not know). Reads only: nothing here writes, and the
//! project's own settings are not read at all — `/moai project` manages those.
//!
//! `install.hosts` names the hosts the user-level install covers: any non-empty subset of
//! `claude`, `codex` and `omp`, each once, in exact lower case. A HOME with no such key keeps
//! today's behaviour, `claude` and `codex` (the hosts `moai update` writes at the user level
//! now), so adding this reader changes no existing answer.

use crate::session::{configured_advisor, role_full_access, FULL_ACCESS_DEFAULT};
use crate::user_config;
use serde_json::{json, Map, Value};
use std::path::Path;

/// The hosts a user-level install can cover, in the order answers list them.
pub const HOSTS: [&str; 3] = ["claude", "codex", "omp"];
/// The hosts a HOME without `install.hosts` covers (what `moai update` writes today).
pub const DEFAULT_HOSTS: [&str; 2] = ["claude", "codex"];
/// Every key `moai setting` knows, in answer order.
pub const KEYS: [&str; 3] = ["install.hosts", "launcher.advisor", "launcher.role_access"];

fn split(key: &str) -> (&str, &str) {
    key.split_once('.').unwrap_or((key, ""))
}

fn describe(key: &str) -> &'static str {
    match key {
        "install.hosts" => "hosts the user-level install covers: any of claude, codex, omp",
        "launcher.advisor" => {
            "advisor of Claude launches: none, fable, opus, sonnet or a claude-* id"
        }
        _ => "access of role sessions: full or prompt",
    }
}

fn default_value(key: &str) -> Value {
    match key {
        "install.hosts" => json!(DEFAULT_HOSTS),
        "launcher.advisor" => json!("none"),
        _ => json!(if FULL_ACCESS_DEFAULT {
            "full"
        } else {
            "prompt"
        }),
    }
}

/// The hosts `value` names, in [`HOSTS`] order: a non-empty list of known names, each once.
pub fn hosts_from(value: &Value) -> Result<Vec<&'static str>, String> {
    let bad = || "install.hosts must be a non-empty list of claude, codex or omp, each once";
    let Value::Array(items) = value else {
        return Err(bad().into());
    };
    let mut picked = Vec::new();
    for item in items {
        let name = item.as_str().and_then(|s| HOSTS.iter().find(|h| **h == s));
        match name {
            Some(h) if !picked.contains(h) => picked.push(*h),
            _ => return Err(bad().into()),
        }
    }
    if picked.is_empty() {
        return Err(bad().into());
    }
    Ok(HOSTS
        .iter()
        .copied()
        .filter(|h| picked.contains(h))
        .collect())
}

/// Whether `value` is allowed for `key` (a key of [`KEYS`]); the launcher keys use the
/// launcher's own check, so the two never disagree.
fn check(key: &str, value: &Value) -> Result<(), String> {
    let (_, field) = split(key);
    match key {
        "install.hosts" => hosts_from(value).map(|_| ()),
        "launcher.advisor" | "launcher.role_access" => {
            let mut one = Map::new();
            one.insert(field.to_string(), value.clone());
            if key == "launcher.advisor" {
                configured_advisor(Some(&one)).map(|_| ())
            } else {
                role_full_access(Some(&one)).map(|_| ())
            }
        }
        _ => Err("unknown setting key".into()),
    }
}

/// The stored value of `key`: `None` when the file, the section or the key is absent.
fn lookup<'a>(
    root: Option<&'a Map<String, Value>>,
    key: &str,
) -> Result<Option<&'a Value>, String> {
    let (section, field) = split(key);
    match root.and_then(|r| r.get(section)) {
        None => Ok(None),
        Some(Value::Object(o)) => Ok(o.get(field)),
        Some(_) => Err(format!(
            "user launcher configuration {section} must be an object"
        )),
    }
}

fn entry(root: Option<&Map<String, Value>>, key: &str) -> Value {
    let base = json!({"key": key, "description": describe(key), "default": default_value(key)});
    let mut out = base.as_object().cloned().unwrap_or_default();
    match lookup(root, key) {
        Ok(None) => {
            out.insert("value".into(), default_value(key));
            out.insert("source".into(), json!("default"));
            out.insert("valid".into(), json!(true));
        }
        Ok(Some(v)) => {
            out.insert("source".into(), json!("file"));
            match check(key, v) {
                Ok(()) => {
                    out.insert("value".into(), v.clone());
                    out.insert("valid".into(), json!(true));
                }
                Err(e) => {
                    out.insert("valid".into(), json!(false));
                    out.insert("error".into(), json!(e));
                }
            }
        }
        Err(e) => {
            out.insert("source".into(), json!("file"));
            out.insert("valid".into(), json!(false));
            out.insert("error".into(), json!(e));
        }
    }
    Value::Object(out)
}

/// How many keys the file holds that the table does not know (names are not shown).
fn unlisted(root: &Map<String, Value>) -> usize {
    let mut n = 0;
    for (section, value) in root {
        let known: Vec<&str> = KEYS
            .iter()
            .map(|k| split(k))
            .filter(|(s, _)| s == section)
            .map(|(_, f)| f)
            .collect();
        match (known.is_empty(), value) {
            (true, _) => n += 1,
            (false, Value::Object(o)) => {
                n += o.keys().filter(|f| !known.contains(&f.as_str())).count()
            }
            (false, _) => {}
        }
    }
    n
}

fn file_json(home: &Path, root: Option<&Map<String, Value>>) -> Value {
    json!({"path": user_config::path(home), "state": if root.is_some() { "present" } else { "absent" }})
}

/// `setting list`: every key with its value and where the value comes from.
pub fn list(home: &Path) -> Result<Value, String> {
    let root = user_config::read_root(home)?;
    let settings: Vec<Value> = KEYS.iter().map(|k| entry(root.as_ref(), k)).collect();
    Ok(json!({
        "file": file_json(home, root.as_ref()),
        "settings": settings,
        "unlisted_keys": root.as_ref().map_or(0, unlisted),
        "writes": false,
        "project_settings": "managed by /moai project, not by this command",
    }))
}

/// `setting get KEY`: one key; a key outside [`KEYS`] is refused.
pub fn get(home: &Path, key: &str) -> Result<Value, String> {
    if !KEYS.contains(&key) {
        return Err(format!(
            "unknown setting key; known keys: {}",
            KEYS.join(", ")
        ));
    }
    let root = user_config::read_root(home)?;
    Ok(json!({
        "file": file_json(home, root.as_ref()),
        "setting": entry(root.as_ref(), key),
        "writes": false,
    }))
}

/// The hosts the user-level install covers for `home`: `install.hosts` when set, else
/// [`DEFAULT_HOSTS`]. A broken file or value is an error and nothing is guessed.
pub fn selected_hosts(home: &Path) -> Result<Vec<&'static str>, String> {
    let root = user_config::read_root(home)?;
    match lookup(root.as_ref(), "install.hosts")? {
        None => Ok(DEFAULT_HOSTS.to_vec()),
        Some(v) => hosts_from(v),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::fs;

    fn home_with(text: Option<&str>) -> tempfile::TempDir {
        let home = tempfile::tempdir().unwrap();
        if let Some(t) = text {
            let path = user_config::path(home.path());
            fs::create_dir_all(path.parent().unwrap()).unwrap();
            fs::write(path, t).unwrap();
        }
        home
    }

    fn setting<'a>(v: &'a Value, key: &str) -> &'a Value {
        v["settings"]
            .as_array()
            .unwrap()
            .iter()
            .find(|s| s["key"] == key)
            .unwrap()
    }

    #[test]
    fn no_file_shows_the_defaults_and_selects_the_hosts_update_writes_today() {
        let home = home_with(None);
        let v = list(home.path()).unwrap();
        assert_eq!(v["file"]["state"], "absent");
        assert_eq!(v["writes"], false);
        for k in KEYS {
            assert_eq!(setting(&v, k)["source"], "default", "{k}");
        }
        assert_eq!(
            setting(&v, "install.hosts")["value"],
            json!(["claude", "codex"])
        );
        assert_eq!(setting(&v, "launcher.role_access")["value"], "full");
        assert_eq!(
            selected_hosts(home.path()).unwrap(),
            vec!["claude", "codex"]
        );
    }

    #[test]
    fn a_file_value_is_shown_with_its_source_and_hosts_come_back_in_a_fixed_order() {
        let home = home_with(Some(
            r#"{"install":{"hosts":["omp","claude"]},"launcher":{"role_access":"prompt","advisor":"opus"}}"#,
        ));
        let v = list(home.path()).unwrap();
        assert_eq!(v["file"]["state"], "present");
        assert_eq!(setting(&v, "launcher.role_access")["value"], "prompt");
        assert_eq!(setting(&v, "launcher.role_access")["source"], "file");
        assert_eq!(setting(&v, "launcher.advisor")["value"], "opus");
        assert_eq!(
            setting(&v, "install.hosts")["value"],
            json!(["omp", "claude"])
        );
        assert_eq!(selected_hosts(home.path()).unwrap(), vec!["claude", "omp"]);
        assert_eq!(v["unlisted_keys"], 0);
    }

    #[test]
    fn only_a_non_empty_set_of_known_hosts_each_once_is_valid() {
        for ok in [
            json!(["claude"]),
            json!(["omp"]),
            json!(["omp", "codex", "claude"]),
        ] {
            assert!(hosts_from(&ok).is_ok(), "{ok}");
        }
        for bad in [
            json!([]),
            json!(["claude", "claude"]),
            json!(["Claude"]),
            json!(["cursor"]),
            json!(["claude", 1]),
            json!("claude"),
            json!(null),
            json!(["claude", "codex", "omp", "omp"]),
        ] {
            assert!(hosts_from(&bad).is_err(), "{bad}");
        }
    }

    #[test]
    fn an_invalid_value_is_flagged_without_being_quoted_and_selection_refuses_it() {
        let secret = "secret-looking-value-9f3a";
        let home = home_with(Some(&format!(
            r#"{{"install":{{"hosts":["{secret}"]}},"launcher":{{"role_access":"{secret}"}}}}"#
        )));
        let v = list(home.path()).unwrap();
        for k in ["install.hosts", "launcher.role_access"] {
            let s = setting(&v, k);
            assert_eq!(s["valid"], false, "{k}");
            assert!(s.get("value").is_none(), "{k}");
        }
        assert!(!v.to_string().contains(secret), "{v}");
        assert!(selected_hosts(home.path()).is_err());
    }

    #[test]
    fn get_refuses_a_key_outside_the_table_and_other_content_is_only_counted() {
        let home = home_with(Some(
            r#"{"launcher":{"role_access":"full","token":"abc"},"other":{"x":1},"flat":true}"#,
        ));
        assert!(get(home.path(), "launcher.token").is_err());
        assert!(get(home.path(), "").is_err());
        assert!(get(home.path(), "install").is_err());
        let g = get(home.path(), "launcher.role_access").unwrap();
        assert_eq!(g["setting"]["value"], "full");
        let l = list(home.path()).unwrap();
        assert_eq!(l["unlisted_keys"], 3);
        assert!(!l.to_string().contains("abc"));
        assert!(!l.to_string().contains("other"));
    }

    #[test]
    fn a_broken_or_special_file_is_refused_and_a_section_must_be_an_object() {
        for text in ["not json", "[1]", "null"] {
            let home = home_with(Some(text));
            assert!(list(home.path()).is_err(), "{text}");
            assert!(selected_hosts(home.path()).is_err(), "{text}");
        }
        let home = home_with(Some(r#"{"install":"claude"}"#));
        assert!(selected_hosts(home.path()).is_err());
        let v = list(home.path()).unwrap();
        assert_eq!(setting(&v, "install.hosts")["valid"], false);
        // A folder where the file belongs.
        let home = tempfile::tempdir().unwrap();
        fs::create_dir_all(user_config::path(home.path())).unwrap();
        assert!(list(home.path()).is_err());
    }

    #[cfg(unix)]
    #[test]
    fn a_symlinked_file_is_refused() {
        let home = home_with(None);
        let path = user_config::path(home.path());
        fs::create_dir_all(path.parent().unwrap()).unwrap();
        let target = home.path().join("elsewhere.json");
        fs::write(&target, r#"{"install":{"hosts":["omp"]}}"#).unwrap();
        std::os::unix::fs::symlink(&target, &path).unwrap();
        assert!(list(home.path()).is_err());
        assert!(selected_hosts(home.path()).is_err());
    }

    #[test]
    fn listing_changes_nothing_under_home() {
        let home = home_with(Some(r#"{"install":{"hosts":["claude"]}}"#));
        let before = fs::read(user_config::path(home.path())).unwrap();
        list(home.path()).unwrap();
        get(home.path(), "install.hosts").unwrap();
        selected_hosts(home.path()).unwrap();
        assert_eq!(fs::read(user_config::path(home.path())).unwrap(), before);
    }
}
