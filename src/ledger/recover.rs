//! Recovering a damaged ledger (Issue #3068 P6; AC7 and a reinforcement of AC6).
//!
//! A ledger that opens as `ledger_unavailable` stops new dispatch (P1). Recovery builds a new,
//! complete ledger from whatever the best readable copy (the damaged file itself, the
//! `.bak` copy, or a ledger-shaped file the operator names) still holds, and brings back only the
//! reservations that can be confirmed. Everything else is held for review. Nothing is deleted
//! and nothing is overwritten.
//!
//! - [`plan`] reads the source with tolerance and writes nothing. Each of the five tables
//!   (`card`, `card_ac`, `card_note`, `card_link`, `event`) and the two reservation tables is
//!   read in primary-key order; a read error stops that table at its last good row and the plan
//!   says how many rows were lost. The result carries a digest over everything the apply will
//!   do.
//! - [`apply`] (under one exclusive per-project lock, [`lock_recovery`]) recomputes the plan, requires the same digest, builds the new ledger at
//!   `<ledger>.recovering` with the builder `register` uses, copies the rows with their explicit
//!   ids, writes one `recover` event, checks the new file, then sets the damaged file aside
//!   (a hard link under `<ledger>.damaged-<UTC>`, so the real name never points at nothing) and
//!   renames the new file over the real name. A kill at any point leaves the old state or the
//!   new state; running the apply again converges.
//! - The decision rule for reservations ([`decide`]) is conservative. A source reservation that
//!   is `held`, `expired_candidate` or `held_for_review` is restored as `held` only when the
//!   journal says the holder's claim is live, the worktree exists on the reservation's branch
//!   and the remote observation agrees; every other row becomes `held_for_review`, which still
//!   occupies the card's holder slot, so no new reservation can take the card.
//!
//! This module never starts a process and never reads the network or `git`: the command layer
//! (`ledger_recover`) gathers the facts and passes them in as plain [`Observations`].

use super::{location, migrate, open, schema, LedgerError, APPLICATION_ID};
use rusqlite::types::Value;
use rusqlite::{Connection, OpenFlags, Transaction};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value as Json};
use std::collections::{BTreeMap, BTreeSet};
use std::path::{Path, PathBuf};

/// Appended to the ledger's path: the file the new ledger is built in.
const RECOVERING: &str = ".recovering";
/// Appended to the ledger's path, then `-<UTC suffix>`: where the damaged file is kept.
const DAMAGED: &str = ".damaged-";
/// The most rejected rows a plan lists by name (the count is always complete).
pub const MAX_LISTED_REJECTED: usize = 200;
/// The product of the decision rule, as it is written to `reservation.status`.
pub const RESTORE_HELD: &str = "restore_held";
pub const HELD_FOR_REVIEW: &str = "held_for_review";

// Reason words of a `held_for_review` decision. They are stable: callers and the recover event
// match on them.
pub const REASON_HOLDER_EMPTY: &str = "holder_empty";
pub const REASON_UNREADABLE_ROW: &str = "unreadable_row";
pub const REASON_GENERATION_BELOW_CARD: &str = "generation_below_card";
pub const REASON_JOURNAL_MISSING: &str = "journal_missing";
pub const REASON_JOURNAL_NOT_LIVE: &str = "journal_not_live";
pub const REASON_WORKTREE_NOT_RECORDED: &str = "worktree_not_recorded";
pub const REASON_WORKTREE_NOT_OBSERVED: &str = "worktree_not_observed";
pub const REASON_WORKTREE_MISSING: &str = "worktree_missing";
pub const REASON_BRANCH_MISMATCH: &str = "branch_mismatch";
pub const REASON_REMOTE_ABSENT: &str = "remote_absent";
pub const REASON_REMOTE_MISMATCH: &str = "remote_mismatch";

/// What the command layer observed outside the ledger. Plain data: the library never calls the
/// network, `git` or the file system of the worktrees.
#[derive(Debug, Clone, Default, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct Observations {
    /// The claim journal's last record per claim (`holder` = the owner id, `generation` = the
    /// epoch), with whether its lease is live.
    #[serde(default)]
    pub journal: Vec<JournalFact>,
    /// The worktrees named by reservations, as found on this machine.
    #[serde(default)]
    pub worktrees: Vec<WorktreeFact>,
    /// What the remote shows for a branch (from an operator-provided observation file).
    #[serde(default)]
    pub remotes: Vec<RemoteFact>,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct JournalFact {
    pub holder: String,
    pub generation: i64,
    pub live: bool,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct WorktreeFact {
    /// The path exactly as the reservation spells it.
    pub path: String,
    pub exists: bool,
    /// The branch checked out there; `None` for a detached HEAD or an unreadable one.
    #[serde(default)]
    pub branch: Option<String>,
    /// Informational: no uncommitted change. It does not decide anything.
    #[serde(default)]
    pub clean: bool,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct RemoteFact {
    pub branch: String,
    /// When given it must equal the reservation's card.
    #[serde(default)]
    pub card_id: Option<i64>,
    /// When given it must equal the reservation's holder.
    #[serde(default)]
    pub holder: Option<String>,
    /// When given it must equal the reservation's generation.
    #[serde(default)]
    pub generation: Option<i64>,
    /// Whether the branch exists on the remote.
    pub exists: bool,
}

/// The facts of one reservation row that the decision rule reads.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ReservationRow {
    pub reservation_id: i64,
    pub card_id: i64,
    pub generation: i64,
    pub holder: String,
    pub status: String,
    pub worktree: Option<String>,
    pub branch: Option<String>,
}

/// What the plan does with one active source reservation.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct ReservationDecision {
    pub reservation_id: i64,
    pub card_id: i64,
    pub generation: i64,
    pub holder: String,
    pub source_status: String,
    /// [`RESTORE_HELD`] or [`HELD_FOR_REVIEW`].
    pub decision: String,
    /// Empty exactly when the decision is [`RESTORE_HELD`].
    pub reasons: Vec<String>,
    /// `true` when the source row could not be read as a reservation (a column holds a value of
    /// another kind than the schema declares). The other fields then hold what could be read
    /// (`0` and the empty text for the rest) and the only reason is [`REASON_UNREADABLE_ROW`].
    pub unreadable: bool,
}

/// The decision rule. `card_generation` is the card's restored generation (the larger of the
/// source card's and the highest reservation generation of that card), so a generation below it
/// has been superseded. `restore_held` only when every condition holds; each failed condition is
/// named, in this fixed order.
pub fn decide(
    row: &ReservationRow,
    card_generation: i64,
    obs: &Observations,
) -> ReservationDecision {
    let mut reasons: Vec<String> = Vec::new();
    if row.holder.trim().is_empty() {
        reasons.push(REASON_HOLDER_EMPTY.into());
    }
    if row.generation < card_generation {
        reasons.push(REASON_GENERATION_BELOW_CARD.into());
    }
    // (a) the journal holds a record for this holder and generation, and its lease is live.
    let mut matching = obs
        .journal
        .iter()
        .filter(|j| j.holder == row.holder && j.generation == row.generation)
        .peekable();
    if matching.peek().is_none() {
        reasons.push(REASON_JOURNAL_MISSING.into());
    } else if !matching.any(|j| j.live) {
        reasons.push(REASON_JOURNAL_NOT_LIVE.into());
    }
    // (b) the worktree exists and is on the reservation's branch.
    match row.worktree.as_deref() {
        None => reasons.push(REASON_WORKTREE_NOT_RECORDED.into()),
        Some(path) => match obs.worktrees.iter().find(|w| w.path == path) {
            None => reasons.push(REASON_WORKTREE_NOT_OBSERVED.into()),
            Some(w) if !w.exists => reasons.push(REASON_WORKTREE_MISSING.into()),
            Some(w) => {
                if row.branch.is_none() || w.branch != row.branch {
                    reasons.push(REASON_BRANCH_MISMATCH.into());
                }
            }
        },
    }
    // (c) the remote observation for the branch agrees; no observation is "cannot confirm".
    let facts: Vec<&RemoteFact> = match row.branch.as_deref() {
        None => Vec::new(),
        Some(branch) => obs.remotes.iter().filter(|r| r.branch == branch).collect(),
    };
    if facts.is_empty() {
        reasons.push(REASON_REMOTE_ABSENT.into());
    } else if !facts.iter().all(|r| {
        r.exists
            && r.card_id.is_none_or(|c| c == row.card_id)
            && r.holder.as_ref().is_none_or(|h| *h == row.holder)
            && r.generation.is_none_or(|g| g == row.generation)
    }) {
        reasons.push(REASON_REMOTE_MISMATCH.into());
    }
    ReservationDecision {
        reservation_id: row.reservation_id,
        card_id: row.card_id,
        generation: row.generation,
        holder: row.holder.clone(),
        source_status: row.status.clone(),
        decision: if reasons.is_empty() {
            RESTORE_HELD
        } else {
            HELD_FOR_REVIEW
        }
        .into(),
        reasons,
        unreadable: false,
    }
}

// ---- the tables ---------------------------------------------------------------------------------

