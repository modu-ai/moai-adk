# Progress — SPEC-FEEDBACK-AUTO-SUBMIT-001

## §E.1 Plan-phase Audit-Ready Signal

- 2026-08-22 (1차) — plan-phase 산출물 3종 작성(Tier M). 근거는 `.moai/reports/t170/` 읽기 전용 렌즈 4종. 카드 전제 4건 반증 → §A 정정으로 기록, 운영자 결정 D1~D4 반영.
- 2026-08-22 (2차, 분리) — **AC 예산 초과로 SPEC을 둘로 분리**. 측정: `### AC-` 32개 > Tier M 상한 16, Tier L 상한 25. `spec-workflow.md` § SPEC Complexity Tier에 따라 완화가 아니라 분할로 대응.
  - 이 SPEC: **Tier L**(5종 — spec/plan/acceptance/design/research)로 재작성. 범위는 피드백 축(확인 게이트·스크러버·마스킹·취약점 분류·로그·큐·스킬 조항 + `feedback.auto_submit` 배선). **AC 23 / 상한 25**, REQ 13.
  - 신설 형제: `SPEC-TODO-ENABLE-FLAG-001`(**Tier M**, 3종). 범위는 todo 축. **AC 11 / 상한 16**, REQ 6.
  - 두 SPEC은 파일 9종을 공유하며, 병합 규율([HARD] 같은 파일의 다른 항목만 추가)은 양쪽 §E.1에 대칭으로 기록했다. `depends_on`은 기능 의존이 없어 미기재하고 그 근거를 §E.1에 남겼다.
  - 남은 결정: §D 결정 D5(웹 노출 경로 A/B) — 착수 승인 시점 확정.

- 2026-08-22 (iter2) — plan-audit iter1 **FAIL 0.75**(Tier L 임계 0.85) + **MP-2 FAIL** 대응 개정. 블로킹 8건(D1 제목 미스크럽 / D1b 탐지기→재작성기 비대칭 / D2 공허 선택자 / D3 unexported 재사용 / MP-2·D8 REQ-12 + 결정 유예 / D4 큐-초안 충돌 / D7 마일스톤 AC 범위 / D5·D6·D11) + 선택 3건(D9·D10·D12) 처리. 강제 주장의 정직성(§E.3 · design.md §1 · plan.md AP-12)은 감사 유지 판정을 받아 **후퇴시키지 않았다**. 결정 D5는 선택지 A로 **해소**(유예 삭제). AC 23 → 24(상한 25). version 0.2.0 → 0.3.0, 항목별 내역은 `spec.md` §G.

- 2026-08-22 (iter3, Tier L 상한 회차) — plan-audit iter2 **FAIL 0.84**(임계 0.85, must-pass 7/7 통과) 대응. 블로킹 2건만 처리: **N1** 제목 의무를 관측 가능하게(REQ-10 문구 + AC-F-019를 두 사본 × 5 grep으로 확장, 새 AC 없음), **N2** 누락된 AC-F-013을 M3 Exit에 복귀. 선택 4건(N3~N6)은 상한 회차라 의도적으로 미처리. AC 24 유지(상한 25), version 0.3.0 → 0.4.0.

- 2026-08-22 (iter4, 최종) — plan-audit iter3 **FAIL 0.863**(임계 0.85, must-pass 7/7) + **Tier L 재시도 상한 도달 → 운영자 결정**: 기계적으로 증명된 2건만 수정, 나머지는 M6 부채 이관. **재감사 없음.** 수정: AC-F-019 ④⑤ 토큰을 base 0/0 실측본으로 교체(`queue.json`, `label`↔`conversation_language` 동일 줄) + "기준 토큰은 base에서 0을 반환해야 한다" 규칙 명문화. 이관: ② `--title` 앵커 부재, ③ 한국어 리터럴 — 감사 인용 문단을 영어 원문 그대로 `acceptance.md`·`plan.md` M6 두 곳에 [HARD]로 고정. AC 24 유지, version 0.4.0 → 0.5.0. plan-phase 종료.

## §E.2 Run-phase Evidence

Cycle: TDD (RED → GREEN → REFACTOR). 워크트리 `.claude/worktrees/t170`, 브랜치 `WT-auto-feedback`, base `3210da7d3`.

### M1 — 설정 데이터 모델

Pre-flight: `grep -rn "auto_submit\|AutoSubmit" --include='*.go' internal/ | grep -v _test` → **0건**
(편집 전 트리에 이 키가 없다는 근거).

RED (`go test ./internal/config/ -run 'TestFeedbackAutoSubmit' -v`, 편집 전 트리):

```
# github.com/modu-ai/moai-adk/internal/config [github.com/modu-ai/moai-adk/internal/config.test]
internal/config/feedback_auto_submit_test.go:36:16: cfg.FeedbackAutoSubmit undefined (type *Config has no field or method FeedbackAutoSubmit)
internal/config/feedback_auto_submit_test.go:50:16: cfg.FeedbackAutoSubmit undefined (type *Config has no field or method FeedbackAutoSubmit)
internal/config/feedback_auto_submit_test.go:64:5: undefined: DefaultFeedbackAutoSubmit
internal/config/feedback_auto_submit_test.go:65:61: undefined: DefaultFeedbackAutoSubmit
internal/config/feedback_auto_submit_test.go:67:42: fc.AutoSubmit undefined (type FeedbackConfig has no field or method AutoSubmit)
internal/config/feedback_auto_submit_test.go:67:56: undefined: DefaultFeedbackAutoSubmit
internal/config/feedback_auto_submit_test.go:68:73: fc.AutoSubmit undefined (type FeedbackConfig has no field or method AutoSubmit)
internal/config/feedback_auto_submit_test.go:68:85: undefined: DefaultFeedbackAutoSubmit
FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]
```

GREEN (같은 명령, 구현 후):

```
=== RUN   TestFeedbackAutoSubmitKeyResolution
=== PAUSE TestFeedbackAutoSubmitKeyResolution
=== RUN   TestFeedbackAutoSubmitCompiledDefault
=== PAUSE TestFeedbackAutoSubmitCompiledDefault
=== CONT  TestFeedbackAutoSubmitKeyResolution
=== CONT  TestFeedbackAutoSubmitCompiledDefault
--- PASS: TestFeedbackAutoSubmitCompiledDefault (0.00s)
--- PASS: TestFeedbackAutoSubmitKeyResolution (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/config	0.383s
```

패키지 트리 전체 (`go test ./internal/config/...`):

```
ok  	github.com/modu-ai/moai-adk/internal/config	3.935s
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile	0.642s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	(cached)
```

`go vet ./internal/config/...` · `GOOS=windows go vet ./internal/config/...` 둘 다 무출력(통과).

납품물:
- `internal/config/types.go` — `FeedbackConfig.AutoSubmit bool` + `yaml:"auto_submit"`.
- `internal/config/defaults.go` — `DefaultFeedbackAutoSubmit = false` 상수 + `NewDefaultFeedbackConfig()`에서 대입.
- `internal/config/feedback_accessors.go` — `(*Config).FeedbackAutoSubmit() bool`.
- `internal/config/feedback_auto_submit_test.go` — AC-F-001 단언 2종.

설계 판단: 접근자에 fallback 분기를 두지 않았다. `FeedbackRepository()`의 빈 문자열 fallback은
"명시된 빈 값"과 "부재"가 구별되지 않는 문자열 키라서 필요하지만, bool 키는 loader가
`NewDefaultFeedbackConfig()`로 wrapper를 seed한 뒤 디코드하는 부분 오버라이드 계약(`loader.go`)이
부재를 이미 컴파일 기본값으로 해석한다. 분기를 추가하면 `false` 명시와 부재가 같은 코드 경로로
합쳐져 죽은 조건이 된다. 테스트가 그 계약을 하드코딩된 구조체가 아니라 `Loader.Load()`로
관측하는 이유가 이것이다.

AC-F-001 판정: **PASS** — `go test ./internal/config/ -run 'TestFeedbackAutoSubmit' -v` 가 위 GREEN 블록을 출력.

### M2 — 스크러버 타입 계약 + 마스킹 변환

Pre-flight (편집 전 트리 `95fc239e3` 대상, 세 건 모두 **0**):

```
$ git ls-tree -r 95fc239e3 --name-only -- internal/feedback | wc -l
       0
$ git grep -c "DefaultEnvDenyList" 95fc239e3 -- '*.go' | wc -l
       0
$ git grep -c "func Scrub(" 95fc239e3 -- '*.go' | wc -l
       0
```

`internal/feedback` 패키지 자체가 없었고, `DefaultEnvDenyList` 접근자도 `Scrub` 함수도 부재였다는 근거.

RED (`go test ./internal/feedback/ -run '<M2 9종>' -v`, 테스트만 있고 구현 없는 트리):

```
# github.com/modu-ai/moai-adk/internal/feedback [github.com/modu-ai/moai-adk/internal/feedback.test]
internal/feedback/scrub_test.go:52:20: undefined: Options
internal/feedback/scrub_test.go:53:9: undefined: Options
internal/feedback/scrub_test.go:59:21: undefined: Result
internal/feedback/scrub_test.go:59:50: undefined: Finding
internal/feedback/scrub_test.go:65:9: undefined: Finding
internal/feedback/scrub_test.go:73:14: undefined: Scrub
internal/feedback/scrub_test.go:73:20: undefined: Input
internal/feedback/scrub_test.go:97:15: undefined: Scrub
internal/feedback/scrub_test.go:97:21: undefined: Input
internal/feedback/scrub_test.go:104:28: undefined: KindSecret
internal/feedback/scrub_test.go:104:28: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/feedback [build failed]
FAIL
```

GREEN — AC별 명령을 각각 1회씩 실행해 관측한 결과:

```
$ go test ./internal/feedback/ -run 'TestFindingsCarryNoRawValue' -v
--- PASS: TestFindingsCarryNoRawValue (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.555s

$ go test ./internal/feedback/ -run 'TestScrubMasksGitHubToken' -v
--- PASS: TestScrubMasksGitHubToken (0.00s)
    --- PASS: TestScrubMasksGitHubToken/body (0.00s)
    --- PASS: TestScrubMasksGitHubToken/title (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.715s

$ go test ./internal/feedback/ -run 'TestScrubMasksGoogleAPIKey' -v
--- PASS: TestScrubMasksGoogleAPIKey (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.661s

$ go test ./internal/feedback/ -run 'TestScrubBenignBodyUntouchedAndAllowed' -v
--- PASS: TestScrubBenignBodyUntouchedAndAllowed (0.00s)
    --- PASS: TestScrubBenignBodyUntouchedAndAllowed/benign_prose (0.00s)
    --- PASS: TestScrubBenignBodyUntouchedAndAllowed/lowercase_prose_run (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.719s

$ go test ./internal/feedback/ -run 'TestMaskOutputShapeMatchesExistingMasker' -v
--- PASS: TestMaskOutputShapeMatchesExistingMasker (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.761s

$ go test ./internal/feedback/ -run 'TestScrubCollapsesHomePath' -v
--- PASS: TestScrubCollapsesHomePath (0.00s)
    --- PASS: TestScrubCollapsesHomePath/first_home (0.00s)
    --- PASS: TestScrubCollapsesHomePath/second_home (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.801s

$ go test ./internal/feedback/ -run 'TestScrubMasksEnvValues' -v
--- PASS: TestScrubMasksEnvValues (0.00s)
    --- PASS: TestScrubMasksEnvValues/process_environment (0.00s)
    --- PASS: TestScrubMasksEnvValues/default_deny_list (0.00s)
    --- PASS: TestScrubMasksEnvValues/env_scrub_extra (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.628s

$ go test ./internal/feedback/ -run 'TestScrubIsIdempotent' -v
--- PASS: TestScrubIsIdempotent (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.613s

$ go test ./internal/feedback/ -run 'TestScrubMasksPrivateKeyBlockEntirely' -v
--- PASS: TestScrubMasksPrivateKeyBlockEntirely (0.00s)
    --- PASS: TestScrubMasksPrivateKeyBlockEntirely/with_terminator (0.00s)
    --- PASS: TestScrubMasksPrivateKeyBlockEntirely/truncated_block (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.588s
```

AC-F-008 네 번째 케이스의 반증 장치 (`TestRewritePatternsAreCaseSensitive`) — 이 케이스가 **무언가를 관측한다**는 근거. 재작성 패턴 전부가 `(?i)` 없이 컴파일되고 그 산문에 매치하지 않으면서, **같은 패턴에 `(?i)`를 붙이면 매치한다**는 것을 같은 테스트가 확인한다:

```
$ go test ./internal/feedback/ -run 'TestRewritePatternsAreCaseSensitive' -v
--- PASS: TestRewritePatternsAreCaseSensitive (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.787s
```

패키지 트리 전체 (`go test ./internal/feedback/... ./internal/sandbox/...`):

```
ok  	github.com/modu-ai/moai-adk/internal/feedback	1.009s
ok  	github.com/modu-ai/moai-adk/internal/sandbox	0.527s
```

`go vet ./internal/feedback/... ./internal/sandbox/...` · `GOOS=windows go vet ./internal/feedback/... ./internal/sandbox/...` 둘 다 무출력, exit 0.
`go test -race ./internal/feedback/` → `ok ... 2.133s`.
`golangci-lint run --timeout=3m ./internal/feedback/... ./internal/sandbox/...` → `0 issues.`

납품물:
- `internal/feedback/scrub.go` — `Input`/`Finding`(`Where` 포함)/`Result`/`Options` 타입 계약 + `Scrub()` + 파이프라인 순서 고정 + 분류 seam(placeholder, M3 소관).
- `internal/feedback/patterns.go` — 정책 객체 통째 재사용 + `AIza` 합집합 + 치환 span 규칙(마커 앵커 → 종료자까지/입력 끝까지, `(?i)` 제거 재컴파일).
- `internal/feedback/env.go` — 이름 어휘 기반 값 마스킹 + `env_scrub_extra` 확장.
- `internal/feedback/paths.go` — `paths.Home()` 기반 홈 축약 + `.moai/` 마커 상향 탐색.
- `internal/sandbox/env.go` — `DefaultEnvDenyList()` 신설(사본 반환). 이 SPEC이 `internal/feedback` 밖을 편집하는 유일한 지점.
- 테스트 4종 파일: `internal/feedback/scrub_test.go`, `internal/feedback/paths_test.go`, `internal/sandbox/env_denylist_test.go`.

