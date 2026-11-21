//! The MoAI git-hook feature (SPEC-20261121-go-path-removal): recognise the git hooks `moai`
//! manages in a repository, preview the change, retire the Go-generation hooks and install the
//! Rust hook texts in their place, with a revert record. The managed names are `pre-commit`,
//! `pre-push` and `commit-msg`; the texts live in `install/git-hooks/` and travel inside the
//! binary through `include_str!`, so an install writes only the repository's own hooks folder.
//!
//! Recognition: a hook beside a `.moai-<name>.sha256` hash file is Go generation (that is how
//! the Go build pinned its hooks); a hook whose header carries the `MoAI` mark without the
//! `(Rust)` mark is Go generation too; a hook with both marks is this generation's; a hook with
//! no marks at all belongs to the user and is never touched — not retired, not overwritten, and
//! its name blocks the install of that hook. A hash file whose hook is gone is a leftover.
//!
//! Whenever a run changes a hook it brings the whole managed set to current: the free names
//! join the install, and the names the user holds stay the user's.
//!
//! The apply saves every retired byte (hook and hash file) in
//! `.git/hooks/moai-retired-<stamp>/`; that folder is the revert record, and the revert moves
//! the bytes back only over hooks this feature installed (byte-equal to its own texts) or that
//! were absent. The update layer and the CLI verbs wire this module later (pair requests on
//! the cards that hold `src/update.rs`, `src/lib.rs` and `src/main.rs`); the tests compile it
//! by path, so it stands alone on `std`, `serde_json` and `sha2`.

use serde_json::{json, Value};
use sha2::{Digest, Sha256};
use std::fs;
use std::io::{Read, Write};
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// The hooks this feature manages, in the order the preview lists them.
pub const MANAGED: [&str; 3] = ["pre-commit", "pre-push", "commit-msg"];

/// The shipped hook texts — the exact bytes an install writes and a revert compares.
pub const PRE_COMMIT_TEXT: &str = include_str!("../../install/git-hooks/pre-commit");
pub const PRE_PUSH_TEXT: &str = include_str!("../../install/git-hooks/pre-push");
pub const COMMIT_MSG_TEXT: &str = include_str!("../../install/git-hooks/commit-msg");

fn text_of(name: &str) -> &'static str {
    match name {
        "pre-commit" => PRE_COMMIT_TEXT,
        "pre-push" => PRE_PUSH_TEXT,
        _ => COMMIT_MSG_TEXT,
    }
}

/// Largest hook read: 64 KiB is read, one byte more is "larger" (and then kept as a user hook).
const MAX_BYTES: usize = 65536;
/// How many header lines the mark check reads.
const HEADER_LINES: usize = 10;
/// The token prefix, so a token of another part is never taken for this one.
const PREFIX: &str = "git-hooks:";
const NOT_REVIEWED: &str =
    "the git-hooks token differs from the preview's; nothing applied, preview again";

/// What one managed hook is.
#[derive(Clone, Copy, PartialEq)]
enum Kind {
    /// Go generation: the hash file, or the header says MoAI without `(Rust)`.
    GoGeneration,
    /// This generation: the header carries `MoAI` and `(Rust)`.
    Current,
    /// No MoAI marks: the user's own hook, never touched.
    User,
    /// No file of that name.
    Absent,
}

/// One managed hook as the scan found it.
struct Found {
    kind: Kind,
    /// `"hash file"` or `"header"` for a Go hook, empty otherwise.
    evidence: &'static str,
    /// The bytes read back (`None` when the file is absent or unreadable).
    bytes: Option<Vec<u8>>,
    /// Whether the Go-generation hash file sits beside the hook.
    hash_present: bool,
}

