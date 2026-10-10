# SPEC-REVIEW-GATE-REPAIR-001 — 구현 계획

## §A 맥락

- **Tier L 판정.** 변경 표면이 네 서브시스템(진입 훅, 턴 종료 게이트, 영수증 저장과 재사용, 리뷰 기준 해석)과 설정 스위치에 걸친다. 인용하는 코드 파일이 15개를 넘고, 요구사항 14개와 인수 기준 16개를 가진다(상한 25). 리서치 산출물(research.md)이 필요하므로 설계 산출물(design.md)과 함께 다섯 개 산출물을 둔다. 여기에 progress.md 와 decision-index.md 를 더한다. decision-index.md 는 `interview.decision_gate: on` 이므로 필수다.
- **현재 상태.** Plan 단계의 산출은 개방된 운영자 판정 여섯 건(Q1–Q6), 인용 전수 재검증, 기준선 셀 두 개 실행이다. 구현 코드는 건드리지 않았다.
- **이 계획이 하지 않는 것.** 판정 없이 구현하지 않는다. 판정이 나온 뒤에도 F1–F3 는 이 계획에 들어가지 않는다.

## §B 되돌리기 어려운 결정 (구현 전 확정)

가장 바뀌기 쉬운 결정부터 순서대로 적는다. 각 항목은 decision-index.md 의 같은 번호와 짝을 이룬다.

1. **Q1 — 두 설정 키가 다른 ref 를 가리킬 때의 기준 규칙.** [NEEDS CLARIFICATION: 키가 다르면 inconclusive 로 둘지, 후보 집합의 최근접 선택으로 진행할지, 설정을 먼저 정렬할지]. 문언대로면 이 저장소에서 모든 판정이 inconclusive 가 된다.
2. **Q6 — SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 기록 처분.** [NEEDS CLARIFICATION: 진입 차단 요구의 폐기를 해당 SPEC 개정 행으로 기록할지, 이 SPEC 에만 기록할지, 대체 선언으로 할지].
3. **Q3 — 실행 단위 스위치의 형태.** [NEEDS CLARIFICATION: 환경 값으로 할지, 주 체크아웃의 미추적 재정의로 할지, 둘을 우선순위로 함께 둘지].
4. **Q2 — 연속 차단 상한 C.** [NEEDS CLARIFICATION: C 의 값]. 현행은 턴 사이 상한이 없다. 카드는 상한을 요구한다.
5. **Q4 — inconclusive 뒤 재검토 간격.** [NEEDS CLARIFICATION: 턴 종료마다 즉시 재시작할지, 고정 쿨다운 뒤 재시작할지].
6. **Q5 — 두 계층 동시 활성 시 doctor 경고.** [NEEDS CLARIFICATION: 이 SPEC 범위에 넣을지, 후속으로 보낼지].

## §C 기술 접근 — 파일 지도

현재 역할은 research.md 의 인용 등록부에서 확인한 것이다. 줄 번호는 기준 트리(2aab5f797)의 값이다.

| 파일 | 현재 역할 (근거) | 관련 마일스톤 |
|---|---|---|
| `internal/cli/codex_review_delay.go` | 진입 훅 `HandleCodexReviewEntry`(140행). 차단 분기 169–183행. 헤더 1–14행은 교착을 설계로 채택했다 | M2 |
| `internal/cli/codex_review_gate.go` | 턴 종료 `HandleCodexReviewGate`(104행). stop_hook_active 109행. 수용 영수증 재사용 159–178행. 진행 중 표식 197–210행. 스위치 읽기 469행 | M3, M5, M6 |
| `internal/cli/codex_review_cache.go` | 유효기간 상수 29행. 상세 파일 읽기 72–76행 | M2, M5 |
| `internal/cli/codex_review_receipt.go` | 영수증 생산자. 키 측정 149행. inconclusive 기록 160–186행. 런타임 설정 전용 발견 사항의 재분류 176행 | M5 |
| `internal/cli/codex_review_scope.go` | `cardMergeBase` 189–212행(`IntegrationTarget` 읽기 191행). origin 처리 204–210행. 변경 집합 해시 341–397행 | M4 |
| `internal/cli/mcp_review_material.go` | `resolveReviewBase` 156–186행. step 0 의 `worktree_base_branch` 읽기 159행. `resolveReviewMergeBase` 189행은 얇은 래퍼 | M4 |
| `internal/cli/mcp_selfreview.go` | 리뷰 범위 해석 호출 119행. 실행 시간 접두 제외 목록 168행 | M4 |
| `internal/cli/mcp_codex.go` | `readCodexReviewGateEnabled` 2503행(빈 루트면 false). 감사 경로의 기준 기록 | M6 |
| `internal/cli/codex_review_tree_scope.go` | 트리와 주 체크아웃 범위 판정 133–137행 | M6 |
| `internal/config/loader_integration_branch.go` | `IntegrationTarget` 123–125행. 값 정책은 Q1 이 정한다. 이 파일은 읽기만 한다 | M4 |
| `.claude/hooks/moai/handle-codex-review-entry.sh` | 33–34행: `CLAUDE_PROJECT_DIR` 기준 workflow.yaml 을 읽는다 | M6 |
| `.claude/hooks/moai/handle-codex-review-gate.sh` | 마지막 줄: 훅 호출에 stderr 리다이렉트가 없다 | M6 |
| `.claude/settings.json` 272행, `internal/template/templates/.claude/settings.json.tmpl` 272행 | 진입 훅 timeout 5 (F2 의 대상, 이 계획은 바꾸지 않는다) | — |
| `internal/template/templates/.claude/hooks/moai/handle-codex-review-*.sh` | 두 래퍼의 배포 미러 | M6, M7 |
| `internal/verify/receipt.go` 63–96행, 70–76행, 90행 | 재사용 판정 입력. 결합 필드 다섯 개. 유효기간 0 은 기본값(10분)으로 바뀐다 | M2, M5 |
| `internal/auditreceipt/storeroot.go` 179–185행, 243행 | 주 체크아웃 해석. 트리에 workflow.yaml 이 있으면 그 트리를 그대로 쓴다 | M6 |

