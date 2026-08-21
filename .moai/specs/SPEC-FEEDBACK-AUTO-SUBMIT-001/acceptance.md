# Acceptance — SPEC-FEEDBACK-AUTO-SUBMIT-001

> 각 AC는 판정을 내리는 **명령 하나**와 한 번의 관측으로 결정되는 **기대치**를 명시한다. 사람의 판단이 개입하는 문구는 쓰지 않는다.
> 이름 붙은 테스트 함수는 해당 마일스톤의 **납품물**이다 — 먼저 쓰고(RED) 판정한다. 존재하지 않는 테스트를 가리키는 `-run` 패턴으로 통과를 주장하는 것, 아무것도 실행하지 않고 통과하는 AC는 둘 다 이 팩토리 런에서 관측된 실패 유형이므로 금지한다.
> 로컬 검증은 패키지 스코프(`go test ./internal/<pkg>/...`)로만. 전 패키지 판정은 CI 몫(CLAUDE.local.md §4).

**AC 총계: 24 / Tier L 상한 25.**

## §D AC Matrix

| AC ID | REQ | Severity | 요약 |
|---|---|---|---|
| AC-F-001 | REQ-1 | MUST-PASS | 키 해석: 부재→false, 명시 true→true |
| AC-F-002 | REQ-2 | MUST-PASS | 확인 게이트가 `gh issue create` 앞에 존재(baseline FAIL 대조 포함) |
| AC-F-003 | REQ-3 | MUST-PASS | `scrub` 5필드 JSON(`title` 포함) + `findings.where` + exit 0 |
| AC-F-004 | REQ-3 | MUST-PASS | 도구 실패 → exit ≠ 0 (fail-closed) |
| AC-F-005 | REQ-3 | MUST-PASS | `findings`에 원문 값 없음 |
| AC-F-006 | REQ-3·4 | MUST-PASS | GitHub 토큰 마스킹 — **본문과 제목 양쪽**(D1 행동 관측) |
| AC-F-007 | REQ-4 | MUST-PASS | `AIza` 키 마스킹 (합집합 증명) |
| AC-F-008 | REQ-4·7 | MUST-PASS | 무해 본문 → 마스킹 0건 **AND** `verdict: ok` + 소문자 산문 과잉 마스킹 없음(D1b) |
| AC-F-009 | REQ-4 | MUST-PASS | 마스킹 출력 형태가 기존 마스커와 일치 |
| AC-F-010 | REQ-5 | MUST-PASS | 홈 경로 축약 + `HOME` 오버라이드 반영 |
| AC-F-011 | REQ-6 | MUST-PASS | 민감 env 값 마스킹 + `env_scrub_extra` 확장 |
| AC-F-012 | REQ-7 | MUST-PASS | 취약점 본문 → `blocked` + SECURITY.md 라우팅 문구 |
| AC-F-013 | REQ-7 | MUST-PASS | 분류가 마스킹 이전 원문을 본다 |
| AC-F-014 | REQ-3 | MUST-PASS | 파이프라인 멱등성 |
| AC-F-015 | REQ-8 | MUST-PASS | 마스킹 로그: 종류·건수 기록, 원문 없음, 권한 `0o600` |
| AC-F-016 | REQ-8 | MUST-PASS | 로그 쓰기 실패가 스크럽을 중단시키지 않음(fail-open) |
| AC-F-017 | REQ-9 | MUST-PASS | 전송 실패 → 큐 적재 |
| AC-F-018 | REQ-9 | MUST-PASS | 이후 성공 → 큐에서 제거 |
| AC-F-019 | REQ-10 | MUST-PASS | 스킬 [HARD] 의무 4종(제목 전달·fail-closed 3문장·실패 분기·라벨 언어)이 두 사본 각각에 존재 |
| AC-F-020 | REQ-11 | MUST-PASS | 마법사 질문 존재·기본 false·개수 고정 테스트 유지 |
| AC-F-021 | REQ-11 | MUST-PASS | 4로케일 번역 완전성 |
| AC-F-022 | REQ-11 | MUST-PASS | 마법사 답변이 실제로 파일에 기록됨(사장 경로 배제) |
| AC-F-023 | REQ-12·13 | MUST-PASS | 웹 등록(라우트 반전 포함) + i18n + 템플릿 미러 + 인벤토리 + `make build` + 중립성 |
| AC-F-024 | REQ-4 | MUST-PASS | 개인키 블록이 헤더부터 종료자까지 통째로 마스킹(D1b 과소 마스킹 방어) |

