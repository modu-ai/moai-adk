# t1210 감사 — tool-policy.yaml 원본 층 `\:` 가드

- 대상: 커밋 `d5989baeb` (기준 `baa054586`), 브랜치 `WT-policy-escape-guard`, 워크트리 `.claude/worktrees/t1210`
- 변경 범위: `internal/config/toolpolicy/escape_guard_test.go`(+113, 테스트 전용), `.moai/reports/t1210/verdict.md`(+47)
- 감사자: sync-auditor (read-only, 추적 파일 무수정. 변이는 모두 `go test -overlay`와 스크래치 사본으로만 수행)
- 감사 시점 HEAD: `d5989baeb`, `git status --short` 0줄

## 판정

**PASS-WITH-DEBT** — 조화평균 **92.4**

| 차원 | 점수 | 판정 | 근거 요약 |
|---|---|---|---|
| Functionality | 95 | PASS | 레인 주장 3건 모두 재현. 가드가 t1207 이전 원본(`C\:` 5건)에서 실제로 빨간불이 됨을 오버레이로 확인 |
| Security | 95 | PASS | 테스트 전용 변경. 거부 규칙 무력화 경로 하나를 원본 층에서 막음. 새 공격 표면 없음 |
| Craft | 88 | PASS | vet·gofmt·golangci-lint 무결. 변이 4종 모두 검출. 재생성 경로 충실도·고정 철자 의존 등 선택 항목 3건 |
| Consistency | 92 | PASS | 템플릿 쪽 가드와 같은 술어(Bash + `\:`), 기존 drift 헬퍼 재사용, 커밋 메시지에 카드 id |

조화평균: 4 / (1/95 + 1/95 + 1/88 + 1/92) = 92.4. 반드시 통과해야 하는 차원(Functionality·Security)은 각각 기준을 넘는다. blocking 항목은 없다. 부채는 F1(빌드 시점 미실행)과 F2(Bash 외 도구 미검사)이며, 둘 다 레인이 verdict.md에 스스로 기록해 두었다.

## 주장별 검증

### 주장 1 — 공백 재현(`DriftCheckAloneMissesIt`)은 실재하며 공허하지 않다

**Evidence**
```
$ go test ./internal/config/toolpolicy -count=1 -run TestToolPolicyEscapeGuard -v
=== RUN   TestToolPolicyEscapeGuard_CommittedPolicy
--- PASS: TestToolPolicyEscapeGuard_CommittedPolicy (0.00s)
=== RUN   TestToolPolicyEscapeGuard_DetectsMutation
--- PASS: TestToolPolicyEscapeGuard_DetectsMutation (0.01s)
=== RUN   TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt
--- PASS: TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.115s
```

공허성 반증 — 변이 M5(변이 단계를 무효화: `clean`을 `clean`으로 치환):
```
== M5 mutation step is a no-op
--- PASS: TestToolPolicyEscapeGuard_CommittedPolicy (0.00s)
    escape_guard_test.go:89: guard on the mutated policy = [], want exactly ["Bash(rm -rf C\\:/:*)"]
--- FAIL: TestToolPolicyEscapeGuard_DetectsMutation (0.06s)
    escape_guard_test.go:111: regenerated settings do not carry "Bash(rm -rf C\\:/:*)"
--- FAIL: TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt (0.01s)
```
재생성 단계가 결정적임을 보이는 변이 M6(설정 사본을 재생성하지 않고 커밋본 그대로 둠):
```
== M6 settings not regenerated
    escape_guard_test.go:104: drift check reported ["deny only-in-yaml: Bash(rm -rf C\\:/:*)" "deny only-in-settings: Bash(rm -rf C:/:*)"]; the reproduction expects set-equal inputs
--- FAIL: TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt (0.02s)
```

해석: 공백은 「원본에 `\:`를 넣고 **설정까지 그 원본으로 다시 만든 경우**」에 한정된다. 원본만 바꾸면 drift 검사가 잡는다(M6). 레인의 서술(「로컬 설정을 그 원본으로 다시 만들면」)은 이 범위와 정확히 일치한다. `clean` 패턴이 없으면 `t.Fatalf`로 크게 실패하고(24–28행), 재생성된 설정에 이스케이프 규칙이 실제로 들어 있는지도 확인한다(110행). JSON 확인식 `strings.ReplaceAll(escaped, `\`, `\\`)`는 JSON 인코더가 역슬래시 하나를 `\\`로 쓰는 것과 맞는다 — M5에서 이 확인이 실패 경로로 발화했으므로 검사가 살아 있다.

### 주장 2 — 가드는 커밋된 원본에서 0건, 변이 사본에서 정확히 1건을 보고한다

