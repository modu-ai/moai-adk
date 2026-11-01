# SPEC-AGENTIC-FACTORY-LOCAL-001 — 설계 자료 (0.6.0)

> 이전 문구(머리, 0.3.0): «# SPEC-AGENTIC-FACTORY-LOCAL-001 — 설계 자료 (0.3.0)». 0.6.0은 문서 끝에 "0.6.0 부록"을 더했을 뿐 아래 계약 사본(§1~§6, §9, §10)은 한 글자도 바꾸지 않았다. 감사받지 않은 판이다.

> **원천 고정(0.2.0).** 계약 경로 `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`, sha256 `f2c1078b7add847b93926ba3356eda62e4fbc44d8aa77221f059f44f6b059acb`
> (측정 명령 `shasum -a 256 /Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`, 측정 시각 `2026-10-30T05:28:04Z`, `date -u`).
> this SPEC revision tracks the contract as pinned; a different hash means the SPEC must be re-aligned.
> 고정한 판의 `## 1.`~`## 6.` 구간(계약 9~112행)과 아래 §1~§6을 줄 단위로 비교했다: 104줄 대 104줄, 차이 없음(progress.md §G).
> 0.1.0 사본에서 바뀐 것은 없다(0.1.0 사본에도 지금의 I7 문장이 이미 들어 있었다).
>
> **읽기 메모(0.3.0, OD-40 — leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible)).** 아래 계약 사본은 한 글자도 바꾸지 않는다. 그 사본의 I8과 §2 `evidence.tree_sha`는
> §10이 대신한다: 결속 대상은 `evidence.subject_sha`(증거가 잰 commit SHA)다. §2 표는 `cards.epoch`, `merge_queue.card_epoch`,
> `evidence.subject_sha`를 포함한 것으로 읽는다. QueueUpdate epoch와 closed/settled 원천은 §10 19th review 문장을 따른다. spec.md §A.5와 같다.
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
| 0.6.0 부록(계약 사본 밖, 착지 코드) | REQ-AFL-001(저장소 묶음)·007(reclaim-cap 출구)·008(Enqueue 거절)·009(evidence_mark, 스키마 2판)·014(trunk 고정, gate-logs, merged 사실, OnTrunk, DONE의 trunk 확인) |

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

---

## 0.6.0 부록 — 착지 코드가 하는 일 (계약 사본 밖)

> 이 부록은 위 계약 사본의 일부가 아니다. 원천은 감사 N19와 리더 지시("SPEC 개정: gate-logs·FailureLog·evidence_mark, REQ-AFL-014 트렁크 고정, 저장소 묶음을 반영")이며, 아래 모든 문장은 이 작성자가 `WT-agentic-factory-integration` HEAD `12eaa6c23f`에서 읽은 파일과 줄에 기댄다. 시험은 돌리지 않았고, 감사도 돌리지 않았다. 의무는 spec.md의 REQ 문장이 정하며, 이 부록은 그 REQ가 가리키는 동작을 설명할 뿐이다. 이 문서와 spec.md가 어긋나면 spec.md가 이긴다(머리말 그대로).

### A. `transitions.evidence_mark` — 두 수열의 순서를 원장 안에서 정함 (REQ-AFL-008·009·014)

