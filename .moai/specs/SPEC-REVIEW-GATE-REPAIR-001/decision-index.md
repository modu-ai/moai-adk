# 결정 색인 — SPEC-REVIEW-GATE-REPAIR-001

> 이 색인은 상태 축을 갖지 않는다. 이 SPEC 의 생명주기는 spec.md 만이 가진다. 각 행의 `Operator verdict` 는 운영자 판정 원문과 보드 기록 번호로 채웠다(spec v0.1.1, 2026-10-10). 판정이 빈 행은 없다. 2회차(spec v0.1.2)에서 Q1 에 개정 판정 행을 더했고, 2회차 기록(F-2, 문서 불일치)은 Q 행 밖에 두었다.

## 개요

| 번호 | 질문 | 라벨 | 분류 | 기본값 (규칙 적용 결과) | 판정 |
|---|---|---|---|---|---|
| Q1 | 두 설정 키가 다른 ref 를 가리킬 때 기준 판정 | FOUNDER | product-level | (a) (rule: 순위 3) | (a) 합집합에서 최근접 후보 → 개정 (가) 합집합 유지 + 활성 흐름 가드 (d-20261010T063222Z-43e0) |
| Q2 | 연속 차단 상한 C 의 값 | FOUNDER | product-level | C = 1 (rule: 순위 3) | C = 1 |
| Q3 | 실행 단위 스위치의 형태 | FOUNDER | product-level | 없음 (rule: 순위 3, 분리되지 않음) | (b) 주 체크아웃의 미추적 재정의 파일 |
| Q4 | inconclusive 뒤 재검토 간격 | FOUNDER | product-level | 없음 (rule: 분리되지 않음) | (a) 간격 없음 |
| Q5 | 두 계층 동시 활성 시 doctor 경고 | FOUNDER | product-level | (b) 추가하지 않음 (rule: 순위 1) | (b) |
| Q6 | SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 기록 처분 | FOUNDER | product-level | (b) 이 SPEC 에만 기록 (rule: 순위 1) | (b) |

보드 판정 D3·D4·D11 은 Q 행이 아니므로 아래 "보드 판정 기록" 절에 같은 방식으로 기록했다.

**권한 등록부 검토 결과.** DECIDED 와 POLICY-COVERED 로 라우팅할 수 있는 행이 없다.

- 권한 등록부의 정의는 `.claude/agents/moai/manager-spec.md:115` 에 있다. 등록부는 제품 문서(`.moai/project/product.md`), 완료 SPEC 의 HISTORY 와 Amendments, `.moai/config/sections/*.yaml` 의 운영자 설정, 헌법이다. 결정 색인 작성 규칙은 `.claude/skills/moai/workflows/plan/clarity-interview.md:63` 과 `:177` 에 있다.
- `AGENTS.md` 는 이 등록부에 속하지 않는다. 근거로 인용하지 않으며, Q1 의 충돌 출처로만 인용한다. `.claude/rules/local/gitflow-lane-protocol.md` 도 등록부 밖이며 충돌 출처로만 인용한다.
- SPEC-CODEX-REVIEW-TARGET-001 §A.7 은 완료된 SPEC 이지만, 그 결정을 뒤집은 카드 t1426 의 항목이 HISTORY 에 없다(`grep -c t1426` = 0). 같은 질문으로 인정되지 않는다.
- 따라서 Q1–Q6 은 모두 FOUNDER 라벨을 유지한다. 운영자 판정은 아래 각 행의 `Operator verdict` 와 보드 기록에 있다. 보드 기록을 권한 앵커로 인용하려면 digest 접두사와 인용 줄을 고정(pin)해야 하는데, 이 SPEC 은 그 고정을 하지 않았다. 라벨을 DECIDED 로 올리지 않는다.

---

