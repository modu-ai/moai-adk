# SPEC-REVIEW-GATE-REPAIR-001 — 인수 기준

## §A 관측 규율 [HARD]

1. **기준 트리.** 모든 셀의 트리 SHA 는 `2aab5f797`(작업 트리 HEAD)다. 측정 시점마다 `git -C <작업 트리> rev-parse --short HEAD` 로 다시 확인한다.
2. **릴리스 게이트 0건(plan 시점).** 이 SPEC 의 행위 기준을 plan 시점에 릴리스 게이트로 승격할 수 없는 이유는 세 가지다.
   - 행위 요구를 단언하는 테스트가 현재 트리에 없다. 테스트 이름 확인으로 확인했다(research.md §5).
   - 재실행 가능한 RED 셀을 만들려면 구현 전 트리에 실패하는 테스트가 있어야 한다. 이 plan 단계는 SPEC 디렉터리 밖을 쓸 수 없으므로 그런 테스트를 추가할 수 없다.
   - 기존 행위 테스트 가운데 원하는 행위를 단언하는 것은 없다. 현재 PASS 인 행위 테스트는 결함을 고정하고 있다(§C 셀 B-2).
3. **분류.** `회귀 가드(재실행 가능, 현재 GREEN)` 은 §C 셀 B-1 로 기록한다. `회귀 가드(재실행 불가 — 현 트리에 테스트 없음)` 은 아래 표에서 `RG-PENDING` 으로 표시한다. 릴리스 게이트는 0건이다. 검토 규칙(verification-completeness.md §2.1)에 따라, 재실행할 수 없는 기준은 릴리스 게이트가 될 수 없다.
4. **셀 네 요소.** 릴리스 게이트로 승격하는 셀은 (1) 명령(단일 호출, 파이프·리다이렉트·`&&` 없음), (2) 명령의 원문 stdout, (3) 종료 코드(별도 필드), (4) 트리 SHA 를 모두 가진다. 이 plan 시점에는 RED 셀을 인용하지 않는다. 셀 B-1 과 B-2 는 회귀 가드 및 기준선 기록이며 RED 셀이 아니다.
5. **go test 의 성격.** `go test` 는 빌드 캐시와 테스트별 임시 디렉터리를 쓴다. 엄격한 의미의 read-only 는 아니다. 실행 전후 `git status --short` 가 같았다(§C 트리 점검). 이 차이를 여기에 명시한다.

## §B AC 매트릭스

요구 연결은 spec.md 의 REQ-RG 번호를 따른다. `RG-PENDING` 은 재실행 불가 회귀 가드, `RG-GUARD` 는 현재 GREEN 인 재실행 가능 회귀 가드를 뜻한다.

### AC-RG-001 — 진입 훅은 fail 영수증이 있어도 차단하지 않는다 (REQ-RG-001)

- **Given** 현재 리뷰 키에 대한 fresh fail 영수증이 있다.
- **When** 진입 훅이 프롬프트를 받는다.
- **Then** 출력에 차단 결정이 없다.
- **분류** RG-PENDING. 결함 고정 기준선: `TestReviewEntry_FreshFailBlocks` 가 차단을 단언하며 PASS 한다(셀 B-2).

### AC-RG-002 — 발견 사항은 키마다 한 번 추가 맥락으로 전달되고 전달 기록이 남는다 (REQ-RG-002)

- **Given** fresh fail 영수증이 있고 이 키의 전달 기록이 없다.
- **When** 첫 프롬프트가 들어온다.
- **Then** 추가 맥락에 상세가 한 번 실리고, 키 단위 전달 기록이 생긴다.
- **분류** RG-PENDING.

### AC-RG-003 — 전달 기록이 있는 키의 상세는 다시 실리지 않는다 (REQ-RG-002)

- **Given** 같은 키에 전달 기록이 있다.
- **When** 같은 키로 다음 프롬프트가 들어온다.
- **Then** 상세가 다시 실리지 않는다.
- **분류** RG-PENDING.

