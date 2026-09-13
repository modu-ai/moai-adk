# t619 sync 감사 보고서 — SPEC-TOOLPOLICY-DRIFT-GUARD-001

- card: t619
- 감사자: sync-auditor (독립 감사, 읽기 전용. 이 파일 외에는 쓰지 않았다)
- 측정 트리: `.claude/worktrees/t619` @ `1ff407dfe47935b34b355a16c4a529d440fe0655`, 브랜치 `WT-toolpolicy-drift`
- 기준(merge-base): `d1b61005d`
- 툴체인: `go version go1.26.4 darwin/arm64`, `golangci-lint has version 2.10.1 built with go1.26.0`
- 평가 프로필: SPEC 프런트매터에 `evaluator_profile` 없음. 프로필 파일은 읽지 않았고 내장 기본값(가중치 40/25/20/15, 필수 통과 차원 Functionality·Security)으로 채점했다(갭 1).

## 판정

**Overall Verdict: PASS** — 가중 조화 평균 **90.7 / 100**

단, AC-TDG-009 는 측정 결과 그대로 **FAIL** 이며, 이 카드의 변경이 만든 회귀가 아니라 **t609(`1ac333952`)가 남긴 선재 결함**으로 귀속한다(§1.3). 이 카드의 판정을 떨어뜨리는 사유로는 쓰지 않았지만, **develop push 전에 수리 카드 t660 이 닫혀야 한다는 통합 게이트 조건은 그대로 유효하다.**

---

## 1. 주장 (Claim)

### 1.1 AC 표

| AC | 판정 | 근거 요약 (이번 감사에서 다시 잰 값) |
|---|---|---|
| AC-TDG-001 | PASS | `make tool-policy-drift-check` rc=0. 두 커밋 트리 테스트 PASS. 오버레이로 base YAML ↔ HEAD settings 를 비교하면 차이 20개, 분할 7/1/12/0 (붉은색 짝이 공허하지 않음) |
| AC-TDG-002 | PASS | HEAD 트리로 빌드한 스크래치 바이너리의 생성 결과 `allow=114 ask=0 deny=48 env_gated_skipped=5`, 정렬 목록 `allow diff_rc=0` `deny diff_rc=0`, ask 길이 `0` `0` |
| AC-TDG-003 | PASS | base↔HEAD settings 두 파일 `settings_diff_rc=0`. 검사 전후 입력 sha256 동일, 검사 전후 `git status` 동일(기존 미추적 1건만) |
| AC-TDG-004 | PASS (귀속) | 감사자는 추적 파일을 변이할 수 없어 워킹 트리 변이는 재실행하지 않았다. progress.md 의 조건부 사슬 기록(`mutate=DONE` → rc=2·차이 1줄 → `restore=DONE` → rc=0·`diff_rc=0`)과, 같은 판정 함수를 격리 사본에서 변이한 `yaml_side_flip`·`settings_side_missing` PASS 를 근거로 삼는다(갭 2) |
| AC-TDG-005 | PASS | Mutation 하위 7개 PASS. 독립 뮤턴트 2종(중복 검사 제거 → `duplicate_settings_*` 3개만 FAIL, 겹침 검사 제거 → `allow_deny_overlap` 1개만 FAIL) |
| AC-TDG-006 | PASS | FailClosed 하위 10개 PASS, SKIP 0. 독립 뮤턴트 2종(빈 집합 검사 제거 → `empty_*` 4개 FAIL, 엄격 목록 디코드 제거 → `wrong_type_settings_list` FAIL) |
| AC-TDG-007 | PASS | 부재 grep 다섯 개 모두 `0`, 양성 `tool-policy-drift-check` YAML `1`·types.go `2`, 대조 `165`·`1` |
| AC-TDG-008 | PASS | `Makefile:34` `build:` 선행에 `tool-policy-drift-check`, `make -n build` 에서 검사가 `templ generate`·`go build` 보다 앞(5행 < 7행 < 9행), `go_code` 블록 15줄에 `.claude/settings.json` |
| AC-TDG-009 | **FAIL — 선재 결함 귀속, 범위 밖** | toolpolicy 패키지 rc=0·RoundTrip PASS 1·FAIL 0, config rc=0·PASS 1, **cli rc=1** (`filter_ask`) |
| AC-TDG-010 | PASS | list rc=0, MultiEdit `0`, 삭제 대상 deny `0`, env_gate Write deny `1`, env_gate `5`, 해시 `5e0cba52…6d27` 불변, 일곱 도구 allow 전부 존재, Read deny `5`, 항목 `167` |

