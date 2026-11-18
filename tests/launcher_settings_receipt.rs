//! SPEC-20261119-launcher-settings-receipt (card C-47 slice 4): the receipt the model stage leaves
//! when it saves the launcher settings, and the revert that puts the earlier file back. Every
//! file lives in a temporary HOME; the real HOME is never read or written.

#![cfg(unix)]

use moai_cli::home_root;
use moai_cli::update_models::{revert, settings_path, stage};
use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::fs;
use std::io::Cursor;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use tempfile::TempDir;

const REF: &str = "user-pick:update-models";
/// A text that must never appear in a result or a printed line.
const MARKER: &str = "BEFORE-MARKER-TEXT";

struct Home {
    _tmp: TempDir,
    path: PathBuf,
}

fn home() -> Home {
    let tmp = TempDir::new().unwrap();
    let path = fs::canonicalize(tmp.path()).unwrap().join("home");
    fs::create_dir_all(path.join(".claude")).unwrap();
    Home { _tmp: tmp, path }
}

fn write_settings(home: &Path, body: &str, mode: u32) -> PathBuf {
    let path = settings_path(home);
    fs::create_dir_all(path.parent().unwrap()).unwrap();
    fs::write(&path, body).unwrap();
    fs::set_permissions(&path, fs::Permissions::from_mode(mode)).unwrap();
    path
}

fn receipts(home: &Path) -> PathBuf {
    home_root::path(home).join("receipts")
}

fn sha(text: &str) -> String {
    format!("{:x}", Sha256::digest(text.as_bytes()))
}

/// Saves the claude defaults (tool pick `1`, four Enters, a yes).
fn save(home: &Path) -> (Value, String) {
    let mut out = Vec::new();
    let result = stage(
        home,
        &mut Cursor::new("1\n\n\n\n\ny\n".to_string()),
        &mut out,
    );
    (result, String::from_utf8(out).unwrap())
}

fn receipt_file(home: &Path, id: &str) -> PathBuf {
    receipts(home).join(format!("{id}.json"))
}

fn names(dir: &Path) -> Vec<String> {
    let mut found: Vec<String> = fs::read_dir(dir)
        .map(|entries| {
            entries
                .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
                .collect()
        })
        .unwrap_or_default();
    found.sort();
    found
}

#[test]
fn a_save_leaves_a_receipt_with_the_text_before_and_no_text_in_the_result() {
    let h = home();
    let before = format!("{{ \"note\" : \"{MARKER}\" ,\"launcher\":{{}} }}\n");
    let path = write_settings(&h.path, &before, 0o600);
    let (result, out) = save(&h.path);
    assert_eq!(result["state"], json!("applied"), "{result}");
    let id = result["receipt"].as_str().unwrap().to_string();
    assert!(id.starts_with("models-"), "{id}");

    let file = receipt_file(&h.path, &id);
    let receipt: Value = serde_json::from_str(&fs::read_to_string(&file).unwrap()).unwrap();
    assert_eq!(receipt["operation"], json!("update.models"));
    assert_eq!(receipt["path"], json!(path.display().to_string()));
    assert_eq!(receipt["before_text"], json!(before));
    assert_eq!(receipt["before_sha256"], json!(sha(&before)));
    let after = fs::read_to_string(&path).unwrap();
    assert_ne!(after, before);
    assert_eq!(receipt["after_sha256"], json!(sha(&after)));
    assert_eq!(receipt["permission_reference"], json!(REF));
    assert_eq!(
        fs::metadata(&file).unwrap().permissions().mode() & 0o777,
        0o600
    );
    assert_eq!(
        fs::metadata(receipts(&h.path))
            .unwrap()
            .permissions()
            .mode()
            & 0o777,
        0o700
    );
    assert!(
        !result.to_string().contains(MARKER) && !out.contains(MARKER),
        "no file text in the result or the printed lines"
    );
}

#[test]
fn reverting_a_save_into_an_absent_file_removes_it() {
    let h = home();
    let path = settings_path(&h.path);
    assert!(!path.exists());
    let (result, _) = save(&h.path);
    let id = result["receipt"].as_str().unwrap().to_string();
    assert!(path.is_file());
    let receipt: Value =
        serde_json::from_str(&fs::read_to_string(receipt_file(&h.path, &id)).unwrap()).unwrap();
    assert_eq!(receipt["before_text"], Value::Null, "{receipt}");
    assert_eq!(receipt["before_sha256"], Value::Null);

    let reverted = revert(&h.path, &id, REF).unwrap();
    assert_eq!(reverted["state"], json!("reverted"), "{reverted}");
    assert!(!path.exists(), "the file the save created is removed");
}