### AC-RG-004 — 턴 종료는 fresh fail 에서 상세를 이유로 차단한다 (REQ-RG-003)

- **Given** fresh fail 영수증이 있고 stop-hook-active 플래그가 거짓이다.
- **When** 턴이 끝난다.
- **Then** 차단과 상세가 나온다.
- **분류** RG-GUARD. 가드: `TestReviewGate_CacheHitReusesFailVerdict`(현재 GREEN). 단언의 끝부분은 이 문서에서 다시 읽지 않았다(research.md §6, Gap).

### AC-RG-005 — C 번째 연속 차단 뒤에는 허용하고 미검토를 드러낸다 (REQ-RG-004)

- **Given** 같은 키에서 턴 종료가 C 회 연속 차단되었다.
- **When** C+1 번째 턴 종료가 온다.
- **Then** 턴 종료를 허용하고, systemMessage 에 미검토 상태가 나온다. C 의 값은 Q2 를 따른다.
- **분류** RG-PENDING. Q2 가 기록되기 전에는 값이 정해지지 않는다.

### AC-RG-006 — 세 소비자가 같은 기준 merge-base 를 낸다 (REQ-RG-005)

- **Given** 기준 키 두 개가 같은 ref 를 가리키고 카드가 main 에서 잘린 픽스처가 있다.
- **When** 턴 종료 범위, card 범위 리뷰, 감사 기준을 각각 계산한다.
- **Then** 세 결과의 merge-base SHA 가 같다.
- **분류** RG-PENDING. 부분 가드: `TestCardMergeBaseConfiguredMain` 과 `TestCodexAudit_BaseBranchResolutionChain`(셀 B-1). 두 가드는 소비자 사이의 일치를 검사하지 않는다.

### AC-RG-007 — 최근접 공통 조상을 고른다 (REQ-RG-006)

- **Given** 카드가 main 에서 잘렸고, develop 이 그 뒤로 전진한 픽스처가 있다.
- **When** 기준을 계산한다.
- **Then** 선택된 기준의 `rev-list --count <기준>..HEAD` 가 두 후보 가운데 최소다.
- **분류** RG-PENDING.

### AC-RG-008 — 키가 다른 ref 를 가리키면 inconclusive 이고 두 이름을 댄다 (REQ-RG-007, Q1 에 따라 조정)

- **Given** `worktree_base_branch` 와 `develop_branch` 가 서로 다른, 존재하는 ref 를 가리킨다.
- **When** 판정한다.
- **Then** inconclusive 가 나오고 두 이름이 함께 나온다.
- **분류** RG-PENDING. Q1 이 이 문언을 버리면 이 기준은 폐기되고 Q1 기록을 근거로 대체 기준으로 바뀐다.

### AC-RG-009 — 기준 후보가 없으면 inconclusive 이고 다른 기준으로 넘어가지 않는다 (REQ-RG-008)

- **Given** 어떤 후보 ref 도 트리에서 해석되지 않는다.
- **When** 판정한다.
- **Then** inconclusive 가 나오고 다른 기준이 선택되지 않는다.
- **분류** RG-PENDING. 관측 근거: 현재 `resolveReviewBase` 는 "no base branch resolves" 오류를 낸다(research.md §3, 부분 일치).

### AC-RG-010 — inconclusive 영수증은 재사용되지 않는다 (REQ-RG-009, REQ-RG-010)

- **Given** 같은 키의 inconclusive 영수증이 있다.
- **When** 턴이 끝난다.
- **Then** 재검토가 시작되고, "reusing the cached inconclusive verdict" 로 허용되지 않는다.
- **분류** RG-PENDING. 결함 고정 기준선: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 가 PASS 하며 stderr 에 "reusing the cached inconclusive verdict for the unchanged tree (skip #1)" 가 나온다(셀 B-2).

### AC-RG-011 — 재검토 간격은 Q4 의 규칙대로 동작한다 (REQ-RG-010)

