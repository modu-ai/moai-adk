# SPEC-REVIEW-GATE-REPAIR-001 — 구현 계획

## §A 맥락

- **Tier L 판정.** 변경 표면이 네 서브시스템(진입 훅, 턴 종료 게이트, 영수증 저장과 재사용, 리뷰 기준 해석)과 설정 스위치에 걸친다. 인용하는 코드 파일이 15개를 넘고, 요구사항 16개와 인수 기준 20개를 가진다(각각 상한 25). 리서치 산출물(research.md)이 필요하므로 설계 산출물(design.md)과 함께 다섯 개 산출물을 둔다. 여기에 progress.md 와 decision-index.md 를 더한다. decision-index.md 는 `interview.decision_gate: on` 이므로 필수다.
- **현재 상태.** 운영자 판정 Q1–Q6 과 보드 판정 D3·D4·D11 은 decision-index.md 에 기록되었다. 플랜 감사 1회차는 FAIL(score 0.625)이었고, 이 개정은 그 D1–D12 를 수리한다. 구현 코드는 건드리지 않았다. 재감사를 통과해야 Implementation Kickoff 를 열 수 있다.
- **이 계획이 하지 않는 것.** 구현 코드 변경은 run 단계의 몫이다. F1·F2·F3·F5 는 구현 범위 밖이다. F4 의 규칙 문장 정정과 F1–F5 의 후속 등록은 M7 에서 한다.

## §B 되돌리기 어려운 결정 (구현 전 확정)

가장 바뀌기 쉬운 결정부터 순서대로 적는다. 각 항목은 decision-index.md 의 같은 번호와 짝을 이룬다. 여섯 판정은 모두 기록되었다. 판정 원문과 보드 기록 번호는 decision-index.md 의 `Operator verdict` 행에 있다.

1. **Q1 — 두 설정 키가 다른 ref 를 가리킬 때의 기준 규칙.** 판정: (a) 합집합에서 최근접 후보 (보드 기록 d-20261010T053034Z-c50f). 두 키의 후보 합집합에서 최근접 후보를 고르며, 키 불일치는 판정을 막지 않는다. 결과: REQ-RG-007, design.md §4, AC-RG-007·008.
2. **Q6 — SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 기록 처분.** 판정: (b) (보드 기록 d-20261010T054047Z-3fc2). 이 SPEC 의 HISTORY 와 F5 에만 기록하고 대상 SPEC 은 편집하지 않는다. 결과: spec.md §A.4, F5.
3. **Q3 — 실행 단위 스위치의 형태.** 판정: (b) 주 체크아웃의 미추적 재정의 파일 (보드 기록 d-20261010T053834Z-422b). 결과: REQ-RG-011·012, design.md §5, AC-RG-012·013.
4. **Q2 — 연속 차단 상한 C.** 판정: C = 1 (보드 기록 d-20261010T053834Z-422b). 차단 횟수는 키당 총량으로 제한된다(D3). 결과: REQ-RG-003·004, AC-RG-004·005·017.
5. **Q4 — inconclusive 뒤 재검토 간격.** 판정: (a) 간격 없음 (보드 기록 d-20261010T053834Z-422b). 결과: REQ-RG-010, AC-RG-010·011.
6. **Q5 — 두 계층 동시 활성 시 doctor 경고.** 판정: (b) (보드 기록 d-20261010T054047Z-3fc2). 이 SPEC 에서 doctor 검사를 추가하지 않고 F4 로 등록한다. 결과: spec.md §E, F4.

보드 판정 D3·D4·D11 도 같은 보드 기록(d-20261010T054047Z-3fc2)으로 기록되었다. 차단 횟수는 키당 총량이며 허용이 되돌리지 않는다(REQ-RG-004). 검토 없이 허용되는 턴 종료는 항상 메시지를 낸다(REQ-RG-016). 스위치의 기본값은 커밋된 값이다(REQ-RG-011).

## §C 기술 접근 — 파일 지도

현재 역할은 research.md 의 인용 등록부에서 확인한 것이다. 줄 번호는 기준 코드 트리(`2aab5f797`)의 값이다.

