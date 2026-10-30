---
id: SPEC-AGENTIC-FACTORY-LOCAL-001
title: "AgenticFactory 로컬 원장 — Issue 없는 단일 SQLite 권위·claim·병합 큐 하위 계약"
version: "0.1.0"
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

## §A 범위와 관계

### A.1 규범 원천

- 사용자 결정 D1~D4 — user decision (the /goal typed in the leader session, per the leader's transcription; original not read by this author). 기록 파일: `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/user-decision-local-adk-20261030.md`. 이 작성자는 그 파일만 읽었다. 목표 문장(파일에 옮겨 적힌 것): "로컬 전용 ADK 재설계 — Issue 없이, 병목 없이 -> 설계 대로 진행 해주고, …". 표의 처분:
  - D1 로컬 전용: AgenticFactory의 Issue·원격 코드(forge, mirror, ghintake, ctxledger, labels, gitref) 퇴역 — Adopted
  - D2 상위 DEC-AF-06/07/17/18을 로컬 원장 권위로 개정 — Adopted
  - D3 카드 완료는 병합 큐 증거로; 리더 승인 receipt 퇴역 — Adopted
  - D4 전역 직렬 slot을 충돌 키 잠금으로 대체 — Adopted
  - D5 원격 게시 주기(push/PR/release) — Not decided. 이 SPEC은 정하지 않으며, 아무것도 게시하지 않는다(상위 OD-31).
- 리더의 구현 계약(leader's implementation contract): `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`. 불변식 I1~I8, §2 스키마, §3 API·Claim·충돌 키, §4 lane 소유, §5 카드 흐름, §6 기본값, §7 단계. 그 §1~§6을 이 SPEC의 `design.md`로 옮겼다(출처 머리말 포함).
- 같은 파일의 "Leader's derived scope (provisional, reversible)"(배포 = 로컬 설치, held 카드 복구 등)는 리더의 잠정 해석이며 이 SPEC의 요구가 아니다.

### A.2 상위 SPEC과의 관계

- 상위: SPEC-AGENTIC-FACTORY-001(0.7.0). 상위의 Issue·원격 의존 REQ/AC의 처분은 상위 §J 표에 있고, 대체 REQ가 이 SPEC에 있을 때 그 표의 "child id" 칸이 이 SPEC의 REQ-AFL id를 가리킨다.
- 상위 DEC-AF-06/07/17/18은 0.7.0에서 로컬 원장 권위로 다시 썼다(옛 문장은 그 행에 '이전 결정(superseded)'으로 남음). D1~D4는 상위 DEC-AF-30~33이다.
- 프론트매터에는 `related_specs`만 적었다. `depends_on`은 쓰지 않는다.

### A.3 기준 커밋

코드 사실은 통합 끝 `37b67f6d39`에서 읽었다(`git log -1 --format='%h %s' 37b67f6d39` → "37b67f6d39 merge(lane-18): F6 transitions outside the table are refused (5b164ccd14, includes 338c7c33fb) (Issue #1948)").

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
| REQ-AFL-005 | I5 살아 있는 소유자 보호 | While the liveness of a lease holder, judged from its recorded PID and process-start fingerprint by an adapter that reuses `internal/agenticfactory/claim/presence.go`, is not `Dead`, the ledger SHALL NOT reclaim that lease; when the liveness is unknown, the ledger SHALL treat the holder as alive and SHALL write a `stalled` note. |
| REQ-AFL-006 | I6 core 단계 재사용 | The ledger SHALL represent stages only with the positions of `internal/agenticfactory/core` (READY, BUILDING/{SPEC,IMPLEMENT,FIX}, REVIEWING, MERGING, DONE, CANCELLED, and park) and SHALL apply a stage write only when `core.CheckMove` accepts the move. |
| REQ-AFL-007 | I7 지우지 않음·재회수 상한 | The ledger SHALL never delete a row, SHALL count reclaims of each card in `reclaim_count`, and when a card's third reclaim is recorded, it SHALL park that card for the operator. |
| REQ-AFL-008 | I8 증거의 tree SHA 결속 | The ledger SHALL bind every evidence row to a tree SHA, and evidence whose tree SHA differs from the tree being judged SHALL NOT count. |

### C.2 원장 구조와 API (계약 §2·§3)

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AFL-009 | §2 스키마 | The ledger schema SHALL consist of the tables `cards`, `transitions`, `leases`, `lease_keys`, `evidence`, `merge_queue`, `locks`, and `notes` with the columns of `design.md` §2, SHALL make `lease_keys.key` the primary key, SHALL restrict `merge_queue.state` to `waiting`, `verifying`, `landing`, `landed`, `failed` and `notes.kind` to `landed`, `decision`, `gate_wait`, `stalled` by CHECK constraints, and SHALL store `cards.spec_json` in the shape stated in `design.md` §2. |
| REQ-AFL-010 | §3 API·Claim | The `Store` SHALL expose the operations of `design.md` §3; when `Claim` runs, it SHALL, in one transaction, first reap Dead holders, then pick the highest-priority READY, unparked card whose dependencies are DONE and whose conflict keys are all free, insert the lease and its keys, bump the epoch, move the card to BUILDING/SPEC for class C or BUILDING/IMPLEMENT for class A or B, and return the card id and epoch, and it SHALL return `ErrNoWork` when no card qualifies; `Snapshot` SHALL write its copy with `VACUUM INTO` a temporary file and then rename that file to the destination. |
| REQ-AFL-011 | §3 충돌 키 | `KeysFor` SHALL map each path to the key of its nearest Go package directory and a path matching a shared pattern to that shared key — `shared:go.mod` (go.mod, go.sum), `shared:templates` (internal/template/templates/**), `shared:build` (Makefile, .github/**), `shared:rules` (.claude/rules/**); when a lane is about to edit outside its held keys, it SHALL call `ExpandKeys` before the edit, and when a requested key is held by another card, `ExpandKeys` SHALL return `ErrKeyConflict` and SHALL park no card. |
| REQ-AFL-012 | D4 충돌 키 잠금 | Where two cards hold disjoint conflict keys, AgenticFactory SHALL let both be claimed and worked on at the same time and SHALL NOT order card work through a global serial slot; when `ExpandKeys` returns `ErrKeyConflict`, the lane SHALL wait on that card only and SHALL remain free to claim other work. |

### C.3 카드 흐름과 병합 큐 (계약 §5, D3)

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AFL-013 | §5 카드 흐름 | A claimed card SHALL move through BUILDING with a RED evidence record showing its acceptance tests failing, then a GREEN fix loop of at most 3 iterations after which the card is parked with reason `fix-limit`, then the gates (build, vet, lint, affected and reverse-dependency tests, widened when shared keys are held), then REVIEWING only when `spec.risk` is non-empty or the diff exceeds 400 changed lines, then MERGING through `Enqueue`; when the merge queue reports a failure, the card SHALL return to BUILDING/FIX with that `fail_reason`. |
| REQ-AFL-014 | D3 병합 큐 완료 | When a card is in MERGING, the merge queue SHALL verify the merge result in a candidate worktree outside the landing lock, then take the lock `merge-queue:landing`, re-check trunk, update trunk with a compare-and-set `git update-ref <trunk> <new> <old>`, record a `landed` note, move the card to DONE, and release it; card completion SHALL be judged from this merge-queue evidence, no leader approval receipt SHALL be required for it, and after a restart the merge queue SHALL reconcile its rows against git ancestry. |

### C.4 퇴역·이관·명령·기본값 (계약 §4·§6, D1)

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AFL-015 | §4 Issue·원격 코드 퇴역 | When the new path is green (판정 기준 미정, OD-35), the code paths `forge`, `mirror`, `effect/ghintake`, `effect/ctxledger`, `effect/labels`, `gitref`, the GitHub lookups of `identity`, the GitHub search of `intake`, and the GitHub routes of `resident` SHALL be retired by removing every import of a package first and deleting the package last, in one reviewed commit per package. |
| REQ-AFL-016 | importer | The importer in `internal/agenticfactory/ledger/importer` SHALL bring backlog.db queue items, factory.db card rows, and claim JSON documents into the ledger, SHALL open every source read-only and change none of them, and SHALL provide a dry-run mode as its first deliverable (dry-run의 원장 쓰기 여부는 미정, OD-35). |
| REQ-AFL-017 | `moai af` 명령 | The `moai af` command SHALL provide verbs from which every step of the REQ-AFL-013 card flow (claim, stage move, evidence record, `Enqueue`, `Release`) is reachable; the verb names and their grammar are 미정 (OD-33). |
| REQ-AFL-018 | §6 기본값 | Each of the values heartbeat throttle 30 s, lease 15 min, renew 5 min, reclaim cap 3, fix-loop cap 3, review trigger 400 changed lines, landing lock TTL 10 min, and snapshot interval 30 min while the store is open SHALL be a named constant, and no code on the local path SHALL write any of these values as an inline number. |

### C.5 관찰과 heartbeat (계약 §4 lane-20·lane-21)

| ID | 묶음 | 요구사항 |
|---|---|---|
| REQ-AFL-019 | 흐름 읽기 모형 | The flow read model in `internal/agenticfactory/monitor/flow` SHALL compute WIP, age per stage, wait per state, and stalled detection taking its input only from `ledger.Reader`. |
| REQ-AFL-020 | 보드 | The board in `internal/web` SHALL show cards, leases, keys, queue, notes, and flow metrics read from `ledger.Reader` and SHALL perform no write. |
| REQ-AFL-021 | 훅 heartbeat 기록기 | The hook heartbeat writer SHALL write a session's heartbeat at most once per 30 s, and when that session's stamp file is younger than 30 s, it SHALL skip the database write. |

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
| OD-35 | 계약이 말하지 않는 세부 11가지(상위 OD-35 행에 나열) | REQ-AFL-003·004·007·010·011·013·015·016·018·021, AC-AFL-003·013·015·017·018·021 |

## §H 확인 공백

- 사용자 결정 원문은 읽지 않았다(리더의 옮겨 적음만 읽음). 리더 보고 v2(artifact)는 읽지 않았다.
- 시험은 하나도 돌리지 않았다. 이름 21개가 없다는 것은 한 번의 `git grep` 결과다(acceptance.md §B).
- REQ-AFL-003의 "트랜잭션 안에서 외부 명령 없음, 에이전트 턴을 넘는 트랜잭션 없음"은 AC-AFL-003의 시험 하나로 덮이지 않는다. 그 부분의 확인 방법은 미정이다(OD-35 (j); 계약 §4에 적힌 검토 경로는 lane-19의 읽기 전용 검토뿐이다).
- §6 상수 가운데 원장 패키지 밖에 둘 것(heartbeat 30 s, fix-loop 3, review 400, landing TTL 10 min)의 패키지는 미정이므로 AC-AFL-018의 시험은 원장 패키지 상수만 덮는다(OD-35 (d)).

## §I 범위 밖

### Out of Scope — 원격 게시

- push, PR, 태그, 공개 릴리스. D5는 정해지지 않았고(상위 OD-31) 이 SPEC은 아무것도 게시하지 않는다.

### Out of Scope — 배포와 held 카드 처리

- `make build VERSION=v3.2.0-rc.39`, 백업, 로컬 설치, held 카드 165장의 복구·가져오기·분류(계약 §7 P3·P4). 이것은 리더의 잠정 해석에 든 단계이며 이 SPEC의 요구가 아니다.

### Out of Scope — 규칙 문서와 template

- lane 루프의 규칙 문서·template 수정(계약 §4 lane-17 둘째 산출물). 이 SPEC은 규칙 파일을 고치지 않는다.

### Out of Scope — 옛 원장 직접 수정

- 옛 원장(backlog.db, factory.db)을 SQL로 직접 고치는 것. importer는 읽기만 한다(REQ-AFL-016).
