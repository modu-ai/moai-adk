# SPEC-TUI-CHECKLIST-001 — 설계

이 문서는 run 단계가 따를 형태를 정한다. 함수·파일 이름은 방향을 보여 주는 예시이며, 같은 계약을 지키면 run 단계가 기존 관례에 맞춰 바꿀 수 있다. 이름이 바뀌면 acceptance.md 의 선택자도 함께 고친다. 바꿀 수 없는 것은 계약이다 — 이벤트 순서, 캡처 규칙, 보존 줄 목록, 대체 규칙, 키 이름 규칙. 인용 행번호는 `cb30e05b3e` 기준이며, 델타 1회차에서 HEAD `d241917d17`, 델타 2회차에서 HEAD `b422a13d53`(셋 다 코드는 같음 — `b422a13d53` 까지의 커밋은 SPEC 디렉터리만 바꿨다)에서 다시 읽었다.

## §1 화면 모양 (리치 모드, 깨끗한 init)

```
◆ MoAI-ADK init
  darwin/arm64 · v3.3.0 · ~/work/demo
◆ 템플릿
  ✓ 프로젝트 구조 감지
  ✓ 프로젝트 검증
  ✓ 템플릿 배포
◆ 설정
  ✓ 자율 등급
  ✓ 프로필과 언어
  ✓ 선택 기능 동의
  ✓ 하네스와 배포 모드
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

- 23줄이다. 상한은 24줄(REQ-TCL-010, 깨끗한 실행의 정의는 spec.md §3.3).
- `설정` 그룹의 항목 순서는 코드 순서를 따른다. R3 는 삽입만 하므로(plan.md §B) 화면 순서가 곧 코드 순서다: 자율 등급 `init.go:928` → 프로필 `:958` → 선택 기능 동의(Jev `:970`, 참여 `:980`) → 하네스와 배포 모드(하네스 `:991`, 배포 모드 `:1001`, evolution `:1009`) → 전역 설정 `:1014`.
- 실행 중인 항목은 `  ⠋ 템플릿 배포` 처럼 스피너 프레임으로 한 줄만 그리고, 그 줄만 제자리에서 다시 그린다.
- 실패한 항목은 `  ✗ …` 아래에 들여쓴 상세 줄을 모두 펼친다. 경고가 있는 항목은 `  ! …` 아래에 경고 문장을 펼친다. 보존 줄(§3.3.1)은 항목이 성공해도 펼친다.
- 항목 밖에서 기록된 수집기 경고(예: 위저드 전 git 미설치 경고 `init.go:427`)는 요약 아래 `경고:` 목록으로 다시 나열한다(결정 D-6).
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
func (i *ChecklistItem) Warn(msg string)    // 경고 기록(종료 시 펼침, 줄 끝 `!`)
func (i *ChecklistItem) Keep(line string)   // 보존 안내 기록(종료 시 펼침, 글리프 불변)
func (i *ChecklistItem) Done()
func (i *ChecklistItem) Fail(msg string)
func (i *ChecklistItem) Skip(reasonKey string)
func (c *Checklist) Summary(s ChecklistSummary) // 항목 밖 경고 목록 포함
```

- 렌더러는 언제나 리치 모드로 그린다. 켤지 말지는 CLI 쪽 세션이 정한다(§3.1). 대체 출력은 렌더러가 아니라 지금의 코드 경로가 맡으므로 렌더러에 평문 모드를 두지 않는다(YAGNI).
- 고루틴과 티커가 없다. 스피너는 이벤트가 올 때만 한 칸 움직인다. `Spinner` 의 무상태 계약(`internal/tui/status.go:49-50`)을 그대로 따른다.
- 항목은 한 번에 하나만 열린다. 새 `Item` 이 오면 열린 항목을 `Done` 으로 닫는다(암묵 종료). 그래서 호출 지점마다 닫는 줄을 따로 넣지 않아도 되고, init.go 변경이 한 줄 삽입으로 끝난다(plan.md §B).
- 종료 전환(`Done`/`Fail`/`Skip`)을 두 번 부르면 `ProgressLine` 처럼 panic 한다(`internal/tui/progress_line.go:173-177` 선례).
- 줄 다시 그리기는 `progressClearPrefix`(`internal/tui/progress_line.go:41`)를 재사용한다. 완료된 줄 뒤에는 다시 그리기 시퀀스를 내지 않는다.
- 색은 `Theme` 토큰만 쓰고 `downsample` 을 거친다. `NO_COLOR` 에서는 `MonochromeTheme`(`internal/tui/detect.go:59-61`)이므로 SGR 이 나오지 않는다. 줄 지우기 시퀀스는 색이 아니므로 남는다.
- 폭 계산은 `internal/tui/internal` 의 `StringWidth` 를 쓴다(CJK 2칸, `internal/tui/internal/runeguard.go:39-41`).

## §3 이벤트 모델과 CLI 쪽 세션 — `internal/cli/checklist_session.go` (새 파일)

### §3.1 활성 판정과 입력 수집 (REQ-TCL-011, REQ-TCL-012)

다섯 입력은 한 함수 `checklistEnv(cmd, verbose)` 가 모은다. init 과 update 는 이 함수만 부르고, 입력을 따로 읽지 않는다. 다음이 모두 참일 때만 체크리스트를 켠다. 하나라도 거짓이면 세션은 빈 세션(모든 메서드가 아무 일도 하지 않음)이 되고, 명령은 기준선 출력 그대로 동작한다.

1. 프로세스 stdout 과 stderr 가 각각 터미널이다 — 두 fd 를 따로 isatty 로 확인한다. stderr 판정에 stdout 결과를 쓰지 않는다.
2. `cmd.OutOrStdout()`·`cmd.ErrOrStderr()` 가 프로세스 스트림 그대로다(테스트가 버퍼를 끼우면 거짓).
3. `CI` 환경 변수가 비어 있다(`internal/cli/init_participation_wizard.go:34` 와 같은 판정).
4. `TERM` 이 `dumb` 가 아니다.
5. `--verbose` 가 꺼져 있다.

시접(패키지 변수, `internal/cli/init_executor_seam.go:26-36` 과 같은 방식 — 바꾸는 테스트는 `t.Parallel` 금지, `t.Cleanup` 으로 복원):