| 파일 | 현재 역할 (근거) | 관련 마일스톤 |
|---|---|---|
| `internal/cli/codex_review_delay.go` | 진입 훅 `HandleCodexReviewEntry`(140행). 차단 분기 169–183행. 헤더 1–14행은 교착을 설계로 채택했다 | M2 |
| `internal/cli/codex_review_gate.go` | 턴 종료 `HandleCodexReviewGate`(104행). stop_hook_active 109행. 수용 영수증 재사용 159–178행. 진행 중 표식 197–210행. 스위치 읽기 469행. 빈 루트 해석 `reviewGateConfigRoot` 489행 | M3, M5, M6 |
| `internal/cli/codex_review_cache.go` | 유효기간 상수 29행. 상세 파일 읽기 72–76행 | M2, M5 |
| `internal/cli/codex_review_receipt.go` | 영수증 생산자. 키 측정 149행. inconclusive 기록 160–186행. 런타임 설정 전용 발견 사항의 재분류 176행 | M5 |
| `internal/cli/codex_review_scope.go` | `cardMergeBase` 189–212행(`IntegrationTarget` 읽기 191행). origin 처리 204–210행. 변경 집합 해시 341–397행 | M4 |
| `internal/cli/mcp_review_material.go` | `resolveReviewBase` 156–186행. step 0 의 `worktree_base_branch` 읽기 159행. `resolveReviewMergeBase` 189행은 얇은 래퍼 | M4 |
| `internal/cli/mcp_selfreview.go` | 리뷰 범위 해석 호출 119행. 실행 시간 접두 제외 목록 168행 | M4 |
| `internal/cli/mcp_codex.go` | `readCodexReviewGateEnabled` 2503행(빈 루트면 false). 감사 경로의 기준 기록 | M6 |
| `internal/cli/codex_review_tree_scope.go` | 트리와 주 체크아웃 범위 판정 133–137행 | M6 |
| `internal/config/loader_integration_branch.go` | `IntegrationTarget` 123–125행. 값 정책은 Q1 이 정했다. 이 파일은 읽기만 한다 | M4 |
| `internal/auditreceipt/storeroot.go` | `IdentifyPrimaryCheckout` 142행(주 체크아웃 해석, 재정의 파일 경로에 쓴다). `StoreRoot` 242–243행(트리에 workflow.yaml 이 있으면 그 트리를 돌려준다, 재정의 경로에는 쓰지 않는다) | M6 |
| `.claude/hooks/moai/handle-codex-review-entry.sh` | 33–34행: `CLAUDE_PROJECT_DIR` 기준 workflow.yaml 을 읽는다. 주 체크아웃 해석으로 바뀐다 | M6 |
| `.claude/hooks/moai/handle-codex-review-gate.sh` | 마지막 줄: 훅 호출에 stderr 리다이렉트가 없다 | M6 |
| `.claude/settings.json` 272행, `internal/template/templates/.claude/settings.json.tmpl` 272행 | 진입 훅 timeout 5 (F2 의 대상, 이 계획은 바꾸지 않는다) | — |
| `internal/template/templates/.claude/hooks/moai/handle-codex-review-{entry,gate}.sh` | 두 래퍼의 배포 미러. 같은 커밋에서 갱신한다(spec.md §D.3) | M6 |
| `internal/verify/receipt.go` 63–96행, 70–76행, 90행 | 재사용 판정 입력. 결합 필드 다섯 개. 유효기간 0 은 기본값(10분)으로 바뀐다 | M2, M5 |
| `.claude/rules/moai/workflow/factory-dispatch-detail.md` 157행 과 템플릿 미러 `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-detail.md` | 두 계층 동시 활성 문장. F4 의 정정 대상 | M7 |

후보 집합, 거리 척도, 동점 규칙은 design.md §4 가 정한다. 파일 목록은 run 단계에서 확정한다. 목록이 Tier L 의 상한을 넘으면 Tier 판정을 다시 한다.

## §D 마일스톤 (우선순위 순)

각 마일스톤은 RED 를 먼저 기록하고 GREEN 으로 넘어간다. RED 는 구현 전 트리에서 실행한 출력을 원문으로 남긴다(tdd-result-contract.md). 마일스톤의 순서는 되돌리기 어려운 결정을 먼저 두고, 기계적인 정리는 뒤에 둔다. 각 마일스톤은 한 커밋이며, 두 면을 가진 파일은 그 커밋에서 함께 갱신한다(spec.md §D.3).

### M1 — 판정 기록 확인 (우선순위 High, 차단 게이트)

- Q1–Q6 과 보드 판정 D3·D4·D11 은 decision-index.md 에 기록되었다(plan 단계 완료). run 단계는 여섯 Q 행의 `Operator verdict:` 가 채워져 있는지 읽어 확인한다.
- 완료 조건: 여섯 Q 행의 Operator verdict 가 모두 채워져 있고, 보드 판정 세 건이 같은 색인에 있다.
- 재감사(plan-audit 2회차) 통과가 이 마일스톤의 전제다. 재감사 전에는 run 단계로 넘어가지 않는다.
- Q6 처분은 (b) 이므로 대상 SPEC 의 편집을 여는 별도 작업은 없다.

