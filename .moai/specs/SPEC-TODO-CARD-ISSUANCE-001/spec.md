---
id: SPEC-TODO-CARD-ISSUANCE-001
title: "카드 발행 품질 — 발행 시점의 겹침·중복 제시, 카드 관계 그래프, 묶음 직렬 경로, 과분할 억제 규칙"
version: "0.5.0"
status: draft
created: 2026-10-03
updated: 2026-10-04
author: MoAI
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/factory, internal/graph, internal/homestate, internal/web, internal/template/templates, .claude/rules, .claude/skills, .claude/agents"
lifecycle: spec-anchored
tags: "card-issuance, todo-queue, relation-graph, card-bundling, card-merge, issuance-rules, card-t1454"
tier: L
card: t1454
depends_on: [SPEC-TODO-ANALYSIS-001, SPEC-TODO-AUTO-PICK-001, SPEC-TODO-ARCHIVE-QUERY-001, SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-RELATION-PICKUP-FILTER-001, SPEC-GTD-AUTONOMY-001, SPEC-WEB-TODO-QUEUE-001, SPEC-TODO-SURFACE-POLISH-001]
related_specs: [SPEC-FACTORY-ATOMIC-LEASE-001, SPEC-TODO-CLAIM-LEASE-001, SPEC-TODO-TRANSITION-STAMPS-001, SPEC-JEV-CONSUMERS-001, SPEC-INSTRUCTION-BUDGET-SCOPE-001, SPEC-KANBAN-PR-CARD-TRACEABILITY-001]
---

# SPEC: 카드 발행 품질

## HISTORY

