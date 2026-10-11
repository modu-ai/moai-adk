---
id: SPEC-ALWAYS-LOADED-BUDGET-001
title: "plan — 배포 표면 상시 로드 지시문 예산"
version: "0.7.0"
created: 2026-10-03
updated: 2026-10-04
---

# plan.md — SPEC-ALWAYS-LOADED-BUDGET-001

## §A. 분류

- **Tier L** — 영향 파일이 15개를 넘는다(템플릿 규칙 13 + 로컬 미러 13 + companion + 훅 Go + 테스트 + 경로 고정 테스트 갱신). 상시 지시문 체계, 곧 헌법급 표면을 바꾼다.
- **클래스 C** — 여러 하위 시스템(템플릿 규칙, SessionStart 훅, 템플릿 테스트)에 걸친 설계 변경이다.
- 개발 방식: 예산 가드와 역할 가드는 TDD(RED 먼저), 규칙 분할은 DDD(구속 원장으로 행동 보존을 고정한 뒤 이동).

## §B. 카드 t1450 과의 경계 — 흡수와 잔류

| t1450 항목 | 판정 | 이유 |
|---|---|---|
| (a) 역할 한정 규칙의 SessionStart 주입 — `factory-dispatch`(구 `kanban-dispatch`)·`cross-session-messaging` | **흡수** | 이 카드의 핵심 설계와 같다. 두 카드가 같은 파일을 따로 고치면 충돌한다 |
| (a) 같은 묶음의 `goal-directive`·`moai-mcp-tools` | **흡수(주입이 아니라 분할로)** | 둘 다 역할 한정이 아니다. `/moai goal` 은 일반 세션이 쓰고, `project_root` 의무는 워크트리의 모든 에이전트에 걸린다. 역할 경로로 보내면 일반 세션이 의무를 잃으므로 M3 의 core+companion 분할 대상으로 다룬다 |
| (a) 누락 시 리더가 규칙 없이 뜨는 위험의 가드 테스트 | **흡수** | REQ-ALB-009·011 |
| (b) 출력 스타일 본문 축약 | **잔류** | 출력 스타일은 이 SPEC 이 재는 지시문 파일 집합 밖이다. 다른 계측(첫 턴 토큰)과 다른 소유 파일을 갖는다 |
| (c) 스킬 목록 예산·에이전트 설명 상한 | **잔류** | 스킬 목록·Agent 도구 설명은 별개 런타임 메커니즘이다 |
| (d) `AGENTS.md` 와 규칙 중복 — 규칙 쪽 | **흡수** | 규칙 문단을 `AGENTS.md` 포인터로 바꾸는 것은 규칙 분할 작업의 일부다(`design.md` §4) |
| (d) `AGENTS.md` 본문 쪽, Codex 자족성 | **잔류** | Codex 지시문 예산과 자족성 계약이 걸린 별개 설계다 |
| 「각 항목 전후 첫 턴 토큰 실측」 | **흡수(이 SPEC 항목 한정)** | REQ-ALB-021 |

t1450 이 남은 항목만 다루도록 카드 본문을 조정하는 것은 리드의 큐 작업이다(이 SPEC 은 큐를 바꾸지 않는다).

## §C. 착지 순서 제약

- **이 SPEC 의 분할이 t1453 M4 보다 먼저 착지한다.** t1453 M4 는 `factory-dispatch*`(구 `kanban-dispatch*`)·`AGENTS.md`·`CLAUDE.md`·규칙의 develop 서술을 고친다. 순서가 뒤집히면 t1453 이 고친 문장이 이 SPEC 의 분할에서 축자 이동 대상이 되어 원장 기준선이 움직이고, 두 카드가 같은 문단을 두 번 편집한다. t1453 은 이 SPEC 착지 뒤 develop 을 흡수하고 새 배치(상시 core / 역할 core / companion) 위에서 문구를 고친다.
- **t1399(런처 진입 `-f`/`-l`, 칸반 모드 제거)가 먼저 착지하면** M2 착수 전에 develop 을 흡수하고, 그 시점의 역할 표지 집합으로 REQ-ALB-007·011 을 구현한다. t1399 가 늦으면 현재 표지 집합으로 구현하고 t1399 가 표지를 바꿀 때 역할 가드(REQ-ALB-011)가 그 변경을 잡는다(`decision-index.md` Q8).
- 규칙 편집은 모든 세션의 프롬프트 캐시를 무효화한다. 병합은 배치 경계에서 한다.
- 흡수할 때마다 §C.2 재고정 절차를 따른다.