전 항목 MUST-PASS. Tier L이며 배포 템플릿과 공개 채널 전송 경로를 동시에 건드린다. 보안 조항(REQ-3~8)은 포괄 AC 1건으로 묶지 않고 AC-F-005~016·F-024로 쪼갰다.

**REQ-4 하위 조항(치환 span 규칙)의 양방향 관측**: 과소 마스킹은 **AC-F-024**(개인키 블록 전체), 과잉 마스킹은 **AC-F-008 네 번째 케이스**(소문자 산문)가 본다. 감사관이 요구한 두 관측을 AC 두 개로 벌리지 않고 하나는 기존 오탐 대조 AC에 접었다 — 관측은 유지하면서 상한 25를 지키기 위함이고, 두 단언이 분리돼 있어 실패 시 어느 방향인지 구별된다.

## §D.2 Given-When-Then Scenarios

### AC-F-001 — 키 해석 (기본 + 명시 오버라이드)

```
Given auto_submit 키가 없는 임시 프로젝트
When  Loader.Load() 로 읽는다
Then  FeedbackAutoSubmit() == false
Given 같은 프로젝트의 feedback.yaml 에 "auto_submit: true" 를 기록한다
When  다시 읽는다
Then  FeedbackAutoSubmit() == true 이고 FeedbackRepository() 는 기존 값을 유지한다
```
`go test ./internal/config/ -run 'TestFeedbackAutoSubmit' -v` → PASS

### AC-F-002 — 확인 게이트 존재 (baseline 대조 필수)

```
Given .claude/skills/moai/workflows/feedback.md
When  AskUserQuestion / gh issue create / auto_submit 의 등장 줄 번호를 비교한다
Then  gh issue create 줄 번호보다 작은 위치에 auto_submit 조건부 AskUserQuestion 항목이 존재한다
```
```bash
grep -n 'AskUserQuestion\|gh issue create\|auto_submit' .claude/skills/moai/workflows/feedback.md
```
[HARD] 이 AC는 **편집 전 트리에서 FAIL해야 한다**(오늘 `AskUserQuestion`은 `:52`/`:156`/`:178`뿐). 편집 전 FAIL과 편집 후 PASS를 둘 다 관측해 기록한다.

### AC-F-003 — 스크러버 계약

```
Given 임의의 제목 문자열과 본문 문자열
When  echo "<body>" | moai feedback scrub --title "<title>"
Then  stdout 이 verdict/title/body/findings/reason 5필드를 가진 단일 JSON 객체이고 종료 코드가 0이다
  And findings 의 각 항목이 where 를 담는다 ("title" | "body")
```
`go test ./internal/cli/ -run 'TestFeedbackScrubContract' -v` → PASS
+ 바이너리 스모크:
```bash
echo 'hello' | ./bin/moai feedback scrub --title 'a title' | jq -e '.verdict and (.title|type=="string") and (.findings|type=="array")'   # exit 0
```
`title` 필드가 계약에 있다는 것이 D1(제목이 스크럽을 우회하던 경로)이 닫혔다는 구조적 증거다. 실제 마스킹 관측은 AC-F-006이 한다.

### AC-F-004 — 도구 실패는 fail-closed

```
Given 스크러버가 정책을 로드할 수 없는 조건
When  moai feedback scrub 을 실행한다
Then  종료 코드가 0이 아니고, stdout 에 verdict: ok 를 담은 JSON 이 나오지 않는다
```
`go test ./internal/cli/ -run 'TestFeedbackScrubToolFailureExitsNonZero' -v` → PASS

### AC-F-005 — findings 에 원문 값 없음

```
Given 본문에 ghp_ 로 시작하는 36자 토큰 1건
When  Scrub 을 호출한다
Then  Result.Findings 의 어떤 필드도 그 토큰(또는 8자 이상 부분열)을 담지 않는다
```
`go test ./internal/feedback/ -run 'TestFindingsCarryNoRawValue' -v` → PASS

