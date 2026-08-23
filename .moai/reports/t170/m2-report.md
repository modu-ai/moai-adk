# M2 완료 보고 — SPEC-FEEDBACK-AUTO-SUBMIT-001

> 스크러버 타입 계약 + 마스킹 변환. 워크트리 `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`, base `3210da7d3`, M2 착수 시 HEAD `95fc239e3`.

## 1. Claim (주장)

M2가 계획한 6개 파일 + 테스트 3종이 착지했고, M2에 배정된 AC 9건(AC-F-005·006·007·008·009·010·011·014·024)이 각각 자기 명령으로 PASS한다. `internal/feedback` 밖 편집은 `internal/sandbox/env.go`의 접근자 1개로 한정된다.

## 2. Evidence (증거)

RED — 테스트만 있고 구현이 없던 트리:

```
# github.com/modu-ai/moai-adk/internal/feedback [github.com/modu-ai/moai-adk/internal/feedback.test]
internal/feedback/scrub_test.go:52:20: undefined: Options
internal/feedback/scrub_test.go:59:21: undefined: Result
internal/feedback/scrub_test.go:59:50: undefined: Finding
internal/feedback/scrub_test.go:73:14: undefined: Scrub
internal/feedback/scrub_test.go:104:28: undefined: KindSecret
internal/feedback/scrub_test.go:104:28: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/feedback [build failed]
```

GREEN — AC별 명령 9건 각각 1회 실행, 전부 PASS(전문은 `progress.md` §E.2 M2):

| AC | 명령 | 관측 |
|---|---|---|
| AC-F-005 | `go test ./internal/feedback/ -run 'TestFindingsCarryNoRawValue' -v` | `--- PASS` / `ok ... 0.555s` |
| AC-F-006 | `go test ./internal/feedback/ -run 'TestScrubMasksGitHubToken' -v` | `--- PASS` + body/title 서브테스트 2건 |
| AC-F-007 | `go test ./internal/feedback/ -run 'TestScrubMasksGoogleAPIKey' -v` | `--- PASS` / `ok ... 0.661s` |
| AC-F-008 | `go test ./internal/feedback/ -run 'TestScrubBenignBodyUntouchedAndAllowed' -v` | `--- PASS` + 서브테스트 2건 |
| AC-F-009 | `go test ./internal/feedback/ -run 'TestMaskOutputShapeMatchesExistingMasker' -v` | `--- PASS` / `ok ... 0.761s` |
| AC-F-010 | `go test ./internal/feedback/ -run 'TestScrubCollapsesHomePath' -v` | `--- PASS` + 서브테스트 2건 |
| AC-F-011 | `go test ./internal/feedback/ -run 'TestScrubMasksEnvValues' -v` | `--- PASS` + 서브테스트 3건 |
| AC-F-014 | `go test ./internal/feedback/ -run 'TestScrubIsIdempotent' -v` | `--- PASS` / `ok ... 0.613s` |
| AC-F-024 | `go test ./internal/feedback/ -run 'TestScrubMasksPrivateKeyBlockEntirely' -v` | `--- PASS` + 서브테스트 2건 |

패키지·정적 검사:

```
$ go test ./internal/feedback/... ./internal/sandbox/...
ok  	github.com/modu-ai/moai-adk/internal/feedback	1.009s
ok  	github.com/modu-ai/moai-adk/internal/sandbox	0.527s

$ go vet ./internal/feedback/... ./internal/sandbox/...            → 무출력, exit 0
$ GOOS=windows go vet ./internal/feedback/... ./internal/sandbox/... → 무출력, exit 0
$ go test -race ./internal/feedback/                                → ok ... 2.133s
$ golangci-lint run --timeout=3m ./internal/feedback/... ./internal/sandbox/... → 0 issues.
```

## 3. Baseline-attribution (baseline 귀속)

편집 전 트리는 `95fc239e3`(M1 착지 커밋). 세 건 모두 0으로 실측:

```
$ git ls-tree -r 95fc239e3 --name-only -- internal/feedback | wc -l   → 0
$ git grep -c "DefaultEnvDenyList" 95fc239e3 -- '*.go' | wc -l        → 0
$ git grep -c "func Scrub(" 95fc239e3 -- '*.go' | wc -l               → 0
```

RED 블록은 이 트리에 테스트 파일만 올린 상태에서 관측했고, GREEN 블록은 구현 후 같은 워크트리에서 관측했다. 인용한 수치는 전부 이번 실행의 출력이며 다른 마일스톤·다른 시점의 값을 옮겨오지 않았다.

AC-F-008 네 번째 케이스(과잉 마스킹 방향)는 통과 자체로는 아무것도 관측하지 않을 수 있어 — 패턴이 애초에 매치하지 않으면 무조건 통과한다 — 반증 장치 `TestRewritePatternsAreCaseSensitive`를 함께 넣었다. 이 테스트는 (a) 재작성 패턴 전부가 `(?i)` 없이 컴파일되고 그 산문에 매치하지 않으며, (b) **같은 패턴에 `(?i)`를 붙이면 매치한다**를 같은 실행에서 확인한다. (b)가 깨지면 테스트가 실패하므로 "관측하는 것이 없는 통과"가 배제된다.

## 4. Gaps (미검증)

- **분류기는 구현하지 않았다.** `classify()`는 항상 `ok`를 반환하는 placeholder이며 `@MX:TODO`로 표시했다. AC-F-012·F-013은 M3 소관이고, M2가 고정한 것은 "분류가 마스킹 이전 원문을 본다"는 **순서**뿐이다.
- **`ResolveProjectRoot` / `Options.ProjectRoot`는 프로덕션 호출자가 없다.** M4(마스킹 로그·재시도 큐)가 소비할 때까지 단위 테스트로만 덮여 있다.
- **바이너리 스모크·CLI 계약은 실행하지 않았다.** `moai feedback scrub`은 M5 소관이며 M2 범위 밖이다.
- **전 패키지 스위트를 돌리지 않았다.** 로컬 검증은 패키지 스코프로만 했다(CLAUDE.local.md §4). 전 패키지 판정은 CI 몫이다.
- **`internal/hook` 스위트를 돌리지 않았다.** `internal/feedback`이 `hook.DefaultSecurityPolicy()`를 읽기만 하고 hook 쪽을 편집하지 않았으므로 영향 패키지로 보지 않았다.

## 5. Residual-risk (잔여 위험)

- **`MaskSecret`은 값의 첫 1자와 끝 4자를 남긴다.** 채택 제약(호출 가능한 기존 마스커가 이것 하나)과 AC-F-009(형태 일치)가 겹쳐 선택지가 없었지만, 공개 이슈로 나가는 텍스트에 토큰 꼬리 4자가 남는다는 뜻이다. 형태를 바꾸려면 세 마스커를 통합하는 별도 카드가 필요하다.
- **env 값 8자 하한.** 8자 미만 자격증명은 마스킹되지 않는다. 산문 파괴를 피하려는 의도적 교환이며, 그런 값이 실재하면 미탐이다.
- **`AWS_` 접두사 미채택.** `AWS_SECRET_ACCESS_KEY` 같은 이름이 deny list에 열거돼 있지 않으면 그 값은 이름 어휘로 걸리지 않는다(`AKIA` 형태의 액세스 키 ID는 정규식이 잡는다). 열거 목록을 늘리는 것은 `internal/sandbox`의 소관이라 이 SPEC에서 건드리지 않았다.
- **본문 종류에 따른 과잉/과소 마스킹은 양방향 AC 2건으로만 관측된다.** 실제 사용자 보고서 분포는 표본이 없다.
- **`hook` 정책이 나중에 패턴을 추가하면 재작성 경로도 자동으로 넓어진다.** 정책 객체를 통째로 받는 설계의 의도된 성질이지만, 새 패턴이 대소문자에 민감하지 않은 형태라면 산문 훼손이 이쪽으로 먼저 나타난다.
