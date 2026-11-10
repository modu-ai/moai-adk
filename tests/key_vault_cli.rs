//! #2393 S1: the file backend of the key vault through the real binary. Every case uses a
//! temporary HOME and a synthetic key made at run time; the real home, the OS key store and
//! any legacy key file are never read. Every answer (stdout and stderr) is checked to hold
//! no key text.
#![cfg(unix)]

use serde_json::Value;
use std::fs;
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::{Command, Output, Stdio};

const BIN: &str = env!("CARGO_BIN_EXE_moai-cli");

/// A synthetic key that cannot be a real one.
fn synthetic() -> String {
    format!("moai-test-key-{}-{}", std::process::id(), line!())
}

fn home() -> (tempfile::TempDir, PathBuf) {
    let dir = tempfile::TempDir::new().unwrap();
    let home = fs::canonicalize(dir.path()).unwrap();
    (dir, home)
}

fn run(home: &Path, args: &[&str], stdin: Option<&str>) -> Output {
    let mut child = Command::new(BIN)
        .args(args)
        .env("HOME", home)
        .env_remove("ANTHROPIC_AUTH_TOKEN")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .unwrap();
    {
        let mut pipe = child.stdin.take().unwrap();
        if let Some(text) = stdin {
            pipe.write_all(text.as_bytes()).unwrap();
        }
    }
    child.wait_with_output().unwrap()
}

/// The envelope, after checking that neither stream holds `secret`.
fn envelope(out: &Output, secret: &str) -> Value {
    let stdout = String::from_utf8_lossy(&out.stdout);
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(!stdout.contains(secret), "key text in stdout");
    assert!(!stderr.contains(secret), "key text in stderr");
    serde_json::from_str(&stdout).unwrap_or_else(|e| panic!("{e}: {stdout} {stderr}"))
}

fn entry(home: &Path) -> PathBuf {
    home.join(".moai-cli/vault/glm.key")
}

fn mode(path: &Path) -> u32 {
    fs::metadata(path).unwrap().permissions().mode() & 0o777
}

#[test]
fn set_then_status_then_remove_round_trip_without_echoing_the_key() {
    let (_d, home) = home();
    let key = synthetic();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["ok"], true, "{s}");
    assert_eq!(s["result"]["backend"], "file");
    assert_eq!(s["result"]["entry"], "absent");

    let set = envelope(
        &run(&home, &["key-set", "glm"], Some(&format!("  {key}\n"))),
        &key,
    );
    assert_eq!(set["result"]["written"], true, "{set}");
    assert_eq!(set["result"]["value_shown"], false);
    assert_eq!(
        fs::read_to_string(entry(&home)).unwrap(),
        format!("{key}\n")
    );
    assert_eq!(mode(&entry(&home)), 0o600);
    assert_eq!(mode(&home.join(".moai-cli/vault")), 0o700);
    // No temporary file is left next to the entry.
    let names: Vec<String> = fs::read_dir(home.join(".moai-cli/vault"))
        .unwrap()
        .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
        .collect();
    assert_eq!(names, ["glm.key"]);

    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "present", "{s}");
    assert_eq!(s["result"]["blocked"], Value::Null);
    // Status says nothing about the value's size either.
    let text = s.to_string();
    assert!(!text.contains(&key.len().to_string()) || !text.contains("len"));

    // Replacing keeps one entry and mode 0600.
    let key2 = format!("{key}-2");
    envelope(&run(&home, &["key-set", "glm"], Some(&key2)), &key2);
    assert_eq!(
        fs::read_to_string(entry(&home)).unwrap(),
        format!("{key2}\n")
    );
    assert_eq!(mode(&entry(&home)), 0o600);

    // Without --yes nothing is removed.
    let r = envelope(&run(&home, &["key-remove", "glm"], None), &key2);
    assert_eq!(r["result"]["removed"], false);
    assert!(entry(&home).exists());
    let r = envelope(&run(&home, &["key-remove", "glm", "--yes"], None), &key2);
    assert_eq!(r["result"]["removed"], true);
    assert!(!entry(&home).exists());
}

