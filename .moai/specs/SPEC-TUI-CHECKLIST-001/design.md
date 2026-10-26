# SPEC-TUI-CHECKLIST-001 — 설계

이 문서는 run 단계가 따를 형태를 정한다. 함수·파일 이름은 방향을 보여 주는 예시이며, 같은 계약을 지키면 run 단계가 기존 관례에 맞춰 바꿀 수 있다. 이름이 바뀌면 acceptance.md 의 선택자도 함께 고친다. 바꿀 수 없는 것은 계약이다 — 이벤트 순서, 캡처 규칙, 대체 규칙, 키 이름 규칙. 인용 행번호는 워크트리 HEAD `cb30e05b3e` 기준이다.

## §1 화면 모양 (리치 모드, 깨끗한 init)

```
◆ MoAI-ADK init
  darwin/arm64 · v3.3.0 · ~/work/demo
◆ 템플릿
  ✓ 프로젝트 구조 감지
  ✓ 프로젝트 검증
  ✓ 템플릿 배포
◆ 설정
  ✓ 프로필과 언어
  ✓ 선택 기능 동의
  ✓ 하네스와 배포 모드
  ✓ 자율 등급
  ✓ 전역 설정
◆ 훅
  ✓ pre-push 훅
  ✓ pre-commit 훅
◆ MCP
  ✓ MCP 서버 항목
◆ 검증
  ✓ 매니페스트 기록

  완료 · 파일 2개 생성 · 디렉터리 13개 · 경고 0 · 1.4s
  배포 모드: plugin
  다음 단계: cd demo → moai cc → /moai plan "…"
```

- 23줄이다. 상한은 24줄(REQ-TCL-010).
- 실행 중인 항목은 `  ⠋ 템플릿 배포` 처럼 스피너 프레임으로 한 줄만 그리고, 그 줄만 제자리에서 다시 그린다.
- 실패한 항목은 `  ✗ …` 아래에 들여쓴 상세 줄을 모두 펼친다. 경고가 있는 항목은 `  ! …` 아래에 경고 문장을 펼친다.
- 요약의 `→` 는 상태 글리프가 아니라 본문 문자다. run 단계가 허용 목록 검사(AC-TCL-023)에 걸리는지 확인하고, 걸리면 `·` 로 바꾼다.

### §1.1 카드 문구와 다른 글리프를 고른 이유

| 카드 문구 | 채택 | 근거 |
|---|---|---|
| 그룹 표식 `➤` | `◆` | `➤` 는 U+27A4 로, 카탈로그 이모지 검사가 금지하는 U+2700–27BF 범위(`internal/tui/i18n_test.go:106`)에 들고 상태 글리프 허용 목록(`internal/tui/glyphs.go:9-10`)에도 없다. `◆` 는 허용 목록에 있고 update 정체성 띠(`internal/cli/update_tux.go:78`)가 이미 쓴다. |
| 건너뜀 `–` | `○` (`GlyphSkip`) | U+2013 은 허용 목록에 없다. 이 렌더러는 대기 항목을 그리지 않으므로 `○` 는 "건너뜀" 하나만 뜻한다. |
| 실행 중 스피너 또는 `/` | `Spinner` 프레임(평소 `⠋`, 모션 감소 시 `●`) | 기존 `Spinner` 를 재사용한다(`internal/tui/status.go:51-72`). |

## §2 렌더러 API — `internal/tui/checklist.go` (새 파일 1개)

```go
type ChecklistOptions struct {
    Out   io.Writer        // 그릴 대상. 채널은 호출자가 고른다(§5)
    Theme *Theme           // nil 이면 LightTheme (기존 원칙과 같음)
    Lang  string           // 카탈로그 언어. 빈 값·미지원 값은 en
    Clock func() time.Time // nil 이면 time.Now. 골든은 고정 시계를 넣는다
}

func NewChecklist(opts ChecklistOptions) *Checklist
func (c *Checklist) Header(titleKey string, env EnvSummary)
func (c *Checklist) Item(groupKey, itemKey string) *ChecklistItem // 그룹이 바뀌면 그룹 제목을 먼저 그린다
func (i *ChecklistItem) Update()            // 스피너 한 칸 전진, 그 줄만 다시 그림
func (i *ChecklistItem) Detail(line string) // 접힌 상세 한 줄 기록
func (i *ChecklistItem) Warn(msg string)    // 경고 기록(종료 시 펼침)
func (i *ChecklistItem) Done()
func (i *ChecklistItem) Fail(msg string)
func (i *ChecklistItem) Skip(reasonKey string)
func (c *Checklist) Summary(s ChecklistSummary)
```