- 0.1.0 — 2026-10-03 — 카드 t1454(운영자 요청 10-02 밤, Class C, Tier L, 단일 카드 유지)의 plan 단계 산출물을 처음 작성했다. 근거는 세 개의 읽기 전용 렌즈와, 이 계획이 같은 트리 `2de0a2cb6`(전체 SHA `2de0a2cb613b04765a1554f86685a3b48e0be806`)에서 다시 잰 값이다. 카드가 인용한 측정 수치는 다른 세션의 scratchpad 에서 왔으므로, git 쪽 수치는 이 트리에서 재현했고(전부 일치) 큐 쪽 수치는 스냅숏 위에서 다시 쟀다(드리프트 기록). 카드 본문의 낡은 전제 여섯 가지는 §A.3 에서 바로잡았다. 결정 열다섯 가지는 `decision-index.md` 와 `design.md` §11 에 선택지·근거·작업 기본값으로 두었고, 이 SPEC 에는 미해결 질문 표지를 두지 않는다(카드의 리더가 감사 근거로 판정한다). 요구 24개, 수용 기준 24개(출시 차단 21, 회귀 가드 3).
- 0.2.0 — 2026-10-03 — 이터레이션 1 독립 plan-audit(FAIL 0.72, Tier L 기준 0.85, 보고서 `.moai/reports/t1454/plan-audit-iter1.md` — 로컬 전용·미추적이라 권위 인용이 아니라 출처 포인터다)의 지적 D1~D14 에 대응했다(감사 지적 번호이며 이 SPEC 의 결정 D1~D15 와는 별개다). 블로킹 일곱 건: (D1) 기준선이 `.gitignore:235` 의 `.moai/reports/*` 에 걸려 추적될 수 없었으므로 추적되는 `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` 으로 옮기고, 출하되는 허브 파일 목록은 M0 가 만들고 시험이 측정 명령과 대조하는 임베드 데이터 파일로 바꾸고, 완료 정의 #2 를 커밋 그래프가 증언할 수 있는 명령으로 다시 썼다. M5 초안 경로도 같은 이유로 추적 경로로 옮겼다. (D2) M5 게이트에서 `--first-parent` 를 없애 어느 부모 경로로든 닿는 술어로 바꾸고 고정 SHA 위의 두 번째 부모 양성 대조를 더했다. (D3) RED 행 L1·L23 을 감사가 만들 수 없는 추적 상태 질의로 옮겼다. (D4) 수치는 로컬 전용 규칙에만 쓰고 배포 사본은 메커니즘만 쓰는 것으로 한 가지로 정했다. (D5) REQ-TCI-013 을 실제 쓰기 동사(`relate`, `add`)가 위반할 수 있는 제약으로 좁혔다. (D6) 처분 컬럼을 finding 표 두 곳으로 넓혀 REQ-TCI-008·009 와 AC-TCI-008 에 넣었다. (D7) 출시 차단 17개 기준에 green-path 명령과 통과 시 출력의 모양을 적었다. 선택 지적: (D8) AC-TCI-021 을 조건부로 분류했다. (D9) 결정 기록 Q6 의 선호 문장을 지웠다. (D10) §A.3 에 일곱째 정정을 더했다. (D11) REQ-TCI-007·016 의 구현 세부를 설계로 옮겼다. (D12) M0 가 개수마다 측정 명령을 옆에 적는다. (D13) REQ-TCI-016 의 불변식을 (트리, 도달 가능한 이력) 위에서 서술하고 CI 와의 일치는 미검증으로 표시했다. (D14) 규칙이 발행 세션에 닿는지 보는 기준을 AC-TCI-021 에 더했다. 이 반복이 새로 찾은 것: 카드→파일 간선 층도 first-parent 만 걸으면 같은 맹점을 가지므로 REQ-TCI-016 과 AC-TCI-016 이 도달 가능한 모든 병합을 걷고 흡수 병합은 제외하도록 요구한다. 요구 24개, 수용 기준 24개(출시 차단 20, 조건부 1, 회귀 가드 3).
- 0.3.0 — 2026-10-03 — 이터레이션 2 독립 plan-audit(FAIL 0.82, Tier L 기준 0.85, 보고서 `.moai/reports/t1454/plan-audit-iter2.md` — 로컬 전용·미추적이라 출처 포인터일 뿐 권위 인용이 아니다)의 지적 D15~D25 에 대응했다(감사 지적 번호이며 이 SPEC 의 결정 D1~D15 와는 별개다). 블로킹 셋: (D15) 기준선 커밋 B 의 순서 증인이 B 보다 앞선 제품 커밋을 보지 못했다(`B^..HEAD` 는 `B^` 에서 닿는 커밋을 제외한다) — 증인 5 를 더했다: 카드 id 를 담고 `internal`·`cmd`·`.claude` 를 바꾸는 커밋이 `B^` 에서 닿는 이력에 0개여야 한다. 증인 3 의 틀린 문장을 고치고 증인 4 의 측정 트리를 기준선 머리의 `card-head:` 줄로 정의했다. (D16) REQ-TCI-002(c) 의 양성 겹침에 기준이 없었다 — AC-TCI-003 에 공유 경로 항목과 측정된 `none` 시험을, AC-TCI-002 에 렌더 확인을 더했고 "예상 파일"의 정의를 Module A 에 적었다. (D17) REQ-TCI-013 마지막 절에 기준이 없었다 — AC-TCI-013 에 `parent-of`·`follow-up-of`·`merged-into` 거절 시험을 더했다. 선택 지적: (D18) AC-TCI-008 에 속성 없는 카드가 SQL NULL 로 저장된다는 단언, (D19) 저장소 전역 ref 개수를 트리의 성질이 아니라 읽은 시각의 값으로 귀속, (D20) M5 게이트 선택자의 id 뒤를 앵커하고 흡수 방향 병합을 제외하는 판독을 더하고 고정 SHA 음성 대조를 만듦, (D21) AC-TCI-022(d) 를 `//`·`url(`·`@import`·`srcset` 까지 넓히고 한계를 명시, (D22) 출하 허브 목록의 "사용자 프로젝트에 일치 경로 없음" 전제를 검증하지 못한 수용 잔여로 바꿈, (D23) 허브 목록 적재의 동작 시험과 건너뜀 정책 시험을 추가하고 한계를 명시, (D24) 측정 트리 정의와 `ls` 셸 의존 문장 정정, (D25) 시험 패키지 귀속 정정. 이 반복이 새로 찾은 것: AC-TCI-021 (f) 가 `TestGoldenCommittedArtifactsMatchEmission` 을 `internal/template` 에서 찾았지만 그 시험은 `internal/template/agentemit` 에 있다(`internal/template` 의 `-list` 는 3개를 낸다 — `acceptance.md` AC-TCI-021 에서 두 패키지로 갈랐다). 문서 수준 핀을 `ad02a5677` 로 다시 고정하고 증거 행의 `HEAD`·`develop` 을 SHA 로 바꿨다. 요구 24개, 수용 기준 24개(출시 차단 20, 조건부 1, 회귀 가드 3).
- 0.4.0 — 2026-10-03 — 이터레이션 3 독립 plan-audit(FAIL 0.83, Tier L 기준 0.85, 보고서 `.moai/reports/t1454/plan-audit-iter3.md` — 로컬 전용·미추적이라 출처 포인터일 뿐 권위 인용이 아니다)의 지적 D26~D33 에 대응했다(감사 지적 번호이며 이 SPEC 의 결정 D1~D15 와는 별개다). 리더가 감사 상한에서 델타 감사 한 번(이터레이션 4)을 더 허용했으므로 범위를 바꾸지 않고 좁게 고쳤다. 블로킹 셋: (D26) AC-TCI-001 순서 증인 3 은 두 질의의 개수가 같으면 통과시켰지만 개수의 일치는 집합의 일치가 아니다 — 곁가지의 제품 커밋을 카드 id 가 적힌 병합으로 들이면 개수는 `2`·`2` 로 같은데 집합이 다르다(스크래치 저장소에서 실행해 확인). 증인 3 을 `--full-history` SHA 목록 비교로 바꾸고 완료 정의 #2·plan §F.12·design §12.4 를 맞췄다; 옛 형태가 정당한 이력에 거짓으로 실패하는 것도 실제 이력(카드 t1460 대역)으로 보였다. (D27) M5 게이트의 `A` 가 `--grep=absorb` 로 메시지 전체를 봐서 본문에만 `absorb` 가 든 진짜 착지(t1439)를 흡수 방향으로 세어 게이트가 닫힌 채 읽혔다 — `A` 를 병합 제목 줄에 `--grep` 하나로 걸고 t1439 를 양성 대조로 더했으며, 판독 2 의 점검도 같은 형태로 맞추고 `T ≥ 1` 인데 `T − A = 0` 이면 목록을 읽게 했다. (D28) AC-TCI-022 (b) 는 락 파일의 바이트·mtime 으로 락 획득을 증언했지만 advisory flock 은 그 파일에 아무것도 남기지 않는다 — 락을 쥔 채 `GET /todo?view=graph` 가 2초 안에 200 을 내는 시험(`TestTodoGraphViewDoesNotWaitOnQueueLock`)과 어휘 검사로 바꿨다. 선택 지적: (D29) "여섯 마일스톤"을 "일곱"으로, (D30) 겹침 음성 입력(파일 이름·문자열 접두사)과 입력이 섞인 경우의 집합 단위 판정을 AC-TCI-003 에, (D31) 템플릿 사본 수치 검사의 범위를 새 `card-issuance.md` 사본으로 좁혀 명시, (D32) 건너뜀 정책 시험이 측정 본문을 직접 구동하도록, (D33) 제품 경로 필터에 `pkg`·`scripts`·`.codex` 를 더했다. 이 반복이 더한 것: M5 압축이 카드 id 기준선 짝(`gtd.md t696` 등 넷)을 지우면 `card_id_leak_test.go` 의 기준선 시험이 붉어진다는 점을 plan 에 적었다. 요구 24개, 수용 기준 24개(출시 차단 20, 조건부 1, 회귀 가드 3).
- 0.5.0 — 2026-10-04 — 이터레이션 4 독립 plan-audit(FAIL 0.84, Tier L 기준 0.85, 보고서 `.moai/reports/t1454/plan-audit-iter4.md` — 로컬 전용·미추적이라 출처 포인터일 뿐 권위 인용이 아니다)의 지적 D34~D38 에 대응했다(감사 지적 번호이며 이 SPEC 의 결정 D1~D15 와는 별개다). 리더가 두 번째 감사 상한에서 카드를 보류했다가 재개하며 요구한 것은 지적 수리와 develop `30ce3a02d` 흡수에 맞춘 재기반(트리 `68a4d8137`)이고, 이 수리 뒤 독립 감사 한 번이 남는다 — 범위는 바꾸지 않았다. 블로킹 둘: (D34) 증인 5 가 기본 이력 단순화라 곁가지에서 제품 커밋을 넣었다 되돌려 B 앞에서 병합하면 증인 1~5 가 모두 통과했다 — 스크래치 세 변형(AUD1~AUD3)에서 확인하고 `--full-history` 를 더하되, 단독형은 이 카드의 실제 이력(M0 앞의 develop 흡수 병합 `68a4d8137` 이 카드 id 를 담는다)에서 정당한 M0 를 거짓으로 실패시켜(C42: `1`) `--full-history --simplify-merges` 로 채택했고(C43: `0`) 변이 MU-123 으로 이름 붙였다; (D35) 증인 3 이 B 뒤 제품 커밋 목록의 비어 있지 않음을 요구하는데 plan 이 AC-TCI-001 을 M0 종료 증거로 삼아 M0 종료에서는 항상 거짓으로 실패했다 — 순서 증인 1~5 의 읽는 시점을 병합 전 한 번으로 정하고 M0 종료 증거를 AC-TCI-001 (a)~(e) 로 갈랐다(spec·plan·design·acceptance·spec-compact 가 같은 문장). 선택 지적: (D36) REQ-TCI-021 의 착륙 정의를 게이트가 실제로 쓰는 어휘 규칙으로 적고 거짓 열림·거짓 닫힘 비율(SHA 로 재측정)과 `T ≥ 1`·`T − A = 0` 목록의 처분을 더했다, (D37) 측정 시험 래퍼의 소스 읽기 단언, (D38) 진행 중 레인 카드 수 문구와 설계 표 위치. 재기반: `internal/kanban` → `internal/factory`, `kanban-dispatch*.md` → `factory-dispatch*.md`(카드 t1399)와 줄 번호·크기·미러(생성 플러그인 페이로드 하나가 늘었다)·게이트·핀·가드 목록을 다시 읽어 고쳤다. 흡수로 죽은 전제는 없었고 이터레이션 4 까지의 서술 둘이 처음부터 틀렸음을 바로잡았다(루트 `.codex/` 는 추적되지 않는다; `factory-dispatch.md` 의 분기는 `TestAutoPickMirrorParity` 가 가드한다). 요구 24개, 수용 기준 24개(출시 차단 20, 조건부 1, 회귀 가드 3), 변이 123개.

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

**A3 — SPEC-TODO-AUTO-PICK-001 (v0.5.0)**

- **직전 완료 버전**: 0.5.0 — 카드 t1458 의 SPEC-FACTORY-ATOMIC-LEASE-001 이 이 SPEC 을 `completed → in-progress → completed` 로 한 번 개정해 `## Amendments` 절을 이미 만들었다(0.4.1 → 0.5.0, 대체·강화된 조항의 기록). 이 SPEC 의 개정 행은 그 절에 둘째 항목으로 덧붙는다
- **prior_completed_sha**: `4293b2d73` — 이터레이션 5 가 읽은 대상 `progress.md` §E.4 의 `sync_commit_sha` 다(0.5.0 재종료 뒤에도 같다). sync 시점에 다시 읽는다
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

