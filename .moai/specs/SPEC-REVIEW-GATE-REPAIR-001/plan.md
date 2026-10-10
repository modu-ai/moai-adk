# SPEC-REVIEW-GATE-REPAIR-001 — 구현 계획

## §A 맥락

- **Tier L 판정.** 변경 표면이 네 서브시스템(진입 훅, 턴 종료 게이트, 영수증 저장과 재사용, 리뷰 기준 해석)과 설정 스위치에 걸친다. 인용하는 코드 파일이 15개를 넘고, 요구사항 7개와 인수 기준 11개를 가진다(각각 상한 25). v0.2.0 에서 기준 선택 요구 4개(v0.1.2 의 폐기 식별자, HISTORY v0.2.0 행)와 인수 기준 10개(v0.1.2의 AC-RG-006–009·016·020–024)를 t1662 로 이관했다. 리서치 산출물(research.md)이 필요하므로 설계 산출물(design.md)과 함께 다섯 개 산출물을 둔다. 여기에 progress.md 와 decision-index.md 를 더한다. decision-index.md 는 `interview.decision_gate: on` 이므로 필수다. [ID mapping v0.2.2: 적용되는 009–016 대응 없음. 이 항목의 v0.1.2 폐기 식별자 네 개는 v0.2.0 에서 폐기된 식별자다.]
- **현재 상태.** 운영자 판정 Q1–Q6 과 보드 판정 D3·D4·D11 은 decision-index.md 에 기록되었다. 플랜 감사 1회차는 FAIL(score 0.625)이었고, 이 개정은 그 D1–D12 를 수리한다. 플랜 감사 2회차(codex, 영수증 rcpt-0efbcef22b780f388f658933)는 FAIL 이었고, 이 개정(spec v0.1.2)은 그 F-1·F-2 를 수리한다. 구현 코드는 건드리지 않았다. 재감사를 통과해야 Implementation Kickoff 를 열 수 있다. v0.2.0 은 분할 판정 d-20261010T070728Z-6cdf(영실이 판단, 운영자 위임)를 적용한 축소본이다. 계획 감사 3회차의 F-3 은 결함이므로 기준 브랜치 선택(카드 (2)항)을 t1662 로 떼어 냈다. 이 개정 뒤 계획은 다음 계획 감사가 통과하기 전까지 audit-ready 가 아니다.
- **이 계획이 하지 않는 것.** 구현 코드 변경은 run 단계의 몫이다. F1·F2·F3·F5 는 구현 범위 밖이다. F4 의 규칙 문장 정정과 F1–F5 의 후속 등록은 M7 에서 한다.

## §B 되돌리기 어려운 결정 (구현 전 확정)

가장 바뀌기 쉬운 결정부터 순서대로 적는다. 각 항목은 decision-index.md 의 같은 번호와 짝을 이룬다. 여섯 판정은 모두 기록되었다. 판정 원문과 보드 기록 번호는 decision-index.md 의 `Operator verdict` 행에 있다.

1. **[t1662 로 이관, 판정 d-20261010T070728Z-6cdf] Q1 — 두 설정 키가 다른 ref 를 가리킬 때의 기준 규칙.** 판정: (a) 합집합에서 최근접 후보 (보드 기록 d-20261010T053034Z-c50f). 개정 판정: (가) 합집합 유지 + 활성 흐름 가드 (영실이 판단, 운영자 위임, 보드 기록 d-20261010T063222Z-43e0; 보류 d-20261010T063038Z-f1a1 을 해소). 후보 집합은 `manual.workflow` 등 활성 흐름 키가 정하는 네 절로 고른다. 두 키의 후보 합집합에서 최근접 후보를 고르며, 키 불일치는 판정을 막지 않는다. 결과: t1662 로 이관(REQ-ACE-005). 이 SPEC 에서 제거된 ID 는 spec.md HISTORY v0.2.0 에 있다.
2. **Q6 — SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 기록 처분.** 판정: (b) (보드 기록 d-20261010T054047Z-3fc2). 이 SPEC 의 HISTORY 와 F5 에만 기록하고 대상 SPEC 은 편집하지 않는다. 결과: spec.md §A.4, F5.
3. **Q3 — 실행 단위 스위치의 형태.** 판정: (b) 주 체크아웃의 미추적 재정의 파일 (보드 기록 d-20261010T053834Z-422b). 결과: 후속 카드 t1674 로 이관.
4. **Q2 — 연속 차단 상한 C.** 판정: C = 1 (보드 기록 d-20261010T053834Z-422b). 차단 횟수는 키당 총량으로 제한된다(D3). 결과: REQ-RG-003·004, AC-RG-004·005·017.
5. **Q4 — inconclusive 뒤 재검토 간격.** 판정: (a) 간격 없음 (보드 기록 d-20261010T053834Z-422b). 결과: 후속 카드 t1674 로 이관.
6. **Q5 — 두 계층 동시 활성 시 doctor 경고.** 판정: (b) (보드 기록 d-20261010T054047Z-3fc2). 이 SPEC 에서 doctor 검사를 추가하지 않고 F4 로 등록한다. 결과: spec.md §E, F4.