- 렌더러는 언제나 리치 모드로 그린다. 켤지 말지는 CLI 쪽 세션이 정한다(§3.1). 대체 출력은 렌더러가 아니라 지금의 코드 경로가 맡으므로 렌더러에 평문 모드를 두지 않는다(YAGNI).
- 고루틴과 티커가 없다. 스피너는 이벤트가 올 때만 한 칸 움직인다. `Spinner` 의 무상태 계약(`internal/tui/status.go:49-50`)을 그대로 따른다.
- 항목은 한 번에 하나만 열린다. 새 `Item` 이 오면 열린 항목을 `Done` 으로 닫는다(암묵 종료). 그래서 호출 지점마다 닫는 줄을 따로 넣지 않아도 되고, init.go 변경이 한 줄 삽입으로 끝난다(plan.md §B).
- 종료 전환(`Done`/`Fail`/`Skip`)을 두 번 부르면 `ProgressLine` 처럼 panic 한다(`internal/tui/progress_line.go:173-177` 선례).
- 줄 다시 그리기는 `progressClearPrefix`(`internal/tui/progress_line.go:41`)를 재사용한다. 완료된 줄 뒤에는 다시 그리기 시퀀스를 내지 않는다.
- 색은 `Theme` 토큰만 쓰고 `downsample` 을 거친다. `NO_COLOR` 에서는 `MonochromeTheme`(`internal/tui/detect.go:59-61`)이므로 SGR 이 나오지 않는다. 줄 지우기 시퀀스는 색이 아니므로 남는다.
- 폭 계산은 `internal/tui/internal` 의 `StringWidth` 를 쓴다(CJK 2칸, `internal/tui/internal/runeguard.go:39-41`).

## §3 이벤트 모델과 CLI 쪽 세션 — `internal/cli/checklist_session.go` (새 파일)

### §3.1 활성 판정 (REQ-TCL-011, 012)

다음이 모두 참일 때만 체크리스트를 켠다. 하나라도 거짓이면 세션은 빈 세션(모든 메서드가 아무 일도 하지 않음)이 되고, 명령은 지금 출력 그대로 동작한다.

1. 프로세스 stdout 과 stderr 가 모두 터미널이다(`os.Stdout`·`os.Stderr` fd 를 isatty 로 확인).
2. `cmd.OutOrStdout()`·`cmd.ErrOrStderr()` 가 프로세스 스트림 그대로다(테스트가 버퍼를 끼우면 거짓).
3. `CI` 환경 변수가 비어 있다(`internal/cli/init_participation_wizard.go:34` 와 같은 판정).
4. `TERM` 이 `dumb` 가 아니다.
5. `--verbose` 가 꺼져 있다.

판정 함수는 패키지 변수 시접으로 두어 테스트가 강제로 켤 수 있게 한다(`internal/cli/init_executor_seam.go:26-36` 과 같은 방식: 바꾸는 테스트는 `t.Parallel` 금지, `t.Cleanup` 으로 복원).

### §3.2 캡처 규칙 (REQ-TCL-014)

세션이 켜지면 `cmd.SetOut`·`cmd.SetErr` 로 세션 작성기를 끼운다.

- **항목이 열려 있는 동안** 쓰인 바이트는 그 항목의 상세로 기록된다(줄 단위). 화면에는 나오지 않는다.
- **항목이 열려 있지 않을 때** 쓰인 바이트는 즉시, 바꾸지 않고 실제 작성기로 넘긴다. 위저드 전 경고(`init.go:427`), 사용자 자산 보장 출력(`init.go:784`), 세션 워크트리 안내(`init.go:405-423`), update 종료 블록(`update.go:176`)이 여기에 해당한다.
- 기존 `ProgressLine` 호출은 그대로 둔다. 세션 작성기는 `*os.File` 이 아니므로 `ProgressLine` 은 비TTY 형태(이스케이프 없는 줄, `internal/tui/progress_line.go:218-224`)로 쓰고, 그 줄이 상세로 접힌다.

### §3.3 경고 신호 (REQ-TCL-015)

