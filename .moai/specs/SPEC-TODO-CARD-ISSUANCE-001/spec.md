---
id: SPEC-TODO-CARD-ISSUANCE-001
title: "카드 발행 품질 — 발행 시점의 겹침·중복 제시, 카드 관계 그래프, 묶음 직렬 경로, 과분할 억제 규칙"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: MoAI
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/graph, internal/homestate, internal/web, internal/template/templates, .claude/rules, .claude/skills, .claude/agents"
lifecycle: spec-anchored
tags: "card-issuance, todo-queue, relation-graph, card-bundling, card-merge, issuance-rules, card-t1454"
tier: L
card: t1454
depends_on: [SPEC-TODO-ANALYSIS-001, SPEC-TODO-AUTO-PICK-001, SPEC-TODO-ARCHIVE-QUERY-001, SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-RELATION-PICKUP-FILTER-001, SPEC-GTD-AUTONOMY-001, SPEC-WEB-TODO-QUEUE-001, SPEC-TODO-SURFACE-POLISH-001]
related_specs: [SPEC-TODO-CLAIM-LEASE-001, SPEC-TODO-TRANSITION-STAMPS-001, SPEC-JEV-CONSUMERS-001, SPEC-INSTRUCTION-BUDGET-SCOPE-001, SPEC-KANBAN-PR-CARD-TRACEABILITY-001]
---

# SPEC: 카드 발행 품질

## HISTORY

- 0.1.0 — 2026-10-03 — 카드 t1454(운영자 요청 10-02 밤, Class C, Tier L, 단일 카드 유지)의 plan 단계 산출물을 처음 작성했다. 근거는 세 개의 읽기 전용 렌즈와, 이 계획이 같은 트리 `2de0a2cb6`(전체 SHA `2de0a2cb613b04765a1554f86685a3b48e0be806`)에서 다시 잰 값이다. 카드가 인용한 측정 수치는 다른 세션의 scratchpad 에서 왔으므로, git 쪽 수치는 이 트리에서 재현했고(전부 일치) 큐 쪽 수치는 스냅숏 위에서 다시 쟀다(드리프트 기록). 카드 본문의 낡은 전제 여섯 가지는 §A.3 에서 바로잡았다. 결정 열다섯 가지는 `decision-index.md` 와 `design.md` §11 에 선택지·근거·작업 기본값으로 두었고, 이 SPEC 에는 미해결 질문 표지를 두지 않는다(카드의 리더가 감사 근거로 판정한다). 요구 24개, 수용 기준 24개(출시 차단 21, 회귀 가드 3).

## Amendments

이 SPEC 은 아래 완료 SPEC 의 행동·예약 항목을 바꾼다. 각 행은 대상 SPEC 의 `## Amendments` 절로 **sync 단계에서** 옮겨 적을 계획이며(manager-spec 재위임, `completed → in-progress (amendment)` 전이, `amendment_of:` 자기 참조), 이 SPEC 의 plan 단계는 대상 SPEC 을 편집하지 않는다. `prior_completed_sha` 는 대상 SPEC 의 `progress.md` §E.4 `sync_commit_sha` 에서 읽은 값이다. 조건부 행은 결정 D5·D2·D6 의 판정에 따라 적용 여부가 갈리며, 적용하지 않기로 한 행은 sync 때 "미적용"으로 기록한다.

**A1 — SPEC-TODO-ANALYSIS-001 (v1.0.0), 조건부(D5)**

- **직전 완료 버전**: 1.0.0 (`status: completed`)
- **prior_completed_sha**: `b6716a748`
- **사유**: REQ-TA-010 은 기계 분석기(analyser)를 `add` 와 `analyze` 로 한정한다. 이 SPEC 의 발행 시점 제시(REQ-TCI-002, -004, -005)는 분석기가 아니라 읽기 전용 조회 경로로 만든다. D5 가 "`gtd engage` 가 분석기에 합류한다"를 고르면 REQ-TA-010 의 문면이 실제로 바뀐다.
- **범위**: 기본 작업값(제시는 별도 읽기 경로, engage 는 제시만 하고 분석기는 돌지 않는다)에서는 문면 변경이 없고 이 행은 "명확화"로 기록한다. D5 가 분석기 합류를 고르면 REQ-TA-010 에 `gtd engage` 를 추가하고 "`list` 는 두 번 연속 실행해도 같은 출력" 불변식은 그대로 둔다.

**A2 — SPEC-TODO-ANALYSIS-001 (v1.0.0), 범위는 D6 가 정한다**

