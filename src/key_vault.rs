//! API key vault, stage 1: the file backend and `key status | set | remove` (Issue #2393;
//! design `docs/key-vault-design-2026-11-10.md` §3; user decision #2389 c6774899518 — OS
//! store and file both supported, file by default).
//!
//! An entry is `~/.moai-cli/vault/<provider>.key`: one line, the value only. The backend comes
//! from `key_store` in `~/.moai-cli/config` (TOML): `file` (also when the file or the key is
//! absent), `keyring` or `auto`. The OS-store backends arrive in a later stage with their
//! dependency review; until then they answer `key_store_unavailable`.
//!
//! Safety rules of the file backend (design §3.2), checked on the opened handle so a path
//! swapped between check and read is not followed:
//! - the entry is opened without following a symlink and must be a regular file;
//! - on Unix its owner must be the owner of `HOME`, and it may carry no group or other
//!   permission bits; the vault folder may carry none either (the answer names the `chmod`
//!   fix, never the value);
//! - at most [`VALUE_FILE_MAX`] bytes are read (a bigger file is refused, not truncated);
//! - a write goes to a new `0600` file in the vault folder (created `0700`), then is renamed
//!   over the entry, so a reader sees the old value or the new one.
//!
//! No value, length or fingerprint ever appears in an answer, error or log. A value comes in
//! on stdin only (never argv), and an interactive terminal is refused: nothing here can turn
//! echo off without a terminal library, so the value would be shown while typed.

use serde_json::{json, Value};
use std::fs;
use std::io::{Read, Write};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// Folder of the file backend under the home state directory.
pub const VAULT_DIR: &str = "vault";
/// The config file under the home state directory that may name the backend.
pub const CONFIG_FILE: &str = "config";
/// Largest entry file that is read (the launcher's key file bound, #2359).
pub const VALUE_FILE_MAX: u64 = 65_536;

/// A refusal with a fixed, machine-readable code (design §3.3) and the next action.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Blocked {
    pub code: &'static str,
    pub reason: String,
    pub next: String,
}

impl Blocked {
    fn new(code: &'static str, reason: impl Into<String>, next: impl Into<String>) -> Self {
        Self {
            code,
            reason: reason.into(),
            next: next.into(),
        }
    }

    pub fn to_json(&self) -> Value {
        json!({"code": self.code, "reason": self.reason, "next": self.next})
    }
}

/// The selected backend.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum Backend {
    File,
    Keyring,
    Auto,
}

impl Backend {
    pub fn name(self) -> &'static str {
        match self {
            Backend::File => "file",
            Backend::Keyring => "keyring",
            Backend::Auto => "auto",
        }
    }
}

/// A provider name: 1–32 of `a-z`, `0-9`, `-`, starting with a letter. It becomes a file
/// name, so nothing else is accepted.
pub fn check_provider(provider: &str) -> Result<()> {
    let ok = (1..=32).contains(&provider.len())
        && provider.as_bytes()[0].is_ascii_lowercase()
        && provider
            .bytes()
            .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'-');
    if ok {
        Ok(())
    } else {
        Err("provider must be 1-32 characters of a-z, 0-9 and '-', starting with a letter".into())
    }
}

/// The home directory: `HOME`, or `USERPROFILE` on Windows when `HOME` is unset.
pub fn home() -> Result<PathBuf> {
    std::env::var_os("HOME")
        .filter(|h| !h.is_empty())
        .or_else(|| {
            cfg!(windows)
                .then(|| std::env::var_os("USERPROFILE"))
                .flatten()
        })
        .map(PathBuf::from)
        .ok_or_else(|| "HOME is not set".to_string())
}

fn state_dir(home: &Path) -> PathBuf {
    home.join(crate::HOME_STATE_DIR)
}

fn vault_dir(home: &Path) -> PathBuf {
    state_dir(home).join(VAULT_DIR)
}

fn entry_path(home: &Path, provider: &str) -> PathBuf {
    vault_dir(home).join(format!("{provider}.key"))
}

/// The backend named by `key_store` in the config; `file` when the config or the key is
/// absent. An unreadable config, a non-string value or an unknown name is an error rather
/// than a silent default, so a typo cannot select a weaker store than intended.
pub fn backend(home: &Path) -> Result<Backend> {
    let path = state_dir(home).join(CONFIG_FILE);
    let text = match fs::symlink_metadata(&path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Backend::File),
        Err(_) => return Err("~/.moai-cli/config is unreadable".into()),
        Ok(m) if !m.is_file() || m.len() > VALUE_FILE_MAX => {
            return Err("~/.moai-cli/config is not a regular file within 64 KiB".into())
        }
        Ok(_) => fs::read_to_string(&path).map_err(|_| "~/.moai-cli/config is unreadable")?,
    };
    let doc: toml_edit::DocumentMut = text
        .parse()
        .map_err(|_| "~/.moai-cli/config is not valid TOML")?;
    match doc.get("key_store") {
        None => Ok(Backend::File),
        Some(item) => match item.as_str() {
            Some("file") => Ok(Backend::File),
            Some("keyring") => Ok(Backend::Keyring),
            Some("auto") => Ok(Backend::Auto),
            _ => Err(
                "key_store in ~/.moai-cli/config must be \"file\", \"keyring\" or \"auto\"".into(),
            ),
        },
    }
}

