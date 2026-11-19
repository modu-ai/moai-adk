//! The BOARD 02 observation store (SPEC-20261120-board-observation-store): one file beside the
//! ledger in the project folder, with the ledger folder's rules applied again — owner-only
//! permissions, a refusal when a link stands anywhere on the path. Rows are unique per (source,
//! source event id), so feeding the same observation file twice adds nothing and the same id
//! with different content is kept apart as a named conflict. The kinds are the six fixed
//! `name@1` values and never change; rows older than the 30-day retention are pruned while the
//! ledger stays permanent. A commit advances the cursor and appends its rows in one
//! transaction inside this store only; a failed batch leaves both where they were. The read
//! side opens its own read-only connection and hands out an opaque cursor that is returned
//! verbatim and never recomputed (BOARD 01's contract). This module is declared from
//! `src/lib.rs` in a later slice (that file is reserved by another card meanwhile), so it is
//! written with `moai_cli::` absolute paths only and must not use `crate::`.

use moai_cli::ledger::folders;
use moai_cli::ledger::location;
use moai_cli::ledger::LedgerError;
use rusqlite::{Connection, OpenFlags, OptionalExtension};
use std::fmt;
use std::path::{Path, PathBuf};

/// How long observations live; the ledger is permanent.
pub const RETENTION_DAYS: i64 = 30;

/// The six fixed kinds. A version never changes once fixed.
pub const KINDS: [&str; 6] = [
    "lane-observation@1",
    "pr-state@1",
    "local-check@1",
    "bundle-run@1",
    "gap@1",
    "wake@1",
];

/// One observation as the collector hands it over: where it came from, the event it answers,
/// what it is about, when it was seen, its kind and its content (references and hashes only —
/// never log originals, secrets or dialogue).
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Observation {
    pub source: String,
    pub source_event_id: String,
    pub scope: String,
    pub observed_at: i64,
    pub kind: String,
    pub content: String,
}

/// What one commit did.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct Append {
    pub added: usize,
    pub unchanged: usize,
    /// `(source/source_event_id, reason)` for ids that came back with different content.
    pub conflicts: Vec<(String, String)>,
}

/// One stored row, as the read side returns it.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Row {
    pub seq: i64,
    pub source: String,
    pub source_event_id: String,
    pub scope: String,
    pub observed_at: i64,
    pub kind: String,
    pub content_hash: String,
    pub content: String,
}

/// What can go wrong: a path or file problem (`Unavailable`), a bad input (`InvalidInput`) or
/// a sqlite failure (`Io`). Kinds map onto the envelope words the product already uses.
#[derive(Debug)]
pub enum StoreError {
    Unavailable(String),
    InvalidInput(String),
    Io(String),
}

impl fmt::Display for StoreError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            StoreError::Unavailable(m) | StoreError::InvalidInput(m) | StoreError::Io(m) => {
                write!(f, "{m}")
            }
        }
    }
}

impl std::error::Error for StoreError {}

impl From<rusqlite::Error> for StoreError {
    fn from(e: rusqlite::Error) -> Self {
        StoreError::Io(e.to_string())
    }
}

impl From<LedgerError> for StoreError {
    fn from(e: LedgerError) -> Self {
        StoreError::Unavailable(e.to_string())
    }
}

/// The store's own file inside the project folder, beside `ledger.db`.
pub fn store_path(home: &Path, project_id: &str) -> Result<PathBuf, StoreError> {
    Ok(location::project_dir(home, project_id)?.join("observation.db"))
}

