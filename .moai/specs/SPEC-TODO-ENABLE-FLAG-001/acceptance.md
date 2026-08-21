# Acceptance — SPEC-TODO-ENABLE-FLAG-001

> 각 AC는 판정을 내리는 **명령 하나**와 한 번의 관측으로 결정되는 **기대치**를 명시한다.
> 이름 붙은 테스트 함수는 해당 마일스톤의 **납품물**이다 — 먼저 쓰고(RED) 판정한다. 존재하지 않는 테스트를 가리키는 `-run` 패턴, 아무것도 실행하지 않고 통과하는 AC는 둘 다 금지한다.
> [HARD] 충족 불가능한 문구의 AC를 쓰지 않는다 — "사용 안 함이면 todo 안내가 전혀 뜨지 않는다"는 AC는 이 문서에 **없다**(`spec.md` §A.1 정정 P4).
> 로컬 검증은 패키지 스코프로만. 전 패키지 판정은 CI 몫(CLAUDE.local.md §4).

**AC 총계: 11 / Tier M 상한 16.**

## §D AC Matrix

| AC ID | REQ | Severity | 요약 |
|---|---|---|---|
| AC-T-001 | REQ-1 | MUST-PASS | 키 해석 4케이스: 부재→활성, `false`→비활성, `true`→활성, 잘못된 값→활성 |
| AC-T-002 | REQ-2 | MUST-PASS | SessionStart 백로그 문구 억제 + 대조 |
| AC-T-003 | REQ-2 | MUST-PASS | statusline TODO 세그먼트 억제 + 대조 + OR 판정 |
| AC-T-004 | REQ-2 | MUST-PASS | 라우팅 조건이 자동 라우팅으로 한정됨 + 목록 3줄 내용 불변(내용 단언) |
| AC-T-005 | REQ-3·2 | MUST-PASS | 플래그 false 여도 명령 등록 유지 + add→list 왕복 동작 |
| AC-T-006 | REQ-4 | MUST-PASS | 마법사 질문 존재·기본 true·개수 고정 테스트 유지 |
| AC-T-007 | REQ-4 | MUST-PASS | 4로케일 번역 완전성 |
| AC-T-008 | REQ-4 | MUST-PASS | 마법사 답변이 실제로 `workflow.yaml` 에 기록됨(사장 경로 배제) |
| AC-T-009 | REQ-5 | MUST-PASS | 필드 등록(존재·TypeBool·PersistSeam) + i18n 4로케일 — 실재 테스트 2개 |
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
Given 같은 키를 bool 로 해석되지 않는 값(enabled: maybe)으로 바꾼다
When  다시 읽는다
Then  TodoEnabled() == true (섹션 폴백 → 기본값 → 활성)
```
`go test ./internal/config/ -run 'TestTodoEnabled' -v` → PASS
네 케이스가 한 표에 들어간다. 첫 케이스가 `*bool` 요구(부재 ≠ false)를, 네 번째가 D3의 잘못된 값 결과를 증명한다. 네 번째 케이스는 `workflow` **섹션 전체**가 기본값으로 되돌아가는 경로를 탄다(`internal/config/loader.go:226-237`) — 그 폭발 반경은 이 SPEC이 만든 것도 바꾸는 것도 아니지만, 테스트가 그 경로를 실제로 지난다는 사실을 주석으로 남긴다.

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

### AC-T-004 — 라우팅 조건이 자동 라우팅으로 한정 + 목록 3줄 내용 불변

```
Given 소스 SKILL.md 와 템플릿 미러
When  todo 라우팅 조건 문장과 그 한정 범위를 grep 한다
Then  두 파일 모두 조건 문장이 1건 이상이다
  And 그 문장이 명시적 /moai todo 호출은 정상 동작한다는 한정을 담는다 (REQ-2 D2)
  And 아래 세 목록 줄이 각각 정확히 1건씩 그대로 존재한다
