//! Schema version rules of the ledger (Issue #3068 P1) and the one repair of an existing file
//! that exists so far (SPEC-20261121-ledger-card-landed-migration).
//!
//! - [`check_version`] decides whether a file's `user_version` may be opened. Version 1 is still
//!   the only version.
//! - [`upgrade_done_guard`] gives a version 1 ledger made before the M-Flow `card.landed`
//!   exception the current `card_done_guard` trigger. `open_write` runs it on every writable
//!   open; a ledger that already has the exception is not touched.
//!
//! The upgrade runs in one `BEGIN IMMEDIATE` transaction, and the lock that statement takes is
//! held until the commit. Inside it, after the trigger was read again under the lock, SQLite
//! itself makes the backup: a second SQLite connection runs `VACUUM INTO` a new file
//! `ledger.db.bak-<UTC>` beside the ledger (the reserved lock keeps every other writer out and
//! nothing has been written yet, so the copy is a consistent snapshot), and the copy is read
//! back; only then is the trigger replaced and a `meta` row written. Only the process holding the
//! lock makes a backup, so concurrent opens leave one backup and one upgrade.
//!
//! **Never open the ledger file with `std::fs` while a connection holds a lock on it.** POSIX
//! record locks belong to the process, not to the file descriptor: closing any descriptor of the
//! file drops every lock the process holds on it, which would let another process write in the
//! middle of the transaction (https://www.sqlite.org/howtocorrupt.html, section 2.2, read
//! 2026-11-21). SQLite's own connections are safe because they share one lock state per file
//! inside a process and keep the descriptor open until the last lock goes. The backup file
//! itself is a different file, so `std::fs` may create and flush it.
//!
//! Any error rolls the transaction back. The backup this call made is removed only after the
//! trigger is read again and shows the rollback; if the commit may have happened, or the file
//! cannot be read to tell, the backup stays. No row of any data table is touched.
//! `PRAGMA user_version` stays 1 (SPEC decision D-1), so an older build still opens the file.

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

/// The points of an upgrade at which a test may look or fail; each one is inside the write
/// transaction, so the lock must be held at every one of them.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Stage {
    /// The lock is taken and the trigger was read again; nothing is written or copied yet.
    Locked,
    /// The backup exists and reads back; the trigger is still the old one.
    BackedUp,
    /// The old trigger is dropped and the new one is not created yet.
    GuardDropped,
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
    upgrade_done_guard_with(conn, ledger, |_, _| Ok(()))
}

