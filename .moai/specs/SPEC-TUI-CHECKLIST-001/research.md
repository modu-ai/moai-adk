# SPEC-TUI-CHECKLIST-001 — 조사

모든 file:line 은 워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`, HEAD `cb30e05b3e`(로컬 `main` 과 같음)에서 이번 실행 중에 읽은 것이다. 읽지 않은 행번호는 적지 않았다. 델타 1회차(HEAD `d241917d17`, 코드는 `cb30e05b3e` 와 같음)에서 §8 의 행을 더 읽었고 §7 의 t1966 ref 를 다시 쟀다. 델타 2회차(HEAD `b422a13d53`, 코드는 같음)에서 §7 을 다시 재고, §8 을 정정·보강하고, §9(보존 줄 고정 검색과 궤도 추적)와 §10(일부러 만든 저장소의 양성·음성 사례)을 더했다.

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

> [DEFERRED → t1978] (v0.2.0) update 연결은 카드 t1978 로 넘어갔다. 이 절과 §8 의 update 사실은 증거로 남기며, 이 SPEC 의 살아 있는 요구사항이 아니다. 이 SPEC 이 update 에 대해 지키는 것은 "출력이 기준선 그대로"(REQ-TCL-012) 하나뿐이다.

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

`git log origin/main -5 -- internal/cli/init.go` 결과(초안 때의 측정. 증거 원장의 행은 아니다 — 원장 L15 는 `git diff -U0 cb30e05b3e -- internal/cli/init.go` 다):

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
| 델타 2회차(HEAD `b422a13d53`) | `git rev-parse WT-moai-init-moai` | `4e3b0a871488d45f2b56e76dbf29958abd3317cb` (종료 0) — 2회차 감사 E-4 와 같은 값 |

ref 는 움직였지만 init.go hunk 머리는 그대로다. 델타 2회차에도 같은 두 호출(`git diff -U0 --output=<스크래치>/t1966-init.diff cb30e05b3e...WT-moai-init-moai -- internal/cli/init.go` → 빈 출력, 종료 0; `/usr/bin/grep -n "^@@" <스크래치>/t1966-init.diff` → 종료 0)로 다시 쟀고, 아래 열 줄과 같은 머리가 나왔다. `git log --oneline --grep=t1969 cb30e05b3e..WT-moai-init-moai` 는 2회차 감사가 빈 출력으로 쟀다(E-4) — t1966 커밋은 아직 이 카드 id 를 언급하지 않는다. 델타 1회차 재측정은 파이프를 쓰지 않으려고 두 번의 호출로 나눴다: `git diff -U0 --output=<스크래치>/t1966-init.diff cb30e05b3e...WT-moai-init-moai -- internal/cli/init.go`(빈 출력, 종료 0) → `/usr/bin/grep -n "^@@" <스크래치>/t1966-init.diff`(종료 0). 둘째 명령의 출력(줄 번호는 diff 파일 안의 위치)은 아래 열 줄과 같은 머리를 담았다 — `5:@@ -103,5 …`, `15:@@ -237,5 …`, `25:@@ -259,0 …`, `33:@@ -332,0 …`, `53:@@ -561,0 …`, `74:@@ -593,0 …`, `78:@@ -596,9 …`, `95:@@ -730,20 …`, `116:@@ -775,6 …`, `124:@@ -866 …`.

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

## §8 신호를 거치지 않는 직접 쓰기 줄 (델타 1회차 HEAD `d241917d17`, 델타 2회차 HEAD `b422a13d53` 정정)

델타 1회차의 아래 목록은 `update_template_sync.go` 한 파일만 훑었다(2회차 감사 N1). 완전한 목록은 §9 의 고정 검색과 design.md §3.3.1 의 분류다.

1회차 감사 D3 이 든 지점을 각 줄을 읽어 다시 확인했다. 결과는 design.md §3.3.1 보존 줄 목록의 근거다.

- `init.go:858-864` — 재개 경로. `p.Info("Initialized MoAI project (resumed: …)")`(`:862`), `flushUpdateNotice(p)`(`:863`), `return nil`(`:864`). **델타 2회차 정정(2회차 감사 N3):** 델타 1회차는 "열린 보고기 항목이 있는 채로 성공 반환한다"고 적었으나 사실이 아니다. 재개 경로는 `executor.Execute` 가 "already initialized" 오류를 낸 뒤에만 탄다(`init.go:829`, `:854`). 그 오류는 검증 단계에서 나고, 실행기는 오류를 돌려주기 전에 보고기를 실패로 닫는다 — `internal/core/project/phase.go:111` `StepStart("Validation", "Validating project structure")` → `:112` `phaseValidate` → `:194` `return fmt.Errorf("%w", ErrProjectExists)` → `:113` `pe.reporter.StepError(err)` → `:114` `return nil, err`(원장 L21, 같은 파일 `:90-142`·`:172-199` 를 읽음). 기존 보고기도 `StepError` 에서 `r.spin.Fail("%v", err)` 로 줄을 닫는다(`internal/cli/reporter.go:102-118` 을 읽음). 그래서 재개가 시작될 때 열린 항목은 없고, 재개 단계(`resumeInitializedProject`, `internal/cli/init_resume.go`)의 출력은 넘김 규칙으로 보인다. 실행기의 보고기 이벤트 이름은 `"Detection"`(`:66`), `"Methodology Detection"`(`:82`, 모드를 주지 않았을 때), `"Validation"`(`:103` 명시 모드 검증, `:111` 구조 검증), `"Initialization"`(`:122`)이다. 이 순서는 코드 읽기이며 실행 관측이 아니다(plan.md §I G-11).
- [DEFERRED → t1978, 그리고 정정됨] `update_template_sync.go:249-255` — 본문 `runTemplateSyncWithReporter` 안의 "이미 최신" 조기 반환. `plVersion.Done("Already up-to-date")`(`:250`) 뒤 결과 알약(`:252`)을 쓰고 `return nil`(`:255`). 델타 2회차의 이 문단은 "체크리스트가 켜지면 `templates.version` 항목이 열린 채 이 알약이 쓰인다"고 적었으나 **틀렸다**: 실제 `moai update` 는 `update.go:604` `runTemplateSyncWithProgress` 를 거치고, 그 바깥 함수가 같은 조건을 `:1260` 에서 먼저 보고 `:1262` 에서 알약을 쓴 뒤 `:1266` 에서 돌아온다 — 본문의 `:249-255` 는 본문을 직접 부르는 테스트만 닿는다(3회차 감사 M2, design.md §3.4, 4회차 감사 N5). 이 사실은 update 연결의 배경이며 이 SPEC 의 요구사항이 아니다.
- `update_template_sync.go:1081-1218` — 단계 루프 뒤의 꼬리. 진행 막대(`:1081`), 결과 알약(`:1144` `renderUpdateOutcome`), 조정 결과(`:1148`), 훅 설치(`:1177` `installPrePushHookOptional`, `:1182` pre-commit), 배포 모드 기록(`:1211` `stampMigrationDeployMode` → `internal/cli/update_migrate.go:154`·`:161`), 스탬프 검사(`:1216` `warnStampSkew` → `internal/cli/update_template_sync_guard.go:105`), `return nil`(`:1218`). 훅 설치는 `--no-hooks` 가 아니면 매번 `"Pre-push hook installed (.git/hooks/pre-push)"`(`internal/cli/hook_install.go:262`)를 쓴다 — `:255-284` 를 읽음. 이것이 결정 D-10 의 근거다. 본문 함수는 `runTemplateSyncWithReporter` 이고, 그 뒤 `runTemplateSyncWithProgress` 가 `:1323` 에서 `verifyMigrationFinalState` 를 부른다(`:1296-1325` 를 읽음).
- `init.go:885`, `:890` — `backup.StageDeployedSettingsSnapshot` / `StageDeployedMCPSnapshot(…, cmd.ErrOrStderr())`. 쓰기 실패는 `internal/cli/update/backup/file_snapshot.go:95`, `:99` 의 `warnLine` 으로 간다. `warnLine` 본체는 `internal/cli/update/backup/settings_snapshot.go:111-116`, 쓰는 줄은 `:115` `fmt.Fprintf(warn, "%s %v\n", prefix, err)`.
- `init.go:944`, `:945` — `backup.SettleSettingsSnapshot` / `SettleMCPSnapshot`. 승격 실패는 `file_snapshot.go:139` 의 `warnLine`.
- `init.go:1034` — `emitWorktreeAdvisory(cmd.OutOrStdout(), …)` → `internal/cli/worktree_advisory.go:57` `emitSeverityLine(out, sevNote, …)`. stdout 으로 쓰는 안내다.
- `init.go:1058` — `flushUpdateNotice(p)` → `internal/cli/init_update_notice.go:107` `p.Info("A newer moai binary is available: %s (current: %s). Run 'moai update' to upgrade and refresh templates.", …)`. 바로 위 `:1045` 의 `mcp.entry` 항목이 열린 채라면 접힌다.
- `internal/cli/update_template_sync.go` — `/usr/bin/grep -n -e "errOut" -e "SymWarning"` 결과에서 확인한 직접 쓰기: `:666` `"  settings.local.json env.PATH refresh failed: %v\n"`, `:672`·`:679` 스냅숏 단계(`errOut` 전달), `:697` `"  manifest save after deploy: %v\n"`, `:948` `"  %s deployment_mode re-assert warning: %v\n"`(`out` 으로, `uikit.SymWarning()`), `:996`·`:1016`·`:1056` `"  manifest retrack (…): %v\n"`. `:958`·`:961` 은 `emitSeverityLine(errOut, sevWarn, …)` 이라 경고 신호 경로(REQ-TCL-015)가 덮는다.
- `internal/cli/update_noise.go:111-114` — `--verbose` 일 때만 `emitSeverityLine(errOut, sevWarn, …, "3-way merge failed for %s, falling back to 2-way", …)`. verbose 전용 stderr 줄의 근거(1회차 감사 D4). `:119-120` — 대체 횟수가 임계값에 닿으면 `sevNote` 안내 한 번.
- `internal/cli/init_warnings.go:58-63` — `Count()` 는 `len(w.warnings)`. `Collect`(`:52-56`)와 `Warn`(`:41-47`)만 이 목록에 더한다. `emitSeverityLine` 은 수집기를 거치지 않으므로(`init.go:901`, `:907`) 기존 완료 카드의 경고 수(`init.go:1087` `p.Count()`)에 들어가지 않는다(1회차 감사 D6, 결정 D-9).
- `internal/cli/update_action_block.go:88-91` — `renderUpdateTerminalBlock` 은 장부가 비면 아무것도 쓰지 않는다. [DEFERRED → t1978] 단, 동기화가 끝까지 도는 update 에서는 장부에 세 행이 무조건 기록되어(`update_template_sync.go:1165`, `:1167`, `:1190`) 장부가 비지 않는다 — 4회차 감사 N1 (나), design.md §4.1. `internal/cli/update_tux.go:77-82` — 정체성 띠는 한 줄(`◆ MoAI-ADK` + 버전 알약 + `runtime.Version()`).
- 백업 디렉터리 시각 형식: `internal/cli/update/backup/retention_realdata_test.go:53` `time.Parse("20060102_150405", name)` — 테스트 파일에서 읽은 형식이며, 이름을 만드는 생산 코드는 이번에 읽지 않았다. AC-TCL-013 정규식 `\b\d{8}_\d{6}\b` 의 근거다.

## §9 보존 줄 목록의 고정 검색과 궤도 추적 (델타 2회차, HEAD `b422a13d53`)

2회차 감사 N1 은 델타 1회차의 보존 줄 목록이 `update_template_sync.go` 한 파일만 훑어 실제 직접 쓰기 경고(`update_migrate.go:154`·`:161`, `update_template_sync_guard.go:35`·`:105`)를 빠뜨렸다고 지적했고, 감사 자신의 검색도 `internal/cli` 일부에 그쳤다고 적었다(감사 Gaps 마지막 항목). 그래서 `internal/cli` 와 그 하위 패키지 전체(테스트 파일 제외)를 한 번의 호출로 훑었다.

검색 모양 열두 개는 이 코드베이스가 직접 쓰기 경고·오류·안내를 내는 모양이다 — 경고·오류 글리프(`SymWarning()`, `SymError()`), 경고·오류 체크 줄(`CheckLine("warn"`, `CheckLine("err"`), 스냅숏 경고 도우미(`warnLine(`), 오류 작성기로의 직접 쓰기(`Fprintf(errOut`, `Fprintln(errOut`, `Fprint(errOut`, `Fprintf(warn,`), 명령 stderr 로의 직접 쓰기(`OrStderr(),`), 안내 심각도(`sevNote,`), 그리고 대소문자를 가리지 않는 `warning:` 접두. 마지막 모양은 첫 검색(11개 모양, 295줄)에서 빠진 `update_recovery_manifest.go:79` 의 `"Warning: could not write"` 와 `update/backup` 의 `fmt.Fprintf(os.Stderr, "Warning: …")` 를 보고 더했다. 첫 검색 결과는 판정에 쓰지 않는다 — 아래가 원장 L20 의 측정이다.

명령(원장 L20, 한 번의 호출, 종료 0, 360줄):

```
/usr/bin/grep -rni --exclude='*_test.go' -e "SymWarning()" -e "SymError()" -e "CheckLine(\"warn\"" -e "CheckLine(\"err\"" -e "warnLine(" -e "Fprintf(errOut" -e "Fprintln(errOut" -e "Fprint(errOut" -e "Fprintf(warn," -e "OrStderr()," -e "sevNote," -e "warning:" internal/cli
```

출력 원문(줄 순서는 파일 시스템의 순회 순서다. 완전성 하위 테스트는 집합으로 비교하므로 순서에 기대지 않는다):

```text
internal/cli/todo_history.go:165:		if _, werr := fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err); werr != nil {
internal/cli/todo_history.go:177:		if _, werr := fmt.Fprintf(errOut, "history: answered by %s; no archive is available\n", vouch.Store); werr != nil {
internal/cli/todo_history.go:182:			if _, werr := fmt.Fprintf(errOut, "history: answered by %s; its archive tables are missing; no archive is available\n", vouch.Store); werr != nil {
internal/cli/todo_history.go:260:		if _, werr := fmt.Fprintf(errOut,
internal/cli/todo_history.go:295:		if _, err := fmt.Fprintf(errOut,
internal/cli/update_participation.go:85:// here as a warning: an update must not fail because the ask could not be
internal/cli/severity_line.go:8:// Warning: / warning: / hint: — each Fprintln'd with its own casing, some on
internal/cli/update_action_block.go:81:	l.reference = append(l.reference, updateActionRow{sev: sevNote, msg: fmt.Sprintf(format, args...)})
internal/cli/web_port.go:94:	_, _ = fmt.Fprintf(errOut, "Stopping existing moai web instance (PID %d) and reusing port %d\n", pid, port)
internal/cli/integration_base_freshness.go:133:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "base-freshness: not measured — cannot resolve the working directory")
internal/cli/integration_base_freshness.go:136:	return checkBaseFreshness(cmd.ErrOrStderr(), root, dir, cardID, integrationBranch, requestedBranch, omittedBranchTakesTarget)
internal/cli/factory_card.go:2354:						_, _ = fmt.Fprintln(cmd.ErrOrStderr(), refusal.Error())
internal/cli/factory_card.go:2381:						_, _ = fmt.Fprintln(cmd.ErrOrStderr(), holdLine)
internal/cli/factory_card.go:2411:	return factoryNextWriteOutput(cmd.OutOrStdout(), cmd.ErrOrStderr(), factoryCardRoot(), card, openPRs)
internal/cli/factory_card.go:4253:		_, _ = fmt.Fprintf(w, "%s warning: skipped %d unparseable line(s)\n", factoryRecordUnavailableTag, r.UnavailableSkipped)
internal/cli/testdata/pm3-record-t1640.md:15:RESULT PM-3: FAIL ;; COMMAND: cd "$HOME/.moai/scratch/t1640-pm3/repo" && claude --worktree "$HOME/.moai/worktrees/projects/pm3-scratch/pm3tree" -p "Run the shell commands pwd and git rev-parse --show-toplevel and reply with their two output lines only" --output-format text ;; OUTPUT: (stdout empty; stderr verbatim:) `Warning: claude.ai connectors are disabled because ANTHROPIC_API_KEY or another auth source is set and takes precedence over your claude.ai login · Unset it to load your organization's connectors` + `Error creating worktree: Invalid worktree name "/Users/goos/.moai/worktrees/projects/pm3-scratch/pm3tree": each "/"-separated segment must be non-empty and contain only letters, digits, dots, underscores, and [the name-charset set]` ;; EXIT: 1 ;; TOOLS: claude 2.1.303 (Claude Code) at /Users/goos/.local/bin/claude ;; TREE: aa3f9a79b484ac4fb64f13abc02a932ec2876b89
internal/cli/epic.go:85:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/spec_view.go:79:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", e)
internal/cli/spec_view.go:82:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", e)
internal/cli/spec_view.go:87:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", e)
internal/cli/todo_triage.go:494:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_triage.go:539:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/todo_issuance.go:279:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), presText)
internal/cli/mx_query.go:57:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "InvalidQuery: --kind value ' '%s' is invalid. Allowed values: note, warn, anchor, todo, legacy, debt\n", kind)
internal/cli/mx_query.go:104:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/mx_query.go:187:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
internal/cli/mx_query.go:206:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/harness.go:259:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness apply --execute: %s\n", execErr.Error())
internal/cli/update_manifest_retrack.go:57:			_, _ = fmt.Fprintf(errOut, "  manifest retrack skip %s: %v\n", rel, err)
internal/cli/web.go:103:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/web.go:128:	if err := ensurePortFree(cmd.ErrOrStderr(), webPort, !webNoReuse); err != nil {
internal/cli/contract_decide.go:83:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai contract: decide: %v\n", err)
internal/cli/contract_decide.go:97:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), res.JevNotice)
internal/cli/contract_revoke.go:66:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai contract: revoke: %s has no signed contract; nothing written\n", specID)
internal/cli/plan.go:73:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/init_resume.go:168:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: global settings env update failed: %v\n", globalErr)
internal/cli/init_resume.go:174:	if pushErr := installPrePushHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr(), cmd.ErrOrStderr()); pushErr != nil {
internal/cli/init_resume.go:175:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Pre-push hook installation failed: %v\n", pushErr)
internal/cli/init_resume.go:177:	if commitErr := installPreCommitHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr(), cmd.ErrOrStderr()); commitErr != nil {
internal/cli/init_resume.go:178:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Pre-commit hook installation failed: %v\n", commitErr)
internal/cli/init_resume.go:201:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: manifest retrack (config sections) failed: %v\n", retrackErr)
internal/cli/todo_claim.go:160:	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_analysis.go:140:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/hook_install.go:278:		emitSeverityLine(out, sevNote, th, "existing pre-push hook preserved (no MoAI-ADK marker found)")
internal/cli/glm.go:331:			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), factoryRelaunchSupersededNote)
internal/cli/glm.go:363:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "auto mode requires a supported Claude model; GLM models are not supported (see Claude Code permission-modes docs)")
internal/cli/glm.go:364:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "use 'moai cc --permission-mode auto' instead")
internal/cli/glm.go:371:	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: moai glm uses GLM models for the MAIN SESSION. Known limitations:")
internal/cli/glm.go:372:	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  - Main session context window: 1M (glm-5.3, glm-5.3-flash)")
internal/cli/glm.go:373:	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "  - Z.AI concurrency is limited (1-3 in-flight requests per paid tier)")
internal/cli/glm.go:374:	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Mixed Claude/GLM teammate roles require verified teammate routing support.")
internal/cli/glm.go:390:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai: entering the existing worktree %s\n", entryDir)
internal/cli/update/merge/merge.go:310:			_, _ = fmt.Fprintf(out, "  %s %s merge failed, preserving user version: %v\n", uikit.SymWarning(), fb.Path, mergeErr)
internal/cli/update/merge/merge.go:327:			_, _ = fmt.Fprintf(out, "  %s %s merged with conflicts (user version preferred)\n", uikit.SymWarning(), fb.Path)
internal/cli/update/backup/file_snapshot.go:95:		warnLine(warn, s.writeFailedPrefix, err)
internal/cli/update/backup/file_snapshot.go:99:		warnLine(warn, s.writeFailedPrefix, err)
internal/cli/update/backup/file_snapshot.go:139:		warnLine(warn, s.promoteFailedPrefix, err)
internal/cli/update/backup/backup.go:122:		_, _ = fmt.Fprintf(os.Stderr, "Warning: could not save template defaults: %v\n", err)
internal/cli/update/backup/settings_snapshot.go:111:func warnLine(warn io.Writer, prefix string, err error) {
internal/cli/update/backup/settings_snapshot.go:115:	_, _ = fmt.Fprintf(warn, "%s %v\n", prefix, err)
internal/cli/update/backup/snapshot.go:95:			_, _ = fmt.Fprintf(os.Stderr, "Warning: snapshot mkdir %s: %v\n", filepath.Dir(dst), mkErr)
internal/cli/update/backup/snapshot.go:101:			_, _ = fmt.Fprintf(os.Stderr, "Warning: snapshot read %s: %v\n", relPath, readErr)
internal/cli/update/backup/snapshot.go:105:			_, _ = fmt.Fprintf(os.Stderr, "Warning: snapshot write %s: %v\n", relPath, writeErr)
internal/cli/update/backup/restore.go:125:			_, _ = fmt.Fprintf(os.Stderr, "Warning: skipping symlink backup entry %s (containment)\n", backupPath)
internal/cli/update/backup/restore.go:158:			_, _ = fmt.Fprintf(os.Stderr, "Warning: skipping restore target %s escaping config dir (containment)\n", targetPath)
internal/cli/update/backup/restore.go:222:			_, _ = fmt.Fprintf(os.Stderr, "Warning: merge failed for %s, restoring backup\n", relPath)
internal/cli/update/backup/restore.go:248:			_, _ = fmt.Fprintf(os.Stderr, "Warning: skipping symlink backup entry %s (containment)\n", backupPath)
internal/cli/update/backup/restore.go:282:			_, _ = fmt.Fprintf(os.Stderr, "Warning: skipping restore target %s escaping config dir (containment)\n", targetPath)
internal/cli/update/backup/restore.go:308:			_, _ = fmt.Fprintf(os.Stderr, "Warning: merge failed for %s, restoring backup\n", relPath)
internal/cli/harness_validate.go:60:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
internal/cli/harness_validate.go:130:			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), errMsg)
internal/cli/cc.go:278:			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), factoryRelaunchSupersededNote)
internal/cli/cc.go:341:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai: entering the existing worktree %s\n", entryDir)
internal/cli/update.go:231:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Update checker", "not available", "using current version", &th))
internal/cli/update.go:367:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Worktree migration preview", "failed", migrationErr.Error(), &th))
internal/cli/update.go:381:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Reconciliation preview", "failed", previewErr.Error(), &th))
internal/cli/update.go:458:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Deny-rule migration", "failed", stripErr.Error(), &th))
internal/cli/update.go:463:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Deny-rule migration", "failed", normErr.Error(), &th))
internal/cli/update.go:485:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "v2 detection", "failed", fpErr.Error(), &th))
internal/cli/update.go:508:				ErrOut: cmd.ErrOrStderr(),
internal/cli/update.go:536:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai: user-asset phase failed (continuing): %v\n", err)
internal/cli/update.go:589:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai: user-asset phase failed (continuing with the project phase): %v\n", err)
internal/cli/update.go:647:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Participation ask", "failed", err.Error(), &th))
internal/cli/update.go:712:		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Participation ask", "failed", err.Error(), &th))
internal/cli/update.go:742:		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "v2 detection", "failed", fpErr.Error(), &th))
internal/cli/update.go:774:		ErrOut:           cmd.ErrOrStderr(),
internal/cli/update.go:1148:	// CLAUDE_DISABLE_PATH_WARNING: Suppresses Claude Code's PATH warning globally.
internal/cli/update_version.go:357:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Re-exec", "failed", err.Error(), &th))
internal/cli/update_snapshot_hook.go:33:		// "Warning: " prefix.
internal/cli/launch_chain.go:90:	_, _ = fmt.Fprintf(warn, "Warning: chain node not recorded for this launch: %v\n", err)
internal/cli/update_clean_install.go:491:		_, _ = fmt.Fprintf(errOut, "[clean-reinstall] settings.local.json env.PATH refresh failed: %v\n", err)
internal/cli/update_clean_install.go:550:			_, _ = fmt.Fprintf(out, "[clean-reinstall] unmanaged config restore warning: %v\n", unmanagedErr)
internal/cli/update_clean_install.go:557:			_, _ = fmt.Fprintf(out, "[clean-reinstall] retired model key removal warning: %v\n", stripErr)
internal/cli/update_clean_install.go:590:		_, _ = fmt.Fprintf(out, "[clean-reinstall] settings merge warning: %v\n", mergeErr)
internal/cli/update_clean_install.go:601:			_, _ = fmt.Fprintf(out, "[clean-reinstall] .gitignore merge warning: %v\n", mergeErr)
internal/cli/update_clean_install.go:613:		_, _ = fmt.Fprintf(out, "[clean-reinstall] deny-rule migration warning: %v\n", stripErr)
internal/cli/update_clean_install.go:618:		_, _ = fmt.Fprintf(out, "[clean-reinstall] deny-rule migration warning: %v\n", normErr)
internal/cli/agent_model_flags_retired.go:28:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/agent_model_flags_retired.go:29:			"Warning: --%s is deprecated and has no effect: subagents now inherit the main session's model and effort. "+
internal/cli/factory_anchor_guard.go:297:	_, _ = fmt.Fprintf(warn, "factory leader launch: card-tree anchor override — launching from %s (branch %q)\n",
internal/cli/factory_anchor_guard.go:300:		_, _ = fmt.Fprintf(warn, "factory leader launch: REFUSED — the bypass was NOT recorded in %s (%v); the launch is refused without an audit row\n",
internal/cli/factory_anchor_guard.go:305:	_, _ = fmt.Fprintf(warn, "factory leader launch: bypass recorded in %s\n", config.DefaultLeaderAnchorAuditRelPath)
internal/cli/todo_relate.go:52:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_relate.go:56:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_relate.go:63:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_relate.go:167:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_relate.go:181:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", mutErr)
internal/cli/todo_classify.go:84:		_, _ = fmt.Fprintln(errOut, todoClassificationFallbackNotice)
internal/cli/update_model_key_strip.go:137:		emitSeverityLine(w, sevNote, resolveTheme(), "retained key %q absent from new template (preserved from user config)", ref.Key)
internal/cli/uikit/styles.go:28:func SymError() string { return ErrorStyle.Render(string(tui.GlyphErr)) }
internal/cli/uikit/styles.go:31:func SymWarning() string { return WarnStyle.Render(string(tui.GlyphWarn)) }
internal/cli/decision.go:139:				_, _ = fmt.Fprintf(os.Stderr, "decision read: warning: %d unparseable line(s) in %s\n", res.Unparseable, board)
internal/cli/codex_review_delay.go:273:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex-review-entry: invalid stdin JSON (%v); emitting default output\n", err)
internal/cli/codex_review_delay.go:282:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "codex-review-entry: error:", entryErr)
internal/cli/printer/printer.go:67:	// Warn writes a "Warning: ..." status line to stderr.
internal/cli/printer/printer.go:304:	p.status("warn", p.theme.Warning, tui.StatusIcon("warn"), "Warning: ", format, args...)
internal/cli/update_integrity_probe.go:150:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Integrity", "probe internal error", fmt.Sprint(r), &th))
internal/cli/update_integrity_probe.go:158:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Integrity", fmt.Sprintf("%s (%s)", e.rel, reason), "", &th))
internal/cli/session.go:390:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/codex_review_gate.go:779:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex-review-gate: invalid stdin JSON (%v); emitting default output\n", err)
internal/cli/codex_review_gate.go:788:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "codex-review-gate: error:", gateErr)
internal/cli/hook.go:654:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "spec-status: error:", err)
internal/cli/hook.go:659:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "spec-status: write output:", writeErr)
internal/cli/hook.go:675:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "session-start-compact: invalid stdin JSON (%v); emitting default output\n", err)
internal/cli/hook.go:686:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "session-start-compact: error:", err)
internal/cli/hook.go:846:		_, _ = fmt.Fprintf(errOut, "harness-observe: retention prune spawn failed (non-blocking): %v\n", err)
internal/cli/hook.go:906:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness-observe: event recording failed: %v\n", err)
internal/cli/hook.go:996:		_, _ = fmt.Fprintf(errOut, "harness-observe-stop: event recording failed: %v\n", err)
internal/cli/hook.go:1007:		_, _ = fmt.Fprintf(errOut, "harness-observe-stop: auto-classify failed (non-blocking): %v\n", classifyErr)
internal/cli/hook.go:1009:		_, _ = fmt.Fprintf(errOut, "harness-observe-stop: auto-classify %d patterns → %d promotions\n", patternCount, promoCount)
internal/cli/hook.go:1020:				_, _ = fmt.Fprintf(errOut, "harness-observe-stop: auto-propose failed (non-blocking): %v\n", pErr)
internal/cli/hook.go:1022:				_, _ = fmt.Fprintf(errOut, "harness-observe-stop: auto-propose generated %d proposal(s)\n", n)
internal/cli/hook.go:1134:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness-observe-subagent-stop: event recording failed: %v\n", err)
internal/cli/hook.go:1313:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness-observe-user-prompt-submit: event recording failed: %v\n", err)
internal/cli/hook.go:1421:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness-classify error: %v\n", err)
internal/cli/hook.go:1426:	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness-classify: %d patterns → %d promotions written\n", patternCount, promoCount)
internal/cli/worktree/launch_ledger.go:35:		_, _ = fmt.Fprintf(warn, "Warning: could not prune the launch ledger: %v\n", err)
internal/cli/worktree/clean.go:114:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), lockSourceUnreadableNotice(lockErr))
internal/cli/worktree/clean.go:127:			_, _ = fmt.Fprintf(out, "  Warning: could not check %s: %v\n", wt.Branch, err)
internal/cli/worktree/clean.go:159:				_, _ = fmt.Fprintf(out, "  Warning: could not remove %s: %v\n", wt.Path, err)
internal/cli/worktree/clean.go:245:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), lockSourceUnreadableNotice(lockErr))
internal/cli/worktree/clean.go:292:			_, _ = fmt.Fprintf(out, "  Warning: could not remove %s: %v\n", c.Path, err)
internal/cli/worktree/remove.go:62:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: --force removing %s while %d live session(s) are anchored there:\n%s\n",
internal/cli/worktree/remove.go:77:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: --force removing %s while a live session is anchored there - %s\n", absPath, detail)
internal/cli/worktree/guard.go:178:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "suspect flag written: %s\n", flagPath)
internal/cli/worktree/guard.go:282:	_, _ = fmt.Fprintf(out, "WARNING: this will discard local changes against tracked files at HEAD %s\n", short)
internal/cli/worktree/sweep.go:332:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "base %s is absent on its remote; using the remote default branch %s as the sweep base\n", derived, effective)
internal/cli/worktree/sweep.go:664:		complete, hoistErr := sweepHoistEvidence(cmd.ErrOrStderr(), r.Path)
internal/cli/worktree/sweep.go:677:				_, _ = fmt.Fprintf(out, "  Warning: could not remove %s: %v\n", r.Path, err)
internal/cli/worktree/sweep.go:686:				_, _ = fmt.Fprintf(out, "  Warning: could not remove %s [%s]: %v\n", r.Path, r.Branch, err)
internal/cli/worktree/done.go:491:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: --force removing %s while %d live session(s) are anchored there:\n%s\n",
internal/cli/worktree/done.go:527:				fmt.Sprintf("Warning: could not delete branch: %v", err),
internal/cli/wizard/styles.go:24:		Warning: lipgloss.AdaptiveColor{Light: lt.Warning, Dark: dt.Warning},
internal/cli/wizard/wizard.go:686:		Warning:         th.Warning,
internal/cli/agentlint/workflow_lint.go:119:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "error: malformed workflow.yaml: %v\n", err)
internal/cli/gtd.go:284:				mirrorFactoryAssignment(cmd.Context(), cmd.ErrOrStderr(), root, store, runID, result.CardID, lane)
internal/cli/todo_pr.go:189:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_pr.go:196:	rows := computeTodoPRRows(cmd.ErrOrStderr(), rec, only)
internal/cli/graph.go:167:						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "graph refresh refused — unexplained shrink (answering from the existing artifact): %s\n", refuse.Report.Describe())
internal/cli/graph.go:169:						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "graph refresh failed (answering from the existing artifact): %v\n", rErr)
internal/cli/graph.go:172:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/graph.go:184:				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), pv.Describe())
internal/cli/graph.go:376:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", defect)
internal/cli/todo_hold.go:92:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_hold.go:145:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/factory_relaunch.go:172:				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), relaunchRetireFromRun(ctx, fromRun))
internal/cli/factory_relaunch.go:175:				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), relaunchLaneIgnoredNote(lane))
internal/cli/todo_show.go:61:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_show.go:64:	return renderTodoLookup(cmd.OutOrStdout(), cmd.ErrOrStderr(), rec, normalizeTodoRef(ref), "show")
internal/cli/chain.go:127:		_, _ = fmt.Fprintf(warn, "chain: legacy events retained at %s (canonical ledger already exists; the two are not merged)\n", legacyFile)
internal/cli/chain.go:131:		_, _ = fmt.Fprintf(warn, "chain: migration failed, using legacy path %s: %v\n", legacyFile, err)
internal/cli/spec_lint.go:76:				return argumentError(cmd.ErrOrStderr(), "cannot use --json and --sarif together")
internal/cli/spec_lint.go:78:			if err := validateBaselineFlags(cmd.ErrOrStderr(), baselinePath, updateBaseline, reason, jsonOutput, sarifOutput, strict); err != nil {
internal/cli/spec_lint.go:95:			targets, resolveErr := resolveLintTargets(cmd.ErrOrStderr(), args)
internal/cli/spec_lint.go:109:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "linter error: %v\n", lintErr)
internal/cli/spec_lint.go:219:			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), msg)
internal/cli/spec_lint.go:243:		return argumentError(cmd.ErrOrStderr(), "spec lint: %v", loadErr)
internal/cli/spec_lint.go:329:		case spec.SeverityWarning:
internal/cli/update_dryrun_preview.go:67:		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "[dry-run] stale: "+f.RelPath,
internal/cli/update_dryrun_preview.go:85:			_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "[dry-run] migration candidate: "+f.RelPath,
internal/cli/update_dryrun_preview.go:109:		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "[dry-run] symlinks",
internal/cli/hook_pre_push.go:184:	_, _ = fmt.Fprintf(errOut, "%d of %d commit(s) violate %s convention:\n\n",
internal/cli/hook_pre_push.go:190:			_, _ = fmt.Fprint(errOut, errMsg)
internal/cli/hook_pre_push.go:191:			_, _ = fmt.Fprintln(errOut)
internal/cli/spec_status.go:209:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  warning: skipping %s: parsed status %q is not a valid status\n", specID, currentStatus)
internal/cli/spec_status.go:248:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  error updating %s: %v\n", specID, err)
internal/cli/todo_autodone.go:209:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: todo auto-done: the queue could not be read: %v\n", err)
internal/cli/todo_autodone.go:219:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/todo_autodone.go:244:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: todo auto-done: %v\n", err)
internal/cli/todo_autodone.go:254:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: todo auto-done: %v\n", err)
internal/cli/todo_autodone.go:262:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/launcher.go:85:		"Warning: no model resolved from profile %q — Claude Code will use the "+
internal/cli/launcher.go:209:			_, _ = fmt.Fprintf(launcherStderr, "Warning: failed to record last-used profile: %v\n", err)
internal/cli/launcher.go:256:		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to clear tmux session env: %v\n", err)
internal/cli/launcher.go:306:		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to persist team mode: %v\n", err)
internal/cli/launcher.go:311:			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to inject GLM env into tmux session: %v\n"+
internal/cli/launcher.go:411:		return fmt.Sprintf("Warning: failed to disable team mode: %v", err)
internal/cli/launcher.go:709:			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to sync permission mode setting: %v\n", err)
internal/cli/update_noise.go:85://   - verbose=true  → emit legacy "Warning: 3-way merge failed for X, falling
internal/cli/update_noise.go:113:		// Card t1527 D4: ! severity line replaces the raw "Warning:" prefix.
internal/cli/update_noise.go:120:		emitSeverityLine(errOut, sevNote, resolveTheme(), "'moai update -c' to resync templates for %s", relPath)
internal/cli/init_warnings.go:82:	_, _ = fmt.Fprintln(errOut)
internal/cli/init_warnings.go:83:	_, _ = fmt.Fprintln(errOut, uikit.WarnStyle.Render(b.String()))
internal/cli/memory_drain.go:99:					drainPrintf(cmd.ErrOrStderr(), "warning: %s: %v\n", tree, drainErr)
internal/cli/preference/cmd.go:89:			return runDecayScan(cmd.OutOrStdout(), cmd.ErrOrStderr(), flags)
internal/cli/preference/toggle.go:298:			return runToggle(cmd.OutOrStdout(), cmd.ErrOrStderr(), flags)
internal/cli/harness/execute.go:440:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "harness execute: %s\n", diagnosticForError(err))
internal/cli/factory_bundle.go:36:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: factory bundle: %v\n", err)
internal/cli/todo_disclosure.go:100:	if err := discloseNonAuthoritativeBacklogJSON(cmd.ErrOrStderr(), verb,
internal/cli/todo_disclosure.go:104:	return discloseStaleLocalStores(cmd.ErrOrStderr(), verb,
internal/cli/update_template_sync_guard.go:35:		_, _ = fmt.Fprintf(errOut, "  template tree not recorded: %v\n", err)
internal/cli/update_template_sync_guard.go:105:	_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "template stamp disagrees with deployed bytes",
internal/cli/todo_edit_move.go:62:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_edit_move.go:84:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_edit_move.go:128:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_edit_move.go:145:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/hook_stop_goal.go:94:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "stop-goal marshal: %v\n", err)
internal/cli/gate.go:82:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/gate.go:95:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), waitRes.line)
internal/cli/gate.go:103:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "gate-run lock: release failed: %v\n", err)
internal/cli/gate.go:118:			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output)
internal/cli/gate.go:126:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), output)
internal/cli/audit_plan_cmd.go:231:	_, _ = fmt.Fprintln(c.ErrOrStderr(), msg)
internal/cli/factory_age_alarm.go:69:		return config.CardAgeAlarmSettings{}, fmt.Sprintf("age-alarm warning: %v", err)
internal/cli/factory_age_alarm.go:77:		return nil, fmt.Sprintf("age-alarm warning: events read failed: %v", err)
internal/cli/multi_review_gate.go:275:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "multi-review-gate: invalid stdin JSON (%v); emitting default output\n", err)
internal/cli/multi_review_gate.go:284:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "multi-review-gate: error:", gateErr)
internal/cli/todo_drop.go:89:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_drop.go:94:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_drop.go:139:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_drop.go:218:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/codex_audit_launch.go:896:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "codex audit %s: current directory is not inside a git worktree\n", role)
internal/cli/codex_audit_launch.go:912:		Stdout: cmd.OutOrStdout(), Stderr: cmd.ErrOrStderr(),
internal/cli/profile_setup.go:294:		emitProfileScopeNotice(cmd.ErrOrStderr(), wtPath)
internal/cli/profile_setup.go:299:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/profile_setup.go:415:				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to sync profile to project config: %v\n", err)
internal/cli/profile_setup.go:426:				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Warning: failed to persist project config: %v\n", err)
internal/cli/constitution.go:51:			return runConstitutionGuard(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, registryPath, violationsFlag)
internal/cli/constitution.go:70:		_, _ = fmt.Fprintf(wWarn, "Warning: %s\n", warn)
internal/cli/constitution.go:79:			_, _ = fmt.Fprintf(wWarn, "Warning: dangling reference %q - ID not in registry\n", id)
internal/cli/constitution.go:124:			return runConstitutionList(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, registryPath, zoneFilter, fileFlag, formatFlag)
internal/cli/constitution.go:154:		_, _ = fmt.Fprintf(wWarn, "Warning: %s\n", warn)
internal/cli/constitution.go:240:			return runConstitutionValidate(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, registryPath, constitution.ValidateOptions{
internal/cli/constitution.go:242:				FailOnWarning: failOnWarningFlag,
internal/cli/constitution.go:264:		_, _ = fmt.Fprintf(wWarn, "Warning: %s\n", warn)
internal/cli/constitution.go:484:			return runConstitutionAmend(cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, ruleIDFlag, beforeFlag, afterFlag, evidenceFlag, dryRun)
internal/cli/constitution.go:509:		_, _ = fmt.Fprintf(wWarn, "Warning: %s\n", warn)
internal/cli/update_tux.go:331:			uikit.SymWarning(), len(kept))
internal/cli/update_archive.go:344:				_, _ = fmt.Fprintln(out, tui.CheckLine("warn",
internal/cli/update_archive.go:389:		_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "[dry-run] archive: "+id, "will NOT be archived",
internal/cli/update_archive.go:427:	_, _ = fmt.Fprintln(out, tui.CheckLine("warn", "Legacy skill archive",
internal/cli/todo_why.go:37:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/glm_tools.go:574:				"Warning: could not lock %s; proceeding without cross-process guard (best-effort)\n", configPath)
internal/cli/glm_tools.go:595:					"Warning: %s changed concurrently after %d retries; writing best-effort (a concurrent edit may be overwritten)\n",
internal/cli/glm_tools.go:640:			"Warning: could not open lock file %s (%v); proceeding without cross-process guard\n", lockPath, err)
internal/cli/glm_tools.go:647:			"Warning: could not lock %s (%v); proceeding without cross-process guard\n", configPath, err)
internal/cli/glm_tools.go:840:		_, _ = fmt.Fprintf(os.Stderr, "Warning: MCP auto-enable config path resolution failed: %v\n", err)
internal/cli/glm_tools.go:846:		_, _ = fmt.Fprintf(os.Stderr, "Warning: Z.AI MCP auto-enable failed: %v\n", err)
internal/cli/todo_merge.go:45:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/launch_session_defaults.go:67:			_, _ = fmt.Fprintf(launcherStderr, "Warning: %s %q has surrounding whitespace or control characters; using %q\n",
internal/cli/launch_session_defaults.go:71:		_, _ = fmt.Fprintf(launcherStderr, "Warning: %s %q is not a valid session %s; using the built-in default %q\n",
internal/cli/factory_handoff_recover.go:207:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai: lane %s: %s; nothing was written\n", slot, reason)
internal/cli/worktree_branch_flag.go:122:		if _, writeErr := fmt.Fprintf(warn, "moai: worktree %s already exists; re-entering it\n", treePath); writeErr != nil {
internal/cli/verify_run.go:185:		_, _ = fmt.Fprintf(c.ErrOrStderr(), "verify run: cannot start command: %v\n", err)
internal/cli/verify_run.go:190:		_, _ = fmt.Fprintf(c.ErrOrStderr(), "verify run: command exceeded --timeout %s and was terminated\n", p.timeout)
internal/cli/verify_run.go:203:		_, _ = fmt.Fprintf(c.ErrOrStderr(), "verify run: waiting for the command failed: %v\n", err)
internal/cli/update_template_sync.go:666:						_, _ = fmt.Fprintf(errOut, "  settings.local.json env.PATH refresh failed: %v\n", err)
internal/cli/update_template_sync.go:697:					_, _ = fmt.Fprintf(errOut, "  manifest save after deploy: %v\n", saveErr)
internal/cli/update_template_sync.go:844:						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "moai: migration warning: %v\n", migErr)
internal/cli/update_template_sync.go:948:						_, _ = fmt.Fprintf(out, "  %s deployment_mode re-assert warning: %v\n", uikit.SymWarning(), err)
internal/cli/update_template_sync.go:996:						_, _ = fmt.Fprintf(errOut, "  manifest retrack (.gitignore): %v\n", retrackErr)
internal/cli/update_template_sync.go:1016:				_, _ = fmt.Fprintf(errOut, "  manifest retrack (mergeable set): %v\n", retrackErr)
internal/cli/update_template_sync.go:1056:						_, _ = fmt.Fprintf(errOut, "  manifest retrack (reconciled set): %v\n", retrackErr)
internal/cli/factory_quota.go:197:const integrationQuotaWarningPrefix = "quota warning: "
internal/cli/integration.go:181:	return fmt.Sprintf("[moai:integration-lock] warning: git-flow project with no develop branch configured; the window was recorded against the caller's branch %q. Set git_strategy.manual.develop_branch, or pass --branch <integration-target>.", branch)
internal/cli/integration.go:190:	return fmt.Sprintf("[moai:integration-lock] warning: git_strategy workflow %q is not one of the allowed flows (%s); the window was recorded against the caller's branch. Repair the value in .moai/config/sections/git-strategy.yaml.",
internal/cli/integration.go:204:	return fmt.Sprintf("[moai:integration-lock] warning: the %s integration target is %q, but the window was recorded against the caller's branch %q (acquire does not read that target yet). Pass --branch %s if the window is for the integration branch.",
internal/cli/integration.go:307:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "dropped ticket: %s\n", dropped)
internal/cli/integration.go:530:				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), integrationFallbackWarning(branch))
internal/cli/integration.go:538:				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), integrationInvalidWorkflowWarning(gitFlow.Workflow))
internal/cli/integration.go:547:				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), integrationUnwiredTargetWarning(gitFlow.Workflow, gitFlow.IntegrationTarget, branch))
internal/cli/integration.go:555:				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), line)
internal/cli/integration.go:630:				issueCodemapsDebtCard(cmd.ErrOrStderr(), cmd.ErrOrStderr(), resolveProjectDir())
internal/cli/integration.go:642:			issueCodemapsDebtCard(cmd.OutOrStdout(), cmd.ErrOrStderr(), resolveProjectDir())
internal/cli/factory_run_adopt.go:79:			_, _ = fmt.Fprintf(warn, "factory leader resume: run %s (owner %s) still holds live records — adopting nothing; to adopt one deliberately, relaunch with `moai cc -f --factory-run %s`\n", c.RunID, c.Classification, c.RunID)
internal/cli/factory_run_adopt.go:91:		_, _ = fmt.Fprintf(warn, "factory leader resume: run %s is led by codex — a claude leader does not adopt it; retire it first or relaunch the codex leader deliberately\n", adopt.RunID)
internal/cli/factory_run_adopt.go:103:		_, _ = fmt.Fprintf(warn, "factory leader resume: run %s holds live legacy record %q from a binary before the leader/lane rename — adopting nothing; end its sessions and retire with 'moai factory runs --retire %s' first\n", adopt.RunID, value, adopt.RunID)
internal/cli/todo_auto_lane.go:154:		_, _ = fmt.Fprintln(errOut, holdLine)
internal/cli/todo_auto_lane.go:199:				_, _ = fmt.Fprintf(errOut, "%s: %s refused (%s): %s — the pass ends here\n",
internal/cli/todo_auto_lane.go:206:			_, _ = fmt.Fprintf(errOut, "%s: %s refused (%s): %s — next candidate\n",
internal/cli/state.go:102:			fmt.Fprintf(os.Stderr, "Warning: Checkpoint is stale. Use --resume to force load.\n")
internal/cli/migrate_project_assets.go:672:		_, _ = fmt.Fprintf(errOut, "moai: %s manifest unreadable, cannot verify: %v\n", migrationFinalStateSentinel, err)
internal/cli/migrate_project_assets.go:692:		_, _ = fmt.Fprintf(errOut, "moai: %s %s\n", migrationFinalStateSentinel, p)
internal/cli/todo_classify_llm.go:209:		_, _ = fmt.Fprintln(errOut, todoClassificationFallbackNotice)
internal/cli/graph_check.go:90:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "graph check: config load failed: %v\n", cfgErr)
internal/cli/graph_check.go:100:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "graph check: system error: %v\n", err)
internal/cli/graph_check.go:101:				writeUnreachableStampRecovery(cmd.ErrOrStderr(), err)
internal/cli/hook_install_precommit.go:400:		emitSeverityLine(out, sevNote, th, "existing pre-commit hook preserved (no MoAI-ADK marker found)")
internal/cli/doctor_windows_hook_shell.go:8:// warning: a missing dependency.
internal/cli/update_recovery_manifest.go:79:			_, _ = fmt.Fprintf(g.out, "Warning: could not write %s: %v\n", path, writeErr)
internal/cli/factory_quota_lanes.go:138:		return fmt.Sprintf("%s; warning: no live non-Claude lane (unknown backend: %d); nothing is re-dispatched", line, inv.Unknown)
internal/cli/todo.go:102:	_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/todo.go:303:					return runAutoLaneCycle(cmd.Context(), resolveTodoQueueRoot(), cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
internal/cli/todo.go:315:					return runAutoLaneCycle(cmd.Context(), resolveTodoQueueRoot(), cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
internal/cli/todo.go:959:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), presentation)
internal/cli/todo.go:1005:	if err := discloseStaleLocalStores(cmd.ErrOrStderr(), "add",
internal/cli/todo.go:1038:			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), presText)
internal/cli/todo.go:1040:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo.go:1109:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), presText)
internal/cli/todo.go:1112:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "a real add would refuse: %s already holds this card\n", match.ID)
internal/cli/todo.go:1114:	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "dry-run: nothing was written")
internal/cli/todo.go:1140:	if err := discloseStaleLocalStores(cmd.ErrOrStderr(), "add --pick",
internal/cli/todo.go:1163:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo.go:1216:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo.go:1272:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/todo.go:1431:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo.go:1531:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/todo.go:1538:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/todo.go:1604:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo.go:1704:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo.go:1759:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/contract.go:85:	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "moai contract: "+msg)
internal/cli/contract.go:125:		_, _ = fmt.Fprintf(errOut, "warning: configuration not loaded (%v); using defaults\n", lerr)
internal/cli/contract.go:131:		_, _ = fmt.Fprintln(errOut, "warning: "+w)
internal/cli/contract.go:134:		_, _ = fmt.Fprintf(errOut, "configuration error: %v\n", s.DeciderError)
internal/cli/contract.go:165:			_, _ = fmt.Fprintf(errOut, "warning: constitution registry not loaded (%v)\n", err)
internal/cli/contract.go:418:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "written: %s\nunwritten: %s\n",
internal/cli/feedback_participation.go:102:		_, _ = fmt.Fprintln(errOut, "warn: participation drain failed:", err.Error())
internal/cli/feedback_participation.go:107:		_, _ = fmt.Fprintln(errOut, "warn: participation send failed:", err.Error())
internal/cli/init.go:411:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/init.go:575:				_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Initialization cancelled.")
internal/cli/init.go:901:		emitSeverityLine(cmd.ErrOrStderr(), sevWarn, resolveTheme(), "Could not record the harness selection for the resume: %v", recErr)
internal/cli/init.go:907:		emitSeverityLine(cmd.ErrOrStderr(), sevWarn, resolveTheme(), "Could not record the init resume marker: %v", pendErr)
internal/cli/init.go:1022:	if pushErr := installPrePushHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr(), cmd.ErrOrStderr()); pushErr != nil {
internal/cli/init.go:1027:	if commitErr := installPreCommitHookOptional(opts.ProjectRoot, getBoolFlag(cmd, "no-hooks"), cmd.ErrOrStderr(), cmd.ErrOrStderr()); commitErr != nil {
internal/cli/init.go:1086:	_, _ = fmt.Fprintln(cmd.ErrOrStderr(),
internal/cli/tool_policy.go:125:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: local settings.json not found at %s (skipping local target)\n", localPath)
internal/cli/tool_policy.go:148:							_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Info: template %s has a render-time-conditional permissions block (git_mode gating); skipping template target to preserve the conditional.\n", tmplPath)
internal/cli/tool_policy.go:161:						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Info: template settings.json.tmpl not found at %s (consumer project — skipping template target)\n", tmplPath)
internal/cli/tool_policy.go:166:			return reportBuild(cmd.OutOrStdout(), cmd.ErrOrStderr(), results, asJSONFlag)
internal/cli/worktree_advisory.go:57:	emitSeverityLine(out, sevNote, resolveTheme(), "%s", worktreeAdvisoryText(projectRoot))
internal/cli/update_migrate.go:154:			_, _ = fmt.Fprintf(errOut, "  %s deployment_mode write warning: %v\n", uikit.SymWarning(), err)
internal/cli/update_migrate.go:161:			_, _ = fmt.Fprintf(errOut, "  %s deployment_mode write warning: %v\n", uikit.SymWarning(), err)
internal/cli/todo_landed.go:119:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/todo_landed.go:151:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
internal/cli/spec_audit.go:110:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error running audit: %v\n", err)
internal/cli/goal.go:907:			mirrorFactoryAssignment(cmd.Context(), cmd.ErrOrStderr(), root, store, gitOpts.RunID, linkedCardID, gitOpts.Lane)
internal/cli/goal.go:978:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", warn)
internal/cli/goal.go:1187:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(),
internal/cli/goal.go:1197:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", warn)
internal/cli/goal.go:1409:		_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/update_manifest_heal.go:14:// round 2): the ! severity line replaces the raw "warning:" prefix.
internal/cli/todo_export.go:105:	_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/factory_merge.go:286:					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "factory merge ready: the pre-acquire proof failed (%s) and the window release also failed: %v\n", proofWhy, releaseErr)
internal/cli/todo_undone.go:70:				return todoWriteLine(cmd.ErrOrStderr(), err, "Error: %v\n", err)
internal/cli/astgrep.go:104:		_, _ = fmt.Fprintln(cmd.ErrOrStderr(),
internal/cli/astgrep.go:135:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/astgrep.go:144:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/astgrep.go:153:			_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
internal/cli/leader_approval_gate.go:378:				_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
```

**궤도 추적(design.md §3.3.1 의 "궤도 파일"을 정한 근거).** 궤도 밖 파일을 파일 단위로 제외하려면 init·update 가 그 파일을 부르지 않아야 한다. 이름만 보고 판단하면 틀린다는 것을 이번에 직접 봤다 — `feedback_participation.go` 는 다른 명령(`moai feedback`)의 파일처럼 보이지만 update 가 `internal/cli/update.go:649`·`:714` 에서 그 안의 `runParticipationFlushAtUpdate` 를 부른다. 그래서 궤도 파일 19개에서 작성기(`out`, `errOut`, `warn`, `w`, `cmd.ErrOrStderr()`, `cmd.OutOrStdout()`)를 인자로 넘기는 호출을 모두 뽑고, 그 함수를 정의한 파일을 찾았다. 아래는 그 한 단계 추적의 출력이다(스크래치 스크립트 — 파이프를 쓰므로 원장 행이 아니라 분석 근거다). `<not in internal/cli>` 는 다른 패키지 함수이거나 시접 변수다.

```text
append -> <not in internal/cli>
archiveLegacySkills -> internal/cli/update_archive.go
cleanupSessionWorktree -> internal/cli/session_worktree.go
deployWithMirrorNotice -> internal/cli/mirror_notice.go
dryRunArchiveLegacySkills -> internal/cli/update_archive.go
emitLocalInstructionsAdvisory -> internal/cli/migrate_local_instructions.go
emitSlimModeNotice -> internal/cli/init.go
emitSummary -> internal/cli/init_warnings.go
emitWorktreeAdvisory -> internal/cli/worktree_advisory.go
ensureUserAssetsLockedFn -> <not in internal/cli>
enterSessionWorktree -> internal/cli/session_worktree.go
healManifestBestEffort -> internal/cli/update_manifest_heal.go
installPreCommitHookOptional -> internal/cli/hook_install_precommit.go
installPrePushHookOptional -> internal/cli/hook_install.go
mergeUserFilesSettlingSnapshot -> internal/cli/update_settings_snapshot.go
newRecoveryGuard -> internal/cli/update_recovery_manifest.go
normalizeLegacyRootDenySpecifiers -> internal/cli/update_deny_migration.go
previewReconciliation -> internal/cli/update_dryrun_preview.go
provisionMCPEntryUnlessDeclined -> internal/cli/init.go
reconcileManagedRoots -> internal/cli/update_reconcile.go
recordMergeFallback -> internal/cli/update_noise.go
recordTemplateTree -> internal/cli/update_template_sync_guard.go
renderReconciliationOutcome -> internal/cli/update_tux.go
renderRetainedKeyAdvisory -> internal/cli/update_tux.go
renderUpdateOutcome -> internal/cli/update_tux.go
renderUpdateTerminalBlock -> internal/cli/update_action_block.go
reportArchiveShortfall -> internal/cli/update_archive.go
retrackManifestFiles -> internal/cli/update_manifest_retrack.go
retrackSectionFiles -> internal/cli/update_manifest_retrack.go
runManagedSurfaceIntegrityProbe -> internal/cli/update_integrity_probe.go
runParticipationFlushAtUpdate -> internal/cli/feedback_participation.go
runParticipationStep -> internal/cli/update_participation.go
runUpdateMigrationTrigger -> internal/cli/update_migrate.go
runUpdateRestore -> internal/cli/update_restore.go
runUpdateWorktreeMigration -> internal/cli/update_worktree_migration.go
runUserAssetUpdatePhase -> internal/cli/user_asset_phase.go
runV3ResidueCleanup -> internal/cli/update_residue_cleanup.go
sessionExitAutoMerge -> internal/cli/session_worktree_automerge.go
stampMigrationDeployMode -> internal/cli/update_migrate.go
stripRetiredModelConfig -> internal/cli/update_model_key_strip.go
stripRetiredModelConfigOnVersionMatch -> internal/cli/update_model_key_strip.go
stripRetiredV2DenyEntries -> internal/cli/update_deny_migration.go
verifyMigrationFinalState -> internal/cli/migrate_project_assets.go
warnStampSkew -> internal/cli/update_template_sync_guard.go
writeInstallSummary -> internal/cli/user_asset_phase.go
writeTemplateSnapshotBestEffort -> internal/cli/update_snapshot_hook.go
```

추적이 찾은 파일 가운데 첫 검색에 걸린 것은 `feedback_participation.go` 하나였고, 궤도로 넣었다(제외 행 X3 — update 의 템플릿 동기화 뒤에 쓰인다). 나머지 파일(`session_worktree.go`, `mirror_notice.go`, `migrate_local_instructions.go`, `user_asset_phase.go`, `update_*.go` 등)은 궤도 정의에 이미 들거나 검색 일치가 없었다. 이 추적은 한 단계뿐이다(plan.md §I G-10).

**한계(3회차 감사 M5, v0.2.0 에 보강).** 이 추적은 작성기를 **인자로 넘기는** 호출만 본다. init 이 값으로 들고 다니는 출력 타입의 메서드 — `init.go:388` 이 만든 프린터의 `Warn`(`printer/printer.go:304`) — 는 보지 못해 `printer/printer.go` 가 X9 로 잘못 분류됐다. 그래서 출력 타입을 정의한 `printer/printer.go`·`severity_line.go` 를 궤도에 넣었다(design.md §3.3.1). 두 파일의 일치 세 줄은 위 첫 블록에 이미 있다(`printer.go:67`, `:304`, `severity_line.go:8`).

## §10 순서 증거·범위 검사 명령의 양성·음성 사례 (델타 2회차)

2회차 감사 N5·N7 과 리더 codex_audit P2(= N8)는 명령이 의도한 집합을 내는지를 문제 삼았다. 이 카드의 트리에는 아직 코드 커밋이 없어 세 명령 모두 빈 출력이다(원장 L22·L24·L25). 빈 출력만으로는 명령이 무언가를 잡을 수 있는지 알 수 없으므로, 스크래치 디렉터리에 버리는 저장소를 따로 만들어 같은 명령을 돌렸다. 이 카드의 트리는 건드리지 않았다.

- 양성 사례(브랜치 `card`): 기준 커밋(`internal/cli/init.go` 20줄) → M0(`internal/cli/checklist_fallback_test.go` 와 `internal/cli/testdata/checklist/fallback/init-clean-stderr.golden`, 메시지에 `card t1969`) → M2(`internal/tui/checklist.go`) → M4a(init.go 10행 뒤 3줄 삽입) → M4b(원래 15행 — 지금 18행 — 수정). 모두 메시지에 `card t1969`.
- 음성 사례(브랜치 `neg`): 기준 커밋 → X(`internal/cli/init_warnings.go` 수정, `card t1969`) → M0 → Y(`internal/cli/severity_line.go` 추가, 카드 id 없음).

명령은 acceptance.md 와 같은 형태다(O3: `git log --no-merges --grep=t1969 --format=%h <기준>..HEAD ^main -- internal ":(exclude)internal/cli/testdata" ":(exclude)internal/cli/checklist_fallback_test.go"` 와 SHA 마다 `git merge-base --is-ancestor <M0> <SHA>`; 1회차 O3: 같은 범위에서 `internal/cli/checklist_fallback_test.go` 제외가 없는 것; 완전성: `git log --no-merges --invert-grep --grep=t1969 --format=%h <기준>..HEAD ^main -- internal`; AC-TCL-022: `git diff -U0 <기준> HEAD -- internal/cli/init.go` 와 1회차의 `git log --no-merges --grep=t1969 -p -U0 --format=%h <기준>..HEAD -- internal/cli/init.go`; 병합 검사: `git log --merges --format=%h <기준>..HEAD -- internal/cli/init.go`). 출력 원문:

```text
== POSITIVE: base=fdd84fe M0=97bad5c ==
-- O3 corrected selector:
eae97ad
1c17cd1
e2b6d23
exit=0
is-ancestor 97bad5c eae97ad exit=0
is-ancestor 97bad5c 1c17cd1 exit=0
is-ancestor 97bad5c e2b6d23 exit=0
-- O3 OLD selector (internal minus testdata only):
eae97ad
1c17cd1
e2b6d23
97bad5c
-- N7 invert-grep over protected paths:
exit=0
-- N5 OLD per-commit hunks (git log -p, parent coordinates):
eae97ad
@@ -18 +18 @@ l14
1c17cd1
@@ -10,0 +11,3 @@ l10
-- N5 NEW cumulative hunks (git diff base HEAD, base coordinates):
@@ -10,0 +11,3 @@ l10
@@ -15 +18 @@ l14
-- N5 merge guard:
exit=0
== NEGATIVE: X(code before M0)=e8e758c M0=3fc67d4 Y(no card id)=23f0692 ==
-- O3 corrected selector:
e8e758c
is-ancestor 3fc67d4 e8e758c exit=1
-- N7 invert-grep over protected paths:
23f0692
exit=0
```

읽는 법:

- **O3 (N8).** 양성에서 고친 선택자는 M0 뒤의 코드 커밋 셋만 내고 M0 자신은 내지 않으며, 세 조상 검사가 모두 종료 0 이다. 1회차 선택자는 같은 저장소에서 넷째 줄로 M0 자신(`97bad5c`)을 냈다 — 리더 codex_audit P2 가 지적한 그 충돌이다. 음성에서 고친 선택자는 M0 앞의 코드 커밋 X 를 내고, 조상 검사가 종료 1 이다(빨강).
- **카드 id 완전성 (N7).** 양성에서는 빈 출력, 음성에서는 id 없는 커밋 Y 한 줄. 두 경우 모두 종료 0 이라 판정은 출력으로 한다.
- **AC-TCL-022 좌표 (N5).** 1회차의 커밋별 hunk 는 둘째 커밋(M4b)을 `@@ -18 +18 @@` 로 낸다 — 앞 커밋의 3줄 삽입만큼 밀린 부모 좌표다. 누적 diff 는 같은 변경을 기준 좌표 `@@ -15 +18 @@` 로 낸다. 병합 검사는 빈 출력(종료 0).
- 커밋 SHA 는 실행마다 다르다(커밋 시각이 들어간다). 판정에 쓰는 것은 줄의 개수와 종료 코드다.