- **Given** inconclusive 뒤 짧은 간격 안에서 턴이 끝난다.
- **When** 턴 종료가 온다.
- **Then** Q4 가 정한 대로 재검토 여부가 결정된다.
- **분류** RG-PENDING. Q4 가 기록되기 전에는 판정할 수 없다.

### AC-RG-012 — 스위치가 꺼져 있으면 같은 run 의 두 훅이 모두 검토 없이 허용한다 (REQ-RG-011, 형태는 Q3)

- **Given** run 스위치가 꺼져 있다. 형태는 Q3 가 정한다.
- **When** 같은 run 의 두 레인에서 진입 훅과 턴 종료가 실행된다.
- **Then** 모두 검토를 시작하지 않고 허용하며, 추적 파일 workflow.yaml 의 값과 무관하다.
- **분류** RG-PENDING.

### AC-RG-013 — 셸 층과 Go 층이 같은 스위치 값을 읽는다 (REQ-RG-012)

- **Given** 같은 run 의 스위치 값이 정해져 있다.
- **When** 셸 래퍼와 Go 읽기 경로가 값을 읽는다.
- **Then** 두 결과가 같다.
- **분류** RG-PENDING(부분 가드 포함). 부분 가드: `TestReviewGateReaders_AgreeWithConfigLoader`(현재 GREEN)는 Go 층과 설정 로더의 일치만 본다. 셸 층을 검사하는 가드는 없다.

### AC-RG-014 — 현재 pass 영수증은 재검토 없이 허용한다 (REQ-RG-013)

- **Given** 현재 키에 대한 fresh pass 영수증이 있다.
- **When** 진입 훅 또는 턴 종료가 실행된다.
- **Then** 재검토 없이 허용한다.
- **분류** RG-GUARD. 가드: `TestReviewEntry_FreshPassAllows`(셀 B-1, PASS, 출력 "background verdict pass for the unchanged tree — allowing").

### AC-RG-015 — 비검토 건너뜀의 결과는 바뀌지 않는다 (REQ-RG-014)

- **Given** 게이트가 꺼져 있거나, stop-hook-active 해제이거나, 검토할 변경이 없거나, reviewer 바이너리가 없다.
- **When** 훅이 실행된다.
- **Then** 현재와 같이 허용한다.
- **분류** RG-GUARD. 가드: `TestReviewGate_DisabledAllows`, `TestReviewEntry_DisabledAllows`, `TestReviewGate_LoopPreventionAllows`(셀 B-1, 모두 PASS).

### AC-RG-016 — 기존 기준 결정의 보존은 Q1 에 따른 명시적 변경으로만 바뀐다 (REQ-RG-005, REQ-RG-006)

- **Given** 감사 경로의 단계 0 과 main 고정 card 기준이 현재 기대값으로 고정되어 있다.
- **When** 회귀 가드를 실행한다.
- **Then** 모두 PASS 한다. 기대값을 바꾸는 경우 Q1 기록을 근거로 명시한다.
- **분류** RG-GUARD. 가드: `TestCodexAudit_BaseBranchResolutionChain`(하위 테스트 4개), `TestCardMergeBaseConfiguredMain`(하위 테스트 2개)(셀 B-1).

## §C 기준선 셀 (실행 기록)

### 셀 B-1 — 회귀 가드 (현재 GREEN)

- **명령(단일 호출):** `go test -count=1 -v -run '^(TestReviewEntry_FreshPassAllows|TestReviewGate_DisabledAllows|TestReviewEntry_DisabledAllows|TestReviewGate_LoopPreventionAllows|TestReviewGate_CacheHitReusesFailVerdict|TestReviewGateReaders_AgreeWithConfigLoader|TestCardMergeBaseConfiguredMain|TestCodexAudit_BaseBranchResolutionChain)$' ./internal/cli/`
- **트리 SHA:** `2aab5f797`
- **종료 코드:** `0`(같은 명령을 실행하면서 `echo EXIT=$?` 를 뒤에 붙여 확인. 명령 자체에는 파이프나 리다이렉트가 없다.)
- **스윕 크기:** 최상위 테스트 8개가 실행되었고, 하위 테스트 12개가 실행되었다(`--- PASS` 줄 20개). 빈 스윕이 아니다.
- **원문 stdout:**

