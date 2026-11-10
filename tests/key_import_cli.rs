//! #2393 S3: `key-import-plan` / `key-import-apply` move a legacy GLM key file into the
//! encrypted vault (user decision, Issue #2393 comment 6778671862: `_shared` with `--shared`,
//! else the current project's folder).
//! Preview writes nothing and shows no value; apply needs the preview's digest and a
//! permission reference, reads back and compares, records a receipt without the value, and
//! leaves the legacy file byte- and time-identical. Temporary HOME and synthetic keys only.
#![cfg(unix)]

mod common;
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

/// Runs from HOME (not a git work tree) with `--shared`.
fn run(home: &Path, args: &[&str]) -> Output {
    run_in(home, home, &[args, &["--shared"]].concat())
}

fn run_in(home: &Path, cwd: &Path, args: &[&str]) -> Output {
    Command::new(BIN)
        .args(args)
        .current_dir(cwd)
        .env("HOME", home)
        .env("GIT_CEILING_DIRECTORIES", home)
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .output()
        .unwrap()
}

fn shared_entry(home: &Path) -> PathBuf {
    home.join(".moai-cli/vault/_shared/glm.key")
}

/// Sealed, private, and without the key text.
fn assert_sealed(entry: &Path) {
    let bytes = fs::read(entry).unwrap();
    assert!(bytes.starts_with(b"MOAIKV1\n"));
    for key in [KEY, RUST_KEY] {
        assert!(!bytes.windows(key.len()).any(|w| w == key.as_bytes()));
    }
    assert_eq!(
        fs::metadata(entry).unwrap().permissions().mode() & 0o777,
        0o600
    );
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
/// Every file under HOME: path → (bytes, mtime, mode). The mode is part of "unchanged"
/// (pre-audit N1: the legacy file's permissions are not touched either).
fn snapshot(home: &Path) -> BTreeMap<String, (Vec<u8>, std::time::SystemTime, u32)> {
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
                (
                    fs::read(&p).unwrap_or_default(),
                    m.modified().unwrap(),
                    m.permissions().mode(),
                ),
            );
        }
    }
    out
}

/// Pre-audit B1/B2 (S3): the permission reference is 1–128 bytes without control
/// characters. 128 passes, 129 is refused, a control character is refused; nothing is
/// written by a refusal. Literal sizes.
#[test]
fn the_permission_reference_bounds_are_kept() {
    for (reference, ok) in [
        ("r".repeat(128), true),
        ("r".repeat(129), false),
        ("ok\u{7}bell".to_string(), false),
        ("two\nlines".to_string(), false),
    ] {
        let (_d, h) = home(Some(&format!("GLM_API_KEY={KEY}\n")), None);
        let p = plan(&h);
        let a = apply(&h, p["plan_digest"].as_str().unwrap(), &reference);
        assert_eq!(a["ok"], ok, "{} bytes: {a}", reference.len());
        assert_eq!(shared_entry(&h).exists(), ok);
    }
}

/// Pre-audit B3 (S3): a legacy file of exactly 65,536 bytes is read; 65,537 is not
/// (`size_ok: false`, no key, nothing to import). Literal sizes.
#[test]
fn a_legacy_file_is_read_up_to_64_kib() {
    let line = format!("GLM_API_KEY={KEY}\n");
    let mut body = line.clone().into_bytes();
    body.resize(65_536, b'#');
    let (_d, h) = home(None, None);
    fs::create_dir_all(h.join(".moai")).unwrap();
    fs::write(h.join(".moai/.env.glm"), &body).unwrap();
    let p = plan(&h);
    assert_eq!(p["sources"][0]["size_ok"], true, "{p}");
    assert_eq!(p["sources"][0]["has_key"], true);
    assert_eq!(p["action"], "import");
    body.push(b'#');
    fs::write(h.join(".moai/.env.glm"), &body).unwrap();
    let p = plan(&h);
    assert_eq!(p["sources"][0]["size_ok"], false, "{p}");
    assert_eq!(p["sources"][0]["has_key"], false);
    assert_eq!(p["action"], "nothing_to_import");
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
    // The vault holds the key, sealed and private; it decrypts (status present).
    assert_eq!(p["folder"], "_shared");
    assert_sealed(&shared_entry(&h));
    let s = envelope(&run(&h, &["key-status", "glm"]));
    assert_eq!(s["result"]["entry"], "present", "{s}");
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
    assert_eq!(a["result"]["matched"], true);
    assert_sealed(&shared_entry(&h));
}

/// Without `--shared` the key goes to the working directory's project folder; outside a git
/// work tree that is refused before anything is read or written.
#[test]
fn the_destination_is_the_project_folder_without_shared() {
    let (_d, h) = home(Some(&format!("GLM_API_KEY={KEY}\n")), None);
    let out = envelope(&run_in(&h, &h, &["key-import-plan", "glm"]));
    assert_eq!(out["ok"], false, "{out}");
    let repo = h.join("proj");
    fs::create_dir(&repo).unwrap();
    let init = common::git::command()
        .args(["init", "-q"])
        .current_dir(&repo)
        .status()
        .unwrap();
    assert!(init.success());
    let p = envelope(&run_in(&h, &repo, &["key-import-plan", "glm"]))["result"].clone();
    assert_eq!(p["scope"], "project", "{p}");
    let folder = p["folder"].as_str().unwrap().to_string();
    assert!(folder.starts_with("proj-"), "{folder}");
    let a = envelope(&run_in(
        &h,
        &repo,
        &[
            "key-import-apply",
            "glm",
            "--plan-digest",
            p["plan_digest"].as_str().unwrap(),
            "--permission-reference",
            "user-ok",
        ],
    ));
    assert_eq!(a["ok"], true, "{a}");
    assert_sealed(&h.join(".moai-cli/vault").join(&folder).join("glm.key"));
    assert!(!shared_entry(&h).exists());
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
    // An unusable vault entry (a plain-text file from before the format) blocks instead of
    // being overwritten.
    let (_d4, h4) = home(Some(&format!("GLM_API_KEY={KEY}\n")), None);
    fs::create_dir_all(h4.join(".moai-cli/vault/_shared")).unwrap();
    for dir in [".moai-cli/vault", ".moai-cli/vault/_shared"] {
        fs::set_permissions(h4.join(dir), fs::Permissions::from_mode(0o700)).unwrap();
    }
    fs::write(shared_entry(&h4), "x\n").unwrap();
    fs::set_permissions(shared_entry(&h4), fs::Permissions::from_mode(0o600)).unwrap();
    let p = plan(&h4);
    assert_eq!(p["action"], "blocked", "{p}");
    assert_eq!(
        apply(&h4, p["plan_digest"].as_str().unwrap(), "ok")["ok"],
        false
    );
    assert_eq!(fs::read_to_string(shared_entry(&h4)).unwrap(), "x\n");
}

#[test]
fn only_glm_has_legacy_files() {
    let (_d, h) = home(None, None);
    let v = envelope(&run(&h, &["key-import-plan", "other"]));
    assert_eq!(v["ok"], false);
}