- `warnCollector.Warn`·`Collect`(`internal/cli/init_warnings.go:41-56`)는 세션이 켜져 있으면 세션에 경고를 알린다. 열린 항목이 있으면 그 항목의 경고가 되고, 없으면 경고 한 줄을 즉시 넘긴다.
- `emitSeverityLine`(`internal/cli/severity_line.go:67-69`)은 작성기가 경고 수신 인터페이스를 구현하고 심각도가 `sevWarn`·`sevErr` 이면 같은 방식으로 알린다. 아니면 지금처럼 한 줄을 쓴다. 이 경로가 필요한 이유는 init 의 `emitSeverityLine(…, sevWarn, …)` 두 곳(`init.go:901`, `:907`)이 수집기를 거치지 않기 때문이다.
- 세션이 켜져 있으면 `emitSummary` 패널(`internal/cli/init_warnings.go:69-84`)을 내지 않는다. 모든 경고는 항목 아래 펼침이나 즉시 넘김으로 정확히 한 번 나온다.

### §3.4 종료와 실패 (REQ-TCL-016)

- 세션 종료는 `runInit`·`runUpdate` 에서 가장 먼저 등록하는 `defer` 로 둔다. 그래서 가장 나중에 실행되고, 뒤에 등록된 `emitSummary`·세션 워크트리 정리·update 종료 블록보다 늦게 돈다.
- 요약에 도달하지 못하고 끝나면(오류 반환, 위저드 취소, 조기 반환) 열린 항목을 `Fail` 로 닫고, 아직 화면에 나가지 않은 캡처 바이트를 모두 그대로 쏟아낸 뒤 작성기를 되돌린다.

### §3.5 대화형 입력 경계 (REQ-TCL-018)

- 위저드(`internal/cli/wizard/wizard.go:129` `form.Run()`)와 병합 확인(`internal/cli/update_template_sync.go:1373-1374`)은 `cmd` 작성기를 쓰지 않으므로 캡처되지 않는다.
- 렌더러는 대화형 입력이 도는 동안 끝나지 않은 줄을 갖지 않는다. init 의 첫 항목은 위저드가 끝난 뒤 보고기 이벤트로 열리고, update 의 확인은 `analyze` 항목을 닫은 뒤 열린다.

### §3.6 보고기 연결

init 은 `executor.SetReporter(newSpinnerReporter(p))`(`init.go:728`)로 보고기를 받는다. 이 줄은 t1966 의 변경 범위(`init.go:730-749` 삭제) 바로 옆이라 건드리지 않는다. 대신 `newSpinnerReporter`(`internal/cli/reporter.go:42-44`)가 받은 printer 에서 세션을 꺼낼 수 있으면(수집기가 세션 접근자를 노출) 체크리스트 보고기를 돌려준다. 보고기의 `StepStart` 이름을 §4 표의 `templates` 항목으로 옮긴다.

## §4 그룹·항목 계획표 (SSOT) — `internal/cli/checklist_plan.go`

한 표에 모은다. 그룹당 항목 8개 이하(REQ-TCL-010). 열의 키는 카탈로그 키 접미사다.

| 명령 | 그룹 키 | 항목 키 | 대응하는 지금 코드 |
|---|---|---|---|
| init | `templates` | `detect`, `validate`, `deploy` | 보고기 `StepStart` 이름 Detection·Validation·Initialization (`internal/cli/testdata/tuxiu/postm4/init.notty.stderr.golden:2-9` 이 보여 주는 순서) |
| init | `config` | `profile`, `optins`, `harness`, `autonomy`, `globals` | `init.go:928-1016` 꼬리 단계 묶음 |
| init | `hooks` | `prepush`, `precommit` | `init.go:1022-1029` |
| init | `mcp` | `entry` | `init.go:1045-1048` |
| init | `verify` | `manifest` | `init.go:1073-1075` |
| update | `templates` | `version`, `load`, `manifest`, `analyze` | `update_template_sync.go:242-343` |
| update | `backup` | `config`, `namespace` | 단계 루프의 `Backup` 가로채기(`update_template_sync.go:734-`) |
| update | `deploy` | `validate`, `clean`, `deploy`, `restore`, `reconcile` | 단계 표(`update_template_sync.go:450-`)와 조정(`:1028`) |
| update | `verify` | `gitignore`, `retrack`, `retention` | `update_template_sync.go:948-1016` |

init 의 사용자 자산 보장(`init.go:780-790`)은 항목으로 만들지 않는다. 그 바로 앞 줄(`init.go:780`)을 t1966 이 바꾸기 때문이다. 이번에는 출력이 넘김 규칙으로 지금처럼 보이고, 항목화는 t1966 착지 뒤 후속 카드로 넘긴다(plan.md §E).

