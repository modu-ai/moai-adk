# SPEC-REVIEW-GATE-REPAIR-001 — 인수 기준

## §A 관측 규율 [HARD]

1. **기준 트리.** 모든 셀의 트리 SHA 는 `2aab5f797b75983e132af451da68f69e3426557b`(짧은 형식 `2aab5f797`)다. 측정 시점마다 `git -C <작업 트리> rev-parse HEAD` 로 다시 확인한다.
2. **릴리스 게이트 0건(plan 시점).** 이 SPEC 의 행위 기준을 plan 시점에 릴리스 게이트로 승격할 수 없는 이유는 세 가지다.
   - 행위 요구를 단언하는 테스트가 현재 트리에 없다. 테스트 이름 확인으로 확인했다(research.md §6).
   - 재실행 가능한 RED 셀을 만들려면 구현 전 트리에 실패하는 테스트가 있어야 한다. 이 plan 단계는 SPEC 디렉터리 밖을 쓸 수 없으므로 그런 테스트를 추가할 수 없다.
   - 기존 행위 테스트 가운데 원하는 행위를 단언하는 것은 없다. 현재 PASS 인 행위 테스트는 결함을 고정하고 있다(§C 셀 B-2).
3. **분류.** `회귀 가드(재실행 가능, 현재 GREEN)` 은 RG-GUARD 로 표시한다. `회귀 가드(재실행 불가 또는 행위 미단언)` 은 RG-PENDING 으로 표시한다. 릴리스 게이트는 0건이다. 검토 규칙(verification-completeness.md §2.1)에 따라, 재실행할 수 없는 기준은 릴리스 게이트가 될 수 없다.
4. **셀 네 요소.** 릴리스 게이트로 승격하는 셀은 (1) 명령(단일 호출, 파이프·리다이렉트·`&&` 없음), (2) 명령의 원문 stdout, (3) 종료 코드(별도 필드), (4) 트리 SHA 를 모두 가진다. 이 plan 시점에는 RED 셀을 인용하지 않는다. 셀 B-1 과 B-2 는 회귀 가드 및 기준선 기록이며 RED 셀이 아니다.
5. **go test 의 성격.** `go test` 는 빌드 캐시와 테스트별 임시 디렉터리를 쓴다. 엄격한 의미의 read-only 는 아니다. 실행 전후 `git status --short` 가 같았다(§C 트리 점검). 이 차이를 여기에 명시한다.

## §B AC 매트릭스

요구 연결은 spec.md 의 REQ-RG 번호를 따른다. `RG-PENDING` 은 현재 행위를 단언하지 않거나 재실행할 수 없는 기준, `RG-GUARD` 는 현재 GREEN 인 재실행 가능 회귀 가드를 뜻한다.

### AC-RG-001 — 진입 훅은 fail 영수증이 있어도 차단하지 않는다 (REQ-RG-001)

- **Given** 현재 리뷰 키에 대한 fresh fail 영수증이 있다.
- **When** 진입 훅이 프롬프트를 받는다.
- **Then** 출력에 차단 결정(`decision: block`)이 없다.
- **분류** RG-PENDING. 결함 고정 기준선: `TestReviewEntry_FreshFailBlocks` 가 차단을 단언하며 PASS 한다(셀 B-2). 이 단언은 M2 에서 바뀐다.

### AC-RG-002 — 첫 프롬프트에서 상세가 추가 맥락으로 한 번 전달되고 전달 기록이 생긴다 (REQ-RG-002)

- **Given** fresh fail 영수증이 있고 이 키의 전달 기록이 없다.
- **When** 첫 프롬프트가 들어온다.
- **Then** 추가 맥락(`additionalContext`)에 상세가 한 번 실리고, 키 단위 전달 기록이 생긴다.
- **분류** RG-PENDING.

### AC-RG-003 — 전달 기록이 있는 키의 상세는 다시 실리지 않는다 (REQ-RG-002)

- **Given** 같은 키에 전달 기록이 있다.
- **When** 같은 키로 다음 프롬프트가 들어온다.
- **Then** 추가 맥락에 상세가 다시 실리지 않는다.
- **분류** RG-PENDING.