/// The hooks folder of the repository: `.git/hooks`, or the hooks folder of the real git dir a
/// linked worktree's `.git` file names. Writes nothing.
fn hooks_folder(repo_root: &Path) -> Result<PathBuf> {
    let dot_git = repo_root.join(".git");
    let git_dir = match fs::symlink_metadata(&dot_git) {
        Ok(m) if m.is_dir() => dot_git,
        Ok(_) => {
            let text = fs::read_to_string(&dot_git)
                .map_err(|_| format!("{} could not be read", dot_git.display()))?;
            let line = text
                .lines()
                .find(|l| l.starts_with("gitdir:"))
                .ok_or_else(|| format!("{} does not name a git dir", dot_git.display()))?;
            let stated = PathBuf::from(line["gitdir:".len()..].trim());
            if stated.is_absolute() {
                stated
            } else {
                repo_root.join(stated)
            }
        }
        Err(_) => return Err(format!("{} holds no .git folder", repo_root.display())),
    };
    Ok(git_dir.join("hooks"))
}

/// The first `MAX_BYTES + 1` bytes of a regular file; `None` when it does not open.
fn read_bounded(p: &Path) -> Option<Vec<u8>> {
    let mut bytes = Vec::new();
    fs::File::open(p)
        .ok()?
        .take(MAX_BYTES as u64 + 1)
        .read_to_end(&mut bytes)
        .ok()?;
    Some(bytes)
}

fn hex_of(bytes: &[u8]) -> String {
    let mut h = Sha256::new();
    h.update(bytes);
    h.finalize().iter().map(|b| format!("{b:02x}")).collect()
}

/// A file that must not exist yet, written whole, mode set exactly (the umask is not trusted),
/// flushed to disk.
fn create_new(path: &Path, bytes: &[u8]) -> std::io::Result<()> {
    let mut f = fs::OpenOptions::new()
        .write(true)
        .create_new(true)
        .open(path)?;
    f.write_all(bytes)?;
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        f.set_permissions(fs::Permissions::from_mode(0o755))?;
    }
    f.sync_all()
}

fn absent(p: &Path) -> bool {
    matches!(fs::symlink_metadata(p), Err(e) if e.kind() == std::io::ErrorKind::NotFound)
}

/// Recognise one managed hook in the folder. A hook larger than 64 KiB or unreadable counts as
/// the user's: the feature never touches what it cannot read.
fn recognize(folder: &Path, name: &str) -> Found {
    let bytes = read_bounded(&folder.join(name));
    let hash_present = folder.join(format!(".moai-{name}.sha256")).exists();
    let header: String = bytes
        .as_deref()
        .map(|b| String::from_utf8_lossy(b).to_string())
        .unwrap_or_default()
        .lines()
        .take(HEADER_LINES)
        .collect::<Vec<_>>()
        .join("\n");
    let adk_mark = header.contains("MoAI-ADK");
    let moai_mark = header.contains("MoAI");
    let rust_mark = header.contains("(Rust");
    let (kind, evidence) = if hash_present {
        (Kind::GoGeneration, "hash file")
    } else if adk_mark {
        (Kind::GoGeneration, "header")
    } else if bytes.is_none() {
        (Kind::Absent, "")
    } else if moai_mark && rust_mark {
        (Kind::Current, "")
    } else {
        (Kind::User, "")
    };
    Found {
        kind,
        evidence,
        bytes,
        hash_present,
    }
}

/// `YYYY-MM-DDTHH:MM:SSZ` -> `YYYYMMDDTHHMMSSZ`, or `None` when `now` is not that shape.
fn compact_utc(now: &str) -> Option<String> {
    let b = now.as_bytes();
    let shape = b.len() == 20
        && b.iter().enumerate().all(|(i, c)| match i {
            4 | 7 => *c == b'-',
            10 => *c == b'T',
            13 | 16 => *c == b':',
            19 => *c == b'Z',
            _ => c.is_ascii_digit(),
        });
    shape.then(|| now.chars().filter(|c| c.is_ascii_alphanumeric()).collect())
}

/// `YYYYMMDDTHHMMSSZ`: the stamp a retire folder carries.
fn stamp_ok(stamp: &str) -> bool {
    let b = stamp.as_bytes();
    b.len() == 16
        && b.iter().enumerate().all(|(i, c)| match i {
            8 => *c == b'T',
            15 => *c == b'Z',
            _ => c.is_ascii_digit(),
        })
}