fn os_store_unavailable(backend: Backend) -> Blocked {
    Blocked::new(
        "key_store_unavailable",
        format!(
            "key_store = \"{}\" needs the OS store backend, which is not available in this build",
            backend.name()
        ),
        "set key_store = \"file\" in ~/.moai-cli/config, or wait for the OS store backend",
    )
}

/// The Unix permission and owner rule for one opened object. `what` names it in the answer.
#[cfg(unix)]
fn unix_unsafe(meta: &fs::Metadata, home: &Path, path: &Path, dir: bool) -> Option<Blocked> {
    use std::os::unix::fs::MetadataExt;
    let owner = fs::metadata(home).ok()?.uid();
    let shown = path.display();
    if meta.uid() != owner {
        return Some(Blocked::new(
            "key_file_unsafe",
            format!("{shown} is not owned by the owner of HOME"),
            format!(
                "chown the {} to your user",
                if dir { "folder" } else { "file" }
            ),
        ));
    }
    if meta.mode() & 0o077 != 0 {
        let fix = if dir { "700" } else { "600" };
        return Some(Blocked::new(
            "key_file_unsafe",
            format!("{shown} is readable or writable by other users"),
            format!("chmod {fix} {shown}"),
        ));
    }
    None
}

#[cfg(not(unix))]
fn unix_unsafe(_: &fs::Metadata, _: &Path, _: &Path, _: bool) -> Option<Blocked> {
    // Windows: the profile folder's ACL is inherited (design §3.2); no mode bits to check.
    None
}

/// The vault folder's state: absent is fine (nothing stored yet), anything but a private
/// real directory is unsafe.
fn check_vault_dir(home: &Path) -> std::result::Result<bool, Blocked> {
    let dir = vault_dir(home);
    match fs::symlink_metadata(&dir) {
        Err(_) => Ok(false),
        Ok(m) if !m.is_dir() => Err(Blocked::new(
            "key_file_unsafe",
            format!("{} is not a real folder", dir.display()),
            "move it away; the vault folder is created on the next key set",
        )),
        Ok(m) => match unix_unsafe(&m, home, &dir, true) {
            Some(b) => Err(b),
            None => Ok(true),
        },
    }
}

/// Opens an entry for reading without following a symlink (and without blocking on a FIFO).
fn open_entry(path: &Path) -> std::io::Result<fs::File> {
    let mut options = fs::OpenOptions::new();
    options.read(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK);
    }
    options.open(path)
}

/// The one rule for a value (shared with the launcher, #2359 AC4): surrounding whitespace
/// is dropped; empty, whitespace-only or control-character values are no value.
pub fn usable_value(value: &str) -> Option<String> {
    let value = value.trim();
    (!value.is_empty() && !value.chars().any(char::is_control)).then(|| value.to_string())
}

/// The file backend's value for `provider`: `Ok(None)` when no entry exists, `Err` when the
/// store or the entry breaks a safety rule or holds no usable value.
pub fn read_file_entry(
    home: &Path,
    provider: &str,
) -> std::result::Result<Option<String>, Blocked> {
    if !check_vault_dir(home)? {
        return Ok(None);
    }
    let path = entry_path(home, provider);
    let file = match open_entry(&path) {
        Ok(f) => f,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        Err(_) => {
            // A symlink (O_NOFOLLOW) or an unopenable entry.
            return Err(Blocked::new(
                "key_file_unsafe",
                format!("{} is not a readable regular file", path.display()),
                format!("remove it and run: moai-cli key set {provider}"),
            ));
        }
    };
    let meta = file.metadata().map_err(|_| {
        Blocked::new(
            "key_file_unsafe",
            format!("{} metadata is unreadable", path.display()),
            format!("remove it and run: moai-cli key set {provider}"),
        )
    })?;
    if !meta.is_file() {
        return Err(Blocked::new(
            "key_file_unsafe",
            format!("{} is not a regular file", path.display()),
            format!("remove it and run: moai-cli key set {provider}"),
        ));
    }
    if let Some(b) = unix_unsafe(&meta, home, &path, false) {
        return Err(b);
    }
    if meta.len() > VALUE_FILE_MAX {
        return Err(Blocked::new(
            "key_file_unsafe",
            format!("{} is larger than 64 KiB", path.display()),
            format!("run: moai-cli key set {provider}"),
        ));
    }
    let mut text = String::new();
    file.take(VALUE_FILE_MAX + 1)
        .read_to_string(&mut text)
        .map_err(|_| {
            Blocked::new(
                "key_file_unsafe",
                format!("{} is not readable text", path.display()),
                format!("run: moai-cli key set {provider}"),
            )
        })?;
    match usable_value(&text) {
        Some(v) => Ok(Some(v)),
        None => Err(Blocked::new(
            "key_absent",
            format!("{} holds no usable value", path.display()),
            format!("run: moai-cli key set {provider} (value on stdin)"),
        )),
    }
}