- **직전 완료 버전**: 1.0.0
- **prior_completed_sha**: `b6716a748`
- **사유**: §B.1 의 표는 "한 카드를 다른 카드에 접어 넣기"를 금지로 판정했고(이유: 자동 변경은 눈에 보이지 않는다), §E 의 `### Out of Scope — 카드 흡수·병합` 은 흡수를 "`drop` + `edit` 의 운영자 조합"으로만 남겼다. `.claude/skills/moai/workflows/gtd.md` 의 두 문장(`absorbs` 행의 "`absorbs` does not absorb", 그리고 "Analysis never folds one card into another…")이 이를 문서로 못 박고 있고, 코드는 "병합 동사의 부재"로 이를 강제한다(`internal/cli/todo_relate.go` 머리 주석).
- **범위**: 운영자가 직접 호출하는 명시적 병합 동사(`moai todo merge`, REQ-TCI-019)를 새로 둔다. 분석기·`analyze`·`relate`·레인 세션은 여전히 카드를 접지 못한다 — 이를 코드 형태와 테스트로 계속 강제한다. 개정되는 문면은 §B.1 표의 해당 행, §E 의 해당 Out of Scope 절, REQ-TA-014 가 요구하는 "never folds" 문장(분석이 접지 않는다는 뜻으로 한정하고 운영자 동사의 예외를 덧붙인다)이다.

**A3 — SPEC-TODO-AUTO-PICK-001 (v0.4.1)**

- **직전 완료 버전**: 0.4.1
- **prior_completed_sha**: `4293b2d73`
- **사유**: §B.7, REQ-TAU-013, 그리고 `### Out of Scope — schemas, relation kinds, graph, and the t1454 inputs` 가 "관계 레코드 저장소 통합, `moai graph` 변경, 예상 파일·부모·크기 필드는 카드 t1454 의 몫"이라고 예약했다. 이 SPEC 이 그 예약을 소비한다. 또한 REQ-TAU-013 은 선택 입력 집합이 열려 있어 새 입력이 keep-set 과 임대 경로를 고치지 않는다고 못 박았는데, 묶음(REQ-TCI-018, -020)은 임대 선택 호(arm)가 묶음 예약을 읽도록 바꾼다 — 이것은 "새 입력 추가"가 아니라 임대 경로 변경이므로 D14 의 판정에 따라 문면 개정이 필요하다.
- **범위**: (i) 예약 문장에 "t1454 가 소비했다"를 기록, (ii) 관계 레코드가 하나의 해석기를 통해 읽힌다는 사실과 파일 겹침 입력이 카드가 예상 파일을 가질 때 측정값이 된다는 사실을 기록(결정 기록의 `evidence_refs` 에 측정된 겹침이 들어갈 수 있다), (iii) D14 가 임대 경로 변경을 고르면 REQ-TAU-013 후반("임대 경로를 고치지 않고")을 묶음 예외로 좁힌다. keep-set 은 파일 겹침을 읽지 않는다는 문장은 그대로 둔다.

**A4 — SPEC-RELATION-PICKUP-FILTER-001 (v0.1.0)**

- **직전 완료 버전**: 0.1.0
- **prior_completed_sha**: `dcf744fdc`
- **사유**: §A.3 "두 관계 체계(섞지 말 것)" 표가 `todo relate` 소견과 `gtd_relations` 를 별개 저장소로 못 박았다. D2 의 기본 작업값에서도 두 저장소는 그대로 분리돼 있고, 읽는 쪽에 해석기가 생긴다.
- **범위**: "섞지 말 것"을 "저장은 분리, 읽기는 해석기 하나"로 좁힌다. REQ-RPF-001 의 "소견의 존재가 곧 미해결이다"와 `blocks`/`depends` 의 선택 필터 동작은 그대로다.

**A5 — SPEC-GTD-AUTONOMY-001 (v0.2.0), 조건부(D2)**

- **직전 완료 버전**: 0.2.0
- **prior_completed_sha**: `5ec516165ef5c5e31cdf372b28e6e538f8347719`
- **사유**: D2 의 기본 작업값에서는 `gtd_relations` 를 읽기 전용으로 읽는 해석기가 생길 뿐 쓰기 경로와 아홉 종류 관계는 그대로라 이 행은 필요 없을 수 있다. D2 가 "단일 저장소"를 고르면 관계 종류와 PK 의미가 바뀌므로 개정이 필요하다.
- **범위**: 기본값에서는 "미적용, 읽기 소비자 추가를 기록"으로 닫는다. 단일 저장소를 고르면 아홉 종류 관계 목록과 비공개 투영(`gtd-edges.jsonl`) 문장을 다시 쓴다.

**A6 — SPEC-WEB-TODO-QUEUE-001 (v0.1.0)**

- **직전 완료 버전**: 0.1.0
- **prior_completed_sha**: `f1e71db4b`
- **사유**: REQ-WTQ-003 은 `/todo` 가 큐가 가진 모든 항목을 세 상태로 나열한다고 정했고 REQ-WTQ-001 은 콘솔의 쓰기·락 획득을 금지한다. 관계 그래프 보기(REQ-TCI-023)는 같은 라우트에 보관 카드와 관계를 읽는 두 번째 보기를 더한다.
- **범위**: REQ-WTQ-001(쓰기 금지)과 REQ-WTQ-003(표의 모집단)은 바뀌지 않는다. 읽기 집합이 보관 카드·관계로 넓어졌음과 새 쿼리 매개변수 `view` 를 기록하는 행이다.

