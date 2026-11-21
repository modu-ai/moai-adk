//! SPEC-20261121-retired-hosts-read-safely (REQ-9): the backup code that the hook files and the
//! Go-era state preview share with the Codex config writer. It moved here unchanged so the Codex
//! config module can be deleted without taking the hooks with it.
use std::fs;
use std::path::{Path, PathBuf};

type Result<T> = std::result::Result<T, String>;

/// Size bound for the config file, its receipt and the rendered result.
pub const CONFIG_LIMIT: u64 = 262_144;

/// The copy of the config a replacement keeps (SPEC-20261119-codex-conflict-choices). The hook
/// files keep theirs the same way (SPEC-20261120-hooks-conflict-choices), so the writer is shared.
pub(crate) struct Backup {
    pub(crate) rel: String,
    pub(crate) path: PathBuf,
    pub(crate) bytes: Vec<u8>,
}

/// The folder of the copies, under the project state folder, which the managed `.gitignore`
/// block already ignores; a copy is named by the digest of the bytes it holds.
pub(crate) const BACKUP_DIR: &str = "backups";

/// A regular file of at most [`CONFIG_LIMIT`] bytes, `None` when absent; a link is refused.
fn read_existing(path: &Path) -> Result<Option<Vec<u8>>> {
    match fs::symlink_metadata(path) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
        Err(_) => Err("target metadata unavailable".into()),
        Ok(m) if m.file_type().is_symlink() => Err("symlink target refused".into()),
        Ok(m) if !m.is_file() || m.len() > CONFIG_LIMIT => {
            Err("target is not a bounded regular file".into())
        }
        Ok(_) => fs::read(path)
            .map(Some)
            .map_err(|_| "target read failed".into()),
    }
}

/// Keeps the copy of the config before a replacement. A copy of the same bytes left by an
/// interrupted run is reused; a file of that name with other bytes is never overwritten.
pub(crate) fn write_backup(b: &Backup) -> Result<()> {
    match read_existing(&b.path)? {
        Some(held) if held == b.bytes => Ok(()),
        Some(_) => Err(format!(
            "backup {} holds other bytes; exclusive create refused",
            b.rel
        )),
        // Private: the config can hold secrets (an MCP server's `env`), and the copy must not
        // be more readable than the file it keeps, whatever the original's mode was.
        None => crate::apply::publish_exclusive_mode(&b.path, &b.bytes, Some(0o600)),
    }
}