- TTY 탐지 시접: 작성기마다 터미널 여부를 돌려준다. 명령 수준 테스트는 "stdout 참·stderr 거짓"처럼 두 값을 따로 준다.
- 프로세스 스트림 시접: 조건 2 의 비교 대상(프로세스 stdout·stderr)을 돌려준다. 명령 수준 테스트는 자기가 끼운 버퍼를 프로세스 스트림으로 알려 조건 2 를 참으로 만들고, 나머지 한 조건만 거짓으로 둔다. 이 시접이 없으면 버퍼를 쓰는 모든 명령 수준 테스트에서 조건 2 가 늘 거짓이라 다른 조건의 검사가 공허해진다(AC-TCL-012).
- 강제 켜기 시접: 판정 결과를 통째로 참으로 만든다. `--verbose` 도 이 시접 앞에서는 체크리스트를 끄지 못한다. 그래서 이 시접은 리치 골든(AC-TCL-011)과 연결 테스트(AC-TCL-015, AC-TCL-016, AC-TCL-017, AC-TCL-018, AC-TCL-019, AC-TCL-020, AC-TCL-024)만 쓰고, 조건 하나를 거짓으로 두고 판정을 보는 테스트(AC-TCL-012 명령 수준, AC-TCL-014)는 쓰지 않는다. 그 테스트들은 TTY 탐지 시접과 프로세스 스트림 시접으로 나머지 조건을 참으로 두고, 검사하는 조건 하나만 실제 값으로 둔다.

### §3.2 캡처 규칙 (REQ-TCL-014)

세션이 켜지면 `cmd.SetOut`·`cmd.SetErr` 로 세션 작성기를 끼운다.

- **항목이 열려 있는 동안** 쓰인 바이트는 그 항목의 상세로 기록된다(줄 단위). 화면에는 나오지 않는다. 단, §3.3.1 보존 줄은 종료 때 펼친다.
- **항목이 열려 있지 않을 때** 쓰인 바이트는 즉시, 바꾸지 않고 실제 작성기로 넘긴다. 위저드 전 경고(`init.go:427`), 사용자 자산 보장 출력(`init.go:784`), 세션 워크트리 안내(`init.go:405-423`), update 정체성 띠(`update.go:212-214`), update 종료 블록(`update.go:176`)이 여기에 해당한다.
- 기존 `ProgressLine` 호출은 그대로 둔다. 세션 작성기는 `*os.File` 이 아니므로 `ProgressLine` 은 비TTY 형태(이스케이프 없는 줄, `internal/tui/progress_line.go:218-224`)로 쓰고, 그 줄이 상세로 접힌다.

**항목 경계(결정 D-10).** 열린 항목은 네 경우에만 닫힌다 — 다음 `Item` 이 열릴 때(암묵 `Done`), `Summary` 가 그려질 때(암묵 `Done`), 세션이 끝날 때(§3.4), 단계 보고기가 실패를 알릴 때(`Fail`, §3.6). 그 밖에는 열린 채로 다음 바이트를 상세로 받는다.

- init: 첫 항목은 `executor.Execute`(`init.go:829`) 안의 첫 보고기 이벤트(`internal/core/project/phase.go:66` `StepStart("Detection", …)`)로 열린다. 요약은 R2(`init.go:1077-1087`)에서 그린다. 그래서 `init.go:829` 앞의 출력(위저드 전 경고 `:411`·`:427`, 취소 `:575`, 자산 보장 `:784`, 남은 스냅숏 판정 `:812`·`:815`)과 요약 뒤의 출력은 넘김 규칙으로 바로 보인다.
- update: 항목은 템플릿 동기화 안에서만 열린다(`runTemplateSyncWithProgress` `internal/cli/update_template_sync.go:1244` → 본문 `runTemplateSyncWithReporter`). 요약은 본문이 `return nil`(`:1218`) 하기 직전에 그린다. 단계 루프 뒤 꼬리(`:1081-1216` — 진행 막대, 결과 알약 자리, 훅 설치 `:1177`·`:1182`, 배포 모드 기록 `:1211`, 스탬프 검사 `:1216`)는 마지막으로 열린 항목 안에서 돈다. 진행 막대(`:1081`)와 결과 알약(`:1144` `renderUpdateOutcome`)은 세션이 켜져 있으면 그리지 않고 그 데이터를 요약으로 넘긴다(결정 D-6). 템플릿 동기화 밖의 update 출력(`internal/cli/update.go:604` 앞의 미리보기·마이그레이션 줄, 뒤의 무결성 검사 `:631`·참여 질문 `:647`·종료 블록 `:176`, 최종 상태 검증 `update_template_sync.go:1323`)은 열린 항목이 없으므로 넘김 규칙으로 보인다.

### §3.3 경고 신호와 경고 수 (REQ-TCL-015, REQ-TCL-019)

- `warnCollector.Warn`·`Collect`(`internal/cli/init_warnings.go:41-56`)는 세션이 켜져 있으면 세션에 경고를 알린다.
  - 열린 항목이 있으면 그 항목의 경고가 된다(`Warn` 이 프린터로 내는 즉시 줄은 상세로 접히고, 경고 문장은 항목 아래 한 번 펼친다).
  - 열린 항목이 없으면 기존 동작을 그대로 둔다(`Warn` 은 즉시 줄을 내고, `Collect` 는 아무것도 내지 않는다). 그리고 요약 블록이 그 경고를 한 번 나열한다 — 기존 패널이 하던 일을 요약이 대신한다.
- `emitSeverityLine`(`internal/cli/severity_line.go:67-69`)은 작성기가 경고 수신 인터페이스를 구현하고 심각도가 `sevWarn`·`sevErr` 이면 열린 항목의 경고로 알린다. 아니면 지금처럼 한 줄을 쓴다. 이 경로가 필요한 이유는 init 의 `emitSeverityLine(…, sevWarn, …)` 두 곳(`init.go:901`, `:907`)이 수집기를 거치지 않기 때문이다.
- 세션이 켜져 있으면 `emitSummary` 패널(`internal/cli/init_warnings.go:69-84`)을 내지 않는다.
- **경고 수(결정 D-9).** 요약의 경고 수는 `warnCollector.Count()`(`internal/cli/init_warnings.go:59-63`, `len(w.warnings)`)다. 기존 완료 카드가 같은 수를 보고한다(`init.go:1087` `p.Count()`). 심각도 줄과 보존 줄은 항목 아래 보이지만 이 수에 더하지 않는다 — 기존 화면도 그 줄들을 세지 않았다. update 요약의 경고 수는 기존 update 결과 화면이 보고하는 수를 그대로 옮긴다.

#### §3.3.1 보존 줄 목록 (REQ-TCL-022)

