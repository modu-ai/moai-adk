auditor-model: claude-opus-5-5

# t1972 sync 감사 판정 — moai init 설문·프로필 동기화 잔여 결함 5건 (Class B)

verdict: FAIL
audited_sha: 388967b934342b3c9a9e2a6939bc36331958e588

- 감사 범위: 9b632c1eb3..388967b934 (8e0ca7660f F4+F5, 9a5dc4b672 F6+골든, 388967b934 N6+N7)
- 트리: /Users/goos/.moai/worktrees/projects/moai-adk-go/t1972 (브랜치 WT-moai-init-f4), 감사 전후 `git status --short` 비어 있음, HEAD 불변
- 종합 점수: 0.74 (Functionality 65 / Security 95 / Craft 75 / Consistency 75)
- 판정 근거: F6 수리가 SPEC-INIT-TUX-I18N-001 REQ-ITI-015("select 필드 마지막 옵션 아래 빈 카드 줄 없음")를 새로 어긴다(실측, in-diff). must-pass 차원 Functionality 미달 → FAIL. F4·F5·N6은 수리 정확, N7은 부분.
- 교차 모델: codex_audit 1회(운영자 결정 d-20261106T055816Z-449d), native, delta_base 9b632c1eb3 → `pass`, 결함 0건, receipt `rcpt-88eb22e66235cd24de99b525`. Codex는 정적 검토만 했고(“테스트·빌드·재현 명령은 실행하지 않았어”) 아래 F6-R1은 렌더 실측으로만 드러난다. 두 판정이 엇갈리며, 이 감사는 실측을 근거로 FAIL을 유지한다.

## 차원 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 65 | FAIL | F4·F5·N6 수리 정확, 수리 전 코드에서 새 테스트 전부 RED 확인. F6은 80열 영어에서만 맞고 ko/ja/zh와 80열 외 폭에서 빈 카드 줄을 새로 만들거나 옵션을 여전히 숨긴다(아래 Evidence E7·E8) |
| Security (25%) | 95 | PASS | 비밀값·입력 경계 변화 없음. N6은 깨진 파일을 건드리지 않고 거부(데이터 보존 쪽). F5는 대화형 경로에서 도달 불가(E5) |
| Craft (20%) | 75 | PASS | RED→GREEN 4건 모두 실측. F6 테스트는 하한만 검사하고 상한(빈 줄)·비영어 렌더를 보지 않아 회귀를 못 잡음. 커버리지 미측정 |
| Consistency (15%) | 75 | PASS | gofmt 깨끗. wizard.go 문서 주석이 이제 거짓, F6 테스트가 prod와 다른 lipgloss(v1) 사용 |

## 결함 목록 (structured defect-list)

