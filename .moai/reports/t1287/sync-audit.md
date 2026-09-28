# t1287 독립 감사와 통합 재측정

## Claim

**PASS — AC-SMN-001~004.** 원 담당자의 `73d900b97` 구현은 선언 모델의 `[1m]` 접미와 `inherit`를 요구대로 분류한다. 통합 브랜치의 cherry-pick `13481c8ba`에서도 관련 테스트와 로컬 프로필 스윕이 통과했다.

## Evidence

독립 감사자는 원본 `WT-served-model-norm`의 깨끗한 `73d900b97`에서 다음을 측정했다.

```text
$ go test ./internal/hook -run '^TestServedModel_(Classify|UnknownNeverOK)$' -count=1 -v
PASS
$ go test ./internal/cli -run 'TestRunDiagnosticChecks|TestServedModelCheck_|TestBinaryLag_' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli  7.982s
$ go mod verify
all modules verified
$ go vet ./internal/hook
(출력 없음, exit 0)
```

통합 브랜치에서 다시 잰 hook 테스트 `ok .../internal/hook 0.609s`, CLI 테스트 `ok .../internal/cli 8.318s`, `go vet ./internal/hook` exit 0은 `progress.md` §E.3에 기록했다. 이 트리로 빌드한 `/tmp/moai-t1287-doctor`의 프로필 스윕은 `swept 2363 subagent transcripts: ok 1599, served_drift 706, unknown 11, unmapped 47`이었다. 출력의 `served_drift expected=` 상세 706행에서 `expected=…[1m]` 또는 `expected=inherit`는 0행이었다. 대조 케이스 `opus[1m]` 대 `glm-*`는 테스트에서 여전히 drift다.

## Baseline-attribution

독립 감사의 코드는 원본 `73d900b97`, 통합 재측정의 코드는 이 브랜치의 `13481c8ba`에 후속 증거 커밋을 얹은 트리다. 스윕은 `CLAUDE_CONFIG_DIR=/Users/goos/.moai/claude-profiles/moai-adk`의 2026-09-28 로컬 기록을 대상으로 했다. 과거 카드의 292건 수치를 이번 판의 기준으로 쓰지 않았다.

## Gaps

패키지 전체 커버리지와 `golangci-lint`는 독립 감사에서 완료되지 않았다. 프로필 스윕은 현재 로컬 기록만 보며, 모든 가능한 모델 ID의 의미를 증명하지 않는다. 이 통합 커밋의 원격 CI 판정은 아직 없다.

## Residual-risk

독립 감사자가 임시 overlay 테스트로 `servedMatches("opus", "claude-sonnet-opus-5")`가 참이 되는 기존 별칭 비교를 확인했다. 현재 유효 모델 ID에서 그 형태는 관측되지 않아 이 카드의 AC 차단 사유로 삼지 않았으나, 새 모델 ID가 그 형태를 쓰면 오분류할 수 있다. 필요하면 가족 토큰 위치를 고정하는 별도 수리가 필요하다.