표의 값은 머리에 적은 트리와 큐 스냅숏 시각에 읽은 측정이고 현재 값이 아니다 — 큐 쪽 값은 계속 움직인다. git 쪽 값도 develop 흡수(`68a4d8137`) 뒤에 다시 재지 않았다 — M0 가 실행 시점 트리에서 다시 잰다. 이 SPEC 의 임계값은 이 표가 아니라 M0 가 새로 만드는 기준선 기록에서 읽는다(REQ-TCI-001).

카드의 숫자를 그대로 믿지 않고 다시 잰 결과 두 가지가 새로 드러났다. 첫째, 기계 분석기의 near 구간(token-set Jaccard 0.80 이상 1.0 미만)은 큐 전체 역사에서 한 번도 발화하지 않았다 — 근접 이웃이 그 구간에 있는 카드가 0장이고, 기록된 `near-duplicate` 소견 107건은 전부 Jev 가 판정했다. 둘째, 어휘 겹침은 약한 변별자다. 발행 시점 가정에서 가장 가까운 이전 카드의 점수 중앙값이 0.10 이고, Jev·에이전트가 기록한 관계 상대가 어휘상 상위 3개 안에 드는 비율은 하한 없이 52%, 하한 0.30 에서 6%다. 따라서 M1 의 제시는 "확실한 것"(정확 중복, 보관·dropped 카드 포함)과 "참고용"(어휘 이웃)을 구분해서 보여야 하고, 어휘 이웃이 의미 중복 판정을 대체한다고 약속해서는 안 된다.

### A.2 이 SPEC 이 하는 일

일곱 마일스톤(M0~M6)을 순서대로 한다(실행 순서: M0, M1, M2, M3, M4, M6, 그다음 게이트가 열렸을 때 M5).

- **M0** 카드의 기준선을 **추적되는** 위치(`.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/`)에 반입하고 실행 단계에서 다시 잰다. 출하되는 허브 파일 목록도 같은 커밋이 만든다.
- **M1** `add` 가 막지 않고 알린다 — 유사 카드 상위 3개(보관·dropped 포함), 같은 구성요소의 열린 카드, 진행 중 레인과의 예상 파일 겹침, 이미 덮는 완료 SPEC. `--dry-run` 을 지원한다.
- **M2** 카드에 발행 속성(스폰한 카드·출처·크기 추정·예상 파일·drop 사유)을 가산적으로 싣고 finding 에 처분(수용·병합·기각)을 싣는다. 카드 표 두 곳과 finding 표 두 곳 모두에 같은 이주·보관·parity 규율을 적용한다.
- **M3** 관계를 일곱 종류의 한 어휘로 묶고(종류별 순환·카디널리티 제약), 두 저장소를 해석기 하나로 읽고, 추이 질의 동사와 `moai graph` 의 card→file 간선을 만든다.
- **M4** 묶음 직렬 경로, 운영자 호출 카드 병합 동사, 허브 파일 직렬 소유.
- **M5**(게이트됨) 카드 크기·후속 지적·파생 깊이·동시 진행 한도·발행 체크리스트 규칙. 선행 조건은 명령으로 판정한다.
- **M6** `moai web` `/todo` 에 관계 그래프 보기.

### A.3 바로잡은 낡은 전제

카드 본문과 렌즈 요약 일부가 이 트리의 현실과 어긋나 있었다. 이 SPEC 은 아래 정정을 기준으로 쓴다.

1. 카드는 near-duplicate 기록이 `todo.go:800` 에 있다고 적었다. 실제 위치는 `internal/cli/todo_analysis.go:74`(`case factory.BacklogMatchNear`)이고 임계값은 `internal/factory/backlog_analysis.go:33`(`BacklogNearDuplicateThreshold = 0.80`)이다.
2. t1349 의 세 항목(`show` 동사, `add` 의 대시 시작 본문 파싱, list 기본 상한)은 SPEC-TODO-SURFACE-POLISH-001(completed, 커밋 `1f7eeae49`)이 이미 끝냈다. 이 SPEC 의 범위가 아니다.
3. 카드는 `assign --after` 를 todo 동사처럼 적었다. 실제로는 `moai factory assign` 이다(`internal/cli/factory_card.go:1885` 부근에서 `--prefer`/`--after`/`--spec` 를 읽는다).
4. 카드는 "현 규칙은 병합 금지"라 했지만 `factory-dispatch-detail.md`·`factory-dispatch-mechanics.md`·`sync-auditor.md`·`manager-todo.md` 에는 그런 문장이 없고, 상시 로드 `factory-dispatch.md:37` 에 리더에 대한 한 문장이 있을 뿐이다(소견은 증거일 뿐 — 리더는 관련 카드를 접지 않고 순서를 바꾸지 않고 지우거나 고치지 않는다). 분석 쪽 금지는 `.claude/skills/moai/workflows/gtd.md` 의 두 문장과 병합 동사가 없다는 코드 형태가 만든다(A2 참조). 이터레이션 4 까지는 첫 문장도 "어디에도 없다"고 적었으나 틀렸다(`factory-dispatch.md:37` 은 개명 전 `kanban-dispatch.md:37` 이었다). 두 문장 — 이 줄과 `gtd.md` 의 "Analysis never folds one card into another, …" — 은 `todo_auto_doc_test.go` 가 문장 전체를 고정하므로(`gtd.md` 쪽은 `acceptance.md` G33) 개정은 그 문장을 지우거나 고치지 않고 뒤에 예외 문장을 덧붙인다.
5. t1448 은 이미 이 트리에 착지했다(병합 `4315f0d0e` 는 HEAD 의 조상). t1453 은 아직 착지하지 않았다 — 게이트 명령이 0 을 읽는다(§C Module E, `acceptance.md` L21; 흡수 뒤 핀 `68a4d8137` 에서도 같다).
6. 카드는 `moai graph` 의 "커밋된 `edges.jsonl`"을 전제했지만 `.moai/project/graph/` 는 `.gitignore:344` 로 추적 제외된 파생 산출물이다(`internal/graph/check.go` 도 "untracked derived artifact"라 적는다). 결정론 요구는 그대로이고, "비공개 큐 상태를 커밋된 산출물에 섞지 말 것"은 "`moai graph build` 의 동일 트리 동일 출력 계약을 지킬 것"으로 바뀐다(D3).
7. 카드 본문은 발행 시점 제시가 기존 분류 경로(`ClassifyCardText`)와 기호 추출 도우미(`todoTriageSymbols`)를 **재사용**한다고 적었다. 이 SPEC 은 **둘 다 재사용하지 않는다** — 분류기는 dropped 카드를 거절 대상에서 건너뛴다는 독트린과 고정 시험을 가지므로 건드리지 않고(REQ-TCI-006), 기호 추출 도우미는 최대 4개만 내는 상한 때문에 경로 겹침 용도에 맞지 않는다(`research.md` §4). 대신 제시용 조회를 별도 읽기 전용 경로로 두고 척도(`NormalizeCardText`·`TokenSetJaccard`)만 재사용한다.
8. 이터레이션 1 계획은 기준선과 M5 초안을 `.moai/reports/t1454/` 아래 "추적되는 형태"로 둔다고 적었지만 그 경로는 `.gitignore:235`(`.moai/reports/*`)에 걸려 추적될 수 없다(`git check-ignore -v` 가 종료 0 으로 확인). 감사 산출물 규약(`.moai/docs/audit-artifact-convention.md`)은 보고서를 트리에 강제로 넣거나 무시 규칙을 넓히는 것을 금하므로, 이 SPEC 의 기준선과 초안은 추적되는 SPEC 디렉터리 안으로 옮겼다(REQ-TCI-001, -022).

### A.4 흡수가 바꾼 위치와 전제 (이터레이션 5)

이터레이션 4 까지의 문서는 develop `2de0a2cb6` 위에서 썼다. 이터레이션 5 는 develop `30ce3a02d` 를 흡수한 트리 `68a4d8137` 에서 모든 인용을 다시 읽었다. 바뀐 것과 그대로인 것:

1. **개명(카드 t1399).** 패키지 `internal/kanban` 은 `internal/factory` 로(M8 `8a7d60ba2`; 락 파일 `board_*.go` → `state_lock*.go`, 락 도우미 이름 `boardLock*` → `stateLock*`), 디스패치 규칙 셋은 `kanban-dispatch{,-detail,-mechanics}.md` → `factory-dispatch*.md` 로(M10 `859477bc7`) 바뀌었다. 환경 키 `MOAI_KANBAN`·`MOAI_KANBAN_LABEL` 은 사라졌다. 이 SPEC 은 현재 이름을 쓴다.
2. **크기.** 상시 로드 `factory-dispatch.md` 는 27,744자·27,948바이트(템플릿 사본 27,423자), `factory-dispatch-detail.md` 40,659자(659자 초과), `gtd.md` 41,160자(1,160자 초과 — 이터레이션 4 까지는 37자)다. M5 가 `gtd.md` 에서 줄여야 하는 양이 커졌다.
3. **M5 미러 표면이 하나 늘었다.** `plugins/moai/skills/moai/workflows/gtd.md` 는 템플릿 `gtd.md` 의 생성된 바이트 사본이다(`make plugin-emit`, `internal/template/pluginemit` 의 `TestGoldenCommittedArtifactsMatchEmission`·`TestCommittedVersionMatchesSSOT`). 반대로 이터레이션 4 까지 "생성 Codex TOML(저장소 루트와 템플릿 하위)"이라 적은 것 중 루트 `.codex/agents/moai/*.toml` 은 처음부터 `.gitignore:137` 로 추적되지 않았다 — 추적되는 것은 템플릿 하위뿐이다.
4. **M5 가드 둘이 더 걸린다.** `TestAutoPickMirrorParity`(`internal/cli`)는 갈라진 쌍 `factory-dispatch.md` 가 정확히 한 줄만 다르다는 것을 고정한다 — 이 쌍의 분기가 가드 없이 서 있다는 이터레이션 4 까지의 서술은 틀렸다(가드 없이 서는 분기는 `sync-auditor.md` 쪽이다). `todo_auto_doc_test.go` 는 `gtd.md` 의 "Analysis never folds one card into another, …" 문장 전체를 고정한다 — 개정 행 A2 가 고치려는 문장이다.
5. **대상 SPEC 상태.** SPEC-TODO-AUTO-PICK-001 은 0.4.1 에서 0.5.0 으로 한 번 개정됐다(개정 행 A3 의 직전 완료 버전은 0.5.0).
6. **리스 경로.** SPEC-FACTORY-ATOMIC-LEASE-001(completed)이 `factory next` 선택 호 전체를 큐 락 아래 한 임계 구역에 넣었다(REQ-FAL-001·-007·-010). M4 의 묶음 예약·허브 체인은 그 구역 안의 결정 읽기이고 git 프로세스·큐 쓰기를 더하지 않으며 `internal/homestate` 비시험 파일이 `internal/factory` 에 의존하지 않는다는 조건을 지킨다(`design.md` §7.2).
7. **겹치는 카드.** t1480(picked — 큐·기록 통합, `factory dispatch`, 직렬 슬롯 의미)은 M1·M2·M3·M4 와 같은 파일·같은 의미를 만진다 — 의존 메모는 `plan.md` §D. 범위는 바꾸지 않았다.
8. **그대로인 것.** M2 스키마 층(`backlog_*`·`todo_queue_merge.go`·`todo_merge_procedure.go`)과 M3 그래프 층(`internal/graph`)의 인용 줄은 개명 외에 같다; `todo.go` 의 `runTodoAddAppendRoot` 는 `:787` 에서 `:796` 으로, `factory_card.go` 의 선택·임대 함수들은 `+33` 줄에서 `+182` 줄까지 밀렸다(`research.md` §0 표).
9. **쓰기 전 LLM 분류 단계.** SPEC-TCD-LLM-DECIDER-001 의 `todoPreClassifyLLM`(`internal/cli/todo_classify_llm.go:202`)은 `add` 경로에서 큐 락을 잡기 전에 한 번 호출되고(`todo.go:812`·`:871`) 판정을 정적 결정기에 실어 락 안으로 들인다. M1 의 분석·제시와 M2 의 `add` 플래그는 이 단계 뒤의 결정기를 받으며, 락 안에 네트워크 왕복을 더하지 않는다(그 SPEC 의 REQ-TLD-005 가 지키는 조건이다). 범위는 바꾸지 않았다.

## §B 결정과 작업 기본값

열다섯 결정은 `decision-index.md`(Detect → Explain → Ask, 선호 답 없음)와 `design.md` §11(선택지·측정 근거·작업 기본값)에 있다. 아래 표의 "작업 기본값"은 계획이 이 값으로 진행한다는 뜻이지 판정이 아니다 — 카드의 리더가 감사 근거로 판정하고, 판정이 다르면 해당 요구와 기준의 리터럴만 다시 고정한다.

| ID | 결정 | 계획이 진행하는 작업 기본값 |
|---|---|---|
| D1 | M5 의 순서와 게이트 미충족 시 인도물 | M5 는 마지막 마일스톤이고 게이트 명령(어느 부모 경로로든 닿는 조상 술어)이 판정한다. 미충족이면 규칙 파일은 건드리지 않고 추적되는 초안과 후속 카드 문안만 낸다 |
| D2 | 두 관계 저장소 통합 방식 | 저장소는 분리 유지, 읽기 해석기와 어휘 확장으로 통합 |
| D3 | card→file 간선 운반체 | 커밋된 증거(카드 귀속 병합 커밋)에서만 `edges.jsonl` 에 간선 생성, 비공개 사이드카는 만들지 않는다 |
| D4 | M2 스키마 형태 | nullable JSON 컬럼 하나(`issuance`), finding 은 컬럼 하나(`disposition`) |
| D5 | MCP 알림 전달과 `gtd engage` | MCP 결과 텍스트의 id 줄 뒤에 제시를 덧붙이고, engage 는 제시만 하고 분석기는 돌리지 않는다 |
| D6 | 병합 경로 독트린 개정 범위 | 운영자 호출 동사 하나로 한정, 분석·relate·레인은 계속 금지 |
| D7 | 수치 임계값 | SPEC 은 값을 고정하지 않는다. 값은 M0 의 추적되는 기준선 기록에서 읽고, 값 자체는 로컬 전용 규칙에만 쓰며 배포 사본은 메커니즘만 쓴다 |
| D8 | 유사도 척도와 소음 | 기존 token-set Jaccard 재사용, 표시 하한을 두고 점수와 척도를 항상 함께 표기 |
| D9 | closed-at | 저장하지 않고 기존 두 스탬프에서 읽는 접근자 하나 |
| D10 | 파생 부모의 단일 원천 | 카드 속성이 원천, `parent-of`/`follow-up-of` 는 읽기 투영 |
| D11 | finding 처분의 효과 | 기록만 한다, 카드·순서·픽업 필터에 영향 없음 |
| D12 | 구성요소별 부채 대장 운반체 | 큐의 카드(출처 `debt`)와 구성요소 노드로 표현, 새 저장소 없음 |
| D13 | `add` 새 플래그의 입력 방식 | 명시 플래그만, 본문에서 추론해 저장하지 않는다 |
| D14 | 묶음 메커니즘과 임대 경로 개정 | 팩토리 레코드의 묶음 컬럼과 선택 호의 묶음 예약(임대 경로 개정 인정) |
| D15 | 신규 경로 한정 파일의 40,000자 한도 | 이미 정해짐 — SPEC-INSTRUCTION-BUDGET-SCOPE-001 |

## §C 요구사항 (GEARS)