**Evidence** — 가드 술어를 그대로 복사한 스크래치 프로브(`escape_guard_test.go:58-71`과 동일, `t.Fatalf`만 exit로 대체)를 t1207 이전 원본과 철자 변형에 돌렸다:
```
$ go run . ../pre-t1207-tool-policy.yaml ../current.yaml ../v-double.yaml ../v-single.yaml ../v-plain.yaml ../v-envgated.yaml ../v-read.yaml
pre-t1207-tool-policy.yaml -> bash-guard 5 ["Bash(rm -rf C\\:/:*)" "Bash(rm -rf C\\:/\\* *)" "Bash(del /S /Q C\\:/:*)" "Bash(rmdir /S /Q C\\:/:*)" "Bash(Remove-Item -Recurse -Force C\\:/:*)"] | any-tool 5
current.yaml -> bash-guard 0 [] | any-tool 0
v-double.yaml -> bash-guard 1 ["Bash(rm -rf C\\:/:*)"] | any-tool 1
v-single.yaml -> bash-guard 1 ["Bash(rm -rf C\\:/:*)"] | any-tool 1
v-plain.yaml -> bash-guard 1 ["Bash(rm -rf C\\:/:*)"] | any-tool 1
v-envgated.yaml -> bash-guard 1 ["Bash(rm -rf D\\:/:*)"] | any-tool 1
v-read.yaml -> bash-guard 0 [] | any-tool 1
$ go run . ../v-double-invalid-escape.yaml
LOAD ERROR: tool-policy parse "../v-double-invalid-escape.yaml": yaml: line 160: found unknown escape character
```
(`%q` 출력이므로 `\\`는 실제 역슬래시 하나다.)

실제 패키지 안의 가드로도 확인 — 변이 M4(`CommittedPolicy`가 t1207 이전 원본을 읽도록 오버레이):
```
== M4 committed-policy test pointed at pre-t1207 yaml
    escape_guard_test.go:79: tool-policy.yaml Bash rules escape ':' and cannot match a real drive path: ["Bash(rm -rf C\\:/:*)" "Bash(rm -rf C\\:/\\* *)" "Bash(del /S /Q C\\:/:*)" "Bash(rmdir /S /Q C\\:/:*)" "Bash(Remove-Item -Recurse -Force C\\:/:*)"]
--- FAIL: TestToolPolicyEscapeGuard_CommittedPolicy (0.01s)
```
가드를 눈멀게 한 변이 M1(술어 `\:` → `\;`):
```
== M1 blind predicate
--- PASS: TestToolPolicyEscapeGuard_CommittedPolicy (0.00s)
    escape_guard_test.go:89: guard on the mutated policy = [], want exactly ["Bash(rm -rf C\\:/:*)"]
--- FAIL: TestToolPolicyEscapeGuard_DetectsMutation (0.01s)
```

해석: 가드는 YAML 디코딩 후의 값을 보므로 큰따옴표(`"C\\:"`)·작은따옴표(`'C\:'`)·일반 스칼라(`C\:`) 세 철자를 모두 잡는다. 큰따옴표 안의 `"C\:"`는 YAML 파싱 오류가 되어 `Load` 단계에서 이미 실패한다. env-gated 항목도 잡는다. `CommittedPolicy` 단독으로는 M1에서 공허하게 통과하지만, 같은 함수를 쓰는 `DetectsMutation`이 대신 빨간불이 되므로 쌍으로는 공허하지 않다. 레인 verdict.md의 Gaps 3번째 항목(「과거 상태 원본으로 가드를 돌리지 않았다」)은 이 감사의 M4로 닫혔다.

### 주장 3 — 테스트는 작업 트리를 건드리지 않는다

**Evidence**
```
$ git status --short | wc -l      # M1·M4 실행 직후
       0
$ git status --short | wc -l      # M5·M6 실행 직후
       0
```
코드 판독: 모든 쓰기는 `driftWriteFixture`를 거치고, 그 디렉터리는 `driftCopyCommitted`의 `t.TempDir()`(drift_check_test.go:326-338) 또는 그 파일의 `filepath.Dir`이다. 작업 트리의 원본·설정은 `os.ReadFile`로만 읽는다.

### 정적 검사

```
$ go vet ./internal/config/toolpolicy/ ; echo "vet exit=$?"
vet exit=0
$ gofmt -l internal/config/toolpolicy/; echo "gofmt exit=$?"
gofmt exit=0
$ golangci-lint run ./internal/config/toolpolicy/
0 issues.
$ go test ./internal/config/toolpolicy -count=1
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.257s
```

### Makefile 배선 제외가 정직하게 기록됐는가

```
$ grep -n -A8 '^tool-policy-drift-check' Makefile
66:tool-policy-drift-check: ## Verify tool-policy.yaml and the .claude/settings.json permissions block declare the same sets (read-only; never regenerates)
67-	@go test ./internal/config/toolpolicy/... -run 'TestToolPolicyDrift_(CommittedSettingsMatchYAML|NoDuplicatesOrOverlap)$$' -count=1 \
$ grep -rn 'go test' .github/workflows/*.y*ml | ...
.github/workflows/ci.yml:221:          go test -json -coverprofile=coverage.out -covermode=atomic ./... > test-stream.json || rc=$?
.github/workflows/ci.yml:303:          go test -json -race -count=1 -timeout 20m ./... > test-stream.json || rc=$?
```
`make build` 앞의 drift 검사는 정규식이 두 테스트 이름에 고정돼 있어 `TestToolPolicyEscapeGuard_*`를 돌리지 않는다. CI의 `./...` 실행에는 포함된다. 레인의 Gaps 서술과 일치한다.