/// One table of the copy: its columns in insert order, the `ORDER BY` of the tolerant read, and
/// the id column when the table has an `AUTOINCREMENT` key.
struct Spec {
    name: &'static str,
    cols: &'static [&'static str],
    order: &'static str,
    id: Option<&'static str>,
}

/// Parents before children, so every foreign key finds its row.
const SPECS: [Spec; 7] = [
    Spec {
        name: "card",
        cols: &[
            "card_id",
            "title",
            "body",
            "kind",
            "state",
            "priority",
            "write_scope",
            "generation",
            "rev",
            "created_by",
            "created_at",
            "updated_at",
            "closed_at",
            "source_host",
            "source_project",
            "source_number",
            "source_url",
        ],
        order: "card_id",
        id: Some("card_id"),
    },
    Spec {
        name: "card_ac",
        cols: &[
            "ac_id",
            "card_id",
            "seq",
            "text",
            "verify",
            "status",
            "evidence_ref",
            "withdrawn_reason",
            "updated_at",
        ],
        order: "ac_id",
        id: Some("ac_id"),
    },
    Spec {
        name: "card_note",
        cols: &[
            "note_id",
            "card_id",
            "kind",
            "author",
            "origin",
            "body",
            "ref",
            "created_at",
        ],
        order: "note_id",
        id: Some("note_id"),
    },
    Spec {
        name: "card_link",
        cols: &[
            "link_id",
            "from_card",
            "kind",
            "to_card",
            "target_host",
            "target_project",
            "target_number",
            "pr_state",
            "merge_ref",
            "state",
            "created_at",
            "updated_at",
        ],
        order: "link_id",
        id: Some("link_id"),
    },
    Spec {
        name: "reservation",
        cols: &[
            "reservation_id",
            "card_id",
            "generation",
            "holder",
            "status",
            "acquired_at",
            "lease_until",
            "updated_at",
            "worktree",
            "branch",
        ],
        order: "reservation_id",
        id: Some("reservation_id"),
    },
    Spec {
        name: "reservation_scope",
        cols: &["reservation_id", "pattern"],
        order: "reservation_id, pattern",
        id: None,
    },
    Spec {
        name: "event",
        cols: &[
            "event_id",
            "event_key",
            "at",
            "actor",
            "card_id",
            "kind",
            "payload",
        ],
        order: "event_id",
        id: Some("event_id"),
    },
];

fn spec(name: &str) -> &'static Spec {
    SPECS.iter().find(|s| s.name == name).expect("known table")
}

fn col(spec: &Spec, name: &str) -> usize {
    spec.cols
        .iter()
        .position(|c| *c == name)
        .expect("known column")
}

/// What reading one source table gave.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct TableReport {
    pub table: String,
    /// `ok`, `partial` (stopped at a read error after some rows) or `unreadable` (no row could
    /// be read and the table raised an error).
    pub state: String,
    pub rows_read: u64,
    pub last_good_id: Option<i64>,
    /// How many rows after the last good one the source's own id counter says existed. `None`
    /// when the table has no id counter or the counter could not be read.
    pub lost_estimate: Option<i64>,
    pub error: Option<String>,
    /// Rows the new ledger took.
    pub copied: u64,
    /// Events whose `event_key` the new ledger already held.
    pub skipped_duplicate: u64,
    /// Rows the schema refused (a failed check, a missing parent); they stay in the source.
    pub rejected: u64,
}

/// A source row the new ledger refused.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct Rejected {
    pub table: String,
    pub id: String,
    pub reason: String,
}

/// The state of the real ledger file when the plan was made.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum TargetState {
    /// No file at the ledger path.
    Absent,
    /// The ledger path is not a regular file (a link or a folder): apply refuses.
    NotRegular,
    /// A file that the P1 open positively reports as not this project's usable ledger because of
    /// its content (not a database, corrupt, truncated, another `application_id`, a version
    /// below 1): the one apply replaces.
    Unavailable,
    /// A usable ledger: apply refuses unless it is the result of this very plan.
    Healthy,
    /// Another connection holds the file past the short probe wait. Not confirmed damaged: apply
    /// refuses (replacing it would lose the holder's write).
    Busy,
    /// Written by a newer schema than this build knows: apply refuses.
    SchemaNewer,
    /// A readable ledger of another project: apply refuses.
    IdCollision,
    /// In WAL mode (this build never opens it): apply refuses.
    Wal,
    /// A leftover rollback journal that a writer must roll back first: apply refuses.
    NeedsRecovery,
    /// The file could not be examined for a reason other than its content (an I/O or other SQL
    /// error): not confirmed damaged, apply refuses.
    Unconfirmed,
}

impl TargetState {
    /// Whether an apply may build over this state: only when no ledger is there or the content
    /// of the file is positively reported unusable.
    pub fn replaceable(self) -> bool {
        matches!(self, TargetState::Absent | TargetState::Unavailable)
    }
}

/// One difference between a source table and the table this build creates.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct SchemaDiff {
    pub table: String,
    /// Empty for a whole table that is missing.
    pub column: String,
    /// `TYPE notnull=<0|1> pk=<n>` of this build's column, or `absent`.
    pub expected: String,
    /// The same for the source, or `absent` / `unreadable: <error>`.
    pub found: String,
}

/// A `.creating-<pid>-<n>` file left in the project folder by a `register` whose process is
/// gone. Reported, never removed.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct StaleTemp {
    pub name: String,
    pub pid: i64,
}

/// The plan: read-only facts about the source and everything an apply will do.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct RecoveryPlan {
    pub project_id: String,
    pub source: PathBuf,
    /// The real ledger path the new file replaces.
    pub ledger: PathBuf,
    /// Where the new ledger is built (`<ledger>.recovering`).
    pub target: PathBuf,
    /// `quick_check` of the source: `ok`, its first complaint, or `unreadable: ...`.
    pub source_check: String,
    pub tables: Vec<TableReport>,
    pub unreadable_tables: Vec<String>,
    /// One entry per source reservation that was active.
    pub decisions: Vec<ReservationDecision>,
    pub restore_held: u64,
    pub held_for_review: u64,
    /// `released` and `reclaimed` rows are copied as they are.
    pub copied_as_is: u64,
    /// Cards whose `generation` was raised to the highest reservation generation.
    pub card_generations_raised: u64,
    pub rejected_total: u64,
    pub rejected: Vec<Rejected>,
    /// Cards that are `active` but have no active reservation in the new ledger.
    pub active_cards_without_holder: Vec<i64>,
    pub target_state: TargetState,
    /// The error text behind a [`TargetState`] that is not `absent`, `healthy` or
    /// `unavailable`, else `None`.
    pub target_detail: Option<String>,
    /// Source reservation rows that could not be read as reservations. Active ones become
    /// `held_for_review` (reason `unreadable_row`); the others are rejected, never copied.
    pub unreadable_reservations: u64,
    /// The source's table definitions that differ from this build's (columns and types). A
    /// difference alone does not refuse: the unreadable-row rule is the safety net.
    pub schema_differs: Vec<SchemaDiff>,
    pub stale_temp: Vec<StaleTemp>,
    pub plan_digest: String,
}

/// The source rows and the decisions, kept for the apply (not part of the plan's output).
struct Prepared {
    plan: RecoveryPlan,
    rows: BTreeMap<&'static str, Vec<Vec<Value>>>,
    seq: BTreeMap<String, i64>,
    card_generation: BTreeMap<i64, i64>,
    /// One per source reservation row, in the table's order.
    verdicts: Vec<RowVerdict>,
}