요구는 다섯 모듈이다. GEARS 표기를 쓰고 모든 `REQ-TCI-NNN` 은 `acceptance.md` 의 `AC-TCI-NNN` 하나 이상이 추적한다. 요구 문장은 리포의 한국어 SPEC 관행대로 GEARS 영어 문형으로 적고, 설명과 근거는 한국어로 적는다. 요구는 관측 가능한 행동만 말한다 — 저장 경로·순회 방식 같은 구현 방법은 `design.md` 가 소유한다. 숫자 임계값은 이 문서에 쓰지 않는다 — 값은 M0 가 만드는 기준선 기록에서 읽고(REQ-TCI-001), 계획의 작업 기본값은 `plan.md` §F.11 과 `design.md` §11 에 있다.

### Module A — 기준선과 발행 시점 제시 (M0, M1)

- **REQ-TCI-001** (Ubiquitous): The baseline record under `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/` — a tracked path that is present in every clone and in CI — shall hold, for every figure on which a threshold or rule value of this SPEC rests, the producing command, its verbatim output and the tree SHA it was measured on, and every later requirement or rule that names a threshold shall take its value from that record rather than from this document, and the commit that adds the record shall carry nothing outside it and shall precede every commit that changes shipped behavior.
- **REQ-TCI-002** (Event-driven): **When** a card is admitted through `moai todo add`, `moai gtd add` or the bare `moai todo <text>` fallthrough, the add path shall print on its error stream a presentation of (a) up to three cards most similar to the new text across live, dropped (reason prefix stripped) and archived cards, (b) open cards sharing a component with it, (c) in-flight lane cards whose expected files overlap its own, and (d) completed SPECs that already cover it, each item at or above its display floor and labelled with its source and measure, and its output stream shall remain exactly `<id> <position>` followed by a newline.
- **REQ-TCI-003** (Unwanted): The presentation shall not refuse, delay beyond its time bound, reorder, edit, fold or drop any card or otherwise change the admission, so that the queue after an add is identical to the queue the same add would produce without the presentation apart from the new card's own row and the findings the analyser already records, and any probe that reads outside the queue (git, the SPEC directory) shall run outside the queue's cross-process write lock and under a time bound.
- **REQ-TCI-004** (Capability-gate): **Where** `--dry-run` is passed to `moai todo add` or `moai gtd add`, the add path shall print the same presentation, shall write nothing (the queue file byte-identical, no id consumed, `last_seq` unchanged) and shall exit 0, and for an exact duplicate shall report that a real add would refuse instead of refusing.
- **REQ-TCI-005** (Event-driven): **When** a card is admitted through the MCP `todo_add` tool or through `moai gtd engage`, the admission shall carry the same presentation — in the tool result's text after its first `<id> <position>` line for MCP, on the error stream for engage — without recording any finding on the engage path and without changing the tool result's first line.
- **REQ-TCI-006** (Ubiquitous): The lookup that feeds the presentation shall be a read-only path separate from the duplicate classifier, so that the classifier, the text normalizer, the Jaccard measure and the near-duplicate threshold shall keep their behavior and their tests, including the classifier's skip of dropped cards.

용어 — **예상 파일**. REQ-TCI-002(c) 의 겹침 제시에서 카드의 예상 파일은 (i) 본문이 이름으로 대는 저장소 경로 토큰(읽기 전용 추출이고 저장하지 않는다 — D13)과 (ii) M2 이후의 `files` 속성의 합집합이고, 진행 중 레인 카드는 여기에 (iii) 그 레인 브랜치의 변경 파일(시간 상한 안의 git 탐침)이 더해진다. 겹침 항목은 양쪽 입력이 모두 있을 때만 측정값(겹치는 경로마다 한 항목, 겹침이 없으면 `none`)이고, 한쪽이라도 입력이 없으면 `none` 이 아니라 `unmeasured` 다. 판정은 진행 중 카드 **집합** 단위다 — 후보가 경로를 대고 입력이 있는 진행 중 카드가 하나라도 있으면 입력이 있는 카드만 비교하고(입력이 없는 진행 중 카드는 비교에서 빠지며 `none` 으로도 `unmeasured` 줄로도 보고하지 않는다), 비교할 입력이 하나도 없을 때만 `unmeasured` 다. 경로의 일치는 정규화한 저장소 경로 **전체**의 일치이고 파일 이름이나 문자열 접두사의 일치가 아니다. REQ-TCI-020 의 허브 체인은 (ii)만 읽는다 — 본문 추출만으로는 체인을 만들지 않는다.

용어 — **순서 증인의 읽는 시점**(이터레이션 5). 순서 증인 1~5 는 병합 전에 한 번만 읽는다 — 마지막 마일스톤이 끝난 뒤, 카드 브랜치가 통합 브랜치에 병합되기 전(sync-audit 이 증거를 읽는 시점). M0 종료에서는 읽지 않는다: 증인 3 이 B 뒤의 제품 커밋 목록이 비어 있지 않을 것을 요구하는데 M0 종료에는 제품 커밋이 하나도 없어 두 목록이 비기 때문이다(스크래치 S11). M0 종료 증거는 AC-TCI-001 (a)~(e) 다 — 기준선 파일과 그 안의 줄만 읽으므로 제품 커밋이 없는 트리에서 판정된다.

### Module B — 카드 스키마 (M2)

- **REQ-TCI-007** (Ubiquitous): Every card, live or archived, shall be able to carry optional issuance attributes — the card that spawned it, its origin, a size estimate, its expected files and a drop reason — where absence is represented as absent rather than as an empty value, so that a card carrying none of them is stored, serialized and exported exactly as before this change.
- **REQ-TCI-008** (Event-driven): **When** a queue database written before this change is opened, the engine shall add the missing storage for issuance attributes to both card-bearing tables and the missing storage for dispositions to both finding-bearing tables without losing or altering any existing row, and the pure reader shall return absent values for a database that lacks that storage without running any schema change.
- **REQ-TCI-009** (Ubiquitous): Archiving a card and restoring it shall carry its issuance attributes and the dispositions of its findings unchanged, a merge of two queues shall carry them unchanged, and the in-memory record, the JSON rendering and the SQLite rows of one queue shall agree on them, the engine's parity assertion covering all the new storage.
- **REQ-TCI-010** (Event-driven): **When** a card is dropped, `moai todo drop` shall store the reason in the drop-reason attribute while continuing to write the `[DROPPED — <reason>] ` text prefix, and the closing time of a card shall be readable through one accessor that derives it from the archive or drop stamp and answers "unknown" for a card that carries neither.
- **REQ-TCI-011** (Ubiquitous): A finding shall be able to carry an optional disposition — accept, merge or reject — set only through an explicit operator verb, absent for every existing finding, and recording a disposition shall change no card, no finding relation and no queue order.

### Module C — 관계 모델과 그래프 (M3)