보드 판정 D3·D4·D11 도 같은 보드 기록(d-20261010T054047Z-3fc2)으로 기록되었다. 차단 횟수는 키당 총량이며 허용이 되돌리지 않는다(REQ-RG-004). 검토 없이 허용되는 턴 종료는 항상 메시지를 낸다(REQ-RG-007). 스위치의 기본값은 커밋된 값이다(후속 카드 t1674 에서 구현).

## §C 기술 접근 — 파일 지도

현재 역할은 research.md 의 인용 등록부에서 확인한 것이다. 줄 번호는 기준 코드 트리(`2aab5f797`)의 값이다.

| 파일 | 현재 역할 (근거) | 관련 마일스톤 |
|---|---|---|
| `internal/cli/codex_review_delay.go` | 진입 훅 `HandleCodexReviewEntry`(140행). 차단 분기 169–183행. 헤더 1–14행은 교착을 설계로 채택했다. | M2 |
| `internal/cli/codex_review_delay_test.go` | 기존 테스트 파일. `TestReviewEntry_FreshFailBlocks`(363행)는 결함을 고정하므로 기대값을 바꾼다(셀 B-2). M2 의 기대값 변경 대상(진입 차단 단언 :195–197, :267–269) | M2 |
| `internal/cli/codex_review_entry_delivery_test.go` | 신규 RED 테스트 파일(AC-RG-001–003, 진입 훅 비차단과 전달 기록). M2 에서 만든다 | M2 |
| `internal/cli/codex_review_gate.go` | 턴 종료 `HandleCodexReviewGate`(104행). stop_hook_active 109행. M3 의 턴 종료 상태표(차단 횟수 카운터, 메시지)는 이 파일의 턴 종료 처리에 둔다. 재검토 시작은 t1674 몫이다 | M3 |
| `internal/cli/codex_review_gate_test.go` | 기존 테스트 파일(턴 종료 게이트 가드): `TestReviewGate_DisabledAllows`(40행), `TestReviewGate_LoopPreventionAllows`(65행), `TestReviewGate_FailOpenOnMissingCodex`(119행). M3 의 신규 테스트 함수(AC-RG-004 차단 횟수 1 단언, AC-RG-005, AC-RG-017 신규 가드(허용 뒤 재차단 없음), AC-RG-018·019 메시지 단언)는 이 파일에 둔다. M3 은 기존 테스트의 기대값을 바꾸지 않는다. `TestReviewGate_CacheHitReusesFailVerdict`(`internal/cli/codex_review_cache_test.go:68`)는 M2 파일이므로 M3 에서 고치지 않는다. M7 회귀 실행은 이 파일을 읽기만 한다 | M3 |
| `internal/cli/codex_review_cache.go` | 유효기간 상수 29행. 상세 파일 읽기 72–76행 | M2 |
| `internal/cli/codex_review_cache_test.go` | 기존 테스트 파일. M2 의 기대값 변경 대상(진입 차단 단언 :216–227) | M2 |
| `internal/cli/codex_review_gate_live_test.go` | 기존 테스트 파일(codex 가 있는 환경에서만 실행). 진입 차단 단언 `:129–133` 은 M2 의 기대값 변경 대상이다 | M2 |
| `.claude/settings.json` 272행, `internal/template/templates/.claude/settings.json.tmpl` 272행 | 진입 훅 timeout 5 (F2 의 대상, 이 계획은 바꾸지 않는다) | — |
| `internal/verify/receipt.go` 63–96행, 70–76행, 90행 | 결합 필드 다섯 개. 유효기간 0 은 기본값(10분)으로 바뀐다 | M2 |
| `.claude/rules/moai/workflow/factory-dispatch-detail.md` 157행 과 템플릿 미러 `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md` | 두 계층 동시 활성 문장. F4 의 정정 대상 | M7 |
| `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/progress.md` (공통 승인 파일) | 현재 §E.1 만 채워진 진행 기록. 마일스톤별 기록(RED 출력 §E.2, 기대값 변경의 근거, 회귀 비교표)이 여기에 든다. 모든 마일스톤의 승인 파일 목록에 공통으로 든다 | M1–M3, M7 (공통) |