```
=== RUN   TestReviewGate_CacheHitReusesFailVerdict
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- PASS: TestReviewGate_CacheHitReusesFailVerdict (0.75s)
=== RUN   TestReviewEntry_FreshPassAllows
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
codex review entry: background verdict pass for the unchanged tree — allowing
--- PASS: TestReviewEntry_FreshPassAllows (0.75s)
=== RUN   TestReviewEntry_DisabledAllows
--- PASS: TestReviewEntry_DisabledAllows (0.00s)
=== RUN   TestReviewGate_DisabledAllows
--- PASS: TestReviewGate_DisabledAllows (0.00s)
=== RUN   TestReviewGate_LoopPreventionAllows
--- PASS: TestReviewGate_LoopPreventionAllows (0.00s)
=== RUN   TestCardMergeBaseConfiguredMain
=== RUN   TestCardMergeBaseConfiguredMain/remoteAhead=true
=== RUN   TestCardMergeBaseConfiguredMain/remoteAhead=false
--- PASS: TestCardMergeBaseConfiguredMain (1.75s)
    --- PASS: TestCardMergeBaseConfiguredMain/remoteAhead=true (0.82s)
    --- PASS: TestCardMergeBaseConfiguredMain/remoteAhead=false (0.93s)
=== RUN   TestCodexAudit_BaseBranchResolutionChain
=== RUN   TestCodexAudit_BaseBranchResolutionChain/step1_remote_default_head
=== RUN   TestCodexAudit_BaseBranchResolutionChain/step2_main_when_remote_head_absent
=== RUN   TestCodexAudit_BaseBranchResolutionChain/worktree_base_branch_is_step_zero
=== RUN   TestCodexAudit_BaseBranchResolutionChain/dangling_remote_head_falls_through
--- PASS: TestCodexAudit_BaseBranchResolutionChain (4.64s)
    --- PASS: TestCodexAudit_BaseBranchResolutionChain/step1_remote_default_head (1.06s)
    --- PASS: TestCodexAudit_BaseBranchResolutionChain/step2_main_when_remote_head_absent (1.29s)
    --- PASS: TestCodexAudit_BaseBranchResolutionChain/worktree_base_branch_is_step_zero (1.12s)
    --- PASS: TestCodexAudit_BaseBranchResolutionChain/dangling_remote_head_falls_through (1.17s)
=== RUN   TestReviewGateReaders_AgreeWithConfigLoader
=== RUN   TestReviewGateReaders_AgreeWithConfigLoader/both_on
=== RUN   TestReviewGateReaders_AgreeWithConfigLoader/both_off
=== RUN   TestReviewGateReaders_AgreeWithConfigLoader/codex_only
=== RUN   TestReviewGateReaders_AgreeWithConfigLoader/multi_only
=== RUN   TestReviewGateReaders_AgreeWithConfigLoader/absent
=== RUN   TestReviewGateReaders_AgreeWithConfigLoader/tree_scope_skip
--- PASS: TestReviewGateReaders_AgreeWithConfigLoader (0.02s)
    --- PASS: TestReviewGateReaders_AgreeWithConfigLoader/both_on (0.00s)
    --- PASS: TestReviewGateReaders_AgreeWithConfigLoader/both_off (0.00s)
    --- PASS: TestReviewGateReaders_AgreeWithConfigLoader/codex_only (0.00s)
    --- PASS: TestReviewGateReaders_AgreeWithConfigLoader/multi_only (0.00s)
    --- PASS: TestReviewGateReaders_AgreeWithConfigLoader/absent (0.00s)
    --- PASS: TestReviewGateReaders_AgreeWithConfigLoader/tree_scope_skip (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	8.591s
```