- **REQ-TCI-012** (Ubiquitous): Card relations shall be expressed in one vocabulary of seven kinds — blocks, duplicates, parent-of, follow-up-of, merged-into, supersedes, relates-to — over nodes of type Card, Component, File, Commit, Spec and Finding, with every legacy finding relation mapped onto it at read time (depends and blocks to blocks, near-duplicate and duplicate-forced to duplicates, replaces to supersedes, contains, absorbs and conflicts to relates-to with the legacy name kept as a qualifier) and no stored row rewritten.
- **REQ-TCI-013** (Event-detected): **When** `moai todo relate` is asked to record a self-edge or a cycle among `blocks` pairs (a legacy `depends` record counting as one) or among `supersedes` pairs (a legacy `replaces` record counting as one), **or when** `moai todo add` receives a second `--parent` or `--origin`, a `--parent` that names no card, or an `--origin` outside the closed set, the verb shall refuse with a message naming the kind or flag and the offending pair or value and shall leave the queue file byte-identical, a `duplicates` or `relates-to` pair recorded again in the opposite order shall create no second record, and `parent-of`, `follow-up-of` and `merged-into` shall be written by no relation verb — the first two exist only as projections of the add-time attributes and `merged-into` only through `moai todo merge`.
- **REQ-TCI-014** (Ubiquitous): One resolver shall map a queue card id to its GTD item id and back through the engagement link, so that a single read can show a card's findings together with its `gtd_relations`; the resolver shall add no write path to `gtd_relations` and shall leave `moai gtd organize` unchanged.
- **REQ-TCI-015** (Event-driven): **When** `moai todo trace <id> [--kind <k>] [--depth <n>]` runs, the command shall print every node reachable from the card over the named kinds up to the depth bound in a deterministic order, shall terminate on cycles that legacy data may contain, shall write nothing and shall be available to a lane session like the other read-only verbs.
- **REQ-TCI-016** (Ubiquitous): `moai graph build` shall derive card-to-file edges only from committed evidence — merge commits whose subject attributes them to exactly one card and that are reachable from HEAD by any parent path, and the files each such merge brought in — so that two builds over the same tree and the same reachable commit history produce byte-identical output, the output contains no queue-private state (no unlanded card id, expected file or finding), an absorb merge contributes no edge, and `moai graph check` shall notice a change to that evidence source.
- **REQ-TCI-017** (Ubiquitous): For every record that predates this change, `todo list`, `todo list --json`, `todo why`, `todo export`, the pickup filter and the auto-rank near-duplicate path shall behave and render exactly as before.

### Module D — 묶음과 병합 (M4)

- **REQ-TCI-018** (Event-driven): **When** cards are assigned to one bundle, the factory shall lease them to the bundle's lane one at a time in bundle order, offering the next member to that lane after the previous member's local merge and withholding members from every other lane, while the fleet-wide serial slot keeps its present meaning.
- **REQ-TCI-019** (Event-driven): **When** an operator runs `moai todo merge <into> <from>`, the command shall append the text of `<from>` to `<into>` as an explicit section, record a merged-into relation from `<from>` to `<into>`, drop `<from>` with its reason, and refuse when either card is picked, `<from>` is already merged, `<into>` is closed or already merged, or the pair would form a cycle; and neither the analyser, `analyze`, `relate` nor any lane session shall be able to invoke it.
- **REQ-TCI-020** (State-driven): **While** a hub file — a path in the hub-file list that the baseline step produced and the product ships as an embedded data file — appears in the explicitly recorded `files` attribute of two open cards, the factory shall order them as a bundle chain so that the second is leased only after the first has merged, the keep-set shall continue to read no file overlap, and no shipped code path shall read a SPEC directory or a reports path to obtain that list.

### Module E — 규칙과 웹 보기 (M5, M6)

- **REQ-TCI-021** (Capability-gate): **Where** the M5 gate holds — a merge commit whose subject line names card t1453 and does not contain the word `absorb` — the lexical landing rule the gate commands apply, whose two false directions are measured in §G — is reachable from the working tree's HEAD through any parent path (or the leader's recorded reading of a listed subject stands in for it), read by the gate commands together with their positive controls — the card shall add a path-scoped issuance rule that states, as mechanism only, a card-size floor and ceiling, the follow-up rule (a defect a card creates is fixed inside that card before merge, a pre-existing defect is entered in the per-component debt ledger), a derivation-depth cap, a concurrent-in-flight cap and an issuance checklist, shall write each numeric value only in a local-only rule and take it from the baseline record, shall make the rule reachable from the always-loaded dispatch rule, and shall stay within the instruction budget (the always-loaded `factory-dispatch.md` without net growth, every touched file within 40,000 characters or no larger than before) with every mirror guard green and no card id, SPEC path, report path or measured value in any template copy.
- **REQ-TCI-022** (Capability-gate): **Where** the M5 gate does not hold when M5 would start, the card shall leave the six rule files and their template copies unedited, shall deliver the rule text as a ready-to-apply draft under the tracked path `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/` with exact insertion anchors, and shall record in `progress.md` the gate command output and the follow-up card text the leader is to issue.
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
- [NEW] 추적되는 기준선 기록 `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/`(기준선, 재현 스크립트, 허브 파일 목록)과 그 재현 절차.

### [DELTA] 카드 스키마 (Module B)

- [EXISTING] `ensureColumn` 가산 컬럼 패턴, `backlogItemsTableColumns` 의 v1→v2 재구성 목록, `todoJSONProjection`, JSON 골든.
- [MODIFY] `BacklogItem`·`BacklogFinding`·`BacklogArchivedFinding`·`BacklogArchiveEntry` 와 SQLite 읽기·쓰기·parity 경로(카드 표 둘과 finding 표 둘 모두), 동결 컬럼 튜플 테스트(finding 표 튜플 고정은 새로 더한다), 큐 병합 복사 경로.
- [MODIFY] `moai todo drop` — 사유를 속성에도 저장한다(텍스트 접두사는 그대로).
- [NEW] 발행 속성(스폰한 카드·출처·크기 추정·예상 파일·drop 사유), finding 처분, 닫힘 시각 접근자, `add` 의 명시 플래그와 그 검증.

### [DELTA] 관계 모델과 그래프 (Module C)

- [EXISTING] 소견 어휘 여덟 종류와 `WaitsOnOf`·`FindingsBlocking`·`WaitsOnClosesCycle`, `gtd_relations` 아홉 종류와 `ValidateGTDRelation`, 비공개 `gtd-edges.jsonl` 투영, 커밋 제목 귀속의 단일 지점 `subjectAttribution`.
- [MODIFY] `todo relate` — 새 쓰기 종류(`duplicates`, `supersedes`, `relates-to`)와 처분 동사, 종류별 제약 검사. `moai graph build`·`check` — 새 간선 층과 출처 지문.
- [NEW] 일곱 종류 어휘와 읽기 시 매핑, 카드↔GTD 해석기, `moai todo trace`, card→file 간선 층.
- [REMOVE] 없음. 저장된 행은 하나도 다시 쓰지 않는다.

### [DELTA] 묶음과 병합 (Module D)

- [EXISTING] 직렬 슬롯, `assign --after` 의 이전 카드 병합 가드, keep-set 의 "파일 겹침을 읽지 않는다".
- [MODIFY] 팩토리 `cards` 행(묶음 컬럼 추가)과 임대 선택 호.
- [NEW] `moai factory bundle`(또는 동등한 적재 동사), `moai todo merge`, 허브 파일 체인 생성, 임베드되는 허브 파일 목록 데이터 파일과 그 적재.

### [DELTA] 규칙과 웹 (Module E)

- [EXISTING] 카드 등급 A/B/C 문장, Class A 크기 문구, "one card per worktree", sync-auditor 의 blocking/optional 분류와 `PASS-WITH-DEBT`, 템플릿 미러와 그 가드들.
- [MODIFY] `gtd.md`(+미러), `factory-dispatch.md`(+분기된 미러), `factory-dispatch-detail.md`(+미러), `sync-auditor.md`(+분기된 미러, 생성 Codex TOML), `manager-todo.md`(+미러, 생성 Codex TOML), `gtd.md` 의 생성 플러그인 페이로드 사본, `catalog.yaml` 해시, 경로 고정 시험의 목록.
- [NEW] `card-issuance.md`(경로 한정 규칙, 메커니즘만, +미러), 수치를 싣는 로컬 전용 규칙 `.claude/rules/local/card-issuance-thresholds.md`(미러 없음), 웹 관계 그래프 보기와 자산.

## §E 수정 대상 파일