## §A 배경

### A.1 문제

카드 발행이 "무엇을 발행하느냐"가 아니라 "이미 있는 것과 얼마나 겹치느냐"에서 새고 있다. 카드가 인용한 측정과, 이 계획이 같은 트리에서 다시 잰 값은 다음과 같다(전체 표와 재현 명령은 `research.md` §3).

| 항목 | 카드 인용값(다른 세션 scratchpad) | 이 계획의 재측정(트리 `2de0a2cb6`, 큐 스냅숏 10-03 11:45) | 판정 |
|---|---|---|---|
| 착지 카드(first-parent develop 커밋 보유) | 964 | 964 | git 쪽 재현, 동일 |
| 제품 줄 0 / 50 미만 | 111(12%) / 305(32%) | 111(12%) / 305(32%) | 동일 |
| 줄 기준 공정 산출물 비중 | 63% | 63% | 동일 |
| 같은 주 파일 72시간 묶음 | 64개·148장, 절약 84(엄격형 39) | 64개·148장, 절약 84(엄격형 39) | 동일 |
| 큐 행 수 | 1,164 | 1,174(live 127 + 보관 1,047) | 큐가 계속 자란다 |
| 파생 카드 | 468(40.2%) | 473(40.3%) | 드리프트 |
| 파생의 파생(depth ≥ 2) | 178(15.3%) | 181(15.4%) | 드리프트 |
| near-duplicate 0.8 이상 | 44건 중 37건 완료 | 47건 중 40건 완료 | 드리프트, 전부 `source=jev` |
| 바이트 동일 본문 | 60쌍 | 61그룹(122장) | 정규화 방식 차이 |

카드의 숫자를 그대로 믿지 않고 다시 잰 결과 두 가지가 새로 드러났다. 첫째, 기계 분석기의 near 구간(token-set Jaccard 0.80 이상 1.0 미만)은 큐 전체 역사에서 한 번도 발화하지 않았다 — 근접 이웃이 그 구간에 있는 카드가 0장이고, 기록된 `near-duplicate` 소견 107건은 전부 Jev 가 판정했다. 둘째, 어휘 겹침은 약한 변별자다. 발행 시점 가정에서 가장 가까운 이전 카드의 점수 중앙값이 0.10 이고, Jev·에이전트가 기록한 관계 상대가 어휘상 상위 3개 안에 드는 비율은 하한 없이 52%, 하한 0.30 에서 6%다. 따라서 M1 의 제시는 "확실한 것"(정확 중복, 보관·dropped 카드 포함)과 "참고용"(어휘 이웃)을 구분해서 보여야 하고, 어휘 이웃이 의미 중복 판정을 대체한다고 약속해서는 안 된다.

### A.2 이 SPEC 이 하는 일

여섯 마일스톤을 순서대로 한다(실행 순서: M0, M1, M2, M3, M4, M6, 그다음 게이트가 열렸을 때 M5).

- **M0** 카드의 기준선을 추적되는 형태로 반입하고 실행 단계에서 다시 잰다.
- **M1** `add` 가 막지 않고 알린다 — 유사 카드 상위 3개(보관·dropped 포함), 같은 구성요소의 열린 카드, 진행 중 레인과의 예상 파일 겹침, 이미 덮는 완료 SPEC. `--dry-run` 을 지원한다.
- **M2** 카드에 발행 속성(스폰한 카드·출처·크기 추정·예상 파일·drop 사유)을 가산적으로 싣고 finding 에 처분(수용·병합·기각)을 싣는다.
- **M3** 관계를 일곱 종류의 한 어휘로 묶고(종류별 순환·카디널리티 제약), 두 저장소를 해석기 하나로 읽고, 추이 질의 동사와 `moai graph` 의 card→file 간선을 만든다.
- **M4** 묶음 직렬 경로, 운영자 호출 카드 병합 동사, 허브 파일 직렬 소유.
- **M5**(게이트됨) 카드 크기·후속 지적·파생 깊이·동시 진행 한도·발행 체크리스트 규칙. 선행 조건은 명령으로 판정한다.
- **M6** `moai web` `/todo` 에 관계 그래프 보기.

### A.3 바로잡은 낡은 전제

카드 본문과 렌즈 요약 일부가 이 트리의 현실과 어긋나 있었다. 이 SPEC 은 아래 정정을 기준으로 쓴다.