/// What the copy does with one source reservation row.
#[derive(Debug, Clone, PartialEq, Eq)]
enum RowVerdict {
    /// Not an active row: copied as it is.
    Keep,
    /// Active: inserted with this status (`held` or `held_for_review`).
    Status(&'static str),
    /// Not copied; the reason is recorded as a rejected row.
    Reject(&'static str),
}

// ---- reading the source -------------------------------------------------------------------------

fn unavailable(why: impl Into<String>) -> LedgerError {
    LedgerError::Unavailable(why.into())
}

fn sibling(path: &Path, suffix: &str) -> PathBuf {
    let mut os = path.as_os_str().to_os_string();
    os.push(suffix);
    PathBuf::from(os)
}

/// Opens the source read-only after the checks P1 makes before SQLite sees a file: a regular
/// file (not a link or folder), a SQLite header, not WAL, this build's `application_id` and
/// schema version (both read from the 100-byte header, so a damaged page cannot hide them), no
/// leftover journal (a read-only connection cannot roll one back). The project id inside the
/// file must be this project's.
fn open_source(source: &Path, project_id: &str) -> Result<Connection, LedgerError> {
    let meta = std::fs::symlink_metadata(source).map_err(|e| {
        if e.kind() == std::io::ErrorKind::NotFound {
            unavailable("the source file does not exist")
        } else {
            LedgerError::Io(e.to_string())
        }
    })?;
    if !meta.file_type().is_file() {
        return Err(unavailable("the source is a link or not a regular file"));
    }
    let mut head = [0u8; 100];
    {
        use std::io::Read;
        let read = std::fs::File::open(source).and_then(|mut f| f.read_exact(&mut head));
        if read.is_err() || &head[..16] != open::HEADER_MAGIC {
            return Err(unavailable("the source is not a SQLite database"));
        }
    }
    if head[18] == 2 || head[19] == 2 {
        return Err(LedgerError::WalUnexpected);
    }
    let application_id = i32::from_be_bytes([head[68], head[69], head[70], head[71]]);
    if i64::from(application_id) != APPLICATION_ID {
        return Err(unavailable(format!(
            "application_id is {application_id:#x}, not a MoAI ledger ({APPLICATION_ID:#x})"
        )));
    }
    migrate::check_version(i64::from(i32::from_be_bytes([
        head[60], head[61], head[62], head[63],
    ])))?;
    if std::fs::symlink_metadata(sibling(source, "-journal")).is_ok() {
        return Err(unavailable(
            "needs_recovery: a leftover journal sits beside the source; a writer must roll it back first (open the ledger for writing), then plan again",
        ));
    }
    let conn = Connection::open_with_flags(
        source,
        OpenFlags::SQLITE_OPEN_READ_ONLY | OpenFlags::SQLITE_OPEN_NO_MUTEX,
    )?;
    conn.busy_timeout(open::BUSY_TIMEOUT)?;
    // A file cut short is refused as a whole by SQLite ("malformed") unless the schema is
    // marked writable; the connection is read-only, so this writes nothing and only lets the
    // pages that are still there be read.
    conn.pragma_update(None, "writable_schema", true)?;
    let found: String = conn
        .query_row("SELECT value FROM meta WHERE key = 'project_id'", [], |r| {
            r.get(0)
        })
        .map_err(|e| {
            unavailable(format!(
                "the project id inside the source cannot be read ({e}); a source whose project cannot be confirmed is refused"
            ))
        })?;
    if found != project_id {
        return Err(LedgerError::IdCollision {
            expected: project_id.to_string(),
            found,
        });
    }
    Ok(conn)
}

/// Reads one table in primary-key order. A read error stops the table at the last good row;
/// it never aborts the plan.
fn read_table(
    conn: &Connection,
    spec: &Spec,
    seq: &BTreeMap<String, i64>,
) -> (Vec<Vec<Value>>, TableReport) {
    let mut rows: Vec<Vec<Value>> = Vec::new();
    let mut error: Option<String> = None;
    let sql = format!(
        "SELECT {} FROM {} ORDER BY {}",
        spec.cols.join(", "),
        spec.name,
        spec.order
    );
    match conn.prepare(&sql) {
        Err(e) => error = Some(e.to_string()),
        Ok(mut stmt) => match stmt.query([]) {
            Err(e) => error = Some(e.to_string()),
            Ok(mut cursor) => loop {
                match cursor.next() {
                    Ok(None) => break,
                    Err(e) => {
                        error = Some(e.to_string());
                        break;
                    }
                    Ok(Some(row)) => {
                        let mut values = Vec::with_capacity(spec.cols.len());
                        let mut broken = None;
                        for i in 0..spec.cols.len() {
                            match row.get::<_, Value>(i) {
                                Ok(v) => values.push(v),
                                Err(e) => {
                                    broken = Some(e.to_string());
                                    break;
                                }
                            }
                        }
                        if let Some(e) = broken {
                            error = Some(e);
                            break;
                        }
                        rows.push(values);
                    }
                }
            },
        },
    }
    let last_good_id = spec.id.and_then(|_| match rows.last().map(|r| &r[0]) {
        Some(Value::Integer(n)) => Some(*n),
        _ => None,
    });
    let lost_estimate = match (&error, spec.id, seq.get(spec.name)) {
        (None, _, _) => Some(0),
        (Some(_), Some(_), Some(counter)) => Some((counter - last_good_id.unwrap_or(0)).max(0)),
        _ => None,
    };
    let state = match (&error, rows.is_empty()) {
        (None, _) => "ok",
        (Some(_), true) => "unreadable",
        (Some(_), false) => "partial",
    };
    let report = TableReport {
        table: spec.name.to_string(),
        state: state.to_string(),
        rows_read: rows.len() as u64,
        last_good_id,
        lost_estimate,
        error,
        copied: 0,
        skipped_duplicate: 0,
        rejected: 0,
    };
    (rows, report)
}

/// The source's id counters (`sqlite_sequence`), best effort.
fn read_sequences(conn: &Connection) -> BTreeMap<String, i64> {
    let mut out = BTreeMap::new();
    if let Ok(mut stmt) = conn.prepare("SELECT name, seq FROM sqlite_sequence ORDER BY name") {
        if let Ok(mut cursor) = stmt.query([]) {
            while let Ok(Some(row)) = cursor.next() {
                if let (Ok(name), Ok(seq)) = (row.get::<_, String>(0), row.get::<_, i64>(1)) {
                    out.insert(name, seq);
                }
            }
        }
    }
    out
}

fn int(v: &Value) -> Option<i64> {
    match v {
        Value::Integer(n) => Some(*n),
        _ => None,
    }
}

fn text(v: &Value) -> Option<String> {
    match v {
        Value::Text(s) => Some(s.clone()),
        _ => None,
    }
}

fn opt_text(v: &Value) -> Option<Option<String>> {
    match v {
        Value::Null => Some(None),
        Value::Text(s) => Some(Some(s.clone())),
        _ => None,
    }
}

fn to_json(v: &Value) -> Json {
    match v {
        Value::Null => Json::Null,
        Value::Integer(n) => json!(n),
        Value::Real(f) => json!(f),
        Value::Text(s) => json!(s),
        Value::Blob(b) => json!(b.iter().map(|x| format!("{x:02x}")).collect::<String>()),
    }
}

/// Active source statuses: the ones that occupy the card's holder slot.
fn is_active(status: &str) -> bool {
    matches!(status, "held" | "expired_candidate" | HELD_FOR_REVIEW)
}

// ---- the plan -----------------------------------------------------------------------------------

/// Reads `source` and says what an [`apply`] would build. Writes nothing: not the source, not
/// the project folder. `observations` carries the journal, worktree and remote facts; without
/// them every active reservation becomes `held_for_review`.
pub fn plan(
    home: &Path,
    project_id: &str,
    source: &Path,
    observations: &Observations,
) -> Result<RecoveryPlan, LedgerError> {
    Ok(prepare(home, project_id, source, observations)?.plan)
}

/// The reservation rows of the source that the command layer must find facts for, in the order
/// of the table. Read-only; a source that cannot be opened is the same error as [`plan`].
pub fn queries(
    home: &Path,
    project_id: &str,
    source: &Path,
) -> Result<Vec<ReservationRow>, LedgerError> {
    open::check_home_ready(home)?;
    location::ledger_path(home, project_id)?;
    let conn = open_source(source, project_id)?;
    let seq = read_sequences(&conn);
    let (rows, _) = read_table(&conn, spec("reservation"), &seq);
    Ok(rows
        .iter()
        .filter_map(|r| reservation_row(r))
        .filter(|r| is_active(&r.status))
        .collect())
}

/// A reservation row whose every column holds the kind of value the schema declares.
fn reservation_row(r: &[Value]) -> Option<ReservationRow> {
    let s = spec("reservation");
    // The columns the decision does not read are still checked: a row is readable only when all
    // of it is.
    for name in ["acquired_at", "lease_until", "updated_at"] {
        int(&r[col(s, name)])?;
    }
    Some(ReservationRow {
        reservation_id: int(&r[col(s, "reservation_id")])?,
        card_id: int(&r[col(s, "card_id")])?,
        generation: int(&r[col(s, "generation")])?,
        holder: text(&r[col(s, "holder")])?,
        status: text(&r[col(s, "status")])?,
        worktree: opt_text(&r[col(s, "worktree")])?,
        branch: opt_text(&r[col(s, "branch")])?,
    })
}

/// What can be said of a reservation row that is not readable as a whole: its status when that
/// is text, and the integers and the holder that are fine.
fn unreadable_facts(r: &[Value]) -> (Option<String>, ReservationDecisionIds) {
    let s = spec("reservation");
    (
        text(&r[col(s, "status")]),
        ReservationDecisionIds {
            reservation_id: int(&r[col(s, "reservation_id")]).unwrap_or(0),
            card_id: int(&r[col(s, "card_id")]).unwrap_or(0),
            generation: int(&r[col(s, "generation")]).unwrap_or(0),
            holder: text(&r[col(s, "holder")]).unwrap_or_default(),
        },
    )
}

struct ReservationDecisionIds {
    reservation_id: i64,
    card_id: i64,
    generation: i64,
    holder: String,
}

/// The most differences a plan lists.
const MAX_SCHEMA_DIFFS: usize = 200;

/// `(column name, "TYPE notnull=<n> pk=<n>")` of `table`, in column order. An empty list means
/// the table is not there.
fn table_columns(conn: &Connection, table: &str) -> Result<Vec<(String, String)>, String> {
    let mut stmt = conn
        .prepare(&format!("PRAGMA table_info({table})"))
        .map_err(|e| e.to_string())?;
    let rows = stmt
        .query_map([], |r| {
            Ok((
                r.get::<_, String>(1)?,
                format!(
                    "{} notnull={} pk={}",
                    r.get::<_, String>(2)?.to_uppercase(),
                    r.get::<_, i64>(3)?,
                    r.get::<_, i64>(5)?
                ),
            ))
        })
        .map_err(|e| e.to_string())?;
    rows.collect::<Result<Vec<_>, _>>()
        .map_err(|e| e.to_string())
}

/// The source's definitions of the seven copied tables against the ones this build creates.
/// Reports; it never refuses (an unreadable row is the safety net for a value of the wrong
/// kind).
fn schema_differs(source: &Connection) -> Result<Vec<SchemaDiff>, LedgerError> {
    let mut expected = Connection::open_in_memory()?;
    schema::create_v1(&mut expected, "expected", "recovery-plan")?;
    let mut out = Vec::new();
    for s in &SPECS {
        let want = table_columns(&expected, s.name).unwrap_or_default();
        let found = match table_columns(source, s.name) {
            Ok(f) => f,
            Err(e) => {
                out.push(SchemaDiff {
                    table: s.name.to_string(),
                    column: String::new(),
                    expected: "table".into(),
                    found: format!("unreadable: {e}"),
                });
                continue;
            }
        };
        if found.is_empty() {
            out.push(SchemaDiff {
                table: s.name.to_string(),
                column: String::new(),
                expected: "table".into(),
                found: "absent".into(),
            });
            continue;
        }
        for (name, shape) in &want {
            match found.iter().find(|(n, _)| n == name) {
                Some((_, f)) if f == shape => {}
                Some((_, f)) => out.push(SchemaDiff {
                    table: s.name.to_string(),
                    column: name.clone(),
                    expected: shape.clone(),
                    found: f.clone(),
                }),
                None => out.push(SchemaDiff {
                    table: s.name.to_string(),
                    column: name.clone(),
                    expected: shape.clone(),
                    found: "absent".into(),
                }),
            }
        }
        for (name, shape) in &found {
            if !want.iter().any(|(n, _)| n == name) {
                out.push(SchemaDiff {
                    table: s.name.to_string(),
                    column: name.clone(),
                    expected: "absent".into(),
                    found: shape.clone(),
                });
            }
        }
    }
    out.truncate(MAX_SCHEMA_DIFFS);
    Ok(out)
}

fn prepare(
    home: &Path,
    project_id: &str,
    source: &Path,
    obs: &Observations,
) -> Result<Prepared, LedgerError> {
    open::check_home_ready(home)?;
    let ledger = location::ledger_path(home, project_id)?;
    let conn = open_source(source, project_id)?;
    let source_check = match conn.query_row("PRAGMA quick_check(1)", [], |r| r.get::<_, String>(0))
    {
        Ok(first) => first,
        Err(e) => format!("unreadable: {e}"),
    };
    let seq = read_sequences(&conn);
    let schema_differs = schema_differs(&conn)?;
    let mut rows: BTreeMap<&'static str, Vec<Vec<Value>>> = BTreeMap::new();
    let mut tables = Vec::new();
    for s in &SPECS {
        let (r, report) = read_table(&conn, s, &seq);
        rows.insert(s.name, r);
        tables.push(report);
    }
    drop(conn);

    // The restored generation of each card: its own, raised to the highest reservation
    // generation, so a recovered ledger never issues a generation twice.
    let card_spec = spec("card");
    let mut card_generation: BTreeMap<i64, i64> = BTreeMap::new();
    for r in &rows["card"] {
        if let (Some(id), Some(g)) = (
            int(&r[col(card_spec, "card_id")]),
            int(&r[col(card_spec, "generation")]),
        ) {
            card_generation.insert(id, g);
        }
    }
    let raw_generation = card_generation.clone();
    let mut readable: Vec<ReservationRow> = Vec::new();
    for r in &rows["reservation"] {
        if let Some(row) = reservation_row(r) {
            if let Some(g) = card_generation.get_mut(&row.card_id) {
                *g = (*g).max(row.generation);
            }
            readable.push(row);
        }
    }
    let card_generations_raised = card_generation
        .iter()
        .filter(|(id, g)| raw_generation.get(id) != Some(g))
        .count() as u64;

    // Decide every active reservation.
    let mut decisions: Vec<ReservationDecision> = Vec::new();
    let mut verdicts: Vec<RowVerdict> = Vec::new();
    let mut copied_as_is = 0u64;
    let mut unreadable_reservations = 0u64;
    for r in &rows["reservation"] {
        let Some(row) = reservation_row(r) else {
            // A row that is not a whole reservation cannot be judged by the rule. It must never
            // come back as `held`: an active one is held for review (the schema may still refuse
            // it on insert, then it is listed as rejected); any other is not copied.
            unreadable_reservations += 1;
            let (status, ids) = unreadable_facts(r);
            match status {
                Some(st) if is_active(&st) => {
                    decisions.push(ReservationDecision {
                        reservation_id: ids.reservation_id,
                        card_id: ids.card_id,
                        generation: ids.generation,
                        holder: ids.holder,
                        source_status: st,
                        decision: HELD_FOR_REVIEW.into(),
                        reasons: vec![REASON_UNREADABLE_ROW.into()],
                        unreadable: true,
                    });
                    verdicts.push(RowVerdict::Status(HELD_FOR_REVIEW));
                }
                _ => verdicts.push(RowVerdict::Reject(REASON_UNREADABLE_ROW)),
            }
            continue;
        };
        if !is_active(&row.status) {
            copied_as_is += 1;
            verdicts.push(RowVerdict::Keep);
            continue;
        }
        let generation = card_generation.get(&row.card_id).copied().unwrap_or(0);
        let d = decide(&row, generation, obs);
        verdicts.push(RowVerdict::Status(if d.decision == RESTORE_HELD {
            "held"
        } else {
            HELD_FOR_REVIEW
        }));
        decisions.push(d);
    }

    // Build once in memory with the same code the apply runs on the file: the counts, the
    // rejected rows and the digest are those of the real build.
    let mut mem = Connection::open_in_memory()?;
    mem.execute_batch("PRAGMA foreign_keys = ON; PRAGMA recursive_triggers = ON;")?;
    let prepared_view = View {
        rows: &rows,
        seq: &seq,
        card_generation: &card_generation,
        verdicts: &verdicts,
    };
    let copy = schema::create_v1_with(&mut mem, project_id, "recovery-plan", |tx| {
        copy_rows(tx, &prepared_view)
    })?;
    drop(mem);
    for t in &mut tables {
        if let Some(c) = copy.tables.get(t.table.as_str()) {
            t.copied = c.copied;
            t.skipped_duplicate = c.skipped_duplicate;
            t.rejected = c.rejected;
        }
    }

    let restore_held = decisions
        .iter()
        .filter(|d| d.decision == RESTORE_HELD)
        .count() as u64;
    let held_for_review = decisions.len() as u64 - restore_held;
    let active_cards_without_holder = active_cards_without_holder(&rows, &decisions, &copy);
    let unreadable_tables = tables
        .iter()
        .filter(|t| t.state == "unreadable")
        .map(|t| t.table.clone())
        .collect();

    let (target_state, target_detail) = target_state(project_id, &ledger);
    let digest_input = json!({
        "format": 1,
        "project_id": project_id,
        "tables": SPECS.iter().map(|s| json!({
            "table": s.name,
            "state": tables.iter().find(|t| t.table == s.name).map(|t| t.state.clone()),
            "rows": rows[s.name].iter().map(|r| r.iter().map(to_json).collect::<Vec<_>>()).collect::<Vec<_>>(),
        })).collect::<Vec<_>>(),
        "sequences": seq,
        "decisions": decisions,
        "copy": copy.tables.iter().map(|(k, v)| json!([k, v.copied, v.skipped_duplicate, v.rejected])).collect::<Vec<_>>(),
        "rejected": copy.rejected,
        "generations_raised": card_generations_raised,
        "unreadable_reservations": unreadable_reservations,
        "schema_differs": schema_differs,
    });
    let plan_digest = crate::digest(digest_input.to_string().as_bytes());
    let rejected_total = copy.rejected.len() as u64;
    let listed: Vec<Rejected> = copy
        .rejected
        .iter()
        .take(MAX_LISTED_REJECTED)
        .cloned()
        .collect();
    let plan = RecoveryPlan {
        project_id: project_id.to_string(),
        source: source.to_path_buf(),
        target: sibling(&ledger, RECOVERING),
        ledger,
        source_check,
        tables,
        unreadable_tables,
        decisions,
        restore_held,
        held_for_review,
        copied_as_is,
        card_generations_raised,
        rejected_total,
        rejected: listed,
        active_cards_without_holder,
        target_state,
        target_detail,
        unreadable_reservations,
        schema_differs,
        stale_temp: stale_temp(home, project_id)?,
        plan_digest,
    };
    Ok(Prepared {
        plan,
        rows,
        seq,
        card_generation,
        verdicts,
    })
}

fn active_cards_without_holder(
    rows: &BTreeMap<&'static str, Vec<Vec<Value>>>,
    decisions: &[ReservationDecision],
    copy: &CopyReport,
) -> Vec<i64> {
    let card_spec = spec("card");
    let refused = copy.rejected_ids("reservation");
    let with_holder: BTreeSet<i64> = decisions
        .iter()
        .filter(|d| !refused.contains(&d.reservation_id.to_string()))
        .map(|d| d.card_id)
        .collect();
    let mut out: Vec<i64> = rows["card"]
        .iter()
        .filter(|r| text(&r[col(card_spec, "state")]).as_deref() == Some("active"))
        .filter_map(|r| int(&r[col(card_spec, "card_id")]))
        .filter(|id| !with_holder.contains(id))
        .collect();
    out.sort_unstable();
    out
}

// ---- copying ------------------------------------------------------------------------------------

/// Borrowed view of what the copy needs.
struct View<'a> {
    rows: &'a BTreeMap<&'static str, Vec<Vec<Value>>>,
    seq: &'a BTreeMap<String, i64>,
    card_generation: &'a BTreeMap<i64, i64>,
    verdicts: &'a [RowVerdict],
}

#[derive(Default)]
struct TableCopy {
    copied: u64,
    skipped_duplicate: u64,
    rejected: u64,
}

#[derive(Default)]
struct CopyReport {
    tables: BTreeMap<&'static str, TableCopy>,
    rejected: Vec<Rejected>,
}

impl CopyReport {
    fn rejected_ids(&self, table: &str) -> BTreeSet<String> {
        self.rejected
            .iter()
            .filter(|r| r.table == table)
            .map(|r| r.id.clone())
            .collect()
    }
}

/// The key of a row for the rejected list: its id, or the two key columns of a table without one.
fn rejected_id(s: &Spec, values: &[Value]) -> String {
    values
        .iter()
        .take(if s.id.is_some() { 1 } else { 2 })
        .map(|v| match v {
            Value::Integer(n) => n.to_string(),
            Value::Text(t) => t.clone(),
            other => format!("{other:?}"),
        })
        .collect::<Vec<_>>()
        .join("/")
}

fn is_constraint(e: &rusqlite::Error) -> bool {
    matches!(
        e.sqlite_error_code(),
        Some(rusqlite::ErrorCode::ConstraintViolation)
    )
}

/// Inserts every source row with its explicit id, in parent-first order, then sets the id
/// counters so the next id is past every id the source ever issued. The reservation rows take
/// the status the decision rule gave, the cards the restored generation. A row the schema
/// refuses is recorded and left out; only an unexpected error aborts.
fn copy_rows(tx: &Transaction<'_>, view: &View<'_>) -> Result<CopyReport, LedgerError> {
    let mut report = CopyReport::default();
    for s in &SPECS {
        let marks: Vec<String> = (1..=s.cols.len()).map(|i| format!("?{i}")).collect();
        let conflict = if s.name == "event" {
            " ON CONFLICT(event_key) DO NOTHING"
        } else {
            ""
        };
        let mut insert = tx.prepare(&format!(
            "INSERT INTO {} ({}) VALUES ({}){conflict}",
            s.name,
            s.cols.join(", "),
            marks.join(", ")
        ))?;
        let copy = report.tables.entry(s.name).or_default();
        for (index, r) in view.rows[s.name].iter().enumerate() {
            let mut values = r.clone();
            if s.name == "card" {
                if let Some(restored) = int(&values[0]).and_then(|id| view.card_generation.get(&id))
                {
                    values[col(s, "generation")] = Value::Integer(*restored);
                }
            }
            if s.name == "reservation" {
                match &view.verdicts[index] {
                    RowVerdict::Keep => {}
                    RowVerdict::Status(status) => {
                        values[col(s, "status")] = Value::Text((*status).to_string());
                    }
                    RowVerdict::Reject(why) => {
                        copy.rejected += 1;
                        report.rejected.push(Rejected {
                            table: s.name.to_string(),
                            id: rejected_id(s, &values),
                            reason: (*why).to_string(),
                        });
                        continue;
                    }
                }
            }
            match insert.execute(rusqlite::params_from_iter(values.iter())) {
                Ok(0) => copy.skipped_duplicate += 1,
                Ok(_) => copy.copied += 1,
                Err(e) if is_constraint(&e) => {
                    copy.rejected += 1;
                    report.rejected.push(Rejected {
                        table: s.name.to_string(),
                        id: rejected_id(s, &values),
                        reason: e.to_string(),
                    });
                }
                Err(e) => return Err(e.into()),
            }
        }
    }
    for s in SPECS.iter().filter(|s| s.id.is_some()) {
        let id_col = s.id.expect("filtered");
        let max: Option<i64> =
            tx.query_row(&format!("SELECT max({id_col}) FROM {}", s.name), [], |r| {
                r.get(0)
            })?;
        let wanted = max
            .unwrap_or(0)
            .max(view.seq.get(s.name).copied().unwrap_or(0));
        if wanted > 0 {
            let changed = tx.execute(
                "UPDATE sqlite_sequence SET seq = ?1 WHERE name = ?2",
                rusqlite::params![wanted, s.name],
            )?;
            if changed == 0 {
                tx.execute(
                    "INSERT INTO sqlite_sequence (name, seq) VALUES (?2, ?1)",
                    rusqlite::params![wanted, s.name],
                )?;
            }
        }
    }
    Ok(report)
}

// ---- applying -----------------------------------------------------------------------------------

/// The new ledger, built and checked, waiting at `<ledger>.recovering`.
#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct Built {
    pub target: PathBuf,
    pub plan_digest: String,
    /// `PRAGMA integrity_check` of the new file (`ok` when sound).
    pub integrity_check: String,
    /// Rows `PRAGMA foreign_key_check` reported (0 when sound).
    pub foreign_key_violations: u64,
    pub recover_event_id: Option<i64>,
    /// An unfinished `.recovering` of an earlier try was replaced.
    pub replaced_unfinished: bool,
}