파일 목록은 run 단계에서 확정한다. 목록이 Tier L 의 상한을 넘으면 Tier 판정을 다시 한다.

## §D 마일스톤 (우선순위 순)

각 마일스톤은 RED 를 먼저 기록하고 GREEN 으로 넘어간다. RED 는 구현 전 트리에서 실행한 출력을 원문으로 남긴다(tdd-result-contract.md). 마일스톤의 순서는 되돌리기 어려운 결정을 먼저 두고, 기계적인 정리는 뒤에 둔다.

### M1 — 결정 확정과 기록 (우선순위 High, 차단 게이트)

- Q1–Q6 판정을 decision-index.md 의 Operator verdict 에 기록한다.
- Q6 처분이 개정을 요구하면 해당 SPEC 의 기록 작업은 이 SPEC 의 산출물 밖에서 따로 연다.
- 완료 조건: 여섯 행의 Operator verdict 가 모두 채워져 있다.

### M2 — 진입 훅 비차단과 전달 표식 (우선순위 High, 사용자 흐름 변경과 새 데이터 모델)

- RED: fail 영수증이 있어도 진입 훅은 차단하지 않는다(AC-RG-001). 같은 키의 첫 프롬프트에서 상세가 한 번 전달되고 전달 기록이 남는다(AC-RG-002). 전달 기록이 있는 키에서는 반복하지 않는다(AC-RG-003).
- GREEN: 진입 경로의 차단 분기를 추가 맥락 전달로 바꾸고, 키 단위 전달 기록을 둔다. 기록은 영수증 결합 필드 다섯 개의 조합을 키로 쓴다.
- 기대값 변경: `TestReviewEntry_FreshFailBlocks` 는 결함을 고정하고 있으므로(acceptance.md §C, 셀 B-2) 같은 변경에서 기대값을 바꾼다. 바꾼 내용은 progress.md 에 "이전 기대값 → 새 기대값 → 근거 결정 번호" 형식으로 남긴다.

### M3 — 턴 종료 차단 상한 (우선순위 High, 상태 카운터)

- RED: C 번째 연속 차단 뒤 턴 종료가 허용되고, 미검토 상태가 systemMessage 로 드러난다(AC-RG-005).
- GREEN: 키 단위 연속 차단 카운터를 둔다. 허용, pass 기록, 키 변경에서 0 으로 되돌린다. C 의 값은 Q2 를 따른다.
- 차단 사유의 형식은 바꾸지 않는다(F2 의 범위).

### M4 — 리뷰 기준의 단일 해석 함수 (우선순위 High, Q1 에 의존)

- RED: 턴 종료 범위, card 범위 리뷰, 감사 기준이 같은 merge-base 를 낸다(AC-RG-006). 최근접 후보가 선택된다(AC-RG-007). 키가 다른 ref 를 가리키면 Q1 의 규칙대로 판정된다(AC-RG-008). 후보가 없으면 inconclusive 이며 다른 기준으로 넘어가지 않는다(AC-RG-009).
- GREEN: `cardMergeBase`, `resolveReviewBase`, `resolveReviewMergeBase` 가 하나의 해석 함수를 거치도록 수렴한다. 호출 경로를 모두 바꾼다.
- 기대값 변경: `TestCodexAudit_BaseBranchResolutionChain` 의 단계 0 단언은 Q1 의 결과에 따라서만 바뀐다. 바꾸는 경우 Q1 기록을 근거로 한다.

### M5 — inconclusive 비재사용과 재검토 간격 (우선순위 Medium, Q4 에 의존)

- RED: inconclusive 영수증 뒤의 턴 종료는 허용만 하지 않고 재검토를 시작한다(AC-RG-010). 재검토 간격은 Q4 의 규칙을 따른다(AC-RG-011).
- GREEN: 재사용 판정을 pass 와 fail 로 제한한다. 간격 규칙을 구현한다.
- 기대값 변경: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 는 결함을 고정하므로 기대값을 바꾼다(셀 B-2). 이름의 주장과 단언이 어긋나는 점도 함께 기록한다.

