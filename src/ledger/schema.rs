//! Ledger schema version 1 (Issue #3068 P1): the tables, indexes and triggers of the contract
//! comment 6896940106 with its amendment 1 (comment 6897392517). The internal tables (`fence_log`,
//! `slot_hold`, `reclaim`, `recovery_*`) are not here: the slices that use them add them.
//!
//! Rules the SQL keeps, so no command can get around them:
//! - Rows of `card`, `card_ac`, `card_note`, `card_link`, `reservation` and `event` are never
//!   deleted (`*_no_delete`); notes and events are never changed.
//! - `card.state` moves only along `card_transition`; `done` needs criteria, no open one, and a
//!   merged pull request link — or, under M-Flow, a `card.landed` event (`card_done_guard`).
//! - There is no state check on `INSERT`: a recovery inserts a card with an explicit id, `rev`
//!   and state, then its criteria and links.
//! - `card_touch` and the criterion triggers keep `rev` and `updated_at` current.

use super::{APPLICATION_ID, SCHEMA_VERSION};
use rusqlite::Connection;

/// Tables, indexes and the seeded transition table.
const TABLES: &str = r"
CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL) STRICT;

CREATE TABLE card (
  card_id        INTEGER PRIMARY KEY AUTOINCREMENT,
  title          TEXT    NOT NULL CHECK (length(title) BETWEEN 1 AND 300),
  body           TEXT    NOT NULL DEFAULT '',
  kind           TEXT    NOT NULL DEFAULT 'task'
                   CHECK (kind IN ('task','feature','fix','docs','research','audit','chore')),
  state          TEXT    NOT NULL DEFAULT 'draft'
                   CHECK (state IN ('draft','ready','active','review','done','hold')),
  priority       TEXT    NOT NULL DEFAULT 'p2' CHECK (priority IN ('p0','p1','p2','p3')),
  write_scope    TEXT    NOT NULL DEFAULT '[]'
                   CHECK (json_valid(write_scope) AND json_type(write_scope)='array'),
  generation     INTEGER NOT NULL DEFAULT 0,
  rev            INTEGER NOT NULL DEFAULT 1,
  created_by     TEXT    NOT NULL,
  created_at     INTEGER NOT NULL,
  updated_at     INTEGER NOT NULL,
  closed_at      INTEGER,
  source_host    TEXT    CHECK (source_host IN ('github','gitlab')),
  source_project TEXT    CHECK (source_project IS NULL OR (length(source_project) > 0
                           AND source_project = lower(source_project)
                           AND instr(source_project, ' ') = 0)),
  source_number  INTEGER CHECK (source_number > 0),
  source_url     TEXT,
  CHECK ((source_host IS NULL) = (source_project IS NULL)
     AND (source_host IS NULL) = (source_number  IS NULL)),
  CHECK ((state = 'done') = (closed_at IS NOT NULL)),
  UNIQUE (source_host, source_project, source_number)
) STRICT;

CREATE TABLE card_ac (
  ac_id            INTEGER PRIMARY KEY AUTOINCREMENT,
  card_id          INTEGER NOT NULL REFERENCES card(card_id),
  seq              INTEGER NOT NULL CHECK (seq >= 1),
  text             TEXT    NOT NULL CHECK (length(text) > 0),
  verify           TEXT    NOT NULL DEFAULT '',
  status           TEXT    NOT NULL DEFAULT 'open' CHECK (status IN ('open','met','withdrawn')),
  evidence_ref     TEXT,
  withdrawn_reason TEXT,
  updated_at       INTEGER NOT NULL,
  UNIQUE (card_id, seq),
  CHECK (status <> 'met' OR evidence_ref IS NOT NULL),
  CHECK (status <> 'withdrawn' OR withdrawn_reason IS NOT NULL)
) STRICT;

CREATE TABLE card_note (
  note_id        INTEGER PRIMARY KEY AUTOINCREMENT,
  card_id        INTEGER NOT NULL REFERENCES card(card_id),
  kind           TEXT    NOT NULL CHECK (kind IN ('decision','report','observation','comment')),
  author         TEXT    NOT NULL,
  origin         TEXT    NOT NULL DEFAULT 'local' CHECK (origin IN ('local','import')),
  body           TEXT    NOT NULL CHECK (length(body) > 0),
  ref            TEXT,
  created_at     INTEGER NOT NULL
) STRICT;
CREATE UNIQUE INDEX card_note_ref ON card_note(card_id, ref) WHERE ref IS NOT NULL;