후보 집합, 거리 척도, 동점 규칙은 t1662 가 정한다(REQ-ACE-005). 파일 목록은 이 §C 가 정한다. run 단계가 목록 밖 파일을 바꿔야 하면 멈추고 보고한다(마일스톤 종료 시 트리 변화 점검). 목록이 Tier L 의 상한을 넘으면 Tier 판정을 다시 한다.

`.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/progress.md` 는 M1–M3·M7 모든 마일스톤의 공통 승인 파일이다(위 표의 마지막 행). M4 는 t1662 로 이관되어 이 SPEC 에 없다. 따라서 각 마일스톤의 승인 파일 목록은 그 마일스톤에 묶인 행과 이 공통 행을 합친 것이다.

## §D 마일스톤 (우선순위 순)

각 마일스톤의 RED 선행은 커밋 그래프만으로 검증한다. 마일스톤 커밋 M 의 첫 부모 P 의 트리에 덮어쓰기 단위를 얹어 기록된 명령을 실행하면, 기록된 단언에서 EXPECTED_RED 로 실패해야 한다. 덮어쓰기 단위는 마일스톤 커밋이 추가하거나 바꾼 _test.go 파일의 그 커밋 시점 내용이다. 덮어쓰기 단위의 테스트는 첫 부모의 비테스트 코드만으로 컴파일되어야 한다. 컴파일 실패는 tdd-result-contract.md 의 TOOL_FAILURE 이며 RED 로 세지 않는다. 검사 대상은 RED 단계가 있는 M2·M3 이다(M1·M7 은 RED 단계가 없어 제외한다). 그 실행의 명령, 종료 코드, 원문 출력은 progress.md §E.2 에 남긴다. 마일스톤의 순서는 되돌리기 어려운 결정을 먼저 두고, 기계적인 정리는 뒤에 둔다. 각 마일스톤은 한 커밋이며, 두 면을 가진 파일은 그 커밋에서 함께 갱신한다(spec.md §D.3).

### M1 — 판정 기록 확인 (우선순위 High, 차단 게이트)

- Q2–Q6 과 보드 판정 D3·D4·D11 은 decision-index.md 에 기록되었다(plan 단계 완료). Q1 은 t1662 로 이관되었다. run 단계는 Q2–Q6 행의 `Operator verdict:` 가 채워져 있는지 읽어 확인한다.
- 완료 조건: Q2–Q6 행의 Operator verdict 가 모두 채워져 있고, 보드 판정 세 건이 같은 색인에 있다. Q1 의 개정 판정 행과 2회차 기록(d-20261010T063222Z-43e0), 3회차 분할 기록(d-20261010T070728Z-6cdf)도 같은 색인에 있다.
- 다음 plan-audit 통과가 이 마일스톤의 전제다(수리 2회차 뒤의 재감사). 재감사 전에는 run 단계로 넘어가지 않는다.
- Q6 처분은 (b) 이므로 대상 SPEC 의 편집을 여는 별도 작업은 없다.

### M2 — 진입 훅 비차단과 전달 기록 (우선순위 High, 사용자 흐름 변경과 새 데이터 모델)

