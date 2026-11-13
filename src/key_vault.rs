//! API key vault (Issue #2393): encrypted files under `~/.moai-cli/vault/`, and
//! `key-status | key-set | key-remove` (user decisions, Issue #2393 comments 6777274025 and
//! 6778671862: no OS key store; one encrypted file format on every OS).
//!
//! Layout:
//! - `~/.moai-cli/vault/.master.key`: 32 random bytes made by the first `key-set`, the key
//!   that encrypts every entry (choice A of 6777331104);
//! - `~/.moai-cli/vault/<folder>/<provider>.key`: one entry, where `<folder>` is
//!   `<name>-<first 8 hex of clone_id>` for a project (see [`folder_name`]) or `_shared`;
//! - an entry is `MOAIKV1\n`, a 24-byte random nonce, then the XChaCha20-Poly1305 ciphertext
//!   and tag of the value. The associated data names the folder and the provider, so an entry
//!   copied to another folder or provider name does not decrypt.
//!
//! What this protects and what it does not: an entry copied, synced, pasted or backed up
//! without `.master.key` cannot be read. A program running as the same user can read both
//! files, so it is not kept out (the decision records this limit).
//!
//! Safety rules, checked on the opened handle so a path swapped between check and read is
//! not followed: no symlink, a regular file; on Unix the owner is the owner of `HOME` and no
//! group or other permission bits on the files and folders; bounded sizes. A write goes to a
//! new `0600` file in the same folder, then is renamed over the entry.
//!
//! No value, length or fingerprint ever appears in an answer, error or log. A value comes in
//! on stdin only (never argv), and an interactive terminal is refused: nothing here can turn
//! echo off without a terminal library, so the value would be shown while typed.

use chacha20poly1305::aead::{Aead, AeadCore, KeyInit, OsRng, Payload};
use chacha20poly1305::{XChaCha20Poly1305, XNonce};
use serde_json::{json, Value};
use std::fs;
use std::io::{Read, Write};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// Folder of the vault under the home state directory.
pub const VAULT_DIR: &str = "vault";
/// The machine key file inside the vault folder.
pub const MASTER_KEY_FILE: &str = ".master.key";
/// The folder for keys shared by every project.
pub const SHARED_DIR: &str = "_shared";
/// Largest stdin value read (and the legacy key file bound of #2359).
pub const VALUE_FILE_MAX: u64 = 65_536;
/// Largest value `key-set` stores (pre-audit B1: whatever is stored reads back).
pub const VALUE_MAX: usize = 65_535;
const MAGIC: &[u8] = b"MOAIKV1\n";
const NONCE_LEN: usize = 24;
const TAG_LEN: usize = 16;
const MASTER_LEN: usize = 32;
/// Largest entry file read: the header, the nonce, the largest value and the tag
/// (8 + 24 + 65,535 + 16 = 65,583 bytes).
pub const ENTRY_FILE_MAX: u64 = (MAGIC.len() + NONCE_LEN + VALUE_MAX + TAG_LEN) as u64;

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

/// Where an entry lives: a project's folder or the shared one.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Scope {
    Project(String),
    Shared,
}

impl Scope {
    pub fn folder(&self) -> &str {
        match self {
            Scope::Project(f) => f,
            Scope::Shared => SHARED_DIR,
        }
    }

