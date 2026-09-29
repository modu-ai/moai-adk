-- ac_tst_012_seed.sql — SPEC-TODO-TRANSITION-STAMPS-001 AC-TST-012 fixture.
--
-- A queue database in the PRE-change physical shape (the layout this
-- repository shipped before the transition-stamp columns existed): the four
-- core tables, the landing column on both card-bearing tables, and NO
-- picked_at / dropped_at / archived_at / landing_verdict column anywhere.
-- The golden `list --json` snapshot (ac_tst_012_golden.json) was captured by
-- rendering this fixture through `todo list --json` on the pre-change tree
-- and committed BEFORE the implementation landed (the ordering rule of
-- verification-claim-integrity §2.3).
--
-- The test materializes this file with the raw sqlite driver — NOT through
-- the engine — so the post-change binary exercises its pure-read column
-- gating against a database that genuinely predates the new columns.
-- Timestamps are fixed so the rendered record is deterministic.

CREATE TABLE meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped')),
  landing  TEXT
);
CREATE TABLE findings (
  subject_id TEXT  NOT NULL,
  related_id TEXT  NOT NULL,
  relation   TEXT  NOT NULL,
  source     TEXT  NOT NULL,
  score      REAL  NOT NULL,
  note       TEXT  NOT NULL DEFAULT '',
  at         TEXT  NOT NULL
);
CREATE TABLE archived_items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL,
  position INTEGER NOT NULL,
  landing  TEXT
);
CREATE TABLE archived_findings (
  archive_seq INTEGER NOT NULL,
  position    INTEGER NOT NULL,
  subject_id  TEXT  NOT NULL,
  related_id  TEXT  NOT NULL,
  relation    TEXT  NOT NULL,
  source      TEXT  NOT NULL,
  score       REAL  NOT NULL,
  note        TEXT  NOT NULL DEFAULT '',
  at          TEXT  NOT NULL
);
CREATE INDEX idx_items_state ON items(state);

INSERT INTO meta(key, value) VALUES ('schema_version', '1');
INSERT INTO meta(key, value) VALUES ('last_seq', '4');

INSERT INTO items(seq, id, text, added_at, spec_id, state, landing) VALUES
  (1, 't1', 'queued card',        '2026-09-28T09:00:00Z', NULL,                        'queued',  NULL),
  (2, 't2', 'picked card',        '2026-09-28T09:05:00Z', 'SPEC-EXAMPLE-001',          'picked',  NULL),
  (3, 't3', '[DROPPED — stale] dropped card', '2026-09-28T09:10:00Z', NULL, 'dropped', NULL);

INSERT INTO archived_items(seq, id, text, added_at, spec_id, state, position, landing) VALUES
  (1, 't4', 'archived card', '2026-09-28T09:00:00Z', NULL, 'picked',
   0,
   '{"ref":"origin/develop","ref_head":"abc1234567890abcdef1234567890abcdef1234","observed_at":"2026-09-28T10:00:00Z","sha":"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef","sha_source":"operator"}');
