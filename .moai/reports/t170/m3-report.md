# M3 보고 — 취약점 분류기 (SPEC-FEEDBACK-AUTO-SUBMIT-001)

- 카드: t170 · 워크트리 `.claude/worktrees/t170` · 브랜치 `WT-auto-feedback`
- 편집 전 트리: `e51475068` (M2 착지 지점)
- 대상 AC: AC-F-012, AC-F-013, AC-F-008(M2 회귀)

## Claim (주장)

1. `internal/feedback/classify.go` 가 설계 §6 의 신호 3종을 **마스킹 이전 원문**에 대해 판정하고, 차단 시 `SECURITY.md` 의 두 문장 + Advisories URL 을 `Result.Reason` 에 싣는다. (AC-F-012)
2. 분류가 원문을 본다는 사실이 **관측 가능**하다 — 시크릿 1건만 신호로 갖는 본문은 `blocked` 이고, 같은 본문의 마스킹본을 분류기에 단독으로 넣으면 `ok` 다. (AC-F-013)
3. 축퇴("전부 차단") 구현이 배제된다 — 평범한 버그 리포트와 단일 신호 근접 케이스가 `ok` 로 남는다. (AC-F-008 회귀 + 오탐 대조 4케이스)
4. M2 가 고정한 파이프라인 순서는 바뀌지 않았고, `scrub.go` 편집은 placeholder 교체 1줄 + `@MX:TODO` 제거뿐이다.
5. `internal/hook` 은 편집하지 않았다 (AP-13 준수).

## Evidence (증거)

RED — 구현 없는 트리에서 `go test ./internal/feedback/ -run 'TestClassify' -v`:

```
# github.com/modu-ai/moai-adk/internal/feedback [github.com/modu-ai/moai-adk/internal/feedback.test]
internal/feedback/classify_test.go:48:37: undefined: advisoriesURL
internal/feedback/classify_test.go:92:71: too many arguments in call to classify
	have (Input, Options)
	want (Input)
internal/feedback/classify_test.go:127:54: too many arguments in call to classify
	have (Input, Options)
	want (Input)
FAIL	github.com/modu-ai/moai-adk/internal/feedback [build failed]
FAIL
```

GREEN:

```
$ go test ./internal/feedback/ -run 'TestClassifyBlocksVulnerabilityReport' -v
--- PASS: TestClassifyBlocksVulnerabilityReport (0.00s)
    --- PASS: TestClassifyBlocksVulnerabilityReport/advisory_identifier (0.00s)
    --- PASS: TestClassifyBlocksVulnerabilityReport/vulnerability_vocabulary (0.00s)
    --- PASS: TestClassifyBlocksVulnerabilityReport/key-file_mention_plus_vocabulary (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.835s

$ go test ./internal/feedback/ -run 'TestClassifyReadsPreMaskBody' -v
--- PASS: TestClassifyReadsPreMaskBody (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.682s

$ go test ./internal/feedback/ -run 'TestClassifyDoesNotBlockOrdinaryReports' -v
--- PASS: TestClassifyDoesNotBlockOrdinaryReports (0.00s)
    --- PASS: TestClassifyDoesNotBlockOrdinaryReports/single_path_mention (0.00s)
    --- PASS: TestClassifyDoesNotBlockOrdinaryReports/ordinary_bug_report (0.00s)
    --- PASS: TestClassifyDoesNotBlockOrdinaryReports/prose_about_a_token_prefix (0.00s)
    --- PASS: TestClassifyDoesNotBlockOrdinaryReports/single_vocabulary_term (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.464s

$ go test ./internal/feedback/ -run 'TestScrubBenignBodyUntouchedAndAllowed' -v
--- PASS: TestScrubBenignBodyUntouchedAndAllowed (0.00s)
    --- PASS: TestScrubBenignBodyUntouchedAndAllowed/benign_prose (0.00s)
    --- PASS: TestScrubBenignBodyUntouchedAndAllowed/lowercase_prose_run (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.648s

$ go test -count=1 ./internal/feedback/... ./internal/sandbox/...
ok  	github.com/modu-ai/moai-adk/internal/feedback	1.130s
ok  	github.com/modu-ai/moai-adk/internal/sandbox	0.638s

$ go test -race ./internal/feedback/
ok  	github.com/modu-ai/moai-adk/internal/feedback	2.131s

$ go vet ./internal/feedback/... ./internal/sandbox/... ; echo "vet exit=$?"
vet exit=0

$ GOOS=windows go vet ./internal/feedback/... ./internal/sandbox/... ; echo "winvet exit=$?"
winvet exit=0

$ golangci-lint run --timeout=3m ./internal/feedback/... ./internal/sandbox/...
0 issues.

$ grep -c 'security/advisories/new' internal/feedback/classify.go
1

$ gofmt -l internal/feedback/
(무출력)
```

