//! #2393: the `glm` launcher reads the key from the variable, then the encrypted vault — the
//! current project's folder, then `_shared` (user decision, Issue #2393 comment 6778671862).
//! A legacy key file is never read (6778150969): with no vault entry its presence ends
//! `blocked: legacy_key_only`. The source is reported by name (`key_source`), never the
//! value. An entry that is present but unsafe or does not decrypt stops the launch instead of
//! falling back. Temporary HOME, synthetic keys only; entries are made by the binary's own
//! `key-set`, so they carry the real format.
#![cfg(unix)]
mod common;
use serde_json::Value;
use std::fs;
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::sync::OnceLock;

const PROJECT_KEY: &str = "moai-test-project-key-not-real";
const SHARED_KEY: &str = "moai-test-shared-key-not-real";
const LEGACY_KEY: &str = "moai-test-legacy-key-not-real";
const ENV_KEY: &str = "moai-test-env-key-not-real";

/// #2655: one fresh git project per test process — cargo-nextest runs every test as its own
/// process, and a shared fixed folder made their `git init` runs collide on `.git/config`'s
/// lock. A `static` is never dropped, so the folder stays; it is made under
/// `CARGO_TARGET_TMPDIR` so what stays is inside `target/` (cleared by `cargo clean`), never in
/// the system temp folder (audit 6847550852 B1).
fn project() -> &'static Path {
    static PROJECT: OnceLock<tempfile::TempDir> = OnceLock::new();
    PROJECT
        .get_or_init(|| {
            let p = tempfile::Builder::new()
                .prefix("launcher-key-vault-project-")
                .tempdir_in(env!("CARGO_TARGET_TMPDIR"))
                .unwrap();
            let init = common::git::command()
                .args(["init", "-q"])
                .current_dir(p.path())
                .status()
                .unwrap();
            assert!(init.success());
            p
        })
        .path()
}

fn bin(home: &Path) -> Command {
    let mut cmd = Command::new(crate::common::MOAI_BIN);
    cmd.current_dir(project())
        .env("HOME", home)
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove("ANTHROPIC_AUTH_TOKEN");
    cmd
}

/// Stores `value` for the test project, or in `_shared`, with the binary's `key-set`; returns
/// the entry path.
fn store(home: &Path, shared: bool, value: &str) -> PathBuf {
    let mut cmd = bin(home);
    cmd.args(["key-set", "glm"]);
    if shared {
        cmd.arg("--shared");
    }
    let mut child = cmd
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(value.as_bytes())
        .unwrap();
    let out = child.wait_with_output().unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    assert_eq!(v["ok"], true, "{v}");
    PathBuf::from(v["result"]["path"].as_str().unwrap())
}

fn home(legacy: bool) -> tempfile::TempDir {
    let dir = tempfile::tempdir().unwrap();
    if legacy {
        let h = dir.path();
        fs::create_dir_all(h.join(".moai")).unwrap();
        fs::write(
            h.join(".moai/.env.glm"),
            format!("GLM_API_KEY={LEGACY_KEY}\n"),
        )
        .unwrap();
    }
    dir
}

/// `(exit code, stdout JSON, all output)` of `glm -a --dry-run`; no key text may appear.
fn dry(home: &Path, env_key: Option<&str>) -> (i32, Value, String) {
    let mut cmd = bin(home);
    if let Some(k) = env_key {
        cmd.env("ANTHROPIC_AUTH_TOKEN", k);
    }
    dry_command(cmd)
}

fn dry_command(mut cmd: Command) -> (i32, Value, String) {
    let out = cmd
        .args(["glm", "-a", "--name", "vault-test", "--dry-run"])
        .output()
        .unwrap();
    let stdout = String::from_utf8_lossy(&out.stdout).to_string();
    let text = format!("{stdout}{}", String::from_utf8_lossy(&out.stderr));
    for key in [PROJECT_KEY, SHARED_KEY, LEGACY_KEY, ENV_KEY] {
        assert!(!text.contains(key), "key text in launcher output");
    }
    (
        out.status.code().unwrap_or(-1),
        serde_json::from_str(&stdout).unwrap_or(Value::Null),
        text,
    )
}