```
```bash
grep -c 'workflow.todo.enabled' .claude/skills/moai/SKILL.md                                    # >= 1
grep -c 'workflow.todo.enabled' internal/template/templates/.claude/skills/moai/SKILL.md        # >= 1
grep -c '명시적' .claude/skills/moai/SKILL.md                                                     # >= 1 (한정 문장)
```

**목록 3줄 내용 단언** — 줄 번호가 아니라 **내용**으로 고정한다(D5). 라우팅 절이 추가되면 줄 번호는 이동하지만 내용은 이동하지 않으며, `git diff` 와 달리 커밋 이후에도 판정이 유지된다. 아래 세 문자열은 편집 전 트리(`e7fb0e1d2`)에서 `sed -n '6p;81p;105p'` 로 읽은 값이다:

```bash
grep -Fxc -e '  feedback, review, clean, codemaps, gate, e2e, harness, goal, todo) to' .claude/skills/moai/SKILL.md   # 정확히 1
grep -Fc  -e '- **todo** (aliases: backlog): Backlog queue — the slash surface covers two acts' .claude/skills/moai/SKILL.md   # 정확히 1
grep -Fc  -e '- Backlog language (add to the backlog, note this for later, what should I work on next, remind me to) routes to **todo**' .claude/skills/moai/SKILL.md   # 정확히 1
```
세 명령 모두 `1` 이어야 한다. **`-e` 는 생략할 수 없다** — `-` 로 시작하는 패턴은 옵션으로 파싱돼 `invalid option` 으로 exit 2 가 된다(이 AC를 작성하며 실측). 셋 다 편집 전 트리(`e7fb0e1d2`)에서 각각 `1` 을 반환함을 확인했다. `0` 이면 목록 줄이 편집됐다는 뜻이고(범위 밖 위반), `2` 이상이면 중복 추가다. 사람이 diff 줄을 분류할 필요가 없는 단일 관측이다.

### AC-T-005 — CLI 등록 유지 + 플래그 false 하 동사 왕복

```
Given workflow.todo.enabled: false (프로젝트 루트는 t.TempDir())
When  rootCmd 에서 "todo" 명령을 찾고 --help 를 실행한다
Then  명령이 존재하고 종료 코드가 0이다
Given 같은 조건
When  moai todo add "<문구>" 후 moai todo 로 목록을 읽는다
Then  추가한 카드가 목록에 나타나고 두 호출 모두 종료 코드 0이다
```
`go test ./internal/cli/ -run 'TestTodoCommandRegisteredRegardlessOfFlag|TestTodoVerbsUnaffectedByFlag' -v` → PASS
첫 케이스가 REQ-3(등록 유지)를, 두 번째가 REQ-2 D2 조항의 **행동**(이름으로 부른 기능은 플래그와 무관하게 동작)을 관측한다. 슬래시 표면은 산문이라 기계 관측이 불가능하지만, 그 표면이 위임하는 실제 동작은 여기서 관측된다 — 이 왕복이 실패하면 "명시적 호출은 정상 동작한다"는 조항이 거짓이 된다.

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
Then  존재하고 Type 이 TypeBool, Persist.Kind 가 PersistSeam 이다   ← TestWorkflowTodoEnabledFieldRegistered 가 관측
  And 그 필드의 .title/.desc 가 4로케일 모두에 있다                  ← TestI18nKeySetParity 가 관측
```
```bash
go test ./internal/settings/ -run 'TestWorkflowTodoEnabledFieldRegistered' -v   # PASS
go test ./internal/web/ -run 'TestI18nKeySetParity' -v                          # PASS
```

두 관측 모두 **실재하는 이름**을 가리켜야 한다. iter1에서 이 AC는 `TestSchemaLabel` 을 가리켰는데 `internal/web/schema_label_test.go` 에는 그런 이름의 함수가 없어(`TestSchemaEmptyLabelParity:16` / `TestI18nKeySetParity:74` / `TestI18nSegmentKeysRemovedFromWebDictionary:133`) `-run` 이 0개 테스트에 매칭되고 "no tests to run" 으로 조용히 초록이 됐다. 판정 전 확인:

```bash
grep -c 'func TestI18nKeySetParity' internal/web/schema_label_test.go                        # 1
grep -c 'func TestWorkflowTodoEnabledFieldRegistered' internal/settings/schema_sections_test.go  # 1 (M5 납품물)
```
`TestSchemaCurrentValuesReadsAllSections` 는 이 AC에서 **제외**한다 — 무관한 13개 키의 고정 맵을 볼 뿐 신규 필드에 영향받지 않아 아무것도 관측하지 못한다.

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
Then  두 질문이 모두 존재한 채로 4개 테스트가 통과하고 vet 종료 코드가 0이다
```
```bash
go test ./internal/cli/wizard/ -run 'TestQuestionOrder|TestReconfigureQuestions|TestWizardQuestionTranslationCompleteness|TestTodoEnabledQuestion|TestFeedbackAutoSubmitQuestion' -v
GOOS=windows go vet ./internal/config/... ./internal/hook/... ./internal/statusline/... ./internal/cli/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/settings/...
```

`spec.md` §E.1의 충돌 해소 규칙 5조가 실제로 지켜졌는지를 판정하는 유일한 관측이다. **두 번째로 착지하는 쪽(해소 소유자)은 해소 직후 이 AC를 반드시 재실행한다** — 재실행 없이 해소를 끝났다고 보고할 수 없다. `-v` 출력에서 `TestTodoEnabledQuestion` 과 `TestFeedbackAutoSubmitQuestion` 이 **둘 다 RUN 으로 찍히는지** 확인한다: 해소 중 한쪽 항목을 잃으면 그 질문의 테스트가 조용히 실패하거나(개수 고정) 번역 완전성에서 뒤늦게 드러난다. 실패 시 해소를 되돌리고 리드에게 블로커로 보고하며, 테스트를 고쳐 통과시키지 않는다(규칙 5조).

## §D.3 Indirect Verification

- **대조 케이스가 축퇴 통과를 배제한다**: AC-T-002·003의 "키 제거 시 나타난다" 절이 없으면, 원래부터 출력되지 않는 조건에서 테스트가 헛통과한다.
- **AC-T-003 세 번째 케이스**가 기존 `statusline.yaml` 억제 경로의 생존을 증명한다.
- **AC-T-008이 사장 경로를 배제한다**: 답변이 파일에 도달하는지를 직접 관측한다.
- **범위 밖 표면은 AC로 확인하지 않는다**: 상시 로드 룰과 스킬 목록은 억제 대상이 아니므로 "여전히 보인다"를 통과 조건으로 삼지도, "안 보인다"를 요구하지도 않는다. AC-T-004의 내용 단언이 **건드리지 않았음**만 확인한다.
- **`-run` 패턴이 실재 테스트에 매칭되는지 먼저 본다**: iter1에서 AC-T-009가 존재하지 않는 `TestSchemaLabel` 을 가리켜 "no tests to run" 으로 조용히 초록이었다. 각 AC의 판정 명령을 돌리기 전에 `grep -c 'func <TestName>' <file>` 이 `1` 인지 확인한다 — 특히 이번에 추가된 납품 테스트 3개(`TestWorkflowTodoEnabledFieldRegistered`, `TestTodoVerbsUnaffectedByFlag`, `TestTodoCommandRegisteredRegardlessOfFlag`).
- **커밋 이후에도 판정이 유지되는가**: iter1의 AC-T-004는 working-tree `git diff` 로 판정해 커밋 순간 무조건 통과했다. 대체된 내용 단언(`grep -Fxc` / `grep -Fc`)은 트리 상태와 무관하게 같은 값을 낸다.

## §D.4 Closure Gate (Definition of Done)

- [ ] §D.2 전 AC 실행, 관측 출력이 기대와 일치.
- [ ] 패키지 스코프 테스트 전부 초록(`config`, `hook`, `statusline`, `cli`, `cli/wizard`, `core/project`, `settings`, `web`, `template`).
- [ ] AC-T-011의 `GOOS=windows go vet` exit 0.
- [ ] `golangci-lint run --timeout=2m` clean.
- [ ] `make build` exit 0.
- [ ] M6 템플릿 결정이 커밋 메시지에 기록됨.
- [ ] 완료 보고가 억제 범위를 **런타임 표면 3종으로 한정해** 서술하고, 상시 로드 룰·스킬 목록이 범위 밖임을 명시했다. "todo 안내를 전부 껐다"는 표현을 쓰지 않았다.
- [ ] 형제 SPEC이 이미 착지했다면 AC-T-011을 병합 트리에서 실행했고, 충돌을 해소했다면 해소 **직후** 다시 실행했다(`spec.md` §E.1 규칙 4).
- [ ] 이 SPEC이 두 번째 착지자라면 §E.1 해소 규칙 5조를 따랐다 — 양쪽 항목 보존, 재배치 없음, 테스트를 고쳐 통과시키지 않음.
- [ ] PR 경로(Route B)로 착지했다 — 이 저장소는 전 티어 PR이며 `main` 직접 푸시는 거부된다.

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
