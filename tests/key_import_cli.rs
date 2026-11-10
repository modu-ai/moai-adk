//! #2393 S3: `key-import-plan` / `key-import-apply` move a legacy GLM key file into the vault.
//! Preview writes nothing and shows no value; apply needs the preview's digest and a
//! permission reference, reads back and compares, records a receipt without the value, and
//! leaves the legacy file byte- and time-identical. Temporary HOME and synthetic keys only.
#![cfg(unix)]

use serde_json::Value;
use std::collections::BTreeMap;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Output};

const BIN: &str = env!("CARGO_BIN_EXE_moai-cli");
const KEY: &str = "moai-test-import-key-not-real";
const RUST_KEY: &str = "moai-test-import-rust-key-not-real";

fn home(go: Option<&str>, rust: Option<&str>) -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::tempdir().unwrap();
    let h = fs::canonicalize(dir.path()).unwrap();
    for (rel, text) in [(".moai/.env.glm", go), (".moai-cli/.env.glm", rust)] {
        if let Some(text) = text {
            fs::create_dir_all(h.join(rel).parent().unwrap()).unwrap();
            fs::write(h.join(rel), text).unwrap();
        }
    }
    (dir, h)
}

fn run(home: &Path, args: &[&str]) -> Output {
    Command::new(BIN)
        .args(args)
        .env("HOME", home)
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .output()
        .unwrap()
}

fn envelope(out: &Output) -> Value {
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    for key in [KEY, RUST_KEY] {
        assert!(!text.contains(key), "key text in output");
    }
    serde_json::from_slice(&out.stdout).unwrap_or_else(|e| panic!("{e}: {text}"))
}

/// Every file under HOME: path → (bytes, mtime).
fn snapshot(home: &Path) -> BTreeMap<String, (Vec<u8>, std::time::SystemTime)> {
    let mut out = BTreeMap::new();
    let mut stack = vec![home.to_path_buf()];
    while let Some(p) = stack.pop() {
        let m = fs::symlink_metadata(&p).unwrap();
        if m.is_dir() {
            for e in fs::read_dir(&p).unwrap() {
                stack.push(e.unwrap().path());
            }
        } else {
            out.insert(
                p.to_string_lossy().into_owned(),
                (fs::read(&p).unwrap_or_default(), m.modified().unwrap()),
            );
        }
    }
    out
}

fn plan(home: &Path) -> Value {
    let v = envelope(&run(home, &["key-import-plan", "glm"]));
    assert_eq!(v["ok"], true, "{v}");
    v["result"].clone()
}

fn apply(home: &Path, digest: &str, permission: &str) -> Value {
    envelope(&run(
        home,
        &[
            "key-import-apply",
            "glm",
            "--plan-digest",
            digest,
            "--permission-reference",
            permission,
        ],
    ))
}

#[test]
fn preview_writes_nothing_and_apply_copies_keeping_the_original() {
    let (_d, h) = home(Some(&format!("export GLM_API_KEY=\"{KEY}\"\n")), None);
    let before = snapshot(&h);
    let p = plan(&h);
    assert_eq!(snapshot(&h), before, "preview writes nothing");
    assert_eq!(p["action"], "import", "{p}");
    assert_eq!(
        p["source"],
        h.join(".moai/.env.glm").to_string_lossy().as_ref()
    );
    assert_eq!(p["sources"][0]["has_key"], true);
    assert_eq!(p["sources"][1]["exists"], false);
    assert_eq!(p["destination"]["entry"], "absent");
    assert_eq!(p["original_kept"], true);
    assert_eq!(p["writes"], false);

    let a = apply(&h, p["plan_digest"].as_str().unwrap(), "user-ok-2393");
    assert_eq!(a["ok"], true, "{a}");
    assert_eq!(a["result"]["matched"], true);
    // The vault holds the key, private.
    let entry = h.join(".moai-cli/vault/glm.key");
    assert_eq!(fs::read_to_string(&entry).unwrap(), format!("{KEY}\n"));
    assert_eq!(
        fs::metadata(&entry).unwrap().permissions().mode() & 0o777,
        0o600
    );
    // The legacy file is byte- and time-identical.
    let go = h.join(".moai/.env.glm").to_string_lossy().into_owned();
    assert_eq!(snapshot(&h)[&go], before[&go]);
    // The receipt names paths and the result, never the value.
    let receipt = PathBuf::from(a["result"]["receipt"].as_str().unwrap());
    let r: Value = serde_json::from_slice(&fs::read(&receipt).unwrap()).unwrap();
    assert_eq!(r["matched"], true);
    assert_eq!(r["permission_reference"], "user-ok-2393");
    assert!(!fs::read_to_string(&receipt).unwrap().contains(KEY));
    assert_eq!(
        fs::metadata(&receipt).unwrap().permissions().mode() & 0o777,
        0o600
    );
    // Done once: the next preview has nothing left to do.
    assert_eq!(plan(&h)["action"], "already_in_vault");
}