경고 신호를 거치지 않고 작성기에 직접 쓰는 줄 가운데, 항목이 성공해도 접히면 안 되는 줄이다. run 단계는 각 지점이 세션에 "보존 줄"로 알리게 한다(방법은 run 단계가 고른다 — 예: 세션 작성기의 보존 메서드, 또는 호출 지점의 작은 도우미). 문자열 모양으로 "경고처럼 보이는 줄"을 고르지 않는다.

**목록의 완전성은 기억이 아니라 고정 검색으로 정한다(델타 2회차, 2회차 감사 N1).** 아래 한 번의 호출이 `internal/cli` 와 그 하위 패키지 전체(테스트 파일 제외)에서 직접 쓰기 경고·오류·안내의 모양을 찾는다. HEAD `b422a13d53` 에서 360줄, 종료 0 을 냈다(acceptance.md 원장 L20, 출력 원문은 research.md §9).

```
/usr/bin/grep -rni --exclude='*_test.go' -e "SymWarning()" -e "SymError()" -e "CheckLine(\"warn\"" -e "CheckLine(\"err\"" -e "warnLine(" -e "Fprintf(errOut" -e "Fprintln(errOut" -e "Fprint(errOut" -e "Fprintf(warn," -e "OrStderr()," -e "sevNote," -e "warning:" internal/cli
```

검색의 모든 일치는 아래 둘 가운데 정확히 한 곳에 분류된다 — 보존 행(P1~P23), 또는 제외 행(X1~X10). 분류는 "파일 + 그 줄의 메시지 조각"으로 하고 행번호로 하지 않는다(코드 수정으로 행이 밀려도 분류가 살아 있게). init·update 궤도 파일(아래 정의)의 일치는 한 줄씩 분류하고, 궤도 밖 파일은 파일 단위로 X9 에 둔다.

- **궤도 파일**: `internal/cli/init*.go`, `internal/cli/update*.go`, `internal/cli/update/**`, 이 카드가 새로 만드는 `internal/cli/checklist*.go`, 그리고 `hook_install.go`, `hook_install_precommit.go`, `worktree_advisory.go`, `feedback_participation.go`, `migrate_project_assets.go`, `migrate_local_instructions.go`, `session_worktree.go`, `mirror_notice.go`, `user_asset_phase.go`, `uikit/styles.go`. 궤도 밖 파일이 init·update 에서 불리지 않는다는 판단은 궤도 파일에서 작성기를 넘기는 호출을 한 단계 따라가 확인했다(research.md §9 둘째 블록). 그 확인으로 `feedback_participation.go` 가 궤도로 들어왔다(update `internal/cli/update.go:649`·`:714`).
- **검색 모양 밖의 줄**: 위 12개 모양에 걸리지 않는 직접 쓰기(예: `p.Info` 안내)는 검색이 찾지 못한다. 그런 줄 가운데 보존해야 하는 것은 열거 행(P3, P4, P22)으로 따로 둔다. 모양 밖의 새 경고 줄은 남은 위험이다(plan.md §H R-3).

보존 행. "위치" 는 그 줄이 쓰일 때 열린 항목이다(design §3.2 항목 경계). "항목 밖" 행은 넘김 규칙으로 보이며, 세션은 그 줄이 정확히 한 번 보이는지만 지킨다.

| # | 호출 지점 | 실제로 쓰는 줄 (분류 조각) | 작성기 | 종류 | 위치 |
|---|---|---|---|---|---|
| P1 | `init.go:885`, `:890` | `internal/cli/update/backup/file_snapshot.go:95`, `:99` → `warnLine`(`internal/cli/update/backup/settings_snapshot.go:115` `fmt.Fprintf(warn, "%s %v\n", prefix, err)`) | stderr | 경고 | `templates.deploy` |
| P2 | `init.go:944`, `:945` | `file_snapshot.go:139` → `warnLine` | stderr | 경고 | 열린 R3 항목(`config.autonomy`) |
| P3 | `init.go:862` (재개 성공) | `p.Info("Initialized MoAI project (resumed: …)")` — 열거 행 | stderr | 안내 | 항목 밖(§3.4 재개 경로) |
| P4 | `init.go:863`, `:1058` | `internal/cli/init_update_notice.go:107` `p.Info("A newer moai binary is available …")` — 열거 행 | stderr | 안내 | `:863` 항목 밖, `:1058` `mcp.entry` |
| P5 | `init.go:1034` | `internal/cli/worktree_advisory.go:57` `emitSeverityLine(out, sevNote, …)` | stdout | 안내 | `hooks.precommit` |
| P6 | `update_template_sync.go:666` | `"settings.local.json env.PATH refresh failed"` | errOut | 경고 | 배포 단계 항목 |
| P7 | `update_template_sync.go:672`, `:679` | `file_snapshot.go:95`, `:99` → `warnLine` | errOut | 경고 | 배포 단계 항목 |
| P8 | `update_template_sync.go:697` | `"manifest save after deploy"` | errOut | 오류 | 배포 단계 항목 |
| P9 | `update_template_sync.go:948` | `"deployment_mode re-assert warning"` | out(stdout) | 경고 | 복원 단계 항목 |
| P10 | `update_template_sync.go:996`, `:1016`, `:1056` | `"manifest retrack (.gitignore)"`, `"(mergeable set)"`, `"(reconciled set)"` | errOut | 오류 | `:996` 은 `verify.gitignore`, `:1016` 은 `verify.retrack`, `:1056` 은 `verify.reconcile` (§4 코드 순서) |
| P11 | `update_template_sync.go:884` (복원 콜백) | `internal/cli/update_noise.go:120` `"'moai update -c' to resync templates for"` | errOut | 안내 | 복원 단계 항목 |
| P12 | `update_template_sync.go:844` | `"moai: migration warning"` (`cmd.ErrOrStderr()`) | stderr | 경고 | 백업 단계 항목 |
| P13 | `update_template_sync.go:695` (`recordTemplateTree`) | `internal/cli/update_template_sync_guard.go:35` `"template tree not recorded"` | errOut | 경고 | 배포 단계 항목 |
| P14 | `update_template_sync.go:1005` (`mergeUserFilesSettlingSnapshot`) | `internal/cli/update/merge/merge.go:310` `"merge failed, preserving user version"`, `:327` `"merged with conflicts (user version preferred)"` | out(stdout) | 경고 | `verify.retrack` (병합 `:1005` 에서 열린다, §4) |
| P15 | `update_template_sync.go:928` (`renderRetainedKeyAdvisory`) | `internal/cli/update_tux.go:329-334` `"setting(s) kept values that differ from the current template default"` 과 그 아래 `· 섹션: 키` 목록 줄 | out(stdout) | 경고 | 복원 단계 항목 |
| P16 | `update_template_sync.go:512` (`archiveLegacySkills`), `:548`·`:583` (`reportArchiveShortfall`) | `internal/cli/update_archive.go:344` `"archive drift backup: "`, `:427` `"Legacy skill archive"` | out(stdout) | 경고 | 검증·정리 단계 항목 |
| P17 | `update_template_sync.go:995`, `:1015`, `:1055` (`retrackManifestFiles`) | `internal/cli/update_manifest_retrack.go:57` `"manifest retrack skip"` | errOut | 경고 | `:995` 은 `verify.gitignore`, `:1015` 은 `verify.retrack`, `:1055` 은 `verify.reconcile` (§4 코드 순서) |
| P18 | `update_template_sync.go:1211` (`stampMigrationDeployMode`) | `internal/cli/update_migrate.go:154`, `:161` `"deployment_mode write warning"` | errOut | 경고 | 꼬리 — 마지막으로 열린 항목(결정 D-10) |
| P19 | `update_template_sync.go:1216` (`warnStampSkew`) | `update_template_sync_guard.go:105` `"template stamp disagrees with deployed bytes"` | out(stdout) | 경고 | 꼬리 — 마지막으로 열린 항목 |
| P20 | `init.go:1022`, `update_template_sync.go:1177` (`installPrePushHookOptional`) | `internal/cli/hook_install.go:278` `"existing pre-push hook preserved"` | init stderr, update stdout | 안내 | init `hooks.prepush`, update 꼬리 |
| P21 | `init.go:1027`, `update_template_sync.go:1182` (`installPreCommitHookOptional`) | `internal/cli/hook_install_precommit.go:400` `"existing pre-commit hook preserved"` | init stderr, update stdout | 안내 | init `hooks.precommit`, update 꼬리 |
| P22 | `update_template_sync.go:1260-1266` (이미 최신 — 실제 명령이 닿는 바깥 함수 `runTemplateSyncWithProgress` 의 조기 반환, `update.go:604`) | `:1262` 결과 알약 `report.OutcomeAlreadyUpToDate` — 열거 행 | out(stdout) | 안내 | 항목 밖 — 본문보다 앞이라 열린 항목이 없고, 넘김 규칙으로 정확히 한 번 보인다(§3.4). 본문의 같은 모양 `:249-255` 는 생산 호출이 닿지 않아 열거하지 않는다 |
| P23 | `init.go:859` → `internal/cli/init_resume.go` | `:168` `"Warning: global settings env update failed"`, `:175` `"Pre-push hook installation failed"`, `:178` `"Pre-commit hook installation failed"`, `:201` `"Warning: manifest retrack (config sections) failed"` | stderr | 경고·오류 | 항목 밖(재개는 검증 항목이 `✗` 로 닫힌 뒤 돈다, §3.4) |