### Q1: 두 설정 키가 다른 ref 를 가리킬 때, 리뷰 기준 판정은 무엇인가?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER. 검증 가능한 권한 앵커가 없다.)
Why unresolved: 카드 M2 의 문언("두 키가 다르면 inconclusive")을 이 저장소의 설정에 적용하면 모든 판정이 inconclusive 가 된다. 관측된 값: `worktree_base_branch` 는 `main`(`.moai/config/sections/git-strategy.yaml:5`), `develop_branch` 는 `develop`(같은 파일 :16). 턴 종료 기준 `IntegrationTarget` 은 git-flow 수동 프로필에서 `develop` 이다(`internal/config/loader_integration_branch.go:123`). 감사 경로는 `worktree_base_branch` 를 첫 단계로 읽는다(`internal/cli/mcp_review_material.go:159`, 카드 t1426 커밋 `f82a078bb`).
Resolved by: d-20261010T063222Z-43e0 (운영자 위임 판정 (가), 아래 Operator verdict (amendment) 행). 이 Why unresolved 문단은 판정 이전의 원문이다.

충돌하는 커밋 문서 출처가 둘 있다. (1) `AGENTS.md:94–95` 와 `:166–167`: 카드 브랜치는 로컬 `main` 에서 자르고, 카드 PR 은 `main` 으로 보내며, `develop` 은 2026-10-05 GitHub Flow 전환 뒤 legacy 라고 적는다. (2) `.claude/rules/local/gitflow-lane-protocol.md:16`: "카드 워크트리는 `develop`에서 만든다. `main`이 아니다." 이 파일은 `.claude/rules/local/` 아래 로컬 전용 문서이며 git 추적 파일이다. 권한 등록부 밖이므로 충돌 출처로만 인용한다. 두 출처는 카드 워크트리의 기준 브랜치를 서로 다르게 적는다.

완료된 SPEC-CODEX-REVIEW-TARGET-001 §A.7 은 감사 경로의 기준 결정을 기록하고 있으며, 카드 t1426 이 그 결정을 뒤집었지만 그 SPEC 의 HISTORY 에는 기록되지 않았다.

선택지 (c) 는 운영자가 정해야 하는 cutover 단계다. SPEC-GITHUB-FLOW-DEFAULT-001 의 M5 묶음이며, 그 계획은 이를 CUTOVER-TIME 으로 분류하고 "운영자가 직접 수행하는 단계"의 목록에 넣었다(`.moai/specs/SPEC-GITHUB-FLOW-DEFAULT-001/plan.md:12`, `:105`, `:111`, `:154–156`). 계획 문서의 문언만 확인했고 실행 상태는 관측하지 않았다.

Options:
- (a) 최근접 후보를 두 키의 후보 합집합에서 고른다. 키 불일치는 판정을 막지 않는다.
- (b) 문언대로: 키가 다른 ref 를 가리키면 inconclusive 를 두 이름과 함께 낸다. 이 저장소에서는 설정이 정렬되기 전까지 검토가 열리지 않는다.
- (c) 설정 정렬을 먼저 한다(git-flow 잔재 정리). SPEC-GITHUB-FLOW-DEFAULT-001 M5 묶음이며 운영자가 수행하는 cutover 단계다. 이 SPEC 은 정렬 뒤에 진행한다.

Default: (a) (rule: 순위 3, 사용자에게 보이는 표면이 작은 것 — `.claude/agents/moai/manager-spec.md:119`). 순위 1(현행 동작 보존): 세 선택지 모두 현행의 턴 종료 결과를 바꾼다. (c) 는 설정 값을 바꾸므로 그 값을 읽는 쪽의 현행 동작을 바꾸고, (a) 는 키 불일치 구성에서 턴 종료 기준을 바꾸며, (b) 는 이 저장소의 모든 턴 종료 판정을 inconclusive 로 만든다. 순위 1 로는 선택지가 나뉘지 않는다. 순위 2(이 SPEC 커밋 하나의 되돌리기): (c) 의 되돌리기는 SPEC-GITHUB-FLOW-DEFAULT-001 M5 묶음의 되돌리기이며 이 SPEC 의 커밋이 아니므로 탈락한다. (a)와 (b)는 이 SPEC 의 커밋 되돌리기로 원상복구된다. 순위 3: (b) 는 이 저장소의 모든 턴 종료에 inconclusive 메시지를 내고 판정을 쓰지 않으므로 사용자에게 보이는 표면이 크다. (a) 가 순위 3 을 만족한다. 이 기본값은 추천이 아니며, 운영자 판정 (a) 와 같은 방향이지만 근거는 순위 3 의 적용 결과다.
Alternate: (b).
Operator verdict: (a) 합집합에서 최근접 후보 — 보드 기록 d-20261010T053034Z-c50f
Operator verdict (amendment): (가) 합집합 유지 + 활성 흐름 가드 — 영실이 판단(운영자 위임), board d-20261010T063222Z-43e0

