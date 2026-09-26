# t1210 — tool-policy.yaml 쪽 `\:` 가드 추가

## Claim

1. t1207이 넣은 `\:` 가드(`internal/template/settings_test.go`)는 렌더된 템플릿만 검사했다. 원본인 `.moai/config/sections/tool-policy.yaml`에 `\:`가 든 Bash 규칙이 들어오고 로컬 설정을 그 원본으로 다시 만들면, 드리프트 검사는 두 파일의 집합이 같다는 이유로 통과시킨다(감사 F1). 이 공백을 테스트로 재현했다.
2. 원본 층에 같은 가드를 추가했다(`internal/config/toolpolicy/escape_guard_test.go`). 가드는 현재 커밋된 원본에 `\:`가 없음을 확인하고, 드리프트 검사가 놓친 바로 그 변이를 잡아낸다.

## Evidence

**공백 재현** — `TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt`는 커밋된 원본과 설정을 임시 디렉터리에 복사한다. 원본의 `rm -rf C:/:*` 한 줄을 `rm -rf C\:/:*`로 바꾸고, `Load` → `BuildPermissions`로 설정 사본을 다시 만든다. 그 결과 `driftSetDiff`는 차이 0건을 보고했고, 재생성된 설정에는 `Bash(rm -rf C\:/:*)`가 실제로 들어 있었다.

```
=== RUN   TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt
--- PASS: TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt (0.01s)
```

(이 테스트의 PASS는 「드리프트 검사가 변이를 통과시킨다」는 공백이 실재한다는 뜻이다.)

**가드** — `escapedColonSpecifiers`는 원본의 모든 항목을 본다. 환경변수로 켜지는 항목(env-gated)도 포함하며, Bash 규칙 중 `args_pattern`에 `\:`가 든 것을 모은다.

```
go test ./internal/config/toolpolicy -run 'TestToolPolicyEscapeGuard' -count=1 -v
--- PASS: TestToolPolicyEscapeGuard_CommittedPolicy (0.00s)
--- PASS: TestToolPolicyEscapeGuard_DetectsMutation (0.01s)
--- PASS: TestToolPolicyEscapeGuard_DriftCheckAloneMissesIt (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	0.113s
```

`DetectsMutation`은 같은 변이 입력에서 가드가 정확히 `[Bash(rm -rf C\:/:*)]` 하나를 보고하는지 확인한다.

**패키지 전체** — `go test ./internal/config/toolpolicy -count=1` → `ok … 0.368s`. `go vet`와 `gofmt -l`은 출력 없음, `golangci-lint run ./internal/config/toolpolicy/` → `0 issues.`

## Baseline-attribution

작업트리 `.claude/worktrees/t1210`, 브랜치 `WT-policy-escape-guard`, 기준 로컬 develop `baa054586`. 위 결과는 모두 이 트리에서 이번 실행으로 관측했다.

## Gaps

- 가드는 `go test`로만 돈다. `make build` 앞의 `tool-policy-drift-check`는 드리프트 테스트 2건만 골라 돌리므로, 로컬 빌드 시점에는 이 가드가 실행되지 않는다. CI의 전체 테스트에서 잡힌다. 템플릿 쪽 가드도 같은 방식이다. Makefile은 카드 범위 밖이라 고치지 않았다.
- 제품 코드(`Load`, `BuildPermissions`)에서 `\:` 입력을 거부하는 방식은 택하지 않았다. `moai tool-policy build` 동작이 바뀌기 때문이다. 가드는 테스트 층에만 있다.
- 커밋된 원본 파일 자체를 과거 상태(t1207 이전, `C\:` 5건)로 되돌려 가드를 돌리지는 않았다. 변이 테스트가 같은 입력 모양을 대신 검증한다.
- 제품 코드를 바꾸지 않았으므로 `make build`는 다시 돌리지 않았다.

## Residual-risk

- 가드는 `\:` 한 가지 형태만 본다. 다른 이스케이프(예: `\*` 외의 역슬래시 조합)가 같은 방식으로 규칙을 무력화할 수 있는지는 조사하지 않았다.
- 원본 층 가드는 Bash 도구 규칙만 본다. 앞으로 `PowerShell(...)` 규칙이 원본에 들어오면 같은 가드가 필요할 수 있다.