1. 카드는 near-duplicate 기록이 `todo.go:800` 에 있다고 적었다. 실제 위치는 `internal/cli/todo_analysis.go:74`(`case kanban.BacklogMatchNear`)이고 임계값은 `internal/kanban/backlog_analysis.go:33`(`BacklogNearDuplicateThreshold = 0.80`)이다.
2. t1349 의 세 항목(`show` 동사, `add` 의 대시 시작 본문 파싱, list 기본 상한)은 SPEC-TODO-SURFACE-POLISH-001(completed, 커밋 `1f7eeae49`)이 이미 끝냈다. 이 SPEC 의 범위가 아니다.
3. 카드는 `assign --after` 를 todo 동사처럼 적었다. 실제로는 `moai factory assign` 이다(`internal/cli/factory_card.go:1703` 부근에서 `--prefer`/`--after`/`--spec` 를 읽는다).
4. 카드는 "현 규칙은 병합 금지"라 했지만 `kanban-dispatch*.md`·`sync-auditor.md`·`manager-todo.md` 어디에도 그런 문장이 없다. 금지는 `.claude/skills/moai/workflows/gtd.md` 의 두 문장과 병합 동사가 없다는 코드 형태가 만든다(A2 참조).
5. t1448 은 이미 이 트리에 착지했다(병합 `4315f0d0e` 는 HEAD 의 조상). t1453 은 아직 착지하지 않았다 — 게이트 명령이 0 을 읽는다(§C Module E, `acceptance.md` L21).
6. 카드는 `moai graph` 의 "커밋된 `edges.jsonl`"을 전제했지만 `.moai/project/graph/` 는 `.gitignore:344` 로 추적 제외된 파생 산출물이다(`internal/graph/check.go` 도 "untracked derived artifact"라 적는다). 결정론 요구는 그대로이고, "비공개 큐 상태를 커밋된 산출물에 섞지 말 것"은 "`moai graph build` 의 동일 트리 동일 출력 계약을 지킬 것"으로 바뀐다(D3).

## §B 결정과 작업 기본값

열다섯 결정은 `decision-index.md`(Detect → Explain → Ask, 선호 답 없음)와 `design.md` §11(선택지·측정 근거·작업 기본값)에 있다. 아래 표의 "작업 기본값"은 계획이 이 값으로 진행한다는 뜻이지 판정이 아니다 — 카드의 리더가 감사 근거로 판정하고, 판정이 다르면 해당 요구와 기준의 리터럴만 다시 고정한다.

| ID | 결정 | 계획이 진행하는 작업 기본값 |
|---|---|---|
| D1 | M5 의 순서와 게이트 미충족 시 인도물 | M5 는 마지막 마일스톤이고 게이트 명령이 판정한다. 미충족이면 규칙 파일은 건드리지 않고 초안과 후속 카드 문안만 낸다 |
| D2 | 두 관계 저장소 통합 방식 | 저장소는 분리 유지, 읽기 해석기와 어휘 확장으로 통합 |
| D3 | card→file 간선 운반체 | 커밋된 증거(카드 귀속 병합 커밋)에서만 `edges.jsonl` 에 간선 생성, 비공개 사이드카는 만들지 않는다 |
| D4 | M2 스키마 형태 | nullable JSON 컬럼 하나(`issuance`), finding 은 컬럼 하나(`disposition`) |
| D5 | MCP 알림 전달과 `gtd engage` | MCP 결과 텍스트의 id 줄 뒤에 제시를 덧붙이고, engage 는 제시만 하고 분석기는 돌리지 않는다 |
| D6 | 병합 경로 독트린 개정 범위 | 운영자 호출 동사 하나로 한정, 분석·relate·레인은 계속 금지 |
| D7 | 수치 임계값 | SPEC 은 값을 고정하지 않고 M0 의 기준선 기록을 가리킨다 |
| D8 | 유사도 척도와 소음 | 기존 token-set Jaccard 재사용, 표시 하한을 두고 점수와 척도를 항상 함께 표기 |
| D9 | closed-at | 저장하지 않고 기존 두 스탬프에서 읽는 접근자 하나 |
| D10 | 파생 부모의 단일 원천 | 카드 속성이 원천, `parent-of`/`follow-up-of` 는 읽기 투영 |
| D11 | finding 처분의 효과 | 기록만 한다, 카드·순서·픽업 필터에 영향 없음 |
| D12 | 구성요소별 부채 대장 운반체 | 큐의 카드(출처 `debt`)와 구성요소 노드로 표현, 새 저장소 없음 |
| D13 | `add` 새 플래그의 입력 방식 | 명시 플래그만, 본문에서 추론해 저장하지 않는다 |
| D14 | 묶음 메커니즘과 임대 경로 개정 | 팩토리 레코드의 묶음 컬럼과 선택 호의 묶음 예약(임대 경로 개정 인정) |
| D15 | 신규 경로 한정 파일의 40,000자 한도 | 이미 정해짐 — SPEC-INSTRUCTION-BUDGET-SCOPE-001 |

## §C 요구사항 (GEARS)

요구는 다섯 모듈이다. GEARS 표기를 쓰고 모든 `REQ-TCI-NNN` 은 `acceptance.md` 의 `AC-TCI-NNN` 하나 이상이 추적한다. 요구 문장은 리포의 한국어 SPEC 관행대로 GEARS 영어 문형으로 적고, 설명과 근거는 한국어로 적는다. 숫자 임계값은 이 문서에 쓰지 않는다 — 값은 M0 가 만드는 기준선 기록에서 읽고(REQ-TCI-001), 계획의 작업 기본값은 `plan.md` §F.11 과 `design.md` §11 에 있다.

