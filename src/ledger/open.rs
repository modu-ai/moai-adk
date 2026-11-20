//! Registering, opening and writing a ledger (Issue #3068 P1).
//!
//! - [`register`] is the only call that creates a folder or the file.
//! - [`open_read_only`] and [`open_write`] never create anything. An absent ledger is
//!   [`LedgerError::Absent`]; a file that is not this project's ledger is
//!   [`LedgerError::Unavailable`] and is left byte for byte as it was.
//! - Every connection sets `foreign_keys=ON` and `recursive_triggers=ON` (so a `REPLACE` that
//!   removes a row still runs the delete triggers), uses the rollback journal
//!   (`journal_mode=DELETE`; a WAL file is refused from its header before SQLite sees it) and
//!   waits at most [`BUSY_TIMEOUT`] for another writer.
//! - [`WriteLedger::with_write_tx`] is the one way to write: `BEGIN IMMEDIATE`, then the DB clock
//!   (`unixepoch()`), then the caller's statements, then commit; an error rolls back.

use super::{location, migrate, schema, LedgerError, APPLICATION_ID};
use rusqlite::{Connection, OpenFlags, Transaction, TransactionBehavior};
use std::io::Read;
use std::path::Path;
use std::time::Duration;

/// How long a connection waits for another process's write lock before `Busy`.
pub const BUSY_TIMEOUT: Duration = Duration::from_secs(5);

/// The name `register` gives its half-built file: `.creating-<pid>-<n>`. A writer that died
/// leaves one behind; [`super::recover::stale_temp`] reports those (it never removes them).
pub(crate) const TEMP_PREFIX: &str = ".creating-";

/// The first 16 bytes of every SQLite file.
pub(crate) const HEADER_MAGIC: &[u8; 16] = b"SQLite format 3\0";

/// What [`register`] found or made.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Registered {
    /// The folder and the file were made by this call.
    Created,
    /// A valid ledger of this project was already there; nothing was changed.
    AlreadyExists,
}

/// A read-only connection to a ledger. Writes on it fail with [`LedgerError::ReadOnly`].
pub struct ReadLedger {
    conn: Connection,
    project_id: String,
    schema_version: i64,
}

/// A read-write connection to an existing ledger. Change it only through
/// [`WriteLedger::with_write_tx`].
pub struct WriteLedger {
    conn: Connection,
    project_id: String,
}

impl ReadLedger {
    /// The connection, for `SELECT`s.
    pub fn conn(&self) -> &Connection {
        &self.conn
    }
    pub fn project_id(&self) -> &str {
        &self.project_id
    }
    /// The file's `user_version`.
    pub fn schema_version(&self) -> i64 {
        self.schema_version
    }
}

impl WriteLedger {
    /// The connection, for `SELECT`s outside a write transaction.
    pub fn conn(&self) -> &Connection {
        &self.conn
    }
    pub fn project_id(&self) -> &str {
        &self.project_id
    }