---

### Q2: 턴 종료 연속 차단 상한 C 의 값은?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 "연속 차단을 상한으로 막는다"고만 적었고 값을 주지 않았다. 현행 동작: 한 턴 안에서 차단된 뒤의 재종료는 stop-hook-active 플래그로 허용된다(`internal/cli/codex_review_gate.go:109`). 턴을 넘는 연속 차단에는 현재 상한이 없다. 같은 키의 fail 이 남아 있으면 새 턴마다 다시 차단된다. 다른 축의 상한도 있다. 측정되지 않은 결과의 연속 횟수 상한 `StopUnmeasuredCap = 3`(`internal/codexwiring/stop_budget.go:153`)은 차단 횟수가 아니므로 직접 대응하지 않는다.
Options:
- (a) C = 1
- (b) C = 2
- (c) C = 3
- (d) 다른 값(운영자 지정)
Default: C = 1 (rule: 순위 3, 사용자에게 보이는 표면이 작은 것 — `.claude/agents/moai/manager-spec.md:119`). 순위 1(현행 동작 보존)은 C 의 어떤 값으로도 충족되지 않는다. 현행에는 턴 사이 상한이 없기 때문이다. 순위 2(단일 되돌리기)는 세 선택지 모두 만족한다. 순위 3 에서 차단 횟수가 가장 적은 C = 1 이 선택된다. 이 기본값은 추천이 아니다.
Alternate: C = 2 또는 C = 3.
Operator verdict: C = 1 — 보드 기록 d-20261010T053834Z-422b

---

### Q3: 실행 단위 스위치의 형태는?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 "환경 값 또는 주 체크아웃의 미추적 재정의 중 하나"를 허용했다. 현재 스위치는 추적 파일의 키다(`.moai/config/sections/workflow.yaml` 302행, HEAD 값 `true`). 셸 층은 `CLAUDE_PROJECT_DIR` 의 workflow.yaml 을 읽고(`.claude/hooks/moai/handle-codex-review-entry.sh:33-34`), Go 층은 `auditreceipt.StoreRoot` 가 가리키는 트리의 workflow.yaml 을 읽는다(`internal/cli/codex_review_gate.go:469`, `:489`). 두 방식의 전파 조건이 다르다. 환경 값은 레인 터미널마다 설정해야 할 수 있다. 미추적 재정의는 주 체크아웃 한 곳에 두지만, 사용자에게 보이는 파일이 하나 생긴다.

판정 기록 뒤 명시한 사항(보드 기록 d-20261010T054047Z-3fc2, D11 = 커밋된 설정값을 따름):
- 재정의 파일이 없을 때의 기본값은 커밋된 `.moai/config/sections/workflow.yaml` 의 `codex.review_gate.enabled` 값이다. HEAD 에서 그 값은 `true` 이므로(302행) 파일이 없으면 검토가 켜진다. 현재 Go 층은 빈 루트를 비활성으로 읽는다(`internal/cli/mcp_codex.go:2503-2505`). 그래서 기본값을 명시해야 조용한 비활성을 막는다.
- 재정의 파일은 주 체크아웃의 `.moai/state/codex-review-gate/run-switch.json` 에 둔다. 이 경로는 `.gitignore` 401행으로 무시된다. 이름과 형식은 run 단계의 설계 세부다.
- "run" 은 factory run 하나다. `--run` 플래그가 가리키는 factory run id 로 지칭되며, 활성 run 은 하나다(`internal/cli/factory_card.go:1622` 의 도움말 "default: the single active run").
Options:
- (a) 환경 값(이름과 읽기 순서는 run 단계가 정한다)
- (b) 주 체크아웃의 미추적 재정의 파일
- (c) 둘을 우선순위로 함께 둔다: 환경 값이 있으면 그것을, 없으면 재정의 파일을 쓴다
Default: 없음 (rule: 순위 3 의 결과가 두 선택지를 나누지 않는다 — `.claude/agents/moai/manager-spec.md:119`). 순위 1 은 두 선택지 모두 현행의 추적 키를 바꾸므로 나뉘지 않는다. 순위 2 는 두 선택지 모두 이 SPEC 의 커밋으로 되돌릴 수 있다. 순위 3: 환경 값은 파일을 늘리지 않지만 레인마다 설정이 필요하고, 재정의 파일은 한 곳에 있지만 파일을 하나 추가한다.
Alternate: 없음.
Operator verdict: (b) 주 체크아웃의 미추적 재정의 파일 — 보드 기록 d-20261010T053834Z-422b

