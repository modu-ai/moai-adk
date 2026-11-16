//! The local project ledger (Issue #3068 P1; schema contract comment 6896940106 and its
//! amendment 1, comment 6897392517).
//!
//! One SQLite file per project, `<HOME>/.moai/projects/<project_id>/ledger.db`. The module is
//! the only place that builds that path ([`location::ledger_path`]) and the only place that
//! holds SQL: the card commands (#3070) and the host connector (#3071) call the functions here.
//!
//! - [`open::register`] is the one call that creates folders and the file.
//! - [`open::open_read_only`] never creates anything and never writes (no journal, no WAL).
//! - [`open::open_write`] opens an existing ledger; [`open::WriteLedger::with_write_tx`] is the
//!   one way to change it (`BEGIN IMMEDIATE`, a busy timeout, the DB clock).
//! - A file that is not this project's ledger is `ledger_unavailable` and is never replaced.
//!
//! Nothing here calls a model, the network or `git`.

pub mod command;
pub mod folders;
pub mod import;
pub mod location;
pub mod migrate;
pub mod open;
pub mod recover;
pub mod reservation;
pub mod schema;
pub mod summary;

pub use location::ledger_path;
pub use open::{
    new_event_key, open_read_only, open_write, register, ReadLedger, Registered, WriteLedger,
};

/// `PRAGMA application_id` of every ledger: the ASCII bytes `MOAI`.
pub const APPLICATION_ID: i64 = 0x4D4F_4149;

/// `PRAGMA user_version` this build writes and reads.
pub const SCHEMA_VERSION: i64 = 1;

/// Why a ledger call did not succeed. [`LedgerError::code`] is the stable word callers match on;
/// `Display` is the one human line (its prefix keeps `contract::classify` honest: `blocked` for
/// the refusals that need a person, `not_found` for an absent ledger, `invalid_input` for a
/// rejected value).
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum LedgerError {
    /// No ledger file exists for this project. Callers treat it as "no cards", not as a failure.
    Absent,
    /// The file is not this project's usable ledger: not a database, corrupt, another
    /// `application_id`, a version below 1, a link or a folder. Nothing is replaced or repaired.
    Unavailable(String),
    /// The file is a MoAI ledger of another project (two projects share a `project_id`).
    IdCollision { expected: String, found: String },
    /// The file is in WAL mode (header bytes 18-19); this build never opens it.
    WalUnexpected,
    /// The file was written by a newer schema than this build knows.
    SchemaNewer { found: i64, supported: i64 },
    /// Another process holds the write lock past the busy timeout. Sending the call again may
    /// succeed.
    Busy,
    /// A write was refused by a `CHECK`, `UNIQUE`, foreign key or trigger of the schema.
    Constraint(String),
    /// A write was attempted on a read-only connection.
    ReadOnly,
    /// `project_id` is not a portable id (letters, digits, `-`, `_`, at most 80).
    InvalidProjectId,
    /// The HOME is in a state this slice does not write into (the Go generation's, or a stopped
    /// move), or it is not an existing folder.
    HomeNotReady(String),
    /// `<home>/.moai`, `<home>/.moai/projects` or `<home>/.moai/projects/<id>` is a symbolic link
    /// (also a dangling one) or was swapped for one while a step ran. Nothing is read, written or
    /// created through it. The text names the component.
    PathLink(String),
    /// `projects` or `projects/<id>` exists but another user owns it: this build neither
    /// changes its mode nor registers into it.
    FolderNotOurs(String),
    /// A file system call failed.
    Io(String),
    /// Any other SQLite error.
    Sql(String),
}

impl LedgerError {
    /// The stable machine word.
    pub fn code(&self) -> &'static str {
        match self {
            LedgerError::Absent => "ledger_absent",
            LedgerError::Unavailable(_) => "ledger_unavailable",
            LedgerError::IdCollision { .. } => "id_collision",
            LedgerError::WalUnexpected => "wal_unexpected",
            LedgerError::SchemaNewer { .. } => "schema_newer",
            LedgerError::Busy => "busy",
            LedgerError::Constraint(_) => "constraint",
            LedgerError::ReadOnly => "read_only",
            LedgerError::InvalidProjectId => "invalid_project_id",
            LedgerError::HomeNotReady(_) => "home_not_ready",
            LedgerError::PathLink(_) => "path_link",
            LedgerError::FolderNotOurs(_) => "folder_not_ours",
            LedgerError::Io(_) => "io",
            LedgerError::Sql(_) => "sql",
        }
    }
}