- **F6-R1** [Medium] [blocking] [in-diff, 재현 있음] `internal/cli/wizard/wizard.go:319-344` — `selectHeight`가 모든 로케일의 최댓값을 고정 74열 기준으로 잰다. 80열에서 init 2쪽(autonomy_tier) 마지막 옵션 아래 빈 카드 줄(`┃` 뒤 공백뿐)이 ko 0→2, ja 0→1, zh 1→4줄로 늘고, 영어도 120열에서 4줄 생긴다. REQ-ITI-015 위반. 이 설계(로케일 최댓값 + 고정 76열 행 계산)는 t1966에서 감사 차단 사유 B1("fixed 76-column row counting")·D1(짧은 로케일의 빈 카드 줄)로 되돌려진 바로 그 계보다(커밋 5c3f0ebf27 메시지). 요구 수정: 다음 중 하나를 고르고 진행 기록에 남긴다 — (a) 높이를 실행 중 로케일과 실제 폭으로 다시 잰다, (b) 범위를 좁힌다: 80열·로케일 최댓값 가정을 `@MX:DEBT`(+`@MX:CEILING`/`@MX:UPGRADE`)로 남기고 REQ-ITI-015 이탈을 운영자가 명시 수용, 문서 주석 정정, (c) t1966 선례대로 F6을 별도 카드로 분리하고 이 카드에서 되돌린다.
- **F6-R2** [Medium] [blocking — R1과 같은 원인] [in-diff, 재현 있음] 같은 위치 — 60열 터미널에서는 영어 마지막 옵션의 줄바꿈 부분("requires sandbox proof (Docker/gVisor/etc.)")이 여전히 가려진다. 카드 F6("감싼 옵션 줄을 세지 않음")이 80열에서만 고쳐졌다. R1의 수정으로 함께 처리된다.
- **F6-R3** [Low] [optional] [in-diff] `wizard.go:309-312` — 문서 주석이 여전히 "renders every option with no blank card rows (REQ-ITI-015, AC-ITI-016)"라고 적는다. 새로 덧붙인 문단(로케일 최댓값)과 모순되며 실측상 거짓. 요구 수정: R1의 선택에 맞춰 주석을 사실대로 고친다.
- **F6-R4** [Low] [optional] [in-diff] `internal/cli/wizard/select_height_t1972_test.go` — 하한만 단정하므로 과대 계산(빈 카드 줄)을 잡지 못하고, 골든은 영어 하나뿐이다. 또 설명 높이를 `github.com/charmbracelet/lipgloss`(v1)로 재는데 prod는 `charm.land/lipgloss/v2`를 쓴다. 요구 수정: 로케일별 렌더에서 빈 카드 줄 0을 단정하는 테스트(또는 R1-(b)를 고르면 그 한도를 고정하는 테스트), lipgloss v2로 통일.
- **F5-R1** [Low] [optional] [in-diff] `internal/cli/update_wizard.go:279-285` — 새 분기는 대화형 재구성 경로에서 도달하지 않는다. huh는 Enter/Next에서 `validate`를 부르고(`huh/v2@v2.0.3/field_input.go:353`), 위저드 검증기는 빈 값을 질문 기본값으로 바꿔 저장한다(`wizard.go:478-481`). 그러므로 provider gitlab이면 `result.GitLabInstanceURL`은 이미 `https://gitlab.com`이다. RED 테스트는 위저드가 만들 수 없는 결과를 손으로 만든다. 수리 자체는 무해하고 일관성 면에서 옹호 가능하지만, 진행 기록의 "reproduced YES"는 사용자 경로 재현이 아니다. 요구 수정: 진행 기록에 "저장기 단위 재현, 대화형 경로에서는 이미 기본값 복원"이라고 정정.
- **F5-R2** [Medium] [optional] [outside] `internal/cli/update_wizard.go:55-73` — 재구성 위저드가 저장된 `instance_url`을 미리 채우지 않아, 셀프호스팅 URL을 쓰던 사용자가 Enter만 눌러도 `https://gitlab.com`으로 덮어쓴다(수리 전부터 있던 동작, 작성자가 Gaps에 기록함). 데이터 손실성이라 운영자에게 보일 것. 이 카드 판정에는 영향 없음.
- **N6-R1** [Low] [optional] [in-diff] `internal/profile/sync.go:57-62` — 거부가 함수 전체를 일찍 끝내므로 깨진 user.yaml이 뒤의 language·statusline 동기화까지 막는다(실측: ConversationLang ko 요청에도 language.yaml이 en으로 남음). `internal/cli/launcher.go:682`는 오류를 버리므로 그 경로에서는 조용히 일어난다. 깨진 파일 앞에서 멈추는 것도 옹호 가능하므로 optional. 요구 수정(선택): 이름 동기화만 건너뛰고 나머지 섹션은 진행한 뒤 오류를 모아 반환하거나, 현재 동작을 의도로 주석에 적는다.
- **N7-R1** [Low] [optional] [touched-file] `internal/cli/init_t1966_test.go:3` — 파일 머리 주석 "(items 1, 2, 7)"이 78행의 Item 5 테스트를 빠뜨린다. 카드가 말한 "부정확한 테스트 주석 1개"의 후보로 더 유력하다. 작성자가 고친 Item 7 주석은 지금 정확하다(E4).
- **F4-R1** [Info] [optional] [touched-file] `internal/cli/init.go:96-97` — "the wizard presents the same three pages"는 사실과 다르다(스테퍼 "3 / 5", `RunWithDefaults` 주석은 "four pages"). 테스트 주석이 아닌 prod 주석.
- **F4-R2** [Info] [optional] [outside] docs-site 표(`docs-site/content/en/cli-reference/init.md:53`, `getting-started/cli.md:70`)의 "(default: true)"는 표의 설명 칸이라 이중 표기가 아니다. 조치 불필요.

## Claim / Evidence / Baseline-attribution

### C1. F4: `moai init --help`가 기본값을 한 번만 보인다 — 확인
- 렌더 위치: fang이 0이 아닌 기본값에 `" (" + f.DefValue + ")"`를 붙인다.
  - `grep -n 'DefValue' /Users/goos/go/pkg/mod/charm.land/fang/v2@v2.0.1/*.go` →
    `help.go:405: if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" && f.DefValue != "[]" {` / `help.go:406: help += styles.FlagDefault.Render(" (" + f.DefValue + ")")`