### Module A — 기준선과 발행 시점 제시 (M0, M1)

- **REQ-TCI-001** (Ubiquitous): The baseline record under `.moai/reports/t1454/baseline/` shall hold, for every figure on which a threshold or rule value of this SPEC rests, the producing command, its verbatim output and the tree SHA it was measured on, and every later requirement or rule that names a threshold shall take its value from that record rather than from this document.
- **REQ-TCI-002** (Event-driven): **When** a card is admitted through `moai todo add`, `moai gtd add` or the bare `moai todo <text>` fallthrough, the add path shall print on its error stream a presentation of (a) up to three cards most similar to the new text across live, dropped (reason prefix stripped) and archived cards, (b) open cards sharing a component with it, (c) in-flight lane cards whose expected files overlap its own, and (d) completed SPECs that already cover it, each item at or above its display floor and labelled with its source and measure, and its output stream shall remain exactly `<id> <position>` followed by a newline.
- **REQ-TCI-003** (Unwanted): The presentation shall not refuse, delay beyond its time bound, reorder, edit, fold or drop any card or otherwise change the admission, so that the queue after an add is identical to the queue the same add would produce without the presentation apart from the new card's own row and the findings the analyser already records, and any probe that reads outside the queue (git, the SPEC directory) shall run outside the queue's cross-process write lock and under a time bound.
- **REQ-TCI-004** (Capability-gate): **Where** `--dry-run` is passed to `moai todo add` or `moai gtd add`, the add path shall print the same presentation, shall write nothing (the queue file byte-identical, no id consumed, `last_seq` unchanged) and shall exit 0, and for an exact duplicate shall report that a real add would refuse instead of refusing.
- **REQ-TCI-005** (Event-driven): **When** a card is admitted through the MCP `todo_add` tool or through `moai gtd engage`, the admission shall carry the same presentation — in the tool result's text after its first `<id> <position>` line for MCP, on the error stream for engage — without recording any finding on the engage path and without changing the tool result's first line.
- **REQ-TCI-006** (Ubiquitous): The lookup that feeds the presentation shall be a read-only path separate from the duplicate classifier, so that the classifier, the text normalizer, the Jaccard measure and the near-duplicate threshold shall keep their behavior and their tests, including the classifier's skip of dropped cards.

### Module B — 카드 스키마 (M2)

- **REQ-TCI-007** (Ubiquitous): Every card, live or archived, shall be able to carry optional issuance attributes — the card that spawned it, its origin, a size estimate, its expected files and a drop reason — stored additively (absence is a SQL NULL and a nil value, never an empty object or string) through the engine's additive-column path rather than the table-rebuild column list, so that a card carrying none of them serializes byte-identically to before.
- **REQ-TCI-008** (Event-driven): **When** a queue database written before this change is opened, the engine shall add the missing storage to both card-bearing tables without losing or altering any existing row, and the pure reader shall return nil attributes for a database that lacks it without running any schema change.
- **REQ-TCI-009** (Ubiquitous): Archiving a card and restoring it shall carry its issuance attributes unchanged, and the in-memory record, the JSON rendering and the SQLite rows of one queue shall agree on them, the engine's parity assertion covering the new storage.
- **REQ-TCI-010** (Event-driven): **When** a card is dropped, `moai todo drop` shall store the reason in the drop-reason attribute while continuing to write the `[DROPPED — <reason>] ` text prefix, and the closing time of a card shall be readable through one accessor that derives it from the archive or drop stamp and answers "unknown" for a card that carries neither.
- **REQ-TCI-011** (Ubiquitous): A finding shall be able to carry an optional disposition — accept, merge or reject — set only through an explicit operator verb, absent for every existing finding, and recording a disposition shall change no card, no finding relation and no queue order.

### Module C — 관계 모델과 그래프 (M3)

- **REQ-TCI-012** (Ubiquitous): Card relations shall be expressed in one vocabulary of seven kinds — blocks, duplicates, parent-of, follow-up-of, merged-into, supersedes, relates-to — over nodes of type Card, Component, File, Commit, Spec and Finding, with every legacy finding relation mapped onto it at read time (depends and blocks to blocks, near-duplicate and duplicate-forced to duplicates, replaces to supersedes, contains, absorbs and conflicts to relates-to with the legacy name kept as a qualifier) and no stored row rewritten.
- **REQ-TCI-013** (Event-detected): **When** a relation write would give a card a self-edge, close a cycle in a kind that forbids cycles, or exceed the cardinality of its kind (a card with a second parent, a second origin or a second merged-into target), the write verb shall refuse with a message naming the kind and the offending pair and shall leave the queue file byte-identical.
- **REQ-TCI-014** (Ubiquitous): One resolver shall map a queue card id to its GTD item id and back through the engagement link, so that a single read can show a card's findings together with its `gtd_relations`; the resolver shall add no write path to `gtd_relations` and shall leave `moai gtd organize` unchanged.
- **REQ-TCI-015** (Event-driven): **When** `moai todo trace <id> [--kind <k>] [--depth <n>]` runs, the command shall print every node reachable from the card over the named kinds up to the depth bound in a deterministic order, shall terminate on cycles that legacy data may contain, shall write nothing and shall be available to a lane session like the other read-only verbs.
- **REQ-TCI-016** (Ubiquitous): `moai graph build` shall derive card-to-file edges only from committed evidence — card-attributed first-parent merge commits and the files they changed — so that two builds on one tree produce byte-identical output, the output contains no queue-private state (no unlanded card id, expected file or finding), and `moai graph check` shall notice a change to that evidence source.
- **REQ-TCI-017** (Ubiquitous): For every record that predates this change, `todo list`, `todo list --json`, `todo why`, `todo export`, the pickup filter and the auto-rank near-duplicate path shall behave and render exactly as before.

### Module D — 묶음과 병합 (M4)

- **REQ-TCI-018** (Event-driven): **When** cards are assigned to one bundle, the factory shall lease them to the bundle's lane one at a time in bundle order, offering the next member to that lane after the previous member's local merge and withholding members from every other lane, while the fleet-wide serial slot keeps its present meaning.
- **REQ-TCI-019** (Event-driven): **When** an operator runs `moai todo merge <into> <from>`, the command shall append the text of `<from>` to `<into>` as an explicit section, record a merged-into relation from `<from>` to `<into>`, drop `<from>` with its reason, and refuse when either card is picked, already merged, or the pair would form a cycle; and neither the analyser, `analyze`, `relate` nor any lane session shall be able to invoke it.
- **REQ-TCI-020** (State-driven): **While** a hub file — a path the baseline record lists as touched by more cards than its hub threshold — appears in the expected files of two open cards, the factory shall order them as a bundle chain so that the second is leased only after the first has merged, and the keep-set shall continue to read no file overlap.

### Module E — 규칙과 웹 보기 (M5, M6)

- **REQ-TCI-021** (Capability-gate): **Where** the M5 gate holds — the nearest card-t1453 merge is an ancestor of the working tree's HEAD, read by the gate command with its positive control — the rule files named by the card and their template copies shall state a card-size floor and ceiling, the follow-up rule (a defect a card creates is fixed inside that card before merge, a pre-existing defect is entered in the per-component debt ledger), a derivation-depth cap, a concurrent-in-flight cap and an issuance checklist, each numeric value taken from the baseline record, within the instruction budget (the always-loaded `kanban-dispatch.md` without net growth, every touched file within 40,000 characters or no larger than before) and with every mirror guard green.
- **REQ-TCI-022** (Capability-gate): **Where** the M5 gate does not hold when M5 would start, the card shall leave the six rule files and their template copies unedited, shall deliver the rule text as a ready-to-apply draft under `.moai/reports/t1454/m5-draft/` with exact insertion anchors, and shall record in `progress.md` the gate command output and the follow-up card text the leader is to issue.
- **REQ-TCI-023** (Event-driven): **When** the `/todo` page is requested with `?view=graph`, the console shall render a server-side relation graph of cards (live, dropped and archived) and their relations from the unified read seam, bounded in node count, GET-only, with no write and no lock acquisition, no network fetch, embedded assets and every new string present in all locales.
- **REQ-TCI-024** (Ubiquitous): The existing `/todo` table, its sorts, its detail pane and its live refresh shall render and behave exactly as before for the same queue.

## §D 브라운필드 델타

### [DELTA] 발행 시점 제시 (Module A)

- [EXISTING] `ClassifyCardText` 의 정확·근접 분류와 dropped 카드 건너뛰기 — 특성 테스트만 유지한다(REQ-TCI-006).
- [EXISTING] `add` 의 stdout 한 줄 `<id> <position>` 과 정확 중복 거절 — 바이트 단위로 보존한다.
- [MODIFY] `add` 경로(`internal/cli/todo.go` 의 `scanTodoAddArgs`, `runTodoAddAppendRoot`) — `--dry-run` 을 파싱하고 락 밖에서 제시를 출력한다.
- [MODIFY] MCP `todo_add` 핸들러 — 버려지던 stderr 버퍼의 제시를 결과 텍스트에 덧붙인다.
- [MODIFY] `moai gtd engage` — 제시만 출력한다(기록·거절 없음).
- [NEW] 읽기 전용 이웃 조회(보관·dropped 포함, 접두사 제거), 같은 구성요소 조회, 완료 SPEC 조회, 진행 중 레인 겹침 입력, 제시 렌더러.
- [NEW] 기준선 기록 `.moai/reports/t1454/baseline/` 과 그 재현 스크립트.

### [DELTA] 카드 스키마 (Module B)

- [EXISTING] `ensureColumn` 가산 컬럼 패턴, `backlogItemsTableColumns` 의 v1→v2 재구성 목록, `todoJSONProjection`, JSON 골든.
- [MODIFY] `BacklogItem`·`BacklogFinding`·`BacklogArchiveEntry` 와 SQLite 읽기·쓰기·parity 경로, 동결 컬럼 튜플 테스트.
- [MODIFY] `moai todo drop` — 사유를 속성에도 저장한다(텍스트 접두사는 그대로).
- [NEW] 발행 속성(스폰한 카드·출처·크기 추정·예상 파일·drop 사유), finding 처분, 닫힘 시각 접근자, `add` 의 명시 플래그.

### [DELTA] 관계 모델과 그래프 (Module C)

- [EXISTING] 소견 어휘 여덟 종류와 `WaitsOnOf`·`FindingsBlocking`·`WaitsOnClosesCycle`, `gtd_relations` 아홉 종류와 `ValidateGTDRelation`, 비공개 `gtd-edges.jsonl` 투영.
- [MODIFY] `todo relate` — 새 종류와 처분 동사, 종류별 제약 검사. `moai graph build`·`check` — 새 간선 층과 출처 지문.
- [NEW] 일곱 종류 어휘와 읽기 시 매핑, 카드↔GTD 해석기, `moai todo trace`, card→file 간선 층.
- [REMOVE] 없음. 저장된 행은 하나도 다시 쓰지 않는다.

### [DELTA] 묶음과 병합 (Module D)

- [EXISTING] 직렬 슬롯, `assign --after` 의 이전 카드 병합 가드, keep-set 의 "파일 겹침을 읽지 않는다".
- [MODIFY] 팩토리 `cards` 행(묶음 컬럼 추가)과 임대 선택 호.
- [NEW] `moai factory bundle`(또는 동등한 적재 동사), `moai todo merge`, 허브 파일 체인 생성.

### [DELTA] 규칙과 웹 (Module E)

- [EXISTING] 카드 등급 A/B/C 문장, Class A 크기 문구, "one card per worktree", sync-auditor 의 blocking/optional 분류와 `PASS-WITH-DEBT`, 템플릿 미러와 그 가드들.
- [MODIFY] `gtd.md`(+미러), `kanban-dispatch.md`(+분기된 미러), `kanban-dispatch-detail.md`(+미러), `sync-auditor.md`(+분기된 미러, 생성 Codex TOML), `manager-todo.md`(+미러, 생성 Codex TOML), `catalog.yaml` 해시.
- [NEW] `card-issuance.md`(경로 한정 규칙, +미러), 웹 관계 그래프 보기와 자산.

## §E 수정 대상 파일

M1: `internal/cli/todo.go`, `internal/cli/todo_issuance.go`(신규), `internal/cli/mcp_todo.go`, `internal/cli/gtd.go`, `internal/kanban/backlog_issuance.go`(신규)와 각 `_test.go`.
M2: `internal/kanban/backlog_store.go`, `backlog_sqlite.go`, `backlog_migrate.go`, `backlog_schema_freeze_test.go`, `internal/kanban/todo_queue_merge.go`·`todo_merge_procedure.go`(항목 복사 경로), `internal/cli/todo.go`, `todo_drop.go`, `todo_claim.go`(JSON 투영), `todo_export.go`.
M3: `internal/kanban/backlog_relation.go`(신규), `backlog_store.go`, `internal/cli/todo_relate.go`, `todo_trace.go`(신규), `internal/graph/card_file.go`(신규), `graph.go`, `meta.go`, `internal/cli/graph.go`.
M4: `internal/homestate/factory.go`·`card_record.go`·`card_picked.go`·`card_transition.go`, `internal/cli/factory_card.go`, `internal/cli/todo_merge.go`(신규), `internal/cli/todo.go`(읽기 전용 동사 목록과 서브커맨드 등록).
M6: `internal/web/todo_queue_read.go`, `todo_view.go`, `screens.templ`(+생성물 `screens_templ.go`), `assets.go`, `assets/` 신규 자산, `assets/i18n.js`.
M5(게이트 충족 시): `.claude/rules/moai/workflow/card-issuance.md`(신규), `gtd.md`, `kanban-dispatch.md`, `kanban-dispatch-detail.md`, `.claude/agents/moai/sync-auditor.md`, `manager-todo.md`와 각 `internal/template/templates/` 사본, `internal/template/catalog.yaml`, 생성 `.codex/agents/moai/*.toml`(저장소 루트와 템플릿 하위).

## §F 범위 밖 (Exclusions)

### Out of Scope — t1349 의 세 항목

- `show` 동사, `add` 의 대시 시작 본문 파싱, list 기본 상한은 SPEC-TODO-SURFACE-POLISH-001(completed, 커밋 `1f7eeae49`)이 끝냈다. 이 SPEC 은 건드리지 않는다.

### Out of Scope — 다른 카드의 규칙 파일 작업