### M2 — 진입 훅 비차단과 전달 기록 (우선순위 High, 사용자 흐름 변경과 새 데이터 모델)

- 요구: REQ-RG-001, REQ-RG-002. 인수 기준: AC-RG-001, AC-RG-002, AC-RG-003.
- RED: fail 영수증이 있어도 진입 훅은 차단하지 않는다(AC-RG-001). 같은 키의 첫 프롬프트에서 상세가 추가 맥락으로 한 번 전달되고 전달 기록이 남는다(AC-RG-002). 전달 기록이 있는 키에서는 반복하지 않는다(AC-RG-003).
- GREEN: 진입 경로의 차단 분기를 추가 맥락 전달로 바꾸고, 키 단위 전달 기록을 둔다. 기록은 영수증 결합 필드 다섯 개의 조합을 키로 쓴다.
- 기대값 변경: `TestReviewEntry_FreshFailBlocks` 는 결함을 고정하고 있으므로(acceptance.md §C, 셀 B-2) 같은 변경에서 기대값을 바꾼다. 바꾼 내용은 progress.md 에 "이전 기대값 → 새 기대값 → 근거 결정" 형식으로 남긴다. 근거는 REQ-RG-001 이다.
- 범위 밖: 진입 훅의 timeout 과 시간 상한은 F2 의 범위다. 이 마일스톤은 건드리지 않는다.

### M3 — 턴 종료 상태표: 차단 횟수, 재검토 시작, 메시지 (우선순위 High, 상태 카운터)

- 요구: REQ-RG-003, REQ-RG-004, REQ-RG-015, REQ-RG-016. 인수 기준: AC-RG-004, AC-RG-005, AC-RG-017, AC-RG-018, AC-RG-019.
- RED: 차단 횟수가 C(= 1)에 이른 뒤 턴 종료가 허용되고 미검토 상태가 systemMessage 로 드러난다(AC-RG-005). 허용 뒤의 다음 턴 종료도 차단하지 않는다(AC-RG-017). absent·stale·key-mismatch 턴 종료는 재검토를 시작하고 허용한다(AC-RG-018). 검토 없이 허용되는 턴 종료(스위치 꺼짐, 검토할 변경 없음, stop-hook 해제, reviewer 없음)는 상태를 명명한 메시지를 낸다(AC-RG-019).
- GREEN: 리뷰 키 단위 차단 횟수 카운터를 둔다. 차단할 때 1 올린다. 허용은 값을 바꾸지 않는다. 같은 키의 fresh pass 기록에서 제거한다. 키가 바뀌면 새 키는 0 에서 시작한다. C = 1(Q2). 메시지 상태 어휘는 design.md §7 을 쓴다.
- 차단 사유의 형식은 바꾸지 않는다(F2 의 범위).
- 기대값 변경: 현재 테스트에는 허용 뒤 재차단이 없는지를 단언하는 가드가 없다. 그 거동은 신규 가드(AC-RG-017)로 추가한다. 기존 `TestReviewGate_CacheHitReusesFailVerdict` 는 첫 차단만 본다(RG-GUARD).

### M4 — 리뷰 기준의 단일 해석 함수 (우선순위 High, Q1 판정 반영)

- 요구: REQ-RG-005, REQ-RG-006, REQ-RG-007, REQ-RG-008. 인수 기준: AC-RG-006, AC-RG-007, AC-RG-008, AC-RG-009, AC-RG-016, AC-RG-020.
- RED: 턴 종료 범위, card 범위 리뷰, 감사 기준이 같은 merge-base 를 낸다(AC-RG-006). 카드가 main 에서 잘렸으면 main 이 선택된다(AC-RG-007). 카드가 develop 에서 잘렸고 키가 다르면 최근접 develop 이 선택되며 inconclusive 가 아니다(AC-RG-008). 후보가 없으면 inconclusive 이며 다른 기준으로 넘어가지 않는다(AC-RG-009). 동점이고 merge-base SHA 가 다르면 inconclusive 다(AC-RG-020).
- GREEN: `cardMergeBase`, `resolveReviewBase`, `resolveReviewMergeBase` 가 하나의 해석 함수(design.md §4)를 거치도록 수렴한다. 호출 경로를 모두 바꾼다.
- 기대값 변경: `TestCodexAudit_BaseBranchResolutionChain` 의 단계 0 과 단계 1–2 단언은 Q1 결과와 REQ-RG-008 에 따라서만 바뀐다. 바뀌는 항목마다 근거를 progress.md 에 적는다(AC-RG-016).
- 공허 검사 방지: 실행된 테스트 수(`--- PASS` 줄 수)를 기록한다.