M0: `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/baseline/`(신규, 이 디렉터리 밖은 건드리지 않는다).
M1: `internal/cli/todo.go`, `internal/cli/todo_issuance.go`(신규), `internal/cli/mcp_todo.go`, `internal/cli/gtd.go`, `internal/factory/backlog_issuance.go`(신규)와 각 `_test.go`.
M2: `internal/factory/backlog_store.go`, `backlog_sqlite.go`, `backlog_migrate.go`, `backlog_schema_freeze_test.go`, `internal/factory/todo_queue_merge.go`·`todo_merge_procedure.go`(항목·소견 복사 경로), `internal/cli/todo.go`, `todo_drop.go`, `todo_claim.go`(JSON 투영), `todo_export.go`.
M3: `internal/factory/backlog_relation.go`(신규), `backlog_store.go`, `internal/cli/todo_relate.go`, `todo_trace.go`(신규), `internal/graph/card_file.go`(신규), `graph.go`, `meta.go`, `internal/cli/graph.go`.
M4: `internal/homestate/factory.go`·`card_record.go`·`card_picked.go`·`card_transition.go`, `internal/homestate/hub_files.go`(신규)와 `hub_files.txt`(신규, 임베드 데이터), `internal/cli/factory_card.go`, `internal/cli/todo_merge.go`(신규), `internal/cli/todo.go`(읽기 전용 동사 목록과 서브커맨드 등록).
M6: `internal/web/todo_queue_read.go`, `todo_view.go`, `screens.templ`(+생성물 `screens_templ.go`), `assets.go`, `assets/` 신규 자산, `assets/i18n.js`.
M5(게이트 충족 시): `.claude/rules/moai/workflow/card-issuance.md`(신규), `.claude/rules/local/card-issuance-thresholds.md`(신규, 로컬 전용), `gtd.md`, `factory-dispatch.md`, `factory-dispatch-detail.md`, `.claude/agents/moai/sync-auditor.md`, `manager-todo.md`와 각 `internal/template/templates/` 사본, `internal/template/workflow_rule_paths_pinned_test.go`, `internal/template/catalog.yaml`, 생성 `internal/template/templates/.codex/agents/moai/*.toml`(루트 `.codex/` 는 추적되지 않는다), 생성 플러그인 페이로드 `plugins/moai/skills/moai/workflows/gtd.md`, 그리고 압축이 카드 id 기준선 짝(`gtd.md t696`, `factory-dispatch.md t1330`, `factory-dispatch-detail.md t133`·`t224`)의 문장을 지울 때만 `internal/template/card_id_leak_test.go`(`cardIDBaseline` 항목 삭제; `t696` 짝이면 `TestCardIDBaselineIsPerFileAndPerLiteral` 의 상수도 다른 짝으로). M5(게이트 미충족 시): `.moai/specs/SPEC-TODO-CARD-ISSUANCE-001/m5-draft/`(신규, 이 디렉터리 밖은 건드리지 않는다).

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

### Out of Scope — 사용자 프로젝트의 허브 목록·임계값 설정

- 배포되는 사용자 프로젝트가 자기 허브 파일 목록이나 카드 크기·깊이·동시 진행 수치를 설정하는 키는 만들지 않는다. 출하되는 허브 목록은 이 저장소의 측정이다. 일치하는 경로가 없는 프로젝트에서는 체인이 만들어지지 않지만, 목록의 경로가 사용자 프로젝트에도 있는지(예: 흔한 Go 배치인 `internal/config/defaults.go`)는 측정하지 못했다 — 수용된 잔여이고 §G 에 적었다. 배포 규칙은 메커니즘만 싣고, 값은 각 프로젝트가 자기 측정으로 정하는 몫이다.

### Out of Scope — 배포·문서 표면

- 릴리스·CI 워크플로, README 4종, docs-site 는 이 SPEC 의 범위가 아니다. 사용자 문서 정정은 sync 단계의 후속 항목이다.

## §G 공백과 잔여 위험

측정하지 못했거나 판정이 열려 있는 것을 숨기지 않는다(전체 목록은 `research.md` §10).