CREATE TABLE card_link (
  link_id        INTEGER PRIMARY KEY AUTOINCREMENT,
  from_card      INTEGER NOT NULL REFERENCES card(card_id),
  kind           TEXT    NOT NULL CHECK (kind IN ('blocks','split_of','relates','pr')),
  to_card        INTEGER REFERENCES card(card_id),
  target_host    TEXT    CHECK (target_host IN ('github','gitlab')),
  target_project TEXT    CHECK (target_project IS NULL OR (length(target_project) > 0
                           AND target_project = lower(target_project)
                           AND instr(target_project, ' ') = 0)),
  target_number  INTEGER CHECK (target_number > 0),
  pr_state       TEXT    CHECK (pr_state IN ('open','merged','closed')),
  merge_ref      TEXT,
  state          TEXT    NOT NULL DEFAULT 'active' CHECK (state IN ('active','retired')),
  created_at     INTEGER NOT NULL,
  updated_at     INTEGER NOT NULL,
  CHECK ((kind = 'pr') = (to_card IS NULL)),
  CHECK ((kind = 'pr') = (target_host IS NOT NULL AND target_project IS NOT NULL AND target_number IS NOT NULL)),
  CHECK (pr_state IS NULL OR kind = 'pr'),
  CHECK (pr_state <> 'merged' OR merge_ref IS NOT NULL),
  CHECK (to_card IS NULL OR to_card <> from_card)
) STRICT;
CREATE UNIQUE INDEX card_link_card ON card_link(from_card, kind, to_card) WHERE to_card IS NOT NULL;
CREATE UNIQUE INDEX card_link_pr   ON card_link(from_card, target_host, target_project, target_number) WHERE kind = 'pr';

CREATE TABLE card_transition (
  from_state TEXT NOT NULL,
  to_state   TEXT NOT NULL,
  PRIMARY KEY (from_state, to_state)
) STRICT;
INSERT INTO card_transition (from_state, to_state) VALUES
  ('draft','ready'), ('ready','draft'), ('ready','active'), ('active','ready'),
  ('active','review'), ('review','active'), ('review','done'), ('done','ready'),
  ('draft','hold'), ('ready','hold'), ('active','hold'), ('review','hold'),
  ('hold','draft'), ('hold','ready');

CREATE TABLE reservation (
  reservation_id INTEGER PRIMARY KEY AUTOINCREMENT,
  card_id        INTEGER NOT NULL REFERENCES card(card_id),
  generation     INTEGER NOT NULL CHECK (generation >= 1),
  holder         TEXT    NOT NULL,
  status         TEXT    NOT NULL CHECK (status IN ('held','released','expired_candidate','reclaimed','held_for_review')),
  acquired_at    INTEGER NOT NULL,
  lease_until    INTEGER NOT NULL,
  updated_at     INTEGER NOT NULL,
  worktree       TEXT,
  branch         TEXT,
  UNIQUE (card_id, generation)
) STRICT;
CREATE UNIQUE INDEX reservation_one_active ON reservation(card_id)
  WHERE status IN ('held','expired_candidate','held_for_review');
CREATE TABLE reservation_scope (
  reservation_id INTEGER NOT NULL REFERENCES reservation(reservation_id),
  pattern        TEXT    NOT NULL,
  PRIMARY KEY (reservation_id, pattern)
) STRICT;

CREATE TABLE event (
  event_id       INTEGER PRIMARY KEY AUTOINCREMENT,
  event_key      TEXT    NOT NULL UNIQUE,
  at             INTEGER NOT NULL,
  actor          TEXT    NOT NULL,
  card_id        INTEGER REFERENCES card(card_id),
  kind           TEXT    NOT NULL,
  payload        TEXT    NOT NULL DEFAULT '{}' CHECK (json_valid(payload))
) STRICT;
";

/// The triggers up to the done guard (contract section 2 and amendment 1 section A-4). The guard
/// is its own constant so a ledger made before its `card.landed` exception can be given the same
/// text ([`super::migrate::upgrade_done_guard`]).
const TRIGGERS_BEFORE_GUARD: &str = r"
CREATE TRIGGER card_no_delete BEFORE DELETE ON card
BEGIN SELECT RAISE(ABORT, 'rows are never deleted'); END;
CREATE TRIGGER card_ac_no_delete BEFORE DELETE ON card_ac
BEGIN SELECT RAISE(ABORT, 'rows are never deleted'); END;
CREATE TRIGGER card_note_no_delete BEFORE DELETE ON card_note
BEGIN SELECT RAISE(ABORT, 'rows are never deleted'); END;
CREATE TRIGGER card_link_no_delete BEFORE DELETE ON card_link
BEGIN SELECT RAISE(ABORT, 'rows are never deleted'); END;
CREATE TRIGGER reservation_no_delete BEFORE DELETE ON reservation
BEGIN SELECT RAISE(ABORT, 'rows are never deleted'); END;
CREATE TRIGGER event_no_delete BEFORE DELETE ON event
BEGIN SELECT RAISE(ABORT, 'rows are never deleted'); END;

CREATE TRIGGER card_note_no_update BEFORE UPDATE ON card_note
BEGIN SELECT RAISE(ABORT, 'notes are append-only'); END;
CREATE TRIGGER event_no_update BEFORE UPDATE ON event
BEGIN SELECT RAISE(ABORT, 'events are append-only'); END;