통과 9 / 실패 1 (실패 1건은 선재 결함 귀속).

### 1.2 차원 점수

| 차원 | 점수 | 판정 | 핵심 근거 |
|---|---|---|---|
| Functionality (40%) | 88 | PASS | 범위 안 AC 9개 전부 재측정 통과. AC-TDG-009 는 측정상 FAIL 이나 이 변경의 회귀가 아님을 입력 동일성으로 확인(감점 반영) |
| Security (25%) | 95 | PASS | 적용 권한 바이트 불변(`settings_diff_rc=0`), 검사는 읽기 전용(sha·status 불변), Critical/High 0. Low 1건(F2, 기존 코드) |
| Craft (20%) | 90 | PASS | 커버리지 89.1%(기준 85), `0 issues.`, 뮤턴트 4종 모두 기대한 하위 테스트만 붉음, 원인별 센티널 분리. Low 2건(F3·F4) |
| Consistency (15%) | 92 | PASS | 정정된 주장이 생성기 실제 동작(권한 영역만 치환)과 일치, `agents-emit-check` 선례와 같은 배치·명명 |

가중 조화 평균: 1 / (0.40/88 + 0.25/95 + 0.20/90 + 0.15/92) = **90.7**

필수 통과 방화벽: Functionality 88, Security 95 — 둘 다 통과.

### 1.3 AC-TDG-009 처리 방식과 이유

**판정: 측정상 FAIL, 이 카드 변경의 회귀가 아닌 선재 결함(t609 귀속), 이 카드 판정에서는 범위 밖으로 처리한다.**

이유는 네 가지 관측이 모두 같은 방향을 가리키기 때문이다.

1. 실패하는 테스트의 코드가 바뀌지 않았다: `git diff --stat d1b61005d HEAD -- internal/cli/` → 출력 없음, `cli_diff_rc=0`.
2. 실패 조건(커밋 YAML 의 `decision: ask` 항목 존재)이 base 와 HEAD 에서 같다: base `0`, HEAD `0`. 이 카드는 allow·deny 항목만 바꿨다.
3. ask 를 없앤 커밋이 base 의 조상이다: `1ac333952^` 에서 `6`, `git merge-base --is-ancestor 1ac333952 d1b61005d` → `ancestor_rc=0`. 그 커밋은 `tool-policy.yaml` 에서 ask 를 지웠지만 `internal/cli/tool_policy_test.go` 는 건드리지 않았다(`git show --stat` 4개 파일에 없음).
4. 실패 출력이 "ask 항목이 없다"는 원인과 정확히 맞는다: `filter_ask` 출력이 머리글 한 줄뿐이다.

AC-TDG-009 의 요구는 "영향받는 기존 테스트 **무회귀**"다. 이 변경은 그 테스트의 입력도 코드도 바꾸지 않았으므로 회귀를 만들지 않았다. 반면 AC 명령의 문자 그대로의 기대값(`rc=0`)은 충족되지 않으므로 PASS 로 바꿔 적지 않고 FAIL 로 남긴다. 수리는 `internal/cli` 테스트 수정이라 이 SPEC 의 파일 범위(§4·§5) 밖이며, 리드가 발행한 t660 소관이다. **이 PASS 판정은 t660 이 develop push 전에 닫혀야 한다는 통합 조건을 면제하지 않는다.**

---

## 2. 증거 (Evidence)

모든 명령은 워크트리 최상위에서 이번 감사 중 실행했다. 출력은 발췌 없이 그대로 옮기되, 긴 목록은 해당 줄만 남겼다.