- **무엇인가.** `evidence`와 `transitions`는 따로 세는 수열이다. `evidence_mark`는 transition이 쓰인 순간 그 카드의 가장 새 evidence 행 id다. 그래서 "어떤 증거가 어떤 이동 뒤에 기록되었는가"를 시계가 아니라 원장 안의 수로 정한다(`internal/agenticfactory/ledger/schema.go:41-49` 주석과 열).
- **쓰는 곳.** 새 transition은 그 순간 카드의 최대 evidence id를 mark로 쓴다(`claim.go:228-238`). Move가 함께 싣는 증거는 transition보다 먼저 들어가 mark 앞에 놓인다(같은 곳). mark가 없는 행(이전 판의 행)은 카드의 모든 증거를 더 오래된 것으로 센다(`queue.go:241-243` `offerMarkSQL`, `:292-293`).
- **이전(판 1 → 2).** `Open`은 `PRAGMA user_version`이 코드가 아는 가장 새 판(3, 저장소 이동을 기록한 원장의 판 — 부록 D)보다 크면 거절하고(`schema.go:147-149`), 2보다 작으면 한 쓰기 트랜잭션 안에서 열을 더하고 채운 뒤 판을 2로 쓴다(`:153-155`, `:161-177`). (2026-11-01 정정: 이전 문구는 "코드의 판(2)보다 크면 거절"이었고 줄 번호는 @`0f108104c1`에서 다시 읽었다.) 채우는 값은 그 transition의 시각 이전에 기록된 카드의 가장 새 evidence id이며, 같은 밀리초의 증거는 MERGING으로 가는 이동에서는 뒤로, 다른 이동에서는 앞으로 센다(`:155-177`, 규칙은 `:173`). 까닭(코드 주석 `:158-162`): BUILDING으로 되돌아간 이동에서는 같은 밀리초 증거를 앞으로 세는 쪽이 DONE을 막는 쪽이고, 제출(MERGING) 이동에서는 같은 밀리초의 green을 뒤로 세어야 더 새 작업이 더 오래된 것으로 보이지 않는다.
- **제출(offer)이란.** 후보가 큐에 처음 나온 transition이다: MERGING으로 가며 사유가 `enqueued <sha>`인 첫 행(`queue.go:238-239`, `:245-257`); 그런 행이 없으면 MERGING으로 들어간 가장 늦은 행(`:259-281`). 같은 커밋을 다시 내도 첫 제출의 mark는 움직이지 않는다(`:263-264` 주석).
- **mark를 쓰는 판단.**
  - 가장 새 작업(`newestWorkIs`, `queue.go:86-111`): 카드의 가장 새 green이 착지한 커밋에 대한 것이거나, 그 green id가 첫 제출의 mark 이하일 때만 착지한 커밋이 아직 카드의 가장 새 작업이다. green이 없으면 반박할 작업이 없다.
  - 착지가 서 있음(`landingStands`, `:113-155`): (1) 후보에 대한 green·gate·review 각각의 가장 새 기록이 실패가 아니다(`noFailingNewestRun`, `:207-221`); (2) 제출 뒤 BUILDING으로 되돌아갔다면 가장 늦은 그런 이동의 mark보다 큰 id의 통과 green·gate가 있다(`reworkMark`, `:283-302`; BUILDING 안의 이동은 되돌아감이 아님); (3) 제출 뒤 다른 커밋에 대한 실패가 있었거나, 제출 뒤 MERGING을 떠나 다른 커밋의 증거를 남겼다면, 그 증거보다 큰 id의 통과 green과 통과 gate가 후보에 있다(`measuredAfterOtherWork`, `:157-205`; MERGING을 떠나지 않은 동안 다른 커밋의 통과 기록은 아무 의무도 만들지 않는다).
  - 현재 epoch 측정(`measuredInEpoch`, `:223-236`): 그 카드 epoch로 기록된 통과 green·gate가 하나 이상 있고 가장 새 실행이 실패가 아니다. 이전 epoch의 실행은 이전 보유자의 것이다.
- **누가 그 판단을 쓰나.** merged·closed·settled 사실을 원장 표에서 끌어내는 `factsTx`(`queue.go:31-84`)와 `Enqueue`(`:329-507`: 최신 증거 `:441-459`, 실패한 최신 실행 `:460-466`, 재제출 `:467-483`, 이전 epoch에 착지한 행 `:379-405`·`:484-488`).
- **이어받기와 OnTrunk.** 이전 카드 epoch에 landed가 된 행을 현재 보유자가 같은 커밋으로 다시 낼 때, 저장소는 호출자가 `OnTrunk(true)`를 넘긴 경우에만 행을 landed로 두고 현재 epoch로 옮긴다; 아니면 행은 `waiting`으로 다시 시작한다(`queue.go:319-325`, `:379-405`, `:484-488`). 원장은 git을 부르지 않으므로(REQ-AFL-008) trunk 질문은 동사가 한다(아래 C).

### B. 실패 기록 — `FailureLog`와 `gate-logs` (REQ-AFL-014 0.6.0 (ii))