- 후보 빌드: `timeout 300 go build -o <scratch>/moai-cand ./cmd/moai` → `build exit=0` (트리 388967b934에서 빌드, 경로로 실행)
- `NO_COLOR=1 ./moai-cand init --help | grep ...` →
  ```
  29:    --enable-design          Enable design workflow (true)
  30:    --enable-lsp             Enable LSP integration (true)
  31:    --enforce-quality        Enforce quality gates (true)
  ```
- 다른 문자열 플래그의 "(default: ...)"는 DefValue가 ""라 fang이 덧붙이지 않으므로 이중 표기 없음(같은 출력 33·34·44·46·51행).
- 테스트를 고정하는 다른 골든·도움말 스냅샷: `git grep -n 'Enforce quality gates\|Enable design workflow' 388967b934` → 코드 2행 + docs-site 표 + SPEC progress 인용뿐, 도움말 골든 없음.
- RED: 수리 전 init.go를 overlay로 바꿔 `go test -run 'TestInitBoolFlag|TestReconfigureBlank|TestInitCmd_EnableLSP' ./internal/cli/` →
  `init_help_default_t1972_test.go:23: --enable-design defaults to true and the renderer prints that; its usage must not state the default again: "Enable design workflow (default: true)"` (enforce-quality도 같음)
- 잘못된 구현 통과 가능성: 테스트는 bool·기본값 true 플래그의 usage에 "default" 부분문자열이 있는지만 본다. "on by default" 같은 표현도 걸리므로 우회는 어렵다. 범위가 bool-true로 한정된 점은 카드 범위와 맞는다.

### C2. F5: 빈 GitLab URL이면 저장기가 기본값을 쓴다 — 확인(단, 대화형 경로에선 도달 불가)
- RED(overlay로 수리 전 update_wizard.go): `instance_url on disk = "https://gitlab.corp.example", want "https://gitlab.com"` 및 `instance_url on disk = "", want "https://gitlab.com"`
- 도달성: `sed -n 340,360p .../huh/v2@v2.0.3/field_input.go` → `case key.Matches(msg, i.keymap.Next, i.keymap.Submit): value := i.textinput.Value(); i.err = i.validate(value)`; `wizard.go` 검증기 `if v == "" && defVal != "" { v = defVal } ... saveAnswer(qID, v, result, locale)`. provider gitlab ⇒ URL 질문이 보이고(Condition 동일) ⇒ 검증기가 기본값을 저장. `applyWizardConfig`의 유일한 호출자는 `update_wizard.go:122`.
- 저장된 셀프호스팅 URL 덮어쓰기: 재구성 위저드는 github/gitlab 사용자명만 미리 채운다(`update_wizard.go:55-73`), URL 질문은 기본값 `https://gitlab.com`이 입력란에 미리 들어간다(`buildInputField`의 `value = q.Default`). 수리 전·후 같은 동작 → F5-R2(outside).

### C3. F6: wrapped 옵션 줄을 센다 — 80열 영어에서만 성립, 다른 로케일·폭에서 회귀(F6-R1/R2)
- RED(overlay로 수리 전 wizard.go):
  ```
  select_height_t1972_test.go:64: selectHeight = 4, want at least 6 (wrapped option rows are not counted)
  select_height_t1972_test.go:64: selectHeight = 4, want at least 5 (wrapped option rows are not counted)
  select_height_t1972_test.go:88: question "report_format" locale "ko": selectHeight = 6, need at least 8
  select_height_t1972_test.go:88: question "autonomy_tier" locale "ja": selectHeight = 6, need at least 8
  ```
  골든도 수리 전 코드에서 8~10행이 빈 줄로 나옴(`want "┃   safety checks ..." got "      "`).
- 골든 diff(`git show 9a5dc4b672 -- .../init-regroup-second-group.golden`): 빈 줄 3행 → `┃   safety checks`, `┃   Bypass permissions            - Skip all prompts; requires sandbox proof`, `┃   (Docker/gVisor/etc.)`. 이전에 가려졌던 행이 보이게 된 것뿐 — 작성자 주장 확인.
- 로케일별 높이(overlay 프로브 `TestZZAudit_HeightsPerLocale`):
  ```
  report_format  base=9 max=9 old=6 | ja=8 ko=8 zh=8
  autonomy_tier  base=9 max=9 old=6 | ja=8 ko=7 zh=5
  ```
