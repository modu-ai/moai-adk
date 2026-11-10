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
/// Largest value `key-set` stores: the entry is the value plus a newline, so a stored value
/// is always within [`VALUE_FILE_MAX`] when read back (pre-audit B1, Issue #2393).
pub const VALUE_MAX: usize = 65_535;

/// A refusal with a fixed, machine-readable code (design §3.3) and the next action.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Blocked {
    pub code: &'static str,
    pub reason: String,
    pub next: String,
}

impl Blocked {
    /// No vault entry, but a legacy key file exists (`file` relative to HOME). Its content is
    /// not read (user decision, Issue #2393 comment 6778150969).
    pub fn legacy_key_only(file: &str) -> Self {
        Blocked::new(
            "legacy_key_only",
            format!(
                "the vault holds no glm key; ~/{file} exists but is no longer read"
            ),
            "preview the move with `moai-cli key-import-plan glm` (then `key-import-apply`), or store a key with `moai-cli key-set glm` (value on stdin); the old file is left as it is",
        )
    }

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
    let shown = path.display();
    let Ok(home_meta) = fs::metadata(home) else {
        // Without HOME's owner nothing can be checked: refuse rather than call it safe.
        return Some(Blocked::new(
            "key_file_unsafe",
            format!("cannot read HOME to check the owner of {shown}"),
            "check that HOME points at your home folder",
        ));
    };
    match owner_mode_problem(meta.uid(), meta.mode(), home_meta.uid()) {
        None => None,
        Some(OwnerModeProblem::Owner) => Some(Blocked::new(
            "key_file_unsafe",
            format!("{shown} is not owned by the owner of HOME"),
            format!(
                "chown the {} to your user",
                if dir { "folder" } else { "file" }
            ),
        )),
        Some(OwnerModeProblem::Mode) => {
            let fix = if dir { "700" } else { "600" };
            Some(Blocked::new(
                "key_file_unsafe",
                format!("{shown} is readable or writable by other users"),
                format!("chmod {fix} {shown}"),
            ))
        }
    }
}

/// What is wrong with an entry's owner and mode, if anything (pure, so the owner rule can be
/// tested without a second user — pre-audit N1).
#[cfg_attr(not(unix), allow(dead_code))]
#[derive(Debug, PartialEq, Eq)]
enum OwnerModeProblem {
    Owner,
    Mode,
}