impl std::fmt::Display for LedgerError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            LedgerError::Absent => {
                f.write_str("ledger not found: no ledger is registered for this project")
            }
            LedgerError::Unavailable(why) => {
                write!(f, "ledger blocked: ledger_unavailable: {why}")
            }
            LedgerError::IdCollision { expected, found } => write!(
                f,
                "ledger blocked: id_collision: the file belongs to project {found:?}, not {expected:?}"
            ),
            LedgerError::WalUnexpected => f.write_str(
                "ledger blocked: wal_unexpected: the file is in WAL mode; this build opens only rollback-journal ledgers",
            ),
            LedgerError::SchemaNewer { found, supported } => write!(
                f,
                "ledger blocked: schema_newer: the file is schema version {found}, this build supports {supported}"
            ),
            LedgerError::Busy => f.write_str(
                "ledger busy: another process holds the ledger past the busy timeout (unavailable for now); send the call again",
            ),
            LedgerError::Constraint(why) => write!(f, "ledger write refused: {why}"),
            LedgerError::ReadOnly => {
                f.write_str("ledger write refused: the connection is read-only")
            }
            LedgerError::InvalidProjectId => {
                f.write_str("invalid project id: use letters, digits, `-` and `_` (at most 80)")
            }
            LedgerError::HomeNotReady(why) => write!(f, "ledger blocked: home_not_ready: {why}"),
            LedgerError::PathLink(why) => write!(
                f,
                "ledger blocked: path_link: {why} is a symbolic link; the ledger never follows one"
            ),
            LedgerError::FolderNotOurs(why) => write!(
                f,
                "ledger blocked: folder_not_ours: {why} belongs to another user; nothing was changed"
            ),
            LedgerError::Io(why) => write!(f, "ledger file write failed: {why}"),
            LedgerError::Sql(why) => write!(f, "ledger sql error: {why}"),
        }
    }
}

impl std::error::Error for LedgerError {}

impl From<rusqlite::Error> for LedgerError {
    fn from(e: rusqlite::Error) -> Self {
        use rusqlite::ErrorCode as C;
        let text = e.to_string();
        match e.sqlite_error_code() {
            Some(C::NotADatabase) | Some(C::DatabaseCorrupt) => {
                LedgerError::Unavailable(format!("the file is not a readable database ({text})"))
            }
            Some(C::DatabaseBusy) | Some(C::DatabaseLocked) => LedgerError::Busy,
            Some(C::ConstraintViolation) => LedgerError::Constraint(text),
            Some(C::ReadOnly) => LedgerError::ReadOnly,
            _ => LedgerError::Sql(text),
        }
    }
}

impl From<LedgerError> for String {
    fn from(e: LedgerError) -> String {
        e.to_string()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn every_error_has_a_distinct_code_and_a_kind_the_envelope_can_use() {
        use crate::contract::{classify, Kind};
        let cases: Vec<(LedgerError, &str, Kind)> = vec![
            (LedgerError::Absent, "ledger_absent", Kind::NotFound),
            (
                LedgerError::Unavailable("x".into()),
                "ledger_unavailable",
                Kind::Blocked,
            ),
            (
                LedgerError::IdCollision {
                    expected: "a".into(),
                    found: "b".into(),
                },
                "id_collision",
                Kind::Blocked,
            ),
            (LedgerError::WalUnexpected, "wal_unexpected", Kind::Blocked),
            (
                LedgerError::SchemaNewer {
                    found: 2,
                    supported: 1,
                },
                "schema_newer",
                Kind::Blocked,
            ),
            (LedgerError::Busy, "busy", Kind::Io),
            (
                LedgerError::Constraint("x".into()),
                "constraint",
                Kind::InvalidInput,
            ),
            (LedgerError::ReadOnly, "read_only", Kind::InvalidInput),
            (
                LedgerError::InvalidProjectId,
                "invalid_project_id",
                Kind::InvalidInput,
            ),
            (
                LedgerError::HomeNotReady("x".into()),
                "home_not_ready",
                Kind::Blocked,
            ),
            (
                LedgerError::PathLink("x".into()),
                "path_link",
                Kind::Blocked,
            ),
            (
                LedgerError::FolderNotOurs("x".into()),
                "folder_not_ours",
                Kind::Blocked,
            ),
            (LedgerError::Io("x".into()), "io", Kind::Io),
        ];
        let mut seen = std::collections::HashSet::new();
        for (error, code, kind) in cases {
            assert_eq!(error.code(), code);
            assert!(seen.insert(code), "{code} is used twice");
            assert_eq!(classify(&error.to_string()), kind, "{error}");
        }
    }
}
