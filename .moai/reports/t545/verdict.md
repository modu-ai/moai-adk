# t545 판정 — /settings 렌더 폼 name 유일성 고정 테스트

- 카드: t545 (Tier S · Class C, 회귀 가드 보강형)
- 워크트리: `.claude/worktrees/t545`, 브랜치 `WT-settings-form-names`
- 기준 트리: `5b9607218` (local develop `247d0985a` 를 `--no-ff` 로 병합한 커밋, 메시지에 t545)
- 산출물: `internal/web/settings_form_names_test.go` (신규 테스트 1건, 제품 코드 변경 없음)

## Claim

1. 가드가 없는 층을 develop 트리에서 재현했다. 중복 폼값 거절 가드는 저장 파서(`parseSchemaForm`)에만 있고 렌더 층에는 없다. 텍스트 필드 name 을 두 번 출하하는 뮤턴트를 넣어도 기존 `internal/web` 스위트는 전부 통과한다.
2. 신규 테스트 `TestSettingsRenderFormNamesUnique` 는 원본 트리에서 통과하고, 같은 뮤턴트에서는 렌더 단계에서 실패하며 중복된 name 세 개를 정확히 지목한다.
3. 복원한 트리에서 신규 테스트와 영향 패키지 `./internal/web/...` 전체가 통과한다.
4. 제품 코드 수리는 필요 없다. 원본 렌더에는 settings 폼 안의 비라디오 name 중복이 0건이다.

## Evidence

### 가드 위치 (재현 전제)

```
internal/web/schemaform.go:336:			errs[f.Name] = "duplicate form values submitted"
internal/web/schemaform.go:340:			errs[f.Name] = "duplicate form values submitted"
internal/web/handlers.go:403:	schemaEdits, schemaErrs := parseSchemaForm(r, currentSchemaValues)
```

`parseSchemaForm` 의 비테스트 호출처는 `handlers.go:403` 한 곳이다. 가드는 `r.PostForm[f.Name]` 또는 `f.Name+"__present"` 의 값이 둘 이상이면 해당 필드를 거절한다. 렌더 결과를 검사하는 테스트는 없었다. 부재 스캔(`absence-scan.txt`)은 53건을 찾았고, 대조군으로 파서 가드 테스트 두 건이 걸렸으며 나머지는 폼 name 과 무관했다.

### 원본 렌더 측정 (임시 프로브, 커밋 전 삭제)

`probe-pristine.txt`:

```
PROBE forms_in_page=3
PROBE settings-form: controls=198 distinct_names=125 dup_radio_only_names=54 dup_other_names=0
PROBE page: controls=199 distinct_names=126 dup_radio_only_names=54 dup_other_names=0
```

### 뮤턴트 (`mutant.patch`)

`internal/web/fieldsets_templ.go` 의 `default:` 분기(텍스트 행)에서 `schemaTextRow(...).Render` 를 한 번 더 호출한다. 생성 파일을 직접 고친 이유는 템플 재생성을 거치지 않고 렌더 출력만 바꾸기 위해서다.

`probe-mutant.txt`:

```
PROBE settings-form: controls=201 distinct_names=125 dup_radio_only_names=54 dup_other_names=3
PROBE settings-form dup_other: feedback.repository input:text
PROBE settings-form dup_other: git_strategy.worktree_base_branch input:text
PROBE settings-form dup_other: workflow.audit.codex.model input:text
```

### 뮤턴트 아래 기존 스위트 (가드 없는 층의 재현)

`gap-mutant-existing-tests.txt` (종료 코드 0):

```
ok  	github.com/modu-ai/moai-adk/internal/web	22.145s
```

`gap-mutant-existing-tests-v.txt` 에서 최상위 `--- PASS` 453건, `--- FAIL` 0건. 파서 가드 테스트도 그대로 통과했다.

```
--- PASS: TestParseSchemaFormDuplicateFormValuesNotSilentlyFirst (0.00s)
--- PASS: TestParseSchemaFormDuplicateCompanionRejected (0.00s)
```

### 신규 테스트: 원본 GREEN → 뮤턴트 RED → 복원 GREEN

| 단계 | 명령 | 종료 코드 | 증거 |
|---|---|---|---|
| 원본 | `go test ./internal/web/ -run 'TestSettingsRenderFormNamesUnique' -v -count=1` | 0 | `new-test-pristine-green.txt` |
| 뮤턴트 | 같은 명령 | 1 | `new-test-mutant-red.txt` |
| 복원 | 같은 명령 | 0 (`--- PASS` 1건) | `new-test-restored-green.txt` |

뮤턴트 단계 출력:

```
    settings_form_names_test.go:111: the /settings render ships duplicate form-control names inside #settings-form:
          feedback.repository x2 (input:text)
          git_strategy.worktree_base_branch x2 (input:text)
          workflow.audit.codex.model x2 (input:text)
        parseSchemaForm rejects the whole save when a name is submitted more than once; fix the template that renders the duplicate, not the parser
--- FAIL: TestSettingsRenderFormNamesUnique (0.01s)
```

복원 확인: `internal/web/fieldsets_templ.go` sha256 `0c38c94e3fcf73b95c97c742b15b268f80527ca07f9115dfbd746a7a505e781d` (뮤턴트 적용 전과 동일), `git diff --quiet -- internal/web/fieldsets_templ.go` 통과.