    /// The `key-set` flag that selects this scope, for the next-action hints.
    fn flag(&self) -> &'static str {
        match self {
            Scope::Project(_) => "",
            Scope::Shared => " --shared",
        }
    }

    pub fn name(&self) -> &'static str {
        match self {
            Scope::Project(_) => "project",
            Scope::Shared => "shared",
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

/// The project folder name (decision 2 of 6778671862, choice 4 of 6777331104):
/// `<name>-<first 8 hex of clone_id>`. `<name>` is the profile's `project_id` when there is
/// one, else the primary worktree's folder name; it is lowercased, every character other than
/// `a-z`, `0-9` and `-` becomes `-`, runs of `-` collapse, `-` is trimmed at both ends, it is
/// cut to 40 characters, and an empty result becomes `project`. The result can never be
/// empty, start with `.`, hold `/`, or be `..` or `_shared`.
pub fn folder_name(project_id: Option<&str>, primary_dir: &str, clone_hex: &str) -> String {
    let raw = project_id
        .filter(|p| !p.trim().is_empty())
        .unwrap_or(primary_dir);
    let mut name = String::new();
    for ch in raw.chars().flat_map(char::to_lowercase) {
        let c = if ch.is_ascii_lowercase() || ch.is_ascii_digit() {
            ch
        } else {
            '-'
        };
        if c == '-' && (name.is_empty() || name.ends_with('-')) {
            continue;
        }
        name.push(c);
    }
    let mut name: String = name.trim_end_matches('-').chars().take(40).collect();
    while name.ends_with('-') {
        name.pop();
    }
    if name.is_empty() {
        name = "project".into();
    }
    let hex: String = clone_hex
        .chars()
        .filter(char::is_ascii_hexdigit)
        .take(8)
        .collect::<String>()
        .to_ascii_lowercase();
    format!("{name}-{hex}")
}

/// The project scope of `dir` from the workspace identity (`identity`), or `None` outside a
/// git work tree. No git command is started here beyond the identity's own.
pub fn project_scope(dir: &Path) -> Result<Option<Scope>> {
    let id = crate::identity::identity(dir)?;
    if id["identity_status"] == "no_git" {
        return Ok(None);
    }
    if id["identity_status"] != "ok" {
        return Err("the workspace identity could not be resolved".into());
    }
    if id["work_tree"].is_null() {
        return Ok(None);
    }
    let clone = id["clone_id"].as_str().unwrap_or("");
    let clone_hex = clone.strip_prefix("sha256:").unwrap_or(clone);
    if clone_hex.len() < 8 {
        return Err("the workspace identity has no clone id".into());
    }
    // The primary worktree is the folder that holds the common git directory.
    let anchor = Path::new(id["clone_anchor"].as_str().unwrap_or(""));
    let primary = if anchor.file_name().is_some_and(|n| n == ".git") {
        anchor.parent().unwrap_or(anchor)
    } else {
        anchor
    };
    let primary_name = primary
        .file_name()
        .map(|n| n.to_string_lossy().trim_end_matches(".git").to_string())
        .unwrap_or_default();
    Ok(Some(Scope::Project(folder_name(
        id["project_id"].as_str(),
        &primary_name,
        clone_hex,
    ))))
}

/// The scope a `key-*` command works on: `--shared`, else the project of `dir`. Outside a
/// git work tree only `--shared` is possible.
pub fn scope_for(dir: &Path, shared: bool) -> Result<Scope> {
    if shared {
        return Ok(Scope::Shared);
    }
    project_scope(dir)?.ok_or_else(|| {
        "not inside a git work tree: run it in the project, or use --shared for a key every project uses".to_string()
    })
}

/// [`scope_for`] the current directory.
pub fn scope_here(shared: bool) -> Result<Scope> {
    let dir = std::env::current_dir().map_err(|_| "the current directory is unreadable")?;
    scope_for(&dir, shared)
}

fn state_dir(home: &Path) -> PathBuf {
    home.join(crate::HOME_STATE_DIR)
}

fn vault_dir(home: &Path) -> PathBuf {
    state_dir(home).join(VAULT_DIR)
}

fn master_path(home: &Path) -> PathBuf {
    vault_dir(home).join(MASTER_KEY_FILE)
}

fn scope_dir(home: &Path, scope: &Scope) -> PathBuf {
    vault_dir(home).join(scope.folder())
}

pub fn entry_path(home: &Path, scope: &Scope, provider: &str) -> PathBuf {
    scope_dir(home, scope).join(format!("{provider}.key"))
}

/// The Unix permission and owner rule for one opened object.
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

/// What is wrong with an object's owner and mode, if anything (pure, so the owner rule can be
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

/// A vault folder's state: absent is fine (nothing stored there yet), anything but a private
/// real directory is unsafe.
fn check_dir(home: &Path, dir: &Path) -> std::result::Result<bool, Blocked> {
    match fs::symlink_metadata(dir) {
        Err(_) => Ok(false),
        Ok(m) if !m.is_dir() => Err(Blocked::new(
            "key_file_unsafe",
            format!("{} is not a real folder", dir.display()),
            "move it away; the folder is created on the next key-set",
        )),
        Ok(m) => match unix_unsafe(&m, home, dir, true) {
            Some(b) => Err(b),
            None => Ok(true),
        },
    }
}

/// Opens a file for reading without following a symlink (and without blocking on a FIFO).
fn open_nofollow(path: &Path) -> std::io::Result<fs::File> {
    let mut options = fs::OpenOptions::new();
    options.read(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.custom_flags(libc::O_NOFOLLOW | libc::O_NONBLOCK);
    }
    options.open(path)
}

/// The bytes of a private file of at most `max` bytes: `Ok(None)` when it does not exist,
/// `Err` when it breaks a safety rule. `hint` is the next action named in a refusal.
fn read_private(
    home: &Path,
    path: &Path,
    max: u64,
    hint: &str,
) -> std::result::Result<Option<Vec<u8>>, Blocked> {
    let unsafe_file = |why: &str| {
        Blocked::new(
            "key_file_unsafe",
            format!("{} {why}", path.display()),
            hint.to_string(),
        )
    };
    let file = match open_nofollow(path) {
        Ok(f) => f,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(None),
        // A symlink (O_NOFOLLOW) or an unopenable file.
        Err(_) => return Err(unsafe_file("is not a readable regular file")),
    };
    let meta = file
        .metadata()
        .map_err(|_| unsafe_file("metadata is unreadable"))?;
    if !meta.is_file() {
        return Err(unsafe_file("is not a regular file"));
    }
    if let Some(b) = unix_unsafe(&meta, home, path, false) {
        return Err(b);
    }
    if meta.len() > max {
        return Err(unsafe_file(&format!("is larger than {max} bytes")));
    }
    let mut bytes = Vec::new();
    file.take(max + 1)
        .read_to_end(&mut bytes)
        .map_err(|_| unsafe_file("is unreadable"))?;
    if bytes.len() as u64 > max {
        return Err(unsafe_file(&format!("is larger than {max} bytes")));
    }
    Ok(Some(bytes))
}

/// The one rule for a value (shared with the launcher, #2359 AC4): surrounding whitespace
/// is dropped; empty, whitespace-only or control-character values are no value.
pub fn usable_value(value: &str) -> Option<String> {
    let value = value.trim();
    (!value.is_empty() && !value.chars().any(char::is_control)).then(|| value.to_string())
}

/// The associated data of an entry: the format, the folder and the provider.
fn aad(folder: &str, provider: &str) -> Vec<u8> {
    format!("moai-cli vault v1|{folder}|{provider}").into_bytes()
}

/// The machine key, when present and valid. A file of another length is unsafe.
fn load_master(home: &Path) -> std::result::Result<Option<[u8; MASTER_LEN]>, Blocked> {
    let path = master_path(home);
    let hint = "restore the vault's .master.key from your backup; without it the stored keys cannot be read";
    match read_private(home, &path, MASTER_LEN as u64, hint)? {
        None => Ok(None),
        Some(bytes) => bytes.try_into().map(Some).map_err(|_| {
            Blocked::new(
                "key_file_unsafe",
                format!("{} is not a {MASTER_LEN}-byte key", path.display()),
                hint,
            )
        }),
    }
}

/// Every `*.key` entry under the vault folder, one folder deep.
fn has_entries(home: &Path) -> bool {
    let Ok(dirs) = fs::read_dir(vault_dir(home)) else {
        return false;
    };
    dirs.flatten().any(|d| {
        fs::read_dir(d.path()).is_ok_and(|mut files| {
            files.any(|f| f.is_ok_and(|f| f.path().extension().is_some_and(|e| e == "key")))
        })
    })
}

/// The machine key for a write: the existing one, else a new random one made with
/// `create_new` (so two concurrent first writes never make two keys). When entries exist but
/// the key does not, nothing is made: a new key could never read those entries.
fn ensure_master(home: &Path) -> Result<[u8; MASTER_LEN]> {
    let to_text = |b: Blocked| format!("{}: {}", b.code, b.reason);
    if let Some(key) = load_master(home).map_err(to_text)? {
        return Ok(key);
    }
    if has_entries(home) {
        return Err(format!(
            "key_file_unsafe: the vault holds entries but {} is missing; restore it, or remove the entries before storing a new key",
            master_path(home).display()
        ));
    }
    let key = XChaCha20Poly1305::generate_key(&mut OsRng);
    let path = master_path(home);
    let mut options = fs::OpenOptions::new();
    options.write(true).create_new(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600).custom_flags(libc::O_NOFOLLOW);
    }
    match options.open(&path) {
        Ok(mut f) => {
            let written = f.write_all(&key).and_then(|_| f.sync_all());
            if written.is_err() {
                let _ = fs::remove_file(&path);
                return Err(format!("cannot write {}", path.display()));
            }
        }
        // Another writer made it first: use that one.
        Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => {}
        Err(_) => return Err(format!("cannot create {}", path.display())),
    }
    load_master(home)
        .map_err(to_text)?
        .ok_or_else(|| format!("{} vanished while it was made", path.display()))
}