### 2.1 Functionality

```
$ make tool-policy-drift-check; echo "make_rc=$?"   (앞뒤로 shasum -a 256)
25e19e906f639044c38371fadec796f080cab2630b0d295478a62bdc47e73c39  .claude/settings.json
edff7e725e02738f78a77d8d9959a2f933b2e6fe705ce8d044e251048c544cdf  .moai/config/sections/tool-policy.yaml
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.179s
make_rc=0
25e19e906f639044c38371fadec796f080cab2630b0d295478a62bdc47e73c39  .claude/settings.json
edff7e725e02738f78a77d8d9959a2f933b2e6fe705ce8d044e251048c544cdf  .moai/config/sections/tool-policy.yaml
```

```
$ go test ./internal/config/toolpolicy/ -run 'TestToolPolicyDrift' -count=1 -v > <scratch>/drift-v.log 2>&1
drift_rc=0
--- PASS 줄 21 / --- FAIL 줄 0 / --- SKIP 줄 0
--- PASS: TestToolPolicyDrift_CommittedSettingsMatchYAML (0.00s)
--- PASS: TestToolPolicyDrift_NoDuplicatesOrOverlap (0.00s)
--- PASS: TestToolPolicyDrift_Mutation (0.01s)   하위 7개 PASS
    settings_side_missing, yaml_side_flip, unmutated, duplicate_settings_allow,
    duplicate_settings_ask, duplicate_settings_deny, allow_deny_overlap
--- PASS: TestToolPolicyDrift_FailClosed (0.01s)   하위 10개 PASS
    missing_yaml, missing_settings, malformed_yaml, malformed_settings_json,
    wrong_type_settings_list, no_permissions_region, empty_yaml_allow,
    empty_yaml_deny, empty_settings_allow, empty_settings_deny
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.193s
```

붉은색 기준선 재현 (감사자 전용 오버레이 테스트. 파일은 스크래치에만 있고 `go test -overlay` 로만 컴파일했다. base YAML 은 `git show d1b61005d:.moai/config/sections/tool-policy.yaml` 로 스크래치에 추출):

```
$ go test -overlay <scratch>/ov_red.json ./internal/config/toolpolicy/ -run 'TestAuditOverlay_BaseYAMLRed$' -count=1 -v
AUDIT total=20 allow-only-in-settings=7 allow-only-in-yaml=1 deny-only-in-yaml=12 deny-only-in-settings=0
AUDIT allow only-in-yaml: MultiEdit
AUDIT allow only-in-settings: CronCreate   (… CronDelete, CronList, EnterPlanMode, EnterWorktree, ExitPlanMode, ExitWorktree)
AUDIT deny only-in-yaml: Glob(./secrets/**)   (… Glob/Grep/Write × 4경로 = 12)
--- PASS: TestAuditOverlay_BaseYAMLRed (0.00s)
```

AC-TDG-002 (생성기는 스크래치 루트에만 썼다. 바이너리는 이 트리에서 `go build -o <scratch>/moai-head ./cmd/moai` → `build_rc=0`):

```
$ <scratch>/moai-head tool-policy build --repo-root <scratch>/root --policy .moai/config/sections/tool-policy.yaml --local-only
  <scratch>/root/.claude/settings.json [json]: allow=114 ask=0 deny=48 env_gated_skipped=5
tool-policy build: regenerated 1 target(s)
gen_rc=0
allow diff_rc=0   (114줄)
deny diff_rc=0    (48줄)
ask 길이: 0 / 0
```

AC-TDG-003:

```
$ git --no-optional-locks diff --exit-code d1b61005d HEAD -- .claude/settings.json internal/template/templates/.claude/settings.json.tmpl; echo "settings_diff_rc=$?"
settings_diff_rc=0
$ git --no-optional-locks status --porcelain --untracked-files=all   (검사 전과 후 동일)
?? .moai/reports/plan-html/SPEC-TOOLPOLICY-DRIFT-GUARD-001-plan.html
```

AC-TDG-007:

```
structurally preventing      → types.go:0, tool-policy.yaml:0
are both generated           → 0
ANALOGOUS drift class        → 0
comment (audit surface) are generated → 0
settings.json.tmpl#permissions → 0
tool-policy-drift-check      → types.go:2, tool-policy.yaml:1
.claude/settings.json#permissions → 165
Package toolpolicy           → 1
```

AC-TDG-008:

```
Makefile:34  build: agents-emit-check commands-emit-check tool-policy-drift-check templ-generate ## Build the binary
make -n build | grep -c TestToolPolicyDrift_                       → 1
make -n build | grep -c TestGoldenCommittedArtifactsMatchEmission  → 2
make -n build 순서: 5행 go test …TestToolPolicyDrift_… / 7행 templ generate / 9행 go build
go_code 블록 줄 수 → 15, 블록 안 - '.claude/settings.json' 존재 (ci.yml:92), - '.moai/**' 존재
CI test job: ci.yml:120 `if: needs.detect.outputs.go_code == 'true'`, ci.yml:210 `go test -json … ./...`
```

AC-TDG-009:

```
$ go test -count=1 -v ./internal/config/toolpolicy/... > <scratch>/ac009-pkg.log 2>&1
pkg_rc=0 / --- PASS: TestRoundTripEquivalence_JSON 1 / --- FAIL: 0
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.392s

$ go test -count=1 ./internal/config/ -run 'TestAuditLoaderCompleteness' -v > <scratch>/ac009-cfg.log 2>&1
cfg_rc=0 / --- PASS: TestAuditLoaderCompleteness 1
ok  	github.com/modu-ai/moai-adk/internal/config	0.456s

$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -timeout 600s ./internal/cli/ -run 'TestToolPolicyList_QueryFilters' -v > <scratch>/cli.log 2>&1
cli_rc=1
    tool_policy_test.go:97: output missing substring "ask";
        got:
        TOOL  ARGS  RISK  DECISION  OWNER  AUDIT
--- FAIL: TestToolPolicyList_QueryFilters (0.00s)
    --- FAIL: TestToolPolicyList_QueryFilters/filter_ask (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_deny (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_allow (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_irreversible (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/all_entries (0.00s)
    --- PASS: TestToolPolicyList_QueryFilters/filter_tool_Bash (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.926s
```

귀속 판독:

```
$ git --no-optional-locks diff --stat d1b61005d HEAD -- internal/cli/; echo "cli_diff_rc=$?"
cli_diff_rc=0                                     (출력 없음)
git show d1b61005d:…/tool-policy.yaml | grep -c 'decision: ask'   → 0
git show HEAD:…/tool-policy.yaml      | grep -c 'decision: ask'   → 0
git show 1ac333952^:…/tool-policy.yaml | grep -c 'decision: ask'  → 6
git merge-base --is-ancestor 1ac333952 d1b61005d                  → ancestor_rc=0
git show --stat 1ac333952:
 .claude/settings.json | 8 - / .moai/config/sections/tool-policy.yaml | 47 ----- /
 .moai/reports/t609/verdict.md | 207 +++ / internal/template/templates/.claude/settings.json.tmpl | 8 -
```

AC-TDG-010:

```
$ <scratch>/moai-head tool-policy list --policy .moai/config/sections/tool-policy.yaml --format json
list_rc=0
MultiEdit 0 / env_gate 없는 네 경로 Glob·Grep·Write deny 0 / env_gate Write deny 1 / env_gate 5
env_gate 정렬 JSON sha256 5e0cba521c5c81a2d7bb82fbba52c59b027d2329b2c0bf8b912ed6b2150e6d27
allow 일곱 도구 ["CronCreate","CronDelete","CronList","EnterPlanMode","EnterWorktree","ExitPlanMode","ExitWorktree"]
Read deny 5 / 전체 항목 167
```

### 2.2 Security

- 적용 권한 불변: §2.1 AC-TDG-003 `settings_diff_rc=0`.
- 검사 읽기 전용: 검사 전후 입력 sha256·`git status` 동일(§2.1).
- 이스케이프·Cf 스캔 (perl, 대조군 포함):