### AC-F-006 — GitHub 토큰 마스킹

```
Given 본문 "token is ghp_<36자> here"
When  Scrub 을 호출한다
Then  Result.Body 에 원본 토큰이 없고 Findings 에 Kind "secret" Where "body" Count 1 이 있다
Given 같은 토큰이 본문이 아니라 제목에 들어 있다
When  Scrub 을 호출한다
Then  Result.Title 에 원본 토큰이 없고 Findings 에 Kind "secret" Where "title" 이 있다
```
`go test ./internal/feedback/ -run 'TestScrubMasksGitHubToken' -v` → PASS
두 번째 케이스가 D1의 행동 관측이다 — 제목이 스크럽 대상에서 빠지면 여기서 FAIL한다. `Where` 단언까지 있어야 확인 게이트가 "제목에서 가려졌다"를 말할 수 있다.

### AC-F-007 — AIza 마스킹 (합집합 증명)

```
Given 본문 "key AIza<35자>"
When  Scrub 을 호출한다
Then  Result.Body 에 원본 키가 없다
```
`go test ./internal/feedback/ -run 'TestScrubMasksGoogleAPIKey' -v` → PASS
이 패턴은 Go 정책 목록에 없고 astgrep 룰에만 있으므로, 통과는 합집합이 실제로 적용됐음을 증명한다.

### AC-F-008 — 무해 본문 양축 오탐 대조

```
Given 본문 "the ghp_ prefix is how GitHub tokens start. moai init 실행 시 마법사가 두 번 뜹니다."
When  Scrub 을 호출한다
Then  Result.Body 가 입력과 바이트 동일
  And Result.Findings 가 비어 있다        (마스킹 오탐 없음)
  And Result.Verdict == "ok"              (분류 오탐 없음)
```
```
Given 본문에 소문자 산문 연속이 들어 있다 — AWS 키 접두사를 **소문자로** 쓴 뒤 영숫자 16자가 이어지는 형태
      (리터럴은 여기 적지 않는다; 아래 주석 참조. 테스트는 그 문자열을 코드에서 조립한다)
When  Scrub 을 호출한다
Then  Result.Body 가 입력과 바이트 동일하다   (대소문자 비대칭 과잉 마스킹 없음)
```
`go test ./internal/feedback/ -run 'TestScrubBenignBodyUntouchedAndAllowed' -v` → PASS

앞 세 단언은 각각 다른 축이라 실패 시 어느 축인지 구별되며, "전부 마스킹"과 "전부 차단" 축퇴 구현을 동시에 배제한다.

네 번째 케이스는 D1b의 **과잉 마스킹** 방향 관측이다. 원본 정책은 `compilePatterns`(`internal/hook/pre_tool.go:104`)가 `(?i)`를 붙여 AWS 키 패턴을 대소문자 무시로 만든다 — 탐지기에서는 무해하지만 재작성기에서는 소문자 산문을 먹는다. 통과하려면 재작성 경로가 대소문자 민감 패턴을 `(?i)` 없이 재컴파일해야 한다(REQ-4 하위 조항). 과소 마스킹 방향은 AC-F-024가 본다.

> **이 케이스를 문서에 리터럴로 적지 않는 이유(실측)**: 이 AC를 작성하며 해당 소문자 문자열을 이 파일에 그대로 쓰려다 PreToolUse 훅에 `Content contains sensitive data` 로 **차단당했다**. 탐지기가 그 소문자 연속을 자격증명으로 판정한 것이며, D1b가 서술한 대소문자 비대칭이 살아 있는 트리에서 관측된 셈이다. 테스트는 문자열을 소스에서 조립하고(예: 접두사 상수 + 16자 더미), 문서는 형태만 서술한다.

### AC-F-009 — 출력 형태 통일

```
Given 8자 초과 시크릿 1건
When  Scrub 을 호출한다
Then  마스킹 결과가 채택한 기존 마스커 함수의 반환값과 정확히 같다
```
`go test ./internal/feedback/ -run 'TestMaskOutputShapeMatchesExistingMasker' -v` → PASS
(테스트는 채택 함수를 직접 호출해 기대값을 만든다 — 형태 문자열 하드코딩 금지.)

### AC-F-010 — 홈 경로 축약 + HOME 오버라이드

