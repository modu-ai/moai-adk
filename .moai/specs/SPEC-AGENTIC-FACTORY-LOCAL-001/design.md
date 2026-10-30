# SPEC-AGENTIC-FACTORY-LOCAL-001 — 설계 자료 (0.1.0)

> **출처 머리말.** 아래 §1~§6은 리더의 구현 계약(leader's implementation contract)
> `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`의 `## 1. Invariants`부터
> `## 6. Defaults`까지를 **바꾸지 않고 옮긴 것**이다(영어 원문 그대로). 이 작성자는 그 파일을 읽었고, 그 파일이
> 가리키는 리더 보고 v2(artifact DjXe5ZKufqx2GWg4wLiwE7)는 읽지 않았다. 계약의 권위는 사용자 결정 D1~D4이다 —
> user decision (the /goal typed in the leader session, per the leader's transcription; original not read by this author),
> `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/user-decision-local-adk-20261030.md`.
> 계약 머리말의 기준 커밋은 `agentic-factory-integration` @ `37b67f6d39`이다. 계약 §7(단계)은 plan.md에 요약했다.
> 이 문서가 spec.md와 어긋나면 spec.md의 REQ 문장이 이긴다. 계약이 말하지 않는 곳은 spec.md·상위 OD-33~35에 '미정'으로 적었다.

## REQ 대응

| 계약 절 | REQ |
|---|---|
| §1 I1~I8 | REQ-AFL-001~008 |
| §2 Schema | REQ-AFL-009 |
| §3 API·Claim | REQ-AFL-010 |
| §3 Conflict keys | REQ-AFL-011, REQ-AFL-012 |
| §4 Lane ownership | REQ-AFL-019(lane-20), REQ-AFL-020(lane-21), REQ-AFL-021(lane-20), REQ-AFL-016·017(lane-22) |
| §4 Retirement | REQ-AFL-015 |
| §5 Card flow | REQ-AFL-013, REQ-AFL-014 |
| §6 Defaults | REQ-AFL-018 |

---

## 1. Invariants (every lane)

- I1 One local SQLite file per project is the only authority: `<moaiHome>/db/<projectKey>/af/ledger.db`.
  No GitHub/GitLab Issue reads or writes anywhere in the new path.
- I2 Every claim bumps a fencing epoch; every later write carries `card_id AND epoch` in its WHERE
  clause; a stale epoch changes 0 rows and returns `ErrStaleEpoch`.
- I3 All writes run in `BEGIN IMMEDIATE`, short, no external command inside a transaction, no
  transaction held across an agent turn. busy_timeout 5000 ms. SQLite >= 3.35 checked at Open.
- I4 Time comes from the DB clock (`strftime('%Y-%m-%dT%H:%M:%fZ','now')`), not the caller.
- I5 A live process is never reclaimed automatically. Reclaim only when the PID+start fingerprint
  says Dead. Unknown liveness = treat as alive and write a `stalled` note.
- I6 Stages reuse `internal/agenticfactory/core` (READY, BUILDING/{SPEC,IMPLEMENT,FIX}, REVIEWING,
  MERGING, DONE, CANCELLED, plus park). Every stage write goes through `core.CheckMove` (after lane-18's
  F6 tightening: explicit composite list, park exits single-edge, E20 needs caller evidence).
- I7 cards, transitions, evidence and notes rows are never deleted; lease and lease_keys rows are removed by Release/Reap; cards carry `reclaim_count`; 3 reclaims parks the card (`reclaim-cap`) for the operator.
- I8 Evidence is bound to a tree SHA; a different tree SHA invalidates it.

## 2. Schema (owner: lane-18, package `internal/agenticfactory/ledger`)

```sql
cards(id TEXT PK, title TEXT NOT NULL, stage TEXT NOT NULL, sub_stage TEXT, parked INTEGER NOT NULL DEFAULT 0,
      park_reason TEXT, priority INTEGER NOT NULL DEFAULT 0, spec_json TEXT NOT NULL,
      reclaim_count INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)
transitions(seq INTEGER PK AUTOINCREMENT, card_id TEXT NOT NULL, from_stage TEXT, from_sub TEXT,
      to_stage TEXT NOT NULL, to_sub TEXT, epoch INTEGER NOT NULL, actor TEXT NOT NULL, reason TEXT, at TEXT NOT NULL)
leases(card_id TEXT PK, session_id TEXT NOT NULL, epoch INTEGER NOT NULL, pid INTEGER, pid_start TEXT,
      worktree TEXT NOT NULL, heartbeat_at TEXT NOT NULL, next_wake_at TEXT, expires_at TEXT NOT NULL)
lease_keys(key TEXT PK, card_id TEXT NOT NULL)            -- UNIQUE key = conflict lock
evidence(id INTEGER PK AUTOINCREMENT, card_id TEXT NOT NULL, kind TEXT NOT NULL, tree_sha TEXT NOT NULL,
      base_sha TEXT, command TEXT, result TEXT NOT NULL, path TEXT, epoch INTEGER NOT NULL, at TEXT NOT NULL)
merge_queue(card_id TEXT PK, candidate_sha TEXT NOT NULL, trunk_before TEXT, state TEXT NOT NULL
      CHECK(state IN ('waiting','verifying','landing','landed','failed')), fail_reason TEXT,
      enqueued_at TEXT NOT NULL, updated_at TEXT NOT NULL)
locks(name TEXT PK, session_id TEXT NOT NULL, epoch INTEGER NOT NULL, pid INTEGER, pid_start TEXT,
      expires_at TEXT NOT NULL)                           -- e.g. 'merge-queue:landing'
notes(seq INTEGER PK AUTOINCREMENT, kind TEXT NOT NULL CHECK(kind IN ('landed','decision','gate_wait','stalled')),
      card_id TEXT, text TEXT NOT NULL, at TEXT NOT NULL)
```

`spec_json`: `{"acceptance":[...test names or commands...], "paths":[...], "stop":"...", "class":"A|B|C",
"risk":["security"|"permission"|"data"|...]}`.

## 3. API (lane-18 publishes `ledger/api.go` FIRST as its own commit; others code against it)

```go
type Store struct{ ... }
func Open(ctx context.Context, path string) (*Store, error)
func (s *Store) Close() error
func (s *Store) AddCard(ctx context.Context, c Card) error
func (s *Store) Claim(ctx context.Context, r ClaimRequest) (Claimed, error)      // ErrNoWork
func (s *Store) Heartbeat(ctx context.Context, cardID string, epoch int64, nextWake *time.Time) error
func (s *Store) Move(ctx context.Context, cardID string, epoch int64, to core.Position, actor, reason string, ev ...Evidence) error
func (s *Store) ExpandKeys(ctx context.Context, cardID string, epoch int64, keys []string) error // ErrKeyConflict
func (s *Store) AddEvidence(ctx context.Context, cardID string, epoch int64, e Evidence) error
func (s *Store) Release(ctx context.Context, cardID string, epoch int64) error   // drops lease + keys
func (s *Store) Reap(ctx context.Context, alive func(pid int, start string) Liveness) ([]Reclaimed, error)
func (s *Store) AddNote(ctx context.Context, n Note) error
func (s *Store) NotesSince(ctx context.Context, seq int64) ([]Note, error)
func (s *Store) AcquireLock(ctx context.Context, name string, holder LockHolder, ttl time.Duration) (int64, error)
func (s *Store) ReleaseLock(ctx context.Context, name string, epoch int64) error
func (s *Store) Enqueue(ctx context.Context, cardID string, epoch int64, candidateSHA string) error
func (s *Store) QueueUpdate(ctx context.Context, cardID string, lockEpoch int64, state QueueState, trunkBefore, failReason string) error
func (s *Store) Snapshot(ctx context.Context, dst string) error                  // VACUUM INTO tmp, rename
type Reader interface { Cards(ctx) ([]Card, error); Leases(ctx) ([]Lease, error); Queue(ctx) ([]QueueItem, error); Notes(ctx, sinceSeq int64) ([]Note, error) }
```

Claim is ONE transaction: pick the highest-priority READY, unparked card whose dependencies are DONE and
whose conflict keys are all free; insert lease + keys; bump epoch; move READY -> BUILDING/SPEC (class C)
or BUILDING/IMPLEMENT (A/B); return id + epoch. Lazy reap of Dead holders runs at the start of Claim.

Conflict keys (`KeysFor(paths)`): nearest Go package directory per path; shared keys
`shared:go.mod` (go.mod, go.sum), `shared:templates` (internal/template/templates/**),
`shared:build` (Makefile, .github/**), `shared:rules` (.claude/rules/**). An out-of-scope edit calls
ExpandKeys before editing; a conflict parks nothing — the lane waits on that card only and claims other work.

## 4. Lane ownership (one writer per package)

| Lane | Owns | First deliverable |
|---|---|---|
| lane-17 | SPEC-AGENTIC-FACTORY-001 revision (via manager-spec), rule docs + templates for the new lane loop | SPEC delta: DEC-AF-06/07/17/18 rewritten to local authority; REQ/AC for I1-I8, §2-§6 |
| lane-18 | `internal/agenticfactory/ledger` (+ F6 tightening in core/claim already in flight) | `ledger/api.go` types-only commit, then Open/schema/Claim/Move/Heartbeat/Reap/locks/Snapshot with tests |
| lane-20 | liveness + flow: hook heartbeat writer (throttled 30 s per session, skip DB write when a per-session stamp file is < 30 s old), PID-fingerprint liveness adapter (reuse `claim/presence.go`), flow read model (WIP, age per stage, wait per state, stalled detection) in `internal/agenticfactory/monitor/flow` | flow package against `ledger.Reader`; hook writer after api.go lands |
| lane-21 | `internal/web` board from `ledger.Reader` (cards, leases, keys, queue, notes, flow metrics) | read-only board page + tests with a fake Reader |
| lane-22 | merge queue `internal/agenticfactory/mergeq` (candidate worktree verify, landing lock, CAS `git update-ref <trunk> <new> <old>`, restart reconcile against git ancestry, landed note) + importer `internal/agenticfactory/ledger/importer` (backlog.db queue items, factory.db card rows, claim JSON docs -> ledger, read-only on sources) + CLI `moai af` verbs + integration | mergeq against api.go; importer dry-run mode first |
| lane-19 | read-only review of every commit before it reaches integration | — |

Retirement (after the new path is green, owner lane-22 for wiring, lane-18 for claim/controlref):
forge, mirror, effect/ghintake, effect/ctxledger, effect/labels, gitref, identity GitHub lookups,
intake GitHub search; resident GitHub routes. Remove imports first, delete packages last, in one
reviewed commit per package.

## 5. Card flow (what the CLI verbs drive)

claim -> BUILDING (RED evidence commit: acceptance tests fail) -> GREEN (fix loop max 3, then park
`fix-limit`) -> gates (build, vet, lint, affected + reverse-dependency tests; widen on shared keys)
-> REVIEWING only when spec.risk is non-empty or diff > 400 changed lines (cross-model review on the
diff, correctness scope) -> MERGING via Enqueue -> mergeq verifies the merge result in a candidate
worktree outside the lock, then takes `merge-queue:landing`, re-checks trunk, CAS-updates, records
landed note, Move -> DONE, Release. Failure returns the card to BUILDING/FIX with fail_reason.

## 6. Defaults (named constants, not inline numbers)

heartbeat throttle 30 s; lease 15 min; renew 5 min; reclaim cap 3; fix-loop cap 3; review trigger
400 changed lines; landing lock TTL 10 min; snapshot every 30 min while the store is open.