/// The selected backend's value for `provider` (what the launcher will read, stage 2).
pub fn read(home: &Path, provider: &str) -> std::result::Result<Option<String>, Blocked> {
    check_provider(provider)
        .map_err(|e| Blocked::new("key_absent", e, "use a valid provider name"))?;
    let backend = backend(home).map_err(|e| {
        Blocked::new(
            "key_store_unavailable",
            e,
            "fix key_store in ~/.moai-cli/config",
        )
    })?;
    match backend {
        Backend::File => read_file_entry(home, provider),
        other => Err(os_store_unavailable(other)),
    }
}

/// `key status`: the backend and whether an entry exists and passes the safety rules. Never
/// the value, its length or a fingerprint.
pub fn status(home: &Path, provider: &str) -> Result<Value> {
    check_provider(provider)?;
    let backend = backend(home)?;
    let path = entry_path(home, provider);
    let (entry, blocked) = match backend {
        Backend::File => match read_file_entry(home, provider) {
            Ok(Some(_)) => ("present", None),
            Ok(None) => ("absent", None),
            Err(b) if b.code == "key_absent" => ("empty", Some(b)),
            Err(b) => ("unsafe", Some(b)),
        },
        other => ("unknown", Some(os_store_unavailable(other))),
    };
    Ok(json!({
        "provider": provider,
        "backend": backend.name(),
        "entry": entry,
        "path": path.to_string_lossy(),
        "blocked": blocked.map(|b| b.to_json()),
        "value_shown": false,
    }))
}

/// Reads the value for `key set` from `input`: at most [`VALUE_FILE_MAX`] bytes, one usable
/// value. The caller refuses an interactive terminal before calling this.
pub fn read_value(mut input: impl Read) -> Result<String> {
    let mut bytes = Vec::new();
    input
        .by_ref()
        .take(VALUE_FILE_MAX + 1)
        .read_to_end(&mut bytes)
        .map_err(|_| "stdin read failed")?;
    if bytes.len() as u64 > VALUE_FILE_MAX {
        return Err("the value exceeds 64 KiB".into());
    }
    let text = String::from_utf8(bytes).map_err(|_| "the value is not UTF-8 text")?;
    usable_value(&text).ok_or_else(|| {
        "no usable value on stdin (empty, whitespace only, or a control character)".to_string()
    })
}

/// Creates the vault folder `0700` when absent (its parent too), and checks an existing one.
fn ensure_vault_dir(home: &Path) -> Result<PathBuf> {
    let dir = vault_dir(home);
    if check_vault_dir(home).map_err(|b| format!("{}: {}", b.code, b.reason))? {
        return Ok(dir);
    }
    let mut builder = fs::DirBuilder::new();
    builder.recursive(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::DirBuilderExt;
        builder.mode(0o700);
    }
    builder
        .create(&dir)
        .map_err(|_| format!("cannot create {}", dir.display()))?;
    check_vault_dir(home).map_err(|b| format!("{}: {}", b.code, b.reason))?;
    Ok(dir)
}