fn connect_write(path: &Path) -> Result<Connection, StoreError> {
    if let Ok(m) = std::fs::symlink_metadata(path) {
        if m.file_type().is_symlink() {
            return Err(StoreError::Unavailable(format!(
                "a link stands where the observation store belongs: {}",
                path.display()
            )));
        }
    }
    let fresh = !path.exists();
    let conn = Connection::open(path)?;
    conn.busy_timeout(std::time::Duration::from_secs(2))?;
    conn.pragma_update(None, "journal_mode", "DELETE")?;
    conn.execute_batch(
        "CREATE TABLE IF NOT EXISTS meta(key TEXT PRIMARY KEY, value TEXT NOT NULL);
         INSERT OR IGNORE INTO meta(key, value) VALUES ('schema_version', '1');
         CREATE TABLE IF NOT EXISTS observation(
             seq INTEGER PRIMARY KEY,
             source TEXT NOT NULL,
             source_event_id TEXT NOT NULL,
             scope TEXT NOT NULL,
             observed_at INTEGER NOT NULL,
             kind TEXT NOT NULL,
             content_hash TEXT NOT NULL,
             content TEXT NOT NULL,
             UNIQUE(source, source_event_id));
         CREATE TABLE IF NOT EXISTS cursor(
             id INTEGER PRIMARY KEY CHECK (id = 1),
             value TEXT NOT NULL);",
    )?;
    if fresh {
        set_mode_0600(path);
    }
    Ok(conn)
}

#[cfg(unix)]
fn set_mode_0600(path: &Path) {
    use std::os::unix::fs::PermissionsExt;
    let _ = std::fs::set_permissions(path, std::fs::Permissions::from_mode(0o600));
}

#[cfg(not(unix))]
fn set_mode_0600(_path: &Path) {}

/// The write side of the observation store. One holder at a time; the collector holds it
/// while it stores what it gathered.
pub struct Store {
    path: PathBuf,
    conn: Connection,
}

impl Store {
    /// Opens (creating if absent) the store in the project folder, applying the ledger folder
    /// rules: links anywhere on the path are refused, folders are owner-only.
    pub fn open(home: &Path, project_id: &str) -> Result<Store, StoreError> {
        folders::refuse_links(home, project_id)?;
        folders::create_chain(home, project_id)?;
        let path = store_path(home, project_id)?;
        let conn = connect_write(&path)?;
        Ok(Store { path, conn })
    }

    /// The store file's path.
    pub fn path(&self) -> &Path {
        &self.path
    }

    /// Appends `rows` and sets the cursor to `cursor` in one transaction: every row is
    /// validated first, so a bad kind rolls the whole batch back and moves nothing. The same
    /// (source, source event id) with the same content is an `unchanged` no-op; with different
    /// content the existing row stays and the pair is named in `conflicts`.
    pub fn commit(&mut self, rows: &[Observation], cursor: &str) -> Result<Append, StoreError> {
        for r in rows {
            if !KINDS.contains(&r.kind.as_str()) {
                return Err(StoreError::InvalidInput(format!(
                    "unknown observation kind {} (fixed kinds: {})",
                    r.kind,
                    KINDS.join(", ")
                )));
            }
        }
        let tx = self.conn.transaction()?;
        let mut out = Append::default();
        for r in rows {
            let hash = moai_cli::digest(r.content.as_bytes());
            let existing: Option<String> = tx
                .query_row(
                    "SELECT content_hash FROM observation
                     WHERE source = ?1 AND source_event_id = ?2",
                    (&r.source, &r.source_event_id),
                    |row| row.get(0),
                )
                .optional()?;
            match existing {
                None => {
                    tx.execute(
                        "INSERT INTO observation
                         (source, source_event_id, scope, observed_at, kind, content_hash, content)
                         VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7)",
                        (
                            &r.source,
                            &r.source_event_id,
                            &r.scope,
                            r.observed_at,
                            &r.kind,
                            &hash,
                            &r.content,
                        ),
                    )?;
                    out.added += 1;
                }
                Some(h) if h == hash => out.unchanged += 1,
                Some(_) => out.conflicts.push((
                    format!("{}/{}", r.source, r.source_event_id),
                    "the same id came back with different content; the first row stays".to_string(),
                )),
            }
        }
        tx.execute(
            "INSERT INTO cursor(id, value) VALUES (1, ?1)
             ON CONFLICT(id) DO UPDATE SET value = excluded.value",
            (cursor,),
        )?;
        tx.commit()?;
        Ok(out)
    }

    /// The stored collection cursor, verbatim.
    pub fn cursor(&self) -> Option<String> {
        self.conn
            .query_row("SELECT value FROM cursor WHERE id = 1", [], |row| {
                row.get(0)
            })
            .optional()
            .unwrap_or(None)
    }

    /// Removes the collection cursor (the rows stay); collecting again must give the same
    /// result.
    pub fn clear_cursor(&mut self) -> Result<(), StoreError> {
        self.conn.execute("DELETE FROM cursor WHERE id = 1", [])?;
        Ok(())
    }

    /// Deletes observations older than the retention and returns how many went. Never touches
    /// the ledger file.
    pub fn prune(&mut self, now_unix: i64) -> Result<usize, StoreError> {
        let cutoff = now_unix - RETENTION_DAYS * 24 * 3600;
        Ok(self
            .conn
            .execute("DELETE FROM observation WHERE observed_at < ?1", (cutoff,))?)
    }

    /// How many observations the store holds.
    pub fn count(&self) -> usize {
        self.conn
            .query_row("SELECT COUNT(*) FROM observation", [], |row| {
                row.get::<_, i64>(0)
            })
            .unwrap_or(0) as usize
    }
}