- E7 — 80열, 1쪽에서 언어를 고른 뒤 2쪽 렌더(overlay 프로브, head): ko에서 마지막 옵션 `(Docker/gVisor 등)` 아래 `┃` 빈 줄 2개, ja 1개, zh 4개. 같은 프로브를 수리 전 코드로: ko·ja는 빈 카드 줄 0개지만 "권한 우회"/"권限をバイパス" 이하가 가려짐, zh는 빈 카드 줄 1개.
- E8 — 영어, 폭 60/120(head):
  ```
  === width 60 ===
  ┃   Bypass permissions            - Skip all prompts;
  (이후 "requires sandbox proof (Docker/gVisor/etc.)" 가려짐)
  === width 120 ===
  ┃   Bypass permissions            - Skip all prompts; requires sandbox proof (Docker/gVisor/etc.)
  ┃
  ┃
  ┃
  ┃
  ```
- 요구사항: `.moai/specs/SPEC-INIT-TUX-I18N-001/spec.md:196` — "REQ-ITI-015 (Ubiquitous): The init wizard, the reconfigure wizard and the profile wizard **shall** render ... no empty card row below a select field's last option."
- 선행 이력: `git log -1 --format=%B 5c3f0ebf27` — "revert(wizard): remove the selectHeight lineage ... audit blockers B1 (fixed 76-column row counting) and B2 (blank line between fields) motivated the revert", 그리고 앞선 D1 커밋 "the maximum over all locales, leaving an empty card row under the last option in every shorter locale".
- 테스트가 동어반복인가: 아니다. `rowsLowerBound`는 runewidth 표시폭 올림 나눗셈이라 구현의 lipgloss 줄바꿈과 독립된 하한이고, 수리 전 코드에서 실제로 실패했다. 다만 상한이 없어 E7·E8을 못 잡는다(F6-R4).

### C4. N6: 무효 YAML user.yaml을 건드리지 않고 거부 — 확인, 기존 정상 파일은 계속 통과
- RED(overlay로 수리 전 sync.go): `sync over a duplicate-key user.yaml reported success: user: name: "Carol" name: "Bob"`
- 로더 전제 실측(overlay 프로브): `WARN failed to load user config, using defaults error="parse user.yaml: config: invalid YAML syntax"` / `LoadRaw err=<nil> userName=""` — 작성자 측정(로더가 거부하지 않고 기본값으로 진행) 확인.
- 경계 파일(head):
  ```
  bom:      err=<nil> after="﻿user:\n  name: \"Carol\"\n"
  multidoc: err=<nil> after="user:\n  name: \"Carol\"\n---\nother: 1\n"
  flow:     err=<nil> after="{user: {name: Carol}}\n"
  empty:    err=sync user name: yamlpatch: ...: empty or non-document YAML   (yamlpatch의 기존 거부, 프로브와 무관)
  comment:  err=sync user name: yamlpatch: ...: empty or non-document YAML   (같음)
  toplevseq: err=sync user name: user.yaml is not valid YAML, fix it first: ... cannot unmarshal !!seq   (전에도 yamlpatch "not a mapping"으로 실패하던 형태)
  absent:   err=<nil> after="user:\n    name: Carol\n"
  ```
  → BOM·다중 문서·flow·부재 파일은 이전처럼 통과. 새로 거부되는 정상 파일 없음.
- 범위 확장(N6-R1): `dup+lang: err=sync user name: user.yaml is not valid YAML ... language.yaml="language:\n  conversation_language: en\n"`

### C5. N7: 고친 주석은 정확 — 단 카드가 뜻한 주석인지는 미확정
- 새 Item 7 주석("usage string does not state the default; the help renders the registered value itself")은 C1의 렌더 출력과 일치.
- 같은 파일 3행 머리 주석 "(items 1, 2, 7)"은 78행 Item 5를 빠뜨려 부정확(N7-R1). t1966 커밋 5c3f0ebf27이 이 파일을 만들었고 Item 5는 그 같은 커밋에 있다.

### C6. 지정 명령 실행 결과
- `timeout 300 go test -count=1 ./internal/cli/wizard/` → `ok  github.com/modu-ai/moai-adk/internal/cli/wizard 6.023s`
- `timeout 300 go test -count=1 ./internal/profile/` → `ok  github.com/modu-ai/moai-adk/internal/profile 1.953s`
- `timeout 560 go test -count=1 ./internal/cli/ -run 'TestInitCmd_|TestInitBoolFlag|TestReconfigureBlank|TestValidateInitFlags|TestRunInit_'` → `ok  github.com/modu-ai/moai-adk/internal/cli 153.188s` (정규식 차이는 Gaps 참조)
- `gofmt -l internal/cli internal/profile` → 출력 없음, `exit=0`
- `moai verify audit-plan --project-root <tree>` → `"config_status": "ok"`, `"model": "codex"`, codex gate `required`, claude/glm `off`, `"cross_model_active": true`