설계 판단 5건:

1. **채택한 마스커는 `github.MaskSecret`이다.** 후보 3종 중 `maskAPIKey`(`internal/cli/glm.go`)와 `maskPartial`(`internal/cli/glm_tools.go`)은 **unexported**라 `internal/feedback`에서 호출할 수 없고, `internal/cli`를 import하면 M5에서 역방향 순환이 된다. AC-F-009가 "형태 문자열 하드코딩 금지"를 요구하므로 테스트가 직접 호출할 수 있어야 한다는 제약과 합쳐지면 선택지는 하나뿐이다. 네 번째 형태는 만들지 않았다(AP-5).
2. **재작성 경로는 정책 패턴을 `(?i)` 없이 재컴파일한다 — 전량이다.** "대소문자 민감 패턴만" 골라내려면 어떤 패턴이 민감한지 판정하는 규칙이 필요한데, 글자를 포함한 패턴은 전부 `(?i)`의 영향을 받으므로 그 판정은 결국 "글자를 포함하는가"와 같다. 전량 제거가 같은 결과를 더 단순하게 낸다. 실제 자격증명은 발급 시스템이 정한 표준 대소문자로 나오므로 탐지력 손실이 없다.
3. **블록 종료자는 span 규칙이지 탐지 패턴이 아니다.** `-----BEGIN`/`-----END` 마커를 `patterns.go`에 상수로 둔 것은 **이미 정책이 적중시킨 매치의 치환 범위를 늘리기 위해서**이며, 새 탐지 패턴을 추가하는 것이 아니다(AP-4 위반 아님). 종료자가 없는 잘린 키는 더 안전한 경우가 아니라 더 위험한 경우이므로 입력 끝까지 마스킹한다.
4. **샌드박스의 `AWS_` 접두사 규칙은 의도적으로 채택하지 않았다.** 자식 프로세스에서 `AWS_REGION`을 떼는 것은 무해하지만, 텍스트 재작성기에서는 본문에 등장한 `us-east-1`을 가려 산문을 훼손한다. 접두사가 비밀이 아닌 변수까지 덮으므로 과잉 마스킹이다. 이름 어휘(deny list ∪ `env_scrub_extra`)만 쓴다.
5. **env 값 마스킹에 최소 길이 8자 하한을 두었다.** 8자 미만 값은 자격증명일 확률이 낮고 평범한 단어일 확률이 높아, 마스킹하면 사용자가 제출하려던 보고서가 갈가리 찢긴다. 8자 미만 시크릿에 대한 이론적 미탐과 산문 파괴를 맞바꾼 판단이며 상수로 노출했다.

AC-F-005 판정: **PASS** — `go test ./internal/feedback/ -run 'TestFindingsCarryNoRawValue' -v` 가 위 블록을 출력.
AC-F-006 판정: **PASS** — `go test ./internal/feedback/ -run 'TestScrubMasksGitHubToken' -v` 가 위 블록을 출력.
AC-F-007 판정: **PASS** — `go test ./internal/feedback/ -run 'TestScrubMasksGoogleAPIKey' -v` 가 위 블록을 출력.
AC-F-008 판정: **PASS** — `go test ./internal/feedback/ -run 'TestScrubBenignBodyUntouchedAndAllowed' -v` 가 위 블록을 출력.
AC-F-009 판정: **PASS** — `go test ./internal/feedback/ -run 'TestMaskOutputShapeMatchesExistingMasker' -v` 가 위 블록을 출력.
AC-F-010 판정: **PASS** — `go test ./internal/feedback/ -run 'TestScrubCollapsesHomePath' -v` 가 위 블록을 출력.
AC-F-011 판정: **PASS** — `go test ./internal/feedback/ -run 'TestScrubMasksEnvValues' -v` 가 위 블록을 출력.
AC-F-014 판정: **PASS** — `go test ./internal/feedback/ -run 'TestScrubIsIdempotent' -v` 가 위 블록을 출력.
AC-F-024 판정: **PASS** — `go test ./internal/feedback/ -run 'TestScrubMasksPrivateKeyBlockEntirely' -v` 가 위 블록을 출력.

미검증(M2 시점): `Result.Verdict`는 placeholder 분류기가 항상 `ok`를 낸다 — AC-F-012·F-013은 M3 소관이며 M2는 그 순서만 고정했다. `ResolveProjectRoot`·`Options.ProjectRoot`는 M4(로그·큐)가 소비하기 전까지 프로덕션 호출자가 없다.

### M3 — 취약점 분류기

Pre-flight (편집 전 트리 `e51475068` 대상):

```
$ git ls-tree -r e51475068 --name-only -- internal/feedback/classify.go | wc -l
       0
```

`classify.go` 자체가 없었고, 분류 seam 은 `scrub.go` 안에 항상 `ok` 를 돌려주는 placeholder(`@MX:TODO` 표시)로만 있었다는 근거. AC-F-012 의 grep 토큰(`security/advisories/new`)은 그 파일이 부재였으므로 base 에서 **0** 이다 — 산문에 이미 있는 단어를 세는 공허한 검사가 아니다.

RED (`go test ./internal/feedback/ -run 'TestClassify' -v`, 테스트만 있고 구현 없는 트리):

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

GREEN — AC별 명령을 각각 1회씩 실행해 관측한 결과:

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

$ grep -c 'security/advisories/new' internal/feedback/classify.go
1
```

패키지 트리 전체 (`go test -count=1 ./internal/feedback/... ./internal/sandbox/...`):

```
ok  	github.com/modu-ai/moai-adk/internal/feedback	1.130s
ok  	github.com/modu-ai/moai-adk/internal/sandbox	0.638s
```

`go vet ./internal/feedback/... ./internal/sandbox/...` · `GOOS=windows go vet ./internal/feedback/... ./internal/sandbox/...` 둘 다 무출력, exit 0.
`go test -race ./internal/feedback/` → `ok ... 2.131s`.
`golangci-lint run --timeout=3m ./internal/feedback/... ./internal/sandbox/...` → `0 issues.`
`gofmt -l internal/feedback/` → 무출력.

AC-F-008 판정: **PASS** — `go test ./internal/feedback/ -run 'TestScrubBenignBodyUntouchedAndAllowed' -v` 가 위 블록을 출력.
AC-F-012 판정: **PASS** — `go test ./internal/feedback/ -run 'TestClassifyBlocksVulnerabilityReport' -v` 가 위 블록을 출력.
AC-F-013 판정: **PASS** — `go test ./internal/feedback/ -run 'TestClassifyReadsPreMaskBody' -v` 가 위 블록을 출력.

납품물:
- `internal/feedback/classify.go` — 신호 3종(시크릿 패턴 적중 / 시크릿 보유 파일 부류 언급 / 취약점 어휘·식별자) + 임계값 판정 + `SECURITY.md` 인용 거부 메시지. 어휘·가중치·URL 전부 패키지 상수·패키지 var.
- `internal/feedback/classify_test.go` — 차단 3케이스(AC-F-012) + **오탐 대조 4케이스**(축퇴 배제) + 순서 가드(AC-F-013).
- `internal/feedback/scrub.go` — placeholder 를 실제 분류기로 교체(`classify(in, opt)`) + `@MX:TODO` 제거. 파이프라인 순서는 M2 가 고정한 그대로.

설계 판단 5건:

1. **파일 부류 신호는 정책의 `DenyPatterns` 를 통째로 재사용한다.** 패턴 문자열을 골라 복사하면 AP-4 와 같은 형태의 사본이 생기고, 어느 패턴이 "키 파일 부류"인지 판정하는 규칙이 결국 그 사본이 된다. 정책 전체가 이미 "절대 쓰면 안 되는 파일" 집합이므로 그대로 쓴다. `internal/hook` 은 편집하지 않았다(AP-13).
2. **정책 패턴은 산문 전체가 아니라 경로꼴 토큰에만 적용한다.** `\.pem$` · `\.ssh/.*` 는 **파일 경로**로 앵커된 패턴이라 문단 전체에 돌리면 원래 물음과 다른 물음에 답한다. 텍스트에서 경로꼴 토큰을 뽑아 토큰 단위로 매치한다.
3. **파일 부류 점수는 매치한 패턴 수가 아니라 서로 다른 토큰 수다.** 홈 아래 SSH 개인키 경로 하나가 정책 패턴 두 개(`\.ssh/.*`, `id_rsa.*`)를 동시에 만족하므로, 패턴 단위로 세면 언급 한 건이 혼자 임계값에 닿는다 — "신호 2·3이 조합된다"는 설계와 어긋난다.
4. **신호 1은 재작성 패턴(대소문자 민감)으로 본다.** 탐지기 형태(`(?i)`)로 보면 키 모양을 우연히 갖춘 소문자 산문이 차단되는데, 그 차단 뒤에는 자격증명이 없다 — 근거 없는 오탐이다. M2 의 재작성 패턴을 그대로 재사용한다.
5. **어휘는 영어 전용이며, 이는 누락이 아니라 알려진 한계다.** 신호 1·2 는 언어 독립이라 최고 위험 입력을 덮고, 16개 언어를 고정 목록으로 덮는다는 주장은 지킬 수 없다. 소스 주석과 보고서 잔여 위험에 명시했다.

미검증(M3 시점): 임계값 2 는 이 SPEC 이 처음 정한 값이라 실사용 오탐률 근거가 없다. `Options.ProjectRoot` 는 여전히 M4(로그·큐) 소비 전까지 프로덕션 호출자가 없다.

### M4 — 온디스크 산출물 2종

Pre-flight (편집 전 트리 `3bcceffc7`, base `3210da7d3` 대상):

```
$ git ls-tree -r 3210da7d3 --name-only -- internal/feedback/masklog.go internal/feedback/queue.go | wc -l
       0
$ git ls-tree -r 3bcceffc7 --name-only -- internal/feedback/
internal/feedback/classify.go
internal/feedback/classify_test.go
internal/feedback/env.go
internal/feedback/paths.go
internal/feedback/paths_test.go
internal/feedback/patterns.go
internal/feedback/scrub.go
internal/feedback/scrub_test.go
```

두 파일 모두 부재였고 `Options.ProjectRoot` 는 M2 가 만든 뒤 프로덕션 소비자가 없었다 — M4 가 그 소비자다.

RED (`go test ./internal/feedback/ -run 'TestMaskLog|TestQueue' -v`, 테스트만 있고 구현 없는 트리):

```
# github.com/modu-ai/moai-adk/internal/feedback [github.com/modu-ai/moai-adk/internal/feedback.test]
internal/feedback/queue_test.go:15:40: undefined: QueueStore
internal/feedback/queue_test.go:22:15: undefined: NewQueueStore
internal/feedback/queue_test.go:22:29: undefined: QueuePathForRoot
internal/feedback/queue_test.go:54:10: undefined: QueuePathForRoot
internal/feedback/queue_test.go:90:41: undefined: queueFilePerm
internal/feedback/queue_test.go:91:46: undefined: queueFilePerm
internal/feedback/masklog_test.go:46:10: undefined: MaskLogPathForRoot
internal/feedback/masklog_test.go:84:39: undefined: maskLogPerm
internal/feedback/masklog_test.go:85:48: undefined: maskLogPerm
internal/feedback/masklog_test.go:127:23: undefined: MaskLogPathForRoot
internal/feedback/queue_test.go:91:46: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/feedback [build failed]
FAIL
```

GREEN — AC별 명령을 각각 1회씩 실행해 관측한 결과:

```
$ go test ./internal/feedback/ -run 'TestMaskLogRecordsKindAndCountWithoutRawValue' -v
    masklog_test.go:58: mask log entry: 2026-08-23T18:37:28+09:00 | total=1 | kind=secret where=body count=1
