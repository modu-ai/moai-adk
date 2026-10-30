# SPEC-AGENTIC-FACTORY-LOCAL-001 — 설계 자료 (0.2.0)

> **원천 고정(0.2.0).** 계약 경로 `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`, sha256 `f2c1078b7add847b93926ba3356eda62e4fbc44d8aa77221f059f44f6b059acb`
> (측정 명령 `shasum -a 256 /Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`, 측정 시각 `2026-10-30T05:28:04Z`, `date -u`).
> this SPEC revision tracks the contract as pinned; a different hash means the SPEC must be re-aligned.
> 고정한 판의 `## 1.`~`## 6.` 구간(계약 9~112행)과 아래 §1~§6을 줄 단위로 비교했다: 104줄 대 104줄, 차이 없음(progress.md §G).
> 0.1.0 사본에서 바뀐 것은 없다(0.1.0 사본에도 지금의 I7 문장이 이미 들어 있었다).
>
> **출처 머리말.** 아래 §1~§6은 리더의 구현 계약(leader's implementation contract)
> `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`의 `## 1. Invariants`부터
> `## 6. Defaults`까지를 **바꾸지 않고 옮긴 것**이다(영어 원문 그대로). 이 작성자는 그 파일을 읽었고, 그 파일이
> 가리키는 리더 보고 v2(artifact DjXe5ZKufqx2GWg4wLiwE7)는 읽지 않았다. 계약의 권위는 사용자 결정 D1~D4이다 —
> user decision (the /goal typed in the leader session, per the leader's transcription; original not read by this author),
> `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/user-decision-local-adk-20261030.md`.
> 계약 머리말의 기준 커밋은 `agentic-factory-integration` @ `37b67f6d39`이다. 계약 §7(단계)은 plan.md에 요약했다.
> 0.2.0에서 고정한 판의 §9와 §10(제목부터 파일 끝까지)을 이 문서 끝에 바꾸지 않고 덧붙였다. §8은 아래 범위 줄만 두고 옮기지 않았다.
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
| §8 Zombie factory runs | 없음 — 이 SPEC 밖(옛 factory 경로, lane-20) |
| §9 P4 input | 없음 — P4는 spec §I 범위 밖 |
| §10 Leader decisions | REQ-AFL-005·006·007·008·010·014(반영분), 반영하지 않은 항목은 progress.md §G |

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

## 8. (범위 줄 — 옮기지 않음)

> 고정한 계약 §8 "Zombie factory runs in the `-l` chooser"는 이 하위 SPEC 밖이다: 옛 factory 경로(legacy factory path)의 수리이며 소유는 lane-20이다. 본문을 옮기지 않았다.

## 9. P4 input (leader read, 2026-10-30)

Checkpoint `.moai/state/checkpoint/20261029T031506Z/`: `held-ids.txt` lists 165 held cards with their
original state — 101 `picked`, 64 `queued`; snapshots `backlog.json`, `factory-status.json`.
RESTORE.md's legacy path is `moai todo unhold <id>` (+ `pick` for originally picked). Under the new
design the importer reads the hold rows and `held-ids.txt` directly and creates READY ledger cards
(original `picked` → higher priority band), so the old store is not mutated card by card. Triage
(drop obsolete with a recorded reason, reprioritize) runs on the ledger after import, with Jev and
Codex as advisory input and the leader recording each decision as a `decision` note.

## 10. Leader decisions on the API review (lane-19 18th review, 2026-10-30)

- Claim's lazy reap probes liveness OUTSIDE the transaction, then inside BEGIN IMMEDIATE re-checks
  that epoch and heartbeat_at are unchanged before reclaiming. Reap(nil) returns an error.
- Tree-SHA binding (I8) happens at Enqueue: it requires gate evidence with tree_sha == candidateSHA,
  else ErrEvidenceStale. MERGING is entered only through Enqueue.
- Local-path facts are derived by the store from its own tables, never from caller fields:
  merged <- merge_queue.state = landed for the card; closed/settled <- the DONE move mergeq makes after
  writing the landed note. receipt / receipt-accepted are unused on the local path, so the named
  no-handoff compositions are the normal path with external (queue-table) facts.
- Park reasons `fix-limit` and `reclaim-cap` are added to core's parkable set. nextWake is
  informational only. AddCard refuses missing, CANCELLED or cyclic dependencies.
- QueueUpdate epoch source: waiting/verifying/failed check the card lease epoch; landing/landed check
  the `merge-queue:landing` lock epoch.
- The merge-queue trunk is a dedicated ref checked out in no worktree (default `refs/heads/af/trunk`,
  configurable); mergeq refuses to land when `git worktree list --porcelain` shows it checked out.
- B1 (full CLI run) needs the user's own confirmation inside lane-22's session at P3; the leader's
  reading of the goal does not lift a standing instruction given in that session.
- (lane-22 question, 2026-10-30) `candidateSHA` is a **commit** SHA everywhere (Enqueue, merge_queue.candidate_sha,
  mergeq ancestry). The I8 "tree-SHA binding" means: the gate evidence row records the commit SHA it measured
  (`evidence.subject_sha`), and Enqueue requires `subject_sha == candidateSHA`, else ErrEvidenceStale. The store
  never shells out to git; any amend/rebase yields a new commit SHA and therefore needs fresh gate evidence.

### Leader decisions on lane-19's 19th review (2026-10-30)

- Queue fencing (P2): `merge_queue` records `card_epoch` at Enqueue. waiting/verifying/failed writes carry that
  card epoch and the store requires it to equal `cards.epoch` (a reclaimed card fences its stale lander).
  landing/landed writes carry the landing-lock epoch. Reconcile reverts verifying/landing rows to waiting only
  while it holds the `merge-queue:landing` lock, checked by that lock epoch. (lane-18 store, lane-22 mergeq)
- Facts `closed`/`settled` come from the `merge_queue` row of the current card epoch, never from notes; notes
  are informational only. (lane-18)
- Reap on a card already parked at its third reclaim: count increments, park stays, no error. (lane-18)
- I7 at the DB level: BEFORE DELETE triggers on cards, transitions, evidence, notes that RAISE(ABORT). (lane-18)
- `synchronous=FULL`, not NORMAL: a lost epoch bump after power loss would let two claims share one epoch. (lane-18)
- KeysFor rejects absolute paths, `..`, backslashes (ErrInvalid); a `testdata`/`fixtures` segment maps to its
  owning package key. Add a DSN test with a space in the path. (lane-18)
- mergeq: UpdateRef returns ErrRefMoved only for the old-value mismatch, other git failures surface as-is;
  Trunk must be a full `refs/heads/` name and the default branch is refused; the checked-out check is repeated
  inside the lock; `cmd.WaitDelay` set; ReleaseLock failures are logged/returned. (lane-22)
- Accepted as is: the 400-line review trigger is enforced by the runner, not the store; an ancestor candidate
  settles as landed only behind its I8 evidence at Enqueue.

### P2 retirement plan — Issue/remote code (leader, 2026-10-30)

Importer graph measured on integration cef5e75ac7 (`grep` of `"github.com/modu-ai/moai-adk/internal/<pkg>"`):
forge <- e2e, effect/{ctxledger,docstore,ghexec,ghidentity,ghintake,labels}, mirror, monitor/notify, resident, internal/cli;
mirror <- e2e; ghintake <- effect/{ghexec,gitexec}, resident; ctxledger <- e2e, effect/ghexec, resident;
labels <- e2e, resident; gitref <- effect/docstore.

- Order: wiring first, packages last. lane-20 removes remote wiring from resident, monitor/notify, effect/gitexec,
  effect/docstore; lane-22 removes it from internal/cli and e2e. lane-20 deletes the packages only after both land
  in integration and `go build ./... && go vet ./...` is clean there.
- Candidates for deletion: forge, mirror, gitref, effect/{ghintake,ctxledger,labels,ghexec,ghidentity}, and
  effect/docstore if it is remote-only. Each package's role is read before deletion; any local-only piece moves
  rather than disappears.
- `moai factory`/todo old paths stay until the `moai af` verbs replace them (P2 CLI); nothing outside
  internal/agenticfactory and its CLI wiring is retired in this step.

### Leader decisions on lane-19's review of 56e32db128 (zombie fix, 2026-10-30)

- Automatic card transfer to the single live run is the §8 intent. Comments that say otherwise
  (ReconcileActiveRuns, factorymsg/store.go caller, RebindRun "No row is deleted") are corrected to fact.
- leader_approvals do not follow moved cards (REQ-FCR-023 kept; D3 retires receipts in the new AF). The comment
  says so, and the transfer event names the card ids so the loss of approval is visible.
- Transfer only when exactly one owner is Live and none is Indeterminate; an Indeterminate owner preserves the
  dead runs' cards in place.
- Add the missing tests: card-less dead run with dead worker rows (reap in retireRunUnlessLive); RequireStampBasis
  with a peer/boot-dead source; consolidateIntoLive swallowing only ErrRebindRefused; injected failure at
  rebindPointBeforeCommit rolls back worker deletion and the reaped event. The unreachable non-Dead `continue`
  and the double probe are accepted (safe direction).
- (revision, after lane-20's role read) To avoid a cross-lane wait, lane-20 owns the whole remote retirement in one
  change: `moai factory resident` (internal/cli/factory_resident.go, its test, its registration), resident,
  monitor/notify remote parts (ghpr/Inbox/forge cursors), forge (+webhook), mirror, gitref, effect/{ghintake,
  ctxledger,labels,ghexec,ghidentity,gitexec}, docstore cursors/queue, and e2e files whose only subject is those
  packages. effect itself and docstore.Store retire too unless reconcile or monitor/adapt has a non-remote use;
  ghintake.Admit is not moved (no remote publish remains). Out of lane-20's scope: ledger, mergeq, claim, core,
  runner, mergeq e2e, and every other internal/cli file (lane-22). Gate: integration `go build ./... && go vet ./...`.
- (contract correction, lane-18 80a885ca0b) `cards.epoch` is a column of `cards` (Claim bumps it; reclaim does not).
  `KeysFor(paths) ([]string, error)`. RED commits from bundle 2 on carry a build tag so the default build stays green
  while the commit graph witnesses test-first order; bundle 1's same-commit RED is a recorded deviation.

### Leader decisions on lane-19's 20th review (2026-10-30)

- lane-18: (a) one card's error in reapTx/pickTx no longer fails the whole Claim/Reap — that card is skipped with a
  `stalled` note (once per card+epoch) and the rest commits; (b) a Move into DONE/CANCELLED releases the lease and its
  keys in the same tx, pinned by a test; (c) operator resume out of `reclaim-cap` resets `reclaim_count`; (d) an
  expired lease whose holder is Alive also gets one `stalled` note per card+epoch; (e) a card whose dependency is
  CANCELLED is skipped with one note; (f) AddCard writes the first transition row (from empty, reason `added`);
  same-position rows are documented as non-edge kinds; (g) DSN sets `recursive_triggers=ON`.
- lane-22: (a) Reconcile reverts a `verifying` row only when its `card_epoch` differs from `cards.epoch` or the card
  has no live lease — a live lander's verification is left alone; (b) every git subprocess in mergeq/gitcli runs with
  `LC_ALL=C` (and `LANG=C`) so "but expected" matching is locale-independent, with a test under a non-English locale;
  (c) the trunk also may not equal the target of `refs/remotes/origin/HEAD`. The check-to-update-ref gap is accepted.
- lane-20: the global Indeterminate block stays (conservative, operator bypass exists); fix the first sentence of the
  ReconcileActiveRuns doc in the retirement change.