#[test]
fn unusable_values_are_refused_and_nothing_is_written() {
    let (_d, home) = home();
    for bad in ["", "   \n", "a\u{1}b"] {
        let out = run(&home, &["key-set", "glm"], Some(bad));
        let env = envelope(&out, "never-written");
        assert_eq!(env["ok"], false, "{bad:?}: {env}");
        assert!(!entry(&home).exists(), "{bad:?}");
    }
}

#[test]
fn an_entry_with_group_or_other_bits_is_unsafe_with_the_chmod_fix() {
    let (_d, home) = home();
    let key = synthetic();
    envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
    // Each group/other bit alone makes the entry unsafe; 0600 and 0400 pass.
    for bits in [0o640, 0o620, 0o610, 0o604, 0o602, 0o601] {
        fs::set_permissions(entry(&home), fs::Permissions::from_mode(bits)).unwrap();
        let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
        assert_eq!(s["result"]["entry"], "unsafe", "{bits:o}: {s}");
        assert_eq!(s["result"]["blocked"]["code"], "key_file_unsafe");
        assert!(s["result"]["blocked"]["next"]
            .as_str()
            .unwrap()
            .starts_with("chmod 600 "));
    }
    for bits in [0o600, 0o400] {
        fs::set_permissions(entry(&home), fs::Permissions::from_mode(bits)).unwrap();
        let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
        assert_eq!(s["result"]["entry"], "present", "{bits:o}: {s}");
    }
    // An open vault folder is unsafe too.
    fs::set_permissions(entry(&home), fs::Permissions::from_mode(0o600)).unwrap();
    fs::set_permissions(
        home.join(".moai-cli/vault"),
        fs::Permissions::from_mode(0o750),
    )
    .unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
    assert!(s["result"]["blocked"]["next"]
        .as_str()
        .unwrap()
        .starts_with("chmod 700 "));
    // And key-set refuses to write into it.
    let out = envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
    assert_eq!(out["ok"], false, "{out}");
}

#[test]
fn a_symlinked_or_special_entry_is_not_followed() {
    let (_d, home) = home();
    let key = synthetic();
    envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
    // A symlink to a file elsewhere holding a key.
    let outside = home.join("outside.key");
    fs::write(&outside, format!("{key}\n")).unwrap();
    fs::set_permissions(&outside, fs::Permissions::from_mode(0o600)).unwrap();
    fs::remove_file(entry(&home)).unwrap();
    std::os::unix::fs::symlink(&outside, entry(&home)).unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
    // A folder in place of the entry.
    fs::remove_file(entry(&home)).unwrap();
    fs::create_dir(entry(&home)).unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
    let out = envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
    assert_eq!(out["ok"], false, "a folder is not replaced: {out}");
    // A FIFO in place of the entry: opened without blocking, refused as not a regular file
    // (reading it would give an empty value, not an unsafe one).
    fs::remove_dir(entry(&home)).unwrap();
    let made = Command::new("mkfifo").arg(entry(&home)).status().unwrap();
    assert!(made.success());
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
    assert!(s["result"]["blocked"]["reason"]
        .as_str()
        .unwrap()
        .ends_with("is not a regular file"));
    // A symlinked vault folder is refused as not a real folder (not only by its mode bits).
    fs::remove_file(entry(&home)).unwrap();
    fs::remove_dir(home.join(".moai-cli/vault")).unwrap();
    fs::create_dir(home.join("elsewhere")).unwrap();
    std::os::unix::fs::symlink(home.join("elsewhere"), home.join(".moai-cli/vault")).unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
    assert!(s["result"]["blocked"]["reason"]
        .as_str()
        .unwrap()
        .ends_with("is not a real folder"));
}