- 어휘 이웃의 **정밀도**는 측정하지 못했다. 정답 집합이 없고, 유일한 대리 지표인 기록된 관계 125쌍은 Jev·에이전트의 판정이라 재현율만 잴 수 있었다.
- "이미 덮는 완료 SPEC" 조회는 약한 신호다. SPEC 1,029개 중(이 계획의 첫 측정 값) frontmatter 에 `card:` 가 있는 것은 32개뿐이고, 카드→SPEC 링크(`spec_id`)는 큐 1,174장 중 7장에만 있다. 이 항목은 `heuristic` 표지를 달고 나간다.
- 진행 중 레인의 **예상 파일 겹침**은 오늘 입력 자체가 거의 없다. `files` 속성은 어느 카드에도 아직 없고 오늘의 입력은 본문 경로 추출뿐이다 — 열린 카드 37장 중 본문에 경로가 보이는 것은 12장(32%)이며 정확히 같은 경로를 공유하는 쌍은 0이다(이 계획의 첫 측정 값). 그래서 이 항목은 M2 데이터가 쌓일 때까지 대부분 `unmeasured` 로 나가고, 양성 겹침 경로는 고정 입력 시험(AC-TCI-003 (f)·AC-TCI-002 (e))으로만 관측된다.
- git 탐침 비용: 레인 브랜치 하나의 `git diff --name-only` 가 약 0.25초다. `WT-` 브랜치와 워크트리의 수는 저장소 전역 ref 라서 트리의 성질이 아니라 **읽은 시각의 값**이고, 시각마다 움직인다 — 이 계획이 처음 잰 값(큐 스냅숏 10-03 11:45 무렵, 트리 `2de0a2cb6`) 353개·74개, 이터레이션 2 감사가 읽은 값(읽은 시각 미기록) 367개·89개, 이터레이션 3 이 2026-10-03 14:41 에 읽은 값 367개·89개, 이터레이션 5 가 2026-10-04 15:44 에 트리 `68a4d8137` 에서 읽은 값 397개·121개(`git branch --list 'WT-*' | wc -l` 와 `git worktree list | wc -l`). 이터레이션 2 저작자가 `1894984c3` 작업 트리에서 읽었다고 적었던 356개·79개는 읽은 시각이 기록되지 않았고 트리에 귀속할 수 없다. 이 값들은 현재 값이 아니다. 진행 중 레인이 16개면 직렬로 약 4초다 — 제시 경로의 시간 상한과 병렬화는 M1 이 먼저 재야 한다. 같은 방식으로 SPEC 디렉터리 수(1,029 → 1,034 → 1,031 → 1,049; `ls .moai/specs | wc -l` 는 작업 트리의 디렉터리를 세므로 다른 세션이 지우면 줄어든다)와 `card:` 를 가진 SPEC 수(32 → 33 → 36)도 움직였다. 이 값들의 측정 명령(`git branch --list 'WT-*' | wc -l`, `git worktree list | wc -l`, `ls .moai/specs | wc -l`, `git grep -l '^card:' -- '.moai/specs/*/spec.md' | wc -l`)은 M0 가 값 옆에 `command:` 줄과 읽은 시각으로 기록하고, 기준선은 그 값을 "측정 시각의 값"이라 부르며 현재 값이라 부르지 않는다.
- 카드→파일 간선을 만드는 `moai graph build` 의 실행 시간 증가는 측정하지 못했다.
- 카드→파일 간선의 **CI 일치는 미검증**이다. REQ-TCI-016 의 불변식은 (트리, 도달 가능한 이력) 위에서만 말한다. `graph-freshness` 의 `pull_request` 실행은 합성 병합 참조를 체크아웃하고 그 이력의 첫 부모 경로는 카드 브랜치와 다르다 — 로컬과 CI 가 같은 출력을 내는지는 측정하지 못했다(감사 D13).
- 웹 그래프의 레이아웃·번들 크기·브라우저 렌더 증거는 측정하지 못했다.
- 허브 파일 목록의 **완전성**(목록에 빠진 허브)은 시험이 다시 재지 못한다. 시험은 목록에 오른 경로가 기록된 단일 호출 측정 명령으로 기록 개수를 만족하는지(건전성)만 다시 잰다. 완전성은 M0 스크립트 실행 기록에만 기댄다.
- **출하 허브 목록이 사용자 프로젝트와 겹칠 가능성은 검증하지 못했다.** 목록의 경로가 사용자 프로젝트에 없다는 전제는 이 저장소가 볼 수 없는 프로젝트에 대한 주장이라 명령으로 확인할 수 없다 — 후보에는 흔한 Go 배치인 `internal/config/defaults.go` 가 들어 있다(`git ls-files -- internal/config/defaults.go` 가 이 저장소의 경로를 출력하고, `git ls-files | grep -c 'defaults.go$'` 는 이 저장소에서 4 를 출력한다). 그런 프로젝트에서 두 열린 카드가 `--files` 로 그 경로를 명시하면 이 저장소의 측정에서 나온 직렬화가 적용된다. 영향은 두 카드를 한 줄로 세우는 순서 지정이고 거절이 아니며, 입력이 명시 `files` 속성뿐이라(본문 추출은 체인을 만들지 않는다 — Module A 용어) 노출은 그 플래그를 쓴 사용자로 한정된다. 수용한 잔여이고 사용자별 허브 설정은 범위 밖이다.
- 허브 목록 측정 시험(`TestHubFileListMatchesMeasuringCommand`)은 기록된 SHA 를 클론이 갖지 않으면 건너뛴다. CI 의 test 작업은 `fetch-depth: 0` 이고(`.github/workflows/ci.yml:131`) 시험은 `CI` 환경 변수가 있는 실행에서 건너뛰지 않고 실패한다(AC-TCI-019 (f)); 로컬의 얕은 클론에서의 건너뜀은 통과가 아니라 공백이다.
- M5 게이트의 고정 SHA 대조(`b05c3be90…` 위 행들)는 미푸시 브랜치 `WT-github-flow-default` 의 객체가 있는 클론에서만 재현된다. 이식 가능한 런타임 대조는 부등식 한 쌍이다(`acceptance.md` AC-TCI-020 판독 4).
- t1453 이 같은 규칙 파일을 절체 시점에 고친다는 판단은 그 브랜치 끝(`53e3f7ebb`, 이동하는 ref — 이터레이션 4 까지는 `2a5f9c91c`)의 plan 문서를 읽은 것이다. 병합 기준(`46be0b8c8`) 이후 그 브랜치 자신이 여섯 규칙 파일(개명 전 이름 포함)·`todo.go`·`internal/kanban` 을 바꾼 차이는 비어 있었다(`git diff --stat --no-renames`). t1453 이 `merge(t1453)` 형태의 제목으로 착지할지는 확인하지 못했다 — 다른 모양(빨리감기, 스쿼시, 카드 id 없는 제목)이면 게이트가 닫힌 채로 읽는다(안전한 방향). 착지가 되돌려진(revert) 경우는 원래 병합이 조상으로 남아 게이트가 열린 채로 읽힌다 — 이 이력에는 되돌림 사례가 없어 측정하지 못했고(추정), 판독 2 가 SHA 를 `progress.md` 에 적으므로 리더가 읽을 수 있다. 게이트의 착륙 정의는 **어휘 규칙**이다(REQ-TCI-021): 병합 제목 줄에 `absorb` 가 든 병합은 흡수 방향으로 읽혀 판독 1 의 둘째 명령이 제외한다(`acceptance.md` AC-TCI-020). 두 방향의 거짓을 고정한 통합 팁의 첫 부모 사슬로 병합을 갈라(첫 부모가 사슬 밖·둘째 부모가 사슬 위 = 흡수 방향, 반대 = 착지) SHA 만으로 쟀다. **거짓 열림**(흡수 방향인데 제목에 낱말이 없다): 핀 `68a4d8137`(사슬 끝 `30ce3a02d`)에서 흡수 방향 74개 가운데 2개(`c2703f698`, `35e2d7157`; 2.7%), 이터레이션 4 핀 `ad02a5677`(사슬 끝 `42d8474de`)에서 68개 가운데 같은 2개(2.9%) — 감사가 움직이는 ref(`develop` 과 `origin/develop`)로 읽은 "75개 가운데 2개"는 SHA 로 재현되지 않는다. **거짓 닫힘**(착지인데 제목에 낱말이 든다): 현재 핀에서 착지 260개 가운데 7개(2.7%; 제목 모양만 읽은 분류), 이터레이션 4 핀에서 217개 가운데 6개이고 그 가운데 제목이 `into develop` 착지로 읽히는 것이 적어도 셋이다. 그래서 `T ≥ 1` 인데 `T − A = 0` 이면 판독 2 의 목록을 읽고, 제목이 착지로 읽히는 줄이 있으면 **카드 레인은 게이트를 스스로 열지 않는다** — SHA 를 `progress.md` §E.2 에 적어 리더에게 보고하고 리더의 판정(그 SHA 에 판독 2 의 두 점검과 제목을 읽은 결과)이 후보를 대신한다; 판정이 없으면 Mode B 다. 핀 `68a4d8137` 의 카드 id 병합 제목 382개 가운데 제목 줄에 `absorb` 가 든 것은 83개이고 본문에만 든 착지는 31개다(제외되지 않는다).
- 템플릿 사본의 **측정 수치 검사 범위**(D31): `TestCardIssuanceTemplateIsMechanismOnly` 의 수치 검사는 새 `card-issuance.md` 사본 하나에만 건다. M5 가 편집하는 다섯 사본(`gtd.md`, `factory-dispatch*.md`, `sync-auditor.md`, `manager-todo.md`)은 기존 가드(카드 id 누출, 중립성 정규식의 SPEC·REQ id·날짜·16진 단어)가 닿는 범위까지만 기계가 보고, 숫자 임계값이 거기 들어가는지는 M5 커밋의 `git diff` 를 리뷰가 읽는다 — 수용한 잔여다(`acceptance.md` MU-121). REQ-TCI-021 의 "any template copy" 문구는 저자에게 거는 의무이고 기계 검사의 범위가 아니다.
- 흡수(이터레이션 5) 뒤 다시 재지 않은 것: §A.1 표의 git·큐 쪽 값과 `research.md` §3 의 figure 35개(트리 `2de0a2cb6`·큐 스냅숏 시점의 값 — M0 가 실행 시점에 다시 잰다), 허브 후보 중 개명된 파일의 이력(`factory-dispatch.md` 의 경로 한정 `git log` 는 개명 전 이력을 따라가지 않는다 — M0 는 두 이름의 이력을 합쳐 세거나 `--follow` 를 쓴다; 옛 값 `kanban-dispatch.md` 35(10)). 이터레이션 5 가 다시 읽은 것은 코드·규칙 인용, 크기, 미러, 가드, 게이트와 그 통제, 증거 원장이다.
- 잔여 위험: M5 가 게이트에 막혀 SPEC 이 요구는 충족하되 규칙은 아직 효력이 없는 상태로 닫힐 수 있다. 그때 리더가 M5 적용 카드를 발행해야 하며, 완료 보고가 이를 명시해야 한다.

## §H 상호참조

- `research.md` — 증거와 재측정 표, 스크립트, 공백 목록.
- `design.md` — 아키텍처, 데이터 흐름, 스키마, 관계 모델, 간선 운반체, 묶음 훅, D1~D15 선택지 표, 기준선 운반체(§12), M5 게이트 술어(§13).
- `plan.md` — 마일스톤 분해, 게이트 명령, 위험, @MX 계획, 마일스톤별 검증 명령, 작업 기본값 요약.
- `acceptance.md` — 수용 기준, RED-now 원장, 변이 탐침.
- `decision-index.md` — 결정 열다섯 행.
- 인접 SPEC: SPEC-TODO-ANALYSIS-001, SPEC-TODO-AUTO-PICK-001, SPEC-RELATION-PICKUP-FILTER-001, SPEC-GTD-AUTONOMY-001, SPEC-WEB-TODO-QUEUE-001, SPEC-INSTRUCTION-BUDGET-SCOPE-001.