/// Gives the ledger at `ledger` (open as `conn`, writable) the current `card_done_guard`.
/// `Ok(None)` means it already had it and nothing was written. `hook` runs at each [`Stage`]
/// inside the transaction, so a test can look at the lock or make the upgrade fail at its most
/// exposed point; an `Err` from it rolls back.
pub fn upgrade_done_guard_with(
    conn: &mut Connection,
    ledger: &Path,
    mut hook: impl FnMut(Stage, &Transaction<'_>) -> Result<(), LedgerError>,
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
    let mut made = false;
    let work = (|| -> Result<(), LedgerError> {
        hook(Stage::Locked, &tx)?;
        // A refused name (one that exists) is an error before anything is made.
        make_backup(ledger, &backup)?;
        made = true;
        check_backup(&backup)?;
        hook(Stage::BackedUp, &tx)?;
        tx.execute_batch("DROP TRIGGER IF EXISTS card_done_guard;")?;
        hook(Stage::GuardDropped, &tx)?;
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
    let outcome = match work {
        Ok(()) => tx.commit().map_err(LedgerError::from),
        Err(e) => {
            drop(tx); // rolls back
            Err(e)
        }
    };
    match outcome {
        Ok(()) => Ok(Some(Upgraded { backup })),
        Err(e) => {
            if !conn.is_autocommit() {
                let _ = conn.execute_batch("ROLLBACK");
            }
            if !made {
                return Err(e);
            }
            match settle_failed(conn, &backup) {
                Settled::RolledBack | Settled::Unknown => Err(e),
                // The failed call after all committed: the ledger has the trigger and the
                // backup is its record.
                Settled::Committed => Ok(Some(Upgraded { backup })),
            }
        }
    }
}

/// What the file shows after a failed upgrade.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum Settled {
    /// The old trigger is still there: the backup this call made is removed.
    RolledBack,
    /// The new trigger is there: the commit happened, the backup is kept.
    Committed,
    /// The file could not be read to tell: the backup is kept.
    Unknown,
}

/// Reads the trigger again and removes `backup` only when it shows the rollback. A backup that
/// may belong to a committed upgrade is never deleted.
fn settle_failed(conn: &Connection, backup: &Path) -> Settled {
    match done_guard_status(conn) {
        Ok(DoneGuard::NeedsMigration) => {
            remove_backup_files(backup);
            Settled::RolledBack
        }
        Ok(DoneGuard::Current) => Settled::Committed,
        Err(_) => Settled::Unknown,
    }
}

/// Removes a backup this call made and could not use: the file and the `<backup>-journal` a
/// `VACUUM INTO` that stopped half way leaves beside it. Only these two names are touched.
fn remove_backup_files(backup: &Path) {
    let _ = std::fs::remove_file(backup);
    let mut journal = backup.as_os_str().to_owned();
    journal.push("-journal");
    let _ = std::fs::remove_file(PathBuf::from(journal));
}

/// `<ledger>.bak-<stamp>` in the ledger's own folder.
fn backup_path(ledger: &Path, stamp: &str) -> Result<PathBuf, LedgerError> {
    let name = ledger
        .file_name()
        .ok_or_else(|| LedgerError::Io("the ledger path has no file name".into()))?;
    Ok(ledger.with_file_name(format!("{}.bak-{stamp}", name.to_string_lossy())))
}

/// What to do about a backup that cannot be made. The ledger itself was not changed.
const BACKUP_HELP: &str = "free some disk space, make the ledger's folder writable for you, or check its path, then open the ledger again; the ledger itself was not changed";

/// Makes the new file `to` (never an existing one, private to the owner) hold a consistent copy
/// of the ledger `from`, through SQLite: a second SQLite connection runs `VACUUM INTO`. The
/// ledger file is never opened with `std::fs` (see the module doc); `to` is another file and is
/// created and flushed with it. A file this call created and could not fill is removed.
fn make_backup(from: &Path, to: &Path) -> Result<(), LedgerError> {
    let io = |what: &str, e: &dyn std::fmt::Display| {
        LedgerError::Io(format!(
            "ledger backup {} {what} {e}; {BACKUP_HELP}",
            to.display()
        ))
    };
    let mut options = std::fs::OpenOptions::new();
    options.write(true).create_new(true);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600).custom_flags(libc::O_NOFOLLOW);
    }
    // `VACUUM INTO` accepts a file that exists and is empty, so the name is taken first.
    drop(options.open(to).map_err(|e| {
        let why = if e.kind() == std::io::ErrorKind::AlreadyExists {
            "cannot be created, a file of that name exists (move it away or wait one second):"
        } else {
            "cannot be created:"
        };
        io(why, &e)
    })?);
    let copied = (|| -> Result<(), String> {
        let to_text = to.to_str().ok_or("the path is not UTF-8")?;
        let src = Connection::open_with_flags(
            from,
            OpenFlags::SQLITE_OPEN_READ_ONLY | OpenFlags::SQLITE_OPEN_NO_MUTEX,
        )
        .map_err(|e| e.to_string())?;
        src.busy_timeout(super::open::BUSY_TIMEOUT)
            .map_err(|e| e.to_string())?;
        src.execute("VACUUM INTO ?1", [to_text])
            .map_err(|e| e.to_string())?;
        drop(src);
        std::fs::File::open(to)
            .and_then(|f| f.sync_all())
            .map_err(|e| e.to_string())
    })();
    if let Err(e) = copied {
        remove_backup_files(to);
        return Err(io("could not be written:", &e));
    }
    Ok(())
}