/// Where the damaged file was kept.
#[derive(Debug, Clone, Default, PartialEq, Eq, Serialize)]
pub struct SetAside {
    /// `<ledger>.damaged-<suffix>`; `None` when there was no file at the ledger path.
    pub damaged: Option<PathBuf>,
    /// `<ledger>-journal.damaged-<suffix>`: a leftover journal of the damaged file, moved away
    /// so it is never rolled back into the new one.
    pub journal: Option<PathBuf>,
    /// The damaged file had already been set aside by an earlier try (the same file under a
    /// `damaged` name); no second link was made.
    pub reused: bool,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum ApplyState {
    /// The new ledger is in place.
    Applied,
    /// The ledger at the real name is already the result of this plan; nothing was written.
    Unchanged,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct RecoveryResult {
    pub state: ApplyState,
    pub plan_digest: String,
    pub ledger: PathBuf,
    pub built: Option<Built>,
    pub set_aside: Option<SetAside>,
}

fn constraint(word: &str, why: impl std::fmt::Display) -> LedgerError {
    LedgerError::Constraint(format!("{word}: {why}"))
}

/// `suffix` names the kept file: 1 to 40 ASCII letters or digits.
fn valid_suffix(suffix: &str) -> bool {
    !suffix.is_empty() && suffix.len() <= 40 && suffix.bytes().all(|b| b.is_ascii_alphanumeric())
}

/// `YYYYMMDDTHHMMSSZ` now, from SQLite's clock.
fn utc_suffix() -> Result<String, LedgerError> {
    Ok(Connection::open_in_memory()?.query_row(
        "SELECT strftime('%Y%m%dT%H%M%SZ', 'now')",
        [],
        |r| r.get(0),
    )?)
}

/// The wait of the probe that classifies the real ledger. A ledger held by another connection
/// longer than this is `busy`, which is never replaced, so a short wait only makes the answer
/// quicker.
const TARGET_PROBE_BUSY: std::time::Duration = std::time::Duration::from_millis(300);

/// Classifies the file at the real ledger path from what the P1 open says. Only an error that
/// is about the file's content is "damaged" ([`TargetState::Unavailable`]); a lock, a newer
/// schema, another project, WAL, a hot journal or an I/O error is not proof of damage.
fn target_state(project_id: &str, ledger: &Path) -> (TargetState, Option<String>) {
    match std::fs::symlink_metadata(ledger) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => (TargetState::Absent, None),
        Err(e) => (TargetState::Unconfirmed, Some(e.to_string())),
        Ok(m) if !m.file_type().is_file() => (TargetState::NotRegular, None),
        // A file this process cannot even open says nothing about its content.
        Ok(_) if std::fs::File::open(ledger).is_err() => (
            TargetState::Unconfirmed,
            Some("the ledger file cannot be opened for reading".into()),
        ),
        Ok(_) => match open::open_at_timeout(ledger, project_id, false, TARGET_PROBE_BUSY) {
            Ok(_) => (TargetState::Healthy, None),
            Err(LedgerError::Absent) => (TargetState::Absent, None),
            Err(LedgerError::Unavailable(why)) if why.starts_with("needs_recovery") => {
                (TargetState::NeedsRecovery, Some(why))
            }
            Err(LedgerError::Unavailable(_)) => (TargetState::Unavailable, None),
            Err(LedgerError::Busy) => (TargetState::Busy, None),
            Err(e @ LedgerError::SchemaNewer { .. }) => {
                (TargetState::SchemaNewer, Some(e.to_string()))
            }
            Err(e @ LedgerError::IdCollision { .. }) => {
                (TargetState::IdCollision, Some(e.to_string()))
            }
            Err(e @ LedgerError::WalUnexpected) => (TargetState::Wal, Some(e.to_string())),
            Err(e) => (TargetState::Unconfirmed, Some(e.to_string())),
        },
    }
}

/// Whether the usable ledger at the real name carries the `recover` event of `plan_digest`.
fn already_applied(home: &Path, project_id: &str, plan_digest: &str) -> bool {
    open::open_read_only(home, project_id)
        .ok()
        .and_then(|r| {
            r.conn()
                .query_row(
                    "SELECT count(*) FROM event WHERE event_key = ?1",
                    [recover_key(plan_digest)],
                    |row| row.get::<_, i64>(0),
                )
                .ok()
        })
        .is_some_and(|n| n > 0)
}

fn recover_key(plan_digest: &str) -> String {
    format!("recover:{plan_digest}")
}

/// The name of the lock file in the project folder. It is created on the first recovery and
/// never removed: the lock is the advisory `flock` on it, not its existence.
pub const LOCK_NAME: &str = ".recover.lock";

/// The exclusive recovery lock of one project. Holding it is the proof the steps of an apply
/// ask for; it is released when the value is dropped (or the process ends).
#[derive(Debug)]
pub struct RecoveryLock {
    dir: PathBuf,
    _file: std::fs::File,
}

impl RecoveryLock {
    fn covers(&self, home: &Path, project_id: &str) -> Result<(), LedgerError> {
        if location::project_dir(home, project_id)? == self.dir {
            Ok(())
        } else {
            Err(constraint(
                "lock_mismatch",
                "the recovery lock belongs to another project folder",
            ))
        }
    }
}

/// Takes the recovery lock of the project without waiting. The project folder must exist. The
/// loser gets `recovery_in_progress` and has touched nothing.
pub fn lock_recovery(home: &Path, project_id: &str) -> Result<RecoveryLock, LedgerError> {
    open::check_home_ready(home)?;
    let dir = location::project_dir(home, project_id)?;
    if !dir.is_dir() {
        return Err(unavailable(
            "the project folder does not exist; register the project first",
        ));
    }
    let path = dir.join(LOCK_NAME);
    let mut options = std::fs::OpenOptions::new();
    options.write(true).create(true).truncate(false);
    #[cfg(unix)]
    {
        use std::os::unix::fs::OpenOptionsExt;
        options.mode(0o600).custom_flags(libc::O_NOFOLLOW);
    }
    let file = options
        .open(&path)
        .map_err(|e| LedgerError::Io(format!("{}: {e}", path.display())))?;
    match file.try_lock() {
        Ok(()) => Ok(RecoveryLock { dir, _file: file }),
        Err(std::fs::TryLockError::WouldBlock) => Err(constraint(
            "recovery_in_progress",
            "another recovery of this project holds the lock; nothing was changed",
        )),
        Err(std::fs::TryLockError::Error(e)) => Err(LedgerError::Io(e.to_string())),
    }
}

/// The refusal for a real ledger that is not positively unusable.
fn not_damaged(state: TargetState, detail: Option<&str>) -> LedgerError {
    constraint(
        "target_not_damaged",
        format!(
            "the ledger at the real name is `{}`, not confirmed damaged; it is never replaced{}",
            serde_json::to_value(state)
                .ok()
                .and_then(|v| v.as_str().map(str::to_string))
                .unwrap_or_default(),
            detail.map(|d| format!(" ({d})")).unwrap_or_default()
        ),
    )
}

/// Removes the file this call built (and its journal) when the apply stops before the install.
fn discard_scratch(target: &Path) {
    for path in [target.to_path_buf(), sibling(target, "-journal")] {
        if std::fs::symlink_metadata(&path).is_ok_and(|m| m.file_type().is_file()) {
            let _ = std::fs::remove_file(&path);
        }
    }
}

/// Recovers the ledger from `source` when the plan recomputed now has `plan_digest`: builds,
/// checks, sets the damaged file aside and renames the new file into place. Names the kept file
/// with the current UTC time.
pub fn apply(
    home: &Path,
    project_id: &str,
    source: &Path,
    observations: &Observations,
    plan_digest: &str,
) -> Result<RecoveryResult, LedgerError> {
    apply_with_suffix(
        home,
        project_id,
        source,
        observations,
        plan_digest,
        &utc_suffix()?,
    )
}

/// [`apply`] with the suffix of the kept file given (1 to 40 letters or digits). An existing
/// `<ledger>.damaged-<suffix>` that is not this same file is refused, so a second recovery
/// needs a different suffix.
pub fn apply_with_suffix(
    home: &Path,
    project_id: &str,
    source: &Path,
    observations: &Observations,
    plan_digest: &str,
    suffix: &str,
) -> Result<RecoveryResult, LedgerError> {
    if !valid_suffix(suffix) {
        return Err(constraint(
            "invalid_suffix",
            "use 1 to 40 letters or digits",
        ));
    }
    let ledger = location::ledger_path(home, project_id)?;
    // One recovery at a time per project, across processes. The loser changes nothing.
    let lock = lock_recovery(home, project_id)?;
    if target_state(project_id, &ledger).0 == TargetState::Healthy
        && already_applied(home, project_id, plan_digest)
    {
        return Ok(RecoveryResult {
            state: ApplyState::Unchanged,
            plan_digest: plan_digest.to_string(),
            ledger,
            built: None,
            set_aside: None,
        });
    }
    let built = build_recovering(home, project_id, source, observations, plan_digest, &lock)?;
    // The state of the real file can have changed during the build (a writer may have revived
    // it): replace it only while it is still positively unusable.
    let (state, detail) = target_state(project_id, &ledger);
    if !state.replaceable() {
        discard_scratch(&built.target);
        return Err(not_damaged(state, detail.as_deref()));
    }
    let set_aside = match set_aside_damaged(home, project_id, suffix, &lock) {
        Ok(kept) => kept,
        // The new file is this call's own scratch (the lock is held): nothing else is left behind.
        Err(e) => {
            discard_scratch(&built.target);
            return Err(e);
        }
    };
    install_recovered(home, project_id, &lock)?;
    Ok(RecoveryResult {
        state: ApplyState::Applied,
        plan_digest: plan_digest.to_string(),
        ledger,
        built: Some(built),
        set_aside: Some(set_aside),
    })
}

/// Step one of the apply: recomputes the plan, requires `plan_digest`, builds the new ledger at
/// `<ledger>.recovering` and checks it. Changes nothing but that file. An unfinished
/// `.recovering` (and its journal) of an earlier try is a scratch copy of the same source and is
/// replaced; a link or folder at that name is refused.
pub fn build_recovering(
    home: &Path,
    project_id: &str,
    source: &Path,
    observations: &Observations,
    plan_digest: &str,
    lock: &RecoveryLock,
) -> Result<Built, LedgerError> {
    lock.covers(home, project_id)?;
    let p = prepare(home, project_id, source, observations)?;
    if p.plan.plan_digest != plan_digest {
        return Err(constraint(
            "stale_plan",
            "the recomputed digest differs; preview again, nothing was written",
        ));
    }
    match p.plan.target_state {
        TargetState::Healthy => {
            return Err(constraint(
                "target_healthy",
                "the ledger at the real name is usable and is not this plan's result; it is never replaced",
            ))
        }
        TargetState::NotRegular => {
            return Err(constraint(
                "target_not_regular",
                "the ledger path is a link or a folder",
            ))
        }
        TargetState::Absent | TargetState::Unavailable => {}
        other => return Err(not_damaged(other, p.plan.target_detail.as_deref())),
    }
    let target = p.plan.target.clone();
    let mut replaced_unfinished = false;
    for stale in [target.clone(), sibling(&target, "-journal")] {
        match std::fs::symlink_metadata(&stale) {
            Err(_) => {}
            Ok(m) if m.file_type().is_file() => {
                std::fs::remove_file(&stale).map_err(|e| LedgerError::Io(e.to_string()))?;
                replaced_unfinished = true;
            }
            Ok(_) => {
                return Err(constraint(
                    "recovering_not_regular",
                    format!("{} is a link or a folder", stale.display()),
                ))
            }
        }
    }
    if !target.parent().is_some_and(Path::is_dir) {
        return Err(unavailable(
            "the project folder does not exist; register the project first",
        ));
    }
    let view = View {
        rows: &p.rows,
        seq: &p.seq,
        card_generation: &p.card_generation,
        verdicts: &p.verdicts,
    };
    let event_id = open::build_new_with(&target, project_id, |tx| {
        let copy = copy_rows(tx, &view)?;
        let payload = json!({
            "plan_digest": plan_digest,
            "source": source.display().to_string(),
            "tables": p.plan.tables.iter().map(|t| json!({
                "table": t.table, "state": t.state, "rows_read": t.rows_read,
                "lost_estimate": t.lost_estimate, "copied": t.copied,
                "skipped_duplicate": t.skipped_duplicate, "rejected": t.rejected,
            })).collect::<Vec<_>>(),
            "decisions": p.plan.decisions,
            "restore_held": p.plan.restore_held,
            "held_for_review": p.plan.held_for_review,
            "copied_as_is": p.plan.copied_as_is,
            "card_generations_raised": p.plan.card_generations_raised,
            "rejected_total": copy.rejected.len(),
        });
        let inserted = tx.execute(
            "INSERT INTO event (event_key, at, actor, card_id, kind, payload) \
             VALUES (?1, unixepoch(), 'ledger-recover', NULL, 'recover', ?2) \
             ON CONFLICT(event_key) DO NOTHING",
            rusqlite::params![recover_key(plan_digest), payload.to_string()],
        )?;
        Ok(if inserted == 1 {
            Some(tx.last_insert_rowid())
        } else {
            None
        })
    })?;
    let (integrity_check, foreign_key_violations) = check_built(&target, project_id)?;
    Ok(Built {
        target,
        plan_digest: plan_digest.to_string(),
        integrity_check,
        foreign_key_violations,
        recover_event_id: event_id,
        replaced_unfinished,
    })
}

/// `PRAGMA integrity_check` and `PRAGMA foreign_key_check` of a finished file. Any complaint
/// is an error: the file is not put in place.
pub fn check_built(path: &Path, project_id: &str) -> Result<(String, u64), LedgerError> {
    let (conn, _) = open::open_at(path, project_id, false)?;
    let integrity: String = conn.query_row("PRAGMA integrity_check", [], |r| r.get(0))?;
    let mut stmt = conn.prepare("PRAGMA foreign_key_check")?;
    let violations = stmt.query_map([], |_| Ok(()))?.count() as u64;
    if integrity != "ok" || violations != 0 {
        return Err(constraint(
            "check_failed",
            format!("integrity_check `{integrity}`, {violations} foreign key violation(s)"),
        ));
    }
    Ok((integrity, violations))
}

#[cfg(unix)]
fn same_file(a: &std::fs::Metadata, b: &std::fs::Metadata) -> bool {
    use std::os::unix::fs::MetadataExt;
    a.dev() == b.dev() && a.ino() == b.ino()
}

#[cfg(not(unix))]
fn same_file(_a: &std::fs::Metadata, _b: &std::fs::Metadata) -> bool {
    false
}

/// Step two: keeps the damaged file under `<ledger>.damaged-<suffix>` without moving the real
/// name (a hard link, so the real name never points at nothing), and moves a leftover
/// `<ledger>-journal` to `<ledger>-journal.damaged-<suffix>` so it cannot be rolled into the new
/// file. Refuses when either kept name already exists and is not this same file. When the
/// damaged file already has a `damaged` name from an earlier try, that name is reused. A
/// file system without hard links answers `io`.
pub fn set_aside_damaged(
    home: &Path,
    project_id: &str,
    suffix: &str,
    lock: &RecoveryLock,
) -> Result<SetAside, LedgerError> {
    lock.covers(home, project_id)?;
    if !valid_suffix(suffix) {
        return Err(constraint(
            "invalid_suffix",
            "use 1 to 40 letters or digits",
        ));
    }
    let ledger = location::ledger_path(home, project_id)?;
    let meta = match std::fs::symlink_metadata(&ledger) {
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            return Ok(SetAside::default());
        }
        Err(e) => return Err(LedgerError::Io(e.to_string())),
        Ok(m) if m.file_type().is_file() => m,
        Ok(_) => {
            return Err(constraint(
                "target_not_regular",
                "the ledger path is a link or a folder",
            ))
        }
    };
    let kept = sibling(&ledger, &format!("{DAMAGED}{suffix}"));
    let journal = sibling(&ledger, "-journal");
    let kept_journal = sibling(&ledger, &format!("-journal{DAMAGED}{suffix}"));
    let journal_present = std::fs::symlink_metadata(&journal).is_ok();
    if journal_present && std::fs::symlink_metadata(&kept_journal).is_ok() {
        return Err(constraint(
            "suffix_in_use",
            format!("{} exists", kept_journal.display()),
        ));
    }
    // An earlier try may already have kept this very file.
    let dir = ledger.parent().expect("a ledger path has a folder");
    let prefix = format!(
        "{}{DAMAGED}",
        ledger.file_name().expect("file name").to_string_lossy()
    );
    let mut earlier: Vec<PathBuf> = std::fs::read_dir(dir)
        .map_err(|e| LedgerError::Io(e.to_string()))?
        .filter_map(|e| e.ok())
        .filter(|e| e.file_name().to_string_lossy().starts_with(&prefix))
        .map(|e| e.path())
        .filter(|p| {
            std::fs::symlink_metadata(p)
                .is_ok_and(|m| m.file_type().is_file() && same_file(&m, &meta))
        })
        .collect();
    earlier.sort();
    let (damaged, reused) = if let Some(existing) = earlier.into_iter().next() {
        (existing, true)
    } else {
        match std::fs::symlink_metadata(&kept) {
            Ok(_) => {
                return Err(constraint(
                    "suffix_in_use",
                    format!("{} exists", kept.display()),
                ))
            }
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => {}
            Err(e) => return Err(LedgerError::Io(e.to_string())),
        }
        std::fs::hard_link(&ledger, &kept).map_err(|e| LedgerError::Io(e.to_string()))?;
        (kept, false)
    };
    let journal_kept = if journal_present {
        std::fs::rename(&journal, &kept_journal).map_err(|e| LedgerError::Io(e.to_string()))?;
        Some(kept_journal)
    } else {
        None
    };
    Ok(SetAside {
        damaged: Some(damaged),
        journal: journal_kept,
        reused,
    })
}