### AC-RG-004 — 차단 횟수가 C 미만인 fresh fail 턴 종료는 상세를 이유로 차단하고 횟수를 올린다 (REQ-RG-003)

- **Given** fresh fail 영수증이 있고, 이 키의 차단 횟수가 0 이며, stop_hook_active 가 거짓이다.
- **When** 턴이 끝난다.
- **Then** 차단 결정과 상세 사유가 나오고, 이 키의 차단 횟수가 1 이 된다.
- **분류** RG-GUARD(차단 부분: `TestReviewGate_CacheHitReusesFailVerdict`, 셀 B-1 PASS). 횟수 1 증가의 단언은 M3 에서 추가한다.

### AC-RG-005 — 차단 횟수가 C 에 이른 뒤의 턴 종료는 허용하고 미검토를 드러낸다 (REQ-RG-004, REQ-RG-012)

- **Given** 같은 fail 키의 차단 횟수가 C(= 1)이다.
- **When** 같은 키에서 다음 턴 종료가 온다(stop_hook_active 거짓).
- **Then** 차단하지 않고 허용하며, systemMessage 에 차단 횟수 소진이라는 미검토 상태가 나온다.
- **분류** RG-PENDING. 운영자 판정 Q2(C = 1)와 D3(키당 총량)가 기대값을 정한다.

### AC-RG-010 — inconclusive 영수증은 재사용되지 않고 재검토가 시작되며 턴 종료는 메시지와 함께 허용된다 (REQ-RG-005, REQ-RG-006, REQ-RG-012)

- **Given** 현재 키의 inconclusive 영수증이 있고, 진행 중 표식이 없다.
- **When** 턴이 끝난다.
- **Then** 백그라운드 재검토가 시작되고, 턴 종료는 차단 없이 허용되며, systemMessage 에 inconclusive 상태가 나온다. stderr 에 "reusing the cached inconclusive verdict" 가 나오지 않는다.
- **분류** RG-PENDING. 결함 고정 기준선: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 가 PASS 하며 stderr 에 "reusing the cached inconclusive verdict for the unchanged tree (skip #1)" 가 나온다(셀 B-2).

### AC-RG-011 — inconclusive 뒤에는 다음 턴 종료에서 간격 없이 재검토가 다시 시작되고, 살아 있는 진행 중 표식이 있으면 중복되지 않는다 (REQ-RG-006)

- **Given** (i) 현재 키의 inconclusive 영수증이 있다. 직전 턴 종료에서 시작된 검토가 inconclusive 로 완료되었고, 완료와 함께 진행 중 표식이 풀렸다(수동 표식 삭제 없음).
- **When** 다음 턴이 끝난다.
- **Then** 쿨다운 없이 이 턴 종료에서 새 검토가 시작되고, 턴 종료는 허용되며 systemMessage 에 inconclusive 상태가 나온다(운영자 판정 Q4 = (a)).
- **Given** (ii) 현재 키의 inconclusive 영수증이 있고, 검토가 실행 중이다. 진행 중 표식이 남아 있고, 그 표식의 소유자 프로세스가 살아 있다.
- **When** 그 검토가 끝나기 전에 턴이 다시 끝난다.
- **Then** 두 번째 검토는 시작되지 않는다. 턴 종료는 허용되며 systemMessage 에 inconclusive 상태가 나온다. 표식은 동시 실행만 막는다.
- **Given** (iii) 현재 키의 inconclusive 영수증이 있고, 진행 중 표식이 남아 있으나 그 소유자 프로세스가 죽었거나 소유자 pid 를 읽을 수 없다(이전 형식의 표식 포함).
- **When** 다음 턴이 끝난다.
- **Then** 표식은 고아로 회수되고, 이 턴 종료에서 새 검토가 시작되며, 턴 종료는 허용되고 systemMessage 에 inconclusive 상태가 나온다.
- **분류** RG-PENDING.

### AC-RG-012 — 실행 스위치: 재정의 파일이 적용되면 그 값이, 아니면 커밋된 값이 두 훅을 지배한다 (REQ-RG-007)