#[test]
fn the_revert_restores_the_earlier_bytes_and_mode_and_marks_the_receipt() {
    let h = home();
    let before = "{ \"launcher\" : { \"role_access\":\"prompt\" } ,\n\t\"keep\":[1,2,\n3] }";
    let path = write_settings(&h.path, before, 0o640);
    let (result, _) = save(&h.path);
    let id = result["receipt"].as_str().unwrap().to_string();
    assert_ne!(fs::read_to_string(&path).unwrap(), before);
    assert_eq!(
        fs::metadata(&path).unwrap().permissions().mode() & 0o777,
        0o640
    );

    let reverted = revert(&h.path, &id, REF).unwrap();
    assert_eq!(reverted["state"], json!("reverted"), "{reverted}");
    assert_eq!(reverted["receipt"], json!(id));
    assert!(
        !reverted.to_string().contains("role_access"),
        "no file text"
    );
    assert_eq!(fs::read_to_string(&path).unwrap(), before, "byte for byte");
    assert_eq!(
        fs::metadata(&path).unwrap().permissions().mode() & 0o777,
        0o640,
        "the mode is kept"
    );
    assert!(
        receipts(&h.path)
            .join(format!("{id}.reverted.json"))
            .is_file(),
        "the marker is written"
    );
}

#[test]
fn a_file_changed_after_the_save_is_not_reverted() {
    let h = home();
    let before = r#"{"launcher":{}}"#;
    let path = write_settings(&h.path, before, 0o600);
    let (result, _) = save(&h.path);
    let id = result["receipt"].as_str().unwrap().to_string();
    let saved = fs::read_to_string(&path).unwrap();

    let changed = format!("{saved} ");
    fs::write(&path, &changed).unwrap();
    let skipped = revert(&h.path, &id, REF).unwrap();
    assert_eq!(skipped["state"], json!("skipped"), "{skipped}");
    assert_eq!(skipped["reason"], json!("changed_since_apply"));
    assert_eq!(fs::read_to_string(&path).unwrap(), changed, "left alone");
    assert!(
        !receipts(&h.path)
            .join(format!("{id}.reverted.json"))
            .exists(),
        "no marker after a skip"
    );

    // The person undoes their own change, and the revert works.
    fs::write(&path, &saved).unwrap();
    let reverted = revert(&h.path, &id, REF).unwrap();
    assert_eq!(reverted["state"], json!("reverted"), "{reverted}");
    assert_eq!(fs::read_to_string(&path).unwrap(), before);
}

#[test]
fn the_revert_refuses_what_it_cannot_prove_is_its_own() {
    let h = home();
    let before = r#"{"launcher":{}}"#;
    let path = write_settings(&h.path, before, 0o600);
    let (result, _) = save(&h.path);
    let id = result["receipt"].as_str().unwrap().to_string();
    let saved = fs::read_to_string(&path).unwrap();
    let dir = receipts(&h.path);

    // A receipt made by hand: private file in the private folder.
    let plant = |name: &str, value: Value| {
        let file = dir.join(format!("{name}.json"));
        fs::write(&file, value.to_string()).unwrap();
        fs::set_permissions(&file, fs::Permissions::from_mode(0o600)).unwrap();
    };
    let genuine: Value =
        serde_json::from_str(&fs::read_to_string(receipt_file(&h.path, &id)).unwrap()).unwrap();
    let mut other_operation = genuine.clone();
    other_operation["operation"] = json!("update.keys");
    plant("models-1-1", other_operation);
    let mut other_path = genuine.clone();
    other_path["path"] = json!("/somewhere/else/launcher.yaml");
    plant("models-2-2", other_path);

    let refusals: Vec<(&str, &str, &str)> = vec![
        (&id, "", "an empty permission reference"),
        (&id, "placeholder", "a placeholder permission reference"),
        ("../models-1-1", REF, "an id with a path"),
        ("keys-1-1", REF, "an id of another receipt kind"),
        ("models-", REF, "an empty id body"),
        ("models-9-9", REF, "a receipt that does not exist"),
        ("models-1-1", REF, "a receipt of another operation"),
        ("models-2-2", REF, "a receipt for another settings file"),
    ];
    for (given, reference, why) in refusals {
        assert!(revert(&h.path, given, reference).is_err(), "{why}");
        assert_eq!(
            fs::read_to_string(&path).unwrap(),
            saved,
            "{why}: unchanged"
        );
    }

    // A second revert of the same receipt is refused.
    revert(&h.path, &id, REF).unwrap();
    let again = revert(&h.path, &id, REF).unwrap_err();
    assert!(again.contains("already reverted"), "{again}");
    assert_eq!(fs::read_to_string(&path).unwrap(), before);
}