/// Step three: renames `<ledger>.recovering` over the real name. Refuses while a file at the
/// real name has no other name (it would be lost): [`set_aside_damaged`] first.
pub fn install_recovered(
    home: &Path,
    project_id: &str,
    lock: &RecoveryLock,
) -> Result<(), LedgerError> {
    lock.covers(home, project_id)?;
    let ledger = location::ledger_path(home, project_id)?;
    let target = sibling(&ledger, RECOVERING);
    match std::fs::symlink_metadata(&target) {
        Ok(m) if m.file_type().is_file() => {}
        _ => {
            return Err(constraint(
                "nothing_to_install",
                "there is no finished new ledger",
            ))
        }
    }
    match std::fs::symlink_metadata(&ledger) {
        Err(_) => {}
        Ok(m) if m.file_type().is_file() && has_other_name(&m) => {}
        Ok(_) => {
            return Err(constraint(
                "damaged_not_set_aside",
                "the file at the real name has no other name; set it aside first",
            ))
        }
    }
    std::fs::rename(&target, &ledger).map_err(|e| LedgerError::Io(e.to_string()))?;
    if let Ok(dir) = std::fs::File::open(ledger.parent().expect("folder")) {
        let _ = dir.sync_all();
    }
    Ok(())
}

#[cfg(unix)]
fn has_other_name(m: &std::fs::Metadata) -> bool {
    use std::os::unix::fs::MetadataExt;
    m.nlink() > 1
}

