# Acceptance — SPEC-TODO-ENABLE-FLAG-001

> 각 AC는 판정을 내리는 **명령 하나**와 한 번의 관측으로 결정되는 **기대치**를 명시한다.
> 이름 붙은 테스트 함수는 해당 마일스톤의 **납품물**이다 — 먼저 쓰고(RED) 판정한다. 존재하지 않는 테스트를 가리키는 `-run` 패턴, 아무것도 실행하지 않고 통과하는 AC는 둘 다 금지한다.
> [HARD] 충족 불가능한 문구의 AC를 쓰지 않는다 — "사용 안 함이면 todo 안내가 전혀 뜨지 않는다"는 AC는 이 문서에 **없다**(`spec.md` §A.1 정정 P4).
> 로컬 검증은 패키지 스코프로만. 전 패키지 판정은 CI 몫(CLAUDE.local.md §4).

**AC 총계: 11 / Tier M 상한 16.**

## §D AC Matrix

| AC ID | REQ | Severity | 요약 |
|---|---|---|---|
| AC-T-001 | REQ-1 | MUST-PASS | 키 해석: 부재→활성, `false`→비활성, `true`→활성 |
| AC-T-002 | REQ-2 | MUST-PASS | SessionStart 백로그 문구 억제 + 대조 |
| AC-T-003 | REQ-2 | MUST-PASS | statusline TODO 세그먼트 억제 + 대조 + OR 판정 |
| AC-T-004 | REQ-2 | MUST-PASS | 스킬 라우팅 조건이 두 사본에 존재하고 목록 메타데이터는 미변경 |
| AC-T-005 | REQ-3 | MUST-PASS | 플래그 false 여도 `moai todo` 명령이 등록·동작 |
| AC-T-006 | REQ-4 | MUST-PASS | 마법사 질문 존재·기본 true·개수 고정 테스트 유지 |
| AC-T-007 | REQ-4 | MUST-PASS | 4로케일 번역 완전성 |
| AC-T-008 | REQ-4 | MUST-PASS | 마법사 답변이 실제로 `workflow.yaml` 에 기록됨(사장 경로 배제) |
| AC-T-009 | REQ-5 | MUST-PASS | 웹 스키마 등록 + i18n 4로케일 |
| AC-T-010 | REQ-6 | MUST-PASS | 템플릿 결정 반영 + 인벤토리 정합 + `make build` + 중립성 |
| AC-T-011 | 전체 | MUST-PASS | 형제 SPEC 병합 후 마법사 2질문 공존 + `GOOS=windows go vet` |

전 항목 MUST-PASS.

## §D.2 Given-When-Then Scenarios

### AC-T-001 — 키 해석 3케이스

```
Given workflow.yaml 에 todo 블록이 없는 임시 프로젝트
When  설정을 읽는다
Then  TodoEnabled() == true
Given workflow.yaml 에 "todo:\n    enabled: false\n" 를 기록한다
When  다시 읽는다
Then  TodoEnabled() == false
Given 같은 키를 true 로 바꾼다
When  다시 읽는다
Then  TodoEnabled() == true
```
`go test ./internal/config/ -run 'TestTodoEnabled' -v` → PASS
세 케이스가 한 표에 들어가며, 첫 케이스가 `*bool` 요구(부재 ≠ false)를 증명한다.

### AC-T-002 — SessionStart 백로그 문구 억제 (대조 포함)

```
Given workflow.todo.enabled: false
  And 백로그에 대기 카드 1건 이상, kanban 환경 + source=="startup" 충족
When  SessionStart 훅 출력을 만든다
Then  출력에 4로케일 백로그 안내 문구 중 어느 것도 포함되지 않는다
Given 같은 조건에서 todo 키만 제거한다
When  다시 출력을 만든다
Then  안내 문구가 포함된다
```
`go test ./internal/hook/ -run 'TestSessionStartKanbanRespectsTodoDisabled' -v` → PASS
대조 케이스가 "원래 안 나오던 것"을 통과로 오독하는 것을 막는다.

### AC-T-003 — statusline 세그먼트 억제 + OR 판정

