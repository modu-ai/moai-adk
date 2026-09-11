# t660 통합 창 기록 — 흡수 트리 재측정

- 카드: t660 · 레인: lane-6 · 브랜치: `WT-policy-ask-fixture`
- 창 지명: 리드, t587 release 뒤. 재측정 스코프는 `TestToolPolicyList_QueryFilters` 1회로 승인됨.
- 도구: `go version go1.26.8 darwin/arm64`

## 1. Claim

1. 로컬 develop `ae1673d8a` 는 흡수 시점에 최신이었다 — 원격 develop 대비 뒤처진 커밋 0.
2. 흡수 병합 `7397aaa33`(트리 `b940d4e1`) 위에서 `TestToolPolicyList_QueryFilters` 6개 하위 테스트가 모두 통과한다.
3. 흡수는 tool-policy 명령·로더·테스트·제품 YAML 어느 것도 바꾸지 않았다.

## 2. Evidence

### 2.1 창 획득과 develop 최신성

```
moai integration acquire --name lane-6
→ release-integration window acquired by e6946f98-39aa-4df4-9a3f-3b9caed0ab4d on WT-policy-ask-fixture

git fetch -q origin develop                               → FETCH_EXIT=0
git rev-parse origin/develop develop                      → 987eb7e4021ab7a74baef33242307087a74238ae / ae1673d8a6e39df68cad43881d6c6eea9da08a41
git rev-list --count --left-right origin/develop...develop → 0	20
```

fetch 뒤에도 원격이 `987eb7e40` 이고 로컬은 뒤처짐 0 · 앞섬 20(다른 레인의 미푸시 병합)이다. 흡수 대상은 로컬 develop 이다.

### 2.2 흡수

```
git merge --no-edit develop   → Merge made by the 'ort' strategy. (80 files changed, 4454 insertions(+), 50 deletions(-), 충돌 없음)
git rev-parse HEAD HEAD^{tree} HEAD^1 HEAD^2
→ 7397aaa337fe1281878f6903dc4b65e5e0066cb9
  b940d4e18f165acedf39a39883db1d952497d5de
  fef3cd67f6eb98fba3eec0ee8cb372e0e63a2a3b
  ae1673d8a6e39df68cad43881d6c6eea9da08a41
```

흡수로 들어온 코드는 `internal/cli/update*.go`, `internal/cli/update/{backup,report}/`, `internal/config/loader_git_mode.go` 와 t574·t587 증거·SPEC 파일이다.

### 2.3 tool-policy 표면 불변 (흡수 전 `fef3cd67f` ↔ 흡수 후 HEAD)

```
git rev-parse HEAD:.moai/config/sections/tool-policy.yaml develop:… fef3cd67f:…   → 3e2e5de5953c12909f171fd2ad3ddee816d16309 (셋 모두)
git hash-object .moai/config/sections/tool-policy.yaml                            → 3e2e5de5953c12909f171fd2ad3ddee816d16309
git rev-parse HEAD:internal/cli/tool_policy_test.go / git hash-object 작업 파일    → 2e3f551e3754df5c39dc5d69e2aabdd66d9928ff (둘 다, 수리 커밋 da42bd79d 의 blob)
git rev-parse fef3cd67f:internal/cli/tool_policy.go HEAD:internal/cli/tool_policy.go
→ 27716a96fc7d92b17411fb3f5749a4959a8ea2ba (둘 다)
git rev-parse fef3cd67f:internal/config/toolpolicy HEAD:internal/config/toolpolicy
→ ce9a0fa246cbdf4eb04e2dc46d1e73cce42c9027 (둘 다)
git diff --stat fef3cd67f HEAD -- internal/cli/tool_policy.go internal/cli/tool_policy_test.go internal/config/toolpolicy .moai/config/sections/tool-policy.yaml
→ (출력 없음), DIFF_EXIT=0
대조군: git diff --stat fef3cd67f HEAD -- internal/cli/update.go
→ internal/cli/update.go | 33 +++------------------------------
```

같은 diff 형식이 실제로 바뀐 파일은 보여 주므로, tool-policy 경로의 빈 출력은 경로 오기가 아니라 변경 없음이다. blob·트리 해시 비교가 같은 결론을 독립적으로 낸다.

### 2.4 재측정 (`window-green.txt`, 사전 확인 `slot-precheck-window.txt`)

사전 확인: 2026-09-10T18:29:22Z, load 33.82 30.42 25.25, 다른 go test 2줄(1건 실행 중), 관측자 대조군 1.

```
unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -count=1 -v -timeout 600s -run '^TestToolPolicyList_QueryFilters'
```

```
--- PASS: TestToolPolicyList_QueryFilters (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_deny (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_tool_Bash (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_ask (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_irreversible (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_allow (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/all_entries (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.041s
EXIT=0
```

## 3. Baseline-attribution

흡수 병합 `7397aaa33`, 트리 `b940d4e18f165acedf39a39883db1d952497d5de`. 재측정 직전 `git status --porcelain --untracked-files=all` 출력 없음. 부하가 있는 상태(load 33.82)라 판정은 이름별 PASS/FAIL 줄만 쓰고 시간값은 쓰지 않는다.

## 4. Gaps

- 이 창에서 뮤턴트는 다시 돌리지 않았다. 흡수가 테스트 파일·명령·로더를 바꾸지 않았으므로(2.3) 창 전 뮤턴트 RED(`mutant-red.txt`)의 대상 코드가 그대로라는 판독에 기댄다.
- `internal/cli` 패키지 전체와 같은 파일의 `InvalidFlagValues`·`JSONFormat` 은 실행하지 않았다(리드: ask 무관 판독으로 기록만).
- 흡수로 들어온 `update*.go` 변경의 검증은 t587 의 몫이며 이 창에서 재지 않았다.

## 5. Residual-risk

- develop 병합 트리는 이 기록을 담은 증거 커밋의 트리와 같아야 한다. 병합 뒤 `git rev-parse <merge>^{tree}` 로 대조해 리드 보고에 싣는다.
- 병합 이후 로컬 develop 에 다른 레인이 더 병합하면 이 재측정은 그 커밋들을 포함하지 않는다.
