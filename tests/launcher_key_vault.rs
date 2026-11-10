//! #2393 S2: the `glm` launcher reads the key from the variable, then the vault, then (only
//! when the vault holds no entry) the legacy key files, read only. The source is reported by
//! name (`key_source`), never the value. A vault that is present but unsafe, or a backend
//! that cannot answer, stops the launch instead of falling back. Temporary HOME, synthetic
//! keys only; the real home and the OS key store are never read.
#![cfg(unix)]
mod common;
use serde_json::Value;
use std::fs;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::Command;
use std::sync::OnceLock;

const VAULT_KEY: &str = "moai-test-vault-key-not-real";
const LEGACY_KEY: &str = "moai-test-legacy-key-not-real";
const ENV_KEY: &str = "moai-test-env-key-not-real";

fn project() -> &'static Path {
    static PROJECT: OnceLock<PathBuf> = OnceLock::new();
    PROJECT.get_or_init(|| {
        let p = Path::new(env!("CARGO_TARGET_TMPDIR")).join("launcher-key-vault-project");
        fs::create_dir_all(&p).unwrap();
        let init = common::git::command()
            .args(["init", "-q"])
            .current_dir(&p)
            .status()
            .unwrap();
        assert!(init.success());
        p
    })
}

/// A temporary HOME with an optional vault entry (mode `vault_mode`) and legacy file.
fn home(vault: Option<(&str, u32)>, legacy: bool) -> tempfile::TempDir {
    let dir = tempfile::tempdir().unwrap();
    let h = dir.path();
    if let Some((value, mode)) = vault {
        let v = h.join(".moai-cli/vault");
        fs::create_dir_all(&v).unwrap();
        fs::set_permissions(&v, fs::Permissions::from_mode(0o700)).unwrap();
        fs::write(v.join("glm.key"), value).unwrap();
        fs::set_permissions(v.join("glm.key"), fs::Permissions::from_mode(mode)).unwrap();
    }
    if legacy {
        fs::create_dir_all(h.join(".moai")).unwrap();
        fs::write(
            h.join(".moai/.env.glm"),
            format!("GLM_API_KEY={LEGACY_KEY}\n"),
        )
        .unwrap();
    }
    dir
}

fn launcher(home: &Path, env_key: Option<&str>) -> Command {
    let mut cmd = Command::new(env!("CARGO_BIN_EXE_moai-cli"));
    cmd.current_dir(project())
        .env("HOME", home)
        .env_remove("CLAUDE_CODE_DISABLE_ADVISOR_TOOL")
        .env_remove("ANTHROPIC_AUTH_TOKEN");
    if let Some(k) = env_key {
        cmd.env("ANTHROPIC_AUTH_TOKEN", k);
    }
    cmd
}

