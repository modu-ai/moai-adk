# 결정 색인 — SPEC-REVIEW-GATE-REPAIR-001

> 이 색인은 상태 축을 갖지 않는다. 이 SPEC 의 생명주기는 spec.md 만이 가진다. 각 행의 `Operator verdict` 는 작성 시점에 비워 둔다. 이 행들은 Implementation Kickoff 를 차단한다.

## 개요

| 번호 | 질문 | 라벨 | 분류 | 기본값 |
|---|---|---|---|---|
| Q1 | 두 설정 키가 다른 ref 를 가리킬 때 기준 판정 | FOUNDER | product-level | 없음 |
| Q2 | 연속 차단 상한 C 의 값 | FOUNDER | product-level | C=1 (규칙 적용) |
| Q3 | 실행 단위 스위치의 형태 | FOUNDER | product-level | 없음 |
| Q4 | inconclusive 뒤 재검토 간격 | FOUNDER | product-level | 없음 |
| Q5 | 두 계층 동시 활성 시 doctor 경고 | FOUNDER | product-level | 추가하지 않음 (규칙 적용) |
| Q6 | SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 기록 처분 | FOUNDER | product-level | (b) 이 SPEC 에만 기록 (규칙 적용) |

**권한 등록부 검토 결과.** DECIDED 와 POLICY-COVERED 로 라우팅할 수 있는 행이 없다.

- `AGENTS.md` §2 는 "카드 브랜치는 로컬 main 에서 자르고 카드 PR 은 main 으로 간다"고 적지만, 이 파일은 권한 등록부(제품 문서, 완료 SPEC 의 HISTORY 와 Amendments, `.moai/config/sections/*.yaml`, 헌법)에 속하지 않는다. 근거로 인용하지 않는다.
- SPEC-CODEX-REVIEW-TARGET-001 §A.7 은 완료된 SPEC 이지만, 그 결정을 뒤집은 카드 t1426 의 항목이 HISTORY 에 없다(`grep -c t1426` = 0). 같은 질문으로 인정되지 않는다.
- 따라서 Q1–Q6 은 모두 FOUNDER 다. 권한 앵커를 검증할 수 없었던 항목은 "escalate, never downgrade" 원칙에 따라 FOUNDER 로 남긴다.

---

### Q1: 두 설정 키가 다른 ref 를 가리킬 때, 리뷰 기준 판정은 무엇인가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER. 검증 가능한 권한 앵커가 없다.)
Why unresolved: 카드 M2 의 문언("두 키가 다르면 inconclusive")을 이 저장소의 설정에 적용하면 모든 판정이 inconclusive 가 된다. 관측된 값은 다음과 같다. `worktree_base_branch` 는 `main`(`.moai/config/sections/git-strategy.yaml:5`), `develop_branch` 는 `develop`(같은 파일 :16)이며, 턴 종료 기준 `IntegrationTarget` 은 git-flow 수동 프로필에서 `develop` 이다(`internal/config/loader_integration_branch.go:123`). 반대로 최근접 선택만 두면 키 불일치가 판정에 드러나지 않는다. 또한 카드 t1426(커밋 `f82a078bb`)은 감사 경로의 기준을 `worktree_base_branch` 우선으로 바꿨고, 이는 완료된 SPEC §A.7 과 충돌하지만 HISTORY 에 기록되지 않았다.
Options:
- (a) 최근접 후보를 두 키의 후보 합집합에서 고른다. 키 불일치는 판정을 막지 않는다.
- (b) 문언대로: 키가 다른 ref 를 가리키면 inconclusive 를 두 이름과 함께 낸다. 이 저장소에서는 설정이 정렬되기 전까지 검토가 열리지 않는다.
- (c) 설정 정렬을 먼저 한다(git-flow 잔재 정리, SPEC-GITHUB-FLOW-DEFAULT-001 M5 범위). 이 SPEC 은 정렬 뒤에 진행한다.
Default: 없음. 규칙의 세 순위를 이 선택지에 적용해도 분리되지 않는다. (a) 와 (b) 모두 현행 동작을 바꾼다. 단일 되돌리기는 모두 가능하다.
Alternate: 없음.
Operator verdict:

---

### Q2: 턴 종료 연속 차단 상한 C 의 값은?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 "연속 차단을 상한으로 막는다"고만 적었고 값을 주지 않았다. 현행 동작을 확인했다. 한 턴 안에서 차단된 뒤의 재종료는 stop-hook-active 플래그로 허용된다(`internal/cli/codex_review_gate.go:109`). 턴을 넘는 연속 차단에는 현재 상한이 없다. 같은 키의 fail 이 남아 있으면 새 턴마다 다시 차단된다. 다른 축의 상한도 있다. 측정되지 않은 결과의 연속 횟수 상한 `StopUnmeasuredCap = 3`(`internal/codexwiring/stop_budget.go:153`)은 차단 횟수가 아니므로 직접 대응하지 않는다.
Options:
- (a) C = 1
- (b) C = 2
- (c) C = 3
- (d) 다른 값(운영자 지정)
Default: C = 1. 규칙 적용 결과다. 첫 순위(현행 동작 보존)에 맞는 선택지가 없다. 현행은 턴 사이 상한이 없고, 카드 M1 은 상한을 요구한다. 둘째 순위(단일 되돌리기)는 모든 선택지가 만족한다. 셋째 순위(사용자에게 보이는 표면이 작은 것)에서는 차단 횟수가 가장 적은 C = 1 이 선택된다. 이 기본값은 추천이 아니며, 운영자 판정 없이 적용되지 않는다(product-level 행).
Alternate: C = 2 또는 C = 3.
Operator verdict:

---

### Q3: 실행 단위 스위치의 형태는?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 "환경 값 또는 주 체크아웃의 미추적 재정의 중 하나"를 허용했다. 현재 스위치는 추적 파일의 per-tree 키다(`.moai/config/sections/workflow.yaml` 302행, HEAD 값 `true`). 셸 층은 `CLAUDE_PROJECT_DIR` 의 workflow.yaml 을 읽고(`.claude/hooks/moai/handle-codex-review-entry.sh:33-34`), Go 층은 `auditreceipt.StoreRoot` 가 가리키는 트리의 workflow.yaml 을 읽는다(`internal/cli/codex_review_gate.go:469`). 두 방식의 전파 조건이 다르다. 환경 값은 레인 터미널마다 설정해야 할 수 있다. 미추적 재정의는 주 체크아웃 한 곳에 두지만, 사용자에게 보이는 파일이 하나 생긴다.
Options:
- (a) 환경 값(이름과 읽기 순서는 run 단계가 정한다)
- (b) 주 체크아웃의 미추적 재정의 파일
- (c) 둘을 우선순위로 함께 둔다: 환경 값이 있으면 그것을, 없으면 재정의 파일을 쓴다
Default: 없음. 규칙의 셋째 순위(사용자에게 보이는 표면이 작은 것)가 두 선택지를 깔끔하게 나누지 않는다. 환경 값은 파일을 늘리지 않지만 레인마다 설정이 필요하고, 재정의 파일은 한 곳에 있지만 파일을 하나 추가한다.
Alternate: 없음.
Operator verdict:

---

### Q4: inconclusive 영수증 뒤의 재검토 간격은?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드 M3 은 inconclusive 를 재사용하지 않는다고만 한다. 그러면 턴이 끝날 때마다 검토가 다시 시작되고, 진행 중 표식은 중복 시작만 막는다. 검토 한 번에 걸리는 시간은 연구 입력(lens-1)에서 3–9분으로 기록되어 있으나 이 세션에서 재측정하지 않았다. 레인 저장소 집계(기계 로컬, 재관측 불가)는 `codex-review` 항목 117건 가운데 inconclusive 48, fail 48, pass 21 이다. 재검토 비용의 실측은 없다.
Options:
- (a) 간격 없음: 다음 턴 종료에서 다시 시작하고, 진행 중 표식으로만 중복을 막는다.
- (b) 고정 쿨다운 뒤 재시작. 후보 값은 검증 신선도 기본값 10분(`internal/verify/freshness.go:7`)이다.
Default: 없음. 규칙의 첫 순위(현행 동작 보존)를 두 선택지 모두 바꾼다. 둘째와 셋째 순위도 두 선택지를 분리하지 않는다.
Alternate: 없음.
Operator verdict:

---

### Q5: 두 리뷰 계층이 함께 켜졌을 때의 doctor 경고를 이 SPEC 에 넣는가?

Label: FOUNDER
Class: product-level (doctor 의 기본 출력이 바뀐다)
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 두 계층을 함께 켜면 doctor 경고가 뜬다고 적었다. 관측 결과 `internal/cli/doctor*.go` 에는 두 계층을 함께 점검하는 검사가 없다(research.md §5). 같은 주장은 규칙 문서 한 줄(`.claude/rules/moai/workflow/factory-dispatch-detail.md:157`)에만 있다. 검사를 추가하면 doctor 출력이 바뀌고, 추가하지 않으면 카드의 전제 하나가 거짓으로 남는다.
Options:
- (a) 이 SPEC 에 doctor 검사를 추가한다.
- (b) 추가하지 않고, 후속 항목으로 등록하며 규칙 문장을 정정한다.
Default: (b) 추가하지 않음. 규칙 적용 결과다. 첫 순위(현행 동작 보존)에 맞는 선택지가 (b)다. 현재 doctor 출력에는 해당 경고가 없다.
Alternate: (a).
Operator verdict:

---

### Q6: SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 기록 처분은?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER. 대상 SPEC 이 in-progress 이므로 HISTORY 개정 행이 없다.)
Why unresolved: M1 은 진입 시 차단을 없애므로, 진행 중인 SPEC-GATE-BOTTLENECK-001 의 REQ-GBN-002("a fresh FAIL at turn entry SHALL block")와 충돌한다. 같은 설계는 진입 훅의 헤더 주석(`internal/cli/codex_review_delay.go` 1–14행)과 AC-GBN-004 에도 있다. 이 SPEC 이 그 요구를 어떻게 기록할지 정해야 한다.
Options:
- (a) 대상 SPEC 에 개정 행을 추가한다(진행 중이므로 같은 파일을 편집한다).
- (b) 이 SPEC 의 HISTORY 와 Follow-up 에만 기록하고, 대상 SPEC 은 건드리지 않는다.
- (c) 이 SPEC 이 대상 SPEC 의 해당 요구를 대체한다고 선언한다. 필요한 구조 변경은 운영자 판정을 거친다.
Default: (b). 규칙 적용 결과다. 첫 순위(현행 기록 보존)에 맞는 선택지가 (b)이며, 다른 SPEC 파일을 건드리지 않는다.
Alternate: (a) 또는 (c).
Operator verdict:
