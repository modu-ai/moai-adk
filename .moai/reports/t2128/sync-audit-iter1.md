---
verdict: FAIL
audited_sha: 4e7aed62464fc88181a15b675ebc24dba3bcde26
card: t2128
score: 0.87
iteration: 1
---

auditor-model: claude-opus-5-5

# t2128 sync 감사 1회차 — 판정 FAIL (점수 0.87)

- 카드: t2128 (Class B 결함, SPEC 없음), 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2128`, 브랜치 `WT-pr-codex-repair-6`
- 감사 대상: `04d8c27a7e..4e7aed6246` (커밋 `d5e205208c` RED 시험 → `15d6a8648a` 수정 → `4e7aed6246` 진행 기록)
- 평가 프로필: `.moai/config/evaluator-profiles/default.md` (Functionality 40 / Security 25 / Craft 20 / Consistency 15, must-pass = Functionality + Security)
- Codex: 1회 호출(허용 2회), receipt `rcpt-e703dc932cb9971ce8459362`, 판정 pass, 지적 0건

## 판정 요지

수정 자체는 옳다. RED 시험이 수정 전 코드에서 정확히 stack overflow로 죽는 것을 별도 복사본에서 재현했고, 수정 후 순환·공유 앵커·merge key·billion-laughs 형태까지 모두 크래시 없이 처리된다. must-pass 두 차원(Functionality, Security)은 통과다.

FAIL의 원인은 하나다. 새 시험 파일의 `os.Stdout.WriteString` 두 줄이 반환값을 버려서 `errcheck`에 걸린다. 이 저장소의 `.golangci.yml`은 errcheck를 켜 두었고 시험 파일 제외가 없으며, CI의 lint 작업(`golangci-lint run --timeout=5m`)은 필수 검사다. 저장소 전체에서 이런 줄은 이 두 줄뿐이다. 고치는 데는 한 줄씩이면 된다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | RED가 수정 전 코드에서 `fatal error: stack overflow`로 실패(아래 E2), 수정 후 패키지 시험 `ok`(E1), 탐침 13건 중 결함 관련 11건 기대대로, 나머지 2건은 수정 전에도 같은 기존 공백(E5, E6) |
| Security (25%) | 92/100 | PASS | 순환 입력의 프로세스 종료(DoS) 제거, 오류문에 값 미노출(`o/n`, `not-a-real-value` 미포함, E4), billion-laughs 10^12 형태 81.917µs(E5); 키 위치 secret 탐지 공백은 기존 결함(F2) |
| Craft (20%) | 70/100 | FAIL | `golangci-lint run ./internal/config/` 새 파일에서 errcheck 2건(E3); 변경 함수 커버리지 guardSecrets 81.8% / afsHasSecretKey 100% / afsWalkSecret 87.0% |
| Consistency (15%) | 85/100 | PASS | gofmt 출력 없음; 저장소 관례는 `_, _ = fmt.Fprintf(...)`인데 새 파일만 반환값을 버림(F1) |

가중 조화평균: 0.40/95 + 0.25/92 + 0.20/70 + 0.15/85 = 0.011550 → 86.6 → **0.87**. 점수와 별개로 BLOCKING 지적(F1)이 남아 있어 판정은 FAIL이다.

## 지적 목록

- **F1 [BLOCKING] [in-diff]** `internal/config/agentic_factory_settings_cycle_test.go:51`, `:54` — `os.Stdout.WriteString(...)`의 오류 반환을 검사하지 않아 errcheck가 잡는다. CI 필수 lint 작업이 실패한다. 신뢰도 높음(로컬 v2.15.0 관측; CI는 v2.1.6이지만 errcheck 기본 제외 목록에 `(*os.File).WriteString`은 없다).
  - 필요한 수정: 두 줄을 `_, _ = os.Stdout.WriteString(...)`로 바꾼다(`.golangci.yml`이 `check-blank: false`이므로 허용됨). 또는 `fmt.Println("RESULT:...")`로 바꾼다. 수정 뒤 `golangci-lint run ./internal/config/`에서 이 파일 항목이 사라지는지 확인한다.
- **F2 [MINOR] [optional] [touched-file, 기존 결함]** `internal/config/agentic_factory_settings.go:464` — 매핑 키 위치는 `.Value == "secret"`만 본다. 키가 alias(`k: &k secret` 뒤 `{*k : v}`)이거나 복합 키(`? {secret: v}`)이면 secret을 못 찾는다. 수정 전 코드에서도 똑같이 `false`(E6)라 이번 diff가 만든 문제가 아니다. 판정에 영향 없음, 후속 카드 후보로만 기록.
- **F3 [MINOR] [optional] [in-diff]** `internal/config/agentic_factory_settings.go:463-469` — 같은 route 안에서 순환이 secret 키보다 앞에 오면 오류가 cycle로 보고되고 secret은 언급되지 않는다(E4 "secret after cycle in same route"). 어느 쪽이든 거부되고 값은 노출되지 않으므로 수용 가능. 정보성 기록.
- **F4 [MINOR] [optional] [outside]** `webhook.routes`가 참조하지 않는 곳의 순환 앵커(`x-anchors: &c {n: *c}`)는 이 리더가 걷지 않으므로 오류 없이 받아들인다(E5 "unreferenced cycle: err=<nil>"). 같은 파일을 `map[string]any`로 디코드하는 resolver 경로는 yaml.v3가 `anchor 'c' value contains itself` 오류를 낸다(크래시 아님). 리더 간 판단이 갈리지만 크래시는 없음. 판정 영향 없음.

## 점검 항목별 결론

1. **RED 선행·실패 이유**: `d5e205208c`는 시험만 추가하고 소스는 수정 전 그대로다(`afsHasSecretKey(n.Alias)` 1건 확인). 그 커밋을 scratch에 풀어 돌리니 `TestAgenticFactorySettings_RefCycle`이 `runtime: goroutine stack exceeds 1000000000-byte limit` / `fatal error: stack overflow`로 실패했다(E2). 진술된 이유와 일치.
2. **순환 알고리즘**: onPath(회색)·clean(검정) 두 집합을 쓰는 표준 DFS 순환 탐지다. 직접 순환, 간접 순환, 시퀀스 경유 순환, clean 처리된 공유 앵커 뒤의 순환 경로, 두 번 쓰인 공유 앵커 안 secret, merge key(`<<: *m` 자기 순환, `<<: *b` / `<<: [*a, *b]`로 들어온 secret), diamond(순환 아님) 모두 기대대로 나왔다(E5). clean 메모 덕에 각 노드를 한 번만 걸어 billion-laughs 형태가 선형이다. 정상 설정의 오탐은 없었다(공유 앵커·diamond·`&secret` 이름의 앵커 모두 `false,false`).
3. **다른 재귀 경로**: `afsLookup`은 경로 세그먼트 수만큼만 돌고 alias를 세그먼트당 한 번 따라간다. `afsChild` 비재귀, `routes()`는 `repository` 한 키만 읽는다, `Decode`는 스칼라에만 호출된다. 파일 로드는 `yaml.Node`로 Unmarshal하므로 재귀 디코드가 없다. 같은 파일을 읽는 resolver(`loadYAMLSections` → `map[string]any`)는 yaml.v3 디코더의 alias 자기포함 검사로 오류를 낸다(E5 마지막 줄). `LoadAgenticFactorySettings` / `AgenticFactoryConfigured`의 비시험 호출자는 아직 없다(grep 결과 없음).
4. **서브프로세스 시험 위생**: 60초 `exec.CommandContext`로 묶이고, 자식은 환경 변수가 없으면 Skip하며, 자식 실행은 `^...Child$`로 고정되어 부모의 `-run`/`-count`와 무관하다(`-count=3` 실행 `ok`). `CombinedOutput`이 끝까지 기다리고 타임아웃 시 컨텍스트가 프로세스를 죽이므로 누수 없음. `os.Args[0]` 재실행은 Windows 시험 바이너리(.exe)에서도 쓰는 일반 기법이지만 Windows에서 실제 실행은 하지 않았다(Gaps).
5. **레인 로컬 검증**: 아래 E1, E3에 원문 기록.

## Claim / Evidence / Baseline / Gaps / Residual-risk

### Claim

- C1: 수정 후 `internal/config` 패키지 시험이 통과한다.
- C2: RED 시험은 수정 전 코드에서 stack overflow로 실패한다.
- C3: 수정은 순환을 탐지하되 정상 공유 앵커는 거부하지 않고, 공유 앵커 안 secret은 계속 거부한다.
- C4: 새 시험 파일이 CI 필수 lint의 errcheck에 걸린다(F1).
- C5: Windows 교차 vet과 gofmt는 깨끗하다.

### Evidence

**E1** — 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2128`, HEAD `4e7aed62464fc88181a15b675ebc24dba3bcde26`, `GOCACHE=<scratchpad>/gocache-lane28`, go1.27.3 darwin/arm64