/// The token the host process receives (a fake `claude` on PATH records it).
fn host_token(home: &Path) -> String {
    let bin_dir = tempfile::tempdir().unwrap();
    let seen = bin_dir.path().join("seen");
    let fake = bin_dir.path().join("claude");
    fs::write(
        &fake,
        format!(
            "#!/bin/sh\nprintf %s \"$ANTHROPIC_AUTH_TOKEN|$Z_AI_API_KEY\" > '{}'\n",
            seen.display()
        ),
    )
    .unwrap();
    fs::set_permissions(&fake, fs::Permissions::from_mode(0o755)).unwrap();
    let out = bin(home)
        .env(
            "PATH",
            format!("{}:/usr/bin:/bin", bin_dir.path().display()),
        )
        .args(["glm", "-a"])
        .output()
        .unwrap();
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(out.status.success(), "{text}");
    for key in [PROJECT_KEY, SHARED_KEY, LEGACY_KEY] {
        assert!(!text.contains(key));
    }
    let record = fs::read_to_string(&seen).unwrap();
    let (a, b) = record.split_once('|').unwrap();
    assert_eq!(a, b, "both names carry the same key");
    a.to_string()
}

/// Decision 4: variable, then project, then shared; the legacy file is never reached.
#[test]
fn the_variable_then_the_project_then_the_shared_folder() {
    let h = home(true);
    store(h.path(), true, SHARED_KEY);
    let (code, v, text) = dry(h.path(), None);
    assert_eq!(code, 0, "{text}");
    assert_eq!(v["key_source"], "vault:shared:glm", "{text}");
    assert_eq!(host_token(h.path()), SHARED_KEY);

    store(h.path(), false, PROJECT_KEY);
    let (code, v, text) = dry(h.path(), None);
    assert_eq!(code, 0, "{text}");
    assert_eq!(v["key_source"], "vault:project:glm", "{text}");
    assert_eq!(host_token(h.path()), PROJECT_KEY);

    let (code, v, text) = dry(h.path(), Some(ENV_KEY));
    assert_eq!(code, 0, "{text}");
    assert_eq!(v["key_source"], "env:ANTHROPIC_AUTH_TOKEN");
}

#[test]
fn without_a_vault_entry_the_legacy_file_is_named_not_read() {
    let h = home(true);
    let file = h.path().join(".moai/.env.glm");
    let before = fs::read(&file).unwrap();
    let mtime = fs::metadata(&file).unwrap().modified().unwrap();
    let (code, v, text) = dry(h.path(), None);
    assert_eq!(code, 1, "{text}");
    assert_eq!(v, Value::Null);
    assert!(text.contains("blocked: legacy_key_only"), "{text}");
    assert!(
        text.contains(".moai/.env.glm") && text.contains("key-set glm"),
        "{text}"
    );
    assert!(text.contains("moai-cli key-import-plan glm"), "{text}");
    // Untouched: same bytes, same time, and no vault was created.
    assert_eq!(fs::read(&file).unwrap(), before);
    assert_eq!(fs::metadata(&file).unwrap().modified().unwrap(), mtime);
    assert!(!h.path().join(".moai-cli").exists());
    // With the key in the variable the legacy file does not matter.
    let (code, v, text) = dry(h.path(), Some(ENV_KEY));
    assert_eq!(code, 0, "{text}");
    assert_eq!(v["key_source"], "env:ANTHROPIC_AUTH_TOKEN");
}