```
Given HOME=/tmp/h1, 본문에 "/tmp/h1/proj/main.go"
When  Scrub 을 호출한다
Then  Body 가 "~/proj/main.go" 를 담고 "/tmp/h1/" 를 담지 않는다
Given HOME 을 /tmp/h2 로 바꾸고 본문에 "/tmp/h2/x" 를 넣는다
When  다시 Scrub 한다
Then  축약 기준이 새 HOME 을 따른다 (paths.Home() 계약)
```
`go test ./internal/feedback/ -run 'TestScrubCollapsesHomePath' -v` → PASS
(비병렬 서브테스트에서만 `t.Setenv` — CLAUDE.local.md §6.)

### AC-F-011 — env 값 마스킹 + 확장

```
Given GITHUB_TOKEN 값이 "abcd1234efgh" 이고 본문이 그 값을 산문으로 담는다
When  Scrub 을 호출한다
Then  Body 에 그 값이 없고 Findings 에 Kind "env" 가 있다
Given security.sandbox.env_scrub_extra 에 MY_CUSTOM_TOKEN 을 등록하고 그 값을 본문에 넣는다
When  다시 Scrub 한다
Then  그 값도 마스킹된다
```
`go test ./internal/feedback/ -run 'TestScrubMasksEnvValues' -v` → PASS

### AC-F-012 — 취약점 → blocked + SECURITY.md 라우팅

```
Given 취약점 보고로 분류되는 본문
When  Scrub 을 호출한다
Then  Result.Verdict == "blocked"
  And Result.Reason 이 GitHub Security Advisories 경로 문자열을 담는다
```
`go test ./internal/feedback/ -run 'TestClassifyBlocksVulnerabilityReport' -v` → PASS
```bash
grep -c 'security/advisories/new' internal/feedback/classify.go   # >= 1
```

### AC-F-013 — 분류는 마스킹 이전 원문을 본다

```
Given 본문이 시크릿 1건만을 분류 신호로 갖는다(그 외 취약점 어휘 없음)
When  Scrub 을 호출한다
Then  Verdict == "blocked" 이다
  And 같은 본문을 미리 마스킹한 문자열을 분류기에 단독으로 넣으면 "ok" 가 나온다
```
`go test ./internal/feedback/ -run 'TestClassifyReadsPreMaskBody' -v` → PASS
두 번째 단언이 순서 역전(조용한 미탐, `design.md` §3)을 잡는 장치다.

### AC-F-014 — 파이프라인 멱등성

```
Given 임의의 본문
When  Scrub 을 두 번 연속 적용한다 (2회차 입력 = 1회차 Body)
Then  두 번째 Body 가 첫 번째 Body 와 바이트 동일하다
```
`go test ./internal/feedback/ -run 'TestScrubIsIdempotent' -v` → PASS
재시도 큐에서 꺼낸 본문의 재스크럽과 확인 게이트 옵션 3(수정 후 재스크럽)이 이 성질에 의존한다.

### AC-F-015 — 마스킹 로그 내용 + 권한

```
Given 마스킹이 1건 이상 발생한 Scrub 실행 (프로젝트 루트는 t.TempDir())
When  .moai/logs/feedback-mask.log 를 읽고 os.Stat 로 모드를 본다
Then  마지막 줄이 RFC3339 시각·종류·건수를 담고 마스킹된 원문 값을 담지 않는다
  And 파일 perm 이 0o600 이다 (Windows 에서는 권한 단언만 skip)
```
`go test ./internal/feedback/ -run 'TestMaskLog' -v` → PASS

### AC-F-016 — 로그 fail-open

```
Given .moai/logs 가 쓰기 불가인 상태
When  Scrub 을 호출한다
Then  Scrub 이 에러 없이 정상 Result 를 반환한다
```
`go test ./internal/feedback/ -run 'TestMaskLogFailureIsFailOpen' -v` → PASS

### AC-F-017 — 전송 실패 → 큐 적재

```
Given 전송이 실패했다고 표시된 상태 (프로젝트 루트는 t.TempDir())
When  큐에 적재한다
Then  .moai/state/feedback/queue.json 이 존재하고 항목 1건을 담으며 그 본문이 마스킹된 본문이다
```
`go test ./internal/feedback/ -run 'TestQueueEnqueuesOnSendFailure' -v` → PASS