--- PASS: TestMaskLogRecordsKindAndCountWithoutRawValue (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.568s

$ go test ./internal/feedback/ -run 'TestMaskLog' -v
--- PASS: TestMaskLogRequiresProjectRoot (0.00s)
--- PASS: TestMaskLogSkipsCleanScrub (0.00s)
2026/08/23 18:37:06 WARN feedback: cannot create mask log directory dir=/var/folders/.../001/.moai/logs error="mkdir /var/folders/.../001/.moai/logs: not a directory"
--- PASS: TestMaskLogRecordsKindAndCountWithoutRawValue (0.00s)
--- PASS: TestMaskLogFailureIsFailOpen (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.414s

$ go test ./internal/feedback/ -run 'TestMaskLogFailureIsFailOpen' -v
2026/08/23 18:37:09 WARN feedback: cannot create mask log directory dir=/var/folders/.../001/.moai/logs error="mkdir /var/folders/.../001/.moai/logs: not a directory"
--- PASS: TestMaskLogFailureIsFailOpen (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.405s

$ go test ./internal/feedback/ -run 'TestQueueEnqueuesOnSendFailure' -v
--- PASS: TestQueueEnqueuesOnSendFailure (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.411s

$ go test ./internal/feedback/ -run 'TestQueueResolvesOnSuccess' -v
--- PASS: TestQueueResolvesOnSuccess (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.422s
```

경계 테스트 3건(AC 외 추가): `TestQueueRefusesBlockedResult`(차단 판정은 큐에 들어가지 못함) · `TestQueueNeverReadsPreScrubDraft`(D4 경계) · `TestQueueMutateSerializesConcurrentEnqueues`(잠금이 read-modify-write 전체를 덮음). 전부 PASS.

패키지 트리 전체 (`go test -count=1 ./internal/feedback/... ./internal/sandbox/...`):

```
ok  	github.com/modu-ai/moai-adk/internal/feedback	0.693s
ok  	github.com/modu-ai/moai-adk/internal/sandbox	0.267s
```

`go vet ./internal/feedback/... ./internal/sandbox/...` · `GOOS=windows go vet ./internal/feedback/... ./internal/sandbox/...` 둘 다 무출력, exit 0.
`go test -race -count=1 ./internal/feedback/` → `ok ... 1.890s`.
`golangci-lint run --timeout=3m ./internal/feedback/... ./internal/sandbox/...` → `0 issues.`
`gofmt -l internal/feedback/` → 무출력.
`git status --short` → 신규 4파일 + `scrub.go` 수정만. 실제 트리의 `.moai/logs/` 에는 아무것도 쓰이지 않았다(테스트는 전부 `t.TempDir()` 루트를 명시적으로 전달).

AC-F-015 판정: **PASS** — `go test ./internal/feedback/ -run 'TestMaskLogRecordsKindAndCountWithoutRawValue' -v` 가 위 블록을 출력.
AC-F-016 판정: **PASS** — `go test ./internal/feedback/ -run 'TestMaskLogFailureIsFailOpen' -v` 가 위 블록을 출력.
AC-F-017 판정: **PASS** — `go test ./internal/feedback/ -run 'TestQueueEnqueuesOnSendFailure' -v` 가 위 블록을 출력.
AC-F-018 판정: **PASS** — `go test ./internal/feedback/ -run 'TestQueueResolvesOnSuccess' -v` 가 위 블록을 출력.

납품물:
- `internal/feedback/masklog.go` — `.moai/logs/feedback-mask.log` 잠금 없는 append, `0o600`, 전 실패 경로 `slog.Warn` 강등. 엔트리는 `RFC3339 | total=N | kind=… where=… count=N` 한 줄.
- `internal/feedback/queue.go` — `.moai/state/feedback/queue.json` 단일 JSON + 형제 lock(`queue.lock`) + `Mutate()` read-modify-write + `atomicfile.Replace`, `0o600`. `EnqueueMasked` / `Resolve` / `Load` / `QueuePathForRoot`.
- `internal/feedback/masklog_test.go` · `internal/feedback/queue_test.go` — AC 4건 + 경계 3건.
- `internal/feedback/scrub.go` — `appendMaskLog` 호출 1줄 + `Options.ProjectRoot` 주석 정정(아래 판단 1).

설계 판단 5건:

1. **프로젝트 루트 해석은 `Scrub` 이 아니라 CLI 경계에 둔다.** M2 의 `ProjectRoot` 주석은 "빈 값이면 작업 디렉터리에서 상향 탐색"이라고 적었으나, 그대로 구현하면 프로젝트 아래에서 일어나는 **모든** `Scrub` 호출이 그 프로젝트의 `.moai/` 에 쓴다 — M2·M3 테스트가 실제 리포의 `.moai/logs/` 에 로그를 남긴다(CLAUDE.local.md §6 위반). 빈 값 = **산출물 없음**으로 바꾸고, `--root` 해석과 `ResolveProjectRoot` 폴백은 M5 의 CLI 배선이 진다. `ResolveProjectRoot` 는 그대로 export 되어 있어 M5 가 그대로 쓴다.
2. **두 산출물의 형태를 통일하지 않는다.** 로그는 잠금 없는 append(인터리빙은 한 줄 훼손, 잠금은 스크럽을 막을 수 있으므로 fail-open 위반), 큐는 잠금 있는 단일 JSON(성공 시 **삭제**를 표현해야 하므로 append-only 불가 — AP-7). `design.md` §5 그대로다.
3. **`EnqueueMasked` 는 문자열 두 개가 아니라 `Result` 를 받는다.** `Result` 는 구조적으로 스크러버의 산출물이므로, 원문을 큐에 넣는 호출 형태가 아예 존재하지 않는다. 여기에 `verdict != ok` 거부를 더해, 차단된 보고가 재전송 경로로 우회 게시되는 길을 막는다(`ErrQueueBlockedResult`).
4. **D4 경계는 주석이 아니라 동작으로 관측한다.** 같은 `.moai/state/` 트리에 `feedback-draft-<ts>.md` 를 두고 `Load()` 가 그것을 항목으로 읽지 않음을 단언한다(`TestQueueNeverReadsPreScrubDraft`). 소스에 `feedback-draft` 문자열이 없다는 grep 은 base 에서도 0 이라 아무것도 관측하지 못하므로 채택하지 않았다.
5. **잠금 프리미티브는 `atomicfile.Claim` 을 재사용한다.** 리포의 "정확히 한 호출자만 진행" 프리미티브이고 POSIX·Windows 양쪽에서 원자적이다. `internal/kanban` 의 잠금은 보드/백로그 경로에 묶여 있어(`AcquireBoardLock(root)`) 임의 경로에 재사용할 수 없다 — 형태만 따르고 코드는 신규(`research.md` §2 의 판정 그대로).

미검증(M4 시점): 잠금 경합 상한(25ms × 40 ≈ 1s)은 4-goroutine × 3 회 테스트에서만 관측했고 다중 **프로세스** 경합은 관측하지 않았다. `Resolve` 는 아직 프로덕션 호출자가 없다 — M5 의 큐 동사(`queue resolve`)가 그 소비자다. 잠금 아티팩트는 프로세스가 비정상 종료하면 남으며(POSIX flock 과 달리 자동 해제되지 않는다), stale-lock 정리 경로는 이 SPEC 범위 밖이다.

### M5 — CLI 배선

Pre-flight (편집 전 트리 `55dc0ec0a`, base `3210da7d3` 대상):

```
$ git ls-tree -r 3210da7d3 --name-only -- internal/cli/feedback.go internal/cli/feedback_test.go | wc -l
       0
$ git show 3210da7d3:internal/cli/root.go | grep -c "newFeedbackCmd"
0
$ git show 55dc0ec0a:internal/cli/root.go | grep -c "newFeedbackCmd"
0
```

`moai feedback` 은 base 에도 M4 트리에도 존재하지 않았다. M4 가 남긴 두 소비자 공백(`Options.ProjectRoot` 해석, `QueueStore.Resolve`)을 M5 가 메운다.

RED (`go test ./internal/cli/ -run 'TestFeedbackScrubContract|TestFeedbackScrubToolFailureExitsNonZero' -v`, 테스트만 있고 구현 없는 트리):

```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/feedback_test.go:49:9: undefined: newFeedbackCmd
internal/cli/feedback_test.go:129:52: undefined: feedbackSecurityFileName
internal/cli/feedback_test.go:197:14: undefined: resolveFeedbackRoot
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

GREEN — AC별 명령을 각각 1회씩 실행해 관측한 결과:

```
$ go test ./internal/cli/ -run 'TestFeedbackScrubContract' -v
=== RUN   TestFeedbackScrubContract
--- PASS: TestFeedbackScrubContract (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.647s

$ go test ./internal/cli/ -run 'TestFeedbackScrubToolFailureExitsNonZero' -v
=== RUN   TestFeedbackScrubToolFailureExitsNonZero
--- PASS: TestFeedbackScrubToolFailureExitsNonZero (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.643s
```

경계 테스트 7건(AC 외 추가), 전부 PASS: `TestFeedbackScrubBlockedVerdictExitsZero`(P3 축 분리 — 차단은 종료 코드 0) · `TestFeedbackScrubWritesMaskLogUnderRoot`(M4 잔여 위험 1 — 산출물이 실제로 생기는지) · `TestFeedbackRootFallbackWalksUpToMarker`(`--root` 미지정 시 상향 탐색) · `TestFeedbackQueueVerbsRoundTrip`(enqueue→list→resolve) · `TestFeedbackQueueRefusesBlockedResult` · `TestFeedbackQueueNeverReadsPreScrubDraft`(D4 경계의 CLI 판) · `TestFeedbackQueueRequiresRoot`.

바이너리 스모크 (`go build -o /tmp/claude-501/moai-m5 ./cmd/moai`, exit 0):

```
$ SMOKE_SECRET_NAME=smoke-value-0123456789 moai-m5 feedback scrub \
    --root <scratch> --title 'crash while handling smoke-value-0123456789' < body.txt
{
  "verdict": "ok",
  "title": "crash while handling s...6789",
  "body": "body mentions s...6789 twice: s...6789\n",
  "findings": [
    { "kind": "env", "where": "title", "count": 1 },
    { "kind": "env", "where": "body",  "count": 2 }
  ],
  "reason": ""
}
scrub_exit=0

$ ls -l <scratch>/.moai/logs/feedback-mask.log
-rw-------  1 goos  wheel  97 ...
$ cat <scratch>/.moai/logs/feedback-mask.log
2026-08-23T18:55:11+09:00 | total=3 | kind=env where=title count=1 | kind=env where=body count=2

$ moai-m5 feedback queue enqueue --root <scratch> < result.json
{ "id": "f1", "title": "crash while handling s...6789", ..., "attempts": 0 }
enqueue_exit=0
$ ls -l <scratch>/.moai/state/feedback/queue.json
-rw-------  1 goos  wheel  253 ...
$ moai-m5 feedback queue list --root <scratch>      → items 1건, list_exit=0
$ moai-m5 feedback queue resolve f1 --root <scratch> → { "id": "f1", "removed": true }, resolve_exit=0

$ echo 'hello' | moai-m5 feedback scrub --root <scratch> --title 'a title' \
    | jq -e '.verdict and (.title|type=="string") and (.findings|type=="array")'
true
jq_smoke_exit=0

$ (security.yaml 을 깨뜨린 뒤) moai-m5 feedback scrub --root <scratch> --title 't' < body.txt
exit=1 · stdout 0 바이트 · stderr 272 바이트
```

`--title` 을 실제로 통과시킨다는 관측은 위 스모크의 `findings[where=title]` 이다(계약 필드 존재가 아니라 마스킹 동작). 마스킹된 값 형태(`s...6789`)는 기존 마스커 그대로다.

패키지 트리 전체 (`go test -timeout 900s ./internal/cli/... ./internal/feedback/...`):

```
ok  	github.com/modu-ai/moai-adk/internal/cli	343.442s
ok  	github.com/modu-ai/moai-adk/internal/feedback	5.636s
… (하위 패키지 전부 ok)
--- FAIL: TestConstitutionCrossReference (0.00s)
    agent_lint_test.go:1249: moai-constitution.md should cross-reference agent-authoring.md for effort matrix
FAIL	github.com/modu-ai/moai-adk/internal/cli/agentlint	1.194s
```

**이 FAIL 1건은 M5 소관이 아니다** — 단언 대상이 `.claude/rules/moai/core/moai-constitution.md` 이고(`grep -c agent-authoring` → `0`), M5 diff 는 Go 파일 3개뿐이다(`git status --porcelain` → ` M internal/cli/root.go` + 신규 2). M1~M4 도 이 패키지를 돌리지 않았으므로 이번에 처음 보인 것이지 이번에 생긴 것이 아니다. 규칙 파일 부채로 상위에 보고한다.

`go vet ./internal/cli/... ./internal/feedback/...` → exit 0. `GOOS=windows go vet` 동일 대상 → exit 0.
`go test -race ./internal/feedback/` → `ok ... 1.967s`.
`golangci-lint run --timeout=3m ./internal/cli/...` → `0 issues.`
`gofmt -l internal/cli/feedback.go internal/cli/feedback_test.go internal/cli/root.go` → 무출력.
실제 리포 트리의 `.moai/logs/` 에는 아무것도 쓰이지 않았다(테스트는 `t.TempDir()`, 스모크는 `/tmp` scratch).

AC-F-003 판정: **PASS** — `go test ./internal/cli/ -run 'TestFeedbackScrubContract' -v` 가 위 블록을 출력. 5필드 + `findings[].where` 가 `title`·`body` 양쪽에서 관측됨.
AC-F-004 판정: **PASS** — `go test ./internal/cli/ -run 'TestFeedbackScrubToolFailureExitsNonZero' -v` 가 위 블록을 출력. 바이너리에서도 exit 1 + stdout 0 바이트로 재관측.

납품물:
- `internal/cli/feedback.go` — `feedback` 부모(`--root` persistent) + `scrub`(`--title`, 본문 stdin, JSON stdout) + `queue enqueue|list|resolve`. `resolveFeedbackRoot` / `feedbackScrubOptions` / `emitFeedbackJSON`.
- `internal/cli/root.go` — `rootCmd.AddCommand(newFeedbackCmd())` 1줄 + 주석.
- `internal/cli/feedback_test.go` — AC 2건 + 경계 7건.

설계 판단 5건:

1. **`--root` 는 신뢰하지 않고 검증한다.** 오타 난 경로가 조용히 상향 탐색으로 폴백하면 **다른 프로젝트**의 `.moai/` 에 쓴다. 명시 `--root` 는 `.moai` 존재를 확인하고 아니면 실패(fail-closed), 미지정만 `ResolveProjectRoot` 로 폴백한다. 사용자 경로는 `filepath.Abs` 로만 해석한다(`filepath.Join(cwd, userPath)` 금지 — CLAUDE.local.md §6).
2. **루트 미해석의 처리는 동사마다 다르다.** `scrub` 은 루트가 없어도 진행한다(마스킹은 산출물에 의존하지 않는다 — fail-open). `queue` 동사는 쓰기가 존재 이유이므로 루트 없으면 에러다. 두 축을 섞지 않는다.
3. **정책 로드는 탐지기와 반대로 fail-closed 다.** `security.LoadExtraSecurityConfig` 는 파싱 실패를 기본값으로 강등한다 — 탐지기에서는 놓친 패턴 1건이 "쓰기 1건 허용"이지만, 여기서는 사용자가 설정한 확장 패턴이 조용히 빠진 채 **공개 채널로 나간다**. 그래서 M5 는 같은 파일을 직접 읽고 읽기·파싱 실패를 도구 실패로 올린다. AC-F-004 의 "정책을 로드할 수 없는 조건"이 이것이다.
4. **`queue enqueue` 는 제목/본문이 아니라 스크러버 JSON 을 통째로 받는다.** M4 의 `EnqueueMasked(Result)` 논리를 CLI 표면까지 연장한 것으로, 원문 문자열을 큐에 넣는 **호출 형태 자체가 존재하지 않는다**. D4 경계도 같은 이유로 유지된다 — 큐 동사는 `queue.json` 외에 아무것도 읽지 않으며, 초안(`feedback-draft-*.md`)이 같은 트리에 있어도 항목으로 채택되지 않음을 `TestFeedbackQueueNeverReadsPreScrubDraft` 가 동작으로 관측한다.
5. **`env_scrub_extra` 만 로컬 구조체로 선언하고 패턴 확장은 exported 타입을 재사용한다.** 같은 파일을 두 번 언마샬해 (a) `security.ExtraSecurityConfig`(패턴), (b) `security.sandbox.env_scrub_extra` 를 얻는다. 패턴 문자열은 어디에도 복사하지 않는다(AP-4).

미검증(M5 시점): `--root` 미지정 폴백은 단위 테스트(`resolveFeedbackRoot("", nested)`)와 앰비언트 실행으로만 관측했고, 스모크의 산출물 검증은 명시 `--root` 로 했다. `moai feedback` 을 실제로 호출하는 프로덕션 소비자는 아직 없다 — 스킬 본문(M6)이 그 소비자다. `queue` 동사는 재전송을 **수행하지 않는다**(적재/조회/제거만); 재전송 자체는 스킬 본문의 `gh issue create` 재실행이다.

### M6 — 스킬 본문 + 마법사 질문

착수 HEAD `d2063308b`, base `3210da7d3`, 브랜치 `WT-auto-feedback`, 워크트리 `.claude/worktrees/t170`.

#### 부채 해소 — AC-F-019 토큰 2건 재작성 (판정보다 먼저)

plan.md §F 의 `[HARD] M6 착수 시 먼저 처리할 부채` 2건을 구현 전에 처리했다. base 실측은 소스(`.claude/skills/moai/workflows/feedback.md`)와 템플릿 미러 양쪽에 대해 `git show 3210da7d3:<path>` 로 각각 수행했다.

```
$ git show 3210da7d3:.claude/skills/moai/workflows/feedback.md > base_src.md
$ git show 3210da7d3:internal/template/templates/.claude/skills/moai/workflows/feedback.md > base_tpl.md

$ grep -c -- '--title' base_src.md base_tpl.md                                   # 구 ②
base_tpl.md:0
base_src.md:0
$ grep -cE 'moai feedback scrub[^\n]*--title' base_src.md base_tpl.md            # 신 ②
base_tpl.md:0
base_src.md:0
$ grep -c 'MUST NOT submit' base_src.md base_tpl.md                              # 신 ③-a
base_tpl.md:0
base_src.md:0
$ grep -c '60 seconds' base_src.md base_tpl.md                                   # 신 ③-b
base_tpl.md:0
base_src.md:0
```

**부채 1 — ②의 앵커 부재.** 구 토큰 `grep -c -- '--title'` 는 base 에서 **0/0** 을 반환한다. 즉 "base 0" 이라는 채택 조건 자체는 만족하지만 **의도한 것을 관측하지 못한다**: 파일 어디의 `--title` 이든 잡으므로, 본문만 스크럽하고 `gh issue create ... --title "<제목>"` 만 쓰는 구현이 통과한다 — 스크러버가 제목을 못 받은 채로. 스크러버 호출과 **동일 줄**을 요구하는 `grep -cE 'moai feedback scrub[^\n]*--title'` 로 교체했다(base **0/0**). 이 토큰은 제목이 실제로 스크러버 입력으로 들어가는 줄이 존재해야만 통과한다.

**부채 2 — ③의 한국어 리터럴.** 보조 검사 `grep -n '종료 코드\|verdict\|60초'` 는 **영어 전용 표면**(템플릿 미러)을 한국어 토큰으로 겨냥한다 — 통과시키려면 배포 템플릿의 언어 규칙을 어겨야 하는 자기모순이다. 형제 SPEC `SPEC-TODO-ENABLE-FLAG-001`(acceptance.md `N1`)이 같은 형태를 한국어 리터럴 제거 + 행동 관측 위임으로 해소한 선례를 따라, 영어 토큰 2종으로 교체했다: `grep -c 'MUST NOT submit'` >= **3**(세 조항이 각각 금지를 명시했는지를 **건수**로 센다 — 한 문장만 쓴 본문은 3에 못 미쳐 FAIL 한다)과 `grep -c '60 seconds'` >= 1. 둘 다 base **0/0**. 세 축의 행동 관측은 AC-F-004(도구 실패 → exit ≠ 0)와 AC-F-003(`verdict` 계약)이 이미 지고 있다.

acceptance.md 의 AC-F-019 토큰 목록과 실측표를 위 내용으로 갱신했다. **이는 "manager-spec 이 acceptance.md 를 소유한다"는 경계의 승인된 예외다** — plan.md §F 가 이 재작성을 M6 에 명시적으로 배정했다(`[HARD] M6 착수 시 먼저 처리할 부채`). 부채 문단 자체는 지우지 않고 처리 결과를 덧붙였다.

#### RED

마법사 두 축 모두 구현 심볼 부재로 build failed 상태를 먼저 관측했다.

```
$ go test ./internal/cli/wizard/ -run 'TestFeedbackAutoSubmit' -v
# github.com/modu-ai/moai-adk/internal/cli/wizard [github.com/modu-ai/moai-adk/internal/cli/wizard.test]
internal/cli/wizard/feedback_auto_submit_test.go:38:7: r.FeedbackAutoSubmit undefined (type *WizardResult has no field or method FeedbackAutoSubmit)
internal/cli/wizard/feedback_auto_submit_test.go:41:8: r.FeedbackAutoSubmit undefined (type *WizardResult has no field or method FeedbackAutoSubmit)
internal/cli/wizard/feedback_auto_submit_test.go:47:8: r2.FeedbackAutoSubmit undefined (type *WizardResult has no field or method FeedbackAutoSubmit)
internal/cli/wizard/feedback_auto_submit_test.go:50:10: r2.FeedbackAutoSubmit undefined (type *WizardResult has no field or method FeedbackAutoSubmit)
internal/cli/wizard/feedback_auto_submit_test.go:56:8: r3.FeedbackAutoSubmit undefined (type *WizardResult has no field or method FeedbackAutoSubmit)
FAIL	github.com/modu-ai/moai-adk/internal/cli/wizard [build failed]
FAIL

$ go test ./internal/core/project/ -run 'TestWritePhase1ConfigsPersistsFeedbackAutoSubmit|TestWritePhase1ConfigsSkipsFeedbackWhenUnanswered|TestWritePhase1ConfigsFeedbackNoFile' -v
# github.com/modu-ai/moai-adk/internal/core/project [github.com/modu-ai/moai-adk/internal/core/project.test]
internal/core/project/initializer_feedback_test.go:34:50: undefined: defs.FeedbackYAML
internal/core/project/initializer_feedback_test.go:42:3: unknown field FeedbackAutoSubmit in struct literal of type InitOptions
internal/core/project/initializer_feedback_test.go:79:50: undefined: defs.FeedbackYAML
internal/core/project/initializer_feedback_test.go:108:3: unknown field FeedbackAutoSubmit in struct literal of type InitOptions
internal/core/project/initializer_feedback_test.go:114:55: undefined: defs.FeedbackYAML
FAIL	github.com/modu-ai/moai-adk/internal/core/project [build failed]
FAIL
```

#### GREEN — AC별 판정

**AC-F-002 — 확인 게이트 존재 (편집 전 FAIL / 편집 후 PASS 양쪽 관측)**

편집 전(base `3210da7d3`), 소스 사본:

```
$ grep -n 'AskUserQuestion\|gh issue create\|auto_submit' base_src.md
52:[HARD] Collect the feedback fields — type, title, and description — in ONE AskUserQuestion round …
84:Inputs for the `gh issue create` invocation:
118:The orchestrator executes directly: `gh issue create --repo <resolved-target>` …
156:Use AskUserQuestion after successful submission:
178:- Phase 1: MoAI orchestrator (AskUserQuestion for feedback collection)
```

`auto_submit` 은 **0건**이고 `gh issue create`(:84) 앞의 AskUserQuestion(:52)은 필드 수집 라운드다 — 조건부 게이트가 아니다. **편집 전 FAIL 관측 완료.**

편집 후:

```
$ grep -n 'AskUserQuestion\|gh issue create\|auto_submit' .claude/skills/moai/workflows/feedback.md
52:[HARD] Collect the feedback fields — type, title, and description — in ONE AskUserQuestion round …
108:[HARD] When the resolved `feedback.auto_submit` configuration value is `false` (the shipped default), run ONE AskUserQuestion round before submission. Skip this round only when `auto_submit` is `true`.
143:Inputs for the `gh issue create` invocation:
177:The orchestrator executes directly: `gh issue create --repo <resolved-target>` …
217:Use AskUserQuestion after successful submission:
239:- Phase 1: MoAI orchestrator (AskUserQuestion for feedback collection)
```

`auto_submit` 조건부 AskUserQuestion 이 `:108` — 첫 `gh issue create`(`:143`)보다 앞이다. 판정: **PASS**.

**AC-F-019 — 스킬 [HARD] 조항 (두 사본 각각 전수)**

재작성된 7개 관측을 사본별로 돌린 결과(`SRC`=소스, `TPL`=템플릿 미러):

| 관측 | 토큰 | SRC | TPL | 하한 | 판정 |
|---|---|---|---|---|---|
| ① 스크러버 경유 | `moai feedback scrub` | 1 | 1 | 1 | PASS |
| ② 제목 전달(동일 줄 앵커) | `moai feedback scrub[^\n]*--title` | 1 | 1 | 1 | PASS |
| ③ fail-closed 조항 | `verdict` | 3 | 3 | 1 | PASS |
| ③-a 3문장 금지 절 | `MUST NOT submit` | 3 | 3 | 3 | PASS |
| ③-b 타임아웃 문장 | `60 seconds` | 1 | 1 | 1 | PASS |
| ④ 실패 부류 분기 | `queue.json` | 1 | 1 | 1 | PASS |
| ⑤ 게이트 라벨 언어 | `label[^*]*conversation_language\|conversation_language[^*]*label` | 1 | 1 | 1 | PASS |

verbatim 예외 문장이 ①과 같은 절 범위에 있음(두 사본 동일 줄 번호):

```
$ grep -n 'moai feedback scrub\|ONE explicit exception\|### Step 1: Route' <각 사본>
82:### Step 1: Route the report through the scrubber
87:printf '%s' "<body>" | moai feedback scrub --title "<title>"
94:[HARD] This masking is the ONE explicit exception to the verbatim-preservation rule stated under Issue Language Policy below …
```

두 사본 드리프트는 확대되지 않았다(기존 1줄 그대로):

```
$ diff .claude/skills/moai/workflows/feedback.md internal/template/templates/.claude/skills/moai/workflows/feedback.md
245d244
< Last Updated: 2026-02-07
```

템플릿 중립성:

```
$ grep -rnE 'SPEC-[A-Z]|REQ-[A-Z0-9]|AC-[A-Z]|CLAUDE\.local' internal/template/templates/.claude/skills/moai/workflows/feedback.md
(무출력)
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'Leak|Neutral|Internal' -v   → ok … 1.146s
```

판정: **PASS**.

**AC-F-020 · AC-F-021 — 마법사 질문 + 개수 고정 + 번역 완전성**

```
$ go test ./internal/cli/wizard/ -run 'TestFeedbackAutoSubmitQuestion|TestSaveBoolAnswerFeedbackAutoSubmit|TestFeedbackAutoSubmitTranslationsExist|TestQuestionOrder|TestReconfigureQuestions|TestWizardQuestionTranslationCompleteness' -v
=== RUN   TestFeedbackAutoSubmitQuestion
--- PASS: TestFeedbackAutoSubmitQuestion (0.00s)
=== RUN   TestSaveBoolAnswerFeedbackAutoSubmit
--- PASS: TestSaveBoolAnswerFeedbackAutoSubmit (0.00s)
=== RUN   TestFeedbackAutoSubmitTranslationsExist
--- PASS: TestFeedbackAutoSubmitTranslationsExist (0.00s)
=== RUN   TestQuestionOrder
--- PASS: TestQuestionOrder (0.00s)
=== RUN   TestReconfigureQuestionsOrder
--- PASS: TestReconfigureQuestionsOrder (0.00s)
=== RUN   TestWizardQuestionTranslationCompleteness
--- PASS: TestWizardQuestionTranslationCompleteness (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	0.471s
```

`-run` 선택자 실재 확인(판정보다 먼저):

```
$ grep -c 'func TestQuestionOrder\|func TestReconfigureQuestionsOrder\|func TestWizardQuestionTranslationCompleteness' internal/cli/wizard/questions_test.go internal/cli/wizard/translations_completeness_test.go
internal/cli/wizard/translations_completeness_test.go:1
internal/cli/wizard/questions_test.go:2
```

판정: AC-F-020 **PASS**, AC-F-021 **PASS**.

**AC-F-022 — 답변이 실제로 파일에 기록됨**

```
$ go test ./internal/core/project/ -run 'TestWritePhase1ConfigsPersistsFeedbackAutoSubmit|TestWritePhase1ConfigsSkipsFeedbackWhenUnanswered|TestWritePhase1ConfigsFeedbackNoFile' -v
--- PASS: TestWritePhase1ConfigsSkipsFeedbackWhenUnanswered (0.00s)
--- PASS: TestWritePhase1ConfigsFeedbackNoFile (0.01s)
--- PASS: TestWritePhase1ConfigsPersistsFeedbackAutoSubmit (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/core/project	0.418s
```

첫 테스트는 grep 이 아니라 **실제 로더**(`config.NewLoader().Load(...)` → `cfg.FeedbackAutoSubmit()`)로 되읽어 판정한다 — 런타임 표면이 답을 보는지가 AC 의 요지이기 때문이다. 판정: **PASS**.

#### 회귀 + 정적 검사

```
$ go test -timeout 900s ./internal/cli/wizard/... ./internal/core/project/... ./internal/defs/...
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	3.264s
ok  	github.com/modu-ai/moai-adk/internal/core/project	1.245s
ok  	github.com/modu-ai/moai-adk/internal/defs	0.547s

$ go test -timeout 900s ./internal/cli/...        → 전부 ok, 단 1건:
--- FAIL: TestConstitutionCrossReference (0.00s)
    agent_lint_test.go:1249: moai-constitution.md should cross-reference agent-authoring.md for effort matrix
FAIL	github.com/modu-ai/moai-adk/internal/cli/agentlint	0.443s

$ go test -timeout 900s ./internal/template/...   → 전부 ok
$ go build ./...                                  → exit 0
$ go vet ./internal/cli/... ./internal/core/project/... ./internal/defs/...              → exit 0
$ GOOS=windows go vet ./internal/cli/... ./internal/core/project/... ./internal/defs/... → exit 0
$ golangci-lint run --timeout=3m ./internal/cli/wizard/... ./internal/core/project/... ./internal/defs/... → 0 issues.
$ make build                                      → exit 0 (catalog.yaml 재생성 + bin/moai 재빌드)
$ gofmt -l <M6 편집 파일>                          → 무출력
```

**FAIL 1건은 M6 소관이 아니다.** `TestConstitutionCrossReference` 는 `.claude/rules/moai/core/moai-constitution.md` 에 `agent-authoring.md` 인용을 요구하는데, 그 줄은 커밋 `243eb07ef`(별도 카드, 이미 `release/v3.1.3` 에 착지)가 제거했다. base `3210da7d3` 에서도 붉으며 M5 보고서가 같은 FAIL 을 이미 기록했다.

#### 개수 고정 테스트 4건 갱신 (질문 1개 추가의 기계적 귀결)

`feedback_auto_submit` 이 Page-3 에 들어가면서 페이지 인원수를 고정한 기존 테스트 4건이 붉어졌고, 숫자/멤버십만 최소 편집했다(기존 항목 재배치·재서식 없음 — plan.md AP-10):

- `wizard_test.go:659` `TestStepperTotal_DynamicDenominator` — 16 → 17
- `expansion_test.go:39` `TestPage3QuestionsStructure` — want 목록에 `feedback_auto_submit` 1줄 추가(위치: `todo_enabled` 다음)
- `expansion_test.go:267,277` `TestTotalVisibleQuestions_Page3AlwaysCounted` — 15 → 16, Quality & Workflow 9 → 10
- `restructure_test.go:84` `TestInitPages_Membership` — 멤버십 목록에 `feedback_auto_submit` 추가

plan.md §B Known Issues 3 이 지키라고 한 두 고정(`DefaultQuestions` 5개 / `ReconfigureQuestions` 12개)은 **건드리지 않았고 그대로 초록이다** — 신규 질문은 `Page3Questions` 에만 들어갔다(AP-2 회피).

#### 납품물

스킬 2사본(동일 내용, 기존 1줄 드리프트 유지):
- `.claude/skills/moai/workflows/feedback.md` + `internal/template/templates/.claude/skills/moai/workflows/feedback.md` — 신규 `## Phase 1.5: Pre-Submission Scrub and Confirmation Gate` 4스텝(스크러버 경유 + verbatim 예외 / fail-closed 3문장 / 3옵션 확인 게이트 / 제출·큐잉 분기 + 재전송 전 재스크럽), 그리고 `Issue Creation Command` 에 "마스킹된 값만 넘긴다" [HARD] 1줄.

마법사 살아 있는 경로(AP-1 회피 — `init_autonomy_wizard.go` 계열은 손대지 않았다):
- `internal/cli/wizard/questions.go` — `Page3Questions` "Quality & Workflow" 에 `feedback_auto_submit` (Confirm, Default `"false"`)
- `internal/cli/wizard/types.go` — `WizardResult.FeedbackAutoSubmit *bool`
- `internal/cli/wizard/wizard.go` — `saveBoolAnswer` 분기
- `internal/cli/wizard/translations.go` — ko/ja/zh 3로케일
- `internal/cli/init.go` — `applyWizardPage3ToOpts` 에서 `opts.FeedbackAutoSubmit = result.FeedbackAutoSubmit`
- `internal/core/project/initializer.go` — `InitOptions.FeedbackAutoSubmit *bool`
- `internal/core/project/initializer_expansion.go` — `writeFeedbackAutoSubmitYAML` + `WritePhase1Configs` 호출
- `internal/defs/files.go` — `FeedbackYAML` 상수 1줄 추가
- 신규 테스트 2파일 — `internal/cli/wizard/feedback_auto_submit_test.go`(3종 세트) · `internal/core/project/initializer_feedback_test.go`(3종)

#### 설계 판단 3건

1. **`*bool` 을 쓴다(형제 `todo_enabled` 선례).** `worktree_auto_create` 는 `bool` + 별도 `*Set` 추적 필드 방식인데, 포인터 하나면 "물었는가"와 "답이 무엇인가"를 동시에 담아 필드가 하나 줄어든다. `--non-interactive` 는 nil 이고, nil 이면 배포된 `feedback.yaml` 을 **바이트 단위로 그대로 둔다** — 기본값 `false` 를 굳이 다시 적어 모든 프로젝트 설정에 줄을 하나 늘릴 이유가 없다.
2. **쓰기는 `yamlpatch` 로 upsert 한다.** 배포되는 `feedback.yaml` 은 `repository` 키만 담고 `auto_submit` 이 없다(그 추가는 M8 소관). 키가 이미 있어야 동작하는 패치 헬퍼를 쓰면 **정확히 문제가 되는 경로에서 조용한 no-op** 이 된다. `yamlpatch` 는 중첩 매핑을 upsert 하면서 주석과 키 순서를 보존하고, 테스트가 그 보존을 단언한다.
3. **게이트 라벨 예시는 두 사본 모두 영어다.** design.md §7 의 한국어 표는 "의미를 적은 것이지 리터럴이 아니다"이고, D11 은 템플릿 미러 예시를 영어로 요구한다. 두 사본을 동일하게 유지해야 드리프트가 늘지 않으므로 소스도 영어로 쓰고, "이 라벨을 로케일 세션에 그대로 복사하지 말라"는 지시를 [HARD] 로 함께 실었다.

### M7 — 웹 콘솔 노출

착수 HEAD `38705eb85`, base `3210da7d3`, 브랜치 `WT-auto-feedback`, 워크트리 `.claude/worktrees/t170`.

#### 결정 D5 반전을 명시적으로 기록한다

`feedback` 섹션의 `RouteExcluded` 고정은 **낡은 테스트가 아니라 기록된 결정**이다 — SPEC-WEBCONF-SIMPLIFY-001 M3 이 이 섹션의 탭과 웹 쓰기 경로를 의도적으로 제거했고, 두 테스트가 그것을 고정하고 있었다. 이번 편집이 정당한 유일한 근거는 `spec.md` §D 결정 D5 가 그 결정을 **선택지 A로 명시적으로 뒤집었다**는 사실이며(REQ-12), 반전은 커밋 본문에 이름과 함께 남긴다(plan.md AP-9).

**고정 테스트는 카드가 지목한 2건이 아니라 6건이었다.** 카드는 `sectionroute_test.go:27` 과 `scope_contract_test.go:79` 를 지목했는데, 실측해 보니 같은 결정을 고정하는 지점이 4곳 더 있었다. 넷 다 M3 재분류 목록을 그대로 복사한 형태이므로 같은 반전의 기계적 귀결이며, 범위 확장이 아니다:

| 파일 | 지점 | 편집 |
|---|---|---|
| `internal/settings/sectionroute_test.go` | `:27` 외 3함수 | `feedback` → `RouteSeam`, `SeamSections` 2개, `ExcludedSections` 18 → 17 |
| `internal/web/scope_contract_test.go` | `:79` 외 1함수 | 제외 목록 → seam 단언으로 이동 |
| `internal/settings/sectionwrite_test.go` | `:67`, `:158` | seam 쓰기 거부 목록 2곳에서 제거 |
| `internal/web/tab_layout_test.go` | `:18` | `wantTabOrder` 에 `feedback` 12번째 추가 |
| `internal/web/schema_sections_test.go` | `:112`, `:119` | 렌더 금지 접두사 목록 → 렌더 필수 목록으로 이동 |
| `internal/web/schema_render_test.go` | `:73` | `m3RemovedSections` 면제에서 제거 |

마지막 둘과 `schema_sections_test.go:29` 의 `m3ReclassifiedSeamSections` 는 **면제 목록**이다 — `feedback` 을 남겨두면 재개방된 섹션의 라우트·렌더 불변식이 조용히 검사되지 않으므로 함께 제거했다.

#### 형제 SPEC 공유 파일 확인

`SPEC-TODO-ENABLE-FLAG-001` 이 먼저 착지했는지 base 에서 확인했다: `ExcludedSections()` 에 `feedback` 이 그대로 남아 있었고(RED 출력 `RouteForSection("feedback") = 0`), 중복 제거는 발생하지 않았다. 공유 파일(`i18n.js`, `schema_sections.go`)에는 **신규 항목만 덧붙였고** 기존 항목의 위치·서식은 건드리지 않았다(plan.md AP-10).

#### RED — 구현 전 실패 관측

두 패키지 모두 구현 이전 트리에서 새 기대 방향으로 실패함을 먼저 확인했다.

```
$ go test ./internal/settings/ -run 'TestRouteForSectionTable|TestSeamSectionsMatchesRoutes|TestExcludedSectionsAllRejected|TestFeedbackAutoSubmitFieldRegistered|TestFeedbackSectionSeamWritable'
--- FAIL: TestRouteForSectionTable (0.00s)
    sectionroute_test.go:49: RouteForSection("feedback") = 0, want 2
--- FAIL: TestSeamSectionsMatchesRoutes (0.00s)
    sectionroute_test.go:62: SeamSections() length = 1, want 2 (workflow + feedback)
--- FAIL: TestFeedbackAutoSubmitFieldRegistered (0.00s)
    feedback_autosubmit_test.go:26: SectionFields(SectionFeedback) has no feedback.auto_submit field
--- FAIL: TestExcludedSectionsAllRejected (0.00s)
    sectionroute_test.go:94: ExcludedSections() length = 18, want 17 (19 M3 - workflow - feedback)
--- FAIL: TestFeedbackSectionSeamWritable (0.00s)
    feedback_autosubmit_test.go:56: ApplySchemaEdits(feedback.auto_submit): settings: unknown schema field "feedback.auto_submit"
FAIL	github.com/modu-ai/moai-adk/internal/settings	0.428s

$ go test ./internal/web/ -run 'TestScopeContractEditableSections|TestScopeContractExclusions|TestFeedbackPanelRendered|TestFeedbackAutoSubmitI18nKeysInAllLocales|TestFeedbackPanelFieldsWired'
--- FAIL: TestFeedbackPanelRendered (0.01s)
    feedback_panel_test.go:26: rendered console missing "data-tab=\"feedback\""
    feedback_panel_test.go:26: rendered console missing "data-panel=\"feedback\""
    feedback_panel_test.go:26: rendered console missing "name=\"feedback.repository\""
    feedback_panel_test.go:26: rendered console missing "name=\"feedback.auto_submit\""
    feedback_panel_test.go:26: rendered console missing "name=\"feedback.auto_submit__present\""
--- FAIL: TestFeedbackAutoSubmitI18nKeysInAllLocales (0.00s)
    feedback_panel_test.go:39: i18n.js missing feedback key "f.feedback.auto_submit.title" in all 4 locales
    feedback_panel_test.go:39: i18n.js missing feedback key "f.feedback.auto_submit.desc" in all 4 locales
--- FAIL: TestFeedbackPanelFieldsWired (0.00s)
    feedback_panel_test.go:49: schemaPanelMeta(feedback).ID = "", want "feedback"
    feedback_panel_test.go:53: SectionFields(SectionFeedback) = 1 fields, want >= 2 (repository + auto_submit)
--- FAIL: TestScopeContractEditableSections (0.00s)
    scope_contract_test.go:43: section "feedback": route = 0, want RouteSeam (M7 reopened)
FAIL	github.com/modu-ai/moai-adk/internal/web	0.462s
```

#### GREEN — AC-F-023 웹 절반 판정

**PASS.** AC-F-023 은 웹 절반(스키마·라우트·i18n)과 템플릿 절반(미러·인벤토리·`make build`·중립성)으로 나뉘며, M7 은 **웹 절반만** 진다. 템플릿 절반은 M8 소관이므로 이 판정은 그 부분을 주장하지 않는다.

acceptance.md 가 지정한 5개 선택자를 `-v` 로 돌려 각각 `=== RUN` 이 찍히는지 확인했다(0개 실행 통과 방지 — §D.3):

```
$ go test -count=1 ./internal/settings/ ./internal/web/ -run 'TestSchemaCurrentValuesReadsAllSections|TestI18nKeySetParity|TestRouteForSectionTable|TestExcludedSectionsAllRejected|TestScopeContract|TestFeedbackPanelRendered|TestFeedbackAutoSubmitFieldRegistered|TestFeedbackSectionSeamWritable|TestFeedbackAutoSubmitI18nKeysInAllLocales|TestFeedbackPanelFieldsWired' -v
=== RUN   TestFeedbackAutoSubmitFieldRegistered
=== RUN   TestFeedbackSectionSeamWritable
=== RUN   TestSchemaCurrentValuesReadsAllSections
=== RUN   TestRouteForSectionTable
=== RUN   TestExcludedSectionsAllRejected
--- PASS: TestRouteForSectionTable (0.00s)
--- PASS: TestExcludedSectionsAllRejected (0.00s)
--- PASS: TestFeedbackAutoSubmitFieldRegistered (0.00s)
--- PASS: TestFeedbackSectionSeamWritable (0.00s)
--- PASS: TestSchemaCurrentValuesReadsAllSections (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/settings	0.219s
=== RUN   TestFeedbackPanelRendered
--- PASS: TestFeedbackPanelRendered (0.01s)
=== RUN   TestFeedbackAutoSubmitI18nKeysInAllLocales
--- PASS: TestFeedbackAutoSubmitI18nKeysInAllLocales (0.00s)
=== RUN   TestFeedbackPanelFieldsWired
--- PASS: TestFeedbackPanelFieldsWired (0.00s)
=== RUN   TestI18nKeySetParity
--- PASS: TestI18nKeySetParity (0.05s)
=== RUN   TestScopeContractEditableSections
=== RUN   TestScopeContractExclusions
--- PASS: TestScopeContractEditableSections (0.00s)
--- PASS: TestScopeContractExclusions (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.455s
```

**섹션이 실제로 폼에 도달했다는 관측**은 주장이 아니라 두 지점의 렌더 단언이다: `TestFeedbackPanelRendered` 가 `data-tab="feedback"` · `data-panel="feedback"` · `name="feedback.auto_submit"` · `name="feedback.auto_submit__present"`(bool 의 unchecked↔미제출 구분 동반 필드)를 렌더된 HTML 에서 찾고, `internal/web/schema_sections_test.go` 의 렌더 스모크가 `feedback.` 접두사를 **금지 목록에서 필수 목록으로** 옮겼다. i18n 은 `TestFeedbackAutoSubmitI18nKeysInAllLocales`(신규 2키) + 기존 `TestI18nKeySetParity`(스키마 전 필드 4로케일 전수)가 함께 본다.

#### 전 패키지 회귀 + 빌드

```
$ go test -count=1 ./internal/settings/... ./internal/web/...
ok  	github.com/modu-ai/moai-adk/internal/settings	1.018s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	0.347s
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	0.692s
ok  	github.com/modu-ai/moai-adk/internal/web	2.899s

$ go vet ./internal/settings/... ./internal/web/...                  → exit 0 (무출력)
$ GOOS=windows go vet ./internal/settings/... ./internal/web/...      → exit 0 (무출력)
```

**M6 이 기록한 `TestConstitutionCrossReference` FAIL 은 M7 소관이 아니며 이 결과에 섞이지 않았다** — `internal/cli/agentlint` 패키지 소속이라 위 두 패키지 스코프 밖이고, base 에서도 붉다(원인: 커밋 `243eb07ef` 가 `moai-constitution.md` 에서 `agent-authoring.md` 인용을 제거).

#### 납품물

- `internal/settings/schema_sections.go` — `feedback.auto_submit` (TypeBool, seam) 1필드 추가. 기존 `feedback.repository` 줄은 그대로.
- `internal/settings/sectionroute.go` — `sectionRoutes` 에 `"feedback": RouteSeam` 추가, `SeamSections()` → `{workflow, feedback}`, `ExcludedSections()` 에서 `feedback` 제거(18 → 17).
- `internal/web/schemaform.go` — `consoleTabs()` 12번째 탭(기존 순서 무변경, 말미 추가) + `schemaSectionMetas()` 패널 메타. 아이콘은 기존 `messages-square` 케이스 재사용(신규 SVG 케이스 없음), 라벨은 기존 `sec.feedback.title`/`.desc` 키 재사용(신규 `sec.*` 키 0).
- `internal/web/assets/i18n.js` — `f.feedback.auto_submit.title`/`.desc` 를 en/ko/ja/zh **4로케일 전부**에 추가(각 로케일 `f.feedback.repository.*` 바로 뒤). 기존 항목은 무편집.
- 신규 테스트 2파일 — `internal/settings/feedback_autosubmit_test.go`(필드 등록 + seam 왕복) · `internal/web/feedback_panel_test.go`(패널 렌더 + 4로케일 + 배선).
- 고정 테스트 6파일 갱신 — 위 표.

`.templ` 편집은 없다. `root.templ` 의 패널 루프가 `consoleTabs()` 를 순회하며 미지정 탭을 `default:` 분기에서 `fieldsetSchemaSection` 으로 그리므로, 탭 등록만으로 렌더가 따라온다.

#### 설계 판단 2건

1. **탭을 말미에 붙인다.** 기존 11개 순서를 건드리지 않아 `wantTabOrder` 편집이 1줄로 끝나고, 형제 SPEC 과의 충돌 표면도 최소가 된다(AP-10).
2. **`sec.feedback.desc` 는 손대지 않는다.** 지금 값은 "피드백 워크플로우 대상 저장소"라 필드가 둘이 된 지금은 좁지만, 카드의 [HARD] 는 공유 파일에 **신규 항목만** 추가하라고 했다. 기존 4줄을 고치는 대신 Go 쪽 패널 `Desc` 베이스라인 문구를 두 필드를 포괄하도록 썼다. 좁아진 i18n 설명 4줄은 아래 잔여 위험에 남긴다.


### M8 — Template-First 미러 + 키 인벤토리

착수 HEAD `23c5c18fa`, base `3210da7d3`, 브랜치 `WT-auto-feedback`, 워크트리 `.claude/worktrees/t170`.

#### 편집한 파일

| 파일 | 편집 |
|---|---|
| `internal/template/templates/.moai/config/sections/feedback.yaml` | `auto_submit: false` + 중립 주석 5줄. 기존 `repository` 블록 무편집 |
| `.moai/config/sections/feedback.yaml` | 템플릿과 동일 내용으로 미러 |
| `internal/settings/testdata/sections/feedback.yaml` | 템플릿과 동일 내용으로 미러 (per-key 기대값의 판독 원천) |
| `internal/config/testdata/shipped_key_inventory.yaml` | `feedback.auto_submit` 항목 1건 삽입 (`feedback.repository` 바로 앞, 알파벳 순) |
| `internal/settings/schema_sections_test.go` | `TestSchemaCurrentValuesReadsAllSections` 의 per-key 맵에 `"feedback.auto_submit": "false"` 1줄 |
| `internal/web/assets/i18n.js` | `sec.feedback.desc` 4로케일 문구 확장 (M7 인계분) |
| `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md` | 신규 절 1개 × 4로케일 |

세 YAML 사본은 서로 바이트 동일하다 (`cp` 로 미러).

#### 인벤토리 `evidence` 는 정직하게 적는다

design.md §8 은 이 키를 **Go 코드가 읽지 않는다**고 못박는다 — 스킬 본문이 설정 파일을 읽어 분기하고, `FeedbackAutoSubmit()` 접근자는 현재 프로덕션 호출자가 없다. 그래서 인벤토리 항목은 `class: R` 에 `evidence: none` 이 아니라 그 사실을 그대로 적었다:

```yaml
- path: "feedback.auto_submit"
  class: R
  evidence: "consumed by the skill body (.claude/skills/moai/workflows/feedback.md); no Go production caller"
```

`evidence` 필드는 가드가 파싱만 하고 값은 검사하지 않으므로(`loadAllowlistWithCount` 는 `Path` 만 읽는다) 자유 서술이 가능하다 — 검사되지 않는 필드일수록 `none` 으로 뭉개기 쉬운데, 그러면 나중에 이 키를 보는 사람이 "접근자가 있으니 누군가 읽겠지"로 오독한다.

#### 형제 SPEC 공유 파일 확인

`SPEC-TODO-ENABLE-FLAG-001` 이 먼저 착지했는지 실측했다 — 인벤토리에 `todo.` 로 시작하는 항목은 0건이었고(`grep -n 'todo\.' internal/config/testdata/shipped_key_inventory.yaml` → 무출력), 중복 삽입은 발생하지 않았다. 공유 파일 3종(`shipped_key_inventory.yaml`, `schema_sections_test.go`, `i18n.js`)에는 **신규 항목만** 넣었고 기존 항목의 위치·서식은 건드리지 않았다(plan.md AP-10). `gofmt -w` 이후 diff 가 `1 insertion(+)` 단일 줄인 것이 그 관측이다.

`i18n.js` 의 `sec.feedback.desc` 4줄만은 예외적으로 **기존 줄을 고쳤다**. M7 이 "필드가 둘이 된 지금은 좁다"고 잔여 위험에 남긴 항목이며, 이 SPEC 자신의 변경이 부정확하게 만든 기존 설명 1개를 넓히는 것이라 재배치·재서식이 아니다.

#### 두 anti-rot 가드 — 실행 관측

plan.md §B 4번이 지목한 가드 2건. 두 번째 지목(`schema_label_test.go:96`)은 **경로가 다르다** — `internal/settings/` 가 아니라 `internal/web/schema_label_test.go` 이고, 그 안에 `TestSchemaLabel` 이라는 이름의 테스트는 없다(실측: `TestSchemaEmptyLabelParity:16` / `TestI18nKeySetParity:74` / `TestI18nSegmentKeysRemovedFromWebDictionary:133`). 이름으로 돌렸다면 0개 매칭으로 조용히 `ok` 가 찍혔을 것이므로, 실재하는 이름으로 돌렸다.

```
$ go test ./internal/config/ -run 'TestShippedConfigKeysHaveReaders' -v
=== RUN   TestShippedConfigKeysHaveReaders/non_vacuous_inventory
    shipped_key_reader_test.go:132: non-vacuity: 896 shipped keys, 959 inventory entries, 329 struct fields
--- PASS: TestShippedConfigKeysHaveReaders (0.68s)
    --- PASS: TestShippedConfigKeysHaveReaders/non_vacuous_inventory (0.00s)
    --- PASS: TestShippedConfigKeysHaveReaders/collision_resolution (0.00s)
    --- PASS: TestShippedConfigKeysHaveReaders/accessor_indirection (0.00s)
    --- PASS: TestShippedConfigKeysHaveReaders/unbound_classification (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config	1.057s

$ go test ./internal/web/ -run 'TestSchemaEmptyLabelParity|TestI18nKeySetParity|TestI18nSegmentKeysRemovedFromWebDictionary|TestScopeContract' -v
--- PASS: TestSchemaEmptyLabelParity (0.01s)
--- PASS: TestI18nKeySetParity (0.05s)
--- PASS: TestI18nSegmentKeysRemovedFromWebDictionary (0.00s)
--- PASS: TestScopeContractEditableSections (0.00s)
--- PASS: TestScopeContractExclusions (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.499s
```

키 수가 895 → 896, 인벤토리가 958 → 959 로 각각 1 늘었다 — 이 마일스톤이 넣은 키 1개와 항목 1개다.

#### GREEN — AC-F-023 템플릿 절반 판정

**PASS.** M7 이 웹 절반(스키마·라우트·i18n)을 졌고, M8 은 **템플릿 절반**(미러·인벤토리·`make build`·중립성)을 진다. 이로써 AC-F-023 양쪽이 모두 관측됐다.

acceptance.md 가 지정한 5개 선택자를 `-v` 로 돌려 각각 `=== RUN` 이 찍히는지 확인했다(0개 실행 통과 방지 — §D.3 [HARD]):

```
$ go test ./internal/settings/ ./internal/web/ -run 'TestSchemaCurrentValuesReadsAllSections|TestI18nKeySetParity|TestRouteForSectionTable|TestExcludedSectionsAllRejected|TestScopeContract' -v
=== RUN   TestSchemaCurrentValuesReadsAllSections
=== RUN   TestRouteForSectionTable
=== RUN   TestExcludedSectionsAllRejected
--- PASS: TestExcludedSectionsAllRejected (0.00s)
--- PASS: TestRouteForSectionTable (0.00s)
--- PASS: TestSchemaCurrentValuesReadsAllSections (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/settings	0.351s
=== RUN   TestI18nKeySetParity
--- PASS: TestI18nKeySetParity (0.05s)
=== RUN   TestScopeContractEditableSections
=== RUN   TestScopeContractExclusions
--- PASS: TestScopeContractEditableSections (0.00s)
--- PASS: TestScopeContractExclusions (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.597s
```

나머지 4개 관측:

```
$ grep -n 'auto_submit' internal/template/templates/.moai/config/sections/feedback.yaml
13:    auto_submit: false                                                        # 1건

$ grep -n 'feedback.auto_submit' internal/config/testdata/shipped_key_inventory.yaml
362:- path: "feedback.auto_submit"                                                # 1건

$ grep -rn 'SPEC-FEEDBACK-AUTO-SUBMIT\|REQ-' internal/template/templates/.moai/config/sections/feedback.yaml internal/template/templates/.claude/skills/moai/workflows/feedback.md
                                                                                 # 0건 (exit 1)

$ make build
catalog.yaml updated successfully (12899 bytes)
go build -ldflags "..." -o bin/moai ./cmd/moai                                    # exit 0
```

`make build` 이후 `git status --porcelain` 에 `internal/template/catalog.yaml` 은 나타나지 않았다 — 카탈로그는 에이전트·스킬만 해시하고 config 섹션 YAML 은 대상이 아니라서, 이번 편집으로는 해시가 움직이지 않는다. (해시가 움직였다면 같은 커밋에 실어야 CI parity 가 통과한다.)

#### 중립성 + 회귀 + 크로스 플랫폼

```
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/template	21.247s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	(cached)

$ go test ./internal/config/... ./internal/template/... ./internal/settings/... ./internal/web/...
ok  	github.com/modu-ai/moai-adk/internal/config	1.309s
ok  	github.com/modu-ai/moai-adk/internal/config/atomicfile	0.545s
ok  	github.com/modu-ai/moai-adk/internal/config/toolpolicy	(cached)
ok  	github.com/modu-ai/moai-adk/internal/template	22.277s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.564s
ok  	github.com/modu-ai/moai-adk/internal/settings	0.995s
ok  	github.com/modu-ai/moai-adk/internal/settings/agentfm	(cached)
ok  	github.com/modu-ai/moai-adk/internal/settings/yamlpatch	(cached)
ok  	github.com/modu-ai/moai-adk/internal/web	2.954s

$ go vet ./internal/config/... ./internal/template/... ./internal/settings/... ./internal/web/...         → exit 0 (무출력)
$ GOOS=windows go vet ./internal/config/... ./internal/template/... ./internal/settings/... ./internal/web/...  → exit 0 (무출력)
```

`TestConstitutionCrossReference` 는 `internal/cli/agentlint` 소속이라 위 스코프 밖이고, M6·M7 이 기록한 대로 base 에서도 붉다 — 이번 회차에는 **재관측하지 않았다**(아래 미검증).

#### docs-site 4로케일

en/ko/ja/zh 네 파일 모두에 절 1개씩 넣었다. 기존 "대상 저장소 설정" 절 뒤, "피드백 유형" 절 앞이다. 절 제목은 로케일별로 en `Confirming Before Submission` · ko `제출 전 확인` · ja `送信前の確認` · zh `提交前确认`.

```
$ grep -c '^### ' docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md
en:14  ko:14  ja:14  zh:14        # 13 → 14, 네 로케일 동일

$ grep -c '^## ' docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md
en:10  ko:10  ja:10  zh:10        # 무변경, 네 로케일 동일

$ grep -c 'auto_submit' docs-site/content/{en,ko,ja,zh}/utility-commands/moai-feedback.md
en:2  ko:2  ja:2  zh:2            # 산문 1 + YAML 예시 1, 네 로케일 동일
```

본문에 장식 이모지는 넣지 않았고, Mermaid 는 추가하지 않았다. YAML 예시 블록은 네 로케일에서 바이트 동일하다.

#### 설계 판단 2건

1. **인벤토리 헤더의 `Total entries: 958` 은 고치지 않았다.** 헤더는 `Code baseline: <sha>` 와 한 덩어리로 M1 시점의 스냅샷 기록이고, 어떤 테스트도 이 숫자를 읽지 않는다(가드는 항목을 세어 `minimumShippedKeys` 하한만 본다). 게다가 형제 SPEC 도 항목을 1개 늘리므로 두 카드가 같은 줄을 고치면 충돌한다. 숫자가 스테일해진 사실은 아래 잔여 위험에 남긴다.
2. **docs 에 스크러버 서사를 새로 열지 않았다.** 이 페이지는 지금까지 마스킹을 한 번도 언급한 적이 없어서, 확인 게이트가 **무엇을 보여주는지** 설명하는 데 필요한 최소한("가려낸 값의 요약")만 적고 스크러버 자체의 절은 만들지 않았다. 페이지 전체를 스크러버 관점으로 다시 쓰는 것은 이 마일스톤의 범위가 아니다.

### M9 — 검증 스윕

**이 절의 관측 주체는 오케스트레이터다.** 아래 결과는 오케스트레이터가 `a6682a007`(작업 트리 clean) 에 대해 직접 실행해 관측한 것이고, 이 마일스톤은 그것을 **기록**한다. 이 기록을 작성한 에이전트가 다시 돌려본 것이 아니다 — `verification-claim-integrity.md` §2 에 따라 귀속을 명시한다.

기준점: 브랜치 `WT-auto-feedback`, HEAD `a6682a007`, base `3210da7d3`, `git status --porcelain` 무출력.

#### 스윕 결과

| 검사 | 명령 | 관측 |
|---|---|---|
| 영향 패키지 회귀 | `go test ./internal/feedback/... ./internal/sandbox/... ./internal/config/... ./internal/cli/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/settings/... ./internal/web/... ./internal/template/... -count=1` | 전 패키지 `ok`, 예외 1건 — `internal/cli/agentlint` `TestConstitutionCrossReference` FAIL(아래 귀속 참조) |
| 경쟁 상태 | `go test -race -count=1 ./internal/feedback/...` | `ok github.com/modu-ai/moai-adk/internal/feedback 1.856s` |
| 정적 검사 | 편집한 7개 패키지 트리에 대한 `go vet` | exit 0, 무출력 |
| 크로스 플랫폼 | 같은 7개 트리에 대한 `GOOS=windows go vet` | exit 0, 무출력 |
| 린트 | `golangci-lint run --timeout=5m` (feedback / sandbox / cli / settings / web / core-project) | `0 issues.` |
| 템플릿 중립성 | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` | `ok github.com/modu-ai/moai-adk/internal/template 24.045s`, agentemit `ok` |
| 임베드 빌드 | `make build` | 성공. 직후 `git status --porcelain` 무출력 → `catalog.yaml` 드리프트 없음 |

#### 유일한 실패 1건 — 귀속까지 측정됐다

`internal/cli/agentlint` `TestConstitutionCrossReference` 는 이 SPEC 소관이 **아니다**. 귀속 측정:

- `origin/main` 의 `.claude/rules/moai/core/moai-constitution.md` → `grep -c agent-authoring` = **1**
- `origin/release/v3.1.3` 의 같은 파일 → **0**
- 제거 커밋: `243eb07ef` (t82 M4 예산 다이어트, moai-constitution 18,958 → 15,433). 사라진 줄은 `Per-agent effort calibration: see .claude/rules/moai/development/agent-authoring.md § Effort-Level Calibration Matrix`.

즉 이 붉은 신호는 base(`3210da7d3`, release/v3.1.3 계열)에서도 이미 붉었고, M1~M8 의 어떤 편집도 원인이 아니다. **리드가 복구 카드 `t189` 를 lane-8 에 발행했다.** 이 카드에서는 고치지 않고 외부 블로커로만 기록한다.

#### AC 커버리지

`acceptance.md` 의 AC-F id 는 24개이고, §E.2 는 M1~M8 에서 24개 모두에 PASS 판정을 기록했다. (M6~M8 일부는 `판정: AC-F-0NN **PASS**` 어순, 나머지는 `AC-F-0NN 판정: **PASS**` 어순 — 두 형태 모두 유효한 판정 기록이며 문구를 통일하지 않았다.)

#### 통합·PR 은 리드 보류 중

[HARD] 이 레인은 **커밋에서 멈춘다.** push·PR 개설·릴리스 워크트리 진입·머지 모두 리드의 명시적 보류 대상이다. 리드의 통합 큐 순서는 lane-8(t183→t189) → lane-5(t182) → lane-7(이 카드)이며, 이 레인은 진입 전에 리드에게 알린다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-23
run_commit_sha: a6682a007          # M8 = 현재 HEAD. M9 기록 커밋은 이 블록을 쓰는 커밋이라 여기에 담기지 않는다
run_status: complete-pending-integration
run_base_sha: 3210da7d3
branch: WT-auto-feedback
worktree: .claude/worktrees/t170
cycle: tdd
ac_pass_count: 24
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a          # 통합·push 는 리드 보류 — 이 레인은 커밋에서 멈춘다
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0
cross_platform_build:
  goos_darwin_vet: pass
  goos_windows_vet: pass
  windows_runtime_behavior: not-verified-locally   # CI 소관
m1_to_mN_commit_strategy: one-commit-per-milestone   # M1~M8 각 1커밋 + M9 기록 커밋
external_blocker: t189             # TestConstitutionCrossReference (agentlint) — 이 SPEC 소관 아님
integration_status: held-by-lead   # 큐: lane-8(t183→t189) → lane-5(t182) → lane-7
```

### 검증 주체 귀속 (읽는 감사관을 위해)

아래 스윕 증거는 **오케스트레이터가 `a6682a007` 에 대해 직접 실행해 관측**한 것을 이 SPEC 이 기록한 것이다. run-phase 에이전트의 새 측정이 아니다. 마일스톤별 AC 판정(M1~M8)은 각 마일스톤 에이전트가 그 시점 트리에서 직접 관측한 것으로, §E.2 에 verbatim 출력이 남아 있다.

### 마일스톤 커밋 8건

| M | SHA | 제목 |
|---|---|---|
| M1 | `95fc239e3` | M1 feedback.auto_submit config key |
| M2 | `e51475068` | M2 scrubber type contract and masking transforms |
| M3 | `3bcceffc7` | M3 vulnerability classifier reading the pre-mask body |
| M4 | `55dc0ec0a` | M4 mask log and retry queue |
| M5 | `d2063308b` | M5 moai feedback scrub and queue verbs |
| M6 | `38705eb85` | M6 skill gate clauses and wizard question |
| M7 | `23c5c18fa` | M7 expose feedback section in the web console |
| M8 | `a6682a007` | M8 template mirror, key inventory, and 4-locale docs |

base `3210da7d3` (release/v3.1.3 계열 머지 커밋).

### AC 커버리지 — 24/24 PASS

| AC | 소유 M | 판정 근거 명령 |
|---|---|---|
| AC-F-001 키 해석 | M1 | `go test ./internal/config/ -run 'TestFeedbackAutoSubmit' -v` |
| AC-F-002 확인 게이트 존재 | M6 | 편집 전/후 양축 grep: `grep -n 'AskUserQuestion\|gh issue create\|auto_submit'` (base 사본에서 `auto_submit` 0건 → FAIL 관측, 편집 후 `:108` < `:143` → PASS) |
| AC-F-003 스크러버 계약 | M5 | `go test ./internal/cli/ -run 'TestFeedbackScrubContract' -v` |
| AC-F-004 fail-closed | M5 | `go test ./internal/cli/ -run 'TestFeedbackScrubToolFailureExitsNonZero' -v` + 빌드된 바이너리 재관측(exit 1, stdout 0바이트) |
| AC-F-005 findings 무원문 | M2 | `go test ./internal/feedback/ -run 'TestFindingsCarryNoRawValue' -v` |
| AC-F-006 GitHub 토큰 | M2 | `go test ./internal/feedback/ -run 'TestScrubMasksGitHubToken' -v` |
| AC-F-007 AIza 합집합 | M2 | `go test ./internal/feedback/ -run 'TestScrubMasksGoogleAPIKey' -v` |
| AC-F-008 무해 본문 양축 | M2 (M3 재관측) | `go test ./internal/feedback/ -run 'TestScrubBenignBodyUntouchedAndAllowed' -v` |
| AC-F-009 출력 형태 통일 | M2 | `go test ./internal/feedback/ -run 'TestMaskOutputShapeMatchesExistingMasker' -v` |
| AC-F-010 홈 경로 축약 | M2 | `go test ./internal/feedback/ -run 'TestScrubCollapsesHomePath' -v` |
| AC-F-011 env 값 마스킹 | M2 | `go test ./internal/feedback/ -run 'TestScrubMasksEnvValues' -v` |
| AC-F-012 취약점 라우팅 | M3 | `go test ./internal/feedback/ -run 'TestClassifyBlocksVulnerabilityReport' -v` |
| AC-F-013 분류는 원문을 본다 | M3 | `go test ./internal/feedback/ -run 'TestClassifyReadsPreMaskBody' -v` |
| AC-F-014 멱등성 | M2 | `go test ./internal/feedback/ -run 'TestScrubIsIdempotent' -v` |
| AC-F-015 로그 내용 + 권한 | M4 | `go test ./internal/feedback/ -run 'TestMaskLogRecordsKindAndCountWithoutRawValue' -v` |
| AC-F-016 로그 fail-open | M4 | `go test ./internal/feedback/ -run 'TestMaskLogFailureIsFailOpen' -v` |
| AC-F-017 실패 → 큐 적재 | M4 | `go test ./internal/feedback/ -run 'TestQueueEnqueuesOnSendFailure' -v` |
| AC-F-018 성공 → 큐 제거 | M4 | `go test ./internal/feedback/ -run 'TestQueueResolvesOnSuccess' -v` |
| AC-F-019 스킬 [HARD] 조항 4종 | M6 | 두 사본(SRC/TPL) × 재작성된 7개 grep 관측 + `diff` 드리프트 확인 + 중립성 grep. 기준 토큰은 base 0히트 실측본 |
| AC-F-020 마법사 질문 + 개수 고정 | M6 | `go test ./internal/cli/wizard/ -run 'TestFeedbackAutoSubmitQuestion\|TestQuestionOrder\|TestReconfigureQuestions' -v` |
| AC-F-021 4로케일 번역 완전성 | M6 | `go test ./internal/cli/wizard/ -run 'TestFeedbackAutoSubmitTranslationsExist\|TestWizardQuestionTranslationCompleteness' -v` |
| AC-F-022 답변이 파일에 기록 | M6 | `go test ./internal/core/project/ -run 'TestWritePhase1ConfigsPersistsFeedbackAutoSubmit\|...SkipsFeedbackWhenUnanswered\|...FeedbackNoFile' -v` (실제 로더 재읽기로 판정) |
| AC-F-023 웹 + 템플릿 + 인벤토리 + 빌드 + 중립성 | M7(웹 절반) + M8(템플릿 절반) | M7: 스키마·라우트·i18n 테스트. M8: 미러 `diff` + 인벤토리 가드 + `make build` + `MOAI_TEMPLATE_LEAK_STRICT=1` |
| AC-F-024 개인키 블록 통째 마스킹 | M2 | `go test ./internal/feedback/ -run 'TestScrubMasksPrivateKeyBlockEntirely' -v` |

AC-F-023 은 유일하게 두 마일스톤이 절반씩 지며, 각 마일스톤이 자기 절반만 주장하고 나머지를 주장하지 않는다고 명시했다(§E.2 M7·M8).

### 검증 스윕 증거 (오케스트레이터 관측, `a6682a007`)

- 영향 9개 패키지 `-count=1` 회귀 → 전부 `ok`, 예외 1건(외부 블로커, 아래)
- `go test -race -count=1 ./internal/feedback/...` → `ok … 1.856s`
- 7개 패키지 트리 `go vet` → exit 0, 무출력
- 7개 패키지 트리 `GOOS=windows go vet` → exit 0, 무출력
- `golangci-lint run --timeout=5m` (feedback/sandbox/cli/settings/web/core-project) → `0 issues.`
- `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...` → `ok … 24.045s`, agentemit `ok`
- `make build` → 성공, 직후 `git status --porcelain` 무출력(`catalog.yaml` 드리프트 없음)

### 외부 블로커 1건 — 이 SPEC 소관 아님

`internal/cli/agentlint` `TestConstitutionCrossReference` FAIL. 귀속 측정: `origin/main` 의 `moai-constitution.md` 는 `grep -c agent-authoring` = 1, `origin/release/v3.1.3` 은 0. 제거 커밋 `243eb07ef`(t82 M4 예산 다이어트)가 `Per-agent effort calibration: see .claude/rules/moai/development/agent-authoring.md § Effort-Level Calibration Matrix` 줄을 삭제했다. base 에서도 붉으므로 이 SPEC 의 회귀가 아니다. 리드가 복구 카드 **t189** 를 lane-8 에 발행했고, 이 레인은 손대지 않았다.

### 검증하지 **않은** 것 (명시)

1. **전체 스위트 `go test ./...`** — 로컬에서 돌리지 않았다(CLAUDE.local.md §4 규율). 전 패키지 판정은 CI 몫이며, 깨끗한 환경의 PR head 실행이 더 강한 근거다.
2. **Windows 런타임 동작** — `GOOS=windows go vet` 은 컴파일만 증명한다. 실제 Windows 동작은 CI 매트릭스 소관이다.
3. **브라우저 확인 없음** — 웹 콘솔의 feedback 섹션(M7)과 `sec.feedback.desc` 4로케일 문구(M8)는 파일 내용 검증으로만 관측했다. 콘솔을 띄워 눈으로 본 적 없다.
4. **바이너리 스모크는 M5 범위뿐** — M5 가 `moai feedback scrub` 을 빌드된 바이너리로 재관측했다. 그 이후 `~/go/bin/moai` 재설치나 추가 CLI 스모크는 하지 않았다.
5. **Hugo 빌드 미실행** — M8 의 docs-site 4로케일 편집이 렌더 단계에서 경고 없이 통과하는지 확인하지 않았다(표준 markdown 만 추가; shortcode·Mermaid·이모지 없음).
6. **`TestConstitutionCrossReference` 의 base 붉음 재관측** — 위 귀속은 `grep -c` 로 직접 측정했으나, base 트리에서 그 테스트 자체를 다시 돌린 회차는 없다.
7. **인벤토리 헤더 `Total entries: 958`** — 실제 959 와 어긋난 채로 남겼다(형제 SPEC 과 같은 줄 충돌 회피 + 어떤 테스트도 이 숫자를 읽지 않음).

### 누적 잔여 위험 (M2~M8 보고서 §5 에서 취합)

**마스킹 범위 (M2)**
1. `MaskSecret` 이 값의 첫 1자 + 끝 4자를 남긴다 — 공개 이슈 텍스트에 토큰 꼬리 4자가 남는다. 형태 변경은 세 마스커 통합 카드가 필요하다.
2. env 값 8자 하한 — 8자 미만 자격증명은 미탐. 산문 훼손 회피와의 의도된 교환.
3. `AWS_` 접두사 미채택 — deny list 에 이름이 없으면 이름 어휘로 걸리지 않는다(`AKIA` 형태는 정규식이 잡음). deny list 확장은 `internal/sandbox` 소관.
4. 과잉/과소 마스킹은 양방향 AC 2건으로만 관측됐고 실제 보고서 분포 표본은 없다.
5. `hook` 정책이 나중에 패턴을 추가하면 재작성 경로도 자동으로 넓어진다(설계된 성질이나 산문 훼손이 먼저 나타날 지점).

**분류기 (M3)**
6. 영어 어휘 편중 — 한국어·일본어로 쓰인 순수 산문 취약점 보고는 신호를 발생시키지 않는다.
7. CVE 식별자 단독 차단의 오탐 — "CVE-XXXX-XXXX 때문에 버전 올려주세요" 같은 무해한 요청이 차단된다(비대칭 비용 하에서 의도된 방향).
8. 경로 토큰 추출 정밀도 — 따옴표·괄호 이형은 놓칠 수 있다(과소 탐지 방향, 단독 차단 아님).
9. `classify` 가 `hook.DefaultSecurityPolicy()` 를 한 번 더 빌드 — `Scrub` 1회당 정책 컴파일 2회. 루프 소비자가 생기면 재검토.
10. 거부 메시지 ↔ `SECURITY.md` 동기화 부채(기계적 패리티 검사 없음).

**로그·큐 (M4·M5)**
11. 로그 인터리빙 — 잠금이 없어 동시 스크럽 시 한 줄이 섞일 수 있다(훼손 결과는 판독 불가 한 줄이지 값 유출 아님).
12. fail-open 의 이면 — "로그가 없다"를 "마스킹이 없었다"로 읽으면 안 된다. `slog.Warn` 이 유일한 신호.
13. 큐 재전송 시 재스크럽 — `queue enqueue` 는 stdin JSON 을 신뢰한다. 재전송 직전에 다시 스크럽하지 않으면 유출 경로가 된다(파이프라인이 멱등이므로 재스크럽이 안전).
14. `queue.lock` 잔존 — 비정상 종료 시 자동 해제되지 않는다. stale-lock 정리는 이 SPEC 범위 밖.
15. `internal/cli` 스위트는 343초 — 이 패키지 재실행 시 타임아웃 하한 600초를 지켜야 한다(300초로 걸면 통과하는 트리에서 FAIL).
16. fail-closed 정책 로드가 기존 사용자에게 새 실패 지점이 된다 — `security.yaml` 이 깨진 프로젝트는 "피드백만 안 된다"로 보인다(에러 메시지가 경로+파싱 위치를 싣는 것이 완화책).

**강제력의 한계 (M5·M6·M7)**
17. **스크러버 도입은 규약 강제이지 샌드박스가 아니다.** `moai feedback` 을 거치지 않고 직접 이슈를 여는 경로는 그대로 열려 있다(plan.md AP-12, design.md §1 — 감사 유지 판정을 받아 후퇴시키지 않은 정직성 주장).
18. 게이트 3옵션의 `(권장)` 기본이 "제출하지 않음"이라 이탈률이 오를 수 있다(`auto_submit: true` 가 탈출구).
19. 마법사 질문 문구가 부정형 — "예"가 게이트 **해제**를 뜻해 오독 여지가 있다. 기본 `false` 라 오독 비용은 안전한 쪽으로 떨어진다.
20. **`feedback.auto_submit` 을 Go 코드가 읽지 않는다**(design.md §8). 스킬 본문이 설정을 직접 읽어 분기하므로 설정↔산문 드리프트를 컴파일러도 테스트도 잡지 않는다. 콘솔에서 켠 값이 실제 전송 동작으로 이어지는지는 스킬 실행 경로에서만 관측된다.
21. 웹 재개방이 위조 POST 표면을 다시 연다 — `feedback` 이 `RouteSeam` 이 되며 동의 토글이 웹에서 켜질 수 있다(D5 의도). 전송의 fail-closed 3조항이 여전히 유일한 방어선.

**형제 SPEC 병합 (M6·M7·M8)**
22. `SPEC-TODO-ENABLE-FLAG-001` 과 파일 9종을 공유한다. 이 SPEC 은 [HARD] 규율대로 신규 항목만 추가했으나, 개수 고정 테스트 4건은 숫자가 바뀌었으므로 형제가 나중에 착지하면 같은 4건을 다시 조정해야 한다.
23. `i18n.js` · `schema_sections_test.go` · `shipped_key_inventory.yaml` 은 인접 삽입이라 diff 컨텍스트가 겹칠 수 있다(알파벳 순이라 인벤토리는 멀리 떨어져 있고 맵은 순서 무관).
24. 고정 테스트 6곳 중 하나를 놓쳤을 가능성 — 목록을 "붉어진 지점"에서 역산했다. 다른 패키지가 `ExcludedSections()` 나 탭 목록을 복사했다면 CI 에서만 드러난다.

**문서 (M7·M8)**
25. `sec.feedback.desc` 4로케일 문구가 필드 2개가 된 패널에 대해 좁은 채로 남아 있었으나 M8 에서 4로케일 동시 확장으로 처리했다 — 다만 브라우저 확인은 하지 않았다(위 미검증 3).
26. docs 문구 ↔ 스킬 본문 사이에 기계적 패리티 검사가 없다. 스킬 본문이 바뀌면 문서가 조용히 어긋난다.
27. 인벤토리 `evidence` 자유 서술의 내구성 — 가드가 값을 검사하지 않으므로 `class: R` → `W` 재분류는 사람이 해야 한다.
28. docs 에 스크러버 서사를 새로 열지 않았다 — 마스킹 규칙 전체를 문서에서 찾는 사용자는 이 페이지에서 찾지 못한다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-23
sync_commit_sha: pending-backfill-SPEC-FEEDBACK-AUTO-SUBMIT-001   # 커밋은 자기 해시를 알 수 없다 — 후속 커밋에서 백필
sync_status: complete-pending-integration
sync_base_sha: cdff7f315          # run-phase HEAD(M9 기록 커밋)
branch: WT-auto-feedback
worktree: .claude/worktrees/t170
changelog_entry_position: "[Unreleased] → Summary 4문단 + Added 최상단 SPEC 항목 1건 + 사용자 대면 항목 3건"
b12_self_test_a: "pre-emission grep — `git show HEAD:CHANGELOG.md | grep -c 'SPEC-FEEDBACK-AUTO-SUBMIT-001'` = 0 (중복 없음)"
b12_self_test_b: "AC 개수 대조 — `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` = 24, CHANGELOG 주장 24 일치 (0이 아니므로 공허 비교 아님)"
b12_self_test_c: "파일 경로 확인 — CHANGELOG 가 인용한 spec.md · plan.md · design.md 를 `ls` 로 존재 확인"
frontmatter_status_transitions:
  spec_md: "in-progress → completed"
  plan_md: n/a          # frontmatter 없음 — 측정: `head -3 plan.md` 가 `# Plan — …` 로 시작
  acceptance_md: n/a    # frontmatter 없음 — 같은 방식으로 측정
  progress_md: n/a      # frontmatter 없음 — 같은 방식으로 측정
  note: "병합된 3-phase close — implemented 를 경유하되 별도 커밋 없이 단일 sync 커밋이 completed 까지 싣는다. 이 SPEC 은 4종 중 spec.md 하나만 frontmatter 를 가지므로 전이도 그 한 파일에서만 일어난다. 나머지 3종에 frontmatter 를 새로 만들지 않았다 — 없는 것을 만드는 것은 본문 편집이고, sync 소관이 아니다. `updated:` 는 이미 2026-08-23 이라 값이 바뀌지 않았다(재기입 불필요). 본문 §A~§H 는 손대지 않았다."
docs_site_4locale: verified-not-rewritten   # M8(a6682a007)이 착지시킴, sync 는 패리티만 관측
readme_change: none                          # 근거는 아래 §검증 기록
external_blocker: t189                       # TestConstitutionCrossReference (agentlint) — 이 SPEC 소관 아님, lane-8
integration_status: held-by-lead             # 큐: lane-8(t183→t189) → lane-5(t182) → lane-7(이 카드)
```

### sync 가 바꾼 것 (전부)

| 파일 | 변경 |
|---|---|
| `CHANGELOG.md` | `[Unreleased]` Summary 4문단 + Added 4항목(SPEC close 1 + 사용자 대면 3) |
| `.moai/specs/SPEC-FEEDBACK-AUTO-SUBMIT-001/spec.md` | frontmatter `status`·`updated` 만 |
| `…/plan.md` | 변경 없음 (frontmatter 부재) |
| `…/acceptance.md` | 변경 없음 (frontmatter 부재) |
| `…/progress.md` | 이 §E.4 (frontmatter 부재) |
| `.moai/reports/t170/sync-report.md` | 신규(5-section 보고서) |

**코드 변경 0건.** docs-site 4로케일은 run-phase M8 이 이미 착지시켰으므로 sync 는 다시 쓰지 않고 관측만 했다.

### 검증 기록 (관측한 것)

- **close 전 `moai spec audit --json`** (워크트리 cwd): `total_specs: 640`, MUST-FIX **1건** — `SPEC-CODEX-SKILLS-CANONICAL-001` 의 `SyncStatusDrift`(타 SPEC, 선재). 이 SPEC 의 drift finding **0건**.
- **close 후 `moai spec audit --json`**: 아래 "close 후 관측" 참조.
- `go test ./internal/config/... ./internal/template/...` → 전부 `ok` (config 9.752s / template 52.516s). 문서·CHANGELOG 전용 변경이 이 두 패키지를 움직이지 않음을 확인.
- **4로케일 패리티**: `moai-feedback.md` 4개 로케일 각각 `auto_submit` 2히트 · `^### ` 헤딩 14개로 동일. 새 절(`제출 전 확인` / `Confirming Before Submission` / 일·중 대응절)은 4로케일 모두 존재.
- **README 무변경 근거**: `grep -n feedback README.md README.ko.md README.ja.md README.zh.md` → 4파일 각각 3곳(커맨드 나열 2 + 이슈 링크 1)뿐이고 설정 키를 열거하는 문단이 없다. `README.md:574` 의 `.moai/config/sections/` 표는 "편집하게 되는" 6개 + v3.1.1 추가 4개만 싣고 `feedback.yaml` 을 포함하지 않는다. 따라서 이 SPEC 이 README 에 만드는 부채는 없다.

### 검증하지 **않은** 것 (§E.3 의 7건은 그대로 유효하며, sync 에서 더해진 것만 적는다)

1. **`mcp__moai__spec_audit` 의 `project_root` 가 이 워크트리에서 작동하지 않았다.** `project_root=/…/.claude/worktrees/t170` 을 넘겨 호출했는데 결과가 `total_specs: 627` 이고 `SPEC-FEEDBACK-AUTO-SUBMIT-001` 이 출력에 **0회** 등장했다 — 이 SPEC 디렉터리는 워크트리에만 있고 primary 체크아웃(`/Users/goos/MoAI/moai-adk-go/.moai/specs/`, 628개)에는 없다. 즉 primary 를 감사한 것으로 관측된다. 그래서 판정 근거는 워크트리 cwd 에서 실행한 **CLI** (`moai spec audit --json`, `total_specs: 640`)로 잡았다. MCP 쪽 원인 규명은 이 카드 범위 밖이며 별도 카드감이다.
2. **Hugo 빌드 미실행** — §E.3 의 미검증 5와 동일. sync 에서도 docs-site 를 렌더하지 않았다(이번 sync 는 docs-site 파일을 편집하지 않았다).
3. **전체 스위트·CI** — 로컬에서 `go test ./...` 를 돌리지 않았다(CLAUDE.local.md §4). 통합 후 PR CI 가 판정한다.
4. **push·PR·머지 없음** — 리드의 통합 큐 보류 대상. 이 레인은 sync 커밋에서 멈춘다.

### 잔여 위험 (sync 시점에 새로 생긴 것)

- `sync_commit_sha` 가 플레이스홀더로 남는다. 백필 커밋 전까지 이 값은 감사관에게 "sync 는 끝났으나 해시는 미기입" 상태로 읽혀야 한다. 형제 SPEC 사례(`SPEC-CODEX-SKILLS-CANONICAL-001`)에서 보듯, `status` 가 `completed` 에 도달하지 못한 채 플레이스홀더만 남으면 `SyncStatusDrift` MUST-FIX 가 뜬다 — 이 SPEC 은 같은 커밋에서 `completed` 까지 올려 그 형태를 피했다.
- §E.3 의 누적 잔여 위험 28건은 sync 로 해소되지 않았다. 특히 **20번**(`feedback.auto_submit` 을 Go 코드가 읽지 않는다)은 CHANGELOG 문구가 "convention, not sandbox" 로 정직하게 반영됐을 뿐 기술적으로 달라진 것이 없다.