제외 행. 검색에 걸리지만 보존 행이 아닌 일치와 그 이유.

| # | 일치 | 이유 |
|---|---|---|
| X1 | `init.go:901`, `:907` (`emitSeverityLine(…, sevWarn, …)`) | 경고 신호 경로가 덮는다(REQ-TCL-015, §3.3) |
| X2 | `init.go:1022`, `:1027`, `init_resume.go:174`, `:177`, `update.go:508`, `:774` | 작성기를 넘기는 호출·필드일 뿐 쓰는 줄이 아니다. 쓰는 줄은 P20·P21·P23 과 넘김 규칙이 다룬다 |
| X3 | `update.go:231`, `:367`, `:381`, `:458`, `:463`, `:485`, `:536`, `:589`, `:647`, `:712`, `:742`; `update_integrity_probe.go:150`, `:158`(`update.go:631`); `update_dryrun_preview.go:67`, `:85`, `:109`(`update.go:380`); `update_archive.go:389`(`update.go:373`); `migrate_project_assets.go:672`, `:692`(`update_template_sync.go:1323`); `feedback_participation.go:102`, `:107`(`update.go:649`, `:714`) | 템플릿 동기화 밖에서 쓰인다 — 열린 항목이 없어 넘김 규칙으로 그대로 보인다(§3.2) |
| X4 | `init.go:411`, `:575`, `:1086` | `:411`·`:575` 는 첫 항목보다 앞(넘김), `:1086` 은 활성 때 요약이 대신하는 완료 카드(결정 D-6) |
| X5 | `update_clean_install.go:491`, `:550`, `:557`, `:590`, `:601`, `:613`, `:618`; `update_version.go:357` | 항목을 열지 않는 경로(clean-install, 바이너리 갱신·재실행 — spec.md §5) |
| X6 | `init_warnings.go:82`, `:83` | 활성 때 내지 않는 경고 패널(REQ-TCL-015) |
| X7 | `update_action_block.go:81`, `update/backup/settings_snapshot.go:111`, `uikit/styles.go:28`, `:31`; 주석 일치 `update_participation.go:85`, `update.go:1148`, `update_snapshot_hook.go:33`, `update_noise.go:85`, `:113`, `update_manifest_heal.go:14` | 쓰는 줄이 아니다(장부 행 생성, 함수 선언, 글리프 정의, 주석). `settings_snapshot.go:115`(쓰는 줄)는 P1·P2·P7 의 본체다 |
| X8 | `update_model_key_strip.go:137` (`sevNote`, 보존된 설정 키 안내); `update_recovery_manifest.go:79` (`"Warning: could not write"`, 실패 경로의 복구 안내) | `:137` 은 사용자 행동이 필요 없는 상태 안내라 성공 항목에서 접힌다. `:79` 는 실패 경로에서만 쓰이며 REQ-TCL-016 이 캡처 바이트를 모두 쏟는다 |
| X9 | 궤도 밖 파일 전부(파일 단위; research.md §9 첫 블록에서 궤도 접두가 아닌 모든 줄) | init·update 경로가 부르지 않는 다른 명령의 코드 |
| X10 | `update/backup/backup.go:122`; `update/backup/snapshot.go:95`, `:101`, `:105`; `update/backup/restore.go:125`, `:158`, `:222`, `:248`, `:282`, `:308` (`fmt.Fprintf(os.Stderr, "Warning: …")`) | **캡처 밖 쓰기.** 명령 작성기가 아니라 프로세스 stderr 에 바로 쓰므로 세션이 보지도 접지도 못한다 — 사라지지 않고 정확히 한 번 보인다. 다만 실행 중 줄 위에 섞여 찍힐 수 있다(plan.md §H R-12). 작성기를 받도록 `update/backup` 을 고치는 일은 이 카드 범위 밖이다 |