/// Encrypts `value` for `folder`/`provider` with a fresh random nonce.
fn seal(key: &[u8; MASTER_LEN], folder: &str, provider: &str, value: &str) -> Result<Vec<u8>> {
    let cipher = XChaCha20Poly1305::new_from_slice(key).map_err(|_| "bad machine key length")?;
    let nonce = XChaCha20Poly1305::generate_nonce(&mut OsRng);
    let aad = aad(folder, provider);
    let sealed = cipher
        .encrypt(
            &nonce,
            Payload {
                msg: value.as_bytes(),
                aad: &aad,
            },
        )
        .map_err(|_| "encryption failed")?;
    let mut out = Vec::with_capacity(MAGIC.len() + NONCE_LEN + sealed.len());
    out.extend_from_slice(MAGIC);
    out.extend_from_slice(&nonce);
    out.extend_from_slice(&sealed);
    Ok(out)
}

/// Decrypts an entry. Anything without the header, too short, altered, moved to another
/// folder or provider, or encrypted with another key is refused without a value.
fn open_sealed(
    key: &[u8; MASTER_LEN],
    folder: &str,
    provider: &str,
    bytes: &[u8],
    path: &Path,
    hint: &str,
) -> std::result::Result<String, Blocked> {
    let refuse = |why: &str| {
        Blocked::new(
            "key_file_unsafe",
            format!("{} {why}", path.display()),
            hint.to_string(),
        )
    };
    let Some(rest) = bytes.strip_prefix(MAGIC) else {
        return Err(refuse("is not a moai-cli vault entry (no MOAIKV1 header)"));
    };
    if rest.len() < NONCE_LEN + TAG_LEN {
        return Err(refuse("is too short to be a vault entry"));
    }
    let (nonce, sealed) = rest.split_at(NONCE_LEN);
    let cipher = XChaCha20Poly1305::new_from_slice(key).map_err(|_| refuse("cannot be read"))?;
    let aad = aad(folder, provider);
    let plain = cipher
        .decrypt(
            XNonce::from_slice(nonce),
            Payload {
                msg: sealed,
                aad: &aad,
            },
        )
        .map_err(|_| {
            refuse("cannot be decrypted with this machine's key (altered, moved or made with another key)")
        })?;
    let text = String::from_utf8(plain).map_err(|_| refuse("does not hold text"))?;
    usable_value(&text).ok_or_else(|| {
        Blocked::new(
            "key_absent",
            format!("{} holds no usable value", path.display()),
            hint.to_string(),
        )
    })
}