#[test]
fn no_receipt_means_no_write() {
    // A file whose receipt would pass 60000 bytes cannot be kept: nothing is written.
    let h = home();
    let padding = "x".repeat(70_000);
    let big = format!("{{\"pad\":\"{padding}\",\"launcher\":{{}}}}");
    let path = write_settings(&h.path, &big, 0o600);
    let (result, _) = save(&h.path);
    assert_eq!(result["state"], json!("error"), "{result}");
    assert!(
        result["message"].as_str().unwrap().contains("60000"),
        "{result}"
    );
    assert_eq!(fs::read_to_string(&path).unwrap(), big);
    assert!(names(&receipts(&h.path)).is_empty(), "no receipt");

    // A receipts place that is a file: no receipt can be made, so nothing is written.
    let h = home();
    let before = r#"{"launcher":{}}"#;
    let path = write_settings(&h.path, before, 0o600);
    fs::write(receipts(&h.path), "not a folder").unwrap();
    let (result, _) = save(&h.path);
    assert_eq!(result["state"], json!("error"), "{result}");
    assert_eq!(
        fs::read_to_string(&path).unwrap(),
        before,
        "nothing written"
    );
}

#[test]
fn a_no_and_an_unchanged_run_leave_no_receipt() {
    let h = home();
    let mut out = Vec::new();
    let no = stage(
        &h.path,
        &mut Cursor::new("1\n\n\n\n\nn\n".to_string()),
        &mut out,
    );
    assert_eq!(no["state"], json!("declined"), "{no}");
    assert!(names(&receipts(&h.path)).is_empty());

    let (first, _) = save(&h.path);
    assert_eq!(first["state"], json!("applied"), "{first}");
    assert_eq!(names(&receipts(&h.path)).len(), 1);
    let (again, _) = save(&h.path);
    assert_eq!(again["state"], json!("unchanged"), "{again}");
    assert_eq!(names(&receipts(&h.path)).len(), 1, "no second receipt");
}

#[test]
fn a_failed_restore_changes_nothing_and_stays_revertable() {
    let h = home();
    let before = r#"{"launcher":{}}"#;
    let path = write_settings(&h.path, before, 0o600);
    let (result, _) = save(&h.path);
    let id = result["receipt"].as_str().unwrap().to_string();
    let saved = fs::read_to_string(&path).unwrap();
    let dir = path.parent().unwrap().to_path_buf();

    fs::set_permissions(&dir, fs::Permissions::from_mode(0o555)).unwrap();
    let probe = dir.join(".probe");
    let can_write_anyway = fs::write(&probe, "").is_ok(); // running as the superuser
    let _ = fs::remove_file(&probe);
    let failed = revert(&h.path, &id, REF);
    fs::set_permissions(&dir, fs::Permissions::from_mode(0o755)).unwrap();
    if can_write_anyway {
        return;
    }
    assert!(failed.is_err(), "{failed:?}");
    assert_eq!(fs::read_to_string(&path).unwrap(), saved, "unchanged");
    assert!(
        !receipts(&h.path)
            .join(format!("{id}.reverted.json"))
            .exists(),
        "no marker"
    );
    let reverted = revert(&h.path, &id, REF).unwrap();
    assert_eq!(reverted["state"], json!("reverted"), "{reverted}");
    assert_eq!(fs::read_to_string(&path).unwrap(), before);
}