### AC-F-018 — 성공 → 큐에서 제거

```
Given AC-F-017 직후의 큐 상태(1건)
When  같은 항목의 전송 성공을 표시한다
Then  항목 수가 0 이고 파일이 유효한 JSON 으로 남는다
```
`go test ./internal/feedback/ -run 'TestQueueResolvesOnSuccess' -v` → PASS

### AC-F-019 — 스킬 [HARD] 조항 4종 (두 사본 각각)

```
Given 소스 스킬과 템플릿 미러
When  산문 표면이 지는 [HARD] 의무 4종을 각 사본에서 grep 한다
Then  두 파일 모두에서 5개 관측이 전부 1건 이상이다
```

두 사본 각각에 대해(`SKILL=.claude/skills/moai/workflows/feedback.md`, 그리고 `SKILL=internal/template/templates/.claude/skills/moai/workflows/feedback.md`):

```bash
grep -c 'moai feedback scrub' "$SKILL"        # ① 스크러버 경유            >= 1
grep -c -- '--title' "$SKILL"                 # ② 제목 전달 (REQ-3 · REQ-10) >= 1
grep -c 'verdict' "$SKILL"                    # ③ 3문장 fail-closed 조항     >= 1
grep -c 'feedback-draft' "$SKILL"             # ④ 실패 부류 분기 (REQ-9 D4)  >= 1
grep -c 'conversation_language' "$SKILL"      # ⑤ 라벨 언어 의무 (D11)       >= 1
```

**② 가 이 AC의 핵심이다.** iter2에서 REQ-3·REQ-10·plan M6 세 곳이 제목 스크럽을 지시했는데 관측은 ① 하나뿐이어서, **본문만 파이프하는 스킬 본문이 24개 AC를 전부 통과하고 제목이 마스킹 없이 제출되는 상태**가 가능했다. ②가 없으면 D1은 문서상으로만 닫힌다.

③은 타임아웃·파싱 불가를 포함한 3문장 조항의 존재를 보되, `grep -c 'verdict'` 는 존재만 확인하므로 **문장 수까지** 볼 것: `grep -n '종료 코드\|verdict\|60초' "$SKILL"` 이 세 조건을 각각 담은 줄을 보여야 한다.

추가로 verbatim 보존 규칙의 명시적 예외 문장이 ①과 같은 절 범위에 있음을 `grep -n` 줄 번호로 확인한다.

**두 사본을 각각 도는 이유**: 소스만 고치고 템플릿 미러를 빠뜨리면 배포 사용자는 통제 없는 워크플로를 받는다. 사본별 합산이 아니라 **사본별 전수**다.

### AC-F-020 — 마법사 질문 + 개수 고정 유지

```
Given InitQuestions(root) 결과
When  Page3 그룹에서 feedback_auto_submit 을 찾는다
Then  존재하고 Type 이 QuestionTypeConfirm 이며 Default 가 "false" 이다
  And DefaultQuestions 는 여전히 5개, ReconfigureQuestions 는 12개다
```
`go test ./internal/cli/wizard/ -run 'TestFeedbackAutoSubmitQuestion|TestQuestionOrder|TestReconfigureQuestions' -v` → PASS

### AC-F-021 — 4로케일 번역 완전성

```
Given 신규 질문이 추가된 상태
When  번역 완전성 테스트를 실행한다
Then  통과한다 (ko/ja/zh title+description 이 모두 존재)
```
`go test ./internal/cli/wizard/ -run 'TestWizardQuestionTranslationCompleteness' -v` → PASS

### AC-F-022 — 답변이 실제로 파일에 기록됨

```
Given 마법사 결과가 feedback_auto_submit=true 로 채워졌다
When  WritePhase1Configs 를 실행한다
Then  .moai/config/sections/feedback.yaml 에 auto_submit: true 가 있고 기존 주석이 보존된다
```
`go test ./internal/core/project/ -run 'TestWritePhase1ConfigsPersistsFeedbackAutoSubmit' -v` → PASS
이 AC가 §B Known Issues 1(사장 코드 배선)에 대한 방어다 — 질문이 물어지고 버려지면 여기서 FAIL한다.