/// The value of one scope's entry: `Ok(None)` when no entry exists, `Err` when the vault, the
/// entry or the machine key breaks a rule, or the entry does not decrypt.
pub fn read_entry(
    home: &Path,
    scope: &Scope,
    provider: &str,
) -> std::result::Result<Option<String>, Blocked> {
    check_provider(provider)
        .map_err(|e| Blocked::new("key_absent", e, "use a valid provider name"))?;
    if !check_dir(home, &vault_dir(home))? || !check_dir(home, &scope_dir(home, scope))? {
        return Ok(None);
    }
    let path = entry_path(home, scope, provider);
    let hint = format!(
        "remove it and run: moai-cli key-set {provider}{} (value on stdin)",
        scope.flag()
    );
    let Some(bytes) = read_private(home, &path, ENTRY_FILE_MAX, &hint)? else {
        return Ok(None);
    };
    let Some(key) = load_master(home)? else {
        return Err(Blocked::new(
            "key_file_unsafe",
            format!(
                "{} exists but {} is missing",
                path.display(),
                master_path(home).display()
            ),
            "restore the vault's .master.key from your backup, or remove the entry and store the key again",
        ));
    };
    open_sealed(&key, scope.folder(), provider, &bytes, &path, &hint).map(Some)
}

/// The launcher's read (decision 4 of 6778671862): the project's entry, then the shared one.
/// A project entry that exists but breaks a rule stops here; only an absent one falls
/// through to the shared folder.
pub fn read(
    home: &Path,
    project: Option<&Scope>,
    provider: &str,
) -> std::result::Result<Option<(String, Scope)>, Blocked> {
    if let Some(scope) = project {
        if let Some(value) = read_entry(home, scope, provider)? {
            return Ok(Some((value, scope.clone())));
        }
    }
    Ok(read_entry(home, &Scope::Shared, provider)?.map(|v| (v, Scope::Shared)))
}

