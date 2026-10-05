# t660 판정서 — `TestToolPolicyList_QueryFilters/filter_ask` develop 회귀 수리

- 카드: t660
- 브랜치: `WT-policy-ask-fixture` (워크트리 `.claude/worktrees/t660`)
- 카드 기준점: `CARD_BASE=$(git merge-base develop HEAD)` → `987eb7e40` (작성 시점 판독, 흡수 전)
- 커밋 순서: `aa820b716` RED 증거 → `da42bd79d` 테스트 수리 → (이 판정서를 담은 증거 커밋)
- 도구: `go version go1.26.8 darwin/arm64`
- 병합: 아직 하지 않음. 통합 창 지명 전이다.

## 1. Claim

1. develop `987eb7e40` 에서 `TestToolPolicyList_QueryFilters/filter_ask` 는 실패한다. 원인은 커밋된 `tool-policy.yaml` 에 `decision: ask` 항목이 0개라는 것이며, 나머지 5개 하위 테스트는 통과한다.
2. 수리 후 이 테스트는 제품 YAML 이 아니라 테스트가 소유한 fixture(항목 5개, ask 2개 포함)로 필터를 검증하고, 6개 하위 테스트가 모두 통과한다.
3. fixture 에서 ask 항목만 지우는 뮤턴트를 적용하면 `filter_ask` 가 실패한다(행 0개, 기대 2개). 테스트가 ask 필터를 공허하게 통과시키지 않는다는 뜻이다.
4. 제품 정책 `.moai/config/sections/tool-policy.yaml` 은 바뀌지 않았다(ask 0 은 운영자 결정이므로 되돌리지 않는다).

## 2. Evidence

### 2.1 RED — develop 트리 (`red-develop.txt`, 사전 확인 `slot-precheck-red.txt`)

```
unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -count=1 -v -run '^TestToolPolicyList_QueryFilters
```

```
    tool_policy_test.go:97: output missing substring "ask";
        got:
        TOOL  ARGS  RISK  DECISION  OWNER  AUDIT
--- FAIL: TestToolPolicyList_QueryFilters (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_deny (0.01s)
    --- PASS: TestToolPolicyList_QueryFilters/all_entries (0.02s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_allow (0.02s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_tool_Bash (0.02s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_irreversible (0.02s)
    --- FAIL: TestToolPolicyList_QueryFilters/filter_ask (0.02s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.327s
EXIT=1
```

출력이 헤더 한 줄뿐이다 — 필터가 고른 항목이 없다는 뜻이고, 실패 이유가 "ask 항목 부재"와 일치한다.

제품 YAML 의 항목 수 (같은 패턴의 대조군 포함):

```
/usr/bin/grep -c -E '^[[:space:]]*risk_tier: irreversible' .moai/config/sections/tool-policy.yaml   → 60
/usr/bin/grep -c -E '^[[:space:]]*decision: ask'            .moai/config/sections/tool-policy.yaml   → 0
/usr/bin/grep -c -E '^[[:space:]]*decision: (allow|deny)'   .moai/config/sections/tool-policy.yaml   → 173
```

### 2.2 수리 내용 (`da42bd79d`, `internal/cli/tool_policy_test.go` 한 파일, +66/−13)

- 정책 경로를 `../../.moai/config/sections/tool-policy.yaml` 에서 `t.TempDir()` 안에 쓴 `queryFilterFixture` 로 바꿨다.
- fixture 는 6필드 스키마를 지키는 항목 5개다: deny 1 · allow 2 · ask 2, irreversible 2, `Bash` 3.
- 판정을 "출력 어딘가에 부분 문자열이 있다"에서 두 가지로 강화했다. ① 필터별 정확한 행 수(기존에 선언만 되고 쓰이지 않던 `wantCount`) ② 모든 행이 필터 토큰을 포함.
- audit·owner·args 문자열이 어떤 필터 토큰도 담지 않게 했다. 기계 점검 결과 (`fixture-tokens.txt`, 도구 `tools/fixture_tokens.py`):

```
control_leaks 1
Bash irreversible deny leaks []
Bash read allow leaks []
Read read allow leaks []
Write write ask leaks []
Bash irreversible ask leaks []
entries 5 total_leaks 0
EXIT=0
```

대조군(audit 에 "task list" 를 넣은 가짜 항목)이 1건으로 잡히므로, 0건은 검사기가 못 본 결과가 아니다.

### 2.3 GREEN (`green-fixture.txt`, 사전 확인 `slot-precheck-green.txt`: load 8.48, 다른 go test 0건, 관측자 대조군 1)

```
unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -count=1 -v -timeout 600s -run '^TestToolPolicyList_QueryFilters'
```

```
--- PASS: TestToolPolicyList_QueryFilters (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_deny (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_tool_Bash (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_irreversible (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_allow (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_ask (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/all_entries (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.095s
EXIT=0
```

하위 테스트 6개가 이름으로 전부 찍혔다 — 셀렉터 0매치로 인한 공허 초록이 아니다.

### 2.4 뮤턴트 RED — fixture 에서 ask 제거 (`mutant-apply.txt` → `mutant-red.txt` → `mutant-revert.txt`)

적용 (`tools/mutant_apply.py`):

```
entries_before 5
entries_after  3
ask_lines_before 2
ask_lines_after  0
before dac099f591af1b4d5073932c4f3719f81b250d26b40e483bb5069ecc25cf5ebc
after  f04bae8ba0be63630003edf77045fabd64e52926a460bc18d61d4e5547b8563f
```

같은 명령(2.3)으로 실행:

```
    tool_policy_test.go:110: entry rows = 1; want 2;
    tool_policy_test.go:110: entry rows = 2; want 3;
    tool_policy_test.go:110: entry rows = 3; want 5;
    tool_policy_test.go:110: entry rows = 0; want 2;
--- FAIL: TestToolPolicyList_QueryFilters (0.00s)
    --- FAIL: TestToolPolicyList_QueryFilters/filter_irreversible (0.00s)
    --- FAIL: TestToolPolicyList_QueryFilters/filter_tool_Bash (0.00s)
    --- FAIL: TestToolPolicyList_QueryFilters/all_entries (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_allow (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_deny (0.00s)
    --- FAIL: TestToolPolicyList_QueryFilters/filter_ask (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.963s
EXIT=1
```

`filter_ask` 는 행 0개로 실패했다. ask 항목을 포함해 세던 세 필터(irreversible 2→1, Bash 3→2, 전체 5→3)도 함께 실패했고, ask 와 무관한 allow·deny 는 그대로 통과했다 — 뮤턴트가 건드린 범위와 실패 범위가 정확히 겹친다.

복원 (`tools/mutant_revert.py`):

```
mutated  f04bae8ba0be63630003edf77045fabd64e52926a460bc18d61d4e5547b8563f
original dac099f591af1b4d5073932c4f3719f81b250d26b40e483bb5069ecc25cf5ebc
restored dac099f591af1b4d5073932c4f3719f81b250d26b40e483bb5069ecc25cf5ebc
RESTORED_MATCHES_ORIGINAL True
EXIT=0
```

### 2.5 커밋본 = 실행본, 제품 YAML 불변

```
shasum -a 256 internal/cli/tool_policy_test.go            → dac099f5…5ebc   (GREEN 을 돌린 바이트, 복원 후와 동일)
git rev-parse HEAD:internal/cli/tool_policy_test.go        → 2e3f551e3754df5c39dc5d69e2aabdd66d9928ff
git hash-object internal/cli/tool_policy_test.go           → 2e3f551e3754df5c39dc5d69e2aabdd66d9928ff
git rev-parse develop:.moai/config/sections/tool-policy.yaml HEAD:.moai/config/sections/tool-policy.yaml
  → 3e2e5de5953c12909f171fd2ad3ddee816d16309 (둘 다)
git hash-object .moai/config/sections/tool-policy.yaml     → 3e2e5de5953c12909f171fd2ad3ddee816d16309
git diff --stat HEAD (수리 커밋 전)                          → internal/cli/tool_policy_test.go | 79 (1 file)
```

## 3. Baseline-attribution

- RED: develop `987eb7e40` 트리(워크트리 HEAD = 로컬 develop, origin/develop 과 동일 — 흡수 병합이 "Already up to date"). 2026-09-10T18:07:54Z 사전 확인, load 49.48, 다른 go test 2건 실행 중(`slot-precheck-red.txt`). 이 실행은 판정(PASS/FAIL 이름 줄)만 쓰고 시간값은 쓰지 않는다.
- GREEN·뮤턴트: `aa820b716` 위 작업 트리, 테스트 파일 sha256 `dac099f5…`(= 커밋 `da42bd79d` 의 blob). 2026-09-10T18:14:04Z 사전 확인.
- 세 번의 `internal/cli` 컴파일은 리드가 승인한 슬롯 1회 묶음 안에서 직렬로 실행했다.

## 4. Gaps

- 같은 파일의 `TestToolPolicyList_InvalidFlagValues`·`TestToolPolicyList_JSONFormat` 은 여전히 커밋된 YAML 을 읽는다. 셀렉터가 `QueryFilters` 로 한정돼 이번 슬롯에서는 돌리지 않았다. JSON 테스트는 irreversible 항목(현재 60개)에, 무효 플래그 테스트는 오류 문자열에 기대므로 ask 0 과는 무관해 보이지만, 이는 코드 판독이지 실행 관측이 아니다.
- `internal/cli` 패키지 전체는 실행하지 않았다(레인 규율 — 전체 판정은 develop push 뒤 CI 몫).
- 병합 트리 재측정은 아직이다. 통합 창에서 로컬 develop 흡수 후 같은 명령으로 다시 재야 한다.
- 다른 테스트가 커밋된 정책 파일에 기대는지는 경로 문자열(`tool-policy.yaml`, `../../`, `.moai/config/sections`, `LoadFromProjectDir(`) 검색으로만 확인했다. 실행 중 루트 탐색으로 도달하는 경로는 이 검색으로 보이지 않는다.

## 5. Residual-risk

- 제품 정책이 앞으로 ask 를 다시 쓰든 안 쓰든 이 테스트는 영향을 받지 않는다. 대신 제품 YAML 이 필터 명령과 실제로 맞물리는지는 이 테스트가 더 이상 보지 않는다 — 그 축은 로더 검증(`Validate`)과 JSON 테스트가 부분적으로만 덮는다.
- fixture 토큰 점검은 부분 문자열 기준이다. 나중에 fixture 에 항목을 추가하는 사람이 audit 에 "task" 같은 단어를 넣으면 행-토큰 판정이 느슨해질 수 있다(행 수 판정은 여전히 잡는다). `tools/fixture_tokens.py` 로 다시 점검할 수 있다.
- 텍스트 출력 파싱은 헤더가 `TOOL` 로 시작하고 행마다 줄 하나라는 현재 `renderList` 형식에 기댄다. 형식이 바뀌면 이 테스트가 먼저 알려 준다(헤더 없음 → Fatal).

## Card Cross-Check

| 항목 | card |
|---|---|
| filter_ask develop 회귀 수리 | t660 |