## §C.1 착지 시점

- v3.2.0 릴리스에 넣지 않는다. v3.2.0 컷 뒤 develop 에 착지한다(운영자 판정 Q7). t1453 M4 보다 먼저라는 순서 제약은 그대로다.

## §C.2 원장 앵커와 재고정 절차

원장 기준선은 M0 착수 트리에서 뽑고, 그 커밋 SHA 를 원장 머리의 `anchor` 로 적는다. 이후 develop 을 흡수할 때마다(t1399 병합, v3.2.0 컷 이후 흡수 포함) 다음을 순서대로 한다.

1. 흡수 직후 `git diff --stat <이전 앵커> HEAD -- internal/template/templates/.claude/rules internal/template/templates/CLAUDE.md internal/template/templates/AGENTS.md.tmpl` 을 잰다. 출력이 비면 앵커를 그대로 둔다.
2. 출력이 있으면, 흡수 커밋의 부모 가운데 develop 쪽 커밋을 새 앵커로 삼아 그 트리에서 구속 블록 기준선을 다시 뽑는다. 이 SPEC 의 편집이 섞이지 않은 상류 트리이기 때문이다.
3. 원장에 증감 부록을 단다: 상류에서 새로 생긴 블록(행 추가), 상류에서 사라진 블록(`upstream-removed` 와 그 상류 커밋 SHA), 상류에서 바뀐 블록(변경 전 텍스트 갱신 후 재매핑).
4. 새 앵커 트리에서 예산 테스트 기대값(렌더 독립 소계, 구성원 목록)을 다시 잰다. 새 기준선·부록·기대값을 이 SPEC 의 다른 편집보다 먼저 별도 커밋으로 착지시키고, `anchor` 를 새 SHA 로 바꾼다. AC-ALB-002·003·004·007 은 이 기록을 기대값으로 읽는다.
5. 원장 테스트(REQ-ALB-015)를 다시 돌려 초록을 확인한 뒤에야 규칙 편집을 이어간다.

## §D. 마일스톤 (결정이 바뀔 가능성이 큰 순)

### M0 — 도달 가능성 측정과 중단 지점 (Priority High)

- 첫 행동: 로컬 develop 을 흡수하고 §C.2 로 첫 앵커를 세운다. 이 SPEC 의 모든 기대값은 이 앵커에서 시작한다. plan 단계의 `b5815ca80` 수치(169,018 등)는 기준이 아니라 이력이다.

- 앵커 트리에서 구속 원장 기준선을 만든다: 원장 대상 파일(spec §B — 상시 규칙 13개와 `CLAUDE.md` 템플릿)의 모든 절의 단위(§B 「단위 경계」) 전부를 뽑아 `binding`/`normative`/`rationale` 로 나누고, `binding`·`normative` 행에 `entry_points` 를 채워 `internal/template/testdata/` 아래 고정물로 커밋하고, 단위별 앵커 종류와 앵커 SHA 를 머리에 적는다(REQ-ALB-015, §C.2).
- 블록마다 역할 core / 일반을 판정한다. 판정 규칙은 REQ-ALB-025 — 무표지 리더급 진입점까지 덮이지 않는 블록은 일반이다(`design.md` §2).
- 실측 두 건: 런타임이 SessionStart `additionalContext` 를 잘리지 않고 전달하는 크기(Q4), 그리고 t1399 흡수 뒤의 역할 표지 집합(Q8).
- 후보표: 파일·절마다 비구속 자수, 옮길 companion, 그 companion 의 남은 수용량.
- 원장 행마다 처리 방식(축자 이동 / 의미 보존 압축 재작성)과 전·후 텍스트를 적는다. 재작성 행은 감사자가 의미 보존을 판단할 수 있도록 무엇을 줄였는지 한 줄로 남긴다.
- 하한 산출: 이 SPEC 의 규칙(on-demand 이동 금지, 역할 주입 허용, 의미 보존 압축 재작성 허용, 의무 무손실) 아래 도달 가능한 최저 합계.
- 예산 115,000 은 운영자가 유지했다(Q2·Q6). 압축 재작성까지 적용한 하한이 그래도 115,000 을 넘을 때만 규칙을 고치기 전에 멈추고 리드에게 보고한다(REQ-ALB-022).
- 기준선 원장은 구현보다 먼저 별도 커밋으로 착지한다(커밋 그래프만이 순서를 증언한다).