/// `(exit code, stdout JSON, all output)` of `glm -a --dry-run`; no key text may appear.
fn dry(home: &Path, env_key: Option<&str>) -> (i32, Value, String) {
    let out = launcher(home, env_key)
        .args(["glm", "-a", "--dry-run"])
        .output()
        .unwrap();
    let stdout = String::from_utf8_lossy(&out.stdout).to_string();
    let text = format!("{stdout}{}", String::from_utf8_lossy(&out.stderr));
    for key in [VAULT_KEY, LEGACY_KEY, ENV_KEY] {
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
    let bin = tempfile::tempdir().unwrap();
    let seen = bin.path().join("seen");
    let fake = bin.path().join("claude");
    fs::write(
        &fake,
        format!(
            "#!/bin/sh\nprintf %s \"$ANTHROPIC_AUTH_TOKEN|$Z_AI_API_KEY\" > '{}'\n",
            seen.display()
        ),
    )
    .unwrap();
    fs::set_permissions(&fake, fs::Permissions::from_mode(0o755)).unwrap();
    let out = launcher(home, None)
        .env("PATH", format!("{}:/usr/bin:/bin", bin.path().display()))
        .args(["glm", "-a"])
        .output()
        .unwrap();
    let text = format!(
        "{}{}",
        String::from_utf8_lossy(&out.stdout),
        String::from_utf8_lossy(&out.stderr)
    );
    assert!(out.status.success(), "{text}");
    assert!(!text.contains(VAULT_KEY) && !text.contains(LEGACY_KEY));
    let record = fs::read_to_string(&seen).unwrap();
    let (a, b) = record.split_once('|').unwrap();
    assert_eq!(a, b, "both names carry the same key");
    a.to_string()
}

#[test]
fn the_vault_comes_before_the_legacy_file_and_after_the_variable() {
    let h = home(Some((&format!("{VAULT_KEY}\n"), 0o600)), true);
    let (code, v, text) = dry(h.path(), None);
    assert_eq!(code, 0, "{text}");
    assert_eq!(v["key_source"], "vault:glm", "{text}");
    assert!(v["notes"].to_string().contains("key vault"));
    assert_eq!(host_token(h.path()), VAULT_KEY);

    let (code, v, text) = dry(h.path(), Some(ENV_KEY));
    assert_eq!(code, 0, "{text}");
    assert_eq!(v["key_source"], "env:ANTHROPIC_AUTH_TOKEN");
}

#[test]
fn without_a_vault_entry_the_legacy_file_is_still_read_only() {
    let h = home(None, true);
    let before = fs::read(h.path().join(".moai/.env.glm")).unwrap();
    let mtime = fs::metadata(h.path().join(".moai/.env.glm"))
        .unwrap()
        .modified()
        .unwrap();
    let (code, v, text) = dry(h.path(), None);
    assert_eq!(code, 0, "{text}");
    assert_eq!(v["key_source"], "legacy_key_file");
    assert!(v["notes"].to_string().contains("key-set glm"), "{text}");
    assert_eq!(host_token(h.path()), LEGACY_KEY);
    // Read only: same bytes, same time, and no vault was created.
    assert_eq!(fs::read(h.path().join(".moai/.env.glm")).unwrap(), before);
    assert_eq!(
        fs::metadata(h.path().join(".moai/.env.glm"))
            .unwrap()
            .modified()
            .unwrap(),
        mtime
    );
    assert!(!h.path().join(".moai-cli").exists());
}

#[test]
fn an_unusable_vault_stops_the_launch_instead_of_falling_back() {
    // An entry other users can read, an empty entry, a symlinked entry: each with a legacy
    // file present that would otherwise answer.
    let open = home(Some((&format!("{VAULT_KEY}\n"), 0o644)), true);
    let empty = home(Some(("\n", 0o600)), true);
    let linked = home(Some((&format!("{VAULT_KEY}\n"), 0o600)), true);
    let entry = linked.path().join(".moai-cli/vault/glm.key");
    fs::remove_file(&entry).unwrap();
    std::os::unix::fs::symlink(linked.path().join(".moai/.env.glm"), &entry).unwrap();
    for (h, code_word) in [
        (&open, "key_file_unsafe"),
        (&empty, "key_absent"),
        (&linked, "key_file_unsafe"),
    ] {
        let (code, v, text) = dry(h.path(), None);
        assert_ne!(code, 0, "{code_word}: {text}");
        assert_eq!(v, Value::Null);
        assert!(text.contains(&format!("blocked: {code_word}")), "{text}");
        assert!(text.contains("next:"), "{text}");
    }
}

#[test]
fn an_os_store_backend_that_cannot_answer_blocks_without_a_file_fallback() {
    for name in ["keyring", "auto"] {
        let h = home(Some((&format!("{VAULT_KEY}\n"), 0o600)), true);
        fs::write(
            h.path().join(".moai-cli/config"),
            format!("key_store = \"{name}\"\n"),
        )
        .unwrap();
        let (code, _, text) = dry(h.path(), None);
        assert_ne!(code, 0, "{name}: {text}");
        assert!(text.contains("blocked: key_store_unavailable"), "{text}");
    }
    // The variable still wins, so a user is never locked out by the config.
    let h = home(None, false);
    fs::create_dir_all(h.path().join(".moai-cli")).unwrap();
    fs::write(
        h.path().join(".moai-cli/config"),
        "key_store = \"keyring\"\n",
    )
    .unwrap();
    let (code, v, text) = dry(h.path(), Some(ENV_KEY));
    assert_eq!(code, 0, "{text}");
    assert_eq!(v["key_source"], "env:ANTHROPIC_AUTH_TOKEN");
}

#[test]
fn no_key_anywhere_names_the_vault_command() {
    let h = home(None, false);
    let (code, _, text) = dry(h.path(), None);
    assert_ne!(code, 0);
    assert!(text.contains("blocked: key_absent"), "{text}");
    assert!(text.contains("moai-cli key-set glm"), "{text}");
}

#[test]
fn other_hosts_report_no_key_source() {
    let h = home(Some((&format!("{VAULT_KEY}\n"), 0o600)), false);
    let out = launcher(h.path(), None)
        .args(["cc", "-a", "--dry-run"])
        .output()
        .unwrap();
    let v: Value = serde_json::from_slice(&out.stdout).unwrap();
    // The field is absent (not null): the recorded protocol shape of other hosts is unchanged.
    assert!(v.get("key_source").is_none(), "{v}");
}