    /// Runs `f` in one `BEGIN IMMEDIATE` transaction and commits it when `f` returns `Ok`; an
    /// `Err` (or a panic) rolls everything back. `f` gets the transaction and the DB clock
    /// (`unixepoch()` read right after the lock was won, seconds). Use that value for every
    /// time column of the transaction, never the process clock. When another process holds the
    /// write lock past [`BUSY_TIMEOUT`] the call returns [`LedgerError::Busy`] without running `f`.
    pub fn with_write_tx<T>(
        &mut self,
        f: impl FnOnce(&Transaction<'_>, i64) -> Result<T, LedgerError>,
    ) -> Result<T, LedgerError> {
        let tx = self
            .conn
            .transaction_with_behavior(TransactionBehavior::Immediate)?;
        let now: i64 = tx.query_row("SELECT unixepoch()", [], |row| row.get(0))?;
        let out = f(&tx, now)?;
        tx.commit()?;
        Ok(out)
    }
}

/// The key of an event that has no natural identity: `<kind>:<card_id or 0>:<32 hex digits>`,
/// the random part made by SQLite (`randomblob(16)`), so a recovered ledger never meets an old
/// key (amendment 1, section A-2 (b)). Events with a natural identity (an import, a host close,
/// a journal replay) build their key themselves and a replay of the same key adds nothing.
pub fn new_event_key(
    conn: &Connection,
    kind: &str,
    card_id: Option<i64>,
) -> Result<String, LedgerError> {
    Ok(conn.query_row(
        "SELECT ?1 || ':' || ?2 || ':' || lower(hex(randomblob(16)))",
        rusqlite::params![kind, card_id.unwrap_or(0)],
        |row| row.get(0),
    )?)
}

/// Opens an existing ledger for reading. Never creates a folder or a file, and does not write
/// (no journal, no WAL file): see the module comment for the refusals.
pub fn open_read_only(home: &Path, project_id: &str) -> Result<ReadLedger, LedgerError> {
    open_read_only_within(home, project_id, BUSY_TIMEOUT)
}

/// [`open_read_only`] with the wait for another connection given: a reader that must answer
/// quickly (the status line, card C-65) gives up after `busy` instead of [`BUSY_TIMEOUT`]. The
/// wait applies to every statement of the returned connection.
pub fn open_read_only_within(
    home: &Path,
    project_id: &str,
    busy: Duration,
) -> Result<ReadLedger, LedgerError> {
    let path = location::ledger_path(home, project_id)?;
    super::folders::refuse_links(home, project_id)?;
    let (conn, schema_version) = open_at_timeout(&path, project_id, false, busy)?;
    Ok(ReadLedger {
        conn,
        project_id: project_id.to_string(),
        schema_version,
    })
}

/// Opens an existing ledger for writing. Never creates a folder or a file.
pub fn open_write(home: &Path, project_id: &str) -> Result<WriteLedger, LedgerError> {
    let path = location::ledger_path(home, project_id)?;
    super::folders::refuse_links(home, project_id)?;
    let (conn, _) = open_at(&path, project_id, true)?;
    Ok(WriteLedger {
        conn,
        project_id: project_id.to_string(),
    })
}

/// The explicit registration: makes `<home>/.moai/projects/<project_id>/` and the ledger file
/// in it, and nothing else. The file is built under a temporary name in that folder and linked
/// into place only when complete and checked, so a second `register` (or a crash) never leaves
/// a half-made ledger at the real name. A ledger that is already there is checked and left
/// alone ([`Registered::AlreadyExists`]); a file there that is not this project's ledger is an
/// error, never replaced. Folders it makes are 0700; see [`super::folders`] for the modes it
/// tightens, the symbolic links it refuses and the race that remains.
pub fn register(home: &Path, project_id: &str) -> Result<Registered, LedgerError> {
    Ok(register_with_report(home, project_id)?.0)
}

/// [`register`] that also says which folders it made and which it tightened.
pub fn register_with_report(
    home: &Path,
    project_id: &str,
) -> Result<(Registered, super::folders::FolderReport), LedgerError> {
    static COUNTER: std::sync::atomic::AtomicU64 = std::sync::atomic::AtomicU64::new(0);
    use super::folders;
    let path = location::ledger_path(home, project_id)?;
    let dir = location::project_dir(home, project_id)?;
    check_ready(home, project_id)?;
    let plan = folders::inspect(home, project_id)?;
    if let Some(other) = plan.not_ours.first() {
        return Err(LedgerError::FolderNotOurs(other.display().to_string()));
    }
    match std::fs::symlink_metadata(&path) {
        Ok(_) => {
            let report = folders::tighten_existing(home, project_id)?;
            open_read_only(home, project_id)?;
            return Ok((Registered::AlreadyExists, report));
        }
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
        Err(e) => return Err(LedgerError::Io(e.to_string())),
    }
    let report = folders::create_chain(home, project_id)?;
    let tmp = dir.join(format!(
        "{TEMP_PREFIX}{}-{}",
        std::process::id(),
        COUNTER.fetch_add(1, std::sync::atomic::Ordering::Relaxed)
    ));
    let made = folders::verify_open(home, project_id)
        .and_then(|()| build_new(&tmp, project_id))
        .and_then(|()| {
            folders::verify_open(home, project_id)?;
            match std::fs::hard_link(&tmp, &path) {
                Ok(()) => Ok(Registered::Created),
                // Another register won the race between our check and now: its file stands.
                Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => {
                    open_read_only(home, project_id)?;
                    Ok(Registered::AlreadyExists)
                }
                Err(e) => Err(LedgerError::Io(e.to_string())),
            }
        });
    let _ = std::fs::remove_file(&tmp);
    Ok((made?, report))
}

/// [`check_home_ready`] and no symbolic link at `.moai`, `projects` or `<id>`.
pub(crate) fn check_ready(home: &Path, project_id: &str) -> Result<(), LedgerError> {
    // The link check comes first: a link at `.moai` would otherwise be reported as a HOME that
    // is not ready.
    super::folders::refuse_links(home, project_id)?;
    check_home_ready(home)
}

/// The HOME states this slice writes into: not the Go generation's `~/.moai`, not a stopped
/// move, not a refused one (the verdicts of `state_root::classify_home`). The HOME folder itself
/// must exist: registering never creates a HOME.
pub(crate) fn check_home_ready(home: &Path) -> Result<(), LedgerError> {
    if !home.is_dir() {
        return Err(LedgerError::HomeNotReady(
            "the HOME folder does not exist".into(),
        ));
    }
    let verdict = crate::state_root::classify_home(home);
    if matches!(
        verdict.generation,
        "go" | "mixed" | "interrupted" | "refuse"
    ) {
        return Err(LedgerError::HomeNotReady(format!(
            "~/.moai is `{}`; the ledger waits for the HOME move",
            verdict.generation
        )));
    }
    Ok(())
}

/// Builds a complete, checked ledger at `tmp` (which must not exist).
fn build_new(tmp: &Path, project_id: &str) -> Result<(), LedgerError> {
    build_new_with(tmp, project_id, |_| Ok(()))
}

/// [`build_new`] with `populate` run inside the build transaction (see
/// [`schema::create_v1_with`]); the recovery uses it to fill the new file. Returns what
/// `populate` returned once the finished file opened as this project's ledger.
pub(crate) fn build_new_with<T>(
    tmp: &Path,
    project_id: &str,
    populate: impl FnOnce(&Transaction<'_>) -> Result<T, LedgerError>,
) -> Result<T, LedgerError> {
    let flags = OpenFlags::SQLITE_OPEN_READ_WRITE
        | OpenFlags::SQLITE_OPEN_CREATE
        | OpenFlags::SQLITE_OPEN_NO_MUTEX;
    // The file is made private before SQLite writes a byte into it (an empty file is a valid new
    // database), so it is never readable by others, whatever the umask.
    {
        let mut options = std::fs::OpenOptions::new();
        options.write(true).create_new(true);
        #[cfg(unix)]
        {
            use std::os::unix::fs::OpenOptionsExt;
            options.mode(0o600).custom_flags(libc::O_NOFOLLOW);
        }
        options
            .open(tmp)
            .map_err(|e| LedgerError::Io(format!("{}: {e}", tmp.display())))?;
    }
    let out = {
        let mut conn = Connection::open_with_flags(tmp, flags)?;
        conn.busy_timeout(BUSY_TIMEOUT)?;
        conn.query_row("PRAGMA journal_mode = DELETE", [], |_| Ok(()))?;
        conn.execute_batch("PRAGMA foreign_keys = ON; PRAGMA recursive_triggers = ON;")?;
        schema::create_v1_with(&mut conn, project_id, crate::BUILD_VERSION, populate)?
    };
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        std::fs::set_permissions(tmp, std::fs::Permissions::from_mode(0o600))
            .map_err(|e| LedgerError::Io(e.to_string()))?;
    }
    open_at(tmp, project_id, false)?;
    Ok(out)
}

/// The file is a regular file, a SQLite database, not in WAL mode (header bytes 18 and 19 are
/// the write and read format versions; 2 is WAL). Reads 100 bytes with plain file I/O: SQLite
/// has not touched the file yet.
fn check_header(path: &Path) -> Result<(), LedgerError> {
    let meta = match std::fs::symlink_metadata(path) {
        Ok(m) => m,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Err(LedgerError::Absent),
        Err(e) => return Err(LedgerError::Io(e.to_string())),
    };
    if !meta.file_type().is_file() {
        return Err(LedgerError::Unavailable(
            "the ledger path is a link or not a regular file".into(),
        ));
    }
    let mut head = [0u8; 100];
    let mut file = std::fs::File::open(path).map_err(|e| match e.kind() {
        std::io::ErrorKind::NotFound => LedgerError::Absent,
        _ => LedgerError::Io(format!("the ledger file cannot be opened: {e}")),
    })?;
    match file.read_exact(&mut head) {
        Ok(()) => {}
        // Fewer than 100 bytes: the content is cut short, which is damage.
        Err(e) if e.kind() == std::io::ErrorKind::UnexpectedEof => {
            return Err(LedgerError::Unavailable(
                "the file is not a SQLite database".into(),
            ))
        }
        // Any other read error says nothing about the content (a transient failure).
        Err(e) => {
            return Err(LedgerError::Io(format!(
                "the ledger file cannot be read: {e}"
            )))
        }
    }
    if &head[..16] != HEADER_MAGIC {
        return Err(LedgerError::Unavailable(
            "the file is not a SQLite database".into(),
        ));
    }
    if head[18] == 2 || head[19] == 2 {
        return Err(LedgerError::WalUnexpected);
    }
    Ok(())
}

/// Opens `path` (never creating it), sets the per-connection pragmas, and checks the file is
/// this project's ledger of a version this build can read. Returns the connection and the
/// file's `user_version`.
pub(crate) fn open_at(
    path: &Path,
    project_id: &str,
    writable: bool,
) -> Result<(Connection, i64), LedgerError> {
    open_at_timeout(path, project_id, writable, BUSY_TIMEOUT)
}

/// [`open_at`] with the wait for another connection given (a probe that only classifies a file
/// does not need the full [`BUSY_TIMEOUT`]).
pub(crate) fn open_at_timeout(
    path: &Path,
    project_id: &str,
    writable: bool,
    busy: Duration,
) -> Result<(Connection, i64), LedgerError> {
    check_header(path)?;
    let access = if writable {
        OpenFlags::SQLITE_OPEN_READ_WRITE
    } else {
        OpenFlags::SQLITE_OPEN_READ_ONLY
    };
    // No SQLITE_OPEN_NOFOLLOW: SQLite then refuses a link anywhere in the path, and a HOME
    // behind a link (`/var` on macOS) is normal. The ledger file itself was checked above.
    let conn = Connection::open_with_flags(path, access | OpenFlags::SQLITE_OPEN_NO_MUTEX)?;
    conn.busy_timeout(busy)?;
    conn.execute_batch("PRAGMA foreign_keys = ON; PRAGMA recursive_triggers = ON;")?;
    let version = verify_identity(&conn, project_id, writable).map_err(|e| match e {
        // A read-only connection cannot roll back a journal a crashed writer left behind.
        LedgerError::ReadOnly => LedgerError::Unavailable(
            "needs_recovery: a leftover journal must be rolled back by a writer first".into(),
        ),
        other => other,
    })?;
    Ok((conn, version))
}

fn verify_identity(
    conn: &Connection,
    project_id: &str,
    writable: bool,
) -> Result<i64, LedgerError> {
    let mode: String = if writable {
        conn.query_row("PRAGMA journal_mode = DELETE", [], |r| r.get(0))?
    } else {
        conn.query_row("PRAGMA journal_mode", [], |r| r.get(0))?
    };
    if mode.eq_ignore_ascii_case("wal") {
        return Err(LedgerError::WalUnexpected);
    }
    let application_id: i64 = conn.query_row("PRAGMA application_id", [], |r| r.get(0))?;
    if application_id != APPLICATION_ID {
        return Err(LedgerError::Unavailable(format!(
            "application_id is {application_id:#x}, not a MoAI ledger ({APPLICATION_ID:#x})"
        )));
    }
    let version: i64 = conn.query_row("PRAGMA user_version", [], |r| r.get(0))?;
    migrate::check_version(version)?;
    let found: String = conn
        .query_row("SELECT value FROM meta WHERE key = 'project_id'", [], |r| {
            r.get(0)
        })
        .map_err(meta_lookup_error)?;
    if found != project_id {
        return Err(LedgerError::IdCollision {
            expected: project_id.to_string(),
            found,
        });
    }
    Ok(version)
}

/// What a failed read of `meta.project_id` means: a database that is not one, a corrupt one, a
/// missing table or a missing row is damage (`Unavailable`); a lock is `Busy`; a read-only
/// connection that needs a journal rolled back stays `ReadOnly`; every other error (an I/O
/// failure, any other SQLite error) says nothing about the content and is `Io`.
fn meta_lookup_error(e: rusqlite::Error) -> LedgerError {
    if matches!(e, rusqlite::Error::QueryReturnedNoRows) {
        return LedgerError::Unavailable("the ledger has no project id (no row)".into());
    }
    let text = e.to_string();
    match LedgerError::from(e) {
        busy @ (LedgerError::Busy | LedgerError::ReadOnly) => busy,
        LedgerError::Unavailable(why) => {
            LedgerError::Unavailable(format!("the ledger has no project id ({why})"))
        }
        _ if text.contains("no such table") => {
            LedgerError::Unavailable(format!("the ledger has no project id ({text})"))
        }
        other => LedgerError::Io(format!("the project id could not be read ({other})")),
    }
}