/// The copy opens read-only as a MoAI ledger of this version and passes SQLite's quick check.
fn check_backup(backup: &Path) -> Result<(), LedgerError> {
    let bad = |why: String| {
        LedgerError::Io(format!(
            "ledger backup {}: {why}; {BACKUP_HELP}",
            backup.display()
        ))
    };
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

    /// A small database with this build's identity pragmas and one row.
    fn small_ledger(path: &Path) {
        Connection::open(path)
            .unwrap()
            .execute_batch(&format!(
                "CREATE TABLE t (a); INSERT INTO t VALUES (42); \
                 PRAGMA application_id = {APPLICATION_ID}; PRAGMA user_version = {SCHEMA_VERSION};"
            ))
            .unwrap();
    }

    #[test]
    fn a_backup_is_never_made_over_an_existing_file_and_reads_back_private() {
        let dir = tempfile::tempdir().unwrap();
        let from = dir.path().join("from");
        let to = dir.path().join("to");
        small_ledger(&from);
        std::fs::write(&to, b"already here").unwrap();
        let err = make_backup(&from, &to).unwrap_err().to_string();
        assert!(err.contains("exists"), "{err}");
        assert!(
            err.contains("free some disk space"),
            "the refusal says what to do: {err}"
        );
        assert_eq!(std::fs::read(&to).unwrap(), b"already here");
        std::fs::remove_file(&to).unwrap();
        make_backup(&from, &to).unwrap();
        check_backup(&to).unwrap();
        let n: i64 = Connection::open(&to)
            .unwrap()
            .query_row("SELECT a FROM t", [], |r| r.get(0))
            .unwrap();
        assert_eq!(n, 42);
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            let mode = std::fs::metadata(&to).unwrap().permissions().mode() & 0o777;
            assert_eq!(mode, 0o600);
        }
    }

    #[test]
    fn a_backup_of_something_that_is_not_a_database_is_removed() {
        let dir = tempfile::tempdir().unwrap();
        let from = dir.path().join("from");
        let to = dir.path().join("to");
        std::fs::write(&from, vec![7u8; 4096]).unwrap();
        assert!(make_backup(&from, &to).is_err());
        assert!(
            !to.exists(),
            "a file this call made and could not fill is removed"
        );
    }

    /// Names of the files in `dir`, sorted.
    fn files_in(dir: &Path) -> Vec<String> {
        let mut names: Vec<String> = std::fs::read_dir(dir)
            .unwrap()
            .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
            .collect();
        names.sort();
        names
    }

    #[test]
    fn a_failed_backup_leaves_no_file_beside_the_ledger() {
        // Source is not a database.
        let dir = tempfile::tempdir().unwrap();
        let from = dir.path().join("from");
        std::fs::write(&from, vec![7u8; 4096]).unwrap();
        assert!(make_backup(&from, &dir.path().join("to")).is_err());
        assert_eq!(files_in(dir.path()), ["from"], "not a database");

        // Source is a database that another connection holds exclusively: the backup waits
        // for the busy timeout and gives up.
        let dir = tempfile::tempdir().unwrap();
        let from = dir.path().join("from");
        small_ledger(&from);
        let held = Connection::open(&from).unwrap();
        held.execute_batch("PRAGMA locking_mode = EXCLUSIVE; BEGIN EXCLUSIVE;")
            .unwrap();
        assert!(make_backup(&from, &dir.path().join("to")).is_err());
        drop(held);
        assert_eq!(files_in(dir.path()), ["from"], "source held exclusively");
    }

    #[test]
    fn a_half_made_backup_is_removed_with_its_journal_and_nothing_else() {
        let dir = tempfile::tempdir().unwrap();
        let backup = dir.path().join("ledger.db.bak-20261121T000000Z");
        let journal = dir.path().join("ledger.db.bak-20261121T000000Z-journal");
        let neighbour = dir.path().join("ledger.db.bak-20261121T000001Z");
        let ledger = dir.path().join("ledger.db");
        for f in [&backup, &journal, &neighbour, &ledger] {
            std::fs::write(f, b"x").unwrap();
        }
        remove_backup_files(&backup);
        assert_eq!(
            files_in(dir.path()),
            ["ledger.db", "ledger.db.bak-20261121T000001Z"],
            "only the backup and its journal go"
        );
        // Nothing there: no error, nothing changes.
        remove_backup_files(&backup);
        assert_eq!(files_in(dir.path()).len(), 2);
    }

    #[test]
    fn a_rolled_back_upgrade_removes_the_journal_of_its_backup_too() {
        let dir = tempfile::tempdir().unwrap();
        let backup = dir.path().join("backup");
        let journal = dir.path().join("backup-journal");
        let old = dir.path().join("old.db");
        let c = Connection::open(&old).unwrap();
        c.execute_batch(
            "CREATE TABLE card (state); \
             CREATE TRIGGER card_done_guard BEFORE UPDATE OF state ON card BEGIN SELECT 1; END;",
        )
        .unwrap();
        std::fs::write(&backup, b"b").unwrap();
        std::fs::write(&journal, b"j").unwrap();
        assert_eq!(settle_failed(&c, &backup), Settled::RolledBack);
        assert!(!backup.exists() && !journal.exists());
    }

    #[test]
    fn a_failed_upgrade_removes_its_backup_only_when_the_rollback_is_shown() {
        let dir = tempfile::tempdir().unwrap();
        let backup = dir.path().join("backup");
        // Old trigger still there: rolled back, backup removed.
        let old = dir.path().join("old.db");
        let c = Connection::open(&old).unwrap();
        c.execute_batch(
            "CREATE TABLE card (state); \
             CREATE TRIGGER card_done_guard BEFORE UPDATE OF state ON card BEGIN SELECT 1; END;",
        )
        .unwrap();
        std::fs::write(&backup, b"b").unwrap();
        assert_eq!(settle_failed(&c, &backup), Settled::RolledBack);
        assert!(!backup.exists());
        // New trigger there: the commit happened, the backup stays.
        let new = dir.path().join("new.db");
        let c = Connection::open(&new).unwrap();
        c.execute_batch(
            "CREATE TABLE card (state); \
             CREATE TRIGGER card_done_guard BEFORE UPDATE OF state ON card \
             WHEN 'card.landed' = 'card.landed' BEGIN SELECT 1; END;",
        )
        .unwrap();
        std::fs::write(&backup, b"b").unwrap();
        assert_eq!(settle_failed(&c, &backup), Settled::Committed);
        assert!(backup.exists());
        // A file that cannot be read to tell: the backup stays.
        let junk = dir.path().join("junk.db");
        std::fs::write(&junk, vec![7u8; 4096]).unwrap();
        let c = Connection::open(&junk).unwrap();
        assert_eq!(settle_failed(&c, &backup), Settled::Unknown);
        assert!(backup.exists(), "never deleted on a guess");
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