## Baseline-attribution (baseline 귀속)

- 트리: `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`, 편집 전 HEAD `e51475068`.
- AC-F-012 grep 토큰의 base 0 확인: `git ls-tree -r e51475068 --name-only -- internal/feedback/classify.go | wc -l` → `0`. 파일 부재이므로 토큰 히트도 0 — 이 grep 은 사전구현 트리에서 아무것도 반환하지 않는다.
- RED 는 `classify.go` 작성 **이전**, 테스트 파일만 존재하는 상태에서 캡처한 컴파일 실패 출력이다.
- 위 모든 명령은 이번 실행, 이 트리에서 관측됐다.

## Gaps (미검증)

- **전체 스위트를 돌리지 않았다.** 지시대로 패키지 스코프(`./internal/feedback/...`, `./internal/sandbox/...`)만 실행했다. 전 패키지 판정은 CI 몫이다.
- **임계값 2 의 실사용 오탐률은 측정하지 않았다.** 테스트가 관측하는 것은 8개 케이스의 이산 판정일 뿐, 실제 피드백 분포에 대한 오탐률이 아니다. 이 값은 이 SPEC 이 처음 정하는 값이며 근거는 설계의 비용 비대칭 논변이다.
- **영어 이외 언어 본문의 어휘 신호는 검증하지 않았다** — 어휘 목록이 영어 전용이라 검증할 대상이 없다. 신호 1(시크릿 패턴)·2(경로)는 언어 독립이라 그대로 동작한다.
- **CLI 경계(M5)·스킬 본문(M6)은 이 마일스톤 밖이다.** 분류 결과가 실제로 제출을 막는지는 산문 호출자가 `verdict` 를 읽는지에 달려 있고, 그 배선은 M5·M6 소관이다.
- **`Result.Reason` 문구가 `SECURITY.md` 와 동기화된 상태로 유지되는지 검사하는 기계 장치는 없다.** 두 문장은 상수로 복사돼 있으며, `SECURITY.md` 가 개정되면 수동으로 맞춰야 한다.

## Residual-risk (잔여 위험)

1. **영어 어휘 편중.** 한국어·일본어 등으로 쓰인 취약점 보고는 신호 3 을 전혀 발생시키지 않는다. 시크릿이 본문에 없고 경로 언급도 없는 순수 산문 취약점 보고는 통과한다. 완화는 이 마일스톤 범위 밖(후속 카드 후보).
2. **CVE 식별자 단독 차단의 오탐.** "의존성 CVE-XXXX-XXXX 때문에 버전 올려주세요" 같은 무해한 요청이 차단된다. 설계가 명시한 비대칭 비용(미탐=공개 유출 / 오탐=수동 경로 1회) 하에서 의도된 방향이지만, 실사용에서 잦으면 가중치 재조정 대상이다.
3. **경로 토큰 추출의 정밀도.** 산문 안 경로를 정규식으로 잘라내므로, 따옴표·괄호가 붙은 이형은 놓칠 수 있다(과소 탐지 방향). 신호 2 단독으로는 차단하지 않으므로 즉시 유출로 이어지지는 않는다.
4. **`classify` 가 `hook.DefaultSecurityPolicy()` 를 별도로 한 번 더 빌드한다.** `Scrub` 1회 호출당 정책 컴파일이 2회 일어난다. 일회성 CLI 호출이므로 실사용 비용은 무시할 수준이지만, 이 함수가 루프 안에서 호출되는 소비자가 생기면 재검토 대상이다.
5. **거부 메시지 동기화 부채.** 위 Gaps 마지막 항목의 위험 형태 — `SECURITY.md` 개정이 조용히 드리프트를 만든다.