### M1 — 예산 가드 RED (Priority High)

- `internal/template` 에 `TestDeployedAlwaysLoadedCharBudget` 작성. 배포·렌더링 경로는 패키지의 운영 배포기와 렌더러를 쓴다. `internal/cli` 의 init 경로를 부르면 import 순환이 생기므로 쓰지 않는다.
- M1 RED 커밋은 M0 의 첫 흡수·앵커 커밋 뒤에 착지한다. RED 확인: 고정 렌더 입력으로 실패하며, 원장 머리에 기록된 렌더 독립 소계·구성원 목록과 파일당 초과 파일을 보고해야 한다. 이 RED 커밋은 템플릿을 바꾸지 않는다(`git diff --stat <앵커> <M1 커밋> -- internal/template/templates` 무출력).
- 원장 단위 변이 검사: 구속 블록의 연속 줄 하나를 지운 픽스처에서 원장 테스트가 실패해야 한다(AC-ALB-020).
- 표면 도출의 변이 검사(REQ-ALB-004)를 픽스처 하위 테스트로 둔다.

### M2 — 역할 주입과 가드 (Priority High)

- 역할 한정 규칙 본문에 중립 영역 표지를 넣고, 배포 파일만 읽는 역할 core 생성 함수를 만든다(REQ-ALB-023). 원장은 런타임에 읽지 않는다.
- 런처와 훅이 함께 읽는 역할 표지 레지스트리를 둔다. 이미 공용 상수 집합이 있으면 그것을 레지스트리로 승격하고, 없으면 `internal/config` 에 만든다(REQ-ALB-011).
- SessionStart 주입(`startup`·`clear`·`compact`).
- 무표지 리더급 진입점: `manager-lead` 에이전트 정의, `moai-factory-foreman` 스킬(구 `moai-kanban-foreman`, t1399 개명), `gtd` 워크플로의 `--auto` 경로에 `moai:role-rules-required` 표지와 읽기 지시를 넣고, 배포 트리에서 역할 한정 규칙을 인용하는 파일을 모두 찾아 표지를 요구하는 가드를 둔다(REQ-ALB-024). 에이전트 정의를 고치면 `make agents-emit` 으로 Codex 방출본을 재생성한다.
- 실패 처리(REQ-ALB-009)와 크기 상한 처리(REQ-ALB-010) — 상한값은 Q4 실측 뒤 정한다.
- 역할 가드(REQ-ALB-011): 레지스트리의 리더·레인 표지마다 주입 맥락에 역할 core 블록 전부가 있음을 단정. 기대 블록은 배포 파일의 영역 표지에서 도출.
- 비역할 세션 무주입 단정(REQ-ALB-008).

### M3 — 규칙 분할 (Priority Medium)