/// The read side: its own read-only connection, so reading can never write.
pub struct ReadStore {
    conn: Connection,
}

/// Opens the store read-only for `project_id`; a missing store is an error (the timeline
/// decides what to do without one).
pub fn open_read(home: &Path, project_id: &str) -> Result<ReadStore, StoreError> {
    let path = store_path(home, project_id)?;
    let conn = Connection::open_with_flags(
        &path,
        OpenFlags::SQLITE_OPEN_READ_ONLY | OpenFlags::SQLITE_OPEN_NO_MUTEX,
    )?;
    conn.busy_timeout(std::time::Duration::from_millis(100))?;
    Ok(ReadStore { conn })
}

/// The read cursor's shape: `o-<seq>`. Opaque to callers — handed back verbatim, never
/// recomputed; a foreign string is refused.
fn decode_cursor(c: &str) -> Result<i64, StoreError> {
    c.strip_prefix("o-")
        .and_then(|n| n.parse().ok())
        .filter(|n| *n >= 0)
        .ok_or_else(|| {
            StoreError::InvalidInput(
                "not an observation-store cursor (hand back the one you were given)".to_string(),
            )
        })
}

impl ReadStore {
    /// Up to `limit` rows after `after` (None reads from the start), oldest first, with the
    /// next cursor when more rows follow. Never waits and never writes.
    pub fn read_after(
        &self,
        after: Option<&str>,
        limit: usize,
    ) -> Result<(Vec<Row>, Option<String>), StoreError> {
        let after_seq = match after {
            None => 0,
            Some(c) => decode_cursor(c)?,
        };
        let mut stmt = self.conn.prepare(
            "SELECT seq, source, source_event_id, scope, observed_at, kind, content_hash, content
             FROM observation WHERE seq > ?1 ORDER BY seq LIMIT ?2",
        )?;
        let rows = stmt
            .query_map((after_seq, limit as i64), |r| {
                Ok(Row {
                    seq: r.get(0)?,
                    source: r.get(1)?,
                    source_event_id: r.get(2)?,
                    scope: r.get(3)?,
                    observed_at: r.get(4)?,
                    kind: r.get(5)?,
                    content_hash: r.get(6)?,
                    content: r.get(7)?,
                })
            })?
            .collect::<Result<Vec<_>, _>>()?;
        let next = if rows.len() == limit {
            rows.last().map(|r| format!("o-{}", r.seq))
        } else {
            None
        };
        Ok((rows, next))
    }
}