## Findings

- **F1 [Low] [optional]** `Makefile:67` — 새 가드는 `make build` 시점에 돌지 않는다. 로컬에서 원본과 설정을 함께 고친 뒤 빌드하면 CI 전까지 신호가 없다. 레인이 Gaps에 정직하게 기록했고 카드 범위(테스트 추가) 밖이다. 신뢰도 높음. 필요 조치: 원하면 후속 카드에서 `-run` 정규식을 `TestToolPolicy(Drift_(CommittedSettingsMatchYAML|NoDuplicatesOrOverlap)|EscapeGuard_CommittedPolicy)$$`로 넓힌다.
- **F2 [Low] [optional]** `escape_guard_test.go:66` — 술어가 `e.Tool == "Bash"`로 한정돼 Bash 외 도구 규칙의 `\:`는 보지 않는다(프로브 `v-read.yaml -> bash-guard 0 [] | any-tool 1`). 현재 원본에는 `\:`가 어느 도구에도 0건이고, 템플릿 쪽 가드도 Bash만 보므로 두 층의 범위는 일치한다. Read/Edit 경로 규칙에서 `\:`가 실제로 무력화를 일으키는지는 이 감사에서 측정하지 않았다(가설). 레인이 PowerShell 사례로 Residual-risk에 기록했다. 신뢰도 중간. 필요 조치: 없음. PowerShell 등 새 도구 규칙이 원본에 들어올 때 술어를 넓힌다.
- **F3 [Info] [optional]** `escape_guard_test.go:40-50` — 설정 재생성을 `json.Encoder`로 손수 조립한다. 실제 빌드 경로는 `BuildInto` → `BuildPermissions` → `RenderSettingsJSON`(codegen.go:111, 207-226)이다. 집합 비교 결과는 두 경로 모두 `BuildPermissions`를 거치므로 같지만, `RenderSettingsJSON`(또는 임시 사본에 `BuildInto`)을 쓰면 재현이 실제 `moai tool-policy build --local-only`에 한 단계 더 가까워진다. 신뢰도 높음. 필요 조치: 선택.
- **F4 [Info] [optional]** `escape_guard_test.go:96-113` — `DriftCheckAloneMissesIt`는 「drift 검사에 공백이 있다」는 사실을 통과 조건으로 고정한다. 나중에 누군가 `driftSetDiff`에 `\:` 검사를 넣어 공백을 닫으면 이 테스트가 「reproduction expects set-equal inputs」로 실패한다. 테스트 이름과 주석이 의도를 밝히고 있어 오독 위험은 낮다. 신뢰도 높음. 필요 조치: 선택 — 주석에 「drift 검사가 이 형태를 잡게 되면 이 테스트를 지울 것」 한 줄을 덧붙일 수 있다.
- **F5 [Info] [optional]** `escape_guard_test.go:25` — 변이 대상이 큰따옴표 철자 `args_pattern: "rm -rf C:/:*"`에 고정돼 있다. 원본의 따옴표 스타일이 바뀌면 `t.Fatalf`로 크게 실패하므로 공허 통과는 없다(취성일 뿐). 필요 조치: 없음.

## Baseline-attribution

모든 결과는 이번 실행에서 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1210`(HEAD `d5989baeb`)에 대해 관측했다. t1207 이전 원본은 `git show 9d88a99e9:.moai/config/sections/tool-policy.yaml`로 스크래치에 내보냈다. 변이 사본·오버레이 JSON·프로브는 `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/2cdcb7ed-5896-4d85-a8b7-9a19fb627907/scratchpad/t1210-audit/` 아래에만 있다.

## Gaps

- CI(`./...`, `-race`)에서 이 커밋이 실제로 돈 결과는 보지 않았다. 레인은 push하지 않으므로 아직 CI 실행이 없다.
- `moai tool-policy build --local-only` 바이너리 경로를 끝까지 실행하지는 않았다(F3). 재현은 `BuildPermissions` 층에서 확인했다.
- Claude Code가 Bash 외 도구의 경로 규칙에서 `\:`를 어떻게 해석하는지는 측정하지 않았다(F2).
- darwin 단일 환경. windows 매트릭스는 보지 않았다.

## Residual-risk

- 가드는 `\:` 한 형태만 본다. 다른 역슬래시 조합이 같은 방식으로 규칙을 무력화할 가능성은 조사 범위 밖이다(레인도 기록).
- 로컬 빌드 시점에는 가드가 돌지 않으므로(F1), 원본과 설정을 함께 잘못 고친 상태가 CI 전까지 로컬 develop에 머무를 수 있다.