/// A broken project entry stops the launch even though a shared key (and a legacy file)
/// would answer: other users' bits, a plain-text file, a symlink, a lost machine key.
#[test]
fn an_unusable_project_entry_stops_the_launch_instead_of_falling_back() {
    type Break = fn(&Path, &Path);
    let cases: [(&str, Break); 4] = [
        ("open bits", |entry, _| {
            fs::set_permissions(entry, fs::Permissions::from_mode(0o644)).unwrap()
        }),
        ("plain text", |entry, _| {
            fs::write(entry, format!("{PROJECT_KEY}\n")).unwrap()
        }),
        ("symlink", |entry, home| {
            fs::remove_file(entry).unwrap();
            std::os::unix::fs::symlink(home.join(".moai/.env.glm"), entry).unwrap();
        }),
        ("lost machine key", |_, home| {
            fs::remove_file(home.join(".moai-cli/vault/.master.key")).unwrap()
        }),
    ];
    for (what, break_it) in cases {
        let h = home(true);
        store(h.path(), true, SHARED_KEY);
        let entry = store(h.path(), false, PROJECT_KEY);
        break_it(&entry, h.path());
        let (code, v, text) = dry(h.path(), None);
        assert_eq!(code, 1, "{what}: {text}");
        assert_eq!(v, Value::Null, "{what}");
        assert!(text.contains("blocked: key_file_unsafe"), "{what}: {text}");
        assert!(text.contains("next:"), "{what}: {text}");
    }
}

#[test]
fn no_key_anywhere_names_the_vault_command() {
    let h = home(false);
    let (code, _, text) = dry(h.path(), None);
    assert_ne!(code, 0);
    assert!(text.contains("blocked: key_absent"), "{text}");
    assert!(text.contains("moai-cli key-set glm"), "{text}");
}

#[test]
fn other_hosts_report_no_key_source() {
    let h = home(false);
    store(h.path(), true, SHARED_KEY);
    let out = bin(h.path())
        .args(["cc", "-a", "--dry-run"])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    // The field is absent (not null): the recorded protocol shape of other hosts is unchanged.
    assert!(v.get("key_source").is_none(), "{v}");
}

/// Missing git cannot establish that a project key is absent. Even a shared-only vault
/// must not be selected when the current project's identity is unknown.
#[test]
fn missing_git_stops_the_launch_even_with_a_shared_key() {
    let empty_path = tempfile::tempdir().unwrap();
    for with_project in [false, true] {
        let h = home(false);
        store(h.path(), true, SHARED_KEY);
        if with_project {
            store(h.path(), false, PROJECT_KEY);
        }
        let mut cmd = bin(h.path());
        cmd.env("PATH", empty_path.path());
        let (code, v, text) = dry_command(cmd);
        assert_eq!(code, 1, "project={with_project}: {text}");
        assert_eq!(v, Value::Null, "{text}");
        assert!(text.contains("blocked: key_absent"), "{text}");
        assert!(text.contains("next:"), "{text}");
    }
}

/// An identity error must propagate instead of silently selecting a shared entry.
#[test]
fn an_identity_error_stops_the_launch_even_with_a_shared_key() {
    let h = home(false);
    store(h.path(), true, SHARED_KEY);
    let state = h.path().join(".moai-cli");
    let moved = h.path().join("state-fixture");
    fs::rename(&state, &moved).unwrap();
    std::os::unix::fs::symlink(&moved, &state).unwrap();
    let (code, v, text) = dry(h.path(), None);
    assert_eq!(code, 1, "{text}");
    assert_eq!(v, Value::Null, "{text}");
    assert!(text.contains("blocked: key_absent"), "{text}");
    assert!(text.contains("next:"), "{text}");
}

/// With no project entry, an unsafe shared entry must block before the legacy path.
#[test]
fn an_unsafe_shared_entry_stops_the_launch() {
    let h = home(true);
    let entry = store(h.path(), true, SHARED_KEY);
    fs::set_permissions(&entry, fs::Permissions::from_mode(0o644)).unwrap();
    let (code, v, text) = dry(h.path(), None);
    assert_eq!(code, 1, "{text}");
    assert_eq!(v, Value::Null, "{text}");
    assert!(text.contains("blocked: key_file_unsafe"), "{text}");
    assert!(text.contains("next:"), "{text}");
}

/// A successfully identified non-repository still uses the shared entry.
#[test]
fn outside_a_repository_the_shared_entry_is_still_available() {
    let h = home(false);
    store(h.path(), true, SHARED_KEY);
    let outside = tempfile::tempdir().unwrap();
    let mut cmd = bin(h.path());
    cmd.current_dir(outside.path());
    let (code, v, text) = dry_command(cmd);
    assert_eq!(code, 0, "{text}");
    assert_eq!(v["key_source"], "vault:shared:glm", "{text}");
}