- **무엇이 원장에 남나.** 관문·병합·다른 git 단계가 실패하면 원장·쪽지·실패 사유·동사 출력에는 실패 종류(`gate`·`merge`·`git`), 관문 id, 원인(종료 상태나 시간 초과), 출력 바이트 수, 로그 파일 경로(쓰지 못했으면 "logging failed")만 남는다(`internal/agenticfactory/mergeq/mergeq.go:90-101` `GateError`는 관문 id만, `:127-144` `failReason`, `:204-214` 종류 `git`, `:283` 종류 `merge`, `:293`·`:297` 종류 `gate`). git 출력은 오류 문자열에 들어가지 않는다(`gitcli.go:73-83` `GitError`, `:118-128` `MergeError`).
- **관문 id.** `gate-` + 관문 글의 SHA-256 앞 12 hex(`internal/cli/af_land.go:119-122` `afGateID`). 동사는 관문 글과 id를 병합 큐에 넘긴다(`af_land.go:226`).
- **어디에.** 원장 파일 옆 디렉터리 `gate-logs`(`af_land.go:58-61`; 파일당 상한은 기본값, 합계 상한은 비워 두어 기본값).
- **디렉터리.** Lstat로 보아 링크이거나 디렉터리가 아니면 거절하고, ACL을 지운 뒤 0700으로 둔다(`faillog.go:171-190` `prepareDir`).
- **파일.** 임시 파일(`.tmp-` 접두, `faillog.go:56`)을 만들어 0600으로 바꾸고, ACL을 지운 뒤에 쓰고, 로그 이름(`.log`, `:53`)으로 rename한다(`:103-164`). 이름은 카드·후보·종류마다 하나이며 다음 실패가 덮어쓴다(`:284-302`).
- **가림과 상한.** 글은 `core.Redact` → 제어 문자·escape 제거(`StripControls`, `:272-279`) → `core.Redact` → `core.Redact` 순으로 지난다(`:69-74`). 파일당 출력은 마지막 256 KiB(`:19`), 관문 글 머리는 처음 2 KiB(`:59`).
- **디렉터리 합계.** 로그와 살아 있는 임시 파일을 합쳐 1 MiB(`:50`, `4 * DefaultFailureLogCap`). 넘으면 가장 오래된 `*.log`부터 지우고 방금 쓴 파일은 지우지 않는다; 1분이 지난 임시 파일은 지우고(`:195`), 새 임시 파일은 남겨 합계에 넣는다(`:202-262` `prune`).
- **ACL.** macOS는 `/bin/chmod -N`을 절대 경로로 부른다(`faillog_acl_darwin.go:14-19`); 다른 플랫폼은 아무것도 하지 않는다(`faillog_acl_other.go:8`). 지우기에 실패하면 출력을 버리고 사유가 그렇게 말한다(REQ-AFL-014 (ii); 시험 이름은 acceptance AC-AFL-014 (f)).

### C. trunk 고정과 trunk 질문 (REQ-AFL-014 0.6.0 (i)(iv)(v))

- 병합 큐 라이브러리는 바꿀 수 있는 trunk 칸을 갖는다(`mergeq.go:168` `Merger.Trunk`, 기본 `mergeq.go:40` `DefaultTrunk = "refs/heads/af/trunk"`). 신뢰 경계는 동사다: `moai af land`·`reconcile`은 `--trunk`(`af_land.go:45`)가 전용 ref가 아니면 아무것도 착지하기 전에 거절한다(`internal/cli/af.go:229-238` `afCheckTrunk`, 호출 `af_land.go:294`). 이로써 상위 OD-35 (l)이 착지 코드로 닫혔다(spec §G).
- `enqueue` 동사는 후보가 trunk의 조상인지 저장소에 물어 `OnTrunk`로 넘기고, 저장소가 답하지 못하면 아무것도 바꾸지 않고 거절한다(`af.go:569-583`, `:179-185` `afInTrunk`).
- `move`가 DONE으로 옮길 때 큐 행이 landed면, 동사는 그 커밋이 trunk에 있는지 묻고 없거나 답이 없으면 거절한다(`af.go:420-424`, `:187-211` `afCheckLandedInTrunk`).
- 읽지 않은 것: 계약 §10 20th lane-22 (c)("the trunk also may not equal the target of `refs/remotes/origin/HEAD`")를 코드가 검사하는지는 확인하지 않았다(progress §O).

### D. 저장소 묶음 (REQ-AFL-001 0.6.0)