```
.moai/config/sections/tool-policy.yaml escapes=0 cf=0
drift_check_test escapes=0 cf=0
types.go escapes=0 cf=0
Makefile escapes=0 cf=0
ci.yml escapes=0 cf=0
CONTROL escapes=1 cf=1
```

- 시크릿 경로 deny 제거(12개)에 대한 판단: 적용 중인 settings.json 에는 base 시점부터 이미 없었다(§2.1 오버레이 결과의 `deny only-in-yaml` 12개 = YAML 에만 있던 항목). 따라서 실제로 적용되는 보호는 줄지 않았다. 같은 경로의 `Read(...)`/`Edit(...)` deny 는 YAML 에 남아 있다(Read deny `5`).
- 중복 `"permissions"` 키 처리(F2): `settings_region.go:92` `indexOfPermissionsKey` 는 첫 번째 토큰을 반환하고 `extractPermissions` 는 그 영역만 해석한다. 표준 JSON 파서는 같은 키가 두 번이면 마지막 값을 쓴다. 이 카드가 바꾸지 않은 기존 코드다(`verdict.md` §10 에 기록된 `settings_region.go` diff 없음을 이 감사에서도 재확인하지는 않았다 — 갭 4).

### 2.3 Craft

```
$ go test -count=1 -cover ./internal/config/toolpolicy/...
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.222s	coverage: 89.1% of statements
cover_rc=0

$ golangci-lint run --timeout=3m ./internal/config/toolpolicy/...
0 issues.
lint_rc=0
```

독립 뮤턴트 (sed 로 스크래치 사본을 만들고 `go test -overlay` 로 교체 컴파일. 각 사본은 `false &&` 가 정확히 1곳, 원본은 0곳):

```
mut_dup      (if n > 1 → if false && n > 1)
--- FAIL: TestToolPolicyDrift_Mutation/duplicate_settings_allow
--- FAIL: TestToolPolicyDrift_Mutation/duplicate_settings_ask
--- FAIL: TestToolPolicyDrift_Mutation/duplicate_settings_deny
FAIL	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.409s

mut_overlap  (if allow[spec] → if false && allow[spec])
--- FAIL: TestToolPolicyDrift_Mutation/allow_deny_overlap
FAIL	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.403s

mut_empty    (빈 집합 판정 → if false && …)
--- FAIL: TestToolPolicyDrift_FailClosed/empty_yaml_allow
--- FAIL: TestToolPolicyDrift_FailClosed/empty_yaml_deny
--- FAIL: TestToolPolicyDrift_FailClosed/empty_settings_allow
--- FAIL: TestToolPolicyDrift_FailClosed/empty_settings_deny
FAIL	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.437s

mut_type     (json.Unmarshal 오류 → false && err != nil)
--- FAIL: TestToolPolicyDrift_FailClosed/wrong_type_settings_list
FAIL	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.433s
```

네 뮤턴트 모두 붉어졌고, 붉어진 하위 테스트는 해당 규칙을 지키는 것만이었다(겹침 뮤턴트에서 중복 테스트가 붉어지지 않고, 그 반대도 같다). 단정은 `errors.Is` 로 센티널을 확인하며 다른 센티널을 감싸지 않음도 확인한다 — 공허한 통과가 아니다.

### 2.4 Consistency

생성기가 쓰는 표면 판독:

```
internal/config/toolpolicy/codegen.go:121  out.Write(body[:region.start])
internal/config/toolpolicy/codegen.go:122  out.WriteString(`"permissions": `)
internal/config/toolpolicy/codegen.go:123  out.Write(rendered)
internal/config/toolpolicy/codegen.go:124  out.Write(body[region.end:])
internal/config/toolpolicy/codegen.go:151-154  (템플릿 대상 같은 영역 치환)
internal/config/toolpolicy/codegen.go:247  if err := atomicfile.Write(path, out, defs.FilePerm); err != nil {
internal/cli/tool_policy.go:78-79  "the codegen never rewrites the full file, so PATH, hooks, env, …"
```