- 경고·오류 보존 행이 항목 안에서 쓰이면 그 항목 줄을 `!` 로 끝낸다. 안내 행은 글리프를 바꾸지 않는다. 어느 쪽도 경고 수에 더하지 않는다(결정 D-9).
- 항목 밖 행은 넘김 규칙으로 이미 한 번 보인다. 세션은 그 줄을 다시 펼치지 않는다.
- 심각도 줄 도우미의 warn·err 줄(`emitSeverityLine(…, sevWarn|sevErr, …)`)은 위 검색 모양에 없고 보존 행도 아니다. 이 줄들은 경고 신호 경로(REQ-TCL-015)가 덮는다 — init 의 `init.go:901`·`:907`(X1)뿐 아니라 update 의 두 작성기 모두: stderr 쪽 `hook_install.go:268`·`:273`·`:283`·`:291`, `hook_install_precommit.go:389`·`:395`·`:405`·`:413`, `update_migrate.go:158`, `update_template_sync.go:958`·`:961`, `update_snapshot_hook.go:34`, stdout 쪽 `hook_install.go:293`, `hook_install_precommit.go:415`(트리 `69358ecba8` 에서 다시 읽음). 그래서 세션의 두 작성기가 모두 경고 수신 인터페이스를 구현한다. 판정은 AC-TCL-016 `update` 하위 테스트(3회차 감사 M1).
- AC-TCL-024 는 보존 행마다 하위 테스트를 하나씩 두고, 목록 완전성 하위 테스트(`inventory_complete`)가 위 검색과 같은 12개 모양을 Go 코드로 `internal/cli/**/*.go`(테스트 제외)에 돌려 두 방향을 본다 — (가) 분류되지 않은 일치가 하나라도 있으면 실패, (나) 보존 행의 분류 조각이 그 파일에서 사라졌으면 실패. 궤도 파일의 일치를 파일 단위 제외로 처리하면 실패한다. 테스트 안의 분류 표는 이 표의 행 id 와 같다. 분류 조각은 메시지 문자열이라 호출 지점을 도우미로 감싸도 파일에 남는다. 조각 자체를 바꿔야 하거나 검색이 새 일치를 내면 run 단계는 막힘 보고를 내고, manager-spec 이 이 표를 고친 뒤 테스트 표를 맞춘다(SPEC 본문은 run 단계가 고치지 않는다).

### §3.4 종료 (REQ-TCL-016, REQ-TCL-023)

- 세션 종료는 `runInit`·`runUpdate` 에서 가장 먼저 등록하는 `defer` 로 둔다. 그래서 가장 나중에 실행되고, 뒤에 등록된 `emitSummary`·세션 워크트리 정리·update 종료 블록보다 늦게 돈다. 종료 함수는 명명된 오류 반환값(`runInit` 의 `(err error)`, `init.go:382`)을 읽는다.
- **머리줄은 첫 항목이 열릴 때 한 번 그린다.** 항목이 하나도 열리지 않은 실행은 머리줄도 요약도 그리지 않고, 기존 줄만 넘김 규칙으로 보인다(아래 update 이미 최신, init 취소).
- **요약 없이 끝나는 실행의 수집기 경고(REQ-TCL-015, REQ-TCL-016, 3회차 감사 M3).** 열린 항목 없이 기록된 수집기 경고는 원래 요약이 다시 나열하는데, 요약이 그려지지 않으면 갈 곳이 없다. `Collect` 는 줄을 쓰지 않기 때문이다(`internal/cli/init_warnings.go:49-56`). 그래서 세션 종료는 — 오류·취소·`nil` 조기 반환 모두 — 그런 경고 가운데 문장이 아직 실제 작성기에 나가지 않은 것(즉시 줄이 없는 `Collect` 경고)을 하나씩 정확히 한 번 실제 작성기에 쓴 뒤 작성기를 되돌린다. 즉시 줄이 이미 나간 `Warn` 경고는 다시 쓰지 않는다. 패널은 이 경로에서도 내지 않는다. 실제 사례: init 실행기가 실패하면 `init.go:829-838` 이 `result.Warnings` 를 `p.Collect(w)`(`:837`)로 모은다. 이때 열린 항목은 없다 — 실행기가 `internal/core/project/phase.go:125` `StepError` 로 `templates.deploy` 를 이미 `Fail` 로 닫았기 때문이다(§3.6). 요약은 오류 경로에서 그려지지 않으므로, 이 규칙이 없으면 오늘 패널이 보여 주는 스킬 미러 안내(REQ-IDE-004)가 체크리스트에서 사라진다.
- **오류로 끝나거나 위저드가 취소되면**(요약 전, `err != nil` 또는 취소 신호): 열린 항목이 있으면 `Fail` 로 닫고, 아직 화면에 나가지 않은 캡처 바이트를 모두 그대로 쏟아내고, 위 규칙의 수집기 경고를 쓴 뒤 작성기를 되돌린다.
- **`nil` 로 조기 반환하면**(요약 전): 열린 항목이 있으면 `Done` 으로 닫고, 아직 보이지 않은 보존 줄을 펼치고, 위 규칙의 수집기 경고를 쓴 뒤 작성기를 되돌린다. 상세는 접힌 채 둔다. 이미 끝난 줄은 — `✗` 로 끝난 줄도 — 다시 그리지 않는다(REQ-TCL-002). 실제 코드에서 이 경우는 두 갈래이고, 두 갈래 모두 반환 시점에 열린 항목이 없다. 열린 항목을 `Done` 으로 닫는 조항은 세션 단위 경우(AC-TCL-017 (e))가 지킨다.
  - **update 이미 최신**(실제 명령 경로, 3회차 감사 M2): `runUpdate` → `internal/cli/update.go:604` `runTemplateSyncWithProgress` 가 `update_template_sync.go:1260` 에서 같은 조건(`packageVersion == projectVersion && !forceUpdate && versionMatchSkipsSync`)을 먼저 보고, 맞으면 `:1262` 에서 결과 알약을 쓰고 `:1266` `return true, nil` 한다. 항목은 본문 안에서만 열리므로(§3.2) 이때 열린 항목이 없다 — 알약은 넘김 규칙으로 정확히 한 번 보이고(보존 행 P22, 항목 밖), 머리줄·요약은 그려지지 않는다. 뒤이은 `update.go:623-650`(무결성 검사, 모델 키 정리, 참여 질문)도 열린 항목 없이 넘김 규칙으로 보인다. 본문 `runTemplateSyncWithReporter` 의 같은 모양 조기 반환(`:249-255`)은 실제 명령이 닿지 않는다 — `runTemplateSync(cmd` 는 정의(`:120`)만 있고 생산 호출이 없으며, 본문으로 들어가는 생산 경로는 이 바깥 함수의 같은 조건 검사 뒤뿐이다(`:215` `forceBackup` 도 같은 `force` 플래그). 그 경로는 본문을 직접 부르는 테스트만 닿으므로 판정 근거로 쓰지 않는다(AC-TCL-017 (d)).
  - **init 재개**(`init.go:854-864`): 실제 이벤트 순서는 아래와 같고, 재개가 시작될 때 열린 항목은 없다(research.md §8, 원장 L21).
    1. `internal/core/project/phase.go:66` `StepStart("Detection")` → `:72` `StepComplete` (`templates.detect`)
    2. (모드를 주지 않았으면) `:82` `StepStart("Methodology Detection")` → `:96` `StepComplete`, (주었으면) `:103` `StepStart("Validation", …)` → `:104` `StepComplete` (`templates.validate` 를 연다)
    3. `:111` `StepStart("Validation", "Validating project structure")` → `:112` `phaseValidate` 가 `:194` `ErrProjectExists`("project already initialized", `internal/core/project/errors.go:12`)를 돌려줌 → `:113` `StepError(err)` → `:114` 오류 반환. 보고기 어댑터는 `StepError` 에서 `templates.validate` 를 `Fail("project already initialized")` 로 닫는다(§3.6).
    4. `init.go:829` 오류 → `:854` "already initialized" 확인 → `:858` `resuming` → `:859` `resumeInitializedProject`(그 안의 경고 줄은 항목 밖, 보존 행 P23) → `:862` 재개 안내(P3) → `:863` 갱신 안내(P4) → `:864` `return nil`.
    화면은 이렇게 끝난다 — `✓ 프로젝트 구조 감지`, `✗ 프로젝트 검증` 과 그 아래 들여쓴 `project already initialized`, 재개 단계가 넘김 규칙으로 쓴 줄, `Initialized MoAI project (resumed: …)` 한 번. 요약은 없다. 실패 항목은 재개가 성공해도 `✗` 로 남는다 — 검증은 실제로 실패했고, 재개는 그 실패에서 이어 가는 회복이기 때문이다. 이 줄을 `✓` 로 바꾸려면 보고기가 실패를 보류해야 하는데, 그러면 재개 단계의 경고 줄이 열린 검증 항목의 상세로 접히므로 택하지 않는다.