```
$ go test -count=1 ./internal/config/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/config	17.620s

$ GOOS=windows GOARCH=amd64 go vet ./internal/config/
vet-windows exit=0
$ go vet ./internal/config/
vet-native exit=0
$ gofmt -l internal/config/agentic_factory_settings.go internal/config/agentic_factory_settings_cycle_test.go
gofmt exit=0          (출력 없음)

$ go test -count=1 -v -run '^TestAgenticFactorySettings_RefCycle$' ./internal/config/
=== RUN   TestAgenticFactorySettings_RefCycle
--- PASS: TestAgenticFactorySettings_RefCycle (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.250s

$ go test -count=1 -v -run '^TestAfsHasSecretKey_CycleAndSharing$' ./internal/config/
--- PASS: TestAfsHasSecretKey_CycleAndSharing (0.00s)
    --- PASS: .../self_reference_in_a_map, self_reference_in_a_list, indirect_cycle,
              shared_anchor,_no_cycle, secret_through_a_shared_anchor, plain_secret, no_secret
ok  	github.com/modu-ai/moai-adk/internal/config	0.222s

$ go test -count=3 -run '^TestAgenticFactorySettings_' ./internal/config/
ok  	github.com/modu-ai/moai-adk/internal/config	6.704s

$ go test -count=1 -coverprofile=<scratchpad>/t2128-cover.out -run 'AgenticFactory' ./internal/config/   (exit=0)
$ go tool cover -func=<scratchpad>/t2128-cover.out | grep -E "guardSecrets|afsHasSecretKey|afsWalkSecret"
.../agentic_factory_settings.go:414:		guardSecrets				81.8%
.../agentic_factory_settings.go:438:		afsHasSecretKey				100.0%
.../agentic_factory_settings.go:444:		afsWalkSecret				87.0%
```