- 요구: REQ-RG-001, REQ-RG-002. 인수 기준: AC-RG-001, AC-RG-002, AC-RG-003.
- RED: fail 영수증이 있어도 진입 훅은 차단하지 않는다(AC-RG-001). 같은 키의 첫 프롬프트에서 상세가 추가 맥락으로 한 번 전달되고 전달 기록이 남는다(AC-RG-002). 전달 기록이 있는 키에서는 반복하지 않는다(AC-RG-003).
- GREEN: 진입 경로의 차단 분기를 추가 맥락 전달로 바꾸고, 키 단위 전달 기록을 둔다. 기록은 영수증 결합 필드 다섯 개의 조합을 키로 쓴다.
- 기대값 변경: `TestReviewEntry_FreshFailBlocks` 는 결함을 고정하고 있으므로(acceptance.md §C, 셀 B-2) 같은 변경에서 기대값을 바꾼다. 바꾼 내용은 progress.md 에 "이전 기대값 → 새 기대값 → 근거 결정" 형식으로 남긴다. 근거는 REQ-RG-001 이다. 같은 이유로 `codex_review_cache_test.go:216–227`, `codex_review_delay_test.go:195–197`·`:267–269`, `codex_review_gate_live_test.go:129–133` 의 진입 차단 단언도 바꾼다.
- 범위 밖: 진입 훅의 timeout 과 시간 상한은 F2 의 범위다. 이 마일스톤은 건드리지 않는다.

### M3 — 턴 종료 상태표: 차단 횟수, 메시지 (우선순위 High, 상태 카운터)

- 요구: REQ-RG-003, REQ-RG-004, REQ-RG-007. 인수 기준: AC-RG-004, AC-RG-005, AC-RG-017, AC-RG-018, AC-RG-019, AC-RG-025.
- RED: 차단 횟수가 C(= 1)에 이른 뒤 턴 종료가 허용되고 미검토 상태가 systemMessage 로 드러난다(AC-RG-005). 허용 뒤의 다음 턴 종료도 차단하지 않는다(AC-RG-017). absent·stale·key-mismatch 턴 종료는 허용하고 상태 이름을 메시지로 낸다(AC-RG-018). 현재 키의 fresh inconclusive 턴 종료는 허용하고 systemMessage 에 `inconclusive` 를 낸다(AC-RG-025). 재검토 시작은 t1674 몫이다. 검토 없이 허용되는 턴 종료(스위치 꺼짐, 검토할 변경 없음, stop-hook 해제, reviewer 없음)는 상태를 명명한 메시지를 낸다(AC-RG-019).
- GREEN: 리뷰 키 단위 차단 횟수 카운터를 둔다. 차단할 때 1 올린다. 허용은 값을 바꾸지 않는다. 같은 키의 fresh pass 기록에서 제거한다. 키가 바뀌면 새 키는 0 에서 시작한다. C = 1(Q2). 메시지 상태 어휘는 design.md §7 의 아홉 상태를 쓴다.
- 차단 사유의 형식은 바꾸지 않는다(F2 의 범위).
- 기대값 변경: 현재 테스트에는 허용 뒤 재차단이 없는지를 단언하는 가드가 없다. 그 거동은 신규 가드(AC-RG-017)로 추가한다. 기존 `TestReviewGate_CacheHitReusesFailVerdict` 는 첫 차단만 본다(RG-GUARD).

### M7 — 회귀 정비와 문서 참조 정정 (우선순위 Low, 기계적)

- 요구: REQ-RG-005, REQ-RG-006(회귀). 인수 기준: AC-RG-014, AC-RG-015, AC-RG-019(가드 부분).
- 회귀 가드 전체를 변경 전후로 다시 실행하고 비교표를 progress.md 에 남긴다. 여기에는 `TestReviewGate_FailOpenOnMissingCodex` 처럼 셀 B-1 에 없는 reviewer 부재 가드를 포함한다.
- F4 의 규칙 문장(`.claude/rules/moai/workflow/factory-dispatch-detail.md:157`)을 정정한다. 그 파일과 템플릿 미러를 같은 커밋에서 맞춘다.
- F1–F5 를 후속 카드로 등록하고, F6 은 카드 t1635(3.2-1-2)로 이관한다. 이 SPEC 의 Follow-up 항목을 그대로 옮긴다.

M4 는 t1662 가 맡는다(REQ-ACE-005, 리뷰 기준의 단일 해석 함수). M5 와 M6 은 후속 카드 t1674 로 옮겼다(depends t1620). 이 SPEC 의 마일스톤은 M1–M3 과 M7 이며, 번호 간격은 의도한 것이다.

## §E 자기 검증 — run 단계 명령 형식