### §3.5 대화형 입력 경계 (REQ-TCL-018)

- 위저드(`internal/cli/wizard/wizard.go:129` `form.Run()`)와 병합 확인(`internal/cli/update_template_sync.go:1373-1374`)은 `cmd` 작성기를 쓰지 않으므로 캡처되지 않는다.
- 렌더러는 대화형 입력이 도는 동안 끝나지 않은 줄을 갖지 않는다. init 의 첫 항목은 위저드가 끝난 뒤 보고기 이벤트로 열리고, update 의 확인은 `analyze` 항목을 닫은 뒤 열린다.

### §3.6 보고기 연결

init 은 `executor.SetReporter(newSpinnerReporter(p))`(`init.go:728`)로 보고기를 받는다. 이 줄은 t1966 의 변경 범위(`init.go:730-749` 삭제) 바로 옆이라 건드리지 않는다. 대신 `newSpinnerReporter`(`internal/cli/reporter.go:42-44`)가 받은 printer 에서 세션을 꺼낼 수 있으면(수집기가 세션 접근자를 노출) 체크리스트 보고기를 돌려준다. 보고기의 `StepStart` 이름을 §4 표의 `templates` 항목으로 옮긴다.

이벤트 대응(실행기가 내는 이름은 `internal/core/project/phase.go:66`, `:82`, `:103`, `:111`, `:122`, 원장 L21):

| 보고기 이벤트 | 체크리스트 동작 |
|---|---|
| `StepStart("Detection", …)` | `templates.detect` 를 연다 |
| `StepStart("Methodology Detection", …)`, `StepStart("Validation", …)` | `templates.validate` 를 연다. 이미 열려 있으면 그대로 둔다(한 항목 한 줄 — 명시 모드에서는 `"Validation"` 이 두 번 온다) |
| `StepStart("Initialization", …)` | `templates.deploy` 를 연다 |
| `StepUpdate(msg)` | 열린 항목의 스피너를 한 칸 움직이고 `msg` 를 상세로 기록한다 |
| `StepComplete(msg)` | `msg` 를 상세로 기록한다. 항목을 닫지 않는다 — 닫는 것은 §3.2 항목 경계의 네 경우뿐이다 |
| `StepError(err)` | 열린 항목을 `Fail(err.Error())` 로 닫는다. 기존 보고기가 `r.spin.Fail` 로 줄을 닫는 것(`internal/cli/reporter.go:108`)과 같은 시점이다 |

실패 보류는 두지 않는다. 그 이유는 §3.4 재개 경로에 적었다.

## §4 그룹·항목 계획표 (SSOT) — `internal/cli/checklist_plan.go`

한 표에 모은다. 그룹당 항목 8개 이하(REQ-TCL-010). 열의 키는 카탈로그 키 접미사다. 항목 순서는 코드 실행 순서다.

