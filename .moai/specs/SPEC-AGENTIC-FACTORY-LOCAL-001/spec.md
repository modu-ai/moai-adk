---
id: SPEC-AGENTIC-FACTORY-LOCAL-001
title: "AgenticFactory 로컬 원장 — Issue 없는 단일 SQLite 권위·claim·병합 큐 하위 계약"
version: "0.4.0"
status: draft
created: 2026-10-30
updated: 2026-10-30
author: GOOS
priority: P0
phase: "v3.3.0 target"
module: "internal/agenticfactory"
lifecycle: spec-anchored
tags: "agentic-factory, local-ledger, sqlite, fencing-epoch, conflict-keys, merge-queue, importer, retirement"
tier: L
issue_number: 1948
related_specs: [SPEC-AGENTIC-FACTORY-001, SPEC-AGENTIC-FACTORY-IDENTITY-001]
---

# SPEC-AGENTIC-FACTORY-LOCAL-001 — 로컬 원장 하위 계약

## HISTORY

| 버전 | 날짜 | 작성 | 내용 |
|---|---|---|---|
| 0.1.0 | 2026-10-30 | manager-spec | 새로 만든 하위 SPEC(draft, 감사받지 않음). 상위 SPEC-AGENTIC-FACTORY-001은 Tier L이고 REQ 25개·AC 25개로 두 상한에 이미 닿아 있어(`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier: 상한을 넘는 것은 분할 신호), 로컬 원장 요구는 상위에 더하지 않고 이 하위 SPEC에 둔다. REQ를 먼저 세었다: 불변식 I1~I8 8개 + 계약 §2 스키마 1 + §3 API·Claim 1 + 충돌 키 1 + 충돌 키 잠금(D4) 1 + §5 카드 흐름 1 + 병합 큐(D3) 1 + §4 퇴역 1 + importer 1 + `moai af` 1 + §6 기본값 1 + 흐름 읽기 모형 1 + 보드 1 + 훅 heartbeat 기록기 1 = **21개**. Tier M 상한 16을 넘으므로 **Tier L**(상한 25)로 정했고, design.md·research.md를 둔다. AC도 21개이며 AC마다 시험 이름이 정확히 하나다. 원천: 사용자 결정 D1~D4와 리더의 구현 계약(§A.1). 계약이 말하지 않는 곳은 '미정'으로 적고 상위 OD-33~35에 올렸다. 원격 게시 주기(D5)는 정하지 않았다(상위 OD-31). |
| 0.2.0 | 2026-10-30 | manager-spec | revised, unaudited. plan-audit FAIL 0.62(`.moai/reports/issue-1948/plan-audit-local-001-c445ce9559/verdict.md`, audited_sha `c445ce9559`)의 막는 결함 가운데 이 SPEC 몫인 D1·D8·D9·D10을 고쳤고, 원천 계약을 sha256으로 고정했다(§A.4). 고정한 계약 §10(리더 결정)이 0.1.0 문장과 어긋나는 곳은 계약 문장으로 바꿨다: REQ-AFL-005·006·007·008·010·014·020, AC-AFL-004·005·007·008·020. 0.1.0 문장은 §C.6과 각 AC의 '이전 문구'에 남겼다. REQ 21개·AC 21개 그대로이고 새 id는 없다. 0.1.1이 아니라 0.2.0으로 올린 까닭: 오타 정정이 아니라 REQ의 뜻이 바뀌었다(I8 결속 대상이 commit SHA로, 지우는 행의 범위, Claim의 reap 순서). 고친 곳 대응표는 progress.md §G, 새 질문은 상위 OD-36~42. 감사받지 않았고 PASS라고 적지 않는다. |
| 0.3.0 | 2026-10-30 | manager-spec | revised, unaudited. 상위 OD-36~42를 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible)대로 닫았다(원천: 리더가 직접 쓴 결정 파일 `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/leader-decisions-od36-42-20261030.md`. 사용자·운영자 결정이 아니다; `hold-20261030.md`의 "Hold lifted" 절에 적힌 사용자 답은 리더가 옮겨 적은 것이며 원문은 읽지 않았다). 바꾼 곳: REQ-AFL-005(표지), REQ-AFL-007(Move into DONE/CANCELLED 문장, 표지), REQ-AFL-008(메모, 묶음 이름), REQ-AFL-009(§2 읽기), REQ-AFL-014(표지), REQ-AFL-015(§10 P2로 맞춤), REQ-AFL-020(키 원천), AC-AFL-005·007·015·020, 읽기 메모 §A.5. REQ 21개·AC 21개 그대로이고 새 id는 없다. 감사는 FAIL 0.62 @c445ce9559 하나뿐이며 PASS라고 적지 않는다. 대응표는 progress.md §J. 원천 계약 sha256은 그대로다(§A.4). |
| 0.4.0 | 2026-10-30 | manager-spec | revised, unaudited. plan-audit FAIL 0.72(`.moai/reports/issue-1948/plan-audit-local-001-081eacf037/verdict.md`, @081eacf037)를 고쳤다. 막는 결함: (N1) AC-AFL-010에 Claim lazy reap 사례 L-Live·L-Unknown·L-Dead·L-Changed와 변이(Live 소유자 회수)를 더함; (N2) AC-AFL-002를 여섯 메서드 사례표로 넓히고 변이(`WHERE card_id=?`만 쓰는 Release)를 적음; (N3) plan.md M6과 상위 DEC-AF-30을 REQ-AFL-015·OD-42(계약 §10 P2 revision)에 맞춤. SHOULD-FIX: (N4) 상위 §J 집계 다시 셈; (N5) AC-AFL-004가 패키지 수준 문자열 상수를 접어 읽게 함(착지한 `nowSQL` 상수가 통과하도록); (N6) AC-AFL-007 사례 (iii); (N8) AC-AFL-015 디렉터리 부재·resident; (N13) acceptance §F 'AC ↔ landed test' 표를 새로 두고 각 AC의 RED-now·green path를 지금 사실로 고침 — 코드가 이미 착지한 19개 AC는 `verification-completeness.md` §2.1에 따라 regression-guard; AC-AFL-021은 착지 시험이 AC를 완전히 덮어 이름을 `TestLedgerHeartbeat_ThrottlesToTheStampFile`로 채택(이전 이름은 '이전 문구'); (N14) AC-AFL-014 사례 (b) 잠금 epoch, (c) checkout 거절. 코드 사실은 integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직임)과 lane-18 커밋 `607bd17804`(integration 밖)에서 읽었다. N7·N9·N10·N11·N12는 이번에 고치지 않았다. REQ 21개·AC 21개 그대로, 새 id 없음, REQ 본문은 바꾸지 않음. 원천 계약 sha256 그대로(§A.4). 대응표는 progress.md §K. PASS라고 적지 않는다. |

## §A 범위와 관계

### A.1 규범 원천

- 사용자 결정 D1~D4 — user decision (the /goal typed in the leader session, per the leader's transcription; original not read by this author). 기록 파일: `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/user-decision-local-adk-20261030.md`. 이 작성자는 그 파일만 읽었다. 목표 문장(파일에 옮겨 적힌 것): "로컬 전용 ADK 재설계 — Issue 없이, 병목 없이 -> 설계 대로 진행 해주고, …". 표의 처분:
  - D1 로컬 전용: AgenticFactory의 Issue·원격 코드(forge, mirror, ghintake, ctxledger, labels, gitref) 퇴역 — Adopted
  - D2 상위 DEC-AF-06/07/17/18을 로컬 원장 권위로 개정 — Adopted
  - D3 카드 완료는 병합 큐 증거로; 리더 승인 receipt 퇴역 — Adopted
  - D4 전역 직렬 slot을 충돌 키 잠금으로 대체 — Adopted
  - D5 원격 게시 주기(push/PR/release) — Not decided. 이 SPEC은 정하지 않으며, 아무것도 게시하지 않는다(상위 OD-31).
- 리더의 구현 계약(leader's implementation contract): `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`. 불변식 I1~I8, §2 스키마, §3 API·Claim·충돌 키, §4 lane 소유, §5 카드 흐름, §6 기본값, §7 단계, §8 옛 factory 경로의 좀비 run, §9 P4 입력, §10 리더 결정. 그 §1~§6과 §9·§10을 이 SPEC의 `design.md`로 옮겼다(출처 머리말 포함). §8은 이 SPEC 밖(옛 factory 경로, 소유 lane-20)이라 옮기지 않았다. 고정한 판은 §A.4. (이전 문구(0.1.0): "그 §1~§6을 이 SPEC의 `design.md`로 옮겼다(출처 머리말 포함).")
- 같은 파일의 "Leader's derived scope (provisional, reversible)"(배포 = 로컬 설치, held 카드 복구 등)는 리더의 잠정 해석이며 이 SPEC의 요구가 아니다.

### A.2 상위 SPEC과의 관계

- 상위: SPEC-AGENTIC-FACTORY-001(0.7.0). 상위의 Issue·원격 의존 REQ/AC의 처분은 상위 §J 표에 있고, 대체 REQ가 이 SPEC에 있을 때 그 표의 "child id" 칸이 이 SPEC의 REQ-AFL id를 가리킨다.
- 상위 DEC-AF-06/07/17/18은 0.7.0에서 로컬 원장 권위로 다시 썼다(옛 문장은 그 행에 '이전 결정(superseded)'으로 남음). D1~D4는 상위 DEC-AF-30~33이다.
- 프론트매터에는 `related_specs`만 적었다. `depends_on`은 쓰지 않는다.

### A.3 기준 커밋

코드 사실은 통합 끝 `37b67f6d39`에서 읽었다(`git log -1 --format='%h %s' 37b67f6d39` → "37b67f6d39 merge(lane-18): F6 transitions outside the table are refused (5b164ccd14, includes 338c7c33fb) (Issue #1948)").

### A.4 Source pin (원천 고정)

- 계약 경로: `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`(git이 추적하지 않는 파일).
- sha256: `f2c1078b7add847b93926ba3356eda62e4fbc44d8aa77221f059f44f6b059acb` — 측정 명령 `shasum -a 256 /Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`, 측정 시각 `2026-10-30T05:28:04Z`(`date -u`).
- this SPEC revision tracks the contract as pinned; a different hash means the SPEC must be re-aligned.
- 감사 때(@`c445ce9559`) 이 파일의 sha256은 `2d11dba271d59915b093903fa0f701559bb6ba7a7702131d562f896e3fa6c1cd`였다(감사 verdict Evidence 4). 지금 판과 다르다. 옛 판의 사본이 없어 그 사이 무엇이 바뀌었는지는 보지 못했다(§H).

### A.5 계약 읽기 메모 (OD-40, leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible))

- 고정한 계약 사본(`design.md` §1~§6, §9·§10)은 바꾸지 않는다. 이 메모가 그 사본을 읽는 법이다.
- 증거의 결속 대상은 `evidence.subject_sha`(증거가 잰 commit SHA)이며, I8과 §2의 tree SHA 문언은 §10이 대신한다.
- §2 표는 `cards.epoch`, `merge_queue.card_epoch`, `evidence.subject_sha` 열을 포함한 것으로 읽는다(계약 §10 contract correction·19th review·lane-22 질문 항목).
- QueueUpdate의 epoch 원천과 closed/settled 사실의 원천은 §10의 19th review 문장이 18th review 문장보다 앞선다.

## §B 이 트리에서 읽은 사실 (`@37b67f6d39`)

- `git ls-tree -d --name-only 37b67f6d39 internal/agenticfactory/ internal/agenticfactory/monitor/ internal/web/`: `ledger`, `mergeq`, `monitor/flow` 디렉터리는 없다. 있는 것 가운데 이 SPEC과 닿는 것: `claim`, `core`, `effect`(그 아래 `ctxledger`, `ghintake`, `labels` 등), `forge`, `gitref`, `identity`, `intake`, `mirror`, `resident`, `monitor/adapt`, `monitor/notify`, `internal/web/assets`, `internal/web/testdata`.
- `core.CheckMove`: `internal/agenticfactory/core/move.go:28 @37b67f6d39`(`func CheckMove(from, to Position) error`). `claim/claim.go:478`이 이미 그것을 부른다.
- `internal/agenticfactory/claim/presence.go`는 있다(계약 §4가 활성 판정 어댑터에서 재사용하라고 적은 파일).
- AC가 이름 붙인 시험 21개는 하나도 없다(acceptance.md §B).

## §C 요구사항 (GEARS)

REQ 21개(Tier L 상한 25). 각 REQ의 소유 lane은 design §4(계약 §4)를 따르고 AC에 적었다.

### C.1 불변식 (계약 §1)

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AFL-001 | I1 단일 로컬 권위 | The AgenticFactory local path SHALL keep exactly one SQLite file per project, at `<moaiHome>/db/<projectKey>/af/ledger.db`, as the only authority for its card state, and no package on that path SHALL read or write a GitHub or GitLab Issue. |
| REQ-AFL-002 | I2 fencing epoch | When a claim succeeds, the ledger SHALL increment that card's fencing epoch, and every later write for the card SHALL carry both `card_id` and `epoch` in its WHERE clause, so that a write carrying a stale epoch changes zero rows and returns `ErrStaleEpoch`. |
| REQ-AFL-003 | I3 짧은 트랜잭션 | The ledger SHALL run every write inside `BEGIN IMMEDIATE`, SHALL run no external command inside a transaction, SHALL hold no transaction across an agent turn, and SHALL set busy_timeout to 5000 ms; when `Open` finds a SQLite version older than 3.35, it SHALL refuse to open the store. |
| REQ-AFL-004 | I4 DB 시계 | The ledger SHALL take the time of every recorded event (`created_at`, `updated_at`, `at`, `heartbeat_at`, `enqueued_at`) from the database clock `strftime('%Y-%m-%dT%H:%M:%fZ','now')` and SHALL NOT record a caller-supplied time as the time of an event. |
| REQ-AFL-005 | I5 살아 있는 소유자 보호 | While the liveness of a lease holder, judged from its recorded PID and process-start fingerprint by an adapter that reuses `internal/agenticfactory/claim/presence.go`, is not `Dead`, the ledger SHALL NOT reclaim that lease; when the liveness is unknown, the ledger SHALL treat the holder as alive and SHALL write a `stalled` note, and when the holder of an expired lease is alive, the ledger SHALL write one `stalled` note per card and epoch (계약 §10 20th review lane-18 (d); OD-41 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible)). |
| REQ-AFL-006 | I6 core 단계 재사용 | The ledger SHALL represent stages only with the positions of `internal/agenticfactory/core` (READY, BUILDING/{SPEC,IMPLEMENT,FIX}, REVIEWING, MERGING, DONE, CANCELLED, and park), SHALL apply a stage write only when `core.CheckMove` accepts the move, and SHALL use the park reasons `fix-limit` and `reclaim-cap` as members of core's parkable set (계약 §10). |
| REQ-AFL-007 | I7 지우지 않음·재회수 상한 | The ledger SHALL never delete a row of `cards`, `transitions`, `evidence`, `notes`, or `merge_queue`, SHALL remove `leases` and `lease_keys` rows only through `Release` or `Reap`, SHALL count reclaims of each card in `reclaim_count`, and when a card's third reclaim is recorded, it SHALL park that card for the operator with park reason `reclaim-cap`; when a card is moved into DONE or CANCELLED, the ledger SHALL remove its lease and lease keys by calling the `Release` path inside the same transaction as the move. (leader decision OD-36 — leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible). 고정한 I7 목록은 `cards`·`transitions`·`evidence`·`notes`이고, `merge_queue`와 "only"는 조정자가 더한 것이며 리더가 그대로 두었다. DONE/CANCELLED Move 문장의 원천은 계약 §10 20th review lane-18 (b).) |
| REQ-AFL-008 | I8 증거의 commit SHA 결속 | When `Enqueue` is called for a card, the ledger SHALL accept it only when that card has gate evidence whose `subject_sha` — the commit SHA the evidence measured — equals `candidateSHA`, SHALL otherwise return `ErrEvidenceStale` and leave the card where it is, SHALL let a card enter MERGING only through `Enqueue`, and SHALL NOT shell out to git to decide this (계약 §10 lane-22 질문 항목; OD-40 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): `evidence.subject_sha`가 권위이고 I8·§2의 옛 문언은 §10이 대신한다, §A.5. 메모 — OD-37 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): 상위 REQ-AF-007의 원격 무관 의무(같은 HEAD 리뷰, 빚 없는 독립 감사, QA PASS)는 카드가 `Enqueue` 전에 갖춰야 할 관문 증거가 싣는 절차 의무로 남는다; 이 메모는 위 SHALL 문장에 새 의무를 더하지 않으며, 그 의무를 증거 종류로 모형화할지는 배포 뒤 후속 카드가 정한다.) |

### C.2 원장 구조와 API (계약 §2·§3)

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AFL-009 | §2 스키마 | The ledger schema SHALL consist of the tables `cards`, `transitions`, `leases`, `lease_keys`, `evidence`, `merge_queue`, `locks`, and `notes` with the columns of `design.md` §2, SHALL make `lease_keys.key` the primary key, SHALL restrict `merge_queue.state` to `waiting`, `verifying`, `landing`, `landed`, `failed` and `notes.kind` to `landed`, `decision`, `gate_wait`, `stalled` by CHECK constraints, and SHALL store `cards.spec_json` in the shape stated in `design.md` §2. (OD-40 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): `design.md` §2는 `cards.epoch`·`merge_queue.card_epoch`·`evidence.subject_sha`를 포함한 것으로 읽는다, §A.5.) |
| REQ-AFL-010 | §3 API·Claim | The `Store` SHALL expose the operations of `design.md` §3; when `Claim` runs, it SHALL first probe the liveness of expired lease holders outside any transaction, then, inside one `BEGIN IMMEDIATE` transaction, reclaim a lease judged Dead only when its epoch and `heartbeat_at` are unchanged since the probe, pick the highest-priority READY, unparked card whose dependencies are DONE and whose conflict keys are all free, insert the lease and its keys, bump the epoch, move the card to BUILDING/SPEC for class C or BUILDING/IMPLEMENT for class A or B, and return the card id and epoch, and it SHALL return `ErrNoWork` when no card qualifies; when `Reap` is called with a nil liveness function, it SHALL return an error; `Snapshot` SHALL write its copy with `VACUUM INTO` a temporary file and then rename that file to the destination (계약 §10). |
| REQ-AFL-011 | §3 충돌 키 | `KeysFor` SHALL map each path to the key of its nearest Go package directory and a path matching a shared pattern to that shared key — `shared:go.mod` (go.mod, go.sum), `shared:templates` (internal/template/templates/**), `shared:build` (Makefile, .github/**), `shared:rules` (.claude/rules/**); when a lane is about to edit outside its held keys, it SHALL call `ExpandKeys` before the edit, and when a requested key is held by another card, `ExpandKeys` SHALL return `ErrKeyConflict` and SHALL park no card. |
| REQ-AFL-012 | D4 충돌 키 잠금 | Where two cards hold disjoint conflict keys, AgenticFactory SHALL let both be claimed and worked on at the same time and SHALL NOT order card work through a global serial slot; when `ExpandKeys` returns `ErrKeyConflict`, the lane SHALL wait on that card only and SHALL remain free to claim other work. |

### C.3 카드 흐름과 병합 큐 (계약 §5, D3)

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AFL-013 | §5 카드 흐름 | A claimed card SHALL move through BUILDING with a RED evidence record showing its acceptance tests failing, then a GREEN fix loop of at most 3 iterations after which the card is parked with reason `fix-limit`, then the gates (build, vet, lint, affected and reverse-dependency tests, widened when shared keys are held), then REVIEWING only when `spec.risk` is non-empty or the diff exceeds 400 changed lines, then MERGING through `Enqueue`; when the merge queue reports a failure, the card SHALL return to BUILDING/FIX with that `fail_reason`. |
| REQ-AFL-014 | D3 병합 큐 완료 | When a card is in MERGING, the merge queue SHALL verify the merge result in a candidate worktree outside the landing lock, then take the lock `merge-queue:landing`, re-check trunk, update trunk with a compare-and-set `git update-ref <trunk> <new> <old>`, record a `landed` note, move the card to DONE, and release it; the trunk SHALL be a dedicated ref checked out in no worktree (default `refs/heads/af/trunk`, configurable — 설정 수단은 미정, OD-35 (l)), and the merge queue SHALL refuse to land while `git worktree list --porcelain` shows the trunk checked out; `QueueUpdate` writes to `waiting`, `verifying`, or `failed` SHALL carry the card epoch recorded at `Enqueue`, which SHALL equal `cards.epoch`, and writes to `landing` or `landed` SHALL carry the `merge-queue:landing` lock epoch; the store SHALL derive the facts merged, closed, and settled from its own tables — merged from `merge_queue.state = landed` for the card, closed and settled from the `merge_queue` row of the current card epoch — and never from caller-supplied fields or notes; no leader approval receipt SHALL be required for card completion, and after a restart the merge queue SHALL reconcile its rows against git ancestry (계약 §10; OD-40 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): QueueUpdate epoch와 closed/settled 원천은 §10 19th review 문장을 따른다, §A.5). |

### C.4 퇴역·이관·명령·기본값 (계약 §4·§6, D1)

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AFL-015 | §4 Issue·원격 코드 퇴역 | When the new path is green (판정 기준 미정, OD-35 (e)), the Issue and remote code SHALL be retired in one change covering `moai factory resident` (`internal/cli/factory_resident.go`, its test, and its registration), `resident`, the remote parts of `monitor/notify` (ghpr, Inbox, and forge cursors), `forge` (with its webhook), `mirror`, `gitref`, `effect/ghintake`, `effect/ctxledger`, `effect/labels`, `effect/ghexec`, `effect/ghidentity`, `effect/gitexec`, the `docstore` cursors and queue, and the e2e files whose only subject is those packages; `effect` itself and `docstore.Store` SHALL be retired too unless reconcile or `monitor/adapt` has a non-remote use of them; remote wiring SHALL be removed before packages are deleted, each package's role SHALL be read before its deletion, and any local-only piece SHALL be moved rather than deleted; the change SHALL leave integration `go build ./... && go vet ./...` clean, and `ledger`, `mergeq`, `claim`, `core`, `runner`, the mergeq e2e, and every other `internal/cli` file SHALL stay outside it. (OD-42 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): 고정한 계약 §10 "P2 retirement plan"의 revision 항목 "lane-20 owns the whole remote retirement in one change"를 따른다. §4가 적고 §10이 적지 않은 `identity`의 GitHub 조회와 `intake`의 GitHub 검색이 이 변경에 드는지는 미정이다.) |
| REQ-AFL-016 | importer | The importer in `internal/agenticfactory/ledger/importer` SHALL bring backlog.db queue items, factory.db card rows, and claim JSON documents into the ledger, SHALL open every source read-only and change none of them, and SHALL provide a dry-run mode as its first deliverable (dry-run의 원장 쓰기 여부는 미정, OD-35). |
| REQ-AFL-017 | `moai af` 명령 | The `moai af` command SHALL provide verbs from which every step of the REQ-AFL-013 card flow (claim, stage move, evidence record, `Enqueue`, `Release`) is reachable; the verb names and their grammar are 미정 (OD-33). |
| REQ-AFL-018 | §6 기본값 | Each of the values heartbeat throttle 30 s, lease 15 min, renew 5 min, reclaim cap 3, fix-loop cap 3, review trigger 400 changed lines, landing lock TTL 10 min, and snapshot interval 30 min while the store is open SHALL be a named constant, and no code on the local path SHALL write any of these values as an inline number. |

### C.5 관찰과 heartbeat (계약 §4 lane-20·lane-21)

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AFL-019 | 흐름 읽기 모형 | The flow read model in `internal/agenticfactory/monitor/flow` SHALL compute WIP, age per stage, wait per state, and stalled detection taking its input only from `ledger.Reader`. |
| REQ-AFL-020 | 보드 | The board in `internal/web` SHALL show cards, leases, queue, notes, and flow metrics read from `ledger.Reader`, SHALL show lease keys read from the `Keys []string` field of each `Lease` that `ledger.Reader` returns, SHALL hold no reference to the ledger's writable `Store` type, and SHALL perform no write. (OD-39 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): `ledger.Reader`에는 따로 키 접근자가 없다.) |
| REQ-AFL-021 | 훅 heartbeat 기록기 | The hook heartbeat writer SHALL write a session's heartbeat at most once per 30 s, and when that session's stamp file is younger than 30 s, it SHALL skip the database write. |

### C.6 이전 문구 (0.2.0·0.3.0에서 바뀐 REQ)

0.2.0·0.3.0에서 바꾼 REQ의 이전 문장이다. 지금 효력은 위 표의 문장에 있다. 바꾼 까닭과 원천은 progress.md §G 대응표에 있다.

- 0.1.0 문장, `REQ-AFL-005`: «While the liveness of a lease holder, judged from its recorded PID and process-start fingerprint by an adapter that reuses `internal/agenticfactory/claim/presence.go`, is not `Dead`, the ledger SHALL NOT reclaim that lease; when the liveness is unknown, the ledger SHALL treat the holder as alive and SHALL write a `stalled` note.»
- 0.1.0 문장, `REQ-AFL-006`: «The ledger SHALL represent stages only with the positions of `internal/agenticfactory/core` (READY, BUILDING/{SPEC,IMPLEMENT,FIX}, REVIEWING, MERGING, DONE, CANCELLED, and park) and SHALL apply a stage write only when `core.CheckMove` accepts the move.»
- 0.1.0 문장, `REQ-AFL-007`: «The ledger SHALL never delete a row, SHALL count reclaims of each card in `reclaim_count`, and when a card's third reclaim is recorded, it SHALL park that card for the operator.»
- 0.1.0 문장, `REQ-AFL-008`: «The ledger SHALL bind every evidence row to a tree SHA, and evidence whose tree SHA differs from the tree being judged SHALL NOT count.»
- 0.1.0 문장, `REQ-AFL-010`: «The `Store` SHALL expose the operations of `design.md` §3; when `Claim` runs, it SHALL, in one transaction, first reap Dead holders, then pick the highest-priority READY, unparked card whose dependencies are DONE and whose conflict keys are all free, insert the lease and its keys, bump the epoch, move the card to BUILDING/SPEC for class C or BUILDING/IMPLEMENT for class A or B, and return the card id and epoch, and it SHALL return `ErrNoWork` when no card qualifies; `Snapshot` SHALL write its copy with `VACUUM INTO` a temporary file and then rename that file to the destination.»
- 0.1.0 문장, `REQ-AFL-014`: «When a card is in MERGING, the merge queue SHALL verify the merge result in a candidate worktree outside the landing lock, then take the lock `merge-queue:landing`, re-check trunk, update trunk with a compare-and-set `git update-ref <trunk> <new> <old>`, record a `landed` note, move the card to DONE, and release it; card completion SHALL be judged from this merge-queue evidence, no leader approval receipt SHALL be required for it, and after a restart the merge queue SHALL reconcile its rows against git ancestry.»
- 0.1.0 문장, `REQ-AFL-020`: «The board in `internal/web` SHALL show cards, leases, keys, queue, notes, and flow metrics read from `ledger.Reader` and SHALL perform no write.»
- 0.1.0 문장, `REQ-AFL-009`(0.3.0에서 바뀜): «The ledger schema SHALL consist of the tables `cards`, `transitions`, `leases`, `lease_keys`, `evidence`, `merge_queue`, `locks`, and `notes` with the columns of `design.md` §2, SHALL make `lease_keys.key` the primary key, SHALL restrict `merge_queue.state` to `waiting`, `verifying`, `landing`, `landed`, `failed` and `notes.kind` to `landed`, `decision`, `gate_wait`, `stalled` by CHECK constraints, and SHALL store `cards.spec_json` in the shape stated in `design.md` §2.»
- 0.1.0 문장, `REQ-AFL-015`(0.3.0에서 바뀜): «When the new path is green (판정 기준 미정, OD-35), the code paths `forge`, `mirror`, `effect/ghintake`, `effect/ctxledger`, `effect/labels`, `gitref`, the GitHub lookups of `identity`, the GitHub search of `intake`, and the GitHub routes of `resident` SHALL be retired by removing every import of a package first and deleting the package last, in one reviewed commit per package.»
- 0.2.0 문장, `REQ-AFL-007`: «The ledger SHALL never delete a row of `cards`, `transitions`, `evidence`, `notes`, or `merge_queue`, SHALL remove `leases` and `lease_keys` rows only through `Release` or `Reap`, SHALL count reclaims of each card in `reclaim_count`, and when a card's third reclaim is recorded, it SHALL park that card for the operator with park reason `reclaim-cap`. (author's reconciliation of a contract conflict (I7 vs §3 Release), reversible — OD-36. 고정한 I7 목록은 `cards`·`transitions`·`evidence`·`notes`이며 `merge_queue`는 그 목록에 없다: `merge_queue`와 "only"는 조정자가 더한 것이다.)»
- 0.2.0 문장, `REQ-AFL-008`: «When `Enqueue` is called for a card, the ledger SHALL accept it only when that card has gate evidence whose `subject_sha` — the commit SHA the evidence measured — equals `candidateSHA`, SHALL otherwise return `ErrEvidenceStale` and leave the card where it is, SHALL let a card enter MERGING only through `Enqueue`, and SHALL NOT shell out to git to decide this (계약 §10 lane-22 질문 항목; I8·§2의 `tree_sha` 문언과의 차이는 OD-40).» (묶음 칸: "I8 증거의 tree SHA 결속")
- 0.2.0 문장, `REQ-AFL-020`: «The board in `internal/web` SHALL show cards, leases, queue, notes, and flow metrics read from `ledger.Reader`, SHALL show lease keys (출처 미정 — 계약 §3의 `Reader`에 keys 접근자가 없다, OD-39), SHALL hold no reference to the ledger's writable `Store` type, and SHALL perform no write.»
- REQ-AFL-005·014는 0.3.0에서 괄호 안 표지만 바뀌었다("OD-41", "§10 안의 이른 문장과 늦은 문장의 차이는 OD-40" → leader decision 표기).

## §D 수용 기준 요약

`acceptance.md`에 21개(AC-AFL-001~021). AC-AFL-0NN의 주 REQ는 REQ-AFL-0NN이다. AC마다 시험 이름(Go 시험 함수 이름)은 정확히 하나이며, 21개 모두 `37b67f6d39`에 없다. 실행 명령은 두지 않는다 — 아무 시험도 고르지 않는 명령은 통과처럼 보이기 때문이다(`.claude/rules/moai/development/verification-completeness.md` §1.1).

## §E 추적표 (REQ → AC)

| REQ | 주 AC |
|---|---|
| REQ-AFL-001 | AC-AFL-001 |
| REQ-AFL-002 | AC-AFL-002 |
| REQ-AFL-003 | AC-AFL-003 |
| REQ-AFL-004 | AC-AFL-004 |
| REQ-AFL-005 | AC-AFL-005 |
| REQ-AFL-006 | AC-AFL-006 |
| REQ-AFL-007 | AC-AFL-007 |
| REQ-AFL-008 | AC-AFL-008 |
| REQ-AFL-009 | AC-AFL-009 |
| REQ-AFL-010 | AC-AFL-010 |
| REQ-AFL-011 | AC-AFL-011 |
| REQ-AFL-012 | AC-AFL-012 |
| REQ-AFL-013 | AC-AFL-013 |
| REQ-AFL-014 | AC-AFL-014 |
| REQ-AFL-015 | AC-AFL-015 |
| REQ-AFL-016 | AC-AFL-016 |
| REQ-AFL-017 | AC-AFL-017 |
| REQ-AFL-018 | AC-AFL-018 |
| REQ-AFL-019 | AC-AFL-019 |
| REQ-AFL-020 | AC-AFL-020 |
| REQ-AFL-021 | AC-AFL-021 |

## §F 결정 (상위 DEC 행을 가리킴)

- 상위 DEC-AF-30(D1), DEC-AF-31(D2), DEC-AF-32(D3), DEC-AF-33(D4). 출처 표기는 모두 user decision (the /goal typed in the leader session, per the leader's transcription; original not read by this author).
- 이 SPEC은 그 밖에 아무것도 정하지 않는다. 계약이 말하지 않는 곳은 '미정'이다.

## §G 미결 질문 (상위 OD 행)

| 상위 OD | 대상 | 이 SPEC에서 닿는 곳 |
|---|---|---|
| OD-31 | 원격 게시 주기(push·PR·release), D5 | 전체 — 이 SPEC은 아무것도 게시하지 않는다 |
| OD-33 | `moai af` 동사 이름과 문법 | REQ-AFL-017, AC-AFL-017 |
| OD-34 | Claim의 "의존 카드가 DONE"에서 의존 관계의 원천(`spec_json`과 `cards`에 의존 필드가 없음) | REQ-AFL-010 |
| OD-35 | 계약이 말하지 않는 세부 열두 가지(상위 OD-35 행에 나열; 0.2.0에서 (a)는 계약이 답한 것으로 고쳤고 (l)을 더했다) | REQ-AFL-003·004·010·011·013·014·015·016·018·021, AC-AFL-003·004·013·015·017·018·021 |
| OD-36 | **닫힘(2026-10-30): decided by leader, reversible** — leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): only `Release` and `Reap` delete `leases`/`lease_keys` rows; a Move into DONE/CANCELLED calls the `Release` path inside the same transaction and is not a third path; `cards`, `transitions`, `evidence`, `notes`, `merge_queue` rows are never deleted; lane-17's reconciliation stands. 열려 있던 때의 문언(history): REQ-AFL-007의 지우는 행 범위 — 고정한 I7 목록과 조정자 문언(`merge_queue`·"only")의 차이, 계약 §10 20th (b) "a Move into DONE/CANCELLED releases the lease and its keys"가 `Release`인지 셋째 경로인지 | REQ-AFL-007, AC-AFL-007 |
| OD-37 | **닫힘(2026-10-30): decided by leader, reversible** — leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): no new REQ in LOCAL-001; the remote-independent obligations of REQ-AF-007 (same-HEAD review, debt-free independent audit, QA PASS) stay as process obligations carried by the gate evidence a card needs before `Enqueue` (REQ-AFL-008); a follow-up card after deploy decides whether to model them as evidence kinds. 열려 있던 때의 문언(history): 상위 REQ-AF-007의 원격 무관 의무(같은 HEAD 코드·보안 리뷰 PASS, 빚 없는 독립 감사 PASS, QA PASS) — 이 SPEC에 대신하는 REQ가 아직 없다 | 이 SPEC에 닿는 REQ 없음(no local counterpart yet) |
| OD-38 | **닫힘(2026-10-30): decided by leader, reversible** — leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): IDENTITY-001 is not revised in this run; its online-control-ref dependence is recorded as superseded for the local path in §J; a follow-up card after deploy revises it. 열려 있던 때의 문언(history): SPEC-AGENTIC-FACTORY-IDENTITY-001이 D1·D2 아래 따로 개정되어야 하는지(REQ-AFI-001이 online control ref를 권위로 적음) | REQ-AFL-001, REQ-AFL-015(identity GitHub lookups) |
| OD-39 | **닫힘(2026-10-30): decided by leader, reversible** — leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): `Lease` carries `Keys []string`; `ledger.Reader` needs no separate key accessor; the board reads keys from `Lease`. 열려 있던 때의 문언(history): `ledger.Reader`가 lease 키를 내주는지, `Lease`가 키를 싣는지 | REQ-AFL-020, AC-AFL-020 |
| OD-40 | **닫힘(2026-10-30): decided by leader, reversible** — leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): `evidence.subject_sha` (the commit SHA measured) is authoritative; `tree_sha` wording in I8/§2 is superseded by §10; the §2 table is read as including `cards.epoch`, `merge_queue.card_epoch`, `evidence.subject_sha`; for QueueUpdate epoch and closed/settled sources the later §10 sentence (19th) wins over the 18th. 열려 있던 때의 문언(history): 계약 안의 문언 차이: I8·§2 `tree_sha`와 §10의 `evidence.subject_sha`(commit SHA), §10 18th review 문장과 19th review 문장(QueueUpdate epoch 원천, closed/settled 원천), §2 표에 없는 열(`cards.epoch`, `merge_queue.card_epoch`, `evidence.subject_sha`) | REQ-AFL-008·009·014, AC-AFL-008 |
| OD-41 | **닫힘(2026-10-30): decided by leader, reversible** — leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): §10 20th (d) wins: an overdue lease whose holder is Live gets exactly one `stalled` note per card+epoch; the earlier "no stalled note" wording is superseded. 열려 있던 때의 문언(history): 기한 지난 lease의 소유자가 살아 있을 때 `stalled` 쪽지 — 조정자 지시 "Live: no `stalled` note"와 계약 §10 20th (d) "also gets one `stalled` note per card+epoch"의 차이 | REQ-AFL-005, AC-AFL-005 |
| OD-42 | **닫힘(2026-10-30): decided by leader, reversible** — leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible): §10 P2 retirement plan wins: lane-20 owns the whole retirement in one change, with the wider target set; REQ-AFL-015/AC-AFL-015 are aligned to §10 P2. 열려 있던 때의 문언(history): 퇴역의 소유 lane과 대상 패키지 집합 — 계약 §4 Retirement(이 SPEC의 REQ-AFL-015 원천)와 §10 P2 retirement plan(lane-20이 한 변경으로 전부, 대상 더 넓음)의 차이. 이 판은 REQ-AFL-015를 바꾸지 않았다 | REQ-AFL-015, AC-AFL-015 |

## §H 확인 공백

- 사용자 결정 원문은 읽지 않았다(리더의 옮겨 적음만 읽음). 리더 보고 v2(artifact)는 읽지 않았다.
- 시험은 하나도 돌리지 않았다. 이름 21개가 없다는 것은 한 번의 `git grep` 결과다(acceptance.md §B).
- REQ-AFL-003의 "트랜잭션 안에서 외부 명령 없음, 에이전트 턴을 넘는 트랜잭션 없음"은 AC-AFL-003의 시험 하나로 덮이지 않는다. 그 부분의 확인 방법은 미정이다(OD-35 (j); 계약 §4에 적힌 검토 경로는 lane-19의 읽기 전용 검토뿐이다).
- §6 상수 가운데 원장 패키지 밖에 둘 것(heartbeat 30 s, fix-loop 3, review 400, landing TTL 10 min)의 패키지는 미정이므로 AC-AFL-018의 시험은 원장 패키지 상수만 덮는다(OD-35 (d)).

- (0.2.0) 고정한 계약(§A.4)과 감사 때 계약은 sha256이 다르다. 옛 판 사본이 없어 바뀐 내용을 보지 못했다. 감사가 본 `design.md` §1~§6 차이(빈 줄 하나)는 지금 판에서 다시 재지 않았고, 지금 판과의 비교는 progress.md §G에 있다.
- (0.2.0) 계약 §10 가운데 이 SPEC의 REQ 문장에 넣지 않은 항목과 그 까닭은 progress.md §G "반영하지 않은 §10 항목"에 있다.
- (0.2.0) 시험·빌드는 하나도 돌리지 않았다. 시험 이름 21개 부재는 이번에 다시 돌린 `git grep` 한 번의 결과다(acceptance.md §B).

- (0.3.0) OD-36~42의 결정은 리더가 쓴 결정 파일 하나에서 읽었다. 그 결정이 사용자 답(hold 파일의 "Hold lifted", 리더의 옮겨 적음)과 어떻게 이어지는지는 원문을 읽지 않아 확인하지 않았다.

## §I 범위 밖

### Out of Scope — 원격 게시

- push, PR, 태그, 공개 릴리스. D5는 정해지지 않았고(상위 OD-31) 이 SPEC은 아무것도 게시하지 않는다.

### Out of Scope — 배포와 held 카드 처리

- `make build VERSION=v3.2.0-rc.39`, 백업, 로컬 설치, held 카드 165장의 복구·가져오기·분류(계약 §7 P3·P4). 이것은 리더의 잠정 해석에 든 단계이며 이 SPEC의 요구가 아니다.

### Out of Scope — 규칙 문서와 template

- lane 루프의 규칙 문서·template 수정(계약 §4 lane-17 둘째 산출물). 이 SPEC은 규칙 파일을 고치지 않는다.

### Out of Scope — 옛 원장 직접 수정

- 옛 원장(backlog.db, factory.db)을 SQL로 직접 고치는 것. importer는 읽기만 한다(REQ-AFL-016).