### AC-F-023 — 웹 등록 + 템플릿 + 인벤토리 + 빌드 + 중립성

```
Given 신규 키가 추가되고 feedback 섹션이 RouteSeam 으로 재개방된 상태
When  아래 6개 관측을 수행한다
Then  전부 기대와 일치한다
```
```bash
go test ./internal/settings/ ./internal/web/ -run 'TestSchemaCurrentValuesReadsAllSections|TestI18nKeySetParity|TestRouteForSectionTable|TestExcludedSectionsAllRejected|TestScopeContract' -v   # PASS
grep -n 'auto_submit' internal/template/templates/.moai/config/sections/feedback.yaml          # 1건
grep -n 'feedback.auto_submit' internal/config/testdata/shipped_key_inventory.yaml             # 1건
grep -rn 'SPEC-FEEDBACK-AUTO-SUBMIT\|REQ-' internal/template/templates/.moai/config/sections/feedback.yaml internal/template/templates/.claude/skills/moai/workflows/feedback.md   # 0건
make build                                                                                      # exit 0
```

**[HARD] `-v` 출력에서 위 5개 테스트 이름이 각각 `=== RUN` 으로 찍히는지 확인한다.** iter1의 이 AC는 `TestSchemaLabel` 과 `TestSectionRoute` 를 가리켰는데 둘 다 실재하지 않아(측정: `internal/web/schema_label_test.go` → `TestSchemaEmptyLabelParity:16` / `TestI18nKeySetParity:74` / `TestI18nSegmentKeysRemovedFromWebDictionary:133`; `internal/settings/sectionroute_test.go` → `TestRouteForSectionTable:8` / `TestSeamSectionsMatchesRoutes:51` / `TestExcludedSectionsAllRejected:74`) 선택자 절반이 0개 테스트에 매칭되고 `ok` 로 조용히 통과했다. 판정 전 확인:

```bash
grep -c 'func TestI18nKeySetParity' internal/web/schema_label_test.go            # 1
grep -c 'func TestRouteForSectionTable' internal/settings/sectionroute_test.go   # 1
grep -c 'func TestExcludedSectionsAllRejected' internal/settings/sectionroute_test.go  # 1
```

라우트 반전 확인: `sectionroute_test.go:27` 과 `scope_contract_test.go:79` 의 기대값이 `RouteSeam`(및 excluded 목록에서 제거)으로 갱신돼 있고, 그 갱신이 커밋 메시지에 SPEC-WEBCONF-SIMPLIFY-001 M3 반전으로 명시돼야 한다.

### AC-F-024 — 개인키 블록이 통째로 마스킹된다 (과소 마스킹 방어)

```
Given 본문에 PEM 개인키 블록 전체가 들어 있다 — BEGIN 헤더 + 본문 여러 줄 + END 종료자
      (테스트는 더미 키 블록을 소스에서 조립한다; 실제 키를 픽스처로 커밋하지 않는다)
When  Scrub 을 호출한다
Then  Result.Body 에 헤더도, END 종료자도, 그 사이 본문 어느 줄도 남아 있지 않다
Given 같은 블록에서 END 종료자를 제거한 입력(잘린 키)
When  Scrub 을 호출한다
Then  헤더 이후 입력 끝까지가 마스킹돼 남은 본문 줄이 없다
```
`go test ./internal/feedback/ -run 'TestScrubMasksPrivateKeyBlockEntirely' -v` → PASS

D1b의 **과소 마스킹** 방향 관측이며, 이 SPEC이 낼 수 있는 최악의 결과를 직접 막는다. 원본 패턴(`internal/hook/pre_tool.go:263-264`)은 **헤더 마커만** 매칭하므로, 매칭 span을 그대로 치환하는 순진한 구현은 헤더만 지우고 키 본문 전체를 공개 이슈로 실어 보낸다 — 그 구현은 첫 케이스에서 FAIL한다. 두 번째 케이스는 종료자 부재 시 입력 끝까지 마스킹하는 규칙을 고정한다.

## §D.3 Indirect Verification