```
Given workflow.todo.enabled: false, 백로그 데이터 존재, statusline.yaml 에 backlog 키 없음
When  렌더한다
Then  출력에 "TODO:" 문자열이 없다
Given todo 키를 제거하고 statusline.yaml 에도 backlog 키가 없다
When  렌더한다
Then  "TODO:" 가 나타난다
Given todo 키는 true 이고 statusline.yaml 에 backlog: false 를 둔다
When  렌더한다
Then  "TODO:" 가 없다 (기존 억제 경로가 살아 있다)
```
`go test ./internal/statusline/ -run 'TestRendererBacklogSegmentGating' -v` → PASS
세 번째 케이스가 신규 플래그가 기존 경로를 덮어쓰지 않음을 증명한다(`spec.md` §E.3).

### AC-T-004 — 스킬 라우팅 조건 존재 + 목록 미변경

```
Given 소스 SKILL.md 와 템플릿 미러
When  todo 라우팅 조건 문장을 grep 한다
Then  두 파일 모두 1건 이상이다
  And 스킬 목록 메타데이터 줄(:6, :81, :105 부근)은 편집 전과 동일하다
```
```bash
grep -c 'workflow.todo.enabled' .claude/skills/moai/SKILL.md
grep -c 'workflow.todo.enabled' internal/template/templates/.claude/skills/moai/SKILL.md
git diff --unified=0 .claude/skills/moai/SKILL.md | grep '^[-+].*todo' # 라우팅 절 외 목록 줄 변경 0건
```

### AC-T-005 — CLI 등록 유지

```
Given workflow.todo.enabled: false
When  rootCmd 에서 "todo" 명령을 찾고 --help 를 실행한다
Then  명령이 존재하고 종료 코드가 0이다
```
`go test ./internal/cli/ -run 'TestTodoCommandRegisteredRegardlessOfFlag' -v` → PASS
REQ-3의 결정을 고정한다 — 나중에 "안내를 끄니 명령도 끄자"로 흐르는 것을 막는다.

### AC-T-006 — 마법사 질문 + 개수 고정 유지

```
Given InitQuestions(root) 결과
When  Page3 그룹에서 todo_enabled 를 찾는다
Then  존재하고 Type 이 QuestionTypeConfirm 이며 Default 가 "true" 이다
  And DefaultQuestions 는 여전히 5개, ReconfigureQuestions 는 12개다
```
`go test ./internal/cli/wizard/ -run 'TestTodoEnabledQuestion|TestQuestionOrder|TestReconfigureQuestions' -v` → PASS

### AC-T-007 — 4로케일 번역 완전성

```
Given 신규 질문이 추가된 상태
When  번역 완전성 테스트를 실행한다
Then  통과한다 (ko/ja/zh title+description 이 모두 존재)
```
`go test ./internal/cli/wizard/ -run 'TestWizardQuestionTranslationCompleteness' -v` → PASS

### AC-T-008 — 답변이 실제로 파일에 기록됨

```
Given 마법사 결과가 todo_enabled=false 로 채워졌다
When  WritePhase1Configs 를 실행한다
Then  .moai/config/sections/workflow.yaml 에 todo.enabled: false 가 있다
  And 파일의 기존 주석과 다른 키들이 보존돼 있다
```
`go test ./internal/core/project/ -run 'TestWritePhase1ConfigsPersistsTodoEnabled' -v` → PASS
이 AC가 §B Known Issues 1(사장 코드 배선)의 방어다 — 질문이 물어지고 버려지면 여기서 FAIL한다.

### AC-T-009 — 웹 스키마 등록 + i18n

```
Given settings.AllFields() 목록
When  workflow.todo.enabled 필드를 찾는다
Then  존재하고 Type 이 TypeBool, Persist.Kind 가 PersistSeam 이다
  And i18n 완전성 검사가 4로케일 모두에서 통과한다
```
`go test ./internal/settings/ ./internal/web/ -run 'TestSchemaCurrentValuesReadsAllSections|TestSchemaLabel' -v` → PASS

### AC-T-010 — 템플릿 결정 반영 + 인벤토리 정합 + 빌드

```
Given M6 의 템플릿 결정(싣지 않음 / 싣기)이 확정된 상태
When  아래 관측을 수행한다
Then  결정과 관측이 일치하고 빌드가 성공한다
```
```bash
grep -n 'todo' internal/template/templates/.moai/config/sections/workflow.yaml   # 싣지 않기로 했으면 0건, 싣기로 했으면 1건
go test ./internal/config/ -run 'TestShippedConfigKeysHaveReaders' -v            # PASS (둘 중 어느 결정이든)
grep -rn 'SPEC-TODO-ENABLE-FLAG\|REQ-' internal/template/templates/.moai/config/sections/workflow.yaml internal/template/templates/.claude/skills/moai/SKILL.md   # 0건
make build                                                                        # exit 0
```
첫 grep의 기대값은 결정에 따라 달라지므로, **결정을 커밋 메시지에 명시하고 그 값을 기대치로 고정**한다. 결정 없이 이 AC를 통과시킬 수 없다.