`codegen.go`·`tool_policy.go` 에서 `header|metadata` 쓰기는 0건. 따라서 정정된 주장 — YAML 머리말은 손으로 관리하고 생성기는 권한 블록만 쓴다(`tool-policy.yaml:6-8`, `types.go:4-7`, `types.go:97-101`), 드리프트는 구조적으로 막히지 않고 검사로 잡는다 — 은 코드 동작과 일치한다. `metadata.generated_into` 에서 템플릿 항목을 뺀 것도 `tool_policy.go:134-149` 가 지시문을 가진 템플릿 영역을 건너뛰는 동작과 맞는다. 머리말의 "runs ahead of `make build` and in the CI Go test job" 도 §2.1 AC-TDG-008 판독(Makefile:34, ci.yml:120·210, `.moai/**`·`.claude/settings.json` 필터)과 맞는다.

---

## 3. 기준선 귀속 (Baseline-attribution)

- 트리: `.claude/worktrees/t619` @ `1ff407dfe47935b34b355a16c4a529d440fe0655`. 감사 중 HEAD 이동 없음(시작 시 `git rev-parse HEAD` 판독, 추적 파일 무수정 — 검사 전후 `git status` 동일).
- 비교 기준: merge-base `d1b61005d`, t609 커밋 `1ac333952`.
- 툴체인: go1.26.4 darwin/arm64, golangci-lint 2.10.1.
- 도구 출처(판정 빌드): AC-TDG-002·010 의 `moai` 는 설치본이 아니라 이 트리(`1ff407dfe`)에서 `go build` 로 스크래치에 만든 바이너리를 경로로 호출했다. 판정 빌드와 트리가 같다.
- progress.md 의 수치는 증거로 옮기지 않았다. 예외는 AC-TDG-004 의 워킹 트리 변이 기록 하나이며, 귀속으로 명시했다.
- 이후 통합 창 재측정은 go1.26.8 로 이뤄질 예정이며 이 보고서의 측정 범위가 아니다.

## 4. 미검증 (Gaps)

1. **평가 프로필 미판독**: `.moai/config/evaluator-profiles/` 와 `harness.yaml` `default_profile` 을 읽지 않고 내장 기본값으로 채점했다. 기본값과 다른 임계값·가중치를 가진 프로필이 활성이라면 점수 표기가 달라질 수 있다.
2. **AC-TDG-004 워킹 트리 변이 미재실행**: 감사자는 추적 파일을 바꿀 수 없어, 워킹 트리 YAML 변이 → make 붉은색 → 복원 사슬은 progress.md 기록에 기댄다. 검사가 워킹 트리를 읽는다는 성질은 `driftCommittedPaths` 코드 판독과 격리 사본 변이로만 간접 확인했다.
3. **AC-TDG-009 base 트리 실행 없음**: `internal/cli` 테스트를 `d1b61005d` 트리에서 직접 돌리지 않았다(워크트리 추가 금지). 귀속은 코드 diff 0·YAML ask 수 동일·조상 판정에 기댄다. develop 트리에서도 실행하지 않았다.
4. `settings_region.go` 가 base 대비 바뀌지 않았는지는 변경 파일 목록(`git diff --stat d1b61005d HEAD`, 11개 파일에 없음)으로만 확인했고 개별 diff 명령은 실행하지 않았다.
5. CI 에서 실제로 새 테스트가 도는지는 관측하지 않았다(push 없음). 경로 필터와 test job 명령 판독만 했다.
6. darwin 외 플랫폼(Windows 경로 구분자 등)에서 `driftCommittedPaths` 동작은 재지 않았다. `filepath` 사용으로 문제 가능성은 낮다.
7. 스크래치 로그(`drift-v.log`, `cli.log`, `ac009-*.log`, 뮤턴트 사본, 오버레이 파일)는 추적 경로로 반출하지 않았다. 판정에 쓴 출력은 이 보고서 §2 에 옮겼고, 원본 로그는 인용하지 않는다.