/// `key set` (file backend): writes `value` to a new `0600` file in the vault folder, then
/// renames it over the entry. Answers without the value.
pub fn set(home: &Path, provider: &str, value: &str) -> Result<Value> {
    check_provider(provider)?;
    let backend = backend(home)?;
    if backend != Backend::File {
        let b = os_store_unavailable(backend);
        return Err(format!("{}: {}", b.code, b.reason));
    }
    let value = usable_value(value).ok_or("no usable value")?;
    let dir = ensure_vault_dir(home)?;
    let entry = entry_path(home, provider);
    // Refuse to replace something that is not a regular file (a symlink is not followed by
    // rename, but replacing a folder or a device would hide it).
    if let Ok(m) = fs::symlink_metadata(&entry) {
        if !m.is_file() && !m.file_type().is_symlink() {
            return Err(format!(
                "{} is not a regular file; move it away first",
                entry.display()
            ));
        }
    }
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_nanos())
        .unwrap_or(0);
    let tmp = dir.join(format!(
        ".{provider}.key.tmp-{}-{nanos}",
        std::process::id()
    ));
    let mut options = fs::OpenOptions::new();
    options.write(true).create_new(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600).custom_flags(libc::O_NOFOLLOW);
    }
    let written = (|| -> std::io::Result<()> {
        let mut f = options.open(&tmp)?;
        f.write_all(value.as_bytes())?;
        f.write_all(b"\n")?;
        f.sync_all()?;
        fs::rename(&tmp, &entry)
    })();
    if written.is_err() {
        let _ = fs::remove_file(&tmp);
        return Err(format!("cannot write {}", entry.display()));
    }
    Ok(json!({
        "provider": provider,
        "backend": backend.name(),
        "path": entry.to_string_lossy(),
        "written": true,
        "value_shown": false,
    }))
}

/// `key remove` (file backend): removes the entry file only, after the caller's
/// confirmation (`--yes`). No other file is touched (a legacy key file stays as it is).
pub fn remove(home: &Path, provider: &str, confirmed: bool) -> Result<Value> {
    check_provider(provider)?;
    let backend = backend(home)?;
    if backend != Backend::File {
        let b = os_store_unavailable(backend);
        return Err(format!("{}: {}", b.code, b.reason));
    }
    let entry = entry_path(home, provider);
    let exists = fs::symlink_metadata(&entry).is_ok();
    if !confirmed {
        return Ok(json!({
            "provider": provider,
            "backend": backend.name(),
            "path": entry.to_string_lossy(),
            "exists": exists,
            "removed": false,
            "next": format!("rerun with --yes to remove the {provider} entry"),
        }));
    }
    let removed = if exists {
        fs::remove_file(&entry).map_err(|_| format!("cannot remove {}", entry.display()))?;
        true
    } else {
        false
    };
    Ok(json!({
        "provider": provider,
        "backend": backend.name(),
        "path": entry.to_string_lossy(),
        "exists": exists,
        "removed": removed,
    }))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn provider_names_are_file_safe() {
        for ok in ["glm", "z-ai", "a1", &"a".repeat(32)] {
            assert!(check_provider(ok).is_ok(), "{ok}");
        }
        for bad in [
            "",
            "Glm",
            "1glm",
            "-glm",
            "glm.key",
            "../x",
            "a/b",
            "glm ",
            &"a".repeat(33),
        ] {
            assert!(check_provider(bad).is_err(), "{bad:?}");
        }
    }

    #[test]
    fn a_value_is_trimmed_and_control_characters_refuse_it() {
        assert_eq!(usable_value("  k-1  \n").as_deref(), Some("k-1"));
        for bad in ["", "   ", "\n", "a\u{0}b", "a\tb", "a\u{1b}b"] {
            assert_eq!(usable_value(bad), None, "{bad:?}");
        }
    }

    #[test]
    fn stdin_values_are_bounded_at_64_kib() {
        // 65_536 bytes of value is the largest accepted; one more is refused.
        assert!(read_value(&vec![b'k'; 65_536][..]).is_ok());
        assert!(read_value(&vec![b'k'; 65_537][..]).is_err());
        assert!(read_value(&b"\xff\xfe"[..]).is_err());
        assert!(read_value(&b"   \n"[..]).is_err());
    }

    /// `read` (the launcher's entry point, stage 2) follows the selected backend and never
    /// falls back to the file when an OS store is selected.
    #[test]
    fn read_follows_the_backend_and_does_not_fall_back() {
        let dir = tempfile::TempDir::new().unwrap();
        let home = dir.path();
        let value = format!("moai-test-key-{}", std::process::id());
        set(home, "glm", &value).unwrap();
        assert_eq!(read(home, "glm").unwrap().as_deref(), Some(value.as_str()));
        assert_eq!(read(home, "other").unwrap(), None);
        for name in ["keyring", "auto"] {
            fs::write(
                home.join(".moai-cli/config"),
                format!("key_store = \"{name}\"\n"),
            )
            .unwrap();
            let err = read(home, "glm").unwrap_err();
            assert_eq!(err.code, "key_store_unavailable", "{name}");
            assert!(!err.reason.contains(&value) && !err.next.contains(&value));
        }
        fs::write(home.join(".moai-cli/config"), "key_store = \"vault\"\n").unwrap();
        assert_eq!(read(home, "glm").unwrap_err().code, "key_store_unavailable");
    }
}