/// The preview of one repository's hooks folder. Writes nothing.
///
/// The facts the token covers are the per-hook kind, evidence, byte digest and hash-file
/// presence, so any change to any hook invalidates the preview.
pub fn plan(repo_root: &Path) -> Result<Value> {
    let folder = hooks_folder(repo_root)?;
    let mut go_hooks = Vec::new();
    let mut kept = Vec::new();
    let mut install: Vec<&str> = Vec::new();
    let mut current_hooks = Vec::new();
    let mut hash_leftovers = Vec::new();
    let mut facts = Vec::new();
    let mut replaces = 0usize;
    for name in MANAGED {
        let found = recognize(&folder, name);
        let digest = found
            .bytes
            .as_ref()
            .map(|b| hex_of(b))
            .unwrap_or_else(|| "absent".to_string());
        facts.push(json!({
            "name": name,
            "kind": match found.kind {
                Kind::GoGeneration => "go",
                Kind::Current => "current",
                Kind::User => "user",
                Kind::Absent => "absent",
            },
            "evidence": found.evidence,
            "sha256": digest,
            "hash_file": found.hash_present,
        }));
        match found.kind {
            Kind::GoGeneration => {
                go_hooks.push(json!({"name": name, "evidence": found.evidence}));
                install.push(name);
            }
            Kind::Current => {
                let action = if found.bytes.as_deref() == Some(text_of(name).as_bytes()) {
                    "already current"
                } else {
                    replaces += 1;
                    install.push(name);
                    "replace"
                };
                current_hooks.push(json!({"name": name, "action": action}));
            }
            Kind::User => {
                kept.push(json!({"name": name, "reason": "no MoAI marks"}));
            }
            Kind::Absent => {}
        }
    }
    // The run changes a hook when it retires a Go hook or replaces a drifted one; then the
    // free managed names join the install, and a hash file without its hook is a leftover to
    // clean with the same revert record.
    let changes = !go_hooks.is_empty() || replaces > 0;
    if changes {
        for name in MANAGED {
            if install.contains(&name) {
                continue;
            }
            if recognize(&folder, name).kind == Kind::Absent {
                install.push(name);
            }
        }
        for name in MANAGED {
            let is_retired =
                go_hooks.iter().any(|g| g["name"] == json!(name)) || install.contains(&name);
            if !is_retired && folder.join(format!(".moai-{name}.sha256")).exists() {
                hash_leftovers.push(json!({"name": name}));
            }
        }
    }
    let writes = !install.is_empty();
    let token = if writes {
        Some(format!(
            "{PREFIX}{}",
            &hex_of(json!({ "facts": facts }).to_string().as_bytes())[..16]
        ))
    } else {
        None
    };
    Ok(json!({
        "hooks_folder": folder.display().to_string(),
        "go_hooks": go_hooks,
        "current_hooks": current_hooks,
        "kept": kept,
        "hash_leftovers": hash_leftovers,
        "install": install,
        "writes": writes,
        "confirm_token": token,
        "apply_flag": token.as_ref().map(|t| format!("--confirm-git-hooks {t}")),
        "saved_record": writes.then_some("moai-retired-<stamp>"),
    }))
}