#[cfg_attr(not(unix), allow(dead_code))]
fn owner_mode_problem(uid: u32, mode: u32, home_uid: u32) -> Option<OwnerModeProblem> {
    if uid != home_uid {
        return Some(OwnerModeProblem::Owner);
    }
    (mode & 0o077 != 0).then_some(OwnerModeProblem::Mode)
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
            "move it away; the vault folder is created on the next key-set",
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
                format!("remove it and run: moai-cli key-set {provider}"),
            ));
        }
    };
    let meta = file.metadata().map_err(|_| {
        Blocked::new(
            "key_file_unsafe",
            format!("{} metadata is unreadable", path.display()),
            format!("remove it and run: moai-cli key-set {provider}"),
        )
    })?;
    if !meta.is_file() {
        return Err(Blocked::new(
            "key_file_unsafe",
            format!("{} is not a regular file", path.display()),
            format!("remove it and run: moai-cli key-set {provider}"),
        ));
    }
    if let Some(b) = unix_unsafe(&meta, home, &path, false) {
        return Err(b);
    }
    if meta.len() > VALUE_FILE_MAX {
        return Err(Blocked::new(
            "key_file_unsafe",
            format!("{} is larger than 64 KiB", path.display()),
            format!("run: moai-cli key-set {provider}"),
        ));
    }
    let mut text = String::new();
    file.take(VALUE_FILE_MAX + 1)
        .read_to_string(&mut text)
        .map_err(|_| {
            Blocked::new(
                "key_file_unsafe",
                format!("{} is not readable text", path.display()),
                format!("run: moai-cli key-set {provider}"),
            )
        })?;
    match usable_value(&text) {
        Some(v) => Ok(Some(v)),
        None => Err(Blocked::new(
            "key_absent",
            format!("{} holds no usable value", path.display()),
            format!("run: moai-cli key-set {provider} (value on stdin)"),
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
    let value = usable_value(&text).ok_or_else(|| {
        "no usable value on stdin (empty, whitespace only, or a control character)".to_string()
    })?;
    if value.len() > VALUE_MAX {
        return Err("the value exceeds 65,535 bytes".into());
    }
    Ok(value)
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
    if value.len() > VALUE_MAX {
        return Err("the value exceeds 65,535 bytes".into());
    }
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

// ---- stage 3: moving a legacy key file into the vault (design §4) ----

/// The legacy GLM key files under HOME, in the launcher's order (#2359). Read only; the
/// import never changes, renames or removes them (Q7: after the Go tool is gone, #2383).
pub const LEGACY_GLM_FILES: [&str; 2] = [".moai/.env.glm", ".moai-cli/.env.glm"];
/// Names a legacy key file may use for the token.
pub const LEGACY_KEY_NAMES: [&str; 2] = ["ANTHROPIC_AUTH_TOKEN", "GLM_API_KEY"];

/// The token in a legacy key file's text: the first `NAME=value` line for an accepted name,
/// with an `export ` prefix and surrounding quotes removed. Shared by the launcher (#2359) and
/// the import, so both read a file the same way.
pub fn legacy_token_in(text: &str) -> Option<String> {
    text.lines().find_map(|line| {
        let line = line.trim();
        let line = line.strip_prefix("export ").unwrap_or(line);
        let (key, value) = line.split_once('=')?;
        if !LEGACY_KEY_NAMES.contains(&key.trim()) {
            return None;
        }
        usable_value(value.trim().trim_matches(|c| c == '"' || c == '\''))
    })
}

/// One legacy source as the import sees it: its facts and, when readable, its token. The
/// token never leaves this module's import functions.
struct LegacySource {
    path: PathBuf,
    facts: Value,
    /// Metadata that changes when the file is replaced or rewritten; part of the digest.
    identity: Value,
    token: Option<String>,
}

fn legacy_source(home: &Path, rel: &str) -> LegacySource {
    let path = home.join(rel);
    let mut facts = json!({"path": path.to_string_lossy(), "exists": false});
    let mut identity = Value::Null;
    let mut token = None;
    if fs::symlink_metadata(&path).is_ok() {
        facts["exists"] = json!(true);
        if let Ok(file) = open_entry(&path) {
            if let Ok(meta) = file.metadata() {
                let regular = meta.is_file();
                let size_ok = meta.len() <= VALUE_FILE_MAX;
                facts["regular_file"] = json!(regular);
                facts["size_ok"] = json!(size_ok);
                #[cfg(unix)]
                {
                    use std::os::unix::fs::MetadataExt;
                    facts["private"] = json!(meta.mode() & 0o077 == 0);
                    identity = json!([
                        meta.dev(),
                        meta.ino(),
                        meta.len(),
                        meta.mtime(),
                        meta.mtime_nsec()
                    ]);
                }
                #[cfg(not(unix))]
                {
                    identity = json!([meta.len(), format!("{:?}", meta.modified().ok())]);
                }
                if regular && size_ok {
                    let mut text = String::new();
                    if file.take(VALUE_FILE_MAX).read_to_string(&mut text).is_ok() {
                        token = legacy_token_in(&text);
                    }
                }
            }
        } else {
            // A symlink (not followed) or an unopenable path.
            facts["regular_file"] = json!(false);
        }
        facts["has_key"] = json!(token.is_some());
    }
    LegacySource {
        path,
        facts,
        identity,
        token,
    }
}

/// The import plan and the source it would read (internal: carries the token).
fn import_plan_inner(home: &Path, provider: &str) -> Result<(Value, Option<LegacySource>)> {
    check_provider(provider)?;
    if provider != "glm" {
        return Err("only the glm key has legacy files to import".into());
    }
    let backend = backend(home)?;
    let sources: Vec<LegacySource> = LEGACY_GLM_FILES
        .iter()
        .map(|rel| legacy_source(home, rel))
        .collect();
    let dest = entry_path(home, provider);
    let entry = match backend {
        Backend::File => match read_file_entry(home, provider) {
            Ok(Some(_)) => "present",
            Ok(None) => "absent",
            Err(b) if b.code == "key_absent" => "empty",
            Err(_) => "unsafe",
        },
        _ => "unknown",
    };
    let chosen = sources.iter().position(|s| s.token.is_some());
    let (action, reason) = match (backend, entry, chosen) {
        (Backend::Keyring | Backend::Auto, _, _) => (
            "blocked",
            "key_store_unavailable: the OS store backend is not available in this build",
        ),
        (_, "present", _) => (
            "already_in_vault",
            "the vault already holds a glm key; nothing is imported",
        ),
        (_, "unsafe" | "empty", _) => (
            "blocked",
            "key_file_unsafe: the vault entry exists but is unusable; fix or remove it first",
        ),
        (_, _, None) => ("nothing_to_import", "no legacy file holds a usable key"),
        _ => (
            "import",
            "copy the key from the chosen legacy file into the vault",
        ),
    };
    let chosen_path = chosen.map(|i| sources[i].path.to_string_lossy().into_owned());
    let digest_input = json!({
        "schema_version": 1,
        "provider": provider,
        "backend": backend.name(),
        "action": action,
        "source": chosen_path,
        "source_identity": chosen.map(|i| sources[i].identity.clone()),
        "destination": dest.to_string_lossy(),
        "destination_entry": entry,
    });
    let digest = crate::digest(digest_input.to_string().as_bytes());
    let plan = json!({
        "provider": provider,
        "backend": backend.name(),
        "action": action,
        "reason": reason,
        "sources": sources.iter().map(|s| s.facts.clone()).collect::<Vec<_>>(),
        "source": chosen_path,
        "destination": {"path": dest.to_string_lossy(), "entry": entry},
        "original_kept": true,
        "plan_digest": digest,
        "value_shown": false,
        "writes": false,
    });
    let source = chosen.map(|i| sources.into_iter().nth(i).unwrap());
    Ok((plan, source))
}

/// `key-import-plan`: what would be copied from where to where. Reads the legacy files to
/// know whether they hold a key, writes nothing, and never shows the value.
pub fn import_plan(home: &Path, provider: &str) -> Result<Value> {
    Ok(import_plan_inner(home, provider)?.0)
}

/// `key-import-apply`: re-plans, refuses unless the digest matches (the files did not change
/// since the preview) and a permission reference is given, copies the key into the vault,
/// reads it back and compares, then records a receipt with paths, time, backend and the
/// comparison result only. The legacy file is left exactly as it was.
pub fn import_apply(
    home: &Path,
    provider: &str,
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    if ["", "null", "placeholder"].contains(&permission.trim())
        || permission.len() > 128
        || permission.chars().any(char::is_control)
    {
        return Err("permission reference required".into());
    }
    let (plan, source) = import_plan_inner(home, provider)?;
    if plan["plan_digest"] != json!(plan_digest) {
        return Err(
            "plan digest mismatch: the files changed since the preview; run key-import-plan again"
                .into(),
        );
    }
    if plan["action"] != json!("import") {
        return Err(format!(
            "nothing to apply: {}",
            plan["reason"].as_str().unwrap_or("")
        ));
    }
    let source = source.ok_or("the chosen source vanished")?;
    let token = source
        .token
        .as_deref()
        .ok_or("the chosen source holds no key")?;
    set(home, provider, token)?;
    let matched = matches!(read_file_entry(home, provider), Ok(Some(v)) if v == token);
    let unix_ms = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_millis() as u64)
        .unwrap_or(0);
    let receipt = json!({
        "schema_version": 1,
        "operation": "key.import_apply",
        "provider": provider,
        "backend": plan["backend"],
        "source": source.path.to_string_lossy(),
        "destination": plan["destination"]["path"],
        "matched": matched,
        "original_kept": true,
        "permission_reference": permission,
        "unix_ms": unix_ms,
    });
    let receipt_path = write_home_receipt(
        home,
        &format!("key-import-{unix_ms}-{}", std::process::id()),
        &receipt,
    )?;
    if !matched {
        return Err(format!(
            "the vault read-back did not match; receipt {}",
            receipt_path.display()
        ));
    }
    Ok(json!({
        "provider": provider,
        "backend": plan["backend"],
        "source": source.path.to_string_lossy(),
        "destination": plan["destination"]["path"],
        "matched": true,
        "original_kept": true,
        "receipt": receipt_path.to_string_lossy(),
        "value_shown": false,
    }))
}

/// Writes one receipt file under `~/.moai-cli/receipts/` (folder `0700`, file `0600`, never
/// replaced). It holds no value.
fn write_home_receipt(home: &Path, name: &str, receipt: &Value) -> Result<PathBuf> {
    let dir = state_dir(home).join("receipts");
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
    let path = dir.join(format!("{name}.json"));
    let mut options = fs::OpenOptions::new();
    options.write(true).create_new(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600).custom_flags(libc::O_NOFOLLOW);
    }
    let mut f = options
        .open(&path)
        .map_err(|_| format!("cannot write {}", path.display()))?;
    f.write_all(receipt.to_string().as_bytes())
        .and_then(|_| f.write_all(b"\n"))
        .map_err(|_| format!("cannot write {}", path.display()))?;
    Ok(path)
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
        // A value of 65_535 bytes is the largest accepted (with its newline the entry is
        // 65_536 bytes, the read limit); 65_536 is refused (pre-audit B1). Literal bounds.
        assert!(read_value(&vec![b'k'; 65_535][..]).is_ok());
        let mut with_newline = vec![b'k'; 65_535];
        with_newline.push(b'\n');
        assert!(read_value(&with_newline[..]).is_ok());
        assert!(read_value(&vec![b'k'; 65_536][..]).is_err());
        assert!(read_value(&vec![b'k'; 65_537][..]).is_err());
        assert!(read_value(&b"\xff\xfe"[..]).is_err());
        assert!(read_value(&b"   \n"[..]).is_err());
    }

    /// The owner rule without a second user (pre-audit N1): another owner is refused even
    /// with private bits; the owner with any group/other bit is refused; 0600/0400 pass.
    #[test]
    fn owner_and_mode_rule() {
        assert_eq!(owner_mode_problem(501, 0o600, 501), None);
        assert_eq!(owner_mode_problem(501, 0o400, 501), None);
        assert_eq!(
            owner_mode_problem(0, 0o600, 501),
            Some(OwnerModeProblem::Owner)
        );
        assert_eq!(
            owner_mode_problem(502, 0o600, 501),
            Some(OwnerModeProblem::Owner)
        );
        for bits in [0o640, 0o604, 0o610, 0o601] {
            assert_eq!(
                owner_mode_problem(501, bits, 501),
                Some(OwnerModeProblem::Mode),
                "{bits:o}"
            );
        }
    }

    /// A HOME that cannot be read makes every entry unsafe instead of skipping the checks
    /// (pre-audit N2).
    #[cfg(unix)]
    #[test]
    fn an_unreadable_home_blocks() {
        let dir = tempfile::TempDir::new().unwrap();
        let file = dir.path().join("f");
        fs::write(&file, "x").unwrap();
        let meta = fs::metadata(&file).unwrap();
        let b = unix_unsafe(&meta, &dir.path().join("missing-home"), &file, false).unwrap();
        assert_eq!(b.code, "key_file_unsafe");
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