/// [`read`] for the project of the current directory (the launcher's entry point). A project
/// that cannot be resolved is a refusal, not a silent skip to the shared folder.
pub fn read_here(
    home: &Path,
    provider: &str,
) -> std::result::Result<Option<(String, Scope)>, Blocked> {
    let unresolved = |why: String| {
        Blocked::new(
            "key_absent",
            format!("cannot resolve the project for the vault: {why}"),
            format!("run it inside the project, or store the key with `moai-cli key-set {provider} --shared`"),
        )
    };
    let dir = std::env::current_dir()
        .map_err(|_| unresolved("the current directory is unreadable".into()))?;
    let project = project_scope(&dir).map_err(unresolved)?;
    read(home, project.as_ref(), provider)
}

/// `key-status`: whether the scope's entry exists, passes the safety rules and decrypts.
/// Never the value, its length or a fingerprint.
pub fn status(home: &Path, scope: &Scope, provider: &str) -> Result<Value> {
    check_provider(provider)?;
    let (entry, blocked) = match read_entry(home, scope, provider) {
        Ok(Some(_)) => ("present", None),
        Ok(None) => ("absent", None),
        Err(b) if b.code == "key_absent" => ("empty", Some(b)),
        Err(b) => ("unsafe", Some(b)),
    };
    Ok(json!({
        "provider": provider,
        "scope": scope.name(),
        "folder": scope.folder(),
        "entry": entry,
        "path": entry_path(home, scope, provider).to_string_lossy(),
        "blocked": blocked.map(|b| b.to_json()),
        "value_shown": false,
    }))
}

/// Reads the value for `key-set` from `input`: at most [`VALUE_FILE_MAX`] bytes, one usable
/// value of at most [`VALUE_MAX`] bytes. The caller refuses an interactive terminal first.
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

/// Creates a vault folder `0700` when absent (its parents too), and checks an existing one.
fn ensure_dir(home: &Path, dir: &Path) -> Result<()> {
    let to_text = |b: Blocked| format!("{}: {}", b.code, b.reason);
    if check_dir(home, dir).map_err(to_text)? {
        return Ok(());
    }
    let mut builder = fs::DirBuilder::new();
    builder.recursive(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::DirBuilderExt;
        builder.mode(0o700);
    }
    builder
        .create(dir)
        .map_err(|_| format!("cannot create {}", dir.display()))?;
    check_dir(home, dir).map_err(to_text)?;
    Ok(())
}

/// `key-set`: encrypts `value` and writes it to a new `0600` file in the scope's folder, then
/// renames it over the entry. Makes the machine key on the first write. Answers without the
/// value.
pub fn set(home: &Path, scope: &Scope, provider: &str, value: &str) -> Result<Value> {
    check_provider(provider)?;
    let value = usable_value(value).ok_or("no usable value")?;
    if value.len() > VALUE_MAX {
        return Err("the value exceeds 65,535 bytes".into());
    }
    let vault = vault_dir(home);
    ensure_dir(home, &vault)?;
    let key = ensure_master(home)?;
    let dir = scope_dir(home, scope);
    ensure_dir(home, &dir)?;
    let entry = entry_path(home, scope, provider);
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
    let sealed = seal(&key, scope.folder(), provider, &value)?;
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
        f.write_all(&sealed)?;
        f.sync_all()?;
        fs::rename(&tmp, &entry)
    })();
    if written.is_err() {
        let _ = fs::remove_file(&tmp);
        return Err(format!("cannot write {}", entry.display()));
    }
    Ok(json!({
        "provider": provider,
        "scope": scope.name(),
        "folder": scope.folder(),
        "path": entry.to_string_lossy(),
        "written": true,
        "encrypted": true,
        "value_shown": false,
    }))
}