### M5 — inconclusive 비재사용과 재검토 시작 (우선순위 Medium, Q4 판정 반영)

- 요구: REQ-RG-009, REQ-RG-010. 인수 기준: AC-RG-010, AC-RG-011.
- RED: inconclusive 영수증 뒤의 턴 종료는 재사용으로 허용만 하지 않고 재검토를 시작하며 메시지를 낸다(AC-RG-010). 다음 턴 종료에서 간격 없이 재검토가 다시 시작되고, 진행 중 표식이 있으면 중복되지 않는다(AC-RG-011).
- GREEN: 재사용 판정을 pass 와 fail 로 제한한다. 간격 없는 재시작과 진행 중 표식 존중을 구현한다.
- 기대값 변경: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 는 결함을 고정하므로 기대값을 바꾼다(셀 B-2).
- 이름과 단언의 어긋남(감사 O10): 테스트 이름은 inconclusive 가 pass 증거가 아님을 주장하고, 단언(`internal/cli/codex_review_cache_test.go:262–269`)은 허용(`out.Decision` 이 비어 있음)을 고정한다. 두 주장은 양립하지만, 이름이 가리키는 재사용 금지는 단언에 없다. 이 차이를 progress.md 에 기록한다.

### M6 — 실행 단위 스위치와 두 층의 읽기 (우선순위 Medium, Q3 판정 반영)

- 요구: REQ-RG-011, REQ-RG-012. 인수 기준: AC-RG-012, AC-RG-013.
- RED: 재정의 파일이 off 이면 같은 run 의 두 레인에서 두 훅 모두 검토 없이 허용한다(AC-RG-012). 재정의 파일이 없으면 커밋된 값을 따르며 작업 트리의 미커밋 값은 무시된다(AC-RG-012). 셸 층과 Go 층이 같은 스위치 결과를 낸다(AC-RG-013).
- GREEN: 스위치 읽기를 하나의 해석 절차로 모은다. 재정의 파일의 주 체크아웃 해석과 커밋된 값의 판독을 두 층이 같이 쓴다. 추적 키 `codex.review_gate.enabled` 는 기본값 원천으로만 남긴다.
- 두 면 원칙(spec.md §D.3): 훅 래퍼의 `.claude/` 면과 `internal/template/templates/` 면, 그리고 Go 읽기 경로를 이 마일스톤의 한 커밋에서 함께 갱신한다. 템플릿 미러 갱신은 M7 에 두지 않는다.

### M7 — 회귀 정비와 문서 참조 정정 (우선순위 Low, 기계적)

- 요구: REQ-RG-013, REQ-RG-014(회귀). 인수 기준: AC-RG-014, AC-RG-015, AC-RG-019(가드 부분).
- 회귀 가드 전체를 변경 전후로 다시 실행하고 비교표를 progress.md 에 남긴다. 여기에는 `TestReviewGate_FailOpenOnMissingCodex` 처럼 셀 B-1 에 없는 reviewer 부재 가드를 포함한다.
- F4 의 규칙 문장(`.claude/rules/moai/workflow/factory-dispatch-detail.md:157`)을 정정한다. 그 파일과 템플릿 미러를 같은 커밋에서 맞춘다.
- F1–F5 를 후속 카드로 등록한다. 이 SPEC 의 Follow-up 항목을 그대로 옮긴다.

## §E 자기 검증 — run 단계 명령 형식

- 기준 코드 확인: `git diff --stat 2aab5f797b75983e132af451da68f69e3426557b HEAD -- ':!.moai/specs/SPEC-REVIEW-GATE-REPAIR-001'` 의 출력이 비어 있어야 한다. 비어 있지 않으면 멈추고 보고한다. 설치된 빌드 버전 문자열로 기준을 판정하지 않는다.
- 테스트 호출 형식: `go test -count=1 -v -run '^<TestName>$' ./internal/cli/` 처럼 단일 호출로, 출력을 원문으로 남긴다. 종료 코드는 별도 필드로 기록한다.
- 빌드 출처: 측정은 트리의 소스에서 만든 빌드로 한다. 설치된 빌드의 버전 문자열만으로 판정하지 않는다(verification-claim-integrity.md §2.2).
- 공허 검사 방지: 실행된 테스트 수(`--- PASS` 줄 수)를 함께 기록한다. 0이면 측정 불가로 보고한다.
- 트리 변화 점검: 각 마일스톤 종료 시 `git status --short` 를 본다. `.moai/config/sections/workflow.yaml` 외의 변화가 있으면 멈추고 보고한다.
- 변경 세트 점검: 마일스톤 커밋마다 `git show --stat <sha>` 로 두 면(`.claude/` 와 `internal/template/templates/`)이 같은 커밋에 있는지 확인한다.