### 영향 패키지와 정적 검사 (복원 트리, 프로브 삭제 후)

```
go test ./internal/web/... -count=1   → WEB_PKG_EXIT=0
ok  	github.com/modu-ai/moai-adk/internal/web	42.107s
go vet ./internal/web/                → VET_EXIT=0 (vet-web-final.txt 빈 파일)
gofmt -l internal/web/settings_form_names_test.go → 출력 없음
```

## 설계 메모

**검사 범위는 페이지 전체가 아니라 `#settings-form` 이다.** 브라우저는 제출된 폼에 속한 컨트롤만 보낸다. 소속은 폼 하위 트리 안에 있거나 `form="settings-form"` 속성을 가진 경우다. 파서 가드도 이 폼의 저장 경로 한 곳에서만 호출된다. 페이지 전체로 넓히면 오탐이 난다. `shell.templ` 의 프로필 폼 세 개(생성 326행, 이름 변경 332행, 삭제 342행)는 서로 독립된 폼인데 컨트롤이 모두 `name="profile_name"` 을 쓴다(328·334·344행). 이름 변경·삭제 폼은 `len(vm.RenameTargets) > 0` 일 때만 렌더된다. 기본 테스트 렌더에서는 페이지 전체 폼이 3개이고 페이지 컨트롤이 settings 폼보다 1개 많아(199 대 198), 페이지 전체 스윕도 지금은 통과한다. 프로필이 둘 이상인 설정에서는 페이지 전체 스윕이 `profile_name` 을 중복으로 잘못 보고하게 된다. 이 조건부 렌더 상태는 이번에 측정하지 않았고 템플릿 코드에서 읽은 것이다.

**라디오 그룹은 예외로 둔다.** 같은 name 을 공유하는 라디오 중 체크된 하나만 제출되므로 중복이 아니다. 원본 렌더에는 이런 그룹이 54개 있다. 라디오와 비라디오가 name 을 공유하면 중복으로 본다.

**버튼류는 제외한다.** `submit`·`button`·`image`·`reset` 입력은 누른 버튼의 name 만 제출된다. 저장 버튼(`shell.templ:283-291`)에는 name 속성이 없다.

**공허 통과 방지.** `#settings-form` 을 찾지 못하거나, 라디오 그룹 수 또는 비라디오 name 수가 0이면 테스트가 곧바로 실패한다. 스윕이 아무것도 재지 않은 채 통과하는 경우를 막기 위한 장치다.

이 선택들은 브라우저 제출 규칙과 가드의 단일 호출처에서 바로 따라 나오므로 별도의 설계 판단으로 보지 않았다.

## Baseline-attribution

모든 측정은 이번 실행에서 워크트리 `.claude/worktrees/t545`, HEAD `5b9607218ddc19e1c6a53c6fa6782adfa10e70eb`, 브랜치 `WT-settings-form-names` 위에서 했다. 뮤턴트 측정은 이 트리에 `mutant.patch` 만 얹은 상태이고, 복원 후 해당 파일이 HEAD 와 같음을 sha256 과 `git diff --quiet` 로 확인했다. develop 흡수 후의 재측정은 아직 하지 않았다(아래 Gaps).

## Gaps

- 통합 창에서 develop 을 다시 흡수한 트리의 재측정은 아직 하지 않았다. 창을 받으면 병합 트리에서 신규 테스트와 `./internal/web/...` 를 다시 잰다.
- `make templ-generate` 는 돌리지 않았다. `.templ` 과 `_templ.go` 사이 생성 드리프트는 이 카드 범위 밖이며, 드리프트를 검사하는 테스트도 없다.
- 뮤턴트는 텍스트 행 한 종류만 넣었다. select·textarea·checkbox 중복과 `form=` 속성으로 외부에서 소속되는 컨트롤의 중복은 코드 경로로만 다루고 뮤턴트로 확인하지 않았다.
- 프로필이 둘 이상이라 이름 변경·삭제 폼까지 렌더되는 상태는 렌더해 보지 않았다. 범위 선택의 근거 중 이 부분은 템플릿 코드 판독에 기댄다.
- 전체 `internal/cli` 스위트와 다른 패키지는 돌리지 않았다. 변경은 `internal/web` 테스트 파일 하나다.

## Residual-risk

- `disabled` 컨트롤도 집계에 넣는다. 브라우저는 비활성 컨트롤을 제출하지 않으므로, 앞으로 name 이 같은 비활성 컨트롤을 의도적으로 두면 이 테스트가 과하게 실패한다. 지금 렌더에는 해당 사례가 없다.
- 라디오 그룹끼리의 중복, 예를 들어 같은 name 의 라디오 그룹이 폼 안에 두 번 렌더되는 경우는 한 그룹으로 합쳐 보이므로 잡지 못한다.
- `newTestApp` 의 기본 설정으로 렌더한 결과만 본다. 설정값에 따라 조건부로 나타나는 필드가 중복을 만들면 이 테스트 밖에 있다.
- `html.Parse` 는 중첩 폼을 버린다. 템플릿이 폼을 중첩하면 소속 판정이 브라우저와 어긋날 수 있다.