/// `key-remove`: removes the scope's entry file only, after the caller's confirmation
/// (`--yes`). The machine key and every other file stay (a legacy key file too).
pub fn remove(home: &Path, scope: &Scope, provider: &str, confirmed: bool) -> Result<Value> {
    check_provider(provider)?;
    // The same folder rule as read and set (pre-audit B1, Issue #2393): a vault or scope folder
    // that is a symlink or not private is refused, so a removal never follows a folder link
    // out of the vault. The entry itself is removed as a name (a symlink entry loses only the
    // link).
    let to_text = |b: Blocked| format!("{}: {}", b.code, b.reason);
    let dirs_present = check_dir(home, &vault_dir(home)).map_err(to_text)?
        && check_dir(home, &scope_dir(home, scope)).map_err(to_text)?;
    let entry = entry_path(home, scope, provider);
    let exists = dirs_present && fs::symlink_metadata(&entry).is_ok();
    if !confirmed {
        return Ok(json!({
            "provider": provider,
            "scope": scope.name(),
            "folder": scope.folder(),
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
        "scope": scope.name(),
        "folder": scope.folder(),
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
        if let Ok(file) = open_nofollow(&path) {
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
fn import_plan_inner(
    home: &Path,
    scope: &Scope,
    provider: &str,
) -> Result<(Value, Option<LegacySource>)> {
    check_provider(provider)?;
    if provider != "glm" {
        return Err("only the glm key has legacy files to import".into());
    }
    let sources: Vec<LegacySource> = LEGACY_GLM_FILES
        .iter()
        .map(|rel| legacy_source(home, rel))
        .collect();
    let dest = entry_path(home, scope, provider);
    let entry = match read_entry(home, scope, provider) {
        Ok(Some(_)) => "present",
        Ok(None) => "absent",
        Err(b) if b.code == "key_absent" => "empty",
        Err(_) => "unsafe",
    };
    let chosen = sources.iter().position(|s| s.token.is_some());
    let (action, reason) = match (entry, chosen) {
        ("present", _) => (
            "already_in_vault",
            "the vault already holds a glm key; nothing is imported",
        ),
        ("unsafe" | "empty", _) => (
            "blocked",
            "key_file_unsafe: the vault entry exists but is unusable; fix or remove it first",
        ),
        (_, None) => ("nothing_to_import", "no legacy file holds a usable key"),
        _ => (
            "import",
            "copy the key from the chosen legacy file into the vault",
        ),
    };
    let chosen_path = chosen.map(|i| sources[i].path.to_string_lossy().into_owned());
    let digest_input = json!({
        "schema_version": 1,
        "provider": provider,
        "folder": scope.folder(),
        "action": action,
        "source": chosen_path,
        "source_identity": chosen.map(|i| sources[i].identity.clone()),
        "destination": dest.to_string_lossy(),
        "destination_entry": entry,
    });
    let digest = crate::digest(digest_input.to_string().as_bytes());
    let plan = json!({
        "provider": provider,
        "scope": scope.name(),
        "folder": scope.folder(),
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
pub fn import_plan(home: &Path, scope: &Scope, provider: &str) -> Result<Value> {
    Ok(import_plan_inner(home, scope, provider)?.0)
}

/// `key-import-apply`: re-plans, refuses unless the digest matches (the files did not change
/// since the preview) and a permission reference is given, copies the key into the vault,
/// reads it back and compares, then records a receipt with paths, time, folder and the
/// comparison result only. The legacy file is left exactly as it was.
pub fn import_apply(
    home: &Path,
    scope: &Scope,
    provider: &str,
    plan_digest: &str,
    permission: &str,
) -> Result<Value> {
    if !crate::contract::permission_reference_ok(permission) {
        return Err("permission reference required".into());
    }
    let (plan, source) = import_plan_inner(home, scope, provider)?;
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
    set(home, scope, provider, token)?;
    let matched = matches!(read_entry(home, scope, provider), Ok(Some(v)) if v == token);
    let unix_ms = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_millis() as u64)
        .unwrap_or(0);
    let receipt = json!({
        "schema_version": 1,
        "operation": "key.import_apply",
        "provider": provider,
        "folder": plan["folder"],
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
        "folder": plan["folder"],
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

    /// Decision 2 of 6778671862: `<name>-<8 hex>`. User text becomes a path component, so
    /// every unsafe shape is pinned by a literal.
    #[test]
    fn folder_names_are_safe_path_components() {
        let hex = "0123abcdef99";
        for (id, dir, want) in [
            (Some("moai-adk"), "x", "moai-adk-0123abcd"),
            (None, "MoAI ADK", "moai-adk-0123abcd"),
            (Some("  "), "repo", "repo-0123abcd"),
            (Some("../../etc"), "x", "etc-0123abcd"),
            (Some("a/b"), "x", "a-b-0123abcd"),
            (Some(".hidden"), "x", "hidden-0123abcd"),
            (Some(".."), "x", "project-0123abcd"),
            (Some("_shared"), "x", "shared-0123abcd"),
            (None, "", "project-0123abcd"),
            (Some("한글"), "x", "project-0123abcd"),
            (Some("A--B__C"), "x", "a-b-c-0123abcd"),
        ] {
            assert_eq!(folder_name(id, dir, hex), want, "{id:?} {dir:?}");
        }
        let long = "a".repeat(60);
        let got = folder_name(Some(&long), "x", hex);
        assert_eq!(got, format!("{}-0123abcd", "a".repeat(40)));
        // A name cut inside a run of '-' never ends with '-' before the hex.
        let edge = format!("{}-b", "a".repeat(39));
        assert_eq!(
            folder_name(Some(&edge), "x", hex),
            format!("{}-0123abcd", "a".repeat(39))
        );
        for (id, dir) in [(Some("x/../y"), ""), (None, "."), (None, "/")] {
            let f = folder_name(id, dir, hex);
            assert!(
                !f.contains('/') && !f.starts_with('.') && f != SHARED_DIR,
                "{f}"
            );
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
    fn stdin_values_are_bounded() {
        // 65,535 bytes is the largest value; 65,536 is refused. Literal bounds.
        assert!(read_value(&vec![b'k'; 65_535][..]).is_ok());
        let mut with_newline = vec![b'k'; 65_535];
        with_newline.push(b'\n');
        assert!(read_value(&with_newline[..]).is_ok());
        assert!(read_value(&vec![b'k'; 65_536][..]).is_err());
        assert!(read_value(&vec![b'k'; 65_537][..]).is_err());
        assert!(read_value(&b"\xff\xfe"[..]).is_err());
        assert!(read_value(&b"   \n"[..]).is_err());
        assert_eq!(ENTRY_FILE_MAX, 65_583);
    }

    /// The owner rule without a second user (pre-audit N1).
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

    /// A HOME that cannot be read makes every file unsafe instead of skipping the checks.
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

    fn sealed_entry(home: &Path, scope: &Scope, value: &str) -> PathBuf {
        set(home, scope, "glm", value).unwrap();
        entry_path(home, scope, "glm")
    }

    /// The format: header, fresh nonce per write, AAD bound to folder and provider, tag
    /// checked, and no value in any refusal.
    #[test]
    fn entries_are_sealed_and_bound_to_their_place() {
        let dir = tempfile::TempDir::new().unwrap();
        let home = dir.path();
        let value = format!("moai-test-key-{}", std::process::id());
        let p = Scope::Project("demo-0123abcd".into());
        let path = sealed_entry(home, &p, &value);
        let bytes = fs::read(&path).unwrap();
        assert!(bytes.starts_with(b"MOAIKV1\n"));
        assert_eq!(bytes.len(), 8 + 24 + value.len() + 16);
        assert!(!bytes.windows(value.len()).any(|w| w == value.as_bytes()));
        assert_eq!(
            read_entry(home, &p, "glm").unwrap().as_deref(),
            Some(value.as_str())
        );
        // A second write of the same value uses a new nonce.
        set(home, &p, "glm", &value).unwrap();
        assert_ne!(fs::read(&path).unwrap(), bytes);
        let check = |b: Blocked| {
            assert_eq!(b.code, "key_file_unsafe");
            assert!(!b.reason.contains(&value) && !b.next.contains(&value));
        };
        // One flipped byte in the nonce, the ciphertext or the tag: refused.
        for at in [8, 8 + 24, bytes.len() - 1] {
            let mut bad = bytes.clone();
            bad[at] ^= 1;
            fs::write(&path, &bad).unwrap();
            let b = read_entry(home, &p, "glm").unwrap_err();
            assert!(
                b.reason.contains("cannot be decrypted"),
                "{at}: {}",
                b.reason
            );
            check(b);
        }
        // No header (a plain-text file from before the format): refused, not used.
        fs::write(&path, format!("{value}\n")).unwrap();
        let b = read_entry(home, &p, "glm").unwrap_err();
        assert!(b.reason.contains("no MOAIKV1 header"), "{}", b.reason);
        check(b);
        // Copied to the shared folder or to another provider name: does not decrypt.
        fs::write(&path, &bytes).unwrap();
        // A private shared folder, so only the associated data can refuse the copy.
        ensure_dir(home, &scope_dir(home, &Scope::Shared)).unwrap();
        fs::copy(&path, entry_path(home, &Scope::Shared, "glm")).unwrap();
        let b = read_entry(home, &Scope::Shared, "glm").unwrap_err();
        assert!(b.reason.contains("cannot be decrypted"), "{}", b.reason);
        check(b);
        fs::copy(&path, entry_path(home, &p, "zai")).unwrap();
        let b = read_entry(home, &p, "zai").unwrap_err();
        assert!(b.reason.contains("cannot be decrypted"), "{}", b.reason);
        check(b);
        // Another machine key: does not decrypt.
        fs::write(master_path(home), [7u8; 32]).unwrap();
        let b = read_entry(home, &p, "glm").unwrap_err();
        assert!(b.reason.contains("cannot be decrypted"), "{}", b.reason);
        check(b);
    }

    /// The machine key: made once, 32 bytes, never replaced while entries exist.
    #[test]
    fn the_machine_key_is_made_once_and_never_silently_replaced() {
        let dir = tempfile::TempDir::new().unwrap();
        let home = dir.path();
        sealed_entry(home, &Scope::Shared, "moai-test-a");
        let master = fs::read(master_path(home)).unwrap();
        assert_eq!(master.len(), 32);
        sealed_entry(home, &Scope::Shared, "moai-test-b");
        assert_eq!(fs::read(master_path(home)).unwrap(), master);
        // Lost key with entries left: no new key, the write is refused, the read too.
        fs::remove_file(master_path(home)).unwrap();
        let err = set(home, &Scope::Shared, "glm", "moai-test-c").unwrap_err();
        assert!(err.contains("missing"), "{err}");
        assert!(!master_path(home).exists());
        let b = read_entry(home, &Scope::Shared, "glm").unwrap_err();
        assert_eq!(b.code, "key_file_unsafe");
        // A private key file of another length: unsafe for its length (31 and 33 bytes; 32 is
        // the only length). Mode 0600 so only the length can refuse it.
        for (len, why) in [
            (31usize, "is not a 32-byte key"),
            (33, "is larger than 32 bytes"),
        ] {
            fs::write(master_path(home), vec![1u8; len]).unwrap();
            #[cfg(unix)]
            {
                use std::os::unix::fs::PermissionsExt;
                fs::set_permissions(master_path(home), fs::Permissions::from_mode(0o600)).unwrap();
            }
            let b = load_master(home).unwrap_err();
            assert_eq!(b.code, "key_file_unsafe", "{len}");
            assert!(b.reason.ends_with(why), "{len}: {}", b.reason);
        }
        fs::write(master_path(home), [1u8; 32]).unwrap();
        assert_eq!(load_master(home).unwrap(), Some([1u8; 32]));
    }

    /// Decision 4: the project's entry first, the shared one only when the project has none;
    /// a broken project entry does not fall through.
    #[test]
    fn the_project_comes_before_the_shared_folder() {
        let dir = tempfile::TempDir::new().unwrap();
        let home = dir.path();
        let p = Scope::Project("demo-0123abcd".into());
        assert_eq!(read(home, Some(&p), "glm").unwrap(), None);
        sealed_entry(home, &Scope::Shared, "moai-test-shared");
        let (v, s) = read(home, Some(&p), "glm").unwrap().unwrap();
        assert_eq!((v.as_str(), s), ("moai-test-shared", Scope::Shared));
        assert_eq!(read(home, None, "glm").unwrap().unwrap().1, Scope::Shared);
        sealed_entry(home, &p, "moai-test-project");
        let (v, s) = read(home, Some(&p), "glm").unwrap().unwrap();
        assert_eq!((v.as_str(), s), ("moai-test-project", p.clone()));
        fs::write(entry_path(home, &p, "glm"), b"MOAIKV1\nbroken").unwrap();
        assert_eq!(
            read(home, Some(&p), "glm").unwrap_err().code,
            "key_file_unsafe"
        );
    }

    /// The entry bound: an entry of exactly the largest size is read; one byte more is not.
    #[test]
    fn the_entry_file_bound_is_65583_bytes() {
        let dir = tempfile::TempDir::new().unwrap();
        let home = dir.path();
        let max = "k".repeat(65_535);
        let path = sealed_entry(home, &Scope::Shared, &max);
        assert_eq!(fs::metadata(&path).unwrap().len(), 65_583);
        assert_eq!(
            read_entry(home, &Scope::Shared, "glm")
                .unwrap()
                .unwrap()
                .len(),
            65_535
        );
        let mut bytes = fs::read(&path).unwrap();
        bytes.push(0);
        fs::write(&path, &bytes).unwrap();
        let b = read_entry(home, &Scope::Shared, "glm").unwrap_err();
        assert!(b.reason.contains("larger than 65583 bytes"), "{}", b.reason);
    }
}