### 셀 B-2 — 결함 고정 기준선 (현재 GREEN = 결함이 존재하며 고정되어 있음)

- **명령(단일 호출):** `go test -count=1 -v -run '^(TestReviewEntry_FreshFailBlocks|TestReviewGate_InconclusiveReceiptIsNotPassEvidence)$' ./internal/cli/`
- **트리 SHA:** `2aab5f797`
- **종료 코드:** `0`(셀 B-1 과 같은 방식으로 확인)
- **스윕 크기:** 최상위 테스트 2개, `--- PASS` 줄 2개.
- **원문 stdout:**

```
=== RUN   TestReviewGate_InconclusiveReceiptIsNotPassEvidence
codex review: inconclusive: Looks good to me — no blocking findings.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
codex review gate: reusing the cached inconclusive verdict for the unchanged tree (skip #1)
--- PASS: TestReviewGate_InconclusiveReceiptIsNotPassEvidence (0.90s)
=== RUN   TestReviewEntry_FreshFailBlocks
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- PASS: TestReviewEntry_FreshFailBlocks (1.10s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.766s
```

- **해석:** 두 테스트가 PASS 한다는 것은 두 결함이 현재 트리에 존재하며 테스트가 그것을 기대값으로 고정하고 있다는 뜻이다. inconclusive 영수증이 "reusing the cached" 로 허용되고, fail 영수증이 진입 단계에서 차단된다. 이 두 기대값은 M2 와 M5 에서 의도적으로 바뀐다.

### 트리 점검

- 셀 B-1 과 B-2 실행 뒤 `git -C <작업 트리> status --short` 의 출력: ` M .moai/config/sections/workflow.yaml` (리더의 미커밋 변경 한 줄뿐)
- `git -C <작업 트리> rev-parse --short HEAD`: `2aab5f797`

## §D RED 확립 규율 (run 단계)

- 행위 기준을 릴리스 게이트로 승격하려면, run 단계가 구현 전 트리에서 해당 테스트를 추가하고 실행해 EXPECTED_RED 를 원문으로 기록해야 한다(tdd-result-contract.md). 실패 이유가 목적한 단언과 같아야 한다. 빈 매칭(0개 실행)과 TOOL_FAILURE 는 RED 가 아니다.
- 승격할 때 셀 네 요소(명령, 원문, 종료 코드, 트리 SHA)를 채운다. 트리 SHA 는 구현 전 커밋이다.
- 결함을 고정한 기존 테스트(셀 B-2)의 기대값을 바꾸는 것은 승격과 별개다. 바꾼 기록은 progress.md 에 남긴다.

## §E 완료 정의 (DoD)

- Q1–Q6 의 운영자 판정이 decision-index.md 에 기록되어 있다.
- AC-RG-001 부터 AC-RG-016 까지 각 항목이 "가드 GREEN" 이거나 "RED 기록 후 구현 GREEN" 상태다.
- 결함을 고정한 두 테스트의 기대값 변경이 근거 결정 번호와 함께 기록되어 있다.
- 템플릿 미러가 맞고, F1–F3 가 후속 카드로 등록되어 있다.
- 릴리스 게이트 승격 사례는 run 단계의 RED 셀로만 인용한다.

## §F 잔여 위험

- 셀 B-1 과 B-2 는 이 머신(darwin/arm64, go1.26.8)에서 실행한 결과다. 다른 플랫폼의 출력은 측정하지 않았다.
- 테스트 실행 순서가 출력 순서를 바꿀 수 있다. 판정은 테스트 이름과 PASS 개수로 한다.
- AC-RG-004 가드의 단언 끝부분은 이 문서에서 다시 읽지 않았다.
- 셀 B-2 의 두 번째 테스트는 stderr 로 inconclusive 를 허용한다는 사실을 보여 주지만, 그 판단이 옳은지는 판정하지 않는다.
