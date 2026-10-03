# SPEC-CODEX-GATE-SCOPING-001 — acceptance

> AC 전부 `AC-CGSC-NNN` 라벨의 Given-When-Then. 이진 판정형 — 명령과 관측 대상을 이름으로 적는다. 채택은 두 클래스로 나뉜다(§D.1): **R(RED-first)** 5건 — 구현 전 같은 테스트가 적색이었음을 §D.0 관측 장부의 원문으로 증명하며 release-blocking이다. **G(regression-guard)** 4건 — 「변하지 않음」을 단언하는 보존 기준으로 구현 전 트리에서 적색이 정의상 존재하지 않아(verification-completeness §2.1 undecidable 처분) release-blocking 자격이 없고 특성화 관측으로 이행하며, 채택 시 pass로 기록되지 않는다. R 5건의 적색 테스트는 plan 단계에서 이미 저작돼 트리에 있다 — `internal/cli/codex_review_gate_primary_scope_red_test.go`(001·005·007·008), `internal/template/hook_gate_reports_exclude_test.go`(010). M2/M3/M4가 GREEN으로 뒤집는다(이름이 갈리면 E1이 실측명으로 정정 보고).

## §A 검증 계약

- 판정 근거는 출력 결정·호출 여부·구조화 로그 행·수집 집합이다. reviewer verdict 값 단독은 근거가 아니다(spec §C).
- 픽스처: primary 판별은 `t.TempDir()` 실제 git 저장소(+연결 워크트리 — `newCardScopeFixture` 선례), sync 게이트는 `internal/template/hook_gate_*` 선례의 스크립트 실행 픽스처. OTEL 환경변수 금지, 임시 디렉터리는 전부 `t.TempDir()`.
- RED 관측 명령은 단일 호출형(파이프·리다이렉트·`;`·`&&`·서브셸 없음)이고, exit 코드는 별도 필드로 기록한다(§2.1 4요소).

## §B 에지 케이스

1. primary + detached HEAD → 트리 클래스, primary 스킵 **적용**(판별은 브랜치와 무관 — git-dir 동일성 축).
2. 연결 워크트리 안의 연결 워크트리 → git-dir != common-dir → 비primary, 기존 tree 클래스 동작.
3. `primary_scope` 판독 실패(키 부재 — 플랫·주석·다른 블록 포함 — 미지 값, 읽기 오류, YAML 오류) → 기본 skip 유지 — spec §F.2 정책표와 동일 처분(세 산출물이 같은 표를 인용한다). 파싱 성공해 **명시된** `review`만 복원한다(REQ-CGSC-004). 세션 git 프로브 실패는 별개 축이다(REQ-CGSC-006, 생략 미적용).
4. `.moai/reports`가 디렉터가 아닌 파일명으로 존재 → pathspec exclude 동작 확인(등가 처리).
5. 트리 세션의 변경이 검사 가능 변경 + 설정 표면 드리프트 **혼합** → REQ-CGSC-008 재분류 팔 (자기게이트가 아니라 리뷰 후 판정에서 처리).
6. sync 게이트 초기 커밋(HEAD~1 부재) 상태에서의 `.moai/reports` 픽스처 → 빈 트리 diff 경로와 무관하게 배제.

## §D.0 RED-now 관측 장부 (plan 단계, evidence ledger)

> verification-completeness §2.1의 4요소(단일 호출 명령·원문 stdout·exit 코드·트리 SHA)를 셀이 어긋나지 않게 담는 장부 — R AC의 셀은 여기 번호로 인용한다. 관측 트리: `2de0a2cb613b04765a1554f86685a3b48e0be806`, 브랜치 `WT-codex-gate-scope`, 관측 일자 2026-10-03.

### RED-CGSC-001 — AC-CGSC-001 (관측 형태: authored-test)