- 템플릿 원본에서: 역할 한정 2개의 stub 생성과 전체 본문의 상시 표면 이탈, 상시 core 11개와 `CLAUDE.md` 의 비구속 본문 이관. `factory-dispatch-detail.md`(구 `kanban-dispatch-detail.md`)의 40,000 미만 분할은 Q10(a) 범위 축소로 이 SPEC 밖이다 — 후속 카드로 옮겨졌고(`decision-index.md` Q11, 제거 기록 `progress.md` §E.1), `workflow/spec-workflow.md` 40,052 초과는 리드가 별도로 처리한다. 기존 `paths:` companion 은 원장 대상이 아니며(spec §B), M3 가 companion 에 추가하는 것은 `rationale` 뿐이다(REQ-ALB-017).
- 원장 테스트(REQ-ALB-015)와 `§` 교차 참조 해석 테스트(REQ-ALB-016)를 초록으로.
- 로딩 범위 문구 정정(REQ-ALB-020).

### M4 — 경로 고정 테스트와 미러 (Priority Medium)

- 이동된 문장을 특정 파일에서 찾던 기존 문서 고정 테스트를 새 위치(상시 core ∪ 역할 core ∪ companion)로 갱신. 단정의 의미는 바꾸지 않는다.
- `make build` 후 이 저장소의 `.claude/rules/moai/` 미러 동기화. 바뀌거나 새로 생긴 규칙 쌍(역할 한정 규칙 2개, stub 2개, 분할 companion)을 미러 대칭 테스트에 등록하고 초록 확인(REQ-ALB-018).

### M5 — 측정 기록과 정리 (Priority Low)

- 전후 합계와 첫 턴 맥락 실측을 `progress.md` 에 기록(REQ-ALB-021).
- 템플릿 중립성 가드·내부 내용 누출 테스트 초록 확인.

## §E. 위험

| 위험 | 영향 | 대응 |
|---|---|---|
| 압축 재작성까지 적용해도 115,000 에 닿지 않음(`research.md` §2) | 카드가 M0 에서 멈춤 | M0 를 첫 마일스톤이자 중단 지점으로 둠. 하한을 숫자로 보고 |
| 압축 재작성이 의무의 범위를 좁힘 | 구속 의무가 조용히 약해짐 | 원장 전/후 행을 plan·sync 감사가 한 줄씩 확인, 판단이 갈리는 행은 축자로 되돌림 |
| 역할 표지가 t1399 로 바뀜 | 역할 세션이 규칙 없이 시작 | 가드가 레지스트리에서 표지를 도출하므로 새 표지도 검사 대상, M2 직전 develop 흡수 |
| 무표지 진입점(deputy·foreman·`--auto`)이 규칙 없이 동작 | 리더급 의무 소실 | REQ-ALB-024 표지·읽기 지시와 인용 파일 전수 가드, REQ-ALB-025 분류 규칙 |
| 주입 맥락 크기 한도 미측정 | 역할 core 가 잘려 의무 소실 | Q4 실측 뒤 상한 결정, 상한 초과 시 구속 블록 + Read 지시, 그래도 넘으면 경고 + Read 지시(블록을 자르지 않음) |
| `compact` 뒤 역할 규칙 소실 | 긴 리더 세션이 규칙 없이 진행 | `compact` 에서도 재주입 |
| 문서 고정 테스트 대량 실패 | M4 가 길어짐 | 단정 의미 불변, 위치만 갱신. 테스트 목록은 M0 후보표에 함께 적음 |
| 런타임 주입 크기 한도(Q4)·역할 표지(Q8) 미측정 | 주입 설계가 바뀔 수 있음 | M0 에서 실측한 뒤 M2 착수 |

## §F. 금지 사항

- 구속 의무를 지우지 않는다. 재작성·병합은 원장 행이 있고 의미를 보존할 때만 한다.
- 행동으로 걸리는 규칙에 `paths:` 만 붙여 상시 표면에서 빼지 않는다.
- `sessionCharBudget`·`AlwaysLoadedTokenBudget` 를 바꾸지 않는다.
- 템플릿에 SPEC ID·카드 id·날짜·커밋 해시를 넣지 않는다.

## §G. 교차 참조

- `spec.md` §C 요구사항, `acceptance.md` 인수 조건, `design.md` 배치, `research.md` 실측, `decision-index.md` 미결 결정.
