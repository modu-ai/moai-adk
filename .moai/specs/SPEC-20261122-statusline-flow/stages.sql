-- Stage of every open card under the widened rule of SPEC-20261122-statusline-flow (D-5, D-6).
-- Read-only: sqlite3 -readonly ~/.moai/projects/moai-adk/ledger.db < stages.sql
-- GLOB, not LIKE: the marker words are case-sensitive prefixes, as starts_with is in the renderer.
WITH heads AS (
  SELECT n.card_id, n.note_id,
         CASE WHEN instr(substr(n.body,1,1000), char(10))>0
              THEN substr(n.body,1,instr(substr(n.body,1,1000),char(10))-1)
              ELSE substr(n.body,1,1000) END AS first
  FROM card_note n JOIN card c ON c.card_id=n.card_id WHERE c.state IN ('active','review')
), marks AS (
  SELECT card_id, note_id, CASE
    WHEN first GLOB 'LANE-BLOCKED*' THEN 'B'
    WHEN first GLOB 'AUDIT*' AND instr(first,'verdict: PASS')>0 THEN 'A'
    WHEN first GLOB 'AUDIT*' AND instr(first,'verdict: FAIL')>0 THEN 'W'
    WHEN first GLOB 'LANE-READY-INTEGRATION*' OR first GLOB 'MERGE-READY*' OR first GLOB 'READY *'
      OR first GLOB '감사 요청*'
      OR (first GLOB 'LANE-DONE*' AND (instr(first,'audit requested')>0 OR instr(first,'감사 요청')>0)) THEN 'S'
    WHEN first GLOB 'ACK *' OR first GLOB 'PLAN *' THEN 'W'
  END AS mk FROM heads
), newest AS (SELECT card_id, max(note_id) AS nid FROM marks WHERE mk IS NOT NULL GROUP BY card_id),
landed AS (SELECT DISTINCT card_id FROM event WHERE kind='card.landed'),
res AS (SELECT card_id, max(CASE WHEN status IN ('held','held_for_review') AND lease_until >= unixepoch() THEN 1 ELSE 0 END) AS live
        FROM reservation WHERE status IN ('held','held_for_review','expired_candidate') GROUP BY card_id),
st AS (
  SELECT c.card_id, c.state,
    CASE
      WHEN c.state='draft' THEN 'D' WHEN c.state='ready' THEN 'R' WHEN c.state='hold' THEN 'H'
      WHEN c.card_id IN (SELECT card_id FROM landed) THEN 'L'
      WHEN m.mk='B' THEN 'B' WHEN m.mk='A' THEN 'A' WHEN m.mk='S' THEN 'S'
      WHEN c.state='review' THEN 'S'
      WHEN coalesce(res.live,0)=1 THEN 'W' ELSE 'E' END AS stage
  FROM card c LEFT JOIN newest ON newest.card_id=c.card_id LEFT JOIN marks m ON m.note_id=newest.nid
       LEFT JOIN res ON res.card_id=c.card_id
  WHERE c.state <> 'done')
SELECT stage, count(*) FROM st GROUP BY stage ORDER BY stage;
SELECT 'done_today_utc', count(*) FROM card WHERE state='done' AND closed_at >= (unixepoch()/86400)*86400;
SELECT 'done_total', count(*) FROM card WHERE state='done';
SELECT 'open_total', count(*) FROM card WHERE state<>'done';