update 의 clean-install 경로(`update_clean_install.go`), 바이너리 갱신, `--check`·`--shell-env`·`--config` 경로는 항목을 열지 않는다. 그 경로의 출력은 넘김 규칙으로 지금 그대로 보인다(spec.md §5).

## §5 출력 채널 (REQ-TCL-017)

- init: 체크리스트를 stderr 에 그린다. 진행·경고·완료 카드가 stderr 로 가는 지금 규칙(`init.go:383-387`, `:1085-1087`)과 같다.
- update: 체크리스트를 stdout 에 그린다. update 의 진행 줄은 지금 stdout 으로 간다(`update.go:156`, `internal/cli/testdata/tuxiu/postm4/update.notty.stdout.golden:15-46`).

## §6 i18n 키 계획 (REQ-TCL-009)

키 이름 규칙 `<surface>.<section>.<role>`(`internal/tui/i18n.go:78-79`)을 따른다. 모든 키를 en·ko·ja·zh 네 카탈로그에 같은 이름으로 넣는다.

| 접두사 | 키 |
|---|---|
| `checklist.header.` | `init`, `update` |
| `checklist.group.` | `templates`, `config`, `hooks`, `mcp`, `verify`, `backup`, `deploy` |
| `checklist.item.<group>.` | §4 표의 모든 항목(init·update 가 같은 키를 쓰면 하나로 둔다) |
| `checklist.status.` | `skipped`, `failed` |
| `checklist.summary.` | `done`, `failed`, `created_files`, `created_dirs`, `changed`, `preserved`, `warnings`, `elapsed`, `deploy_mode`, `next_steps`, `backup`, `recover` |

- 언어 해석: update 는 `.moai/config/sections/language.yaml` 의 `conversation_language`, init 은 위저드 답 또는 프로필 로케일을 쓰고 둘 다 없으면 `en`. 미지원 값은 `Translate` 가 `en` 으로 대체한다(`internal/tui/i18n.go:83-99`).
- 값에 서식 지시자(`%d`, `%s`)가 있으면 네 카탈로그의 지시자 개수가 같아야 한다(AC-TCL-008 이 비교).
- ja·zh 문구는 이 카드에서 초안을 넣는다. 원어민 검수는 범위 밖이다(spec.md §5, Gap G-5).

## §7 골든 전략

| 층 | 위치 | 고정 조건 | 갱신 |
|---|---|---|---|
| 렌더러 단위 | `internal/tui/testdata/checklist-*.golden` | 고정 시계, `LightTheme`·`MonochromeTheme`, 고정 환경 요약 | `UPDATE_GOLDEN=1` 과 `MANIFEST.txt`(`internal/tui/golden/index_test.go:43-73`) |
| CLI 대체 경로 | `internal/cli/testdata/checklist/fallback-*.golden` | 버퍼 작성기(비TTY), 고정 프로젝트 디렉터리, 변동 값(백업 시각·버전) 정규화 | M0 에서 **코드 변경 전에** 따로 커밋(§8) |
| CLI 리치 경로 | `internal/cli/testdata/checklist/rich-*.golden` | 활성 판정 시접을 강제로 켜고 버퍼 작성기 사용 | M4·M5 |

골든 이름: `checklist-<surface>-<case>.golden`, case ∈ {clean, fail, warn, nocolor, reduced, ko}. 기존 `TestGoldenCount` 하한(`internal/tui/golden/index_test.go:27`, 100개)은 늘어나는 방향이라 영향이 없다.

## §8 기준선을 먼저 커밋하는 이유

`--verbose`·비TTY 출력이 지금과 바이트 단위로 같다는 주장(REQ-TCL-012, 013)은 변경 전 출력을 기준으로 한다. 기준선과 구현이 한 커밋에 들어가면 순서를 git 이 증언할 수 없다(`verification-claim-integrity.md` §2.3). 그래서 M0 은 대체 경로 골든과 그 골든을 읽는 테스트만 담은 커밋 하나로 끝내고, 그 뒤에 코드 변경을 시작한다. M0 커밋은 그 시점 코드에서 이미 녹색이어야 한다(기준선을 바로 그 코드로 찍었으므로).

기존 `internal/cli/testdata/tuxiu/` 골든(`internal/cli/tuxiu_characterization_test.go:14-17`)은 과거에 손으로 찍어 둔 두 묶음을 서로 비교할 뿐 실행 출력을 읽지 않는다. 그래서 이번 변경을 지켜 주지 못하며, 대체 경로 기준선이 따로 필요하다.