#[cfg(not(unix))]
fn has_other_name(_m: &std::fs::Metadata) -> bool {
    true
}

// ---- leftovers of register ----------------------------------------------------------------------

/// The `.creating-<pid>-<n>` files in the project folder whose process is not running on this
/// machine: a `register` that died. Report only: nothing is removed or renamed. A process that
/// cannot be judged is not listed. An absent folder is an empty list.
pub fn stale_temp(home: &Path, project_id: &str) -> Result<Vec<StaleTemp>, LedgerError> {
    let dir = location::project_dir(home, project_id)?;
    let entries = match std::fs::read_dir(&dir) {
        Ok(e) => e,
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => return Ok(Vec::new()),
        Err(e) => return Err(LedgerError::Io(e.to_string())),
    };
    let mut out: Vec<StaleTemp> = entries
        .filter_map(|e| e.ok())
        .filter_map(|e| {
            let name = e.file_name().to_string_lossy().into_owned();
            let pid = temp_pid(&name)?;
            (crate::home_state::pid_alive(pid) == Some(false)).then_some(StaleTemp { name, pid })
        })
        .collect();
    out.sort_by(|a, b| a.name.cmp(&b.name));
    Ok(out)
}

/// The pid in a `.creating-<pid>-<n>` name; `None` for any other shape.
fn temp_pid(name: &str) -> Option<i64> {
    let rest = name.strip_prefix(open::TEMP_PREFIX)?;
    let (pid, n) = rest.split_once('-')?;
    let digits = |s: &str| !s.is_empty() && s.bytes().all(|b| b.is_ascii_digit());
    if !digits(pid) || !digits(n) {
        return None;
    }
    pid.parse::<i64>()
        .ok()
        .filter(|p| *p > 0 && *p <= i64::from(i32::MAX))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn row(card: i64, generation: i64) -> ReservationRow {
        ReservationRow {
            reservation_id: 1,
            card_id: card,
            generation,
            holder: "h".into(),
            status: "held".into(),
            worktree: Some("/w".into()),
            branch: Some("b".into()),
        }
    }

    fn all_true(r: &ReservationRow) -> Observations {
        Observations {
            journal: vec![JournalFact {
                holder: r.holder.clone(),
                generation: r.generation,
                live: true,
            }],
            worktrees: vec![WorktreeFact {
                path: "/w".into(),
                exists: true,
                branch: Some("b".into()),
                clean: true,
            }],
            remotes: vec![RemoteFact {
                branch: "b".into(),
                card_id: Some(r.card_id),
                holder: Some(r.holder.clone()),
                generation: Some(r.generation),
                exists: true,
            }],
        }
    }

    /// The source is opened read-only: a write through the connection is refused by SQLite.
    #[test]
    fn the_source_connection_cannot_write() {
        let dir = tempfile::tempdir().unwrap();
        let path = dir.path().join("source.db");
        {
            let mut conn = Connection::open(&path).unwrap();
            schema::create_v1(&mut conn, "p", "t").unwrap();
        }
        let conn = open_source(&path, "p").unwrap();
        let err = conn
            .execute_batch("CREATE TABLE should_not_exist (a)")
            .unwrap_err();
        assert!(
            err.to_string().contains("readonly"),
            "a read-only connection refuses the write: {err}"
        );
        let after: i64 = Connection::open(&path)
            .unwrap()
            .query_row(
                "SELECT count(*) FROM sqlite_master WHERE name = 'should_not_exist'",
                [],
                |r| r.get(0),
            )
            .unwrap();
        assert_eq!(after, 0);
    }

    /// The decision table: every condition is the only one that fails once.
    #[test]
    fn restore_held_needs_all_three_confirmations() {
        let r = row(7, 3);
        let ok = all_true(&r);
        assert_eq!(decide(&r, 3, &ok).decision, RESTORE_HELD);
        assert!(decide(&r, 3, &ok).reasons.is_empty());

        let mut no_journal = ok.clone();
        no_journal.journal.clear();
        let mut dead_journal = ok.clone();
        dead_journal.journal[0].live = false;
        let mut other_generation = ok.clone();
        other_generation.journal[0].generation = 2;
        let mut other_holder = ok.clone();
        other_holder.journal[0].holder = "x".into();
        let mut no_worktree = ok.clone();
        no_worktree.worktrees.clear();
        let mut gone_worktree = ok.clone();
        gone_worktree.worktrees[0].exists = false;
        let mut other_branch = ok.clone();
        other_branch.worktrees[0].branch = Some("c".into());
        let mut detached = ok.clone();
        detached.worktrees[0].branch = None;
        let mut no_remote = ok.clone();
        no_remote.remotes.clear();
        let mut other_remote_branch = ok.clone();
        other_remote_branch.remotes[0].branch = "c".into();
        let mut remote_gone = ok.clone();
        remote_gone.remotes[0].exists = false;
        let mut remote_holder = ok.clone();
        remote_holder.remotes[0].holder = Some("x".into());
        let mut remote_generation = ok.clone();
        remote_generation.remotes[0].generation = Some(2);
        let mut remote_card = ok.clone();
        remote_card.remotes[0].card_id = Some(8);
        let cases: Vec<(&str, Observations, &str)> = vec![
            ("no journal record", no_journal, REASON_JOURNAL_MISSING),
            ("journal not live", dead_journal, REASON_JOURNAL_NOT_LIVE),
            (
                "journal of another generation",
                other_generation,
                REASON_JOURNAL_MISSING,
            ),
            (
                "journal of another holder",
                other_holder,
                REASON_JOURNAL_MISSING,
            ),
            (
                "worktree not observed",
                no_worktree,
                REASON_WORKTREE_NOT_OBSERVED,
            ),
            ("worktree missing", gone_worktree, REASON_WORKTREE_MISSING),
            (
                "worktree on another branch",
                other_branch,
                REASON_BRANCH_MISMATCH,
            ),
            ("worktree detached", detached, REASON_BRANCH_MISMATCH),
            ("no remote observation", no_remote, REASON_REMOTE_ABSENT),
            (
                "remote fact for another branch",
                other_remote_branch,
                REASON_REMOTE_ABSENT,
            ),
            (
                "branch gone from the remote",
                remote_gone,
                REASON_REMOTE_MISMATCH,
            ),
            (
                "remote holder differs",
                remote_holder,
                REASON_REMOTE_MISMATCH,
            ),
            (
                "remote generation differs",
                remote_generation,
                REASON_REMOTE_MISMATCH,
            ),
            ("remote card differs", remote_card, REASON_REMOTE_MISMATCH),
        ];
        for (name, obs, reason) in cases {
            let d = decide(&r, 3, &obs);
            assert_eq!(d.decision, HELD_FOR_REVIEW, "{name}");
            assert_eq!(d.reasons, vec![reason.to_string()], "{name}");
        }
    }

    #[test]
    fn a_reservation_below_the_cards_generation_or_without_a_holder_is_held_for_review() {
        let r = row(7, 3);
        let ok = all_true(&r);
        // N: equal to the card's generation passes; N-1 is superseded.
        assert_eq!(decide(&r, 3, &ok).decision, RESTORE_HELD);
        let d = decide(&r, 4, &ok);
        assert_eq!(d.decision, HELD_FOR_REVIEW);
        assert_eq!(d.reasons, vec![REASON_GENERATION_BELOW_CARD.to_string()]);
        for holder in ["", "  "] {
            let mut blank = row(7, 3);
            blank.holder = holder.into();
            let d = decide(&blank, 3, &all_true(&blank));
            assert_eq!(
                d.reasons,
                vec![REASON_HOLDER_EMPTY.to_string()],
                "{holder:?}"
            );
        }
        let mut unrecorded = row(7, 3);
        unrecorded.worktree = None;
        assert_eq!(
            decide(&unrecorded, 3, &all_true(&unrecorded)).reasons,
            vec![REASON_WORKTREE_NOT_RECORDED.to_string()]
        );
        let mut no_branch = row(7, 3);
        no_branch.branch = None;
        let d = decide(&no_branch, 3, &all_true(&no_branch));
        assert_eq!(
            d.reasons,
            vec![
                REASON_BRANCH_MISMATCH.to_string(),
                REASON_REMOTE_ABSENT.to_string()
            ]
        );
    }

    #[test]
    fn with_no_observation_at_all_every_condition_is_named() {
        let d = decide(&row(1, 1), 1, &Observations::default());
        assert_eq!(
            d.reasons,
            vec![
                REASON_JOURNAL_MISSING.to_string(),
                REASON_WORKTREE_NOT_OBSERVED.to_string(),
                REASON_REMOTE_ABSENT.to_string()
            ]
        );
    }

    #[test]
    fn a_live_journal_record_wins_over_a_dead_one_for_the_same_claim() {
        let r = row(7, 3);
        let mut obs = all_true(&r);
        obs.journal.insert(
            0,
            JournalFact {
                holder: "h".into(),
                generation: 3,
                live: false,
            },
        );
        assert_eq!(decide(&r, 3, &obs).decision, RESTORE_HELD);
    }

    #[test]
    fn every_remote_fact_for_the_branch_must_agree() {
        let r = row(7, 3);
        let mut obs = all_true(&r);
        obs.remotes.push(RemoteFact {
            branch: "b".into(),
            card_id: None,
            holder: Some("other".into()),
            generation: None,
            exists: true,
        });
        assert_eq!(
            decide(&r, 3, &obs).reasons,
            vec![REASON_REMOTE_MISMATCH.to_string()]
        );
    }

    #[test]
    fn a_temp_name_has_one_pid_and_one_counter() {
        assert_eq!(temp_pid(".creating-12-0"), Some(12));
        assert_eq!(temp_pid(".creating-2147483647-9"), Some(2_147_483_647));
        for bad in [
            ".creating-2147483648-0",
            ".creating-0-0",
            ".creating--1-0",
            ".creating-1",
            ".creating-1-",
            ".creating-1-2-3",
            ".creating-x-1",
            ".creating-1-x",
            "creating-1-0",
            "ledger",
            "",
        ] {
            assert_eq!(temp_pid(bad), None, "{bad:?}");
        }
    }

    #[test]
    fn suffixes_are_short_and_plain() {
        assert!(valid_suffix("20261116T010203Z"));
        assert!(valid_suffix(&"a".repeat(40)));
        assert!(!valid_suffix(&"a".repeat(41)));
        for bad in ["", "a-b", "a b", "a/b", "..", "é"] {
            assert!(!valid_suffix(bad), "{bad:?}");
        }
    }

    #[test]
    fn values_become_json_without_loss() {
        assert_eq!(to_json(&Value::Null), Json::Null);
        assert_eq!(to_json(&Value::Integer(-3)), json!(-3));
        assert_eq!(to_json(&Value::Text("é".into())), json!("é"));
        assert_eq!(to_json(&Value::Blob(vec![0, 255])), json!("00ff"));
    }
}