## 5. 잔여 위험 (Residual-risk)

- **통합 게이트**: AC-TDG-009 의 붉은색은 이 카드와 무관하게 develop 에 이미 있다. t660 이 닫히기 전에 develop 을 push 하면 CI test job 이 붉어진다.
- **중복 키로 드리프트가 가려질 가능성(F2)**: settings.json 에 `"permissions"` 키가 두 번 있으면 검사는 첫 블록을, Claude Code 는 마지막 블록을 읽을 수 있다. 입력을 바꿀 수 있는 주체는 YAML 도 함께 바꿀 수 있는 작성자이므로 권한 우회 경로는 아니지만, 검사의 판정 대상과 실제 적용 대상이 갈라지는 틈이다.
- **로컬 워킹 사본 드리프트**: 검사가 워킹 트리를 읽으므로 로컬에서 settings.json 이 런타임·수동으로 바뀌면 `make build` 가 막힌다. 운영자가 고른 설계(워킹 트리 판독)이며 결함으로 보지 않는다.
- Grep 경로 deny 4개 제거의 무효성 근거는 plan 판정서 §4 갭 1 그대로다(문서의 무시 목록에 Grep 이 없음). 적용 중인 settings.json 에는 이미 없었으므로 이 카드가 보호를 줄이지는 않았다.

---

## 결함 목록

- **F1** [Medium] [optional — 이 카드 판정 기준 / 통합 게이트에서는 blocking, t660 소관] `internal/cli/tool_policy_test.go:97` — `TestToolPolicyList_QueryFilters/filter_ask` 가 커밋 YAML 에 ask 항목이 있다고 가정하는데, t609(`1ac333952`)가 ask 를 0개로 만든 뒤 계속 붉다. 확신도 높음. 필요한 수정: 테스트를 커밋 YAML 이 아닌 ask 항목을 가진 고정 fixture 로 검증하도록 바꾼다(t660). develop push 전에 닫는다.
- **F2** [Low] [optional] `internal/config/toolpolicy/settings_region.go:92` — `indexOfPermissionsKey` 가 첫 `"permissions"` 키를 반환해, 중복 키가 있는 settings.json 에서 검사가 판정하는 블록과 JSON 파서가 적용하는 블록이 다를 수 있다. 기존 코드이며 이 카드의 변경이 아니다. 확신도 중간(중복 최상위 키에 대한 Claude Code 파서 동작은 관측하지 않음). 필요한 수정: 검사 전용 경로에서 최상위 `"permissions"` 토큰 개수가 1이 아니면 해석 실패 센티널로 실패시키는 규칙을 후속 카드로 검토한다.
- **F3** [Low] [optional] `Makefile:68` — 검사 실패 시 안내 문구가 항상 "permissions differ … reconcile tool-policy.yaml" 이다. 실패 원인이 중복·겹침·입력 부재·해석 실패여도 같은 문구가 나온다. 바로 위 `go test` 출력이 실제 원인을 밝히므로 영향은 작다. 필요한 수정: 문구를 "drift check failed — see the test output above for the cause" 형태로 원인 중립적으로 바꾼다.
- **F4** [Info] [optional] `internal/config/toolpolicy/drift_check_test.go` `driftSetDiff` — YAML 을 `driftReadInput` 으로 한 번 읽어 존재만 확인하고 `Load` 로 다시 읽는다. 원인 구별(부재 vs 해석 실패)을 위한 의도로 보이며 동작상 문제는 없다. 필요한 수정: 없음(기록용).

## 권고

- 이 카드는 sync 진행 가능하다. develop 병합·push 순서에서 t660 선행 조건을 리드가 유지한다.
- F2 는 후속 카드 후보로 리드에게 전달된 C1 과 같은 사안이다. 중복 제기하지 않는다.
- F3 은 다음에 이 Makefile 타깃을 만질 때 함께 정리하면 충분하다.

**최종 판정: PASS (90.7) — AC-TDG-009 는 t609 귀속 선재 결함으로 범위 밖 처리, t660 의 develop push 전 종결 조건 유지.**