---

### Q4: inconclusive 영수증 뒤의 재검토 간격은?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드 M3 은 inconclusive 를 재사용하지 않는다고만 한다. 그러면 턴이 끝날 때마다 검토가 다시 시작되고, 진행 중 표식은 중복 시작만 막는다. 검토 한 번에 걸리는 시간은 연구 입력(lens-1)에서 3–9분으로 기록되어 있으나 이 세션에서 재측정하지 않았다. 레인 저장소 집계(기계 로컬, 재관측 불가)는 `codex-review` 항목 117건 가운데 inconclusive 48, fail 48, pass 21 이다. 재검토 비용의 실측은 없다.
Options:
- (a) 간격 없음: 다음 턴 종료에서 다시 시작하고, 진행 중 표식으로만 중복을 막는다.
- (b) 고정 쿨다운 뒤 재시작. 후보 값은 검증 신선도 기본값 10분(`internal/verify/freshness.go:7`)이다.
Default: 없음 (rule: 순위 1 — 현행은 inconclusive 를 재사용하므로 재검토가 없고, 두 선택지 모두 그 상태를 바꾼다. 순위 2 와 3 은 두 선택지를 나누지 않는다. `.claude/agents/moai/manager-spec.md:119`).
Alternate: 없음.
Operator verdict: (a) 간격 없음 — 보드 기록 d-20261010T053834Z-422b

---

### Q5: 두 리뷰 계층이 함께 켜졌을 때의 doctor 경고를 이 SPEC 에 넣는가?

Label: FOUNDER
Class: product-level (doctor 의 기본 출력이 바뀐다)
Authority anchor: 없음 (FOUNDER)
Why unresolved: 카드는 두 계층을 함께 켜면 doctor 경고가 뜬다고 적었다. 관측 결과, `internal/cli/doctor*.go` 71개 파일 전체 검색에서 두 계층을 함께 점검하는 검사가 없다. 명령: `ls internal/cli/doctor*.go | wc -l` → 71; `grep -l -i -E 'card-review|turn-end|review.gate|reviewgate|codex.review|review_gate' internal/cli/doctor*.go` → 출력 없음(research.md §5 P-6). 같은 주장은 규칙 문서 한 줄(`.claude/rules/moai/workflow/factory-dispatch-detail.md:157`)에만 있다. 검사를 추가하면 doctor 출력이 바뀌고, 추가하지 않으면 카드의 전제 하나가 거짓으로 남는다.
Options:
- (a) 이 SPEC 에 doctor 검사를 추가한다.
- (b) 추가하지 않고, 후속 항목으로 등록하며 규칙 문장을 정정한다.
Default: (b) 추가하지 않음 (rule: 순위 1 — 현재 doctor 출력에는 해당 경고가 없으므로 (b) 가 현행 동작을 보존한다. `.claude/agents/moai/manager-spec.md:119`).
Alternate: (a).
Operator verdict: (b) — 보드 기록 d-20261010T054047Z-3fc2

---

### Q6: SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 기록 처분은?

Label: FOUNDER
Class: product-level
Authority anchor: 없음 (FOUNDER. 대상 SPEC 이 in-progress 이므로 HISTORY 개정 행이 없다.)
Why unresolved: 이 SPEC 의 M2 는 진입 시 차단을 없애므로, 진행 중인 SPEC-GATE-BOTTLENECK-001 의 REQ-GBN-002(`spec.md:43–47`: "a fresh FAIL at turn entry SHALL block")와 충돌한다. 같은 설계는 진입 훅의 헤더 주석(`internal/cli/codex_review_delay.go` 1–14행)과 AC-GBN-004(`spec.md:67–68`)에 있다.

