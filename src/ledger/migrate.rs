//! Schema version rules of the ledger (Issue #3068 P1) and the one repair of an existing file
//! that exists so far (SPEC-20261121-ledger-card-landed-migration).
//!
//! - [`check_version`] decides whether a file's `user_version` may be opened. Version 1 is still
//!   the only version.
//! - [`upgrade_done_guard`] gives a version 1 ledger made before the M-Flow `card.landed`
//!   exception the current `card_done_guard` trigger. `open_write` runs it on every writable
//!   open; a ledger that already has the exception is not touched.
//!
//! The upgrade runs in one `BEGIN IMMEDIATE` transaction. Inside it, after the trigger was read
//! again under the lock, the ledger file is copied to `ledger.db.bak-<UTC>` beside it (the
//! reserved lock keeps every other writer out and nothing has been written yet, so the copy is
//! consistent) and the copy is read back; only then is the trigger replaced and a `meta` row
//! written. Any error rolls the transaction back and removes the copy this call made. No row of
//! any data table is touched. `PRAGMA user_version` stays 1 (SPEC decision D-1), so an older
//! build still opens the file.

use super::{schema, LedgerError, APPLICATION_ID, SCHEMA_VERSION};
use rusqlite::{Connection, OpenFlags, Transaction, TransactionBehavior};
use std::path::{Path, PathBuf};

/// Whether a file at `found` may be opened by this build: `found` equals [`SCHEMA_VERSION`].
/// A newer file is `schema_newer` and is left alone; a file below 1 has no schema (an empty or
/// foreign database) and is `ledger_unavailable`.
pub fn check_version(found: i64) -> Result<(), LedgerError> {
    match found {
        v if v == SCHEMA_VERSION => Ok(()),
        v if v > SCHEMA_VERSION => Err(LedgerError::SchemaNewer {
            found: v,
            supported: SCHEMA_VERSION,
        }),
        v => Err(LedgerError::Unavailable(format!(
            "the file has schema version {v}, not a ledger"
        ))),
    }
}

/// The `meta` key that records an upgrade: `<DB clock seconds>|<backup file name>`.
pub const DONE_GUARD_META_KEY: &str = "done_guard_landed";

/// What a ledger's `card_done_guard` trigger says about the `card.landed` exception.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum DoneGuard {
    /// The trigger carries the exception; nothing to do.
    Current,
    /// The trigger lacks the exception (or is absent); [`upgrade_done_guard`] replaces it.
    NeedsMigration,
}

/// What an upgrade did.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Upgraded {
    /// The backup copy made before the trigger was replaced.
    pub backup: PathBuf,
}

/// Reads the trigger text and says whether it carries the `card.landed` exception. Writes nothing,
/// so it works on a read-only connection too.
pub fn done_guard_status(conn: &Connection) -> Result<DoneGuard, LedgerError> {
    let sql: Option<String> = match conn.query_row(
        "SELECT sql FROM sqlite_master WHERE type = 'trigger' AND name = 'card_done_guard'",
        [],
        |r| r.get(0),
    ) {
        Ok(sql) => Some(sql),
        Err(rusqlite::Error::QueryReturnedNoRows) => None,
        Err(e) => return Err(e.into()),
    };
    Ok(match sql {
        Some(sql) if sql.contains("'card.landed'") => DoneGuard::Current,
        _ => DoneGuard::NeedsMigration,
    })
}

/// [`upgrade_done_guard_with`] with nothing in the middle.
pub fn upgrade_done_guard(
    conn: &mut Connection,
    ledger: &Path,
) -> Result<Option<Upgraded>, LedgerError> {
    upgrade_done_guard_with(conn, ledger, |_| Ok(()))
}

/// Gives the ledger at `ledger` (open as `conn`, writable) the current `card_done_guard`.
/// `Ok(None)` means it already had it and nothing was written. `after_drop` runs in the
/// transaction right after the old trigger was dropped and before the new one is created, so a
/// test can make the upgrade fail at its most exposed point; an `Err` from it rolls back.
pub fn upgrade_done_guard_with(
    conn: &mut Connection,
    ledger: &Path,
    after_drop: impl FnOnce(&Transaction<'_>) -> Result<(), LedgerError>,
) -> Result<Option<Upgraded>, LedgerError> {
    // The common case reads once and returns: no lock, no write.
    if done_guard_status(conn)? == DoneGuard::Current {
        return Ok(None);
    }
    let tx = conn.transaction_with_behavior(TransactionBehavior::Immediate)?;
    // Another process may have upgraded it while this one waited for the lock.
    if done_guard_status(&tx)? == DoneGuard::Current {
        return Ok(None);
    }
    let (now, stamp): (i64, String) = tx.query_row(
        "SELECT unixepoch(), strftime('%Y%m%dT%H%M%SZ', unixepoch(), 'unixepoch')",
        [],
        |r| Ok((r.get(0)?, r.get(1)?)),
    )?;
    let backup = backup_path(ledger, &stamp)?;
    copy_ledger(ledger, &backup)?;
    let upgraded = (|| -> Result<(), LedgerError> {
        check_backup(&backup)?;
        tx.execute_batch("DROP TRIGGER IF EXISTS card_done_guard;")?;
        after_drop(&tx)?;
        tx.execute_batch(schema::CARD_DONE_GUARD)?;
        let name = backup
            .file_name()
            .map(|n| n.to_string_lossy().into_owned())
            .unwrap_or_default();
        tx.execute(
            "INSERT OR REPLACE INTO meta (key, value) VALUES (?1, ?2)",
            rusqlite::params![DONE_GUARD_META_KEY, format!("{now}|{name}")],
        )?;
        Ok(())
    })();
    match upgraded.and_then(|()| tx.commit().map_err(LedgerError::from)) {
        Ok(()) => Ok(Some(Upgraded { backup })),
        Err(e) => {
            // The transaction rolled back (the `Transaction` drop or a failed commit); the copy
            // this call made is no use to anyone, and it is the only file removed.
            let _ = std::fs::remove_file(&backup);
            Err(e)
        }
    }
}

/// `<ledger>.bak-<stamp>` in the ledger's own folder.
fn backup_path(ledger: &Path, stamp: &str) -> Result<PathBuf, LedgerError> {
    let name = ledger
        .file_name()
        .ok_or_else(|| LedgerError::Io("the ledger path has no file name".into()))?;
    Ok(ledger.with_file_name(format!("{}.bak-{stamp}", name.to_string_lossy())))
}

/// Copies `from` to the new file `to` (never an existing one), private to the owner, and flushes it.
fn copy_ledger(from: &Path, to: &Path) -> Result<(), LedgerError> {
    let io = |what: &str, e: std::io::Error| {
        LedgerError::Io(format!("ledger backup {what} {}: {e}", to.display()))
    };
    let mut options = std::fs::OpenOptions::new();
    options.write(true).create_new(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600).custom_flags(libc::O_NOFOLLOW);
    }
    let mut dest = options.open(to).map_err(|e| io("cannot be created:", e))?;
    let copied = (|| {
        let mut src = std::fs::File::open(from)?;
        std::io::copy(&mut src, &mut dest)?;
        dest.sync_all()
    })();
    if let Err(e) = copied {
        let _ = std::fs::remove_file(to);
        return Err(io("could not be written:", e));
    }
    Ok(())
}

/// The copy opens read-only as a MoAI ledger of this version and passes SQLite's quick check.
fn check_backup(backup: &Path) -> Result<(), LedgerError> {
    let bad = |why: String| LedgerError::Io(format!("ledger backup {}: {why}", backup.display()));
    let conn = Connection::open_with_flags(
        backup,
        OpenFlags::SQLITE_OPEN_READ_ONLY | OpenFlags::SQLITE_OPEN_NO_MUTEX,
    )
    .map_err(|e| bad(e.to_string()))?;
    let check: String = conn
        .query_row("PRAGMA quick_check", [], |r| r.get(0))
        .map_err(|e| bad(e.to_string()))?;
    if check != "ok" {
        return Err(bad(format!("does not read back (quick_check: {check})")));
    }
    let app: i64 = conn
        .query_row("PRAGMA application_id", [], |r| r.get(0))
        .map_err(|e| bad(e.to_string()))?;
    let version: i64 = conn
        .query_row("PRAGMA user_version", [], |r| r.get(0))
        .map_err(|e| bad(e.to_string()))?;
    if app != APPLICATION_ID || version != SCHEMA_VERSION {
        return Err(bad("does not read back as this ledger".into()));
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn a_backup_that_is_not_this_ledger_does_not_read_back() {
        let dir = tempfile::tempdir().unwrap();
        // Not a database at all.
        let garbage = dir.path().join("garbage");
        std::fs::write(&garbage, vec![7u8; 4096]).unwrap();
        assert!(check_backup(&garbage).is_err());
        // A database that is not a MoAI ledger (application_id 0).
        let foreign = dir.path().join("foreign");
        Connection::open(&foreign)
            .unwrap()
            .execute_batch("CREATE TABLE t (a); PRAGMA user_version = 1;")
            .unwrap();
        assert!(check_backup(&foreign).is_err());
        // A MoAI ledger of another schema version.
        let other = dir.path().join("other");
        Connection::open(&other)
            .unwrap()
            .execute_batch(&format!(
                "CREATE TABLE t (a); PRAGMA application_id = {APPLICATION_ID}; PRAGMA user_version = 2;"
            ))
            .unwrap();
        assert!(check_backup(&other).is_err());
        // The same pragmas at this version read back.
        let good = dir.path().join("good");
        Connection::open(&good)
            .unwrap()
            .execute_batch(&format!(
                "CREATE TABLE t (a); PRAGMA application_id = {APPLICATION_ID}; PRAGMA user_version = {SCHEMA_VERSION};"
            ))
            .unwrap();
        assert!(check_backup(&good).is_ok());
    }

    #[test]
    fn a_backup_is_never_written_over_an_existing_file() {
        let dir = tempfile::tempdir().unwrap();
        let from = dir.path().join("from");
        let to = dir.path().join("to");
        std::fs::write(&from, b"ledger bytes").unwrap();
        std::fs::write(&to, b"already here").unwrap();
        assert!(copy_ledger(&from, &to).is_err());
        assert_eq!(std::fs::read(&to).unwrap(), b"already here");
        std::fs::remove_file(&to).unwrap();
        copy_ledger(&from, &to).unwrap();
        assert_eq!(std::fs::read(&to).unwrap(), b"ledger bytes");
    }

    #[test]
    fn only_the_current_version_opens_at_its_edges() {
        assert!(matches!(check_version(0), Err(LedgerError::Unavailable(_))));
        assert!(matches!(
            check_version(-1),
            Err(LedgerError::Unavailable(_))
        ));
        assert_eq!(check_version(1), Ok(()));
        assert_eq!(
            check_version(2),
            Err(LedgerError::SchemaNewer {
                found: 2,
                supported: 1
            })
        );
    }
}