- t1453(github-flow 전환)의 절체 시점 편집, t1450(상시 로드 규칙 다이어트·적재 범위 재설계), t1452(병합 창 대기 단축), t1359(`gtd.md` 문서 정정 두 건)는 이 SPEC 의 범위가 아니다. M5 는 t1453 착지를 게이트로 기다리고 나머지는 같은 파일을 고치더라도 별도 카드다.

### Out of Scope — 분석의 자동 변형

- 분석기, `analyze`, `relate`, Jev 소견, 레인 세션이 카드를 접거나 지우거나 고치거나 순서를 바꾸는 기능은 만들지 않는다. 새 병합 동사는 운영자가 직접 호출하는 명시적 행위로만 존재한다.

### Out of Scope — 의미 기반 판정의 확대

- 발행 시점 제시는 어휘 겹침과 구조 단서만 쓴다. 제시 경로에서 모델을 호출하지 않으며, 기존 Jev 소비자(발행 시 기록 전용)는 그대로 둔다.

### Out of Scope — 과거 카드 소급 입력

- 기존 1,174장에 부모·출처·크기·예상 파일을 소급해 채우지 않는다. 닫힘 시각이 없는 과거 카드는 "unknown"으로 남는다.

### Out of Scope — 팩토리 선택 규칙 일반

- 직렬 슬롯의 의미, keep-set, 임대 경로의 다른 호(arm), 쿼터 보류, 백엔드 건너뛰기는 묶음에 필요한 최소 변경(D14) 밖에서 바꾸지 않는다.

### Out of Scope — 배포·문서 표면

- 릴리스·CI 워크플로, README 4종, docs-site 는 이 SPEC 의 범위가 아니다. 사용자 문서 정정은 sync 단계의 후속 항목이다.

## §G 공백과 잔여 위험

측정하지 못했거나 판정이 열려 있는 것을 숨기지 않는다(전체 목록은 `research.md` §10).

- 어휘 이웃의 **정밀도**는 측정하지 못했다. 정답 집합이 없고, 유일한 대리 지표인 기록된 관계 125쌍은 Jev·에이전트의 판정이라 재현율만 잴 수 있었다.
- "이미 덮는 완료 SPEC" 조회는 약한 신호다. SPEC 1,029개 중 frontmatter 에 `card:` 가 있는 것은 32개뿐이고, 카드→SPEC 링크(`spec_id`)는 큐 1,174장 중 7장에만 있다. 이 항목은 `heuristic` 표지를 달고 나간다.
- 진행 중 레인의 **예상 파일 겹침**은 오늘 입력 자체가 거의 없다. 예상 파일 필드는 어느 카드에도 없고, 열린 카드 37장 중 본문에 경로가 보이는 것은 12장(32%)이며 정확히 같은 경로를 공유하는 쌍은 0이다. 이 입력은 M2 데이터가 쌓일 때까지 대부분 `unmeasured` 로 나간다.
- git 탐침 비용: 레인 브랜치 하나의 `git diff --name-only` 가 약 0.25초, `WT-` 브랜치 353개·워크트리 74개가 존재한다. 진행 중 레인이 16개면 직렬로 약 4초다 — 제시 경로의 시간 상한과 병렬화는 M1 이 먼저 재야 한다.
- 카드→파일 간선을 만드는 `moai graph build` 의 실행 시간 증가는 측정하지 못했다.
- 웹 그래프의 레이아웃·번들 크기·브라우저 렌더 증거는 측정하지 못했다.
- t1453 이 같은 규칙 파일을 절체 시점에 고친다는 판단은 그 브랜치 끝(`2a5f9c91c`, 이동하는 ref)의 plan 문서를 읽은 것이다. 측정 시점에 여섯 규칙 파일과 `todo.go` 에 대한 차이는 비어 있었다.
- 잔여 위험: M5 가 게이트에 막혀 SPEC 이 요구는 충족하되 규칙은 아직 효력이 없는 상태로 닫힐 수 있다. 그때 리더가 M5 적용 카드를 발행해야 하며, 완료 보고가 이를 명시해야 한다.

## §H 상호참조

- `research.md` — 증거와 재측정 표, 스크립트, 공백 목록.
- `design.md` — 아키텍처, 데이터 흐름, 스키마, 관계 모델, 간선 운반체, 묶음 훅, D1~D15 선택지 표.
- `plan.md` — 마일스톤 분해, 게이트 명령, 위험, @MX 계획, 마일스톤별 검증 명령, 작업 기본값 요약.
- `acceptance.md` — 수용 기준, RED-now 원장, 변이 탐침.
- `decision-index.md` — 결정 열다섯 행.
- 인접 SPEC: SPEC-TODO-ANALYSIS-001, SPEC-TODO-AUTO-PICK-001, SPEC-RELATION-PICKUP-FILTER-001, SPEC-GTD-AUTONOMY-001, SPEC-WEB-TODO-QUEUE-001, SPEC-INSTRUCTION-BUDGET-SCOPE-001.