#[test]
fn apply_needs_the_preview_digest_and_a_permission_reference() {
    let (_d, h) = home(Some(&format!("GLM_API_KEY={KEY}\n")), None);
    let p = plan(&h);
    let digest = p["plan_digest"].as_str().unwrap().to_string();
    for permission in ["", "null", "placeholder", "  "] {
        let a = apply(&h, &digest, permission);
        assert_eq!(a["ok"], false, "{permission:?}");
    }
    let a = apply(&h, "sha256:0000", "user-ok");
    assert_eq!(a["ok"], false);
    // The legacy file changed after the preview: the old digest no longer applies.
    fs::write(h.join(".moai/.env.glm"), format!("GLM_API_KEY={KEY}-new\n")).unwrap();
    let a = apply(&h, &digest, "user-ok");
    assert_eq!(a["ok"], false, "{a}");
    assert!(a.to_string().contains("digest mismatch"), "{a}");
    assert!(
        !h.join(".moai-cli/vault").exists(),
        "nothing written by a refused apply"
    );
}

#[test]
fn the_launcher_order_picks_the_go_file_first_then_the_rust_one() {
    let (_d, h) = home(
        Some("GLM_API_KEY=\n"),
        Some(&format!("ANTHROPIC_AUTH_TOKEN={RUST_KEY}\n")),
    );
    let p = plan(&h);
    assert_eq!(p["sources"][0]["has_key"], false);
    assert_eq!(
        p["source"],
        h.join(".moai-cli/.env.glm").to_string_lossy().as_ref()
    );
    let a = apply(&h, p["plan_digest"].as_str().unwrap(), "user-ok");
    assert_eq!(a["ok"], true, "{a}");
    assert_eq!(
        fs::read_to_string(h.join(".moai-cli/vault/glm.key")).unwrap(),
        format!("{RUST_KEY}\n")
    );
}

#[test]
fn nothing_to_import_and_blocked_states_write_nothing() {
    // No legacy key at all.
    let (_d, h) = home(Some("OTHER=1\n"), None);
    let p = plan(&h);
    assert_eq!(p["action"], "nothing_to_import");
    assert_eq!(
        apply(&h, p["plan_digest"].as_str().unwrap(), "ok")["ok"],
        false
    );
    // A symlinked legacy file is not followed.
    let (_d2, h2) = home(None, None);
    fs::write(h2.join("elsewhere.env"), format!("GLM_API_KEY={KEY}\n")).unwrap();
    fs::create_dir_all(h2.join(".moai")).unwrap();
    std::os::unix::fs::symlink(h2.join("elsewhere.env"), h2.join(".moai/.env.glm")).unwrap();
    let p = plan(&h2);
    assert_eq!(p["sources"][0]["regular_file"], false, "{p}");
    assert_eq!(p["action"], "nothing_to_import");
    // An OS store backend is selected.
    let (_d3, h3) = home(Some(&format!("GLM_API_KEY={KEY}\n")), None);
    fs::create_dir_all(h3.join(".moai-cli")).unwrap();
    fs::write(h3.join(".moai-cli/config"), "key_store = \"keyring\"\n").unwrap();
    let p = plan(&h3);
    assert_eq!(p["action"], "blocked");
    assert!(p["reason"]
        .as_str()
        .unwrap()
        .starts_with("key_store_unavailable"));
    assert_eq!(
        apply(&h3, p["plan_digest"].as_str().unwrap(), "ok")["ok"],
        false
    );
    assert!(!h3.join(".moai-cli/vault").exists());
    // An unusable vault entry blocks instead of being overwritten.
    let (_d4, h4) = home(Some(&format!("GLM_API_KEY={KEY}\n")), None);
    fs::create_dir_all(h4.join(".moai-cli/vault")).unwrap();
    fs::set_permissions(
        h4.join(".moai-cli/vault"),
        fs::Permissions::from_mode(0o700),
    )
    .unwrap();
    fs::write(h4.join(".moai-cli/vault/glm.key"), "x\n").unwrap();
    fs::set_permissions(
        h4.join(".moai-cli/vault/glm.key"),
        fs::Permissions::from_mode(0o644),
    )
    .unwrap();
    let p = plan(&h4);
    assert_eq!(p["action"], "blocked", "{p}");
    assert_eq!(
        apply(&h4, p["plan_digest"].as_str().unwrap(), "ok")["ok"],
        false
    );
    assert_eq!(
        fs::read_to_string(h4.join(".moai-cli/vault/glm.key")).unwrap(),
        "x\n"
    );
}

#[test]
fn only_glm_has_legacy_files() {
    let (_d, h) = home(None, None);
    let v = envelope(&run(&h, &["key-import-plan", "other"]));
    assert_eq!(v["ok"], false);
}