### M6 — 실행 단위 스위치 (우선순위 Medium, Q3 에 의존)

- RED: 스위치가 꺼져 있으면 같은 run 의 레인 두 곳에서 두 훅 모두 검토 없이 허용한다(AC-RG-012). 셸 층과 Go 층이 같은 값을 읽는다(AC-RG-013).
- GREEN: 스위치 읽기를 한 원천으로 모은다. 트랙된 per-tree 키의 역할은 Q3 결과에 따라 제거하거나 기본값 역할로 줄인다.
- 래퍼 두 개와 Go 읽기 경로를 같은 변경 세트에서 갱신한다.

### M7 — 회귀 정비와 미러 동기화 (우선순위 Low, 기계적)

- 회귀 가드 전체를 변경 전후로 다시 실행하고 비교표를 남긴다(AC-RG-004, 014, 015, 016).
- 템플릿 미러(`internal/template/templates/…`)와 문서 참조를 맞춘다.
- F1–F3 를 후속 카드로 등록한다. 이 SPEC 의 Follow-up 항목을 그대로 옮긴다.

## §E 자기 검증 — run 단계 명령 형식

- 기준 트리 확인: `git -C <작업 트리> rev-parse --short HEAD` 가 기준 SHA(`2aab5f797`)와 같은지, run 시작 시 다시 잰다.
- 테스트 호출 형식: `go test -count=1 -v -run '^<TestName>$' ./internal/cli/` 처럼 단일 호출로, 출력을 원문으로 남긴다.
- 빌드 출처: 측정은 트리의 소스에서 만든 빌드로 한다. 설치된 빌드의 버전 문자열만으로 판정하지 않는다(verification-claim-integrity.md §2.2).
- 공허 검사 방지: 실행된 테스트 수(`--- PASS` 줄 수)를 함께 기록한다. 0이면 측정 불가로 보고한다.
- 트리 변화 점검: 각 마일스톤 종료 시 `git status --short` 를 본다. `.moai/config/sections/workflow.yaml` 외의 변화가 있으면 멈추고 보고한다.

## §F 위험

1. **문언대로의 M2 는 이 저장소에서 항상 inconclusive 다**(Q1). M3 와 결합하면 턴마다 재검토가 돌면서 결과가 계속 inconclusive 가 되는 비용 증가가 생긴다. Q1 과 Q4 가 함께 정해지기 전에는 M4 와 M5 를 시작하지 않는다.
2. **결함을 고정한 테스트의 기대값 변경을 회귀로 오인하기 쉽다.** 두 테스트 모두 현재 PASS 이며(셀 B-2), 그 PASS 가 결함의 증거다. 기대값 변경은 근거 결정 번호와 함께 기록한다.
3. **템플릿 미러 누락.** 래퍼와 설정이 두 면에 있다(§C). 한 면만 바꾸면 배포본과 저장소 본이 갈린다.
4. **스위치의 레인 전파.** 환경 값 방식은 레인 터미널마다 설정이 필요할 수 있다(Q3 의 근거). 전파가 빠지면 검토가 조용히 빠진다.
5. **완료된 결정의 뒤집힘이 기록되지 않는다.** Q1 과 Q6 을 기록 없이 진행하면, 완료된 SPEC 의 결정과 코드가 영구히 어긋난다.
6. **설치 빌드 오인.** 설치된 빌드(4f8aba061)는 트리(2aab5f797)보다 새 커밋이다. 설치 빌드의 결과를 트리의 증거로 인용하면 도구 출처 규칙을 어긴다.
7. **리더의 미커밋 workflow.yaml 변경.** 측정은 HEAD 기준이다. 작업 트리의 변경은 run 단계 시작 전에 리더와 확인한다.
8. **영수증과 로그는 기계 로컬이다.** 레인 저장소의 스냅샷과 상세 파일은 이 저장소의 커밋 대상이 아니다. 인용할 때는 기계 로컬 근거임을 밝힌다.

## §G 안티패턴

- 상한 C 나 재검토 간격을 run 단계에서 임의로 고르는 것(Q2, Q4 위반).
- 기준 함수를 둘로 유지하고 래퍼만 맞추는 것(REQ-RG-005 위반).
- 매칭이 0개인 실행을 통과로 세는 것(공허한 통과).
- 트리 SHA 대신 브랜치 이름 `develop` 에 핀하는 것(움직이는 ref).
- 결함 고정 테스트를 삭제해서 빨간불을 피하는 것. 기대값 변경만 허용된다.

## §H 참조

- spec.md · acceptance.md · design.md · research.md · decision-index.md
- `.claude/rules/moai/workflow/factory-dispatch-detail.md:157` (이중 리뷰 문장)
- `.claude/rules/moai/workflow/tdd-result-contract.md` (RED 기록 형식)
- `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 (도구 출처)