처분 범위에 함께 든 요구가 둘 더 있다. REQ-GBN-004(`spec.md:51–54`)는 "every path above SHALL preserve the fail-open posture (a missing/inconclusive reviewer allows)"를 요구하고, AC-GBN-004 의 inconclusive 절은 실패 시 허용을 단언한다. 이 SPEC 은 REQ-GBN-004 를 뒤집지 않는다. inconclusive 는 허용되기 때문이다(REQ-RG-010). 그래도 두 요구는 처분 대상에 포함한다. 처분은 Q6 = (b) 이며, 이 SPEC 의 HISTORY 와 Follow-up F5 에만 기록하고 대상 SPEC 은 편집하지 않는다. 잔여 위험: 대상 SPEC 은 in-progress 이고, 편집하지 않으면 코드의 진입 차단 동작과 REQ-GBN-002 가 주석 없이 어긋난 상태로 남는다.
Options:
- (a) 대상 SPEC 에 개정 행을 추가한다(진행 중이므로 같은 파일을 편집한다).
- (b) 이 SPEC 의 HISTORY 와 Follow-up 에만 기록하고, 대상 SPEC 은 건드리지 않는다.
- (c) 이 SPEC 이 대상 SPEC 의 해당 요구를 대체한다고 선언한다. 필요한 구조 변경은 운영자 판정을 거친다.
Default: (b) (rule: 순위 1 — 규칙의 문언은 "현행 동작 보존"이다. 세 선택지 모두 M2 의 진입 동작 변경과 함께 가므로 동작으로는 나뉘지 않는다. 순위 2 — (b) 는 대상 SPEC 을 편집하지 않아 이 SPEC 커밋의 되돌리기만으로 원상복구된다. (a) 는 대상 SPEC 파일을 편집하고, (c) 는 대상 SPEC 의 구조 변경을 동반하므로 되돌리기 범위가 이 SPEC 밖으로 넘어간다. `.claude/agents/moai/manager-spec.md:119`).
Alternate: (a) 또는 (c).
Operator verdict: (b) — 보드 기록 d-20261010T054047Z-3fc2

---

## 2회차 판정 기록 (Q 행 밖)

F-2 — plan.md:111 correction approved — 영실이 판단(운영자 위임), d-20261010T063222Z-43e0

문서 불일치(AGENTS.md:94–95 · :166–167 과 `.claude/rules/local/gitflow-lane-protocol.md:16`) — record only, transferred to t1635 — 영실이 판단(운영자 위임), d-20261010T063222Z-43e0

progress.md 를 M1–M7 공통 승인 파일로 plan.md §C 에 등록하고, 절 (a) 우선 판독(키 읽기 실패는 항상 절 (d); manual.workflow 부재·알 수 없음·모순은 두 키가 서로 다른 브랜치일 때에만 절 (d); 절 (a) 가 맞으면 manual.workflow 미판독; 리더 이의 없음 확인) — 리더 판정(승인), d-20261010T065057Z-229e

---

## 보드 판정 기록 (Q 행 밖)

보드 판정 세 건은 Q 행이 아니므로 여기에 원문으로 기록한다. 같은 보드 기록(d-20261010T054047Z-3fc2)에 Q5·Q6 과 함께 있다.

- D3 — Operator verdict: 키당 총 C회로 제한 (허용이 카운터를 되돌리지 않음) — 보드 기록 d-20261010T054047Z-3fc2. 반영: REQ-RG-004, AC-RG-005, AC-RG-017.
- D4 — Operator verdict: 항상 메시지 표시 — 보드 기록 d-20261010T054047Z-3fc2. 반영: REQ-RG-016, AC-RG-005, AC-RG-010, AC-RG-018, AC-RG-019.
- D11 — Operator verdict: 커밋된 설정값을 따름 — 보드 기록 d-20261010T054047Z-3fc2. 반영: REQ-RG-011, REQ-RG-012, AC-RG-012, AC-RG-013.