#[test]
fn an_entry_is_read_up_to_64_kib() {
    let (_d, home) = home();
    let key = synthetic();
    envelope(&run(&home, &["key-set", "glm"], Some(&key)), &key);
    // 65_536 bytes: read; 65_537: refused (literal bounds, not the constant).
    let mut body = key.clone().into_bytes();
    body.resize(65_536, b'k');
    fs::write(entry(&home), &body).unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "present", "{s}");
    body.push(b'k');
    fs::write(entry(&home), &body).unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "unsafe", "{s}");
}

#[test]
fn the_config_selects_the_backend_and_os_stores_are_not_available_yet() {
    let (_d, home) = home();
    fs::create_dir_all(home.join(".moai-cli")).unwrap();
    let config = home.join(".moai-cli/config");
    for (text, backend) in [
        ("", "file"),
        ("other = 1\n", "file"),
        ("key_store = \"file\"\n", "file"),
    ] {
        fs::write(&config, text).unwrap();
        let s = envelope(&run(&home, &["key-status", "glm"], None), "x-never");
        assert_eq!(s["result"]["backend"], backend, "{text:?}: {s}");
    }
    // No OS store is linked in this build (#2393 S4): `keyring` blocks with no file fallback;
    // `auto` falls back to the file because the store is unavailable (design §3.1).
    fs::write(&config, "key_store = \"keyring\"\n").unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), "x-never");
    assert_eq!(s["result"]["backend"], "keyring");
    assert_eq!(
        s["result"]["blocked"]["code"], "key_store_unavailable",
        "{s}"
    );
    let set = envelope(&run(&home, &["key-set", "glm"], Some("v-never")), "v-never");
    assert_eq!(set["ok"], false, "no silent file fallback: {set}");
    assert!(!entry(&home).exists());
    fs::write(&config, "key_store = \"auto\"\n").unwrap();
    let key = "moai-test-auto-key";
    let set = envelope(&run(&home, &["key-set", "glm"], Some(key)), key);
    assert_eq!(set["ok"], true, "{set}");
    assert_eq!(set["result"]["backend"], "auto");
    assert!(entry(&home).exists());
    let s = envelope(&run(&home, &["key-status", "glm"], None), key);
    assert_eq!(s["result"]["entry"], "present", "{s}");
    fs::remove_file(entry(&home)).unwrap();
    for bad in [
        "key_store = \"files\"\n",
        "key_store = 1\n",
        "key_store = [\n",
    ] {
        fs::write(&config, bad).unwrap();
        let s = envelope(&run(&home, &["key-status", "glm"], None), "x-never");
        assert_eq!(s["ok"], false, "{bad:?}: {s}");
    }
}

#[test]
fn provider_names_cannot_leave_the_vault() {
    let (_d, home) = home();
    for bad in ["../x", "a/b", "GLM", ".key"] {
        let out = run(&home, &["key-set", bad], Some("v-never"));
        let env = envelope(&out, "v-never");
        assert_eq!(env["ok"], false, "{bad}");
    }
    assert!(!home.join(".moai-cli").exists());
}

#[test]
fn a_legacy_key_file_is_never_read_or_touched() {
    let (_d, home) = home();
    let key = synthetic();
    for legacy in [".moai/.env.glm", ".moai-cli/.env.glm"] {
        let p = home.join(legacy);
        fs::create_dir_all(p.parent().unwrap()).unwrap();
        fs::write(&p, format!("ANTHROPIC_AUTH_TOKEN={key}\n")).unwrap();
    }
    let before = fs::metadata(home.join(".moai/.env.glm"))
        .unwrap()
        .modified()
        .unwrap();
    let s = envelope(&run(&home, &["key-status", "glm"], None), &key);
    assert_eq!(s["result"]["entry"], "absent", "{s}");
    envelope(&run(&home, &["key-remove", "glm", "--yes"], None), &key);
    assert!(home.join(".moai/.env.glm").exists());
    assert!(home.join(".moai-cli/.env.glm").exists());
    assert_eq!(
        fs::metadata(home.join(".moai/.env.glm"))
            .unwrap()
            .modified()
            .unwrap(),
        before
    );
}