- **Given** 같은 run 의 두 레인이 있고, 커밋된 `codex.review_gate.enabled` 는 HEAD 에서 `true` 다. 작업 트리의 workflow.yaml 값은 HEAD 와 다를 수 있다(미커밋 변경 허용). 재정의 파일의 형식과 실행 식별자 해석은 design.md §5(재정의 파일 스키마와 단일 활성 run 규칙)를 따른다. 시나리오는 다섯이다. (a) 주 체크아웃의 재정의 파일이 이 run 식별자로 `off` 를 싣는다. (b) 재정의 파일이 없다. (c) 재정의 파일이 `on` 을 싣지만 식별자가 다른 run 의 것이다. (d) 재정의 파일이 이 run 식별자로 `on` 을 싣고, 작업 트리의 값은 `false` 다. (e) 이 run 의 카드 워크트리가 아닌 세션(주 체크아웃이나 다른 트리에서 시작한 세션)이 이 run 식별자를 싣고 값이 `off` 인 재정의 파일을 읽는다.
- **When** 두 레인과 (e) 의 세션에서 진입 훅과 턴 종료가 실행된다.
- **Then** (a) 두 훅 모두 검토를 시작하지 않고 허용한다. 턴 종료 게이트의 systemMessage 에 스위치 꺼짐과 재정의 파일 출처가 나온다. 진입 훅은 이 상태에 systemMessage 를 내지 않는다(spec.md REQ-RG-012, design.md §2 의 스위치 꺼짐 행). (b) 커밋된 `true` 를 따라 검토가 실행되며, 작업 트리의 `false` 는 무시된다. (c) 파일은 무시되고 커밋된 `true` 를 따른다. (d) 재정의 값 `on` 이 우선하며 검토가 실행된다. (e) 그 세션은 재정의 파일을 무시하고 커밋된 `true` 를 따라 검토가 실행된다(파일의 `off` 는 적용되지 않는다).
- **분류** RG-PENDING. 현재 코드는 작업 트리의 파일을 읽으므로 (b)와 (d)의 작업 트리 값에 따라 달라진다(research.md §3 #50, #52). (e) 는 현재 해석기에 membership 검사가 없어 RG-PENDING 이다(design.md §5, `internal/cli/factory_card.go:2313–2322`).

### AC-RG-013 — 셸 층과 Go 층이 같은 스위치 결과를 낸다 (REQ-RG-008)

- **Given** 같은 run 에 대해 재정의 값이 `on`, `off`, 또는 부재인 세 상태 각각. 재정의 파일은 design.md §5 의 스키마(`run_id`, `codex_review_gate`)를 따르고, 실행 식별자는 design.md §5 의 단일 활성 run 규칙으로 얻는다.
- **When** 셸 래퍼와 Go 읽기 경로가 스위치 값을 해석한다.
- **Then** 세 상태 각각에서 두 층의 결과가 같다.
- **분류** RG-PENDING. 부분 가드: `TestReviewGateReaders_AgreeWithConfigLoader`(셀 B-1)는 Go 층과 설정 로더의 일치만 본다. 셸 층을 검사하는 가드는 없다.

### AC-RG-014 — 현재 pass 영수증은 재검토 없이 허용한다 (REQ-RG-009)

- **Given** 현재 키에 대한 fresh pass 영수증이 있다.
- **When** 진입 훅 또는 턴 종료가 실행된다.
- **Then** 새 검토를 시작하지 않고 허용한다.
- **분류** RG-GUARD. 가드: `TestReviewEntry_FreshPassAllows`(셀 B-1, PASS).

### AC-RG-015 — 검토 건너뜀의 허용 결정은 바뀌지 않는다 (REQ-RG-010)

- **Given** 게이트가 꺼져 있거나, stop-hook-active 해제이거나, 검토할 변경이 없거나, reviewer 바이너리가 없다.
- **When** 훅이 실행된다.
- **Then** 허용 결정은 현재와 같다.
- **분류** RG-GUARD. 가드: `TestReviewGate_DisabledAllows`, `TestReviewEntry_DisabledAllows`, `TestReviewGate_LoopPreventionAllows`(셀 B-1, 모두 PASS). reviewer 부재의 가드 `TestReviewGate_FailOpenOnMissingCodex` 는 셀 B-1 에 없으므로 M7 에서 실행 대상에 넣는다. 메시지 추가는 AC-RG-019 가 본다.

### AC-RG-017 — 허용 뒤에도 차단 횟수는 되돌아가지 않는다 (REQ-RG-004)

- **Given** 같은 fail 키에서 차단 횟수가 이미 C(= 1)에 도달했고, 그 뒤 턴 종료가 허용되어 메시지가 표시되었다.
- **When** 같은 키에서 턴 종료가 또 온다(stop_hook_active 거짓).
- **Then** 다시 차단하지 않고 허용하며 메시지를 표시한다. 차단 횟수는 여전히 C 다.
- **분류** RG-PENDING. 현재 테스트에 이 단언이 없다(신규 가드).

### AC-RG-018 — 턴 종료에 fresh 영수증이 없으면 재검토를 시작하고 허용한다 (REQ-RG-011, REQ-RG-012)

- **Given** 현재 키의 영수증이 없거나(absent), 유효기간이 지났거나(stale), 결합 필드가 다르다(key-mismatch). 진행 중 표식이 없다.
- **When** 턴이 끝난다.
- **Then** 백그라운드 재검토가 시작되고, 턴 종료는 차단 없이 허용되며, systemMessage 에 해당 상태 이름이 나온다.
- **분류** RG-PENDING. 기존 킥 가드 `TestReviewGate_StaleKeyKicksBackgroundReview` 와 `TestReviewGate_CacheMissKicksBackgroundReview` 는 존재가 확인되었으나(research.md §6) 셀로 실행하지 않았다. 메시지 단언은 없다.

### AC-RG-019 — 검토 없이 허용되는 턴 종료는 상태를 명명한 메시지를 낸다 (REQ-RG-012, REQ-RG-010)

- **Given** 네 상황: (a) 스위치 꺼짐, (b) 검토할 변경 없음, (c) stop-hook 해제, (d) reviewer 없음(reviewer 바이너리 부재).
- **When** 각 상황에서 턴이 끝난다.
- **Then** 네 경우 모두 허용 결정이 현재와 같고, 각 경우의 systemMessage 는 그 경우의 상태 이름을 담는다: 스위치 꺼짐, 검토할 변경 없음, stop-hook 해제, reviewer 없음.
- **분류** RG-PENDING. 가드 기준선: `TestReviewGate_DisabledAllows` 등은 결정만 단언한다(셀 B-1).

## §C 기준선 셀 (실행 기록)

### 셀 B-1 — 회귀 가드 (현재 GREEN)

- **명령(단일 호출):** `go test -count=1 -v -run '^(TestReviewEntry_FreshPassAllows|TestReviewGate_DisabledAllows|TestReviewEntry_DisabledAllows|TestReviewGate_LoopPreventionAllows|TestReviewGate_CacheHitReusesFailVerdict|TestReviewGateReaders_AgreeWithConfigLoader|TestCardMergeBaseConfiguredMain|TestCodexAudit_BaseBranchResolutionChain)$' ./internal/cli/`
- **트리 SHA:** `2aab5f797b75983e132af451da68f69e3426557b`
- **종료 코드:** `0` — 기록 당시 실행 도구가 비정상 종료를 보고하지 않았다는 사실에서 읽었다. 명령 뒤에 `echo EXIT=$?` 를 붙여 따로 캡처하지 않았다. 아래 원문 블록에 EXIT 줄이 없는 이유다. 별도 캡처 기록은 없다(§F 잔여 위험).
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
- **트리 SHA:** `2aab5f797b75983e132af451da68f69e3426557b`
- **종료 코드:** `0` — 셀 B-1 과 같은 방식으로 읽었다. 원문 블록에 EXIT 줄은 없고 별도 캡처는 없다.
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
- **결함 고정 목록 확장(v0.2.6):** 위 두 테스트 외에 기대값이 바뀌는 테스트는 plan.md §D 의 M2·M5·M6 기대값 변경 항목에 적혀 있다: `codex_review_cache_test.go:216–227`, `codex_review_delay_test.go:195–197`·`:267–269`·`:272–292`, `codex_review_gate_live_test.go:129–133`, `internal/hook/review_gate_selfgate_test.go:92–113`, `internal/cli/mcp_codex_test.go:522–552`, `internal/cli/mcp_codex_consoletest_test.go:46–47`·`:76–77`. 이 확장 목록의 원문 기준선은 이 판에서 실행하지 않았다(Gap). run 단계가 착수 전에 같은 형식으로 측정해 progress.md §E.2 에 남긴다.

### 트리 점검

- 셀 B-1 과 B-2 실행 시점의 `git -C <작업 트리> status --short` 출력: ` M .moai/config/sections/workflow.yaml`(앞선 측정 세션이 기록한 리더의 미커밋 변경 한 줄). 이 plan 단계의 같은 명령 출력은 비어 있다(작업 트리 = HEAD). 리더의 변경은 이 워크트리에서 재현되지 않았다.
- `git -C <작업 트리> rev-parse --short HEAD`: plan 단계에서 `6218869fa`(plan 커밋). 기준 코드 트리는 `2aab5f797b75983e132af451da68f69e3426557b` 이며, 두 커밋 사이의 코드 차이는 §D 의 기준 코드 확인 명령으로 본다(plan.md §E).

## §D RED 확립 규율 (run 단계)

- 이 SPEC 은 릴리스 게이트를 승격하지 않는다(plan 시점 릴리스 게이트 0건). 승격 절차는 이 SPEC 의 범위 밖이다.
- 셀 네 요소의 규칙은 승격이 일어나는 SPEC 에서 적용한다.
- 결함을 고정한 기존 테스트(셀 B-2)의 기대값을 바꾸는 것은 승격과 별개다. 바꾼 기록은 progress.md 에 남긴다.

## §E 완료 정의 (DoD)

- Q2–Q6 의 운영자 판정과 보드 판정 D3·D4·D11 이 decision-index.md 에 기록되어 있다. Q1 은 t1662 로 이관되었다(d-20261010T070728Z-6cdf). 2회차 기록(d-20261010T063222Z-43e0)도 같은 색인에 있다.
- AC-RG-001–005, AC-RG-010–015, AC-RG-017–019 각 항목이 "가드 GREEN" 이거나 "마일스톤 커밋의 새 테스트는 첫 부모에서 EXPECTED_RED 로 실패하고 마일스톤 커밋과 최종 트리에서 GREEN 이다(검사 기록은 progress.md §E.2)" 상태다. 덮어쓰기 단위는 마일스톤 커밋이 추가하거나 바꾼 _test.go 파일의 그 커밋 시점 내용이다.
- 결함을 고정한 두 테스트의 기대값 변경이 근거(REQ 번호 또는 보드 판정)와 함께 progress.md 에 기록되어 있다.
- 템플릿 미러가 맞고(M6 의 한 커밋), F1–F5 가 후속 카드로 등록되고 F6 이 t1635 로 이관되어 있다.
- 릴리스 게이트 승격 사례는 이 SPEC 에서 인용하지 않는다.

## §F 잔여 위험

- 셀 B-1 과 B-2 는 이 머신(darwin/arm64, go1.26.8)에서 실행한 결과다. 다른 플랫폼의 출력은 측정하지 않았다.
- 셀의 종료 코드는 별도 캡처가 없다. 기록 당시 도구가 비정상 종료를 보고하지 않았다는 사실에서 읽었다. run 단계의 셀은 종료 코드를 별도 필드로 캡처한다.
- 테스트 실행 순서가 출력 순서를 바꿀 수 있다. 판정은 테스트 이름과 PASS 개수로 한다.
- AC-RG-004 가드(`TestReviewGate_CacheHitReusesFailVerdict`)와 셀 B-2 의 `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 는 단언의 끝부분을 이 문서 작성 중 다시 읽지 않았다(research.md §8 Gap 7). run 단계가 읽는다.
- 셀 B-2 의 두 번째 테스트는 stderr 로 inconclusive 를 허용한다는 사실을 보여 주지만, 그 판단이 옳은지는 판정하지 않는다.
- AC-RG-019 와 AC-RG-005 의 메시지 범위(키가 없는 상태 포함)는 운영자 확인 대상이다(plan.md §F 6).
