# t614 — SPEC AC 파서 정규식 재컴파일 제거

카드: t614 · 브랜치: `WT-ac-parser-regex` · 기준: 로컬 develop `9935e4e3e`
워크트리: `.claude/worktrees/t614` · 측정 환경: darwin/arm64, Apple M4 Max, go1.26.8

## Claim

1. `internal/spec`의 AC 파싱 경로에서 호출마다 재컴파일되던 고정 정규식 6개를 패키지 수준 `var` + `MustCompile`로 승격했다.
2. 같은 벤치마크 기준으로 할당 수가 `ParseAcceptanceCriteria`에서 93.3%, `ExtractRequirementMappings`에서 88.5% 줄었다.
3. 동작 변화는 없다.

## Evidence

### 승격 대상 — 실측한 재컴파일 지점 6개

배차문의 전제는 primary checkout(`main`) 기준이었고, `acIDPattern`은 develop에서 이미 승격돼 있었다. 워크트리에서 다시 재어 남은 지점만 손댔다.

```
$ grep -n 'regexp\.MustCompile' internal/spec/parser.go internal/spec/ears.go   # 수정 전
internal/spec/parser.go:358:var acIDPattern = regexp.MustCompile(...)   ← 이미 패키지 수준 (대상 아님)
internal/spec/parser.go:385:	reqRemover := regexp.MustCompile(`\(?\s*(?:maps|MAPS)\s+REQ-[A-Z0-9-]+\s*\)?`)
internal/spec/parser.go:392:	givenRe := regexp.MustCompile(`(?i)^Given\s+(.+?)(?:,\s*(?:When|then)|$)`)
internal/spec/parser.go:399:	whenRe := regexp.MustCompile(`(?i)^When\s+(.+?)(?:,\s*(?:Then|then)|$)`)
internal/spec/parser.go:406:	thenRe := regexp.MustCompile(`(?i)^Then\s+(.+)`)
internal/spec/ears.go:128:	sectionPattern := regexp.MustCompile(`(?i)maps\s+(REQ-[A-Z0-9-]+(?:\s*,\s*REQ-[A-Z0-9-]+)*)`)
internal/spec/ears.go:130:	reqPattern := regexp.MustCompile(`REQ-([A-Z0-9-]+)`)
```

수정 후 `parser.go`·`ears.go` 안에 함수 본문 `MustCompile`은 0개.

### 벤치마크 — 같은 벤치, 같은 트리, count=6

측정 명령(전·후 동일):

```
go test -run '^$' -bench 'BenchmarkParseAcceptanceCriteria|BenchmarkExtractRequirementMappings' \
  -benchmem -count=6 ./internal/spec/
```

전체 출력: `.moai/reports/t614/bench-before.txt` / `.moai/reports/t614/bench-after.txt`

| 벤치마크 | 지표 | 전 (중앙값) | 후 (중앙값) | 변화 |
|---|---|---|---|---|
| `ParseAcceptanceCriteria` (AC 30줄) | ns/op | 1,253,928 | 365,156 | **−70.9%** (3.43배) |
| | B/op | 1,940,901 | 112,085 | **−94.2%** |
| | allocs/op | 17,496 | 1,168 | **−93.3%** |
| `ExtractRequirementMappings` | ns/op | 6,938 | 1,756 | **−74.7%** |
| | B/op | 9,923 | 805 | **−91.9%** |
| | allocs/op | 96 | 11 | **−88.5%** |

allocs/op는 6회 전부 동일한 값(전 17496 / 후 1168, 전 96 / 후 11)이라 잡음이 없다. `ExtractRequirementMappings`의 ns/op는 측정 전 구간에서 흔들렸으므로(6234~10613) 시간 축은 할당 축보다 약한 근거로 읽는다.

### 벤치마크가 공허하지 않다는 근거

두 벤치 모두 측정 루프 앞에 fixture 가드를 둔다 — 파서가 fixture를 거부하면 `b.Fatalf`로 죽는다. 조기 반환 경로를 재고 있었다면 벤치가 통과하지 못한다.

```go
if got, _ := ParseAcceptanceCriteria(markdown, true); len(got) != 30 { b.Fatalf(...) }
if got := ExtractRequirementMappings(text); len(got) != 3 { b.Fatalf(...) }
```

### 동작 변화 0

정규식 리터럴은 이동만 했고 **한 바이트도 바뀌지 않았다** — `git diff`에서 바뀐 것은 식별자 이름과 선언 위치뿐이다(`reqRemover`→`acReqRemoverPattern`, `sectionPattern`→`reqSectionPattern` 등).

```
$ go vet ./internal/spec/           # 출력 없음 (exit 0)
$ gofmt -l internal/spec/           # 출력 없음
$ golangci-lint run ./internal/spec/...
0 issues.
$ go test ./internal/spec/
ok  	github.com/modu-ai/moai-adk/internal/spec	75.653s
```

`ExtractRequirementMappings`의 godoc이 새 `var` 블록에 붙어 exported 함수가 문서를 잃는 회귀가 중간에 있었고, `var` 블록을 godoc 위로 옮겨 고쳤다. 확인:

```
$ go doc ./internal/spec ExtractRequirementMappings
func ExtractRequirementMappings(text string) []string
    ExtractRequirementMappings extracts (maps REQ-...) pattern from text.
```

## Baseline-attribution

- 트리: 워크트리 `.claude/worktrees/t614`, 기준 커밋 `9935e4e3e`(`git merge-base --is-ancestor` 확인)
- 전 측정: 위 벤치 명령, 소스 수정 **전** 상태(벤치 파일만 추가). 출력 `bench-before.txt`
- 후 측정: 같은 명령, godoc 수리까지 끝난 최종 트리. 출력 `bench-after.txt`
- 테스트·린트: 최종 트리에서 재실행한 값

## Gaps

- **카드 본문을 읽지 못했다.** 배차문이 준 경로 `.moai/state/todo/backlog.db`는 이 저장소에 없고, 실제 큐(`~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db`)는 열리지 않았다. 작업 범위는 배차문 `cmd:` 지시만 근거로 삼았다.
- **배차문이 인용한 "할당 프로파일 96.99%가 regexp.compile"을 직접 재현하지 않았다.** 이 카드의 판정은 자체 벤치의 전·후 대조이지, 그 프로파일의 재측정이 아니다.
- **패키지 전체 판정은 안 했다.** `go test ./internal/spec/`만 돌렸고 전체 스위트는 CI 몫이다(§4).
- **다른 재컴파일 지점은 손대지 않았다.** `closer.go:484`의 `pattern := regexp.MustCompile(...)`는 인자(`field`)로 패턴을 조립하므로 단순 승격 대상이 아니다 — 범위 밖으로 두었다.
- **darwin/arm64 단일 환경 측정.** 다른 아키텍처의 개선폭은 재지 않았다.

## Residual-risk

- 패키지 수준 정규식은 프로세스 수명 내내 메모리에 남는다. 6개 소형 패턴이라 무시할 크기지만 0은 아니다.
- `regexp.Regexp`는 동시 사용이 안전하므로 공유 자체는 경합을 만들지 않는다. 다만 전에는 호출마다 새 인스턴스를 쓰던 코드가 이제 하나를 공유하므로, 향후 누군가 이 변수에 `Longest()` 같은 상태 변경 메서드를 부르면 전역에 영향을 준다 — 현재 코드에는 그런 호출이 없다.
- 벤치 fixture는 합성 문자열이다. 실제 SPEC 문서의 형태 분포와 다르면 개선폭의 절대값은 달라질 수 있다(할당 감소 방향은 구조적이라 바뀌지 않는다).