- **AC-F-002 의 baseline 반증**: 편집 전 FAIL을 관측하지 못하면 §A.1 정정 P1과 모순이므로 재측정 대상이다.
- **AC-F-007 이 합집합을 증명한다**: 추가로 `grep -n 'DefaultSecurityPolicy' internal/feedback/` 이 1건 이상이어야 한다 — 패턴 문자열 복사(AP-4) 방어.
- **AC-F-008·F-013 이 축퇴 구현을 배제한다**: 전부 마스킹/전부 차단하는 구현은 F-008에서, 순서 역전은 F-013 두 번째 단언에서 잡힌다.
- **AC-F-006 두 번째 케이스와 AC-F-003의 `title` 필드가 함께 D1을 닫는다**: 계약에 필드가 있어도 제목이 파이프라인을 지나지 않으면 F-006에서 FAIL한다. 둘 중 하나만으로는 부족하다.
- **`-run` 선택자의 실재 확인이 판정보다 먼저다**: iter1에서 AC-F-023의 선택자 절반(`TestSchemaLabel`·`TestSectionRoute`)이 0개 테스트에 매칭돼 `ok` 로 통과했다. 각 AC의 명령을 돌리기 전에 `grep -c 'func <TestName>' <file>` 이 `1` 인지 확인하고, `-v` 출력에서 `=== RUN` 줄을 센다.
- **AC-F-022 가 사장 경로를 배제한다**: 답변이 파일에 도달하는지를 직접 관측한다.
- **형제 SPEC 병합 확인**: `SPEC-TODO-ENABLE-FLAG-001` 이 먼저 착지했다면, AC-F-020·F-021 을 병합 후 트리에서 다시 돌려 두 질문이 함께 통과함을 확인한다(`spec.md` §E.1 [HARD]).

## §D.4 Closure Gate (Definition of Done)

- [ ] §D.2 전 AC 실행, 관측 출력이 기대와 일치.
- [ ] AC-F-002 의 편집 전 FAIL / 편집 후 PASS 를 둘 다 관측.
- [ ] 패키지 스코프 테스트 전부 초록(`feedback`, `sandbox`, `config`, `cli`, `cli/wizard`, `core/project`, `settings`, `web`, `template`).
- [ ] AC-F-023의 5개 선택자가 `-v` 출력에서 각각 `=== RUN` 으로 찍혔다(0개 실행 아님).
- [ ] 라우트 반전(SPEC-WEBCONF-SIMPLIFY-001 M3)이 커밋 메시지에 명시됐다.
- [ ] `go test -race ./internal/feedback/...` 초록.
- [ ] `GOOS=windows go vet` (plan.md M9 목록) exit 0.
- [ ] `golangci-lint run --timeout=2m` clean.
- [ ] `make build` exit 0.
- [ ] 결정 D5 확정, 선택한 쪽이 커밋 메시지에 기록됨.
- [ ] docs-site 4로케일 문서가 같은 PR에 반영됨.
- [ ] 완료 보고에 §E.3 잔여 위험 3건이 그대로 실려 있고 "마스킹이 강제된다"는 표현을 쓰지 않았다.

## §D.5 Forward-Looking Checks (머지 이후)

- 임시 디렉터리에서 `moai init` 대화형 1회 — 질문이 자기 로케일로 뜨고 답변이 파일에 남는지(수동).
- 웹 콘솔에서 토글 렌더·저장 확인(수동).
- `auto_submit: false` 로 실제 피드백 1건 제출 — 게이트가 마스킹 본문 전문과 findings 요약을 보여주는지(수동).
- 릴리즈 후 `verdict: blocked` 오탐 보고가 들어오면 분류 임계값 재조정 후속 카드.

## §D.6 Quality Gate Criteria (TRUST 5)

- **Tested**: `internal/feedback` 커버리지 85% 이상. 보안 조항은 양극성 쌍으로 커버.
- **Readable**: 공개 함수 godoc. 마스킹 형태를 채택한 이유와 파이프라인 순서 근거를 주석 1줄씩.
- **Unified**: `gofmt` + `golangci-lint` clean. 패턴 컴파일은 기존 `(?i)` 관례.
- **Secured**: `findings`·로그에 원문 값 없음(F-005, F-015), 로그 `0o600`(F-015), 제출 fail-closed(F-004), 분류가 원문을 봄(F-013).
- **Trackable**: 마일스톤 단위 Conventional Commits, footer 에 SPEC ID.