**E2** — RED 커밋 `d5e205208c`를 `git -C <tree> archive`로 scratch(`<scratchpad>/t2128-red`)에 풀어 실행(작업 트리 무변경)

```
$ GOFLAGS=-mod=mod go test -count=1 -run '^TestAgenticFactorySettings_' ./internal/config/
exit=1
--- FAIL: TestAgenticFactorySettings_RefCycle (8.90s)
    agentic_factory_settings_cycle_test.go:78: loading a self-referencing alias crashed the process (exit status 2):
        runtime: goroutine stack exceeds 1000000000-byte limit
        fatal error: stack overflow
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/config	10.830s
```

**E3** — golangci-lint v2.15.0 (CI는 v2.1.6), `.golangci.yml`: errcheck 활성, `check-blank: false`, 사용자 제외 없음

```
$ golangci-lint run ./internal/config/
lint-exit=1
internal/config/agentic_factory_settings_cycle_test.go:51:24: Error return value of `os.Stdout.WriteString` is not checked (errcheck)
internal/config/agentic_factory_settings_cycle_test.go:54:23: Error return value of `os.Stdout.WriteString` is not checked (errcheck)
internal/config/audit_struct_yaml_symmetry_test.go:112:18: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/harness_ceiling_test.go:61:18: inline: Constant reflect.Ptr should be inlined (govet)
internal/config/shipped_key_reader_test.go:369:17: inline: Constant reflect.Ptr should be inlined (govet)
... (shipped_key_reader_test.go 4건 더, 모두 govet inline)
9 issues:
* errcheck: 2
* govet: 7

$ grep -rn "^\s*os\.Std\(out\|err\)\.WriteString(" --include='*.go' .
internal/config/agentic_factory_settings_cycle_test.go:51:		os.Stdout.WriteString("RESULT:nil\n")
internal/config/agentic_factory_settings_cycle_test.go:54:	os.Stdout.WriteString("RESULT:" + err.Error() + "\n")

$ grep -n "golangci" .github/workflows/ci.yml
832:  # Lint: Static analysis with golangci-lint (required check; ...
865:        run: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
883:        run: golangci-lint run --timeout=5m
```

