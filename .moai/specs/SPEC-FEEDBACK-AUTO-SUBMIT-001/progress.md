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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