### Baseline-attribution
- 모든 측정은 트리 388967b934342b3c9a9e2a6939bc36331958e588에서 이번 실행에 수행. 수리 전 대조는 `git show 9b632c1eb3:<path>`로 뽑은 base 파일(wizard.go, sync.go, init.go, update_wizard.go)을 `go test -overlay`로 바꿔 끼워 측정했고, 트리 파일은 건드리지 않았다.
- 도움말 렌더는 같은 트리에서 빌드한 후보 바이너리를 경로로 실행(설치본 미사용). MCP 서버 빌드는 646955327e(codex_audit의 `build_commit`).
- 렌더 프로브는 `ptycaptest.NewFormDriver`(80x40 WindowSizeMsg) 위에서 폭만 바꿔 측정.

## Gaps
- 지정 명령의 `-run` 정규식에서 `TestReconfigureBlankGitLab`을 그대로 쓴 호출은 워크트리 가드가 거부했다("names git in a form too complex to verify"). 대신 `TestReconfigureBlank`로 실행했다(이 접두어를 가진 테스트는 F5 테스트 하나). 가드 거부 1건은 lessons-inbox의 tool_failure 행으로 남았을 수 있다.
- `go list -m -f '{{.Dir}}'`를 `$(...)`로 감싼 호출도 같은 가드가 거부해 분리 실행으로 대체.
- 실행하지 않음: 전체 `internal/cli` 스위트, `-race`, golangci-lint, 다른 OS, 커버리지 측정. Craft 점수의 커버리지 항목은 미측정.
- 실제 TTY에서의 렌더는 보지 않았다. E7·E8은 FormDriver(huh `View()`) 기준이다.
- t1966 원 판정문(F4~N7의 출처)은 디스크에 없어, 카드 본문과 t1966 커밋 메시지로만 대조했다. N7이 가리킨 주석은 확정 불가.
- 운영자 지시에 따라 audit_multi는 호출하지 않았다(audit-plan은 `cross_model_active: true`를 보고). codex_audit는 1회만 호출했다.
- receipt 가드가 이 판정을 거부하더라도 우회하지 않는다.

## Residual-risk
- F5 새 분기는 지금 도달 불가지만, 향후 다른 호출자가 "빈 URL = 유지"의 뜻으로 결과를 넘기면 저장된 셀프호스팅 URL이 gitlab.com으로 덮인다(테스트 1번 사례가 그 덮어쓰기를 고정).
- N6 프로브는 `map[string]any` 파싱이라, yaml.v3가 받아들이지만 yamlpatch 줄 스플라이스가 잘못 다루는 다른 형태(예: 앵커/별칭으로 들어온 name)가 남아 있을 수 있다. 검사하지 않음.
- 스크래치의 프로브 출력 파일(render_head.txt, render_base.txt)은 기계 로컬이다. 판정을 가른 줄은 위 Evidence에 옮겨 적었고, 나머지는 보존되지 않는다.
- Codex는 같은 범위를 통과로 봤다. 이 감사의 FAIL은 렌더 실측(E7·E8)과 REQ-ITI-015 문구에 기대며, 리더가 REQ-ITI-015의 적용 범위를 init 2쪽에 미치지 않는다고 판단하면 F6-R1의 blocking 분류가 달라질 수 있다(AC-ITI-016이 명시한 표면은 init 1쪽·프로필 위저드·재구성 1그룹).

## 반복 이력
- 1회차(이 문서). 이전 sync 감사 없음.

## Operational Notes (unverified)
- F6을 고치는 가장 작은 길은 R1-(c)(분리·되돌림)로 보인다. t1966이 같은 이유로 같은 결정을 했고, 나머지 4건(F4·F5·N6·N7)은 F6 없이 그대로 착지할 수 있다. R1-(a)는 t1966의 liveSelect 계보가 B2(필드 사이 빈 줄)로 다시 막혔던 이력이 있어 비용이 크다.
- N7-R1(머리 주석)과 F4-R1(init.go "three pages")은 한 줄씩이라 같은 수리 커밋에 넣어도 범위가 커지지 않는다.
- F5-R2(저장 URL 미리 채우기)는 별도 카드 후보.