CREATE TRIGGER card_ac_frozen_keys BEFORE UPDATE OF card_id, seq ON card_ac
WHEN NEW.card_id IS NOT OLD.card_id OR NEW.seq IS NOT OLD.seq
BEGIN SELECT RAISE(ABORT, 'a criterion keeps its card and number; withdraw it instead'); END;

CREATE TRIGGER card_state_transition BEFORE UPDATE OF state ON card
WHEN NEW.state <> OLD.state
 AND NOT EXISTS (SELECT 1 FROM card_transition WHERE from_state = OLD.state AND to_state = NEW.state)
BEGIN SELECT RAISE(ABORT, 'illegal card state transition'); END;
";

/// The one trigger of the done rule. A card closes only with a met criterion, no open one, and a
/// merged pull request link or a `card.landed` event. Ledgers made before the event existed hold
/// the text without its last `NOT EXISTS`; the migration replaces it with this one.
pub(crate) const CARD_DONE_GUARD: &str = r"
CREATE TRIGGER card_done_guard BEFORE UPDATE OF state ON card
WHEN NEW.state = 'done' AND OLD.state <> 'done'
BEGIN
  SELECT RAISE(ABORT, 'cannot close: the card has no acceptance criterion that is met')
   WHERE NOT EXISTS (SELECT 1 FROM card_ac WHERE card_id = NEW.card_id AND status = 'met');
  SELECT RAISE(ABORT, 'cannot close: an acceptance criterion is still open')
   WHERE EXISTS (SELECT 1 FROM card_ac WHERE card_id = NEW.card_id AND status = 'open');
  SELECT RAISE(ABORT, 'cannot close: no merged pull request is linked')
   WHERE NOT EXISTS (SELECT 1 FROM card_link
                      WHERE from_card = NEW.card_id AND kind = 'pr'
                        AND state = 'active' AND pr_state = 'merged')
     AND NOT EXISTS (SELECT 1 FROM event
                      WHERE card_id = NEW.card_id AND kind = 'card.landed');
END;
";

/// The triggers after the done guard.
const TRIGGERS_AFTER_GUARD: &str = r"
CREATE TRIGGER card_touch AFTER UPDATE ON card
WHEN NEW.rev = OLD.rev
BEGIN UPDATE card SET rev = OLD.rev + 1, updated_at = unixepoch() WHERE card_id = NEW.card_id; END;
CREATE TRIGGER card_touch_ac_insert AFTER INSERT ON card_ac
BEGIN UPDATE card SET rev = rev + 1, updated_at = unixepoch() WHERE card_id = NEW.card_id; END;
CREATE TRIGGER card_touch_ac_update AFTER UPDATE ON card_ac
BEGIN UPDATE card SET rev = rev + 1, updated_at = unixepoch() WHERE card_id = NEW.card_id; END;
";

/// All triggers, in the order they are created.
fn triggers() -> String {
    format!("{TRIGGERS_BEFORE_GUARD}{CARD_DONE_GUARD}{TRIGGERS_AFTER_GUARD}")
}

/// Everything version 1 holds, in the order it is created.
pub fn ddl_v1() -> String {
    format!("{TABLES}{}", triggers())
}

/// Writes the identity pragmas, the schema and the `meta` rows into an empty database, in one
/// transaction. `created_by` is the product version that made the file.
pub fn create_v1(
    conn: &mut Connection,
    project_id: &str,
    created_by: &str,
) -> Result<(), super::LedgerError> {
    create_v1_with(conn, project_id, created_by, |_| Ok(()))
}

/// [`create_v1`] with a step in the middle: `populate` runs in the same transaction after the
/// tables and the `meta` rows exist and before the triggers do, so it can insert rows with
/// explicit ids, `rev` and times (a recovery, Issue #3068 P6) without the touch triggers
/// changing them. Its result is returned when the transaction commits; an `Err` rolls the
/// whole build back.
pub fn create_v1_with<T>(
    conn: &mut Connection,
    project_id: &str,
    created_by: &str,
    populate: impl FnOnce(&rusqlite::Transaction<'_>) -> Result<T, super::LedgerError>,
) -> Result<T, super::LedgerError> {
    conn.pragma_update(None, "application_id", APPLICATION_ID)?;
    let tx = conn.transaction()?;
    tx.execute_batch(TABLES)?;
    tx.execute_batch(&format!("PRAGMA user_version = {SCHEMA_VERSION};"))?;
    for (key, value) in [
        ("schema_version", SCHEMA_VERSION.to_string()),
        ("project_id", project_id.to_string()),
        ("created_by", created_by.to_string()),
    ] {
        tx.execute(
            "INSERT INTO meta (key, value) VALUES (?1, ?2)",
            rusqlite::params![key, value],
        )?;
    }
    tx.execute(
        "INSERT INTO meta (key, value) VALUES ('created_at', CAST(unixepoch() AS TEXT))",
        [],
    )?;
    let out = populate(&tx)?;
    tx.execute_batch(&triggers())?;
    tx.commit()?;
    Ok(out)
}