/// Apply the previewed change: retire the Go hooks (and the replaced ones) into
/// `moai-retired-<stamp>/`, install the free managed names, and report what was kept. Every
/// condition is checked before the first write.
pub fn apply(repo_root: &Path, token: &str, now: &str) -> Result<Value> {
    let stamp = compact_utc(now)
        .ok_or("the clock is not a UTC time (YYYY-MM-DDTHH:MM:SSZ); nothing written")?;
    let fresh = plan(repo_root)?;
    if fresh["confirm_token"].as_str() != Some(token) {
        return Err(NOT_REVIEWED.into());
    }
    if !fresh["writes"].as_bool().unwrap_or(false) {
        return Err(
            "nothing to apply: the hooks are already in the shape the preview describes".into(),
        );
    }
    let folder = hooks_folder(repo_root)?;
    let retire = folder.join(format!("moai-retired-{stamp}"));
    // Every check, before any write.
    if !absent(&retire) {
        return Err(format!(
            "{} already exists; nothing written",
            retire.display()
        ));
    }
    let go_rows = fresh["go_hooks"].as_array().cloned().unwrap_or_default();
    let mut save_names: Vec<&str> = Vec::new();
    for row in &go_rows {
        let name = row["name"].as_str().ok_or("the preview is malformed")?;
        if recognize(&folder, name).kind != Kind::GoGeneration {
            return Err(format!(
                "{name} is not a Go-generation hook any more; nothing written, preview again"
            ));
        }
        save_names.push(name);
    }
    for name in MANAGED {
        let found = recognize(&folder, name);
        if found.kind == Kind::Current && found.bytes.as_deref() != Some(text_of(name).as_bytes()) {
            save_names.push(name);
        }
    }
    // The first write: the retire folder, then one saved copy per retired hook.
    fs::create_dir(&retire)
        .map_err(|e| format!("creating {}: {e}; nothing written", retire.display()))?;
    let mut retired = Vec::new();
    let mut hash_moved = Vec::new();
    for name in save_names {
        let source = folder.join(name);
        if !absent(&source) {
            let bytes = fs::read(&source).map_err(|e| {
                format!(
                    "reading {name}: {e}; the retire folder is {}",
                    retire.display()
                )
            })?;
            let saved_path = retire.join(name);
            create_new(&saved_path, &bytes).map_err(|e| {
                format!(
                    "saving {}: {e}; the retire folder is {}",
                    saved_path.display(),
                    retire.display()
                )
            })?;
            if read_bounded(&saved_path).as_deref() != Some(bytes.as_slice()) {
                let _ = fs::remove_file(&saved_path);
                return Err(format!(
                    "the saved copy {} did not read back equal; nothing retired",
                    saved_path.display()
                ));
            }
            fs::remove_file(&source).map_err(|e| {
                format!(
                    "removing {name}: {e}; the saved copy is {}",
                    saved_path.display()
                )
            })?;
            retired.push(json!({"name": name}));
        }
        let hash = folder.join(format!(".moai-{name}.sha256"));
        if hash.exists() {
            let hash_bytes = fs::read(&hash).map_err(|e| {
                format!(
                    "reading the hash file of {name}: {e}; the saved hook is {}",
                    retire.join(name).display()
                )
            })?;
            let hash_saved = retire.join(format!(".moai-{name}.sha256"));
            create_new(&hash_saved, &hash_bytes)
                .map_err(|e| format!("saving {}: {e}", hash_saved.display()))?;
            fs::remove_file(&hash).map_err(|e| {
                format!(
                    "removing the hash file of {name}: {e}; the saved copy is {}",
                    hash_saved.display()
                )
            })?;
            hash_moved.push(json!({"name": name}));
        }
    }
    let leftover_rows = fresh["hash_leftovers"]
        .as_array()
        .cloned()
        .unwrap_or_default();
    for row in &leftover_rows {
        let name = row["name"].as_str().ok_or("the preview is malformed")?;
        let hash = folder.join(format!(".moai-{name}.sha256"));
        if hash.exists() {
            let hash_bytes =
                fs::read(&hash).map_err(|e| format!("reading the hash file of {name}: {e}"))?;
            let hash_saved = retire.join(format!(".moai-{name}.sha256"));
            create_new(&hash_saved, &hash_bytes)
                .map_err(|e| format!("saving {}: {e}", hash_saved.display()))?;
            fs::remove_file(&hash).map_err(|e| {
                format!(
                    "removing the hash file of {name}: {e}; the saved copy is {}",
                    hash_saved.display()
                )
            })?;
            hash_moved.push(json!({"name": name}));
        }
    }
    let install_rows = fresh["install"].as_array().cloned().unwrap_or_default();
    let mut installed = Vec::new();
    for row in &install_rows {
        let name = row.as_str().ok_or("the preview is malformed")?;
        let target = folder.join(name);
        if target.exists() {
            fs::remove_file(&target).map_err(|e| {
                format!(
                    "replacing {name}: {e}; the saved copies are in {}",
                    retire.display()
                )
            })?;
        }
        create_new(&target, text_of(name).as_bytes()).map_err(|e| {
            format!(
                "installing {name}: {e}; the saved copies are in {}",
                retire.display()
            )
        })?;
        if read_bounded(&target).as_deref() != Some(text_of(name).as_bytes()) {
            return Err(format!(
                "{} did not read back as the shipped text",
                target.display()
            ));
        }
        installed.push(json!({"name": name}));
    }
    Ok(json!({
        "state": "applied",
        "stamp": stamp,
        "retired": retired,
        "hash_files_moved": hash_moved,
        "installed": installed,
        "kept": fresh["kept"],
        "folder": retire.display().to_string(),
        "revert": format!("revert with the feature's revert and the stamp {stamp}"),
    }))
}