- 기준 코드 확인: `git diff --stat 2aab5f797b75983e132af451da68f69e3426557b HEAD -- ':!.moai/specs/SPEC-REVIEW-GATE-REPAIR-001'` 의 출력이 비어 있어야 한다. 비어 있지 않으면 멈추고 보고한다. 설치된 빌드 버전 문자열로 기준을 판정하지 않는다.
- 테스트 호출 형식: `go test -count=1 -v -run '^<TestName>$' ./internal/cli/` 처럼 단일 호출로, 출력을 원문으로 남긴다. 종료 코드는 별도 필드로 기록한다.
- 빌드 출처: 측정은 트리의 소스에서 만든 빌드로 한다. 설치된 빌드의 버전 문자열만으로 판정하지 않는다(verification-claim-integrity.md §2.2).
- 공허 검사 방지: 실행된 테스트 수(`--- PASS` 줄 수)를 함께 기록한다. 0이면 측정 불가로 보고한다.
- 트리 변화 점검: 각 마일스톤 종료 시 `git status --short` 를 본다. 변화가 그 마일스톤의 승인 파일 목록(§C 파일 지도에서 그 마일스톤에 묶인 파일) 밖에 있으면 멈추고 보고한다. `.moai/config/sections/workflow.yaml` 은 어느 마일스톤에서도 건드리지 않는다(spec.md §F 의 제약, §E 범위 밖).
- 변경 세트 점검: 마일스톤 커밋마다 `git show --stat <sha>` 로 두 면(`.claude/` 와 `internal/template/templates/`)이 같은 커밋에 있는지 확인한다.

## §F 위험

1. **[t1662 로 이관, 판정 d-20261010T070728Z-6cdf] Q1 위험은 개정 판정 (가) 로 해소되었다**(보드 기록 d-20261010T063222Z-43e0). 문언대로 키 불일치를 inconclusive 로 두면 이 저장소의 모든 판정이 inconclusive 가 되는데, 절 (b) 의 합집합이 이를 막는다. 남은 위험은 절 (a)·(c)·(d) 의 구현이다(github-flow 에서 develop 이 섞이는 경우, 읽기 실패가 pass 로 저장되는 경우). 이 구현과 그 검증은 t1662(REQ-ACE-005)가 맡으며, 이 SPEC 은 구현하지 않는다.
2. **결함을 고정한 테스트의 기대값 변경을 회귀로 오인하기 쉽다.** 두 테스트 모두 현재 PASS 이며(셀 B-2), 그 PASS 가 결함의 증거다. 기대값 변경은 근거와 함께 기록한다.
3. **템플릿 미러 누락.** 래퍼가 두 면에 있다. 두 면은 t1674 의 한 커밋에 함께 담는다. 한 면만 바꾸면 배포본과 저장소 본이 갈린다.
4. **스위치 해석의 두 층 불일치.** 셸 층과 Go 층이 주 체크아웃 해석기를 다르게 쓰면 같은 run 에서 두 층이 다른 값을 낸다(후속 카드 t1674 에서 확인한다).
5. **커밋된 값 판독은 현행 동작과 다르다.** 현재 Go 층은 작업 트리의 workflow.yaml 을 읽는다. D11 에 따라 기본값은 HEAD 의 커밋된 값이다. 작업 트리의 미커밋 변경은 무시된다. 이 행위 변경은 후속 카드 t1674 에서 다룬다. 운영자가 해석을 확인할 대상이다.
6. **메시지 범위는 넓다.** REQ-RG-007 은 문언대로 스위치 꺼짐, 검토할 변경 없음, stop-hook 해제 턴 종료에도 메시지를 낸다. 그런 턴 종료가 많으면 메시지가 자주 나온다. 범위는 운영자가 확인할 대상이다(보드 판정 D4 "항상 메시지 표시"의 범위를 그대로 옮긴 것이다: 새 통과 검토 없이 허용되는 모든 턴 종료. 감사 보고서가 든 키 한정 선택지보다 넓다).
7. **재정의 파일의 stale 바인딩.** 단일 활성 run 을 해석하지 못하면 재정의 파일은 무시되고 커밋된 값이 적용된다. 실행 식별자 해석은 design.md §5 의 단일 활성 run 규칙을 따른다(t1674 몫).
8. **설치 빌드 오인.** 설치된 빌드(`4f8aba061`)는 트리보다 새 커밋이다. 설치 빌드의 결과를 트리의 증거로 인용하면 도구 출처 규칙을 어긴다.
9. **리더의 미커밋 workflow.yaml 변경은 이 워크트리에서 재현되지 않았다.** 측정은 HEAD 기준이다. 리더가 같은 파일을 다시 바꾸면 후속 카드 t1674 의 판독 경로가 그 값을 무시하는지 확인한다.
10. **영수증과 로그는 기계 로컬이다.** 레인 저장소의 스냅샷과 상세 파일은 이 저장소의 커밋 대상이 아니다. 인용할 때는 기계 로컬 근거임을 밝힌다.
11. **완료된 결정의 뒤집힘이 SPEC 본문에 없다.** 카드 t1426 의 감사 경로 결정 뒤집힘은 Q1 행과 HISTORY 에만 기록된다. SPEC-CODEX-REVIEW-TARGET-001 의 본문은 편집하지 않는다.
12. **[t1662 로 이관, 판정 d-20261010T070728Z-6cdf] 절 (d) 의 모순 조건과 순서 충돌.** 세 키의 조합 가운데 절 (a)–(c) 에서 모순으로 확인된 조합은 없다. 순서 규칙이 충돌을 정한다. (X1) `manual.develop_branch` 가 `main` 이고 `manual.workflow` 가 `github-flow` 이며 `worktree_base_branch` 가 `main` 이 아니면, 절 (a) 와 절 (c) 가 서로 다른 단일 기준을 낸다. 순서상 절 (a) 가 이긴다. (X2) `manual.develop_branch` 가 `main` 이거나 `worktree_base_branch` 가 `manual.develop_branch` 와 같은데 `manual.workflow` 가 없거나 알 수 없으면, 절 (a) 가 `manual.workflow` 를 읽지 않으므로 절 (d) 의 워크플로 조건은 적용되지 않는다. (X3) `manual.workflow` 가 `git-flow` 인데 `manual.develop_branch` 나 `worktree_base_branch` 가 비어 있거나 없으면, 절 (b) 가 합집합을 남은 한쪽 키로 줄여 적용하며, 절 (d) 의 키 읽기 조건은 적용되지 않는다. X1 과 X3 을 inconclusive 로 바꿀지는 운영자 확인 대상이다. 구체적 모순 조합이 확정될 때까지 이 조합의 픽스처는 t1662(REQ-ACE-005)의 검증 몫으로 남으며, 이 SPEC 은 두지 않는다. (리더 판정 d-20261010T065057Z-229e)
13. **[t1662 로 이관, 판정 d-20261010T070728Z-6cdf] 설정 파일 부재 시 키 읽기.** `git-strategy.yaml` 이 없을 때 설정 로더가 기본값을 돌려주면 키가 읽힌 것으로 보일 수 있다. 절 (d) 의 "키를 읽을 수 없음"이 이 경우를 포함하는지는 운영자 확인 대상이다. 이 경우의 픽스처는 t1662(REQ-ACE-005)의 검증 몫으로 남으며, 이 SPEC 은 두지 않는다.