/// Saves into a file that exists and returns the settings path, the earlier text and the receipt.
fn genuine_save(h: &Home) -> (PathBuf, String, String, Value) {
    let before = r#"{"launcher":{"role_access":"prompt"}}"#.to_string();
    let path = write_settings(&h.path, &before, 0o600);
    let (result, _) = save(&h.path);
    let id = result["receipt"].as_str().unwrap().to_string();
    let receipt: Value =
        serde_json::from_str(&fs::read_to_string(receipt_file(&h.path, &id)).unwrap()).unwrap();
    (path, before, id, receipt)
}

fn plant(home: &Path, name: &str, value: &Value) {
    let file = receipts(home).join(format!("{name}.json"));
    fs::write(&file, value.to_string()).unwrap();
    fs::set_permissions(&file, fs::Permissions::from_mode(0o600)).unwrap();
}

#[test]
fn a_receipt_that_lacks_a_before_field_is_damaged_and_deletes_nothing() {
    let h = home();
    let (path, _before, _id, genuine) = genuine_save(&h);
    let saved = fs::read_to_string(&path).unwrap();

    let without = |keys: &[&str]| {
        let mut r = genuine.clone();
        for k in keys {
            r.as_object_mut().unwrap().remove(*k);
        }
        r
    };
    let with = |key: &str, value: Value| {
        let mut r = genuine.clone();
        r[key] = value;
        r
    };
    let cases: Vec<(&str, Value)> = vec![
        ("no before_text", without(&["before_text"])),
        ("no before_sha256", without(&["before_sha256"])),
        (
            "neither before field",
            without(&["before_text", "before_sha256"]),
        ),
        (
            "absent text, present hash",
            with("before_text", Value::Null),
        ),
        (
            "present text, absent hash",
            with("before_sha256", Value::Null),
        ),
        ("text that is a number", with("before_text", json!(7))),
        ("hash that is a number", with("before_sha256", json!(7))),
        (
            "hash that is not the text's",
            with("before_sha256", json!(sha("other"))),
        ),
    ];
    for (i, (why, receipt)) in cases.iter().enumerate() {
        let name = format!("models-{}-1", 100 + i);
        plant(&h.path, &name, receipt);
        let refused = revert(&h.path, &name, REF).unwrap_err();
        assert_eq!(refused, "models receipt is damaged", "{why}");
        assert_eq!(
            fs::read_to_string(&path).unwrap(),
            saved,
            "{why}: unchanged"
        );
        assert!(
            !receipts(&h.path)
                .join(format!("{name}.reverted.json"))
                .exists(),
            "{why}: no marker"
        );
    }
}

#[test]
fn a_failed_apply_still_names_its_receipt() {
    let h = home();
    let before = r#"{"launcher":{}}"#;
    let path = write_settings(&h.path, before, 0o600);
    let dir = path.parent().unwrap().to_path_buf();
    fs::set_permissions(&dir, fs::Permissions::from_mode(0o555)).unwrap();
    let probe = dir.join(".probe");
    let can_write_anyway = fs::write(&probe, "").is_ok(); // running as the superuser
    let _ = fs::remove_file(&probe);
    let (result, out) = save(&h.path);
    fs::set_permissions(&dir, fs::Permissions::from_mode(0o755)).unwrap();
    if can_write_anyway {
        return;
    }
    assert_eq!(result["state"], json!("error"), "{result}");
    let id = result["receipt"].as_str().expect("the receipt id is named");
    assert!(receipt_file(&h.path, id).is_file(), "the receipt exists");
    assert_eq!(
        fs::read_to_string(&path).unwrap(),
        before,
        "nothing written"
    );
    assert!(!result.to_string().contains("role_access") && !out.contains("role_access"));
}

#[test]
fn a_receipt_id_outside_the_shape_is_refused_as_unknown() {
    let h = home();
    let (path, _before, id, _receipt) = genuine_save(&h);
    let saved = fs::read_to_string(&path).unwrap();
    let long = format!("models-{}", "1".repeat(65));
    let bad = [
        format!("{id}.reverted"),
        format!("{id}/"),
        format!("{id}\n"),
        "models-1 1".to_string(),
        "models-1_1".to_string(),
        "MODELS-1-1".to_string(),
        "models-\u{661}-1".to_string(),
        "models-1/../1".to_string(),
        long,
    ];
    for given in bad {
        let refused = revert(&h.path, &given, REF).unwrap_err();
        assert_eq!(refused, "unknown models receipt id", "{given:?}");
    }
    assert_eq!(fs::read_to_string(&path).unwrap(), saved, "unchanged");
}