### AC-T-011 — 형제 SPEC 공존 + Windows vet

```
Given SPEC-FEEDBACK-AUTO-SUBMIT-001 과 이 SPEC 이 모두 착지한 트리
      (한쪽만 착지했다면 착지한 쪽 기준으로 판정하고, 나중 착지 시 재실행)
When  마법사 테스트와 크로스 vet 을 돌린다
Then  두 질문이 함께 통과하고 vet 종료 코드가 0이다
```
```bash
go test ./internal/cli/wizard/ -run 'TestQuestionOrder|TestReconfigureQuestions|TestWizardQuestionTranslationCompleteness|TestTodoEnabledQuestion|TestFeedbackAutoSubmitQuestion' -v
GOOS=windows go vet ./internal/config/... ./internal/hook/... ./internal/statusline/... ./internal/cli/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/settings/...
```
`spec.md` §E.1의 [HARD] 병합 규율이 실제로 지켜졌는지를 판정하는 유일한 관측이다.

## §D.3 Indirect Verification

- **대조 케이스가 축퇴 통과를 배제한다**: AC-T-002·003의 "키 제거 시 나타난다" 절이 없으면, 원래부터 출력되지 않는 조건에서 테스트가 헛통과한다.
- **AC-T-003 세 번째 케이스**가 기존 `statusline.yaml` 억제 경로의 생존을 증명한다.
- **AC-T-008이 사장 경로를 배제한다**: 답변이 파일에 도달하는지를 직접 관측한다.
- **범위 밖 표면은 AC로 확인하지 않는다**: 상시 로드 룰과 스킬 목록은 억제 대상이 아니므로 "여전히 보인다"를 통과 조건으로 삼지도, "안 보인다"를 요구하지도 않는다. AC-T-004의 두 번째 단언(목록 줄 미변경)이 **건드리지 않았음**만 확인한다.

## §D.4 Closure Gate (Definition of Done)

- [ ] §D.2 전 AC 실행, 관측 출력이 기대와 일치.
- [ ] 패키지 스코프 테스트 전부 초록(`config`, `hook`, `statusline`, `cli`, `cli/wizard`, `core/project`, `settings`, `web`, `template`).
- [ ] AC-T-011의 `GOOS=windows go vet` exit 0.
- [ ] `golangci-lint run --timeout=2m` clean.
- [ ] `make build` exit 0.
- [ ] M6 템플릿 결정이 커밋 메시지에 기록됨.
- [ ] 완료 보고가 억제 범위를 **런타임 표면 3종으로 한정해** 서술하고, 상시 로드 룰·스킬 목록이 범위 밖임을 명시했다. "todo 안내를 전부 껐다"는 표현을 쓰지 않았다.
- [ ] 형제 SPEC이 이미 착지했다면 AC-T-011을 병합 트리에서 실행했다.

## §D.5 Forward-Looking Checks (머지 이후)

- 임시 디렉터리에서 `moai init` 대화형 1회 — 질문이 자기 로케일로 뜨고 기본값이 "사용함"인지(수동).
- 웹 콘솔에서 토글 렌더·저장 확인(수동).
- 플래그를 끈 실제 세션에서 SessionStart 출력과 statusline을 눈으로 확인하고, 상시 로드 룰은 여전히 로드된다는 사실을 함께 기록(경계의 실증).

## §D.6 Quality Gate Criteria (TRUST 5)

- **Tested**: 신규 판독 헬퍼와 억제 가드에 억제/대조 쌍 테스트. 변경 패키지 커버리지 하락 없음.
- **Readable**: `*bool`을 쓴 이유(부재 ≠ false)를 필드 주석 1줄로.
- **Unified**: `gofmt` + `golangci-lint` clean. 판독은 기존 `readMCPToolEnablement` 관용구를 따른다.
- **Secured**: 해당 없음(설정 가시성 변경이며 신뢰 경계를 건드리지 않는다). 기능 제거가 아니라 안내 억제이므로 CLI 진입점이 유지된다(AC-T-005).
- **Trackable**: 마일스톤 단위 Conventional Commits, footer에 SPEC ID.