**E4** — HEAD 사본(`<scratchpad>/t2128-head`)에 탐침 시험 파일을 더해 실행, `LoadAgenticFactorySettings` 전체 경로

```
err=validation error: field "workflow.agentic_factory.webhook.routes[0]": a YAML alias refers back to itself (a cycle), which this file does not allow (the value was not echoed) in .../agentic-factory.yaml   (routes 항목이 앞 route의 순환 맵 alias)
err=validation error: field "workflow.agentic_factory.webhook.routes[1]": a YAML alias refers back to itself (a cycle), ...   (둘째 route에만 순환)
err=validation error: field "workflow.agentic_factory.webhook.routes[0]": a YAML alias refers back to itself (a cycle), ...   (merge key 순환 `<<: *c`)
err=validation error: field "workflow.agentic_factory.webhook.routes[0]": a YAML alias refers back to itself (a cycle), ...   (순환 뒤 secret: not-a-real-value)
routes is self alias: err=validation error: field "workflow.agentic_factory.webhook.routes[0]": a YAML alias refers back to itself (a cycle), ...   (routes: &r [*r])
```

네 건 모두 오류문에 `o/n`, `not-a-real-value`가 들어 있지 않음(탐침의 `UNEXPECTED` 단언 미발동).

**E5** — 같은 탐침, `afsHasSecretKey` 직접 호출 (got → want)

```
merge key self cycle                          found=false cycle=true   (want false true)
merge key brings secret                       found=true  cycle=false  (want true false)
merge list brings secret                      found=true  cycle=false  (want true false)
cycle via sequence                            found=false cycle=true   (want false true)
indirect cycle via sequence                   found=false cycle=true   (want false true)
shared clean then cyclic path                 found=false cycle=true   (want false true)
cycle reached through a shared clean node     found=false cycle=true   (want false true)
secret after cycle point                      found=false cycle=true   (want false true)
secret before cycle point                     found=true  cycle=false  (want true false)
alias key to scalar secret (pre-existing gap) found=false cycle=false  (want true false)  MISMATCH
complex key holding secret (pre-existing gap) found=false cycle=false  (want true false)  MISMATCH
diamond, no cycle                             found=false cycle=false  (want false false)
anchor named secret used as value             found=false cycle=false  (want false false)
laughs 10^12: found=false cycle=false in 81.917µs
unreferenced cycle: err=<nil>
map decode of cycle: err=yaml: anchor 'c' value contains itself
```

(위 표는 탐침 출력 줄 `zz_t2128_probe_test.go:44: got found=... cycle=... (want ...)`을 케이스 이름 순서대로 정렬한 것. 두 MISMATCH는 기존 공백 확인용으로 일부러 `true`를 기대한 탐침이다.)

**E6** — RED 사본(수정 전 함수)에서 같은 키 위치 입력

```
base afsHasSecretKey("k: &k secret\nr: {*k : v}\n") = false
base afsHasSecretKey("? {secret: v}\n: x\n") = false
ok  	github.com/modu-ai/moai-adk/internal/config	0.360s
```

**E7** — Codex (`mcp__moai__codex_audit`, mode native, target baseBranch, delta_base `04d8c27a7ed77afc09217f8623519806d48df3c4`, `git merge-base --is-ancestor` 확인 후 호출, 호출 중 트리 clean)