| 명령 | 그룹 키 | 항목 키 (순서대로) | 대응하는 지금 코드 |
|---|---|---|---|
| init | `templates` | `detect`, `validate`, `deploy` | 보고기 `StepStart` 이름 Detection·Validation·Initialization (`internal/cli/testdata/tuxiu/postm4/init.notty.stderr.golden:2-9` 이 보여 주는 순서) |
| init | `config` | `autonomy`, `profile`, `optins`, `harness`, `globals` | 자율 등급 `init.go:928-939`, 프로필 `:958-960`, 동의(Jev `:970-972`, 참여 `:980-982`), 하네스·배포 모드·evolution `:991-1011`, 전역 설정 `:1014-1016` |
| init | `hooks` | `prepush`, `precommit` | `init.go:1022-1029` |
| init | `mcp` | `entry` | `init.go:1045-1048` |
| init | `verify` | `manifest` | `init.go:1073-1075` |
| update | `templates` | `version`, `load`, `manifest`, `analyze` | 본문 `runTemplateSyncWithReporter` 의 `update_template_sync.go:206-343`. `version` 은 본문 진입(`:206`)에서 열어 `:226` 배포 모드 줄과 `:234` 동기화 머리 줄을 상세로 접는다. `load` `:268`, `manifest` `:282`, `analyze` `:317-343` |
| update | `backup` | `config`, `namespace` | 단계 루프의 `Backup` 가로채기 — `config` `:737`, `namespace` `:768`(마이그레이션 경고 `:844` 까지) |
| update | `deploy` | `validate`, `clean`, `deploy`, `restore` | 단계 표의 Validate Templates(`:467`), 정리 단계(`:498` `cleanManagedPathsStage`), Deploy Templates(`:625`), 그리고 Restore Settings 가로채기의 복원 `:853-959` |
| update | `verify` | `retention`, `gitignore`, `retrack`, `reconcile` | Restore Settings 가로채기 안, 코드 순서대로: `retention` `:960-974`(오래된 백업 정리), `gitignore` `:976-999`, `retrack` `:1000-1017`(병합 `:1005` + 재기록 `:1015`), `reconcile` `:1018-1059`(`len(reconPending) > 0` 일 때만 연다, `:1027`) |

init 의 사용자 자산 보장(`init.go:780-790`)은 항목으로 만들지 않는다(레인 결정 Q3). 그 바로 앞 줄(`init.go:780`)을 t1966 이 바꾸기 때문이다. 이 단계는 `executor.Execute`(`init.go:829`)보다 먼저 돌아 열린 항목이 없으므로 출력이 넘김 규칙으로 지금처럼 보이고, 항목화는 t1966 착지 뒤 후속 카드로 넘긴다.

update 의 clean-install 경로(`update_clean_install.go`), 바이너리 갱신, `--check`·`--shell-env`·`--config` 경로는 항목을 열지 않는다. 그 경로의 출력은 넘김 규칙으로 지금 그대로 보인다(spec.md §5).

update 행의 "대응하는 지금 코드" 열은 각 항목이 열리는 줄과 그 항목이 덮는 구간이다. 표의 항목 순서가 곧 코드 실행 순서이고 화면 순서다(§1 규칙). 델타 3회차에 트리 `69358ecba8` 에서 다시 읽은 순서(3회차 감사 M4): 확인 화면은 바깥 함수 `runTemplateSyncWithProgress`(`update_template_sync.go:1244-1297`)에서 본문보다 먼저 돈다. 본문의 단계 표(`:451-711`)는 Backup → Validate Templates → 정리 → Deploy Templates → Restore Settings 순서다. Restore Settings 가로채기 안의 순서는 복원(`:853`) → 오래된 백업 정리(`:960`) → `.gitignore` 병합(`:977`) → 병합·재기록(`:1005`·`:1015`) → 조정(`:1027`)이다. 2회차 표는 조정을 `deploy` 그룹에, 그보다 앞서 도는 정리 줄을 `verify` 그룹에 두어 화면이 `deploy → verify → deploy` 로 그룹을 다시 열었다. 지금 표는 조정을 `verify` 그룹 끝으로 옮기고 `verify` 항목을 코드 순서로 다시 놓았다. M5 는 이 순서로 항목을 열고, 실제 순서가 표와 다르면 막힘 보고로 이 표를 고친다.

**update 줄 예산(델타 3회차 재계산).** 그룹 제목은 그룹이 바뀔 때 한 번 그린다(§2 `Item`). 표의 그룹 순서 `templates → backup → deploy → verify` 에서 다시 열리는 그룹이 없으므로 그룹 제목은 4줄이다. 2회차 표를 코드 순서대로 그리면 `deploy` 가 다시 열려 그룹 제목 5줄, 합계 25줄로 상한을 넘었다(3회차 감사 M4).

| 줄 | 수 | 근거 |
|---|---|---|
| 정체성 띠 | 1 | 넘김 규칙(`internal/cli/update_tux.go:77-82`) |
| 머리·환경 | 2 | 첫 항목 `templates.version` 이 열릴 때(§3.4) |
| 그룹 제목 | 4 | `templates`, `backup`, `deploy`, `verify` 각 한 번 |
| 항목 | 13 | `templates` 4 + `backup` 2 + `deploy` 4 + `verify` 3(`retention`, `gitignore`, `retrack`). `reconcile` 은 `len(reconPending) > 0` 일 때만 열리는데(`:1027`), 깨끗한 update(spec.md §3.3 — 사용자 수정 없음)에는 조정할 파일이 없다(코드 읽기) |
| 요약 | 3 | 완료 줄·백업 경로·복구 명령 |
| 합계 | 23 | 상한 24 에 1줄 여유 |

- `reconcile` 이 열리는 실행이라도(깨끗한 실행이 아님) 항목 14, 합계 24 로 상한 안이다.
- 본문 진입 직후의 두 줄 — `:226` 배포 모드 줄(기록이 있을 때)과 `:234` 동기화 머리 줄 — 은 `version` 항목을 본문 진입에서 열어 그 상세로 접는다. 2회차 계산은 이 두 줄을 세지 않았다.
- 종료 행동 블록은 깨끗한 실행에서 아무것도 내지 않는다(`internal/cli/update_action_block.go:89-91`).
- 훅 설치 줄(`internal/cli/hook_install.go:262` `"Pre-push hook installed"` — `--no-hooks` 가 아니면 update 마다 나옴)과 pre-commit 쪽 같은 줄은 결정 D-10 에 따라 요약 앞의 꼬리, 곧 마지막으로 열린 항목 안에서 쓰이므로 접히고 예산에 들지 않는다. 요약을 결과 알약 자리(`:1144`)에 그리면 이 두 줄이 요약 뒤로 넘어가 25줄이 된다 — D-10 을 택한 이유다.
- 세지 않은 줄: `--yes` 없이 도는 실행은 본문 진입 전에 바깥 함수의 확인 화면(`:1285-1291` — 빈 줄, `Analyzing merge changes` 구역 머리, 미리보기, 대화형 확인)을 열린 항목 없이 넘김 규칙으로 보인다. 대화형 화면이라(REQ-TCL-018) 위 계산에 넣지 않았다. 깨끗한 실행의 정의(spec.md §3.3)가 `--yes` 를 적지 않는 점은 이 회차 범위 밖의 잔여로 progress.md §E.1 에 남겼다.
- 위 합계는 코드 읽기로 센 값이다. M5 의 실측이 24줄을 넘으면 기준을 느슨하게 하지 않고 manager-spec 에 되돌린다(plan.md §H R-4).

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
| `checklist.summary.` | `done`, `failed`, `created_files`, `created_dirs`, `changed`, `preserved`, `warnings`, `warnings_outside`, `elapsed`, `deploy_mode`, `next_steps`, `backup`, `recover` |

