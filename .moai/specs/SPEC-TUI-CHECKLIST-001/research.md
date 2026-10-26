# SPEC-TUI-CHECKLIST-001 — 조사

모든 file:line 은 워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`, HEAD `cb30e05b3e`(로컬 `main` 과 같음)에서 이번 실행 중에 읽은 것이다. 읽지 않은 행번호는 적지 않았다. 델타 1회차(HEAD `d241917d17`, 코드는 `cb30e05b3e` 와 같음)에서 §8 의 행을 더 읽었고 §7 의 t1966 ref 를 다시 쟀다.

## §0 측정 환경

- `git rev-parse --short HEAD` → `cb30e05b3e`
- `date +%F` → `2026-10-27`
- 이 셸에서 `grep` 은 ugrep 을 감싼 셸 함수다(`type grep` → `grep is a shell function from …/shell-snapshots/snapshot-zsh-….sh`). 일치 0건에서도 종료 코드 0 을 냈다. 그래서 acceptance.md 의 모든 명령은 `/usr/bin/grep` 을 쓴다. `/usr/bin/grep` 은 일치 0건에서 종료 코드 1 을 냈다(acceptance.md 증거 원장 L2~L7).

## §1 재사용할 internal/tui 자산

| 자산 | 위치 | 이 SPEC 에서의 쓰임 |
|---|---|---|
| `ProgressLine` / `ProgressLineHandle` | `internal/tui/progress_line.go:62-167` | 줄 지우기 접두사 `progressClearPrefix = "\r\x1b[2K"`(`:41`)를 재사용. TTY 판정은 `*os.File` 만 TTY 로 본다(`:218-224`) — 세션 작성기를 끼우면 기존 호출이 자동으로 평문 줄을 낸다 |
| `StatusIcon` | `internal/tui/status.go:18-35` | ok/warn/err/run/skip → `GlyphDone`/`GlyphWarn`/`GlyphErr`/`GlyphRun`/`GlyphSkip` |
| 글리프 상수 | `internal/tui/glyphs.go:18-25` | 허용 목록 `✓ ✗ ! · ● ○ ◆ ◇`(`:9-10`) |
| `reducedMotion` | `internal/tui/status.go:39-42` | `MOAI_REDUCED_MOTION` 이 비어 있지 않고 `0` 이 아니면 참 |
| `Spinner` | `internal/tui/status.go:51-72` | 모션 감소 시 `●`, 아니면 `⠋`. 무상태(`:49-50`) |
| `Stepper` | `internal/tui/status.go:128-156` | 이번 설계에서는 쓰지 않는다(항목 줄이 진행 상황을 대신 보여 줌). design.md §2 |
| 테마 해석 | `internal/tui/detect.go:58-80` | `NO_COLOR` → `MonochromeTheme`. 비TTY 에서 배경 질의를 건너뜀(`:160-165`) |
| 카탈로그 | `internal/tui/i18n.go:18-99` | embed 로 묶은 `messages/*.yaml`, `Translate(key, lang)` 대체 순서 lang → en → key |
| 카탈로그 시험 | `internal/tui/i18n_test.go:11-45` | 동치 검사는 ko·en 두 언어만 본다. ja·zh 는 검사 밖 |
| 이모지 검사 | `internal/tui/i18n_test.go:95-113` | ko·en 값만 본다. 금지 범위에 U+2700–27BF 포함(`:106`) |
| 골든 색인 | `internal/tui/golden/index_test.go:23-73` | `testdata` 아래 골든 수 하한 100(`:27`), `MANIFEST.txt` 동기화, `UPDATE_GOLDEN=1` 로 재생성(`:56`) |

카탈로그 상태(`internal/tui/messages/`): en.yaml 2568B, ko.yaml 2810B, ja.yaml 1051B(update 복원 키 5개만, `ja.yaml:6-13`), zh.yaml 847B. 다른 11개 언어 파일은 322B 자리 표시자다. ja.yaml 머리말은 원어민 번역이 필요하다고 적어 두었다(`ja.yaml:2-3`).

## §2 init 출력 경로

- 출력 관문: `p := newWarnCollector(printer.New(printer.WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr())))`(`internal/cli/init.go:388`), `defer p.emitSummary(cmd.ErrOrStderr())`(`:389`).
- printer 는 stderr 작성기의 isatty 로 모드를 정하고 `NO_COLOR` 면 평문으로 낮춘다(`internal/cli/printer/printer.go:11-23`). 데이터는 stdout, 상태는 stderr(`:4-9`).
- 진행 보고기: `executor.SetReporter(newSpinnerReporter(p))`(`init.go:728`). 보고기 타입은 `internal/cli/reporter.go:23-44`.
- 첫 상태 줄: `p.Info("Initializing MoAI project...")`(`init.go:750`).
- 사용자 자산 보장: `init.go:780-790`, 출력 작성기는 `cmd.OutOrStdout()`(`:784`).
- 수집기를 거치지 않는 경고 두 곳: `emitSeverityLine(cmd.ErrOrStderr(), sevWarn, …)`(`init.go:901`, `:907`).
- 꼬리 단계(전부 경고만 내고 실패시키지 않음): 자율 등급 `:928-939`, 프로필 `:958-960`, Jev `:970-972`, 참여 `:980-982`, 하네스 `:991-993`, 배포 모드 `:1001-1003`, evolution `:1009-1011`, 전역 설정 `:1014-1016`, pre-push `:1022-1024`, pre-commit `:1027-1029`, 워크트리 안내 `:1034`, MCP `:1045-1048`, 갱신 안내 `:1058`, 매니페스트 재기록 `:1073-1075`.
- 완료 카드: `init.go:1081-1087`, 카드 본문은 `buildInitSuccessCard`(`internal/cli/init_warnings.go:99-130`) — 디렉터리 수, 파일 수, 배포 모드, 다음 단계, 경고 수.
- 경고 패널: `emitSummary`(`internal/cli/init_warnings.go:69-84`), 경고 0 이면 아무것도 내지 않음.
- `--verbose` 플래그가 init 에는 없다(acceptance.md 원장 L6).
- `InitResult` 는 생성 디렉터리·파일·경고만 담고 "보존 수" 필드가 없다(`internal/core/project/initializer.go:110-118`).

지금 비TTY init stderr 의 모양(과거 캡처, `internal/cli/testdata/tuxiu/postm4/init.notty.stderr.golden:1-24`): `· Initializing MoAI project...` → 보고기 단계 8줄 → 상자 카드 11줄 → 훅 2줄 → 갱신 안내 1줄.

## §3 update 출력 경로

- `--verbose` 는 이미 있다(`internal/cli/update.go:89`). 뜻: 인정된 경고까지 다 보이고 보존 파일 목록을 경로별로 나열. 패키지 변수 `updateVerboseMode` 로 전달(`:166-167`).
- 주 출력 작성기는 stdout: `out := cmd.OutOrStdout()`(`update.go:156`). 종료 행동 블록이 `defer renderUpdateTerminalBlock(out, th)`(`:176`).
- 정체성 띠: `update.go:212-214`, `renderIdentityBand`(`internal/cli/update_tux.go:77-82`).
- 템플릿 동기화 진입: `runTemplateSyncWithProgress(cmd, userAssetsInstalled)`(`update.go:604`), 본체 `internal/cli/update_template_sync.go:206-`.
- `ProgressLine` 호출 10곳(원장 L12): 버전 확인 `:242`, 템플릿 적재 `:268`, 매니페스트 `:282`, 검증 `:487`, 설정 호환 `:549`, 배포 `:630`, 설정 백업 `:737`, 네임스페이스 백업 `:768`, 복원 `:853`, 조정 `:1028`.
- 단계 표: `update_template_sync.go:450-` (`Backup`, `Validate Templates`, …), 단계 루프 `:731-1075`, 끝에 진행 막대 한 번 `:1081`, 결과 표시 `:1083-`.
- 직접 쓰는 줄: `.gitignore` 보존 `:983`, 백업 정리 `:969`, 배포 모드 재주장 경고 `:948`(`uikit.SymWarning()`), 매니페스트 재기록 오류는 `errOut` 으로(`:996`, `:1016`, `:1056`).
- 병합 확인: `confirmViaPreview`(`:1373-1374`)가 `os.Stdin` 을 직접 본다.

지금 비TTY update stdout 의 모양(과거 캡처, `internal/cli/testdata/tuxiu/postm4/update.notty.stdout.golden:1-60`): 60줄. 단계마다 `○ …` / `✓ …` 두 줄, 진행 막대 5줄, 결과·백업·복구 줄, 훅, 아카이브.

## §4 대화형 입력의 출력 대상

- init 위저드는 `form.Run()`(`internal/cli/wizard/wizard.go:129`) — `WithOutput` 이 없다. 같은 파일에서 `os.Stdout`·`os.Stderr`·`WithInput` 일치는 없었다(`grep` 결과 `:129` 한 줄뿐). huh 기본 출력은 `cmd` 작성기가 아니다.
- update 병합 확인은 `os.Stdin` 이 TTY 일 때 Bubble Tea 미리보기를 띄운다(`update_template_sync.go:1373-1374`).
- 결론: `cmd.SetOut`·`SetErr` 캡처가 질문 화면을 삼키지 않는다. 다만 렌더러가 질문 중에 끝나지 않은 줄을 들고 있으면 화면이 섞이므로 REQ-TCL-018 이 이를 막는다.

## §5 TTY·CI 판정 도우미

- `isTerminalWriter`(`internal/tui/progress_line.go:218-224`), `isTerminalFile`(`internal/tui/detect.go:169-171`)은 패키지 내부 함수다.
- CI 판정은 공용 도우미가 없다. `os.Getenv("CI") != ""` 를 두 곳이 직접 쓴다(`internal/cli/init_participation_wizard.go:34`, `internal/cli/update_participation.go:91`).
- `internal/config/envkeys.go:72` 에 `EnvNoColor = "MOAI_NO_COLOR"` 가 있으나 `CI` 상수는 없다.

## §6 지금 출력을 고정하는 테스트

- `internal/cli/tuxiu_characterization_test.go:129-154` 는 과거에 찍어 둔 두 골든 묶음(`testdata/tuxiu/*.golden`, `testdata/tuxiu/postm4/*.golden`)을 서로 비교한다. 실행 출력을 읽지 않으므로 이번 변경을 감지하지 못한다.
- 버퍼 작성기로 init·update 를 돌리는 기존 테스트는 비TTY 경로를 탄다. 활성 판정 1·2번(design.md §3.1)이 거짓이 되므로 이 테스트들은 대체 경로, 즉 지금 출력을 그대로 본다. 이 SPEC 은 그 테스트들을 바꾸지 않는다.

## §7 형제 카드 t1966 와의 겹침

`git log origin/main -5 -- internal/cli/init.go` 결과(원장 L15 의 근거):

```
8c36a7904b 2026-10-23 t1909: remove general moai codex execution/factory/init surface, keep the read-only Codex audit launcher (#1896)
cd19e5494d 2026-10-14 merge: absorb main ec193e4777 into WT-moai-update-repair (card t1619)
75a6a9ff60 2026-10-10 fix(SPEC-USERASSET-DEPLOY-GUARD-001): keep the interrupted init's settings snapshot until the resume promotes it (card t1591)
9812e0dfb1 2026-10-10 fix(SPEC-USERASSET-DEPLOY-GUARD-001): persist the bundle selection before the ensure and restore it on resume (card t1591)
e4ec6c24ec 2026-10-10 fix(SPEC-USERASSET-DEPLOY-GUARD-001): resume after an early ensure failure (card t1591)
```

t1966 은 아직 `origin/main` 에 없다. 브랜치 `WT-moai-init-moai`(공유 ref)를 이 트리에서 읽었다. 트리 격리 가드가 `git -C <t1966 경로>` 를 거부했기 때문이다(Gap G-1).

ref 이력:

| 측정 | 명령 | 출력 |
|---|---|---|
| 초안(HEAD `cb30e05b3e`) | `git rev-parse WT-moai-init-moai` | `d7f4d5c288…` (짧은 SHA 만 기록) |
| 델타 1회차(HEAD `d241917d17`) | `git rev-parse WT-moai-init-moai` | `1eebf7ba6d40025758a2fff3fc082b71b2eaaa39` (종료 0) |

ref 는 움직였지만 init.go hunk 머리는 그대로다. 델타 1회차 재측정은 파이프를 쓰지 않으려고 두 번의 호출로 나눴다: `git diff -U0 --output=<스크래치>/t1966-init.diff cb30e05b3e...WT-moai-init-moai -- internal/cli/init.go`(빈 출력, 종료 0) → `/usr/bin/grep -n "^@@" <스크래치>/t1966-init.diff`(종료 0). 둘째 명령의 출력(줄 번호는 diff 파일 안의 위치)은 아래 열 줄과 같은 머리를 담았다 — `5:@@ -103,5 …`, `15:@@ -237,5 …`, `25:@@ -259,0 …`, `33:@@ -332,0 …`, `53:@@ -561,0 …`, `74:@@ -593,0 …`, `78:@@ -596,9 …`, `95:@@ -730,20 …`, `116:@@ -775,6 …`, `124:@@ -866 …`.

`git diff -U0 cb30e05b3e...WT-moai-init-moai -- internal/cli/init.go` 의 hunk 머리(출력에서 `@@` 줄만 옮김):

```
@@ -103,5 +103,4 @@ func init() {
@@ -237,5 +236,4 @@ func applyWizardPage3ToOpts(cmd *cobra.Command, result *wizard.WizardResult, opt
@@ -259,0 +258,7 @@ func getBoolFlagWithDefault(cmd *cobra.Command, name string, defaultVal bool) bo
@@ -332,0 +338,19 @@ func validateInitFlags(cmd *cobra.Command, _ []string) error {
@@ -561,0 +586,20 @@ func runInit(cmd *cobra.Command, args []string) (err error) {
@@ -593,0 +638,3 @@ func runInit(cmd *cobra.Command, args []string) (err error) {
@@ -596,9 +643,7 @@ func runInit(cmd *cobra.Command, args []string) (err error) {
@@ -730,20 +774,0 @@ func runInit(cmd *cobra.Command, args []string) (err error) {
@@ -775,6 +800 @@ func runInit(cmd *cobra.Command, args []string) (err error) {
@@ -866 +886 @@ func runInit(cmd *cobra.Command, args []string) (err error) {
```

`cb30e05b3e` 기준 t1966 이 손대는 줄: 103-107, 237-241, 259 뒤, 332 뒤, 561 뒤, 593 뒤, 596-604, 730-749, 775-780, 866. 이 카드가 쓸 영역(plan.md §B)은 383-389, 919-1075 사이 삽입, 1077-1087 이다. 가장 가까운 거리는 t1966 의 866 과 이 카드의 919 사이 53줄이다. 728(보고기)·750(첫 Info)·780-790(자산 보장)은 t1966 hunk 와 맞닿아 있어 이 카드는 손대지 않는다.

이 카드 쪽 init.go 변경은 아직 없다: `git diff -U0 cb30e05b3e -- internal/cli/init.go` → 빈 출력, 종료 0 (HEAD `d241917d17`, acceptance.md 원장 L15).

t1966 브랜치의 같은 diff 는 init.go 외에 `internal/cli/wizard/*`, `internal/cli/init_profile_sync.go` 등 22개 파일도 보여 준다. 그 diff 에는 t1966 이 흡수한 `origin/main 2a1898ea22` 의 t1921 파일이 섞여 있다(`git log --oneline -5 WT-moai-init-moai` 셋째 줄). 이 카드가 고칠 파일 가운데 t1966 diff 에 있는 것은 `internal/cli/init.go` 하나뿐이다. (이 단락의 파일 수는 초안 때 ref `d7f4d5c288` 에서 잰 값이며 델타 1회차에 다시 세지 않았다. 1회차 감사는 같은 범위에서 `internal/cli/wizard` 만 11개 파일을 보았다.)

## §8 신호를 거치지 않는 직접 쓰기 줄 (델타 1회차, HEAD `d241917d17`)

1회차 감사 D3 이 든 지점을 각 줄을 읽어 다시 확인했다. 결과는 design.md §3.3.1 보존 줄 목록의 근거다.

- `init.go:858-864` — 재개 경로. `p.Info("Initialized MoAI project (resumed: …)")`(`:862`), `flushUpdateNotice(p)`(`:863`), `return nil`(`:864`). 열린 보고기 항목이 있는 채로 성공 반환한다(1회차 감사 D8).
- `init.go:885`, `:890` — `backup.StageDeployedSettingsSnapshot` / `StageDeployedMCPSnapshot(…, cmd.ErrOrStderr())`. 쓰기 실패는 `internal/cli/update/backup/file_snapshot.go:95`, `:99` 의 `warnLine` 으로 간다. `warnLine` 본체는 `internal/cli/update/backup/settings_snapshot.go:111-116`, 쓰는 줄은 `:115` `fmt.Fprintf(warn, "%s %v\n", prefix, err)`.
- `init.go:944`, `:945` — `backup.SettleSettingsSnapshot` / `SettleMCPSnapshot`. 승격 실패는 `file_snapshot.go:139` 의 `warnLine`.
- `init.go:1034` — `emitWorktreeAdvisory(cmd.OutOrStdout(), …)` → `internal/cli/worktree_advisory.go:57` `emitSeverityLine(out, sevNote, …)`. stdout 으로 쓰는 안내다.
- `init.go:1058` — `flushUpdateNotice(p)` → `internal/cli/init_update_notice.go:107` `p.Info("A newer moai binary is available: %s (current: %s). Run 'moai update' to upgrade and refresh templates.", …)`. 바로 위 `:1045` 의 `mcp.entry` 항목이 열린 채라면 접힌다.
- `internal/cli/update_template_sync.go` — `/usr/bin/grep -n -e "errOut" -e "SymWarning"` 결과에서 확인한 직접 쓰기: `:666` `"  settings.local.json env.PATH refresh failed: %v\n"`, `:672`·`:679` 스냅숏 단계(`errOut` 전달), `:697` `"  manifest save after deploy: %v\n"`, `:948` `"  %s deployment_mode re-assert warning: %v\n"`(`out` 으로, `uikit.SymWarning()`), `:996`·`:1016`·`:1056` `"  manifest retrack (…): %v\n"`. `:958`·`:961` 은 `emitSeverityLine(errOut, sevWarn, …)` 이라 경고 신호 경로(REQ-TCL-015)가 덮는다.
- `internal/cli/update_noise.go:111-114` — `--verbose` 일 때만 `emitSeverityLine(errOut, sevWarn, …, "3-way merge failed for %s, falling back to 2-way", …)`. verbose 전용 stderr 줄의 근거(1회차 감사 D4). `:119-120` — 대체 횟수가 임계값에 닿으면 `sevNote` 안내 한 번.
- `internal/cli/init_warnings.go:58-63` — `Count()` 는 `len(w.warnings)`. `Collect`(`:52-56`)와 `Warn`(`:41-47`)만 이 목록에 더한다. `emitSeverityLine` 은 수집기를 거치지 않으므로(`init.go:901`, `:907`) 기존 완료 카드의 경고 수(`init.go:1087` `p.Count()`)에 들어가지 않는다(1회차 감사 D6, 결정 D-9).
- `internal/cli/update_action_block.go:88-91` — `renderUpdateTerminalBlock` 은 장부가 비면 아무것도 쓰지 않는다. `internal/cli/update_tux.go:77-82` — 정체성 띠는 한 줄(`◆ MoAI-ADK` + 버전 알약 + `runtime.Version()`).
- 백업 디렉터리 시각 형식: `internal/cli/update/backup/retention_realdata_test.go:53` `time.Parse("20060102_150405", name)` — 테스트 파일에서 읽은 형식이며, 이름을 만드는 생산 코드는 이번에 읽지 않았다. AC-TCL-013 정규식 `\b\d{8}_\d{6}\b` 의 근거다.