/// Put the retired hooks back: only over hooks this feature installed (byte-equal to its own
/// texts) or that were absent, by rename, with every condition checked before the first move.
pub fn revert(repo_root: &Path, stamp: &str) -> Result<Value> {
    if !stamp_ok(stamp) {
        return Err("the stamp must look like 20261121T194000Z; nothing changed".into());
    }
    let folder = hooks_folder(repo_root)?;
    let retire = folder.join(format!("moai-retired-{stamp}"));
    let meta = fs::symlink_metadata(&retire)
        .map_err(|_| format!("no saved copy {} here; nothing changed", retire.display()))?;
    if !meta.is_dir() {
        return Err(format!(
            "{} is not a folder; nothing changed",
            retire.display()
        ));
    }
    // What the folder holds: hook files by their managed names, hash files by theirs.
    let mut saved_hooks: Vec<String> = Vec::new();
    let mut saved_hashes: Vec<String> = Vec::new();
    for entry in fs::read_dir(&retire).map_err(|e| format!("reading {}: {e}", retire.display()))? {
        let entry = entry.map_err(|e| format!("reading {}: {e}", retire.display()))?;
        let name = entry.file_name().to_string_lossy().to_string();
        if name.starts_with(".moai-") && name.ends_with(".sha256") {
            saved_hashes.push(name);
        } else if MANAGED.contains(&name.as_str()) {
            saved_hooks.push(name);
        } else {
            return Err(format!(
                "{} holds a file this feature did not save ({name}); nothing changed",
                retire.display()
            ));
        }
    }
    // Every check, before any move.
    for name in &saved_hooks {
        let target = folder.join(name);
        match read_bounded(&target) {
            None => {}
            Some(bytes) if bytes == text_of(name).as_bytes() => {}
            Some(_) => {
                return Err(format!(
                    "the hook at {} is not the one this feature installed; nothing changed",
                    target.display()
                ))
            }
        }
    }
    for name in &saved_hashes {
        let target = folder.join(name);
        if !absent(&target) {
            return Err(format!(
                "{} already exists; nothing changed",
                target.display()
            ));
        }
    }
    // The installed hooks with no saved counterpart go away — but only while they carry this
    // feature's bytes.
    let mut removed = Vec::new();
    for name in MANAGED {
        if saved_hooks.iter().any(|s| s == name) {
            continue;
        }
        let target = folder.join(name);
        match read_bounded(&target) {
            None => {}
            Some(bytes) if bytes == text_of(name).as_bytes() => {
                removed.push(json!({"name": name}));
            }
            Some(_) => {
                return Err(format!(
                    "the hook at {} is not the one this feature installed; nothing changed",
                    target.display()
                ))
            }
        }
    }
    // The moves.
    for name in &saved_hooks {
        let target = folder.join(name);
        if target.exists() {
            fs::remove_file(&target).map_err(|e| format!("making room for {name}: {e}"))?;
        }
        fs::rename(retire.join(name), &target).map_err(|e| format!("putting {name} back: {e}"))?;
    }
    for name in &saved_hashes {
        let target = folder.join(name);
        fs::rename(retire.join(name), &target).map_err(|e| format!("putting {name} back: {e}"))?;
    }
    for row in &removed {
        let name = row["name"].as_str().unwrap_or_default();
        fs::remove_file(folder.join(name)).map_err(|e| format!("removing {name}: {e}"))?;
    }
    fs::remove_dir(&retire).map_err(|e| format!("removing {}: {e}", retire.display()))?;
    Ok(json!({
        "state": "reverted",
        "stamp": stamp,
        "restored": saved_hooks,
        "hash_files_restored": saved_hashes,
        "removed": removed,
    }))
}