## §G 안티패턴

- 차단 횟수 C = 1 이나 간격 없음(Q2; Q4 는 t1674 범위)을 run 단계에서 임의로 바꾸는 것.
- 차단 횟수를 허용마다 되돌리는 것(D3 위반), 또는 키와 무관한 전역 카운터로 바꾸는 것.
- 검토 없이 허용되는 턴 종료 중 일부 상태만 메시지를 내고 나머지를 생략하는 것(D4 위반).
- 작업 트리의 미커밋 workflow.yaml 값을 스위치 기본값으로 읽는 것(D11 위반, t1674 범위).
- 매칭이 0개인 실행을 통과로 세는 것(공허한 통과).
- 트리 SHA 대신 브랜치 이름 `develop` 에 핀하는 것(움직이는 ref).
- 결함 고정 테스트를 삭제해서 빨간불을 피하는 것. 기대값 변경만 허용된다.

## §H 참조

- spec.md · acceptance.md · design.md · research.md · decision-index.md
- `.claude/rules/moai/workflow/factory-dispatch-detail.md:157` (두 계층 문장, F4)
- `.claude/rules/moai/workflow/tdd-result-contract.md` (RED 기록 형식)
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 (도구 출처)
- `.claude/rules/moai/development/verification-completeness.md` §2.1 (RED-now 셀의 네 요소)
- `.moai/specs/SPEC-GATE-BOTTLENECK-001/spec.md` REQ-GBN-002 · REQ-GBN-004 · AC-GBN-004 (편집하지 않는다, F5)