- 명령(단일 호출): `go test -count=1 -v ./internal/cli/ -run '^TestCodexReviewGatePrimaryCheckoutSkip$'`
- exit 코드: 1
- 적색 이유: primary 판별·스킵이 존재하지 않아 게이트가 primary 체크아웃의 트리 세션을 자기게이트→리뷰까지 밟고 BLOCK까지 낸다 — 측정된 피해(t1395 처분 #1-#4)의 재현. `lookups=1 detects=1 reviewed=true`가 리뷰 도달을, `0 tree_scope rows`가 별개 basis 행의 부재를 말한다.
- (a) WHEN: 게이트 enabled + `tree_scope` 명시 review + primary 체크아웃(git-dir == common-dir)에 앉은 비카드 세션의 턴 끝 — 스킵이 자기게이트보다 앞서 결정돼야 의미 있는 자리다.
- (b) 적색 입력: 가검토 가능한 미커밋을 품은 primary 레포 픽스처(`newCardScopeFixture`의 primary)로 `HandleCodexReviewGate` 호출 — 올바른 구현은 스킵 행 1건과 무호출을 낸다.
- (c) 도달성: `go test` exit 1 — M2 착지 전 패키지 스위트의 매 실행에서 적색으로 보인다.
- 원문 stdout:

```text
=== RUN   TestCodexReviewGatePrimaryCheckoutSkip
    codex_review_gate_primary_scope_red_test.go:60: a primary-checkout tree session must ALLOW without a review, got &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate: - [P1] ownership probe finding HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
    codex_review_gate_primary_scope_red_test.go:63: the primary skip must precede the self-gate, the lookup and the review; got lookups=1 detects=1 reviewed=true
    codex_review_gate_primary_scope_red_test.go:67: exactly one primary-skip row expected (basis distinguishable from tree_scope); got 0 tree_scope rows
--- FAIL: TestCodexReviewGatePrimaryCheckoutSkip (1.25s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.391s
FAIL
```

### RED-CGSC-005 — AC-CGSC-005 (관측 형태: authored-test)

- 명령(단일 호출): `go test -count=1 -v ./internal/cli/ -run '^TestCodexReviewGatePrimaryPolicySharedByBothPaths$'`
- exit 코드: 1
- 적색 이유: 두 자동 경로 모두 primary 축을 모른다 — Claude 경로는 리뷰를 밟고(`gate skipped=false, rows 0`), Codex 경로는 receipt 루트를 간다(`chain skipped=false`, `ReceiptRead:true`). 하나의 정책이 양 경로에 존재하지 않다는 관측이다.
- (a) WHEN: 같은 세션 상태(primary 체크아웃, `tree_scope` review)를 두 자동 경로가 각각 평가할 때 — REQ-CGSC-003의 패리티 비교 자리.
- (b) 적색 입력: 동일 primary 픽스처를 `HandleCodexReviewGate`와 `codexReviewMember` 양쪽으로 통과 — 올바른 구현은 양쪽 모두 스킵.
- (c) 도달성: `go test` exit 1 — M2 착지 전 매 실행에서 적색.
- 원문 stdout:

```text
=== RUN   TestCodexReviewGatePrimaryPolicySharedByBothPaths
    codex_review_gate_primary_scope_red_test.go:88: both automatic paths must skip a primary-checkout tree session; gate skipped=false (rows 0), chain skipped=false (outcome {Number:0 Name: Decision:deny Class:unmeasured Status: Reason:moai hook codex-review-gate: not measured on this tree (receipt absent). Run `moai verify codex-review`, then end the turn again. (continuation 1 of 3 before the stop is allowed as unverified) ReceiptRead:true ElapsedMS:0 Err: Advisory: Discards:[]})
--- FAIL: TestCodexReviewGatePrimaryPolicySharedByBothPaths (1.91s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.949s
FAIL
```

### RED-CGSC-007 — AC-CGSC-007 (관측 형태: authored-test)

- 명령(단일 호출): `go test -count=1 -v ./internal/cli/ -run '^TestReviewableFromPorcelainRuntimeConfigOnlyFalse$'`
- exit 코드: 1
- 적색 이유: 런타임 관리 **설정** 표면(`.claude/settings.json`, `.claude/settings.local.json`, `.moai/config/sections/workflow.yaml`)이 트리 자기게이트에서 여전히 검사 가능으로 계수된다(현행 `reviewGateRuntimePrefixes`는 상태 표면만 배제). 대조군 행(`cmd/moai/main.go` → reviewable)이 통과해 올바른 이유의 적색이다.
- (a) WHEN: 트리 클래스 세션의 변경이 설정 표면뿐일 때 자기게이트가 porcelain을 판정하는 순간 — 이 경우 reviewer 호출 자체가 없어야 의미다.
- (b) 적색 입력: 설정 표면 전용 porcelain 페이로드 — 올바른 구현은 false(무검사 ALLOW).
- (c) 도달성: `go test` exit 1 — M3 착지 전 매 실행에서 적색.
- 원문 stdout:

```text
=== RUN   TestReviewableFromPorcelainRuntimeConfigOnlyFalse
=== RUN   TestReviewableFromPorcelainRuntimeConfigOnlyFalse/local_claude_settings
    codex_review_gate_primary_scope_red_test.go:111: runtime-managed config surface ".claude/settings.json" must not count as reviewable on the tree path
=== RUN   TestReviewableFromPorcelainRuntimeConfigOnlyFalse/runtime-written_settings_variant
    codex_review_gate_primary_scope_red_test.go:111: runtime-managed config surface ".claude/settings.local.json" must not count as reviewable on the tree path
=== RUN   TestReviewableFromPorcelainRuntimeConfigOnlyFalse/managed_config_tree
    codex_review_gate_primary_scope_red_test.go:111: runtime-managed config surface ".moai/config/sections/workflow.yaml" must not count as reviewable on the tree path
--- FAIL: TestReviewableFromPorcelainRuntimeConfigOnlyFalse (0.00s)
    --- FAIL: TestReviewableFromPorcelainRuntimeConfigOnlyFalse/local_claude_settings (0.00s)
    --- FAIL: TestReviewableFromPorcelainRuntimeConfigOnlyFalse/runtime-written_settings_variant (0.00s)
    --- FAIL: TestReviewableFromPorcelainRuntimeConfigOnlyFalse/managed_config_tree (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.486s
FAIL
```

### RED-CGSC-008 — AC-CGSC-008 (관측 형태: authored-test)

- 명령(단일 호출): `go test -count=1 -v ./internal/cli/ -run '^TestCodexReviewGateRuntimeDriftFindingsReclassified$'`
- exit 코드: 1
- 적색 이유: 발견 전부가 런타임 관리 설정 표면(`.claude/settings.json:13`, `.moai/config/sections/workflow.yaml:65` — 처분 기록 #2·#4의 실측 형태)만을 겨냥한 fail 판정인데 게이트가 BLOCK을 낸다 — 재분류 경로가 존재하지 않는다는 관측. 전제 단언(파서가 두 발견의 File을 정확히 두 표면으로 추출)이 통과한 뒤의 적색이라 파서 변형에 취약하지 않다.
- (a) WHEN: 트리 스코프 review가 실행된 뒤 verdict를 판정하는 순간 — 스킵이 아니라 리뷰 **후** 판정에서만 의미 있다(혼합 케이스는 §B.5).
- (b) 적색 입력: 설정 표면 전용 발견을 싣고 fail로 합성되는 codex 세션 스크립트 — 올바른 구현은 ALLOW + 재분류 행.
- (c) 도달성: `go test` exit 1 — M3 착지 전 매 실행에서 적색.
- 원문 stdout:

```text
=== RUN   TestCodexReviewGateRuntimeDriftFindingsReclassified
{"basis":"no card branch (unreadable or detached)","gate":"codex-review-gate","scope":"tree"}
    codex_review_gate_primary_scope_red_test.go:155: findings targeting only runtime-managed config surfaces must not block the turn (reclassify + record), got &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate: - [P1] `.claude/settings.json:13` personal PATH entry drifted
        - [P2] `.moai/config/sections/workflow.yaml:65` auto_cleanup local drift HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
--- FAIL: TestCodexReviewGateRuntimeDriftFindingsReclassified (0.11s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.342s
FAIL
```

### RED-CGSC-010 — AC-CGSC-010 (관측 형태: authored-test — 현재 아티팩트 실행 픽스처)

- 명령(단일 호출): `go test -count=1 -v ./internal/template/ -run '^TestSyncPhaseGateExcludesReportsGoFixture$'`
- exit 코드: 1
- 적색 이유: `.moai/reports/lab`에 파킹된 Go 픽스처(go.mod + 깨진 .go)가 세 갈래 전부 — ③ 미추적 팔, ①② tracked 팔(커밋 diff), 템플릿 미러 — 에서 수집에 들어가 모듈 루트로 베트되고 스퓨리어스 블록을 낸다. stub 로그의 `go -C …/.moai/reports/lab vet ./…` 행이 수집 침입 그 자체를 말한다. 대조군 서브테스트(`root_fixture_still_gates`, PASS)가 통과해 게이트 전체가 죽은 게 아니라 reports 수집만 틀렸다는 올바른 이유의 적색이다(D3 정합 — 픽스처가 두 팔을 함께 관측).
- (a) WHEN: sync-phase 커밋 뒤의 턴 끝 게이트가 델타 수집·언어 탐색·모듈 루트를 구성할 때 — reports 픽스처가 어느 집합에도 없어야 의미다.
- (b) 적색 입력: `.moai/reports/lab/{go.mod,broken.go}` 픽스처(미추적 팔 / tracked 팔) — 올바른 구현은 빈 stdout(무블록) + stub 로그에 reports 호출 부재.
- (c) 도달성: `go test` exit 1 — M4 착지 전 매 실행에서 적색.
- 원문 stdout:

```text
=== RUN   TestSyncPhaseGateExcludesReportsGoFixture
=== RUN   TestSyncPhaseGateExcludesReportsGoFixture/untracked_reports_fixture_is_not_collected
    hook_gate_reports_exclude_test.go:147: SYNC_GATE_REPORTS_SWEEP: an untracked Go fixture under .moai/reports drove a spurious block.
        stub log: "stub go -C /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestSyncPhaseGateExcludesReportsGoFixtureuntracked_reports_fixt2998347028/001/.moai/reports/lab vet ./...\nstub go -C /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestSyncPhaseGateExcludesReportsGoFixtureuntracked_reports_fixt2998347028/001/.moai/reports/lab build ./...\n"
        out: "{\"hookSpecificOutput\":{\"hookEventName\":\"Stop\",\"decision\":\"block\",\"reason\":\"go vet failed\"},\"systemMessage\":\"sync-phase quality gate BLOCKED: go vet failed (go vet=7 go build=7 deps_modified=1). Detail: .moai/logs/sync-quality-gate.log\"}\n"
=== RUN   TestSyncPhaseGateExcludesReportsGoFixture/tracked_reports_fixture_is_not_collected
    hook_gate_reports_exclude_test.go:154: SYNC_GATE_REPORTS_SWEEP: a committed Go fixture under .moai/reports drove a spurious block.
        stub log: "stub go -C /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestSyncPhaseGateExcludesReportsGoFixturetracked_reports_fixtur2512011694/001/.moai/reports/lab vet ./...\nstub go -C /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestSyncPhaseGateExcludesReportsGoFixturetracked_reports_fixtur2512011694/001/.moai/reports/lab build ./...\n"
        out: "{\"hookSpecificOutput\":{\"hookEventName\":\"Stop\",\"decision\":\"block\",\"reason\":\"go vet failed\"},\"systemMessage\":\"sync-phase quality gate BLOCKED: go vet failed (go vet=7 go build=7 deps_modified=1). Detail: .moai/logs/sync-quality-gate.log\"}\n"
=== RUN   TestSyncPhaseGateExcludesReportsGoFixture/mirror_carries_the_same_exclusion
    hook_gate_reports_exclude_test.go:161: SYNC_GATE_REPORTS_SWEEP: the template mirror sweeps .moai/reports too.
        stub log: "stub go -C /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestSyncPhaseGateExcludesReportsGoFixturemirror_carries_the_sam944560758/001/.moai/reports/lab vet ./...\nstub go -C /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestSyncPhaseGateExcludesReportsGoFixturemirror_carries_the_sam944560758/001/.moai/reports/lab build ./...\n"
        out: "{\"hookSpecificOutput\":{\"hookEventName\":\"Stop\",\"decision\":\"block\",\"reason\":\"go vet failed\"},\"systemMessage\":\"sync-phase quality gate BLOCKED: go vet failed (go vet=7 go build=7 deps_modified=1). Detail: .moai/logs/sync-quality-gate.log\"}\n"
=== RUN   TestSyncPhaseGateExcludesReportsGoFixture/root_fixture_still_gates
--- FAIL: TestSyncPhaseGateExcludesReportsGoFixture (12.69s)
    --- FAIL: TestSyncPhaseGateExcludesReportsGoFixture/untracked_reports_fixture_is_not_collected (3.85s)
    --- FAIL: TestSyncPhaseGateExcludesReportsGoFixture/tracked_reports_fixture_is_not_collected (4.19s)
    --- FAIL: TestSyncPhaseGateExcludesReportsGoFixture/mirror_carries_the_same_exclusion (2.45s)
    --- PASS: TestSyncPhaseGateExcludesReportsGoFixture/root_fixture_still_gates (2.19s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	12.974s
FAIL
```

### 재분류 기록 — RED가 존재할 수 없어 관측하지 않은 것

AC-CGSC-002·003·009·012는 「pre-SPEC과 동일하게 유지된다」를 단언하는 보존 기준으로, 구현 전 트리에서 그 기준이 이미 충족된다 — 적색으로 만들 입력이 정의상 없다(verification-completeness §2.1 undecidable 처분). 이 네 건은 관측되지 않은 적색을 만들어 내지 않았고, release-blocking 자격을 잃고 G 클래스로 이행했다(§D.1). 파편화된 원문 대신 이 한 단락으로 기록한다.

## §D AC Matrix

### AC-CGSC-001 — Facet 1 재현: primary 착지 비카드 세션은 스킵된다 (R)
REQ: REQ-CGSC-001, REQ-CGSC-002, REQ-CGSC-011
- **Given** 게이트 enabled, `tree_scope` 기본 review, primary-scope 기본 skip, 그리고 세션 CWD가 primary 체크아웃(git-dir == git-common-dir, 검사 가능 미커밋 존재)인 픽스처일 때
- **When** `HandleCodexReviewGate`가 그 입력으로 호출된다
- **Then** 출력은 ALLOW(빈 decision), review RPC 미호출, 구조화 로그가 primary 스킵 basis(트리 경로·정책값 포함, tree_scope 행과 구별되는)를 운반한다
- 검증: `go test -count=1 -v ./internal/cli/ -run '^TestCodexReviewGatePrimaryCheckoutSkip$'` — RED: §D.0 RED-CGSC-001(authored-test 관측).

### AC-CGSC-002 — 연결 워크트리 회귀 (G — regression-guard)
REQ: REQ-CGSC-002, REQ-CGSC-006
- **Given** 같은 설정에서 세션 트리가 연결 워크트리(git-dir != common-dir, 비 WT 브랜치)일 때
- **When** 게이트가 평가된다
- **Then** primary 스킵이 적용되지 않고 pre-SPEC 트리 스코프 동작(자기게이트 → tree_scope 정책)이 유지된다
- 이행: M2 착지 시 연결 워크트리 픽스처(wtnobase 선례)의 특성화 관측 — 구현 전후 동일 출력. 구현 전 적색은 정의상 없다(§D.0 재분류 기록).

### AC-CGSC-003 — 카드 스코프 회귀 (G — regression-guard)
REQ: REQ-CGSC-005
- **Given** WT- 카드 워크트리 세션일 때
- **When** 게이트가 평가된다
- **Then** 카드 스코프 요청(baseBranch target + merge-base, cwd 카드 트리)과 자기게이트가 pre-SPEC과 동일하고, primary 정책은 카드 클래스에서 평가되지 않는다
- 이행: M2 착지 시 `newCardScopeFixture` 카드 케이스 특성화 관측 + `go test ./internal/cli/...` 기존 카드 케이스 green 유지. 구현 전 적색은 정의상 없다(§D.0 재분류 기록).

### AC-CGSC-004 — 명시 복원 (P2)
REQ: REQ-CGSC-004
- **Given** `primary_scope: review`로 설정되었을 때
- **When** primary 착지 트리 클래스 세션의 턴이 끝난다
- **Then** pre-SPEC 트리 스코프 동작이 유지되고(tree_scope가 그대로 지배), 노멀라이저는 미지 값을 review가 아닌 skip-기본으로 정규화한다(§F.2 정책표)
- 검증: `go test -v ./internal/config/... -run '^TestNormalizeCodexReviewGatePrimaryScope$'` + `codex_review_gate_test.go` 복원 케이스.

### AC-CGSC-005 — 두 자동 경로 패리티 (R)
REQ: REQ-CGSC-003
- **Given** primary 스킵 정책이 배치되었을 때
- **When** Claude Stop 훅 경로와 Codex Stop 체인 멤버 6 경로가 같은 세션 상태를 평가한다
- **Then** 둘은 같은 정책 함수를 거쳐 같은 판정을 내린다(명시적 생산자 제외 — 현행 유지)
- 검증: `go test -count=1 -v ./internal/cli/ -run '^TestCodexReviewGatePrimaryPolicySharedByBothPaths$'` — RED: §D.0 RED-CGSC-005. (REQ-CRO-006 패리티 선례 패턴, `codex_review_gate_wiring_test.go` 소관.)

### AC-CGSC-006 — 판별 불가 페일오픈 (P2)
REQ: REQ-CGSC-006, REQ-CGSC-001
- **Given** 세션 디렉터리가 비-git이거나 git 오류일 때
- **When** 스코프가 해상된다
- **Then** primary 스킵은 적용되지 않고 기존 basis("no session tree" 계열)와 동작이 유지된다
- 검증: `go test -v ./internal/cli/... -run '^TestCodexReviewGateNonGitDirNoPrimarySkip$'`.

### AC-CGSC-007 — Facet 2 자기게이트: 설정 표면 전용 변경은 검사 가능 아님 (R)
REQ: REQ-CGSC-007, REQ-CGSC-011
- **Given** 트리 클래스 세션의 변경이 `.claude/settings.json` + `.moai/config/sections/workflow.yaml`뿐일 때
- **When** 트리 자기게이트가 porcelain 페이로드를 판정한다
- **Then** 검사 가능 false → reviewer 미호출 ALLOW. 일반 소스 경로는 계속 검사 가능(대조군)
- 검증: `go test -count=1 -v ./internal/cli/ -run '^TestReviewableFromPorcelainRuntimeConfigOnlyFalse$'` — RED: §D.0 RED-CGSC-007(authored-test 관측, 대조군 포함).

### AC-CGSC-008 — Facet 2 혼합: 런타임 표면 전용 발견은 비블록 (R)
REQ: REQ-CGSC-008, REQ-CGSC-011
- **Given** 트리 스코프 review가 실행됐고 모든 발견이 런타임 관리 설정 표면 경로만 겨냥할 때
- **When** verdict가 판정된다
- **Then** BLOCK이 아니라 ALLOW이고, 재분류(known runtime-managed drift) 행이 기록된다
- 검증: `go test -count=1 -v ./internal/cli/ -run '^TestCodexReviewGateRuntimeDriftFindingsReclassified$'` — RED: §D.0 RED-CGSC-008(authored-test 관측, 파서 전제 단언 포함).

### AC-CGSC-009 — Facet 2 카드 경계: 공유 리스트 불변 (G — regression-guard)
REQ: REQ-CGSC-005
- **Given** 트리 전용 제외 세트가 도입됐을 때
- **When** 카드 스코프 경로 필터와 receipt 바인딩이 평가된다
- **Then** 공유 `reviewGateRuntimePrefixes` 내용은 불변이고, `.moai/config/...` 아래 카드 커밋은 여전히 카드 작업으로 계수된다
- 이행: M3 착지 시 특성화 관측(공유 리스트 6항목 불변 단언 + 카드 경로 계수) + 기존 카드 테스트 green 유지. 구현 전 적색은 정의상 없다(§D.0 재분류 기록).

### AC-CGSC-010 — Facet 3 재현: `.moai/reports` Go 픽스처가 수집에 안 들어온다 (R)
REQ: REQ-CGSC-009
- **Given** 픽스처 리포에 `.moai/reports/lab/go.mod` + 깨진 `.go`가 있을 때 — **미추적 팔(③)과 tracked 팔(①②) 양쪽 형태로**(D3 정합)
- **When** sync 게이트가 델타 수집·언어 탐색·모듈 루트를 구성한다
- **Then** `.moai/reports/` 경로가 어느 집합에도 없고 게이트는 블록하지 않는다. `.moai/reports` 밖의 일반 소스는 계속 게이트한다(대조군)
- 검증: `go test -count=1 -v ./internal/template/ -run '^TestSyncPhaseGateExcludesReportsGoFixture$'` — RED: §D.0 RED-CGSC-010(authored-test 관측; hook_gate_* 실행 픽스처 선례).

### AC-CGSC-011 — Facet 3 쌍둔 패리티 (P2)
REQ: REQ-CGSC-009
- **Given** 템플릿 미러와 추적본이 같은 커밋에서 바뀌었을 때
- **When** `make build` 이후 임베드와 추적본을 비교한다
- **Then** 동일 제외 항목이 양쪽에 존재하고 임베드에 반영된다
- 검증: `go test ./internal/template/...` 기존 패리티 축 green + `grep -c '(top,exclude).moai/reports'` 양 쌍둔 각 1 이상.

### AC-CGSC-012 — Facet 3 회귀: 기존 수집 shape-identical (G — regression-guard)
REQ: REQ-CGSC-010
- **Given** `.moai/reports/` 밖의 변경 집합일 때
- **When** sync 게이트가 수집한다
- **Then** 기존 배제 항목(배열 원문 직독 20항목, :257-266) 불변, 델타·탐지·키가 pre-SPEC과 동일 형태
- 이행: M4 착지 시 기존 `hook_gate_*` 테스트 green 유지 + 쌍둔 패리티(AC-CGSC-011) + 멤버십 검사 — `grep -c -e ':(top,exclude).moai/reports' -e ':(top,exclude).moai/state' .claude/hooks/moai/sync-phase-quality-gate.sh` **baseline 1(2026-10-03 실측, exit 0 — state 항목만 적중, reports 항목 부재의 적) → M4 후 2**. 항목 기준 검사이며 `grep -c` 줄 수 기준이 아니다 — 줄 수 기준(`grep -c ':(top,exclude)'` = 2)은 glob 항목을 못 세어 올바른 구현도 통과 못 하는 불가능 기준이었고 plan-audit iteration-1 D4로 교체됐다. 구현 전 적색은 정의상 없다(§D.0 재분류 기록).

## §D.1 분류

plan-audit iteration-1(MP-8) 수리로 재분류됐다. **R(RED-first, release-blocking)** = 001·005·007·008·010 — 구현 전 동일 테스트의 적색이 §D.0 장부에 원문 stdout·exit 코드·트리 SHA와 함께 관측돼 있고, 미충족 시 run 게이트 불통과. **G(regression-guard)** = 002·003·009·012 — 전부 「변하지 않음」을 단언하는 보존 기준으로 구현 전 트리에서 정의상 green이라 적색 입력이 존재하지 않는다(verification-completeness §2.1 undecidable 처분): release-blocking 자격이 없고, 특성화 관측(구현 전후 동일 출력)으로 이행하며, 채택 시 pass로 기록되지 않는다. **P2(완화 축)** = 004·006·011 — 미충족 시 PASS-WITH-DEBT 후보이나 반전 근거(004)는 감사에서 소명 요구. 재분류 전문은 §D.0 말미와 HISTORY(spec v0.1.1)에 있다.

## §D.2 REQ ↔ AC 추적표

| REQ | AC |
|-----|----|
| REQ-CGSC-001 | AC-CGSC-001, AC-CGSC-006 |
| REQ-CGSC-002 | AC-CGSC-001, AC-CGSC-002 |
| REQ-CGSC-003 | AC-CGSC-005 |
| REQ-CGSC-004 | AC-CGSC-004 |
| REQ-CGSC-005 | AC-CGSC-003, AC-CGSC-009 |
| REQ-CGSC-006 | AC-CGSC-002, AC-CGSC-006 |
| REQ-CGSC-007 | AC-CGSC-007 |
| REQ-CGSC-008 | AC-CGSC-008 |
| REQ-CGSC-009 | AC-CGSC-010, AC-CGSC-011 |
| REQ-CGSC-010 | AC-CGSC-012 |
| REQ-CGSC-011 | AC-CGSC-001, AC-CGSC-007, AC-CGSC-008 (로그 행은 각 Then 절에서 공동 관측) |

## §D.3 간접 검증

- 쌍둔 패리티(011)는 임베드 축 간접 증거 — `make build` 후 `git status`로 추적본·임베드 동시 반영을 관측.
- primary 판별은 `git rev-parse` 실측이 원천이며, seam 테스트는 그 위의 행동 고정 — 실측 픽스처(AC-001 픽스처)와 seam 단위 테스트가 쌍을 이룬다.
- RED-CGSC-010의 stub 로그는 「수집 침입」의 간접 관측면이다 — 블록 JSON이 아니라 `go -C …/.moai/reports/lab vet ./…` 호출 기록이 배제 누락의 직접 흔적이다.

## §D.4 클로저 게이트

- R 5건 green(§D.0 장부 원문과 E8 재확인 일치) + G 4건 특성화 관측 첨부 + 회귀 baseline 대비 신규 적색 0 + 커버리지 3패키지 ≥85% + windows 빌드 통과 — run-phase 종료 조건.

## §D.5 전방위 검사 (forward-looking)

- `primary_scope` 키명은 M1에서 확정된다 — 확정값을 spec.md §B REQ-CGSC-004 본문에 역반영하지 않는다(요구는 행동으로 고정, 키명은 plan 소관). 다만 문서화(sync 단계)에서 실제 키명과 일치시킨다. §F.2 정책표의 「명시적 `review`」 표현은 키명과 무관하게 유지된다.
- Go 미러(`codex_sync_gate.go`)에 수집 제외 목록이 생기는 후속 카드가 나오면 REQ-CGSC-009와의 정합을 그 카드가 소명한다.

## §D.6 품질 게이트 기준

TRUST 5 — Tested: 본 매트릭스 + 커버리지 85%(패키지 3). Readable/Unified: 기존 seam·주석 선례 추종, gofmt. Secured: 외부 입력 없음(훅 stdin 기존 경로), 권한 변경 없음. Trackable: Conventional Commits + 카드 id + `🗿 MoAI`.

## §D.7 Definition of Done

1. R 5건 green(§D.0 장부의 RED 원문과 E8 재확인 동봉) + G 4건 특성화 관측 + P2 3건 처리. 2. spec §D 순서 준수 증적(회귀선 선확정). 3. Template-First 순서 증적(M4 커밋에 미러·추적본 동반). 4. decision-index Q1-Q3 상태 갱신 가능 상태로 인계. 5. progress.md §E.2·§E.3 매꿈(manager-develop 소관).