## §F 위험

1. **문언대로의 키 불일치는 이 저장소에서 항상 inconclusive 다.** Q1 = (a) 로 정해졌으므로 이 위험은 M4 의 RED(AC-RG-008)가 막는다.
2. **결함을 고정한 테스트의 기대값 변경을 회귀로 오인하기 쉽다.** 두 테스트 모두 현재 PASS 이며(셀 B-2), 그 PASS 가 결함의 증거다. 기대값 변경은 근거와 함께 기록한다.
3. **템플릿 미러 누락.** 래퍼가 두 면에 있다(§C). 한 면만 바꾸면 배포본과 저장소 본이 갈린다. D12 에 따라 M6 의 한 커밋에 함께 담는다.
4. **스위치 해석의 두 층 불일치.** 셸 층과 Go 층이 주 체크아웃 해석기를 다르게 쓰면 같은 run 에서 두 층이 다른 값을 낸다(AC-RG-013).
5. **커밋된 값 판독은 현행 동작과 다르다.** 현재 Go 층은 작업 트리의 workflow.yaml 을 읽는다. D11 에 따라 기본값은 HEAD 의 커밋된 값이다. 작업 트리의 미커밋 변경은 무시된다. 이 행위 변경은 M6 에 명시되어 있다. 운영자가 해석을 확인할 대상이다.
6. **메시지 범위는 넓다.** REQ-RG-016 은 문언대로 스위치 꺼짐, 검토할 변경 없음, stop-hook 해제 턴 종료에도 메시지를 낸다. 그런 턴 종료가 많으면 메시지가 자주 나온다. 범위는 운영자가 확인할 대상이다(보드 판정 D4 "항상 메시지 표시"의 범위를 그대로 옮긴 것이다: 새 통과 검토 없이 허용되는 모든 턴 종료. 감사 보고서가 든 키 한정 선택지보다 넓다).
7. **재정의 파일의 stale 바인딩.** 단일 활성 run 을 해석하지 못하면 재정의 파일은 무시되고 커밋된 값이 적용된다. 실행 식별자 해석기는 run 단계가 정한다.
8. **설치 빌드 오인.** 설치된 빌드(`4f8aba061`)는 트리보다 새 커밋이다. 설치 빌드의 결과를 트리의 증거로 인용하면 도구 출처 규칙을 어긴다.
9. **리더의 미커밋 workflow.yaml 변경은 이 워크트리에서 재현되지 않았다.** 측정은 HEAD 기준이다. 리더가 같은 파일을 다시 바꾸면 M6 의 판독 경로가 그 값을 무시하는지 AC-RG-012 로 확인한다.
10. **영수증과 로그는 기계 로컬이다.** 레인 저장소의 스냅샷과 상세 파일은 이 저장소의 커밋 대상이 아니다. 인용할 때는 기계 로컬 근거임을 밝힌다.
11. **완료된 결정의 뒤집힘이 SPEC 본문에 없다.** 카드 t1426 의 감사 경로 결정 뒤집힘은 Q1 행과 HISTORY 에만 기록된다. SPEC-CODEX-REVIEW-TARGET-001 의 본문은 편집하지 않는다.

## §G 안티패턴

- 차단 횟수 C = 1 이나 간격 없음(Q2, Q4)을 run 단계에서 임의로 바꾸는 것.
- 차단 횟수를 허용마다 되돌리는 것(D3 위반), 또는 키와 무관한 전역 카운터로 바꾸는 것.
- 검토 없이 허용되는 턴 종료 중 일부 상태만 메시지를 내고 나머지를 생략하는 것(D4 위반).
- 기준 함수를 둘로 유지하고 래퍼만 맞추는 것(REQ-RG-005 위반).
- 작업 트리의 미커밋 workflow.yaml 값을 스위치 기본값으로 읽는 것(D11 위반).
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