- 언어 해석: update 는 `.moai/config/sections/language.yaml` 의 `conversation_language`, init 은 위저드 답 또는 프로필 로케일을 쓰고 둘 다 없으면 `en`. 미지원 값은 `Translate` 가 `en` 으로 대체한다(`internal/tui/i18n.go:83-99`).
- 값에 서식 지시자(`%d`, `%s`)가 있으면 네 카탈로그의 지시자 개수가 같아야 한다(AC-TCL-008 이 비교).
- ja·zh 값은 en 값과 달라야 한다. 예외는 식별자만 담은 값(초기 허용 목록: `MCP`)뿐이며, 허용 목록은 AC-TCL-008 테스트 안에 둔다.
- ja·zh 문구는 이 카드에서 초안을 넣는다. 원어민 검수는 범위 밖이다(spec.md §5, Gap G-5).

## §7 골든 전략

| 층 | 위치 | 고정 조건 | 갱신 |
|---|---|---|---|
| 렌더러 단위 | `internal/tui/testdata/checklist-*.golden` | 고정 시계, `LightTheme`·`MonochromeTheme`, 고정 환경 요약 | `UPDATE_GOLDEN=1` 과 `MANIFEST.txt`(`internal/tui/golden/index_test.go:43-82`) |
| CLI 대체 경로 | `internal/cli/testdata/checklist/fallback/` | 버퍼 작성기(비TTY), 고정 프로젝트 디렉터리, acceptance.md AC-TCL-013 의 정규화 정규식 | M0 에서 **코드 변경 전에** 따로 커밋(§8). 이후 수정 금지 |
| CLI 리치 경로 | `internal/cli/testdata/checklist/rich/` | 활성 판정 강제 켜기 시접과 버퍼 작성기 | M4·M5 |

대체 골든 파일(M0, 모두 `fallback/` 아래):

- `init-{clean,warn,fail}-{stdout,stderr}.golden` — 6개
- `update-{clean,warn,fail}-{stdout,stderr}.golden` — 6개
- `update-verbose-{stdout,stderr}.golden` — 2개. 고정 실행이 verbose 에서만 나오는 줄을 1줄 이상 내야 한다(예: 3-way 병합 실패 대체 경고 `internal/cli/update_noise.go:111-114`).

`warn` 시나리오는 수집기 경고 1건 이상을, `fail` 시나리오는 실패 1건을 주입한다. 정규화는 줄을 지우지 않는다(줄 수 불변, AC-TCL-013).

리치 골든 파일: `rich/init-clean.golden`, `rich/init-warn.golden`, `rich/update-clean.golden`, `rich/update-fail.golden`.

렌더러 골든 이름: `checklist-<surface>-<case>.golden`, case ∈ {clean, fail, warn, nocolor, reduced, ko}. 기존 `TestGoldenCount` 하한(`internal/tui/golden/index_test.go:27`, 100개)은 늘어나는 방향이라 영향이 없다. `TestManifest` 는 새 골든(`:73`)과 지운 골든(`:80`)을 모두 잡는다.

## §8 기준선을 먼저 커밋하는 이유

`--verbose`·비TTY 출력이 기준선과 바이트 단위로 같다는 주장(REQ-TCL-012, REQ-TCL-013)은 변경 전 출력을 기준으로 한다. 기준선과 구현이 한 커밋에 들어가면 순서를 git 이 증언할 수 없다(`verification-claim-integrity.md` §2.3 — 커밋 그래프만이 순서의 증인이다). 그래서 M0 은 `internal/cli/checklist_fallback_test.go` 와 `internal/cli/testdata/checklist/fallback/` 골든만 담은 커밋 하나로 끝내고, 그 뒤에 코드 변경을 시작한다. M0 커밋은 그 시점 코드에서 이미 녹색이어야 한다(기준선을 바로 그 코드로 찍었으므로).

이 순서를 사후에 깨는 세 가지 길 — 구현 뒤 골든을 다시 찍어 수정 커밋으로 넣기, M0 에 코드 변경을 섞기, CLI 쪽 코드 변경을 M0 보다 먼저 커밋하기 — 은 AC-TCL-013 의 순서 증거 O1·O2·O3 이 각각 막는다.

O1~O3 은 커밋 그래프의 **모양**만 본다. 넷째 길 — 이력을 고쳐 다시 찍은 골든을 M0 에 합치거나, M0 에 새 동작을 손으로 적은 골든을 넣는 것 — 은 모양을 바꾸지 않으므로 O4 가 따로 막는다. 커밋 그래프는 순서를 증언하지만 골든을 누가 무엇으로 만들었는지는 증언하지 못한다(`verification-claim-integrity.md` §2.3). 그래서 O4 는 "골든을 변경 전 바이너리가 만들었다"를 주장하지 않고, 그래프가 확인해 줄 수 있는 더 강한 사실로 바꿔 잰다 — **M0 트리의 코드가 기준선 코드와 같고, 그 트리에서 대체 테스트가 녹색이다.** 그러면 골든은 출처와 상관없이 변경 전 코드의 출력과 바이트 단위로 같다. 새 동작을 담은 골든은 M0 트리에서 빨갛다. 이력을 고치면 M0 의 SHA 가 바뀌어 M0 시점 기록(O4a)의 측정 HEAD 와 어긋난다. 구체 명령은 acceptance.md AC-TCL-013 O4. 골든을 별도 하위 디렉터리에 두는 이유가 이것이다: O1 이 디렉터리 하나에 대한 커밋 수로 판정할 수 있고, 같은 상위 디렉터리에 나중에 들어오는 리치 골든이 O1 을 흐리지 않는다.

기존 `internal/cli/testdata/tuxiu/` 골든(`internal/cli/tuxiu_characterization_test.go:14-17`)은 과거에 손으로 찍어 둔 두 묶음을 서로 비교할 뿐 실행 출력을 읽지 않는다. 그래서 이번 변경을 지켜 주지 못하며, 대체 경로 기준선이 따로 필요하다.