```
"verdict":"pass"   (Codex 결과 원문의 필드)
summary: 기준 커밋 대비 변경을 정적 검토했으나 새로 도입된 결함은 찾지 못했어. ... 테스트·빌드·재현 명령은 실행하지 않았어.
findings: []
audit_receipt: rcpt-e703dc932cb9971ce8459362
resolved_model: gpt-6.1-sol
review_base: delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)
build_lag: the installed binary was built from commit 04d8c27a7, an ancestor of this tree's HEAD 4e7aed624.
```

**E8** — 감사 계획

```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2128
"config_status": "ok", "model": "codex", claude gate off, codex gate required, glm gate off,
"cross_model_active": true, "enforced_required": ["codex"]
```

### Baseline-attribution

- 모든 시험·vet·lint는 이 실행에서 트리 HEAD `4e7aed62464fc88181a15b675ebc24dba3bcde26`(clean, `git status --short` 0줄)에서 직접 돌렸다.
- RED 재현(E2)과 기존 공백 확인(E6)은 커밋 `d5e205208c`의 git archive 사본, 탐침(E4, E5)은 커밋 `4e7aed6246`의 git archive 사본에서 돌렸다. 작업 트리에는 이 판정 파일 말고 아무것도 쓰지 않았다.
- `go` 계열 명령은 트리 소스를 직접 컴파일하므로 도구 빌드 지연과 무관하다. `moai verify audit-plan`과 MCP 서버는 설치 바이너리(커밋 04d8c27a7e, 이 트리 HEAD의 조상)로 돌았다. 이 카드는 moai CLI 동작을 바꾸지 않으므로 계획 판독에 영향이 없다고 보지만, 기록상 지연이 있다.

### Gaps

- CI가 쓰는 golangci-lint v2.1.6으로는 돌리지 않았다(로컬 v2.15.0만). F1이 CI에서도 걸린다는 판단은 errcheck 기본 제외 목록에 대한 지식에 기댄 것이다. govet `inline` 7건이 v2.1.6에서도 나오는지는 확인하지 않았다(기준선 파일들이며 이번 diff와 무관).
- Windows에서 서브프로세스 시험을 실제로 실행하지 않았다(교차 vet만).
- `-race`, 전체 스위트, 다른 패키지는 실행하지 않았다(지시 범위).
- 코드 리뷰 도구 규약상 `audit_multi`는 쓰지 않았다(리더 상시 규칙 `d-20261106T080039Z-305c`: codex_audit 단독, 단계당 2회 이하). Codex 호출은 1회.
- 감사 중 거부된 명령 2건(작업 트리 격리 가드): 변수가 든 복합 셸 명령과 `git archive`를 `cd ... &&` 뒤에 붙인 형태. 둘 다 단순 명령으로 나눠 다시 돌렸고 결과는 위에 그대로 있다. 측정 대체는 없었다.

### Residual-risk

- F2의 키 위치 공백은 남아 있다. 이 리더가 지금은 `repository`만 소비하므로 정책 우회에 그치지만, 후속 카드가 route 맵을 통째로 디코드하게 되면 secret이 alias 키로 들어올 수 있다.
- 순환이 아닌 아주 깊은 중첩은 여전히 재귀 깊이만큼 스택을 쓴다(Go 1GB 한도). 진행 기록의 Gaps에도 적혀 있다.
- scratch에 남은 시험 출력(`<scratchpad>/t2128-config-test.txt`, `t2128-lint.txt`, `t2128-probe.txt`, `t2128-red-run.txt`)은 기기 로컬이며 인용 대상이 아니다. 판정에 쓴 줄은 모두 위에 옮겨 적었다.

## 반복 이력

| 회차 | 감사 SHA | 판정 | 점수 | 요지 |
|---|---|---|---|---|
| 1 | 4e7aed62464fc88181a15b675ebc24dba3bcde26 | FAIL | 0.87 | F1 errcheck 2건(BLOCKING) 외 수정 정확 |

재감사 범위 제안: F1 수정 커밋의 diff와 `golangci-lint run ./internal/config/`, `go test -count=1 ./internal/config/` 결과만 보면 된다.