- **원장 쪽.** 한 행짜리 `repository` 표(`schema.go:71-77`)에 첫 호출자가 정체를 쓰고, 현재 정체와 같은 값은 변화 없음, 현재 정체가 아닌 값은 `ErrWrongRepository`(`repository.go:13`, `:23-40`). 현재 정체는 가장 새 이동 기록의 새 정체이고, 이동이 없으면 첫 묶음이다. 묶이기 전 `Repository`는 ""를, 그 뒤에는 현재 정체를 돌려준다(`:42-47`). 갱신·삭제는 트리거가 거절한다. 표는 `CREATE TABLE IF NOT EXISTS`로 Open마다 생기며 판을 올리지 않는다(`schema.go:71-73` 주석).
- **정체.** symlink를 푼 절대 git common dir(`af.go:245-268` `afGitCommonPath`). 그래서 linked worktree는 주 저장소와 같은 정체다.
- **동사 쪽.** `afRun`이 `move-repository`를 뺀 모든 동사에서 묶는다(`af.go:73-93`, `:103-125`). 묶이기 전에는 원장이 이미 아는 저장소 — lease worktree들의 저장소, `--db`로 준 원장 파일이 (symlink를 푼 자리로) 놓인 디렉터리의 저장소, 행이 이름 대는 커밋을 가진 저장소 — 와 먼저 비교하고, 어긋나면 둘을 이름 대어 거절한다(`:184-266`; 기본 위치의 원장 파일 자리는 단서가 아니고, 같은 저장소의 lease worktree도 단서가 아니며 — lease worktree는 호출자가 적어 넣은 말일 뿐이다 — 다른 저장소의 lease worktree와 상대 경로 lease worktree는 거절하고, 단서가 없는 `--db` 원장은 lease worktree가 있어도 `--bind-repo`를 요구한다). 행을 읽지 못하거나, 커밋 질문에 git이 답하지 못하거나 시간이 다하거나 물은 수와 다른 줄 수·다른 모양으로 답하면 거절이며, 단서나 통과로 읽지 않는다(`:241-251`, `:322-`). 정체가 원장의 것과 같으면 표지를 비교한다: 원장에 표지가 있으면 저장소 git 디렉터리의 `moai-af-repository-id`가 같아야 하고(다르거나 없거나 읽을 수 없으면 거절), 없으면 그 저장소의 표지를 받아 기록한다(없으면 새로 만든다; `:132-160`, `af_repository_marker.go`). 커밋을 찾거나 착지하는 동사는 쓰는 저장소(`--repo`, 없으면 카드의 lease worktree, 없으면 프로젝트 디렉터리)도 묶는다(`af.go:413-`; land·reconcile은 `af_land.go:335-357`). `--repo`는 프로젝트 디렉터리와 git common dir가 같은 경로만 받는다(`af.go:541-`, `os.SameFile`). 줄 번호는 2026-11-01 정정 때 @`0f108104c1`에서 다시 읽었다. (2026-11-02 개정) 이 문단은 개정 전 계약의 코드를 적는다: REQ-AFL-001은 이제 원장에 아무것도 쓰지 않고 `--bind-repo` 없이 끝난 동사가 묶이지 않은 원장을 묶지 않기를 요구하며, 첫 묶음은 `--bind-repo`로 밝힌 호출이나 원장에 쓰고 끝난 동사만 한다. `31a7c1099a`의 코드는 아직 그런 동사(회수할 것이 없는 `reap`)도 묶으며(수리 보고 N55 절), 그 차이는 후속 카드 몫이다.
- **예외.** 프로젝트 디렉터리가 git 저장소가 아니면, git을 부르지 않는 동사는 아무것도 묶지 않고 진행한다(`af.go:86-90`).
- **저장소 이동(N52, 코드 사실).** 저장소를 디스크에서 옮기면 정체가 바뀌어 `move-repository`를 부르기 전까지 다른 모든 동사가 거절된다. 원장은 추가만 되는 `repository_move` 표(갱신·삭제 트리거, `schema.go:78-87`)에 옛 정체·새 정체·사유를 쓰고(`Store.MoveRepository`, `Store.RepositoryMoves`; `repository_move.go`), 첫 묶음 행은 건드리지 않는다. 동사 `moai af move-repository`(`internal/cli/af_move_repository.go`)는 `afRun`을 거치지 않고 원장을 직접 열어(`--db` 존중) 묶지도 비교하지도 않으며, 현재 정체를 읽어 확인한 뒤에만 `MoveRepositoryMarked`를 부른다(2026-11-01 정정): 옛 git 디렉터리가 디스크에 없고, `--reason`이 비지 않았고, 새 저장소(`--bind-repo`, 없으면 프로젝트 디렉터리)가 원장이 기록한 표지를 가졌다(가진 커밋과 상관없이; 커밋은 `commits_checked`·`commits_unchecked`로 세기만 함, `afMoveProof` `:212-270`). 표지가 없는 원장은 대체 규칙을 따른다: 실패하지 않은 큐 행의 후보 커밋을 새 저장소가 모두 가졌다고 git이 답해야 하고(답하지 못하면 거절), `--bind-repo`로 저장소를 밝혀야 하며(프로젝트 디렉터리는 밝힘으로 치지 않음), `--db` 원장 파일이 다른 저장소 안에 있지 않아야 하고, 그 뒤 새 저장소의 표지를 이동과 함께 기록한다. 이 대체 규칙은 밝힘이지 증명이 아니다: 실패하지 않은 큐 행이 없는 표지 없는 원장은 `--bind-repo`로 밝힌 어느 저장소든 받는다. 첫 이동 기록은 같은 트랜잭션에서 `PRAGMA user_version`을 3으로 올려 이동을 모르는 빌드가 그 파일을 거절하게 한다(`repository_move.go:79`). `--from <옛 프로젝트 디렉터리>`(`--db`와 함께 쓰면 거절)는 기본 위치 원장(그리고 `-wal`·`-shm`)을 옛 키에서 새 키로 옮긴 뒤 이동을 기록하고, 기록하지 못하면 되돌리며, 새 키에 원장이 이미 있거나 옛 키에 없으면 아무것도 건드리지 않고 거절한다(`:94-124`, `:162-183`). 묶이지 않은 원장은 옮길 저장소가 없어 거절한다. 어느 거절도 원장을 바꾸지 않는다. 이동 뒤 다른 동사는 현재 정체와 비교한다. lease가 적어 둔 worktree 경로는 다시 쓰지 않으므로, 이동 전에 claim한 카드의 evidence·enqueue는 `--repo <새 저장소 안의 경로>`가 필요하다.
- **열린 위험(이 부록이 닫지 않음).** 원장이 아무것도 모르는 기본 위치 원장은 (2026-11-02 개정 계약) 원장에 쓰고 끝난 첫 동사나 `--bind-repo`로 밝힌 첫 호출이 묶음을 정한다 — 코드는 아직 아무것도 쓰지 않고 끝난 첫 호출도 묶는다(후속 카드)(`--db` 원장은 `--bind-repo`를 요구); 저장소를 디스크에서 옮기면 `move-repository`를 부르기 전까지 뒤의 호출이 거절된다; 훅 heartbeat 기록기(`internal/hook/ledger_heartbeat.go`)는 저장소를 확인하지 않는다. (2026-11-01 정정) 표지 없는 원장의 `move-repository` 대체 규칙은 운영자의 밝힘에 기댄다; `--from` 뒤에도 세션이 claim 때 적어 둔 heartbeat 원장 경로는 옛 경로를 가리킨다(수리 보고 Gaps; 바꾸지 않음, 시험 없음). progress §O.

### E. reclaim-cap park의 출구 (REQ-AFL-007 0.6.0)

- reclaim-cap park에서 park 없이 나가는 이동은 같은 트랜잭션에서 `reclaim_count`를 0으로 한다(`ledger/move.go:126-130`). DONE·CANCELLED 이동은 lease와 키를 지운다(`:118-125`).
- `moai af resume`은 `--guard reclaim-cap`만(`af.go:468-470`, 상수 `core/move.go:24`), 비지 않은 `--reason`만 받고(`:471-473`), reclaim-cap park가 아닌 카드는 거절하며(`:483-485`), 같은 단계·park 없음으로 `resume-guard:` 증거를 실어 `Move`한다(`:486-490`). 동사 정의 전체는 `af.go:446-506`.
