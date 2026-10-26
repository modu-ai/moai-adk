# SPEC-TUI-CHECKLIST-001 — 인수 기준

문서 핀: 원장 L1~L14 의 RED-now 측정은 트리 `cb30e05b3e`(워크트리 `.moai/worktrees/t1969`, 브랜치 `WT-moai-init-update`)에서 했다. 자기 핀이 없는 기준은 이 핀을 따른다. L15~L19 는 트리 `d241917d17a3754541ef5ad52536d68023faf142` 에서, L20~L25 는 트리 `b422a13d5377639d58fc672c986b81c678fe3318` 에서 쟀고 각 행에 핀을 따로 단다. 세 트리는 코드가 같다(`d241917d17`·`b422a13d53` 은 SPEC 디렉터리만 바꾼 plan 커밋 — 1회차 감사 E-3, 2회차 감사 E-1 `git diff --name-only cb30e05b3e HEAD`). L26~L33 은 범위 축소 회차(v0.2.0)에 트리 `6f05252d6ddf62362a60cadbfc3350e0d068dbc4` 에서 쟀다. 그 트리도 코드는 `cb30e05b3e` 와 같다(L32 — 기준 이후 바뀐 경로는 SPEC 디렉터리 여섯 파일뿐). L34~L41 은 델타 4회차(v0.2.1)에 트리 `274441db21a4be3cec01b78c0f9d496893d12479` 에서 쟀다. 그 트리도 코드는 `cb30e05b3e` 와 같다(확인 감사 5회차 E-1 의 L32 재실행 — 바뀐 경로는 SPEC 디렉터리뿐).

**범위(v0.2.0).** 이 파일은 렌더러와 `moai init` 연결을 판정한다. update 연결 쪽 하위 테스트·조건은 `[DEFERRED → t1978]` 로 표시하고 이 SPEC 의 녹색 판정에 넣지 않는다. 이월 목록은 plan.md §K. AC id 는 하나도 지우거나 다시 매기지 않았다.

이 파일은 검증 층이다. 요구사항(REQ)의 GEARS 문장은 spec.md §3 에 있고, 여기서는 각 기준을 Given-When-Then 과 이진 판정으로 적는다. 각 기준 제목의 `maps` 목록이 REQ 대응을 선언한다.

## §A 판정 규칙

- **명령 형식.** 모든 명령은 파이프·리다이렉트·`&&`·`;` 없는 한 번의 호출이다. 이 셸의 `grep` 은 ugrep 셸 함수라 일치 0건에서도 종료 코드 0 을 낸다(research.md §0). 그래서 모든 검색은 `/usr/bin/grep` 으로 적는다.
- **선택자 공허 방지.** `go test -run` 은 일치하는 테스트가 0개여도 `ok … [no tests to run]` 와 종료 코드 0 을 낸다(원장 L11 — AC-TCL-001 의 녹색 명령을 오늘 그대로 돌린 결과다). 그래서 테스트 기반 기준의 녹색 판정은 종료 코드 0 **그리고** 출력에 기준이 이름 붙인 `--- PASS: <테스트명> ` 줄(이름 뒤 공백 포함)이 하나도 빠짐없이 있는 것이다. 선택자는 `^…$` 로 묶고, `-v` 와 `-count=1` 을 항상 붙인다.
- **기준 종류.** release-blocking 기준은 RED-now 네 요소(명령, 그대로의 stdout, 종료 코드, 트리 SHA)를 갖춘다. 오늘 이미 녹색인 범위 경계 검사는 regression-guard 로 분류하고 통과로 세지 않는다.
- **이 카드의 커밋.** 범위 검사는 이 카드의 커밋만 센다 — 메시지에 카드 id `t1969` 가 있고, 병합 커밋이 아니며, 로컬 `main` 에서 닿지 않는 것: `git log --no-merges --grep=t1969 <기준>..HEAD ^main`. `^main` 은 흡수한 `main` 의 커밋을 — 그 메시지가 조율 문구로 `t1969` 를 언급하더라도 — 뺀다(1회차 감사 D7, 2회차 감사 N7 (나)).
- **카드 id 완전성(2회차 감사 N7 (가)).** id 를 빠뜨린 이 카드의 커밋은 위 선택에서 사라진다. 그래서 범위 검사를 쓰는 모든 기준(AC-TCL-013 O3, AC-TCL-019 뒤 절, AC-TCL-021, AC-TCL-022)은 이 검사가 함께 성립할 때만 녹색이다: `git log --no-merges --invert-grep --grep=t1969 --format=%h cb30e05b3e..HEAD ^main -- internal` → 빈 출력. `git log` 는 출력이 있어도 종료 0 이므로 판정은 출력으로 한다 — SHA 가 한 줄이라도 나오면(보호 경로 `internal/` 을 건드린, id 없는 커밋) 빨강. 오늘 빈 출력, 종료 0(원장 L24, regression-guard). 일부러 만든 양성·음성 사례에서 이 명령이 id 없는 커밋 하나를 정확히 낸다(research.md §10).

## §B 증거 원장 (RED-now)

stdout 이 비어 있으면 `(빈 출력)` 으로 적고, stderr 는 따로 적는다.

| ID | 트리 | 명령 | stdout | stderr | 종료 코드 |
|---|---|---|---|---|---|
| L1 | `cb30e05b3e` | `ls internal/tui/checklist.go` | (빈 출력) | `ls: internal/tui/checklist.go: No such file or directory` | 1 |
| L2 | `cb30e05b3e` | `/usr/bin/grep -c "^checklist\." internal/tui/messages/en.yaml` | `0` | (없음) | 1 |
| L3 | `cb30e05b3e` | `/usr/bin/grep -c "^checklist\." internal/tui/messages/ko.yaml` | `0` | (없음) | 1 |
| L4 | `cb30e05b3e` | `/usr/bin/grep -c "^checklist\." internal/tui/messages/ja.yaml` | `0` | (없음) | 1 |
| L5 | `cb30e05b3e` | `/usr/bin/grep -c "^checklist\." internal/tui/messages/zh.yaml` | `0` | (없음) | 1 |
| L6 | `cb30e05b3e` | `/usr/bin/grep -n "\"verbose\"" internal/cli/init.go` | (빈 출력) | (없음) | 1 |
| L7 | `cb30e05b3e` | `/usr/bin/grep -l "checklist" internal/cli/init.go internal/cli/update.go internal/cli/update_template_sync.go internal/cli/init_warnings.go internal/cli/severity_line.go internal/cli/reporter.go` | (빈 출력) | (없음) | 1 |
| L8 | `cb30e05b3e` | `ls internal/cli/testdata/checklist` | (빈 출력) | `ls: internal/cli/testdata/checklist: No such file or directory` | 1 |
| L9 | `cb30e05b3e` | `/usr/bin/grep -rn "func NewChecklist" internal/tui` | (빈 출력) | (없음) | 1 |
| L10 | `cb30e05b3e` | `ls internal/tui/checklist_test.go internal/cli/checklist_session.go internal/cli/checklist_session_test.go` | (빈 출력) | `ls: internal/cli/checklist_session_test.go: No such file or directory` / `ls: internal/cli/checklist_session.go: No such file or directory` / `ls: internal/tui/checklist_test.go: No such file or directory` | 1 |
| L11 | `cb30e05b3e` | `go test ./internal/tui -run '^TestChecklistGoldenClean$' -count=1 -v` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/tui	0.301s [no tests to run]` | (없음) | 0 |
| L12 | `cb30e05b3e` | `/usr/bin/grep -c "ProgressLine(" internal/cli/update_template_sync.go` | `10` | (없음) | 0 |
| L13 | `cb30e05b3e` | `git diff --name-only cb30e05b3e -- internal/template/templates` | (빈 출력) | (없음) | 0 |
| L14 | `cb30e05b3e` | `/usr/bin/grep -c "emitSummary" internal/cli/init.go` | `1` | (없음) | 0 |
| L15 | `d241917d17` | `git diff -U0 cb30e05b3e -- internal/cli/init.go` | (빈 출력) | (없음) | 0 |
| L16 | `d241917d17` | `/usr/bin/grep -rln "TestChecklistInitHunkConfinement" internal/cli` | (빈 출력) | (없음) | 1 |
| L17 | `d241917d17` | `/usr/bin/grep -rln "TestChecklistPreservedLinesExactlyOnce" internal/cli` | (빈 출력) | (없음) | 1 |
| L18 | `d241917d17` | `/usr/bin/grep -rn "func checklistEnv" internal/cli` | (빈 출력) | (없음) | 1 |
| L19 | `d241917d17` | `ls internal/cli/testdata/checklist/fallback` | (빈 출력) | `ls: internal/cli/testdata/checklist/fallback: No such file or directory` | 1 |
| L20 | `b422a13d53` | `/usr/bin/grep -rni --exclude='*_test.go' -e "SymWarning()" -e "SymError()" -e "CheckLine(\"warn\"" -e "CheckLine(\"err\"" -e "warnLine(" -e "Fprintf(errOut" -e "Fprintln(errOut" -e "Fprint(errOut" -e "Fprintf(warn," -e "OrStderr()," -e "sevNote," -e "warning:" internal/cli` | 360줄 — research.md §9 첫 블록에 그대로 옮김 | (없음) | 0 |
| L21 | `b422a13d53` | `/usr/bin/grep -n -e "StepStart(\"Detection\"" -e "StepStart(\"Methodology Detection\"" -e "StepStart(\"Validation\"" -e "StepError(err)" -e "ErrProjectExists" internal/core/project/phase.go` | `66:	pe.reporter.StepStart("Detection", "Analyzing project structure")` / `82:		pe.reporter.StepStart("Methodology Detection", "Determining development mode")` / `103:		pe.reporter.StepStart("Validation", fmt.Sprintf("Using mode: %s", opts.DevelopmentMode))` / `111:	pe.reporter.StepStart("Validation", "Validating project structure")` / `113:		pe.reporter.StepError(err)` / `125:		pe.reporter.StepError(err)` / `194:			return fmt.Errorf("%w", ErrProjectExists)` | (없음) | 0 |
| L22 | `b422a13d53` | `git diff -U0 cb30e05b3e HEAD -- internal/cli/init.go` | (빈 출력) | (없음) | 0 |
| L23 | `b422a13d53` | `git log --merges --format=%h cb30e05b3e..HEAD -- internal/cli/init.go` | (빈 출력) | (없음) | 0 |
| L24 | `b422a13d53` | `git log --no-merges --invert-grep --grep=t1969 --format=%h cb30e05b3e..HEAD ^main -- internal` | (빈 출력) | (없음) | 0 |
| L25 | `b422a13d53` | `git log --no-merges --grep=t1969 --format=%h cb30e05b3e..HEAD ^main -- internal ":(exclude)internal/cli/testdata" ":(exclude)internal/cli/checklist_fallback_test.go"` | (빈 출력) | (없음) | 0 |
| L26 | `6f05252d6d` | `/usr/bin/grep -rn --exclude='*_test.go' -e "PillWarn" -e "PillErr" internal/cli` | `internal/cli/update_wizard.go:39:		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "Project not initialized · Run 'moai init' first", Theme: &th}))` / `internal/cli/update_wizard.go:102:		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "GitHub token was not saved (never stored in plaintext). Run 'gh auth login' to authenticate.", Theme: &th}))` / `internal/cli/update_wizard.go:105:		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: "GitLab token was not saved (never stored in plaintext). Run 'glab auth login' to authenticate.", Theme: &th}))` / `internal/cli/loop.go:233:	pill := tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Label: "일시정지", Theme: &th})` / `internal/cli/loop.go:265:	pill := tui.Pill(tui.PillOpts{Kind: tui.PillErr, Label: "취소됨", Theme: &th})` / `internal/cli/update_tux.go:107:// pills — PillOk "+ N add", PillInfo "~ N update", PillErr "! N conflict"` / `internal/cli/update_tux.go:120:		pills = append(pills, tui.Pill(tui.PillOpts{Kind: tui.PillErr, Label: fmt.Sprintf("! %d conflict", conflict), Theme: &th}))` / `internal/cli/doctor_render.go:148:	pWarn := tui.Pill(tui.PillOpts{Kind: tui.PillWarn, Solid: false, Label: fmt.Sprintf("Warn %d", warnTotal), Theme: &th})` / `internal/cli/doctor_render.go:149:	pErr := tui.Pill(tui.PillOpts{Kind: tui.PillErr, Solid: false, Label: fmt.Sprintf("Fail %d", failTotal), Theme: &th})` | (없음) | 0 |
| L27 | `6f05252d6d` | `ls internal/tui/checklist.go internal/tui/checklist_test.go internal/cli/checklist_session.go internal/cli/checklist_session_test.go internal/cli/checklist_plan.go internal/cli/checklist_fallback_test.go internal/cli/testdata/checklist` | (빈 출력) | `ls: internal/cli/checklist_fallback_test.go: No such file or directory` / `ls: internal/cli/checklist_plan.go: No such file or directory` / `ls: internal/cli/checklist_session_test.go: No such file or directory` / `ls: internal/cli/checklist_session.go: No such file or directory` / `ls: internal/cli/testdata/checklist: No such file or directory` / `ls: internal/tui/checklist_test.go: No such file or directory` / `ls: internal/tui/checklist.go: No such file or directory` | 1 |
| L28 | `6f05252d6d` | `/usr/bin/grep -rln -e "TestChecklistGateCommandLevel" -e "TestChecklistWarningsExactlyOnce" -e "TestChecklistAbortFlushes" -e "TestChecklistPreservedLinesExactlyOnce" -e "TestChecklistChannelPartition" internal/cli` | (빈 출력) | (없음) | 1 |
| L29 | `6f05252d6d` | `/usr/bin/grep -rn -e "func checklistEnv" -e "func NewChecklist" internal/cli internal/tui` | (빈 출력) | (없음) | 1 |
| L30 | `6f05252d6d` | `/usr/bin/grep -n "\"verbose\"" internal/cli/init.go` | (빈 출력) | (없음) | 1 |
| L31 | `6f05252d6d` | `/usr/bin/grep -l "checklist" internal/cli/init.go internal/cli/init_warnings.go internal/cli/severity_line.go internal/cli/reporter.go` | (빈 출력) | (없음) | 1 |
| L32 | `6f05252d6d` | `git diff --name-only cb30e05b3e HEAD` | `.moai/specs/SPEC-TUI-CHECKLIST-001/acceptance.md` / `.moai/specs/SPEC-TUI-CHECKLIST-001/design.md` / `.moai/specs/SPEC-TUI-CHECKLIST-001/plan.md` / `.moai/specs/SPEC-TUI-CHECKLIST-001/progress.md` / `.moai/specs/SPEC-TUI-CHECKLIST-001/research.md` / `.moai/specs/SPEC-TUI-CHECKLIST-001/spec.md` | (없음) | 0 |
| L33 | `6f05252d6d` | `ls internal/cli/testdata/checklist/fallback internal/tui/testdata/checklist-init-clean.golden` | (빈 출력) | `ls: internal/cli/testdata/checklist/fallback: No such file or directory` / `ls: internal/tui/testdata/checklist-init-clean.golden: No such file or directory` | 1 |
| L34 | `274441db21` | `/usr/bin/grep -n -e "emitSlimModeNotice(cmd" -e "Initializing MoAI project" -e "ensureUserAssetsLockedFn(homeDir" -e "emitWorktreeAdvisory(cmd" internal/cli/init.go` | `708:		emitSlimModeNotice(cmd.OutOrStdout())` / `750:	p.Info("Initializing MoAI project...")` / `784:			if err := ensureUserAssetsLockedFn(homeDir, selection, cmd.OutOrStdout()); err != nil {` / `1034:	emitWorktreeAdvisory(cmd.OutOrStdout(), opts.ProjectRoot)` | (없음) | 0 |
| L35 | `274441db21` | `/usr/bin/grep -rn --exclude='*_test.go' --exclude-dir=testdata -F -e "Deploying core templates only (slim mode)" -e "Common skills and agents install into your user folders" -e "Initializing MoAI project..." -e "user-asset ensure" -e "this checkout is shared across concurrent sessions" internal/cli` | `internal/cli/update_action_block.go:18://	    · this checkout is shared across concurrent sessions; ...` / `internal/cli/user_asset_phase.go:54:	writeInstallSummary(out, "user-asset ensure", res)` / `internal/cli/init_resume.go:4:// M6, REQ-SRF-007 + audit DEBT R4): a prior init whose user-asset ensure` / `internal/cli/init_resume.go:44:// the user-asset ensure: the ENTIRE re-check → quarantine → install span` / `internal/cli/init_resume.go:52:// the user-asset ensure: the ENTIRE re-check → quarantine → install span` / `internal/cli/init_resume.go:108:		return fmt.Errorf("user-asset ensure: %w", err)` / `internal/cli/init_resume.go:281:// user-asset ensure, so an interrupted run's resume can read it back.` / `internal/cli/init_resume.go:314:// user-asset ensure window (card t1591, finding B). It sits beside the selection` / `internal/cli/init_resume.go:324:// user-asset ensure runs, so an ensure that fails leaves the rerun a resume route. The` / `internal/cli/init_resume.go:392:	// Finding B (card t1591): a fresh init that died in or after its user-asset ensure` / `internal/cli/init.go:43:		"Deploying core templates only (slim mode). "+` / `internal/cli/init.go:47:		"Common skills and agents install into your user folders. "+` / `internal/cli/init.go:363:// variable so a test can fail or interrupt the user-asset ensure, which runs before` / `internal/cli/init.go:750:	p.Info("Initializing MoAI project...")` / `internal/cli/init.go:844:		// PRIOR attempt died at the user-asset ensure (initialized but` / `internal/cli/worktree_advisory.go:39:		return "this checkout is shared across concurrent sessions; " +` / `internal/cli/worktree_advisory.go:46:	return "this checkout is shared across concurrent sessions; " +` | (없음) | 0 |
| L36 | `274441db21` | `ls internal/cli/testdata/checklist/rich/init-clean.golden internal/cli/testdata/checklist/fallback/init-clean-stdout.golden internal/cli/testdata/checklist/fallback/init-clean-stderr.golden` | (빈 출력) | `ls: internal/cli/testdata/checklist/fallback/init-clean-stderr.golden: No such file or directory` / `ls: internal/cli/testdata/checklist/fallback/init-clean-stdout.golden: No such file or directory` / `ls: internal/cli/testdata/checklist/rich/init-clean.golden: No such file or directory` | 1 |
| L37 | `274441db21` | `/usr/bin/grep -rln -e "TestChecklistScreenBudget" -e "TestChecklistAbortFlushes" -e "TestChecklistWarningsExactlyOnce" -e "checklistSessionStartFn" internal/cli` | (빈 출력) | (없음) | 1 |
| L38 | `274441db21` | `ls internal/tui/checklist_test.go internal/tui/testdata/checklist-probe-fail.golden internal/tui/testdata/checklist-probe-warn.golden` | (빈 출력) | `ls: internal/tui/checklist_test.go: No such file or directory` / `ls: internal/tui/testdata/checklist-probe-fail.golden: No such file or directory` / `ls: internal/tui/testdata/checklist-probe-warn.golden: No such file or directory` | 1 |
| L39 | `274441db21` | `/usr/bin/grep -n -e "catalogCache = map" -e "catalogMu    sync.RWMutex" internal/tui/i18n.go` | `27:	catalogCache = map[string]catalog{}` / `28:	catalogMu    sync.RWMutex` | (없음) | 0 |
| L40 | `274441db21` | `/usr/bin/grep -n -e "user-modified pre-push hook was replaced" -e "confirmViaPreviewFn(analysis" internal/cli/hook_install.go internal/cli/update_template_sync.go` | `internal/cli/hook_install.go:268:				emitSeverityLine(warn, sevWarn, th, "user-modified pre-push hook was replaced; previous hook backed up at %s", installer.lastBackupPath)` / `internal/cli/update_template_sync.go:355:		proceed, err = confirmViaPreviewFn(analysis, projectRoot)` / `internal/cli/update_template_sync.go:1287:		proceed, cerr := confirmViaPreviewFn(analysis, projectRoot)` | (없음) | 0 |
| L41 | `274441db21` | `/usr/bin/grep -n -e "p.Warn(" -e "p.Collect(" internal/cli/init.go` | `427:		p.Warn("git is not installed. Some features (plan/run/sync workflows, branch management) will be limited.\n  %s",` / `531:		p.Warn("failed to read profile preferences: %v", err)` / `788:			p.Warn("Could not resolve the user home; the user-folder asset install was skipped: %v", homeErr)` / `837:				p.Collect(w)` / `935:			p.Warn("Failed to apply autonomy tier bundle: %v", tierErr)` / `938:		p.Warn("Failed to resolve home directory for autonomy tier bundle: %v", homeErr)` / `954:		p.Collect(w)` / `959:		p.Warn("Failed to sync profile to project config: %v", err)` / `971:		p.Warn("Failed to persist the Jev opt-in: %v", err)` / `981:		p.Warn("Failed to persist the participation opt-in: %v", err)` / `992:		p.Warn("Failed to apply harness: %v", err)` / `1002:		p.Warn("Failed to apply deploy mode: %v", err)` / `1010:		p.Warn("Failed to scaffold evolution directory: %v", err)` / `1015:		p.Warn("Failed to update global settings env: %v", err)` / `1023:		p.Warn("Pre-push hook installation failed: %v", pushErr)` / `1028:		p.Warn("Pre-commit hook installation failed: %v", commitErr)` / `1047:		p.Collect("MCP server entry provisioning failed: " + mcpErr.Error())` / `1074:		p.Collect("manifest retrack (config sections) failed: " + retrackErr.Error())` | (없음) | 0 |

- L11 은 RED 가 아니라 **대조군**이다. 아직 없는 테스트를 고르는 선택자가 `PASS` 와 종료 코드 0 을 낸다. §A 의 `--- PASS:` 요구가 여기서 나온다. 종료 코드만 보는 판정은 이 SPEC 의 모든 테스트 기반 기준을 오늘 녹색으로 만든다.
- L12·L14 는 기준선이다. L13 은 오늘 이미 비어 있어 AC-TCL-021 은 regression-guard 다.
- L15 는 델타 1회차의 AC-TCL-022 RED 전제(작업 트리 대 기준)였다. 델타 2회차에서 AC-TCL-022 가 커밋 사이 누적 diff 로 바뀌어 RED 는 L22 가 잰다(종료 0, 빈 출력 = 기준 이후 init.go hunk 0개). L23 은 init.go 를 건드린 병합 커밋이 없음을 잰다(오늘 녹색 — 범위 보호 조건). L16 은 그 녹색 테스트가 아직 없음을 잰다.
- L20 은 보존 줄 목록(design.md §3.3.1)이 분류하는 검색의 기준선이다. RED 가 아니라 목록 완전성 하위 테스트가 Go 코드로 다시 돌릴 검색의 원본이다. L21 은 재개 경로의 실제 이벤트 순서의 근거다(AC-TCL-017 (c)).
- L24·L25 는 오늘 이미 빈 출력이다 — 범위 보호 조건(L24)과 순서 증거 선택자(L25)의 기준선이며 RED 가 아니다. 두 명령이 실제로 무언가를 잡는지는 일부러 만든 저장소에서 확인했다(research.md §10 — 양성 사례 하나, 음성 사례 하나).
- L26 은 RED 가 아니라 검색 모양 두 개(`PillWarn`, `PillErr`, 3회차 감사 M8)를 더한 기준선이다. 보존 줄 목록의 검색은 L20 의 12개 모양과 L26 의 2개 모양을 합친 14개 모양이다. L26 의 9줄 분류는 design.md §3.3.1(X7·X9·X11).
- L27~L31·L33 은 범위 축소 회차의 RED 재측정이다. 바뀐 기준(AC-TCL-001, AC-TCL-010 부터 AC-TCL-014 까지, AC-TCL-016 부터 AC-TCL-020 까지, AC-TCL-024)의 RED 이유가 지금 트리에서도 성립함을 다시 쟀다 — 만들 파일·테스트·함수·플래그·연결이 모두 아직 없다. L28 은 AC-TCL-016·AC-TCL-017 본문에 `69358ecba8` 핀으로 적혀 있던 두 칸을 이 트리에서 다시 잰 것이고, 새 기준 이름 `TestChecklistGateCommandLevel/init_writer_replaced` 의 상위 테스트도 함께 본다. L11(테스트 선택자 대조군)은 이 회차에서 다시 돌리지 않았다(`go test` 를 돌리지 않는 회차 — plan.md §I G-12).
- L32 는 L1~L25 의 옛 핀을 이 트리에도 쓸 수 있는 근거다. 기준 `cb30e05b3e` 이후 바뀐 것은 SPEC 디렉터리뿐이다.
- L34~L41 은 델타 4회차(v0.2.1)의 측정이다. L34 는 따로 세는 직접 출력 줄 네 출처의 init.go 행이고, L35 는 그 줄을 가려내는 다섯 조각의 소스 위치다 — 조각마다 생산 쓰기 지점이 하나(`init.go:43`, `:47`, `:750`, `user_asset_phase.go:54` 의 라벨, `worktree_advisory.go:39`·`:46` 의 두 갈래 문장)이고, 나머지 일치는 주석과 재개 실패 오류(`init_resume.go:108`, 깨끗한 init 에서는 쓰이지 않음)다. L36·L37·L38 은 바뀐 기준(AC-TCL-003, AC-TCL-004, AC-TCL-011, AC-TCL-013, AC-TCL-016, AC-TCL-017)의 RED 재측정이다. L39·L40·L41 은 RED 칸이 아니라 사실 칸이다 — L39 는 probe 카탈로그를 넣을 캐시(design.md §7), L40 은 update `warn`·`fail` 기준선 고정이 타는 경로(AC-TCL-013), L41 은 성공하는 init 에 항목 밖 `Collect` 생산 지점이 없다는 근거(`:837` 실패 경로만 항목 밖, design.md §3.7)와 반례의 출처 `:427` 이다.

## §C 인수 기준

각 기준의 "RED 이유" 는 오늘 왜 빨간지, "녹색 경로" 는 어느 마일스톤이 무엇으로 뒤집는지를 적는다. 뮤턴트는 기준을 통과하면서 요구를 어기는 구현을 시도해 본 결과다. "계속 발화" 는 이 검사가 멈췄을 때 누가 어떻게 알게 되는지다.

### AC-TCL-001 — 깨끗한 실행의 체크리스트 모양 (maps REQ-TCL-001, REQ-TCL-003)

- Given 고정 시계·`LightTheme`·고정 환경 요약과 design.md §4 의 init 계획표, When 모든 항목을 `Done` 으로 끝내고 요약을 그리면, Then 출력이 `internal/tui/testdata/checklist-init-clean.golden` 에 바이트 단위로 같고 머리줄 1개·환경 줄 1개·그룹 제목·항목당 한 줄 구조를 갖는다. (v0.1.3 의 `checklist-update-clean.golden` 은 update 계획표에 묶여 있어 [DEFERRED → t1978]. 렌더러가 계획표를 호출자에게서 받는다는 계약은 init 계획표 밖 키로 그리는 AC-TCL-003·AC-TCL-004 의 probe 계획이 보고(v0.2.1), update 모양 요약은 AC-TCL-005 가 따로 본다.)
- (명령 수준, v0.2.2, 결정 기록 §15.107 — REQ-TCL-003 의 `○` 를 실제 init 이 내는지) Given 강제 켜기 시접으로 켠 활성 init(실제 실행기, en), When (`no_hooks`) `--no-hooks` 로 돌리고 (`mcp_declined`) 비대화형 경로(`opts.MCPProvision` 거짓 — design.md §4 조건부 생략)로 돌리면, Then `no_hooks` 에서 `checklist.item.hooks.prepush`·`checklist.item.hooks.precommit` 의 en 라벨을 담은 항목 줄이 각각 정확히 1번 `○` 로 끝나고 `✓` 로 끝나지 않으며, `Pre-push hook installed`·`Pre-commit hook installed` 는 0번, P5 안내 조각(`init.go:1034`)은 정확히 1번이다. `mcp_declined` 에서 `checklist.item.mcp.entry` 라벨 줄이 `○` 로 끝나고 `✓` 로 끝나지 않으며, `Provisioned the moai MCP server entry` 는 0번이다. 대조로 같은 고정에서 `--no-hooks` 를 빼고 대화형 경로(위저드 시접 `runWizardFn`)로 돌린 `control_not_skipped` 는 세 항목 줄이 모두 `✓` 로 끝난다(양방향).
- 종류: release-blocking. RED 이유: L1·L9·L10 — 렌더러와 테스트가 없다. 이 트리에서 다시 잼: L27(파일 없음, 종료 1), L29(`func NewChecklist` 없음, 종료 1), L33(골든 없음, 종료 1). L11 이 같은 명령의 오늘 출력이며 `--- PASS:` 줄이 없다. 명령 수준 절(트리 `eb2a9dfc21`): 명령 `/usr/bin/grep -rn TestChecklistInitConditionalSkip internal/cli` → stdout 빈 출력, 종료 1 — 테스트가 없다. 건너뛰는 단계가 `nil` 을 돌려준다는 근거는 design.md §4(`hook_install.go:181-183`, `hook_install_precommit.go:297-299`, `init.go:194-196`, `:623`).
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistGoldenClean$' -count=1 -v` → 종료 0, `--- PASS: TestChecklistGoldenClean ` 포함. (M4, 명령 수준) `go test ./internal/cli -run '^TestChecklistInitConditionalSkip$' -count=1 -v` → `--- PASS: TestChecklistInitConditionalSkip/no_hooks `, `--- PASS: TestChecklistInitConditionalSkip/mcp_declined `, `--- PASS: TestChecklistInitConditionalSkip/control_not_skipped `.
- 뮤턴트: 항목마다 두 줄(진행 줄 + 결과 줄)을 남기는 렌더러 → 골든 줄 수가 달라 실패. 골든을 렌더러 출력으로 덮어쓰는 뮤턴트 → AC-TCL-011 줄 예산과 AC-TCL-002 시퀀스 검사가 따로 잡는다. (v0.2.1) init 계획표 모양이나 init 키만 하드코딩해 다른 계획표를 그리지 못하는 렌더러 → 이 기준은 녹색이지만 AC-TCL-003·AC-TCL-004 가 init 계획표 밖의 그룹·항목 키(probe 계획, design.md §7)로 그리므로 probe 골든 불일치로 실패한다(v0.2.0 의 문장은 그 두 기준이 키를 정하지 않아 기계적 차단이 불확실했다 — 확인 감사 5회차 M3). (v0.2.2) 조건부 생략 단계에도 여는 줄만 넣어 다음 `Item` 의 암묵 `Done` 으로 닫는 구현 → 렌더러 골든은 녹색이지만 `no_hooks`·`mcp_declined` 에서 그 항목 줄이 `✓` 로 끝나 실패. 조건과 상관없이 늘 `Skip` 하는 구현 → `control_not_skipped` 의 `✓` 부재로 실패. 렌더러 단위로 `Skip` 만 불러 보는 테스트 뮤턴트 → 이 절은 실제 명령 경로(`--no-hooks` 플래그, 비대화형 `opts.MCPProvision`)를 요구하므로 리뷰에서 걸린다.
- 계속 발화: 기본 `go test ./internal/tui` 실행에 포함. 명령 수준 절은 `internal/cli` 기본 실행. 골든을 새로 만들거나 지우면 `TestManifest` 가 실패한다 — 새 파일은 `internal/tui/golden/index_test.go:73`, 지운 파일은 `:80` (`manifest entry %s is missing on disk`).

### AC-TCL-002 — 실행 중인 줄만 제자리에서 다시 그림 (maps REQ-TCL-002)

- Given 한 항목에 `Update` 를 세 번 부른 뒤 `Done` 하고 다음 항목을 연 렌더러 출력, When 바이트를 훑으면, Then `\r\x1b[2K` 는 정확히 4번(갱신 3번 + 종료 1번) 나오고, 모든 `\n` 바로 다음 위치에는 `\r\x1b[2K` 가 오지 않는다(끝난 줄을 다시 그리지 않는다).
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistRedrawOnlyRunningLine$' -count=1 -v` → `--- PASS: TestChecklistRedrawOnlyRunningLine `.
- 뮤턴트: 화면 전체를 매번 다시 그리는 렌더러 → 시퀀스 수 초과로 실패. 다시 그리기를 전혀 안 하는 렌더러 → 수 부족으로 실패.
- 계속 발화: 패키지 기본 실행에 포함.

### AC-TCL-003 — 실패 시 상세 펼침 (maps REQ-TCL-004)

- Given init 계획표에 없는 그룹·항목 키로 만든 probe 계획(design.md §7 — 그룹 `alpha`·`beta`, 항목 `alpha.one`·`alpha.two`·`beta.one` 처럼 init 표의 어느 키와도 다른 키, 라벨은 시험이 카탈로그 캐시에 넣은 시험 전용 언어 `probe` 의 값)에서 상세 줄 3개를 기록한 뒤 `Fail("boom")` 한 항목과, 상세 줄 2개를 기록하고 `Done` 한 항목, When 그리면, Then 출력이 `checklist-probe-fail.golden` 과 같고, 두 그룹 제목과 두 항목 줄이 probe 라벨로 보이며, 실패 항목의 상세 3줄과 `boom` 은 들여쓴 채 모두 보이고 성공 항목의 상세 2줄은 하나도 보이지 않는다. 테스트가 끝나면 캐시의 `probe` 항목은 지워진다.
- 종류: release-blocking. RED 이유: L1·L10. 이 트리에서 다시 잼: L38(트리 `274441db21`, `checklist_test.go` 와 probe 골든 없음, 종료 1). probe 라벨을 넣을 캐시 변수는 있다(L39, 사실 칸).
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistFailExpandsDetail$' -count=1 -v` → `--- PASS: TestChecklistFailExpandsDetail `.
- 뮤턴트: 실패 메시지만 펼치고 상세를 버리는 렌더러 → 상세 줄 부재로 실패. 모든 항목 상세를 펼치는 렌더러 → 성공 항목 상세 노출로 실패(양방향 검사). (v0.2.1, 확인 감사 5회차 M3) init 계획표를 하드코딩해 호출자가 넘긴 계획 대신 자기 표를 그리는 렌더러, 또는 init 키만 아는 라벨 분기(`switch` 등)로 모르는 키를 빈 라벨로 그리는 렌더러 → probe 그룹·항목 라벨이 없거나 init 그룹이 나와 골든 불일치로 실패. probe 키가 원시 키(`checklist.group.alpha`)로 찍힌 골든 → AC-TCL-023 의 원시 키 검사로 실패.
- 계속 발화: 패키지 기본 실행 + `TestManifest`.

### AC-TCL-004 — 경고만 펼치고 나머지는 접음 (maps REQ-TCL-005)

- Given AC-TCL-003 과 같은 probe 계획(init 계획표 밖 키, design.md §7)에서 경고 하나를 기록한 항목과 경고 없는 항목, When 그리면, Then 경고 항목 줄은 `!` 로 끝나고 그 아래 경고 문장이 보이며, 경고 없는 항목의 상세는 보이지 않는다. 출력은 `checklist-probe-warn.golden` 과 같다.
- 종류: release-blocking. RED 이유: L1·L10. 이 트리에서 다시 잼: L38(트리 `274441db21`, 종료 1).
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistWarnExpandsOnlyWarned$' -count=1 -v` → `--- PASS: TestChecklistWarnExpandsOnlyWarned `.
- 뮤턴트: 경고 수만 요약에 세고 본문을 펼치지 않는 렌더러 → 경고 문장 부재로 실패. (v0.2.1) init 계획표·init 키를 하드코딩한 렌더러 → probe 라벨 부재로 골든 불일치(AC-TCL-003 과 같은 이유).
- 계속 발화: 패키지 기본 실행 + `TestManifest`.

### AC-TCL-005 — 요약 블록 필드 (maps REQ-TCL-006)

- Given 바뀐 수 7, 보존 수 3, 경고 1, 고정 시계로 1.5초 경과, 백업 경로와 복구 명령을 담은 update 요약과, 파일 2·디렉터리 13·배포 모드·다음 단계·항목 밖 경고 문장 1개를 담은 init 요약, When 그리면, Then 각 숫자와 문자열이 요약에 한 번씩 나오고 경과 시간이 `1.5s` 로 보인다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistSummaryFields$' -count=1 -v` → `--- PASS: TestChecklistSummaryFields `.
- 뮤턴트: 백업 경로·복구 명령이나 항목 밖 경고 문장을 빠뜨린 요약 → 문자열 부재로 실패. 시계를 무시하고 `time.Now` 를 쓰는 렌더러 → `1.5s` 불일치로 실패.
- 계속 발화: 패키지 기본 실행.

### AC-TCL-006 — NO_COLOR 에서 SGR 없음 (maps REQ-TCL-007)

- Given `MonochromeTheme`(NO_COLOR 해석 결과)로 만든 렌더러, When 깨끗한 실행·실패·경고를 모두 그리면, Then 출력에 정규식 `\x1b\[[0-9;]*m` 일치가 0건이고 `checklist-init-nocolor.golden` 과 같다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistNoColorHasNoSGR$' -count=1 -v` → `--- PASS: TestChecklistNoColorHasNoSGR `.
- 뮤턴트: 테마를 무시하고 굵게 속성을 직접 넣는 렌더러 → SGR 일치로 실패. 빈 출력을 검사하는 테스트 뮤턴트 → 골든 비교가 막는다.
- 계속 발화: 패키지 기본 실행 + `TestManifest`.

### AC-TCL-007 — 모션 감소에서 점자 프레임 없음 (maps REQ-TCL-008)

- Given `MOAI_REDUCED_MOTION=1`, When 실행 중 항목에 `Update` 를 여러 번 부르면, Then 출력에 U+2800–U+28FF 문자가 하나도 없고 실행 중 줄은 `●` 를 쓴다. 대조로 같은 테스트가 환경 변수를 비우면 `⠋` 가 나온다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistReducedMotion$' -count=1 -v` → `--- PASS: TestChecklistReducedMotion `. 이 테스트는 `t.Setenv` 를 쓰므로 `t.Parallel` 을 쓰지 않는다.
- 뮤턴트: 항상 `●` 만 쓰는 렌더러 → 대조(환경 변수 비움)에서 `⠋` 부재로 실패.
- 계속 발화: 패키지 기본 실행.

### AC-TCL-008 — 네 언어 카탈로그 동치 (maps REQ-TCL-009)

- Given en·ko·ja·zh 카탈로그, When `checklist.` 로 시작하는 키 집합을 비교하면, Then 네 집합이 같고, 값이 모두 비어 있지 않고 키 자신과 다르며, 같은 키의 `%` 서식 지시자가 네 언어에서 인자마다 맞고(지시자를 인자 번호 순으로 늘어놓은 (인자 번호, 동사) 목록이 네 언어에서 같다. 번호 없는 지시자는 나온 순서대로 1, 2, … 번 인자로 치고, 어순 때문에 순서를 바꾸는 언어는 `%[2]s` 처럼 번호를 붙여 같은 목록을 만든다. `%%` 는 지시자가 아니다. 개수만 같은 en `Changed %d` 와 ja `変更 %s件` 은 (1, `d`) 대 (1, `s`) 라 다르다), 그 키의 en 동사에 맞춘 대표 인자(`d` 는 `int` 7, `s` 는 `string` `"x"`, 그 밖의 동사는 테스트 안의 동사별 표)로 네 언어 값을 `fmt.Sprintf` 한 결과에 `%!` 가 없고, 값에 U+1F300–1FAFF·U+2600–27BF·U+1F000–1F2FF 문자가 없으며, ja 값과 zh 값이 같은 키의 en 값과 다르다. 마지막 조건의 예외는 테스트 안의 허용 목록(초기값: 값이 `MCP` 하나뿐인 키)에 이름이 오른 키뿐이다. en 키 수는 design.md §6 표에서 이 카드 범위로 표시된 키 수와 같다(`[DEFERRED → t1978]` 로 표시된 update 전용 키는 넣지 않는다 — t1978 이 네 언어에 함께 더한다).
- 종류: release-blocking. RED 이유: L2·L3·L4·L5 — 네 카탈로그 모두 `checklist.` 키가 0개다.
- 녹색 경로(M1): `go test ./internal/tui -run '^TestChecklistCatalogFourLocales$' -count=1 -v` → `--- PASS: TestChecklistCatalogFourLocales `. 보조 관측: `/usr/bin/grep -c "^checklist\." internal/tui/messages/zh.yaml` → en 과 같은 양수, 종료 0.
- 뮤턴트: ja 에 키를 빠뜨리면 `Translate` 가 en 으로 대체해 화면은 멀쩡해 보인다 → 집합 비교가 잡는다. 값에 키 이름을 그대로 넣는 뮤턴트 → 값 ≠ 키 조건이 잡는다. ja·zh 값을 en 문구 그대로 복사하는 뮤턴트 → ja·zh ≠ en 조건이 잡는다. (v0.2.2, 결정 기록 §15.107) en `Changed %d` 에 ja `変更 %s件` 처럼 지시자 개수는 같고 동사가 다른 값 → v0.2.1 의 개수 비교는 녹색이지만 화면에 `変更 %!s(int=7)件` 이 찍힌다. 이 기준에서는 (인자 번호, 동사) 목록 불일치와 `fmt.Sprintf` 결과의 `%!` 로 실패한다. 사실 칸(트리 `eb2a9dfc21`, `go version go1.26.8 darwin/arm64`): 명령 `go run <스크래치>/fmtprobe/main.go`(두 줄 `fmt.Println(fmt.Sprintf("Changed %d", 7))`, `fmt.Println(fmt.Sprintf("変更 %s件", 7))`) → stdout `Changed 7` / `変更 %!s(int=7)件`, 종료 0. 렌더러가 없는 키를 불러 원시 키가 화면에 남는 뮤턴트 → AC-TCL-023 과 AC-TCL-011 의 골든 원시 키 검사가 잡는다.
- 계속 발화: 패키지 기본 실행. 키를 하나 더하고 한 언어를 잊으면 바로 빨개진다.

### AC-TCL-009 — 한국어 폭 정렬과 미지원 언어 대체 (maps REQ-TCL-009)

- Given `Lang: "ko"` 렌더러와 `Lang: "xx"` 렌더러, When 같은 계획을 그리면, Then ko 출력이 `checklist-init-ko.golden` 과 같고(한글 2칸 폭 기준 정렬), xx 출력은 en 출력과 바이트 단위로 같다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistKoreanAndFallback$' -count=1 -v` → `--- PASS: TestChecklistKoreanAndFallback `.
- 뮤턴트: `len()` 으로 폭을 재는 렌더러 → ko 골든 정렬 불일치로 실패.
- 계속 발화: 패키지 기본 실행 + `TestManifest`.

### AC-TCL-010 — 그룹당 8개 이하 (maps REQ-TCL-010)

- Given `internal/cli` 의 init 계획표(design.md §4), When 그룹별 항목 수와 키를 검사하면, Then 모든 그룹이 1~8개 항목을 갖고, 모든 그룹·항목 키가 en 카탈로그에서 키 자신이 아닌 값으로 풀린다. (update 계획표는 [DEFERRED → t1978].)
- 종류: release-blocking. RED 이유: L7·L10 — 계획표와 세션 코드가 없다. 이 트리에서 다시 잼: L27(`internal/cli/checklist_plan.go` 없음, 종료 1), L31(init 쪽 파일에 `checklist` 없음, 종료 1).
- 녹색 경로(M3): `go test ./internal/cli -run '^TestChecklistPlanGroupCap$' -count=1 -v` → `--- PASS: TestChecklistPlanGroupCap `.
- 뮤턴트: 계획표 대신 호출 지점에 키를 흩어 두는 구현 → 이 검사는 녹색이어도 AC-TCL-011 의 실제 실행 줄 수가 잡는다(표만 보는 한계를 둘째 기준이 메운다).
- 계속 발화: `internal/cli` 기본 실행에 포함(CI 전체 실행).

### AC-TCL-011 — 체크리스트 줄 24 이하, 직접 출력 줄 증가 0 (maps REQ-TCL-010, REQ-TCL-009)

- Given 활성 판정 시접을 강제로 켠, spec.md §3.3 정의대로의 깨끗한 init 실행 — M0 의 대체 골든 `fallback/init-clean-{stdout,stderr}.golden`(AC-TCL-013 `init_clean`)을 찍은 것과 **같은 고정**(임시 홈, `--all` 없음, 실제 사용자 자산 보장, git 있음) — When 두 작성기 출력 각각에 spec.md §3.3 의 줄 단위를 적용하고(덮인 줄 접기, ANSI 제거, `\n` 으로 나눔), 각 줄을 spec.md §3.3 표의 다섯 조각 가운데 하나를 담으면 "직접 출력 줄"(D), 그 밖은 "체크리스트 줄"(C)로 나누면, Then
  - (1) 두 작성기를 합친 C 의 줄 수가 24 이하다.
  - (2) 다섯 조각마다 그 조각을 담은 줄이 리치 출력(두 작성기 합)에 정확히 한 번, 그리고 M0 기준선 골든 두 개(합)에도 정확히 한 번 있다. 그 줄의 앞 공백을 뺀 바이트 문자열이 두 쪽에서 같다(문구 불변 — P5 처럼 항목 아래 들여 펼친 줄도 앞 공백만 다르다). 즉 D 는 기준선 대비 증가 0, 감소 0 이다.
  - (3) 접은 결과가 `internal/cli/testdata/checklist/rich/init-clean.golden` 과 같고, `rich/` 아래 모든 골든에 부분 문자열 `checklist.` 가 없다.
  - (깨끗한 update 실행·`rich/update-clean.golden`·update 줄 예산은 [DEFERRED → t1978] — 델타 4회차 감사 N1, design.md §4.1. 화면 전체 줄 상한은 이 기준이 판정하지 않는다 — 운영자 질문, spec.md §5.)
- 종류: release-blocking. RED 이유: L8·L10. 이 트리에서 다시 잼: L27(`internal/cli/testdata/checklist` 없음, 종료 1), L36(트리 `274441db21` — 리치 골든과 두 기준선 골든 모두 없음, 종료 1), L37(`TestChecklistScreenBudget` 없음, 종료 1). 빨간 이유는 "체크리스트 연결과 골든이 아직 없다" 이다. 직접 출력 줄의 출처(L34)와 조각의 유일한 생산 지점(L35)은 이 트리에서 읽었다 — 그 다섯 줄이 깨끗한 init 에 정확히 한 번씩 나온다는 판단은 코드 읽기이며 M0 기준선이 처음 실측한다(M0 기준선에서 어느 조각이라도 0번이나 2번 이상이면 run 단계는 막힘 보고를 내고 manager-spec 이 §3.3 표를 고친다).
- 녹색 경로(M4): `go test ./internal/cli -run '^TestChecklistScreenBudget$' -count=1 -v` → `--- PASS: TestChecklistScreenBudget/init `, `--- PASS: TestChecklistScreenBudget/direct_lines_unchanged `, `--- PASS: TestChecklistScreenBudget/no_raw_keys `.
- 뮤턴트: 넘김 규칙을 잘못 걸어 항목 안 상세가 새는 구현 → 새는 줄은 조각이 없어 C 에 들어가 (1) 초과, 그리고 골든 불일치로 실패. (v0.2.1, 확인 감사 5회차 B1) 예산을 맞추려고 직접 출력 줄 하나(예: slim 안내)를 체크리스트가 켜졌을 때 숨기는 구현 → (2) 의 조각 0번으로 실패(감소도 막는다). 직접 출력 줄을 항목 상세로 접어 체크리스트 줄처럼 보이게 하는 구현 → 그 조각이 0번이라 (2) 실패. 직접 출력 줄의 문구를 줄여 바꾸는 구현 → (2) 의 바이트 비교로 실패. `init.go:750` 을 항목 안으로 옮기는 구현 → AC-TCL-022 의 영역 검사로 실패(R1·R2·R3 밖). 체크리스트 셈에 직접 출력 줄까지 넣는 측정(v0.2.0 의 셈) → 설계를 충실히 따른 구현이 28줄로 빨개지는, 이 기준이 고친 결함 그 자체다. 다섯 조각 밖의 넘김 줄을 D 로 몰래 분류해 C 를 줄이는 테스트 뮤턴트 → 조각 목록은 spec.md §3.3 표에 고정이라 표 밖 조각은 리뷰에서 걸리고, 그 줄은 기준선에 없으므로 (2) 의 "기준선에 정확히 한 번" 이 빨갛다. 한 작성기(stderr)만 세는 측정 → 깨끗한 init 의 stdout 에는 직접 출력 줄 넷(slim 2, 자산 요약 1, P5 1)이 있어 (2) 의 조각이 0번이 되어 실패(v0.2.0 에서 남은 위험으로 적었던 경우가 이제 기계적으로 막힌다). 카탈로그에 없는 키를 부르는 구현 → 원시 키 검사로 실패. `update` 하위 테스트를 남겨 둔 채 건너뛰는 테스트 → 녹색 경로가 `/update ` 를 요구하지 않으므로 판정에 영향이 없고, 건너뛴 하위 테스트는 `--- SKIP` 으로 남아 리뷰에서 보인다 — 이 카드의 테스트에는 update 하위 테스트를 두지 않는다.
- 계속 발화: `internal/cli` 기본 실행. 골든이 없으면 `t.Fatalf` 로 실패하고 건너뛰지 않는다(기존 `tuxReadGolden` 과 같은 방식, `internal/cli/tuxiu_characterization_test.go:115-122`).

### AC-TCL-012 — 활성 판정과 입력 수집 (maps REQ-TCL-011)

- Given design.md §3.1 의 다섯 조건과 그 입력을 모으는 한 함수 `checklistEnv`, When (1) 판정 단위로 조건마다 하나씩만 거짓으로 바꾼 경우와 모두 참인 경우를 판정하고, (2) 명령 수준에서 TTY 탐지 시접과 프로세스 스트림 시접(design.md §3.1 — 강제 켜기 시접은 쓰지 않는다)을 써 "stdout 비TTY·stderr TTY", "stdout TTY·stderr 비TTY", `CI=1`, `TERM=dumb`, 그리고 "작성기 교체"(프로세스 스트림 시접이 명령에 끼운 버퍼와 **다른** 버퍼를 돌려줘 조건 2 만 거짓, 3회차 감사 M7) 각각으로 init 을 돌리면, Then (1) 모두 참일 때만 활성이고, (2) 다섯 경우 모두 두 작성기 출력이 AC-TCL-013 의 init 깨끗한 실행 대체 골든과 바이트 단위로 같고 어느 작성기에도 `\x1b[` 가 없으며, 대조로 같은 시접에서 모든 조건이 참이면 체크리스트 머리줄이 나온다. "stdout 비TTY" 경우는 터미널에서 `moai init > log` 를 한 상황이다. (update 명령 수준 하위 테스트 `update_*` 는 [DEFERRED → t1978].)
- 종류: release-blocking. RED 이유: L7·L10·L18 — 판정 함수, 입력 수집 함수, 세션 코드가 없다. 이 트리에서 다시 잼: L29(`func checklistEnv` 없음, 종료 1), L28(`TestChecklistGateCommandLevel` 없음, 종료 1).
- 녹색 경로(M3 판정 단위, M4 명령 수준): `go test ./internal/cli -run '^TestChecklistActivationPredicate$' -count=1 -v` → `--- PASS: TestChecklistActivationPredicate `, `--- PASS: TestChecklistActivationPredicate/all_true `, `--- PASS: TestChecklistActivationPredicate/stdout_not_tty `, `--- PASS: TestChecklistActivationPredicate/stderr_not_tty `, `--- PASS: TestChecklistActivationPredicate/writer_replaced `, `--- PASS: TestChecklistActivationPredicate/ci_set `, `--- PASS: TestChecklistActivationPredicate/term_dumb `, `--- PASS: TestChecklistActivationPredicate/verbose_on `. 그리고 `go test ./internal/cli -run '^TestChecklistGateCommandLevel$' -count=1 -v` → `--- PASS: TestChecklistGateCommandLevel/init_stdout_not_tty `, `--- PASS: TestChecklistGateCommandLevel/init_stderr_not_tty `, `--- PASS: TestChecklistGateCommandLevel/init_ci_set `, `--- PASS: TestChecklistGateCommandLevel/init_term_dumb `, `--- PASS: TestChecklistGateCommandLevel/init_writer_replaced `, `--- PASS: TestChecklistGateCommandLevel/control_all_true `.
- 뮤턴트: 조건 하나(예: `TERM=dumb`)를 빼먹은 판정 → 해당 하위 테스트 실패. 늘 비활성인 판정 → `all_true`·`control_all_true` 실패(양방향). 호출 지점이 stderr 판정에 stdout 의 TTY 결과를 넘기는 구현 → `init_stderr_not_tty` 가 체크리스트를 그려 골든 불일치로 실패. `checklistEnv` 가 두 조건을 모두 stderr 작성기의 TTY 결과로 정하는 구현(2회차 감사 N4) → `init_stdout_not_tty` 에서 stderr 는 TTY 라 체크리스트가 켜져 stderr 버퍼에 머리줄과 `\x1b[` 가 들어가 실패. `checklistEnv` 가 조건 2 를 늘 참으로 모으는 구현(3회차 감사 M7) → `init_writer_replaced` 에서 체크리스트가 켜져 실패 — 판정 단위의 `writer_replaced` 는 입력을 직접 받으므로 이 뮤턴트를 잡지 못하고, 명령 수준 하위 테스트가 그 몫을 진다. 판정 단위의 `stdout_not_tty` 도 같은 이유로 수집 함수 뮤턴트를 잡지 못한다.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-013 — 대체 경로는 기준선과 같은 바이트, 기준선이 먼저 (maps REQ-TCL-012)

- Given M0 이 코드 변경 **전에**, 그 둘만 담은 커밋 하나로 넣은 `internal/cli/checklist_fallback_test.go` 와 `internal/cli/testdata/checklist/fallback/` 아래 골든 — init·update 각각 시나리오 세 개(깨끗한 실행 `clean`, 경고 줄 1건 이상 `warn` — init 은 수집기 경고, update 는 경고 줄, 실패 1건 `fail`)의 stdout·stderr 골든 12개, When 버퍼 작성기(비TTY)로 같은 고정 실행을 돌리면, Then 출력이 아래 정규화 뒤 골든과 바이트 단위로 같고, 정규화가 줄 수를 바꾸지 않으며, `warn` 골든에는 주입한 경고 문장이, `fail` 골든에는 주입한 실패 문장이 있다.
- (v0.2.0) `update_*` 세 하위 테스트의 뜻이 바뀌었다. 이 카드는 update 를 연결하지 않으므로 그 셋은 "update 의 대체 경로" 가 아니라 **update 출력이 이 카드의 공유 파일 변경(spec.md §4 C-7)에도 기준선 그대로** 라는 회귀 경계다(REQ-TCL-012 의 update 절). 같은 골든은 t1978 의 update 연결 전 기준선이 된다. 회귀 경계가 공유 파일의 줄을 실제로 담도록, update 의 `clean`·`warn` 고정 실행은 `--force --yes` 로 돌아 템플릿 동기화 본문을 끝까지 탄다 — `--force` 가 없으면 같은 바이너리로 init 한 고정 프로젝트는 `internal/cli/update_template_sync.go:1260` 의 같은 버전 조기 반환을 타서 훅 설치·스냅숏 줄이 골든에 들어가지 않는다(델타 4회차 감사 N1 (가)의 관찰). `--force` 가 출력에 더하는 줄(백업·확인 경로)은 골든에 그대로 담기며, 그 모양은 이 회차에 읽지 않았다(plan.md §I G-13). v0.1.3 의 `fallback/update-verbose-{stdout,stderr}.golden` 두 개는 update `--verbose` 기준(AC-TCL-014 의 update 절)에 묶여 있어 [DEFERRED → t1978] — M0 의 골든은 12개다.
- (v0.2.1, 확인 감사 5회차 S1) 여섯 고정 실행의 주입 방법. M0 은 시험 파일과 골든만 담으므로(O2), 모든 주입은 지금 코드에 있는 시험 경로로 한다 — 새 생산 시접은 없다. 그래서 update 고정을 "문서화된 공백" 으로 낮추지 않고 이름으로 정한다.
  - `init_clean`: spec.md §3.3 의 깨끗한 init(AC-TCL-011 과 같은 고정).
  - `init_warn`: 기존 시접 `applyAutonomyTierBundleFn`(`internal/cli/update_settings_snapshot.go:32`)이 오류를 돌려줌 → `init.go:935` 수집기 경고 `Failed to apply autonomy tier bundle` (주입 문장 양성 대조).
  - `init_fail`: 기존 시접 `newInitPhaseExecutorFn`(`internal/cli/init_executor_seam.go:29`)이 `template deployment` 오류를 내는 실행기를 돌려줌 → `init.go:876-877` (주입 문장 `template deployment` 양성 대조).
  - `update_clean`: 같은 바이너리로 init 하고 `git init` 한 고정 프로젝트에서 `--force --yes`.
  - `update_warn`: 같은 고정에서 init 이 설치한 `.git/hooks/pre-push` 의 끝에 한 줄을 덧붙여(앞 세 줄의 MoAI 표식은 그대로) "사용자가 고친 MoAI 훅" 으로 만든 뒤 `--force --yes` → `hook_install.go:214-224` 가 백업을 만들고 `:268` 이 경고 작성기(stderr)에 `user-modified pre-push hook was replaced` 를 쓴다(원장 L40). 주입 문장 양성 대조는 이 문장이다. 분류 함수(`classifyHook`)가 덧붙인 훅을 "사용자 수정" 으로 판정한다는 것은 `hook_install.go:202-225` 의 코드 읽기이며, M0 이 이 고정에서 그 문장을 얻지 못하면 양성 대조가 빨개지므로 M0 커밋 전에 드러난다 — 그때는 고정이 아니라 이 문장을 manager-spec 이 고친다(막힘 보고).
  - `update_fail`: 같은 고정에서 `--force` 만 주고(`--yes` 없음 — 확인 단계가 돈다) 기존 시접 `confirmViaPreviewFn`(`internal/cli/update_model_key_strip.go:33`)이 오류를 돌려줌 → `update_template_sync.go:1287-1290` 의 `confirm merge for` 오류로 명령이 실패한다(원장 L40). 주입 문장 양성 대조는 `confirm merge for` 다. 이 실패는 템플릿 동기화 본문보다 앞이라 공유 파일의 실패 경로 줄을 타지 않는다.
- (v0.2.1, 확인 감사 5회차 S2) 이 기준의 update 회귀 경계는 위 세 고정 실행이 실제로 닿는 줄에 한한다. 공유 파일의 실패 경로 줄 가운데 세 실행이 닿지 않는 것(`hook_install.go:283`·`:291`·`:293`, `update/backup/settings_snapshot.go:115` `warnLine` 등)은 이 SPEC 이 지키지 않는다 — REQ-TCL-012 의 update 절이 세 고정 실행으로 좁혀진 이유이며, 넓히는 일과 update 연결 뒤의 그 절 개정은 t1978 이다(spec.md §3.4 F-10, plan.md §K K-16).
- 정규화 규칙(테스트가 이 순서로 적용하는 정규식 전부. 이 밖의 치환은 없다):
  - 프로젝트 임시 경로: 테스트가 만든 `t.TempDir()` 절대 경로 문자열 → `<ROOT>` (정규식이 아닌 고정 문자열 치환)
  - 백업 디렉터리 시각: `\b\d{8}_\d{6}\b` → `<TS>` (배치 형식 `20060102_150405`, `internal/cli/update/backup/retention_realdata_test.go:53`)
  - 훅 백업 파일 시각: `\b\d{8}T\d{6}Z\b` → `<TS>` (훅 백업 형식 `20060102T150405Z`, `internal/cli/hook_install_shared.go:176` `now.UTC().Format("20060102T150405Z")`. `update_warn` stderr 의 `hook_install.go:268` 경고 줄이 담는 백업 경로 `pre-push.bak.<시각>` 을 지운다. 바로 위 밑줄 형식 규칙은 이 형식을 잡지 못한다 — 6회차 감사 N-A, 결정 기록 §15.104)
    - 사실 칸(트리 `9f7ae0b16f`): 명령 `python3 -I <스크래치>/ts_probe.py` — 이 정규식을 일부러 만든 다섯 줄(양성 2: 백업 시각, 같은 시각 뒤 `.1` 형제 접미사 / 음성 3: 시각 없음, 9자리 날짜, 7자리 날짜)에 적용. 종료 코드 0. stdout 그대로:

      ```text
      pos changed previous hook backed up at <ROOT>/.git/hooks/pre-push.bak.<TS>
      pos-suffix changed previous hook backed up at <ROOT>/.git/hooks/pre-push.bak.<TS>.1
      neg-none unchanged user-modified pre-push hook was replaced
      neg-9digit unchanged pre-push.bak.202601011T120000Z
      neg-7digit unchanged pre-push.bak.2026010T120000Z
      ```

      확인 도구는 Python `re` 이다. 이 정규식이 쓰는 `\b`·`\d`·`{n}` 은 ASCII 입력에서 Go RE2 와 같은 뜻이지만, Go 로 같은 확인을 돌리지는 않았다 — M0 의 `update_warn` 녹색이 그 확인이다.
    - 뮤턴트: 이 치환 없이 찍은 `update_warn` stderr 골든 → 다음 실행(O4b 의 M0 트리 재실행 포함)에서 백업 시각이 바뀌어 골든 불일치로 빨강.
  - 하이픈 ISO 시각: `\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z\b` → `<ISO_TS>` (예 `2026-10-26T22:09:39Z`. 위 두 `<TS>` 규칙은 이 형식을 잡지 못하고, 자리표를 `<TS>` 와 나눠 어느 규칙이 지웠는지 골든에서 구별된다 — v0.2.2, 결정 기록 §15.107). 사실 칸(트리 `eb2a9dfc21`): 명령 `python3 -I <스크래치>/iso_probe.py` → 종료 0, stdout `pos changed recorded_at <ISO_TS> done` / `pos-eol changed at=<ISO_TS>` / `neg-hookbak unchanged pre-push.bak.20261026T220939Z` / `neg-nozone unchanged at 2026-10-26T22:09:39 local` / `neg-batch unchanged backup 20261026_220939`. 이 형식이 여섯 고정 실행의 출력에 실제로 나오는지는 읽지 않았다 — 나오지 않으면 이 규칙은 아무것도 바꾸지 않는다.
  - 버전: `\bv\d+\.\d+\.\d+(-[0-9A-Za-z.]+)?\b` → `<VER>`
  - Go 런타임: `\bgo1\.\d+(\.\d+)?\b` → `<GO>` (update 정체성 띠의 `runtime.Version()`, `internal/cli/update_tux.go:80`)
  - 경과 시간: `\b\d+(\.\d+)?(ms|s)\b` → `<DUR>`
  - 목록 밖 값: 위 규칙 어느 것에도 맞지 않는데 실행마다 바뀌는 값이 골든 비교에서 드러나면, 테스트는 그 값을 이름으로 적은 막힘 메시지(불일치 줄의 기대·실제 쌍 그대로)로 실패한다 — 조용히 정규화하거나 무시하지 않으며, run 단계는 막힘 보고를 내고 manager-spec 이 이 목록에 규칙을 더한다.
- 종류: release-blocking. RED 이유: L8·L19 — 기준선 골든 디렉터리가 없다. 이 트리에서 다시 잼: L33(`fallback` 없음, 종료 1), L27(`checklist_fallback_test.go` 없음, 종료 1), L36(트리 `274441db21`, `fallback/init-clean-*` 없음, 종료 1). update 주입 경로는 L40(사실 칸)에서 읽었다. 이 기준은 기준선 커밋이 구현 커밋보다 앞서야 성립한다(design.md §8).
- 녹색 경로(M0 에서 녹색으로 들어오고, M4·M6 뒤에도 녹색이어야 함): `go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` → `--- PASS: TestChecklistFallbackMatchesBaseline/init_clean `, `--- PASS: TestChecklistFallbackMatchesBaseline/init_warn `, `--- PASS: TestChecklistFallbackMatchesBaseline/init_fail `, `--- PASS: TestChecklistFallbackMatchesBaseline/update_clean `, `--- PASS: TestChecklistFallbackMatchesBaseline/update_warn `, `--- PASS: TestChecklistFallbackMatchesBaseline/update_fail `.
- 순서 증거(O4a 만 M0 직후, 나머지는 M6 에서 각각 한 번의 호출로 돌리고 progress.md §E.2 에 그대로 기록. O1~O4 와 §A 의 카드 id 완전성 검사가 모두 성립해야 녹색). `<기준선 트리>` 는 `cb30e05b3e`, 재촬영 경로에서는 재촬영 커밋의 부모인 흡수 커밋이다(plan.md §B 1(c)).
  - O1 `git log --format=%H cb30e05b3e..HEAD -- internal/cli/testdata/checklist/fallback internal/cli/checklist_fallback_test.go` → 정확히 한 줄, 그 SHA 가 M0 이다. 둘째 줄이 있으면(구현 뒤 골든 수정·재촬영·테스트 정규화 변경) 빨강. 재촬영 경로에서는 정확히 두 줄(재촬영 커밋과 M0).
  - O2 `git show --name-only --format= <M0>` → `internal/cli/checklist_fallback_test.go` 와 `internal/cli/testdata/checklist/fallback/` 아래 경로만. 다른 경로가 하나라도 있으면(코드 변경 혼입) 빨강.
  - O3 `git log --no-merges --grep=t1969 --format=%h cb30e05b3e..HEAD ^main -- internal ":(exclude)internal/cli/testdata" ":(exclude)internal/cli/checklist_fallback_test.go"` 가 내는 SHA 각각에 대해 `git merge-base --is-ancestor <M0> <SHA>` → 종료 0. 하나라도 종료 1 이면(M0 보다 먼저 들어간 코드 변경) 빨강. 범위에서 M0 이 담는 두 경로를 빼므로 M0 자신은 이 목록에 나오지 않는다 — 1회차 범위(`internal` 에서 testdata 만 뺌)는 M0 의 테스트 파일 때문에 M0 을 냈고 "M0 과 다르다" 조건과 부딪혔다(2회차 감사 N8, 리더 codex_audit P2). 고친 명령은 일부러 만든 저장소에서 양성 사례(M0 뒤 코드 커밋 세 개 → 세 SHA, 모두 종료 0, M0 없음)와 음성 사례(M0 앞 코드 커밋 → 그 SHA, 종료 1)를 의도대로 냈다(research.md §10). 오늘 이 트리에서는 빈 출력(원장 L25). 재촬영 경로에서는 범위를 `internal/cli`(위 두 경로와 `internal/cli/testdata` 제외)로, 조상 기준을 재촬영 커밋으로 바꾼다 — M1·M2 의 `internal/tui` 커밋은 재촬영 커밋보다 앞서도 CLI 출력에 닿지 않기 때문이다.
  - O4 (2회차 감사 N6). 커밋 그래프는 순서를 증언하지만 골든의 출처는 증언하지 못한다(`verification-claim-integrity.md` §2.3). 그래서 "골든을 변경 전 바이너리가 찍었다"를 주장하지 않고, 그래프와 재실행이 확인해 줄 수 있는 사실로 바꿔 잰다 — M0 트리의 코드가 기준선 코드와 같고, 그 트리에서 대체 테스트가 녹색이다. 그러면 골든은 출처와 상관없이 변경 전 코드의 출력과 같다.
    - O4a (M0 커밋 직후, 다음 커밋 전, 레인 기록): `git rev-parse HEAD` 와 `go test ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` 를 각각 한 번 돌려 progress.md §E.2 에 그대로 남긴다 — 여섯 `--- PASS:` 줄과 측정 HEAD(40자). M6 에서 그 측정 HEAD 가 O1 이 내는 SHA 와 같아야 한다. 이력을 고치면 M0 의 SHA 가 바뀌어 빨강이다. 이 기록은 레인의 주장이며 그래프가 증언하지 않는다 — 그래서 O4b 가 따로 있다.
    - O4b (sync 감사가 다시 잼, 판정의 증거): (i) `git diff --name-only <기준선 트리> <M0>~1 -- internal pkg go.mod go.sum` → 빈 출력(M0 앞까지 코드가 기준선과 같음). 범위에 `pkg go.mod go.sum` 이 있는 이유(3회차 감사 M6): `internal/cli` 가 `pkg/` 를 가져다 쓰므로(`init.go`, `update.go`, `update_template_sync.go`) M0 앞의 `pkg/`·모듈 파일 변경도 "M0 트리 코드 = 기준선 코드" 전제를 깬다. (ii) `git archive --output=<스크래치>/m0.tar <M0>` → `mkdir -p <스크래치>/m0` → `tar -xf <스크래치>/m0.tar -C <스크래치>/m0` → `go test -C <스크래치>/m0 ./internal/cli -run '^TestChecklistFallbackMatchesBaseline$' -count=1 -v` → 여섯 `--- PASS:` 줄. 이 저장소의 작업 트리는 건드리지 않는다. 그래서 M0 의 대체 테스트는 git 이력을 읽지 않아야 한다(뽑아낸 트리는 저장소가 아니다).
- 뮤턴트(v0.2.0 추가): M0 앞에 `pkg/` 나 `go.mod` 만 바꾼 커밋을 넣는 절차 → 넓힌 O4b (i) 가 그 경로를 내 빨강(3회차 범위 `-- internal` 은 빈 출력이라 녹색이었다). init 연결을 위해 `hook_install.go` 를 고치면서 세션이 없을 때 update 의 훅 설치 줄 모양을 바꾼 구현 → `--force --yes` 고정 실행이 훅 설치 꼬리까지 돌므로 `update_clean` 하위 테스트가 골든 불일치로 빨강(C-7 회귀 경계). update 고정 실행을 `--force` 없이 찍어 조기 반환만 담는 절차 → 이 기준 문장(`--force --yes`)을 어겨 리뷰에서 걸린다. 기계적으로는 골든에 `Already up-to-date` 결과 알약만 있고 훅 줄이 없는 것으로 드러나므로, 녹색 판정에 "update_clean stdout 골든에 훅 설치 줄 `Pre-push hook installed`(`internal/cli/hook_install.go:262`, `out` 작성기)이 정확히 1번" 단언을 함께 둔다. 고정 프로젝트는 `git init` 한 저장소여야 훅 설치가 돈다 — 저장소가 아닐 때의 설치기 동작은 이 회차에 읽지 않았다(plan.md §I G-13). update 대체 골든을 M0 에서 빼고 init 만 찍는 절차 → 녹색 경로의 `update_*` 세 줄이 없어 빨강. (v0.2.1) `update_warn`·`update_fail` 을 주입 없이 `update_clean` 과 같은 실행으로 찍는 절차 → 주입 문장(`user-modified pre-push hook was replaced`, `confirm merge for`) 양성 대조가 빨강. 경고·실패 고정을 위해 M0 에 생산 시접을 더하는 절차 → O2(M0 은 테스트·골든 경로만)로 빨강.
- 뮤턴트: 기준선을 구현 뒤에 다시 찍어 수정 커밋으로 넣는 절차 → O1 이 두 줄. M0 에 `init.go` 를 섞는 절차 → O2. CLI 쪽 변경(`init_warnings.go` 등)을 M0 보다 먼저 커밋하는 절차 → O3. id 없이 그런 커밋을 넣는 절차 → §A 카드 id 완전성 검사. 구현 뒤 다시 찍은 골든을 리베이스로 M0 에 합치는 절차(2회차 감사 N6 (가)) → O4a 의 측정 HEAD ≠ O1 SHA, 그리고 바뀐 대체 출력을 담은 골든은 O4b 의 M0 트리(기준선 코드)에서 빨강. M0 에 새 동작을 손으로 적은 골든을 넣는 절차(N6 (나)) → O4b 에서 빨강, O4a 기록에도 `--- FAIL:` 이 남는다. 경고 줄을 지우는 넓은 정규화 → 줄 수 불변 조건과 `warn` 골든의 주입 문장 조건이 잡는다.
- 계속 발화: 골든 비교는 `internal/cli` 기본 실행, 골든 부재는 `t.Fatalf`. 순서 증거 O1~O4 는 이 카드 범위의 일회성 검사이며 sync 감사가 다시 돈다(O4b 는 sync 감사만 돈다). `main` 흡수로 기준선이 바뀌어야 하면 plan.md §B 의 재촬영 절차가 이 기준을 다시 쓴다.

### AC-TCL-014 — `--verbose` (maps REQ-TCL-013)

- Given init, When TTY 탐지 시접과 프로세스 스트림 시접으로 활성 조건 다섯 가운데 `--verbose` 를 뺀 네 조건을 참으로 두고(강제 켜기 시접은 쓰지 않는다 — 그 시접은 판정을 통째로 참으로 만들어 `--verbose` 도 끌 수 없다, design.md §3.1) `--verbose` 를 주어 돌리면, Then 체크리스트가 켜지지 않고 init 출력이 `fallback/init-clean-stderr.golden`·`init-clean-stdout.golden` 과 같다. 그리고 같은 시접 상태에서 `--verbose` 없이 돌린 대조 실행은 체크리스트 머리줄을 낸다(같은 고정 실행의 두 출력만 비교한다). init 의 `--help` 출력에 `--verbose` 가 있다. (update `--verbose` 가 대체 경로를 강제하는 절과 그 골든 `fallback/update-verbose-{stdout,stderr}.golden`, 하위 테스트 `update`·`update_verbose_differs` 는 [DEFERRED → t1978].)
- 종류: release-blocking. RED 이유: L6 — init 에 `--verbose` 가 없다. L8·L19 — 기준선이 없다. 이 트리에서 다시 잼: L30(`"verbose"` 가 init.go 에 없음, 종료 1), L33(`fallback` 없음, 종료 1).
- 녹색 경로(M3·M4): `go test ./internal/cli -run '^TestChecklistVerboseForcesFallback$' -count=1 -v` → `--- PASS: TestChecklistVerboseForcesFallback/init `, `--- PASS: TestChecklistVerboseForcesFallback/init_control_without_verbose `. 보조: `/usr/bin/grep -rn "Flags().Bool(\"verbose\"" internal/cli` 결과에 update 등록(`update.go:89`)과 init 등록 두 줄, 종료 0.
- 뮤턴트: init 에 플래그만 등록하고 판정에 쓰지 않는 구현 → 나머지 네 조건이 참이라 체크리스트가 켜져 실패. 판정에서 `--verbose` 를 무시하고 강제 켜기 시접으로 테스트를 녹색으로 만드는 테스트 뮤턴트 → 이 기준이 그 시접을 금지하므로 리뷰에서 걸리고, 그 시접 아래에서는 "켜지지 않는다" 단언 자체가 빨갛다. (v0.2.0) 시접이 네 조건 가운데 하나를 실수로 거짓으로 두어 `--verbose` 와 상관없이 늘 대체 경로가 되는 테스트 뮤턴트 → `init_control_without_verbose` 가 머리줄을 요구하므로 빨강(양방향). 다른 픽스처와 비교해 픽스처 차이로 녹색이 되는 테스트(2회차 감사 N13) → 이 기준은 같은 고정 실행의 두 출력만 비교하므로 성립하지 않는다.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-015 — 항목 안은 접고, 밖은 넘김 (maps REQ-TCL-014)

- Given 활성 세션, When 항목이 열린 동안 `cmd.ErrOrStderr()` 에 `inside` 를 쓰고 항목이 닫힌 뒤 `outside` 를 쓰면, Then 실제 작성기에 `outside` 는 쓴 즉시 바이트 그대로 나오고 `inside` 는 성공 종료 뒤에도 나오지 않는다. 같은 항목을 `Fail` 로 닫은 경우에는 `inside` 가 펼쳐진다.
- 종류: release-blocking. RED 이유: L10.
- 녹색 경로(M3): `go test ./internal/cli -run '^TestChecklistCapturePassThrough$' -count=1 -v` → `--- PASS: TestChecklistCapturePassThrough `.
- 뮤턴트: 모든 바이트를 버퍼에 모았다가 끝에 쏟는 구현 → "쓴 즉시" 검사(쓴 직후 실제 작성기 길이 확인)로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-016 — 경고는 항목 아래 한 번, 항목 밖 경고는 요약에 (maps REQ-TCL-015)

- (하위 테스트 `init`) Given 활성 init 실행 중 수집기 `Warn("w-in")`(항목 안), `Warn("w-out")`(항목 밖), `Collect("c-in")`(항목 안), `Collect("c-out")`(항목 밖), `emitSeverityLine(sevWarn, "s-in")`(항목 안), When 실행이 끝나면, Then 실제 작성기 출력에서 `w-in`·`c-in`·`s-in` 은 각각 정확히 1번(항목 아래 펼침), `w-out` 은 정확히 2번(기존과 같은 즉시 줄 + 요약 나열), `c-out` 은 정확히 1번(요약 나열)이고, `warning(s) during init:` 패널 머리 문자열은 0번이며, 요약의 경고 수는 4 — 수집기 수(결정 D-9, `internal/cli/init_warnings.go:59-63`)이며 `s-in` 은 세지 않는다. 대조로 같은 실행을 비활성으로 돌리면 패널 머리가 1번 나오고 기존 완료 카드의 경고 수도 4다.
- (주입 방법, v0.2.1, 확인 감사 5회차 S3 — design.md §3.7) 실제 init(실제 실행기, 실제 배포)에서 주입한다. 항목 밖 `Warn("w-out")`·`Collect("c-out")` 은 새 시험 전용 시접 `checklistSessionStartFn`(M3 이 `internal/cli/checklist_session.go` 에 두는 패키지 변수, 생산 값 `nil`, 세션 시작 때 수집기와 `cmd` 를 넘겨 한 번 부름)에서 기록한다 — 그때 열린 항목이 없다. 지금 코드에는 성공하는 init 이 항목 밖에서 `Collect` 를 부르는 지점이 없기 때문이다(원장 L41: `p.Collect` 는 `init.go:837` 실패 경로, `:954`·`:1047`·`:1074` 항목 안뿐). 항목 안 `Warn("w-in")`·`Collect("c-in")`·`emitSeverityLine(cmd.ErrOrStderr(), sevWarn, …, "s-in")` 은 기존 시접 `newInitPhaseExecutorFn`(`internal/cli/init_executor_seam.go:29`)을 감싼 실행기가 — 안에 실제 실행기를 두고 — 보고기의 `StepStart("Initialization", …)` 를 넘긴 직후(`templates.deploy` 가 열린 상태) 위 시접으로 받아 둔 수집기·`cmd` 로 부른다. 새 시접은 세션 코드 안에서 불리므로 init.go 에 R1 의 세션 시작 줄 밖의 새 줄을 요구하지 않는다(REQ-TCL-021).
- (하위 테스트 `init_stdout_severity`, v0.2.0) Given 활성 init 실행에서 항목이 열린 동안 심각도 줄 도우미로 `emitSeverityLine(cmd.OutOrStdout(), sevErr, …, "s-out")` 를 한 번 주입하고, When 실행이 끝나면, Then 두 실제 작성기 출력을 합쳐 `s-out` 이 정확히 1번 그 항목 줄 아래 펼쳐져 나오고 그 항목 줄은 `!` 로 끝난다. 근거: REQ-TCL-015 는 init 의 두 명령 작성기 모두를 말하고, init 도 심각도 줄 도우미에 stdout 작성기를 넘기는 경로가 있다(`init.go:1034` → `internal/cli/worktree_advisory.go:57` 의 `emitSeverityLine(out, sevNote, …)` — 보존 행 P5. 그 지점은 안내 심각도라 경고 신호가 아니므로, 이 하위 테스트는 같은 작성기에 `sevErr` 를 주입한다. 주입 방법은 위 문단의 감싼 실행기이며, `StepStart("Initialization", …)` 직후 `cmd.OutOrStdout()` 에 쓴다). 3회차 감사 M1 의 "stderr 세션 작성기에만 수신 인터페이스" 뮤턴트를 init 범위에서 막는 하위 테스트다.
- [DEFERRED → t1978] (v0.1.3 의 하위 테스트 `update`, 3회차 감사 M1 — 아래 두 문단은 t1978 이 쓸 근거로 남기며 이 SPEC 의 판정에 넣지 않는다.) Given 강제 켜기 시접으로 켠 활성 update 실행에서 두 명령 작성기가 모두 세션 작성기이고, 단계 루프 뒤 꼬리 — 결정 D-10 에 따라 마지막으로 열린 항목이 아직 열려 있는 상태 — 에서 심각도 줄 도우미로 `emitSeverityLine(cmd.ErrOrStderr(), sevWarn, …, "u-warn")` 와 `emitSeverityLine(cmd.OutOrStdout(), sevErr, …, "u-err")` 를 각각 한 번 주입하고, When 실행이 끝나면, Then 두 실제 작성기 출력을 합쳐 `u-warn`·`u-err` 가 각각 정확히 1번 그 항목 줄 아래 펼쳐져 나오고, 그 항목 줄은 `!` 로 끝나며, 출력 전체에 `✗` 는 0번이다(항목은 성공으로 닫힌다).
- [DEFERRED → t1978] update 쪽 근거 지점(트리 `69358ecba8` 에서 다시 읽음). update 에서 `out` 은 `cmd.OutOrStdout()`(stdout), `errOut` 은 `cmd.ErrOrStderr()`(stderr)이고, 훅 설치 두 곳은 `out, errOut` 을 넘기므로(`internal/cli/update_template_sync.go:1177`, `:1182`) 훅 함수의 `warn` 인자는 stderr 다.
  - stderr 작성기: `internal/cli/hook_install.go:268`·`:273`(`warn`, sevWarn), `:283`·`:291`(`warn`, sevErr); `internal/cli/hook_install_precommit.go:389`·`:395`(`warn`, sevWarn), `:405`·`:413`(`warn`, sevErr); `internal/cli/update_migrate.go:158`(`errOut`, sevWarn); `internal/cli/update_template_sync.go:958`·`:961`(`errOut`, sevWarn); `internal/cli/update_snapshot_hook.go:34`(`errOut`, sevWarn).
  - stdout 작성기: `internal/cli/hook_install.go:293`, `internal/cli/hook_install_precommit.go:415`(`out`, sevErr).
  - 열린 항목: 훅 두 곳과 `update_migrate.go:158` 은 꼬리(마지막으로 열린 항목), `update_template_sync.go:958`·`:961` 은 `deploy.restore`, `update_snapshot_hook.go:34` 는 `update_template_sync.go:687` 을 거쳐 `deploy.deploy`(init 에서는 `init.go:827` 을 거쳐 `templates.deploy`). `internal/cli/update_manifest_heal.go:18`(`errOut`, sevWarn)은 `update.go:576` 에서 불려 `runTemplateSyncWithProgress`(`update.go:604`)보다 앞이므로 열린 항목이 없다 — REQ-TCL-015 의 항목 안 조항 밖이고, 넘김 규칙으로 지금처럼 즉시 한 줄이 보인다.
  - 두 작성기가 모두 쓰이므로 하위 테스트는 두 작성기 각각에 한 번씩 주입한다.
- 종류: release-blocking. RED 이유: L7·L10 — 세션 코드가 없다. 그리고 이 기준의 테스트가 아직 없다 — L28(트리 `6f05252d6d`, `TestChecklistWarningsExactlyOnce` 포함 다섯 이름, 빈 출력, 종료 1). 3회차 핀 `69358ecba8` 의 같은 측정(빈 출력, 종료 1)과 같다. 이 트리에서 다시 잼: L37(트리 `274441db21`, `TestChecklistWarningsExactlyOnce` 와 새 시접 `checklistSessionStartFn` 모두 없음, 빈 출력, 종료 1). 주입 지점의 근거: L41(사실 칸). 기준선: L14 — 패널 연결이 지금 init.go 에 한 곳 있다.
- 녹색 경로(M4): `go test ./internal/cli -run '^TestChecklistWarningsExactlyOnce$' -count=1 -v` → `--- PASS: TestChecklistWarningsExactlyOnce `, `--- PASS: TestChecklistWarningsExactlyOnce/init `, `--- PASS: TestChecklistWarningsExactlyOnce/init_stdout_severity `. (`/update ` 줄은 [DEFERRED → t1978] — 델타 4회차 감사 N6 의 "M5 뒤에 녹색" 문제가 이 이월로 사라진다.)
- 뮤턴트: `emitSeverityLine` 경로를 빠뜨린 구현 → `s-in` 이 접혀 0번 나와 실패. 패널을 그대로 둔 구현 → 패널 머리 1번으로 실패. 항목 밖 경고를 요약에 나열하지 않는 구현 → `c-out` 0번으로 실패. 심각도 줄을 경고 수에 더하는 구현 → 경고 수 5 로 실패. 경고 수신 인터페이스를 stderr 세션 작성기에만 구현한 구현(3회차 감사 M1) → `init` 하위 테스트는 녹색이지만 `init_stdout_severity` 에서 `s-out` 이 성공한 항목의 상세로 접혀 0번, 항목 줄이 `✓` 라 실패. 항목 아래 펼치면서 원래 줄도 실제 작성기로 넘기는 구현 → 2번으로 실패. (v0.2.1) 주입을 가짜 실행기(실제 배포 없음)로 하는 테스트 뮤턴트 → 이 기준이 감싼 실행기 안에 실제 실행기를 요구하므로 리뷰에서 걸린다. 배포되지 않은 프로젝트에서 꼬리 단계가 추가 경고를 내면 경고 수 4 단언이 빨개지지만, 그러지 않을 수도 있어 기계적 차단은 확실하지 않다(남은 위험). `checklistSessionStartFn` 을 첫 항목이 열린 뒤에 부르는 구현 → `w-out`·`c-out` 이 항목 안 경고가 되어 `w-out` 1번·`c-out` 이 항목 아래로 가 요약 나열 0번으로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-017 — 중도 종료와 성공 조기 반환 (maps REQ-TCL-015, REQ-TCL-016, REQ-TCL-023)

- (v0.2.1, 리더 결정 기록 §15.102, 확인 감사 5회차 B2) REQ-TCL-015 의 요약 없는 실행 조항 두 개 — (i) 요약 없이 끝나는 실행은 항목 밖 수집기 경고를 실제 작성기에 한 번 쓴다, (ii) 문장이 이미 실제 작성기에 나간 경고(즉시 줄이 있는 `Warn`)는 다시 쓰지 않는다 — 는 init 전용 조항이며 이 기준이 판정한다. (a1)·(c)·(e) 의 항목 밖 `Warn` 이 (ii) 를, (a2)·(c)·(e) 의 항목 밖 `Collect` 가 (i) 를 본다. 반례로 막는 구현: "종료 때 수집기의 모든 경고를 다시 쓰는" 세션 — git 이 없는 환경에서 `init.go:427` `p.Warn("git is not installed …")`(원장 L41)이 위저드 전에 즉시 줄로 나간 뒤, init 이 오류로 끝나거나(`init.go:866`·`:877`·`:879`) 재개 경로로 `nil` 을 돌려주면(`:864`) 같은 경고가 두 번 보인다. 아래 항목 밖 `Warn` 은 `:427` 과 같은 모양(항목 밖, 즉시 줄 있음)의 주입이다. 실제로 git 을 빼는 고정은 쓰지 않는다 — `PATH` 에서 git 을 빼면 다른 단계의 동작이 바뀌는데 그 영향을 재지 않았다.
- Given 활성 세션에서 (a) 두 변형 — (a1) 첫 항목을 열기 **전에** 수집기 `Warn("w-err")`(항목 밖 — 즉시 줄이 실제 작성기로 나간다)를 기록하고, 항목을 연 채 상세 `detail-x` 를 쓰고 오류를 반환하는 실행, (a2) 항목을 연 채 상세 `detail-x` 를 쓰고, 보고기 `StepError` 로 그 항목이 `Fail` 로 닫혀 열린 항목이 없는 상태에서 수집기 `Collect("c-err")` 를 부른 뒤 오류를 반환하는 실행(실행기가 실패한 뒤 `result.Warnings` 를 `p.Collect(w)` 로 모으는 `internal/cli/init.go:829-838` 실패 경로와 같은 순서, 3회차 감사 M3), (b) 위저드 취소 시접으로 첫 항목 전에 돌아오는 실행, (c) 실제 init 이 초기화해 둔 프로젝트에 중단 체크포인트(중단된 설치 기록 또는 깨진 사용자 매니페스트, `init.go:850-853` 주석)를 둔 `t.TempDir()` 고정에서 **실제 실행기**로 돈 init — 하위 테스트는 실행기 시접 `newInitPhaseExecutorFn`(`internal/cli/init_executor_seam.go:29`)을 바꾸지 않는다. (v0.2.1) 시험 전용 시접 `checklistSessionStartFn`(design.md §3.7)으로 세션 시작 때 — 첫 항목 전에 — 수집기 `Warn("w-resume")` 과 `Collect("c-resume")` 를 기록한다, (d) [DEFERRED → t1978 — 이 SPEC 은 (d) 를 판정하지 않으며, 아래 (d) 문장과 기대 화면은 t1978 이 쓸 근거로 남긴다] 같은 바이너리로 init 해 이미 현재 버전인 `t.TempDir()` 프로젝트에서 `--force` 없이 `runUpdate` 를 거쳐 돈 **실제 `moai update` 명령** — 생산 호출은 `internal/cli/update.go:604` `runTemplateSyncWithProgress` 하나이고, 그 바깥 함수가 `internal/cli/update_template_sync.go:1260-1266` 에서 결과 알약(`:1262`)을 쓰고 `return true, nil` 하므로 본문 `runTemplateSyncWithReporter` 는 불리지 않는다. 본문의 같은 모양 조기 반환(`:249-255`)은 본문을 직접 부르는 테스트만 닿는 경로라, 이 하위 테스트는 본문을 직접 부르지 않는다(3회차 감사 M2, design.md §3.4), (e) 세션 단위로 첫 항목을 열기 **전에** 수집기 `Warn("w-nil")` 과 `Collect("c-nil")` 을 기록하고(둘 다 항목 밖, v0.2.1), 항목을 연 채 상세 `detail-e` 와 보존 줄 `keep-e` 를 기록하고 `nil` 을 돌려주는 실행 — 실제 명령에는 열린 항목을 가진 `nil` 조기 반환이 없으므로(design.md §3.4) REQ-TCL-023 의 "열린 항목을 done 으로 닫는다" 조항은 이 세션 단위 경우가 본다, When 끝나면, Then
  - (a) 두 변형 모두 `✗` 줄과 `detail-x` 가 나오고, 요약 머리 문자열과 `warning(s) during init:` 패널 머리 문자열은 0번이다. (a1) 에서 `w-err` 는 실제 작성기 출력에 정확히 1번 나온다 — 즉시 줄 한 번이고, 종료 처리가 다시 쓰지 않는다(REQ-TCL-015 조항 (ii)). (a2) 에서 `c-err` 는 실제 작성기 출력에 정확히 1번 나온다(REQ-TCL-015 조항 (i), REQ-TCL-016).
  - (b) 취소 문구(`Initialization cancelled.`, `init.go:575`)가 나오고, 체크리스트 머리줄은 0번이다(항목이 하나도 열리지 않은 실행 — design.md §3.4 "머리줄은 첫 항목이 열릴 때"; v0.2.0 에 (d) 가 이월되어 이 조항을 init 쪽에서 보는 단언을 더했다).
  - (c) 실제 이벤트 순서(design.md §3.4, 원장 L21 — 검증 단계가 `StepError` 를 먼저 낸 뒤 재개가 돈다)대로, `templates.detect` 줄이 `✓` 로, `templates.validate` 줄이 `✗` 로 끝나고 그 아래 들여쓴 `project already initialized`(`ErrProjectExists`, `internal/core/project/errors.go:12`)가 보인다. 출력 전체에서 `✗` 는 정확히 1번, `Initialized MoAI project (resumed` 는 정확히 1번, 요약 머리 문자열은 0번이고, 명령은 `nil` 을 돌려준다. (v0.2.1) `w-resume` 은 정확히 1번(즉시 줄만 — 조항 (ii)), `c-resume` 은 정확히 1번(요약이 없으므로 종료 처리가 쓴 한 번 — 조항 (i)), 패널 머리 문자열은 0번이다. `✗` 줄이 끝난 뒤 그 줄을 다시 그리는 `\r\x1b[2K` 는 없다(REQ-TCL-002·REQ-TCL-023).
  - (d) [DEFERRED → t1978] 체크리스트 머리줄은 0번이다 — 항목이 하나도 열리지 않은 실행이다(design.md §3.4). 이미 최신이라는 결과 알약 문장(`report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, "")`)은 실제 stdout 에 정확히 1번 나온다(넘김 규칙, 보존 행 P22 — 항목 밖). `✗` 는 0번, 요약 머리 문자열은 0번이고, 명령은 `nil` 을 돌려준다. (v0.2.0: P22 가 update 행으로 함께 이월되어 AC-TCL-024 (2) 의 옛 P22 괄호는 지웠다 — 델타 4회차 감사 N3.)
  - (e) 그 항목 줄이 `✓` 로 끝나고, `detail-e` 는 0번(접힘), `keep-e` 는 정확히 1번, `✗` 와 요약 머리 문자열은 0번이다. (v0.2.1) `w-nil` 은 정확히 1번(즉시 줄만 — 조항 (ii)), `c-nil` 은 정확히 1번(종료 처리가 실제 작성기에 쓴 한 번 — 조항 (i))이다.
  - 이 SPEC 이 판정하는 다섯 경우((a1)·(a2)·(b)·(c)·(e)) 모두 끝난 뒤 `cmd.OutOrStdout()`·`cmd.ErrOrStderr()` 가 원래 작성기를 돌려준다.
- 종류: release-blocking. RED 이유: L10 — 세션 코드가 없다. 그리고 이 기준의 테스트가 아직 없다 — L28(트리 `6f05252d6d`, `TestChecklistAbortFlushes` 포함, 빈 출력, 종료 1). 3회차 핀 `69358ecba8` 의 같은 측정(빈 출력, 종료 1)과 같다. 이 트리에서 다시 잼: L37(트리 `274441db21`, `TestChecklistAbortFlushes` 와 시접 `checklistSessionStartFn` 모두 없음, 빈 출력, 종료 1). (c) 의 기대 화면은 L21(실행기 이벤트 순서)과 design.md §3.4 가 근거다. 반례 `init.go:427` 과 항목 밖 `Collect` 가 실패 경로 `:837` 에만 있다는 사실은 L41(사실 칸).
- (d)·(a2) 의 경로 근거 — 앞의 세 칸은 (d) 의 근거라 [DEFERRED → t1978] 와 함께 t1978 이 쓰고, 넷째 칸(`p.Collect(w)`)이 (a2) 의 근거다. 같은 트리 `69358ecba8a9b8f996b3e9b0e4b39ae36728a80f` 에서 각각 한 번의 호출로 잰 사실 칸(RED 칸이 아니라 기대 화면의 근거):
  - `/usr/bin/grep -n "runTemplateSyncWithProgress(" internal/cli/update.go` → stdout `604:	syncSkipped, err := runTemplateSyncWithProgress(cmd, userAssetsInstalled)`, 종료 0 — 생산 호출 한 곳.
  - `/usr/bin/grep -rn "runTemplateSync(cmd" internal/cli/update.go internal/cli/update_template_sync.go` → stdout `internal/cli/update_template_sync.go:120:func runTemplateSync(cmd *cobra.Command) error {`, 종료 0 — 정의만 있고 생산 호출은 없다.
  - `/usr/bin/grep -n "OutcomeAlreadyUpToDate" internal/cli/update_template_sync.go` → stdout `252:		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, ""), Theme: &th}))` / `1262:		_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: report.RenderOutcome(report.OutcomeAlreadyUpToDate, 0, ""), Theme: &th}))`, 종료 0 — 실제 명령이 닿는 것은 `:1262`(바깥 함수)다.
  - `/usr/bin/grep -n "p.Collect(w)" internal/cli/init.go` → stdout `837:				p.Collect(w)` / `954:		p.Collect(w)`, 종료 0 — `:837` 이 실행기 실패 경로의 수집이다(`init_warnings.go:49-56` — `Collect` 는 줄을 쓰지 않는다).
- 녹색 경로(M3·M4): `go test ./internal/cli -run '^TestChecklistAbortFlushes$' -count=1 -v` → `--- PASS: TestChecklistAbortFlushes/error `, `--- PASS: TestChecklistAbortFlushes/error_collect_after_fail `, `--- PASS: TestChecklistAbortFlushes/cancel `, `--- PASS: TestChecklistAbortFlushes/resumed_after_validate_fail `, `--- PASS: TestChecklistAbortFlushes/nil_return_open_item `. (`/update_up_to_date ` 은 [DEFERRED → t1978].)
- 뮤턴트(v0.2.1, 확인 감사 5회차 B2): 종료 때 수집기의 모든 경고를 다시 쓰는 세션(반례 `init.go:427` 의 두 번 보임) → (a1) `w-err` 2번, (c) `w-resume` 2번, (e) `w-nil` 2번으로 실패 — 오류 경로와 실제 재개(`nil` 반환) 경로 모두에서 막힌다. `nil` 조기 반환에서 수집기 경고 쓰기를 빠뜨린 세션 → (c) `c-resume` 0번, (e) `c-nil` 0번으로 실패. 오류 경로에서만 쓰고 `nil` 경로에서는 쓰지 않는 세션 → (a2) 는 녹색이어도 (c)·(e) 가 빨갛다. 항목 밖 `Warn` 의 즉시 줄을 세션이 삼키고 종료 때 한 번 쓰는 세션 → 횟수는 1번이라 이 기준으로는 녹색이지만, "기존 즉시 줄을 그대로 둔다"(REQ-TCL-015)를 어기며 AC-TCL-015 의 "쓴 즉시" 검사(항목 밖 바이트는 즉시 넘김)와 AC-TCL-016 의 `w-out` 2번(즉시 줄 + 요약)이 잡는다.
- 뮤턴트(v0.2.0): (d) 를 지우면서 `nil` 조기 반환 처리까지 지운 세션 → (e) `nil_return_open_item` 의 `keep-e` 0번과 (c) 의 재개 경로(`nil` 반환)로 실패 — `nil` 조기 반환 조항(REQ-TCL-023)은 init 쪽 두 경우가 계속 본다.
- 뮤턴트(3회차 감사 M2·M3 추가분): 세션 종료 때 캡처 바이트만 쏟고 수집기 경고를 보지 않는 구현(2회차 REQ-TCL-016 문장 그대로의 구현) → (a2) 의 `c-err` 0번으로 실패. 오류 경로에서 패널을 그대로 내 `c-err` 를 패널로 보이는 구현 → 패널 머리 1번으로 실패. 패널과 세션 쓰기를 둘 다 하는 구현 → `c-err` 2번으로 실패. [DEFERRED → t1978: 본문 `runTemplateSyncWithReporter` 를 직접 불러 `:249-255` 죽은 경로로 녹색을 만드는 테스트 → 본문이 `templates.version` 을 열어 머리줄 1번이 되므로 빨강. 바깥 함수의 조기 반환 앞에서 항목을 미리 여는 구현 → 머리줄 1번으로 실패. 알약을 버리는 구현 → 알약 0번으로 실패.] 세션 시작 때 머리줄을 바로 그리는 구현 → (b) 취소(항목 없는 실행)에서 머리줄 1번으로 실패.
- 뮤턴트(2회차까지): 작성기를 되돌리지 않는 구현 → 원래 작성기 검사 실패. 취소 때 캡처를 버리는 구현 → 취소 문구 부재로 실패. 가짜 실행기로 `StepError` 없이 재개 경로만 태운 테스트(2회차 감사 N3) → `✗` 정확히 1번 조건이 빨갛다 — 실제 순서를 재현해야만 녹색이다. 재개가 성공하면 실패 줄을 `✓` 로 다시 그리는 구현 → `✗` 0번과 다시 그리기 시퀀스로 실패. 실패를 보류해 재개 단계 출력을 검증 항목 상세로 접는 구현 → 보존 행 P23 의 문장이 사라져 AC-TCL-024 `p23_init_resume_warnings` 로 실패. 모든 `nil` 조기 반환을 실패로 닫는 구현 → (e) 의 `✗` 로 실패. `nil` 조기 반환 때 캡처한 보존 줄을 버리는 구현 → (e) 의 `keep-e` 0번으로 실패. `nil` 조기 반환 때 상세까지 쏟는 구현 → (e) 의 `detail-e` 1번으로 실패. 조기 반환에도 요약을 그리는 구현 → (c)·(e) 의 요약 머리 1번으로 실패(v0.2.0 까지 적었던 (d) 는 [DEFERRED → t1978] 라 판정 근거에서 뺐다 — 확인 감사 5회차 M1). 재개 안내를 접는 구현 → 안내 0번으로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-018 — 출력 채널 (maps REQ-TCL-017)

- Given stdout·stderr 를 서로 다른 버퍼로 둔 활성 실행, When init 을 돌리면, Then 체크리스트 머리줄과 모든 항목 줄은 stderr 버퍼에만 있고 stdout 버퍼에는 체크리스트 머리줄이 0번이다. (update 의 stdout 채널은 [DEFERRED → t1978].)
- 종류: release-blocking. RED 이유: L10. 이 트리에서 다시 잼: L27(세션 파일 없음, 종료 1), L28(`TestChecklistChannelPartition` 없음, 종료 1).
- 녹색 경로(M4): `go test ./internal/cli -run '^TestChecklistChannelPartition$' -count=1 -v` → `--- PASS: TestChecklistChannelPartition/init `.
- 뮤턴트: init 체크리스트를 stdout 에 그리는 구현(데이터 채널 오염) → stdout 머리줄 1번으로 실패. 두 작성기 모두에 그리는 구현 → stdout 머리줄 1번으로 실패. 아무 데도 그리지 않는 구현 → stderr 머리줄 0번으로 실패(양방향).
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-019 — 질문 중 열린 줄 없음, 위저드 불변 (maps REQ-TCL-018)

- Given init 위저드 진입 지점의 시접, When 그 시접이 불릴 때 세션 상태를 기록하면, Then 열린 항목이 0개이고 끝나지 않은 줄(스피너 줄)이 화면에 없다. 그리고 이 카드의 커밋은 위저드 패키지를 바꾸지 않았다. (update 병합 확인 지점은 [DEFERRED → t1978].)
- 종류: release-blocking(앞 절), regression-guard(뒤 절). RED 이유: L10. 이 트리에서 다시 잼: L27(세션 파일 없음, 종료 1).
- 녹색 경로(M4): `go test ./internal/cli -run '^TestChecklistNoOpenLineDuringPrompt$' -count=1 -v` → `--- PASS: TestChecklistNoOpenLineDuringPrompt/init `. 보조: `git log --no-merges --grep=t1969 --name-only --format= cb30e05b3e..HEAD ^main -- internal/cli/wizard` → (빈 출력), 종료 0, 그리고 §A 의 카드 id 완전성 검사가 빈 출력. 이 카드의 커밋만 보므로 t1966 의 위저드 변경을 흡수해도 빨개지지 않는다.
- 뮤턴트: 세션 시작 때(위저드 전) 첫 항목이나 머리줄 스피너를 여는 구현 → 위저드 시접 시점에 열린 항목 1개로 실패. 위저드 전 단계(예: 사용자 자산 보장)를 항목으로 만든 구현 → 같은 이유로 실패(레인 결정 Q3 과도 충돌). [DEFERRED → t1978: `analyze` 항목을 닫지 않고 확인 화면을 띄우는 update 구현.]
- 계속 발화: 앞 절은 기본 실행. 뒤 절은 이 카드 범위 확인용 일회성 검사이며 sync 감사가 다시 돈다(병합 뒤에는 멈춘다 — 의도된 한계).

### AC-TCL-020 — 숫자는 바뀌지 않음 (maps REQ-TCL-019)

- Given 같은 고정 프로젝트에서 한 번은 비활성(대체 경로), 한 번은 활성으로 돌리고 수집기 경고 2건과 심각도 경고 1건을 주입한 init, When 두 출력의 숫자를 뽑으면, Then init 의 디렉터리·파일 수와 경고 수(둘 다 수집기 수 2)가 두 실행에서 같고, 배포된 파일 목록(디스크)도 같으며, 명령 반환값도 같다. (update 의 갱신 파일 수·보존 수·경고 수 동치는 [DEFERRED → t1978].)
- 종류: release-blocking. RED 이유: L10. 이 트리에서 다시 잼: L27(세션 파일 없음, 종료 1).
- 녹색 경로(M4): `go test ./internal/cli -run '^TestChecklistSummaryMatchesLegacyCounts$' -count=1 -v` → `--- PASS: TestChecklistSummaryMatchesLegacyCounts/init `.
- 뮤턴트: 요약이 계획표 항목 수를 "바뀐 수" 로 찍는 구현 → 숫자 불일치로 실패. 심각도 줄을 경고 수에 더하는 구현 → 활성 3 · 비활성 2 로 실패.
- 계속 발화: `internal/cli` 기본 실행.

### AC-TCL-021 — 배포 템플릿 불변 (maps REQ-TCL-020)

- Given 이 카드의 모든 커밋, When `git log --no-merges --grep=t1969 --name-only --format= cb30e05b3e..HEAD ^main -- internal/template/templates` 를 돌리면, Then 출력이 비어 있고 종료 0 이며, §A 의 카드 id 완전성 검사도 빈 출력이다.
- 종류: regression-guard. 오늘 이미 녹색(L13 이 같은 범위의 트리 비교로 빈 출력)이므로 통과로 세지 않는다.
- 녹색 경로: 전 구간 유지. M6 와 sync 감사가 다시 돈다.
- 뮤턴트: 템플릿에 체크리스트 설정 키를 넣는 구현 → 출력에 경로가 나와 실패.
- 계속 발화: 일회성 범위 검사. 병합 뒤에는 멈춘다. 템플릿 중립성 자체는 기존 템플릿 테스트가 계속 지킨다.

### AC-TCL-022 — init.go 변경 영역 제한 (maps REQ-TCL-021)

- Given 이 카드의 커밋들과 plan.md §B 의 영역 기준 SHA(처음 값 `cb30e05b3e`), When (1) `git diff -U0 <영역 기준> HEAD -- internal/cli/init.go` 의 hunk 머리를 읽고, (2) `git log --merges --format=%h <영역 기준>..HEAD -- internal/cli/init.go` 를 돌리면, Then (2) 의 출력이 비어 있고, (1) 의 hunk 가 1개 이상이며, 모든 hunk 의 이전 쪽 범위 `-a,b` 가 한 영역 안에 있다 — `b>0` 이면 시작 줄 `a` 와 끝 줄 `a+b-1` 이 모두, `b=0`(삽입)이면 `a` 가 R1 `[383,389]`, R3 `[919,1075]`, R2 `[1077,1087]` 가운데 하나에 들며, R3 안의 hunk 는 모두 `b=0` 이다. 그리고 §A 의 카드 id 완전성 검사가 빈 출력이다.
- 좌표계(2회차 감사 N5). (1) 은 영역 기준과 HEAD 두 트리를 바로 비교하므로, init.go 를 고친 이 카드의 커밋이 몇 개이든 hunk 좌표는 영역 기준 파일의 행번호다 — REQ-TCL-021 의 "영역 기준 대비"와 같은 좌표다. 1회차의 `git log -p` 는 커밋마다 그 부모 대비 좌표를 내서, 앞 커밋의 삽입만큼 뒤 커밋의 좌표가 밀렸다. 일부러 만든 저장소에서 두 커밋(10행 뒤 3줄 삽입, 원래 15행 수정)을 넣으면 `git log -p` 는 둘째 커밋을 `@@ -18 +18 @@` 로, 누적 diff 는 `@@ -15 +18 @@` 로 낸다(research.md §10). (1) 이 "이 카드의 변경만"을 담는 근거는 둘이다 — 영역 기준 이후 init.go 를 건드린 병합이 없다는 (2), 그리고 `internal/` 을 건드린 커밋이 모두 이 카드의 것이라는 §A 완전성 검사. `main` 흡수는 M3 전에만 하고 그때 영역 기준을 흡수 커밋으로 옮기므로(plan.md §B 1(a)), M4 의 init.go 변경은 언제나 영역 기준 뒤에 있다.
- 종류: release-blocking. RED 이유: L22 — 영역 기준 이후 init.go hunk 가 0개(빈 출력, 종료 0)라 "1개 이상"이 거짓이다. L16 — 녹색 테스트가 아직 없다. 빨간 이유는 "연결이 아직 없다"이며, 범위 밖 변경 때문이 아니다. (2) 는 오늘 빈 출력이다(L23, 범위 보호 조건).
- 녹색 경로(M4): M4 가 넣는 `TestChecklistInitHunkConfinement` 가 위 두 명령의 출력을 해석해 판정한다 → `go test ./internal/cli -run '^TestChecklistInitHunkConfinement$' -count=1 -v` → `--- PASS: TestChecklistInitHunkConfinement `. git 이 없거나 이력이 얕으면 건너뛰지 않고 분명한 메시지로 실패한다.
- 뮤턴트: 보고기 줄(`init.go:728`)을 고친 구현 → 시작 줄 728 이 범위 밖이라 실패. 1070 에서 시작해 1100 까지 지우는 hunk → 끝 줄 1100 이 범위 밖이라 실패. R3 안의 기존 줄을 고치는 hunk(`b>0`) → R3 삽입 전용 조건으로 실패. R3 에 14줄을 넣은 뒤 다른 커밋에서 R2 를 고치는 올바른 구현 → 누적 좌표로 R2 안에 들어 녹색(1회차 명령은 이 구현을 잘못된 이유로 빨갛게 만들었다). 범위 밖 변경을 다른 커밋으로 나눠 밀린 좌표로 범위 안에 숨기는 구현 → 누적 좌표에서는 밀리지 않아 실패. init.go 변경을 `main` 흡수 병합에 섞는 절차 → (2) 가 병합 SHA 를 내 실패. (v0.2.2, 결정 기록 §15.107) 기록된 마지막 실행 뒤 — 테스트를 지운 M6 커밋 뒤 포함 — 다른 커밋에서 `init.go:728` 을 고치는 절차 → 기록은 그 커밋 앞의 녹색이라 기록만 읽는 판정은 녹색이지만, 아래 최종 HEAD 재실행에서 시작 줄 728 hunk 가 나와 실패. 재실행 출력을 기록에서 옮겨 적는 감사 → 재실행의 측정 SHA 가 최종 HEAD 와 달라 실패.
- 계속 발화: 이 카드 범위의 일회성 조율 검사라 테스트는 M6 에서 지운다. 지우기 전 마지막 실행(두 명령·그대로의 출력·측정 SHA)을 progress.md §E.2 에 남긴다. 그 기록만으로는 판정하지 않는다 — 기록 뒤에 들어온 커밋이 init.go 를 바꿀 수 있기 때문이다. (v0.2.2) sync 감사는 판정할 **최종 HEAD** 에서 위 두 명령 (1)·(2) 를 각각 한 번 다시 돌려(테스트가 지워졌으므로 명령을 직접 돌린다) 그대로의 출력·종료 코드·측정 SHA(= 최종 HEAD, 40자)를 자기 판정 파일에 남기고, 같은 영역 조건(hunk 1개 이상, 모든 hunk 가 R1·R3·R2 안, R3 은 `b=0`, (2) 빈 출력)으로 판정한다. 기록과 재실행이 모두 녹색이어야 녹색이고, 재실행 측정 SHA 가 최종 HEAD 가 아니면 빨강이다.

### AC-TCL-023 — 글리프 허용 목록과 원시 키 부재 (maps REQ-TCL-001, REQ-TCL-003, REQ-TCL-009)

- Given AC-TCL-001, AC-TCL-002, AC-TCL-003, AC-TCL-004, AC-TCL-005, AC-TCL-006, AC-TCL-007, AC-TCL-009 가 만드는 모든 렌더러 출력과 `internal/tui/testdata/checklist-*.golden`, When ANSI 를 지우고 문자를 훑으면, Then 상태 자리 문자는 `✓ ✗ ! · ● ○ ◆ ◇` 또는 U+2800–28FF 점자뿐이고, 출력 어디에도 U+27A4(`➤`)와 U+2013(`–`)이 없으며, 모든 렌더러 골든에 부분 문자열 `checklist.` 가 없다.
- 종류: release-blocking. RED 이유: L1·L10.
- 녹색 경로(M2): `go test ./internal/tui -run '^TestChecklistGlyphWhitelist$' -count=1 -v` → `--- PASS: TestChecklistGlyphWhitelist `.
- 뮤턴트: 카드 문구 그대로 `➤` 를 쓴 렌더러 → 실패. 카탈로그에 없는 키를 불러 `Translate` 가 키 이름을 돌려주고 그 출력으로 골든을 찍은 경우 → 원시 키 검사로 실패.
- 계속 발화: `internal/tui` 기본 실행.

### AC-TCL-024 — 보존 줄은 성공 항목에서도 한 번 (maps REQ-TCL-022)

- Given 활성 init 세션과 design.md §3.3.1 의 보존 행 가운데 이 SPEC 범위의 행 — P1, P2, P3, P4, P5, P17(init 지점, v0.2.2), P20(init 지점), P21(init 지점), P23. 각 하위 테스트는 **한 행만** 맡아 그 행의 init 호출 지점 모두에 실패·안내를 차례로 주입한다(예: P4 는 `init.go:863` 과 `:1058` 두 지점, P20 은 init `:1022` 한 지점). When 실행을 끝까지 돌리면, Then
  - (1) 그 행의 문장이 실제 작성기 출력에 정확히 한 번 나온다.
  - (2) 위치가 항목 안인 경고·오류 행은 그 항목 줄을 `!` 로, 안내 행은 그 항목 줄을 `✓` 로 끝낸다. 위치가 항목 밖인 행(P3, P4 의 `:863`, P23)은 글리프를 보지 않는다. (v0.1.3 의 "P22 는 REQ-TCL-023 의 `✓`" 괄호는 P22 가 update 행으로 이월되어 지웠다 — 델타 4회차 감사 N3.)
  - (3) 요약이 그려지는 실행에서 요약의 경고 수는 주입 전과 같다(결정 D-9).
  - (3a) (v0.2.2, 결정 기록 §15.107) `p17_init_manifest_retrack` 은 **성공을 돌려주는** 단계가 쓴 경고를 본다. `init.go:1073` `retrackSectionFiles` 는 파일 하나를 건너뛰면 `internal/cli/update_manifest_retrack.go:57` 에 `  manifest retrack skip <경로>: <오류>` 를 쓰고 `continue` 하므로, 나머지 파일의 `Save` 가 성공하면 `nil` 을 돌려주고 `:1074` `p.Collect` 는 불리지 않는다. 주입: 실제 실행기를 감싼 실행기(design.md §3.7 의 항목 안 주입과 같은 시접 `newInitPhaseExecutorFn`)가 배포 뒤 template_managed 인 `.moai/config/sections/*.yaml` 하나를 읽을 수 없게(`chmod 000`) 만든다 — `manifest.Track` 의 `HashFile` 이 실패해(`internal/manifest/manifest.go:141-143`) `:57` 줄이 나온다. Then 그 줄이 정확히 한 번 `verify.manifest` 항목 아래 펼쳐지고, 그 항목 줄은 `!` 로 끝나며(`✓` 아님), 명령은 `nil` 을 돌려주고, 요약의 경고 수는 주입 전과 같다(보존 행은 경고 수에 더하지 않는다, 결정 D-9). 그 파일을 `:1073` 앞의 다른 단계가 읽어 init 이 먼저 실패하거나 줄이 나오지 않으면, run 단계는 막힘 보고를 내고 manager-spec 이 주입 대상을 고친다(이 회차에 그 단계들의 파일 읽기는 따라가지 않았다). `chmod 000` 은 Windows 와 root 실행에서 읽기를 막지 못하므로 이 하위 테스트는 `//go:build !windows` 파일에 두고, root 로 돌면 건너뛰지 않고 분명한 메시지로 실패한다 — 녹색 경로 판정은 비Windows·비root 실행에서 한다.
  - (4) `inventory_complete` 하위 테스트는 원장 L20 의 12개 모양과 원장 L26 의 2개 모양(`PillWarn`, `PillErr` — 3회차 감사 M8), 모두 14개 모양을 Go 코드로 `internal/cli/**/*.go`(테스트 파일 제외)에 돌려 — (가) 보존 행·제외 행(X1~X11) 어디에도 분류되지 않은 일치가 없고, (나) 모든 보존 행의 분류 조각이 그 파일에 아직 있으며, (다) 궤도 파일(design.md §3.3.1 정의)의 일치가 파일 단위 제외로 처리되지 않았음을 단언한다. X11(이 SPEC 에서 update 행으로 이월한 P6~P19 의 분류 조각과 원장 L26 의 `update_wizard.go`·`update_tux.go:120` 네 줄)도 줄 단위 분류이므로 (다) 를 그대로 따른다. update 의 P22 는 열거 행이라 검색 일치·분류 조각이 없어 X11 에 들지 않는다([DEFERRED → t1978], design.md §3.3.1 — 확인 감사 5회차 M1).
  - (5) (3회차 감사 M9) 궤도 목록이 테스트 안의 목록만으로 정해지지 않도록, `inventory_complete` 는 design.md §3.3.1 의 궤도 항목 열여섯 개마다 하위 테스트 하나를 갖고, 각 하위 테스트는 그 항목이 기존 파일 1개 이상에 맞고 그 파일들의 모든 일치가 줄 단위로 분류됨을 단언한다. 궤도 항목을 테스트에서 지우면 녹색 경로의 한 줄이 사라진다.
- 종류: release-blocking. RED 이유: L17 — 이 테스트가 없다. L7·L10 — 세션 코드가 없다. 이 트리에서 다시 잼: L28(`TestChecklistPreservedLinesExactlyOnce` 없음, 종료 1), L27(세션 파일 없음, 종료 1). 오늘(비활성 경로)은 모든 문장이 보이지만, 체크리스트가 켜진 경로 자체가 없어 이 기준의 활성 조건을 만들 수 없다. L20·L26 이 완전성 하위 테스트가 다시 돌릴 검색의 기준선이다.
- 녹색 경로(M4): `go test ./internal/cli -run '^TestChecklistPreservedLinesExactlyOnce$' -count=1 -v` → 아래 26줄이 모두 있어야 한다.
  `--- PASS: TestChecklistPreservedLinesExactlyOnce/p01_init_snapshot_stage `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/p02_init_snapshot_settle `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/p03_init_resumed_notice `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/p04_init_update_notice `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/p05_init_worktree_advisory `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/p17_init_manifest_retrack `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/p20_prepush_preserved `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/p21_precommit_preserved `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/p23_init_resume_warnings `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_init_files `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_update_files `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_update_pkg `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_checklist_files `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_hook_install `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_hook_install_precommit `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_worktree_advisory `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_feedback_participation `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_migrate_project_assets `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_migrate_local_instructions `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_session_worktree `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_mirror_notice `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_user_asset_phase `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_uikit_styles `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_printer `, `--- PASS: TestChecklistPreservedLinesExactlyOnce/inventory_complete/orbit_severity_line `.
  (v0.1.3 의 `p06_update_env_path_refresh` 부터 `p19_update_stamp_skew` 까지 열네 줄과 `p22_update_up_to_date`, 그리고 P17·P20·P21 의 update 지점 주입은 [DEFERRED → t1978]. P17 의 init 지점은 v0.2.2 에 `p17_init_manifest_retrack` 으로 들어왔다.)
- 뮤턴트: 네 지점만 세션에 알리고 나머지 행은 그대로 둔 구현(2회차 감사 N2) → 알리지 않은 행의 하위 테스트가 그 문장 0번으로 빨갛다 — 행마다 하위 테스트가 있으므로 어느 한 행도 빠질 수 없다. 테스트 표에서 행 하나를 지워 하위 테스트를 없앤 테스트 뮤턴트 → 그 행의 검색 일치가 미분류가 되어 `inventory_complete` (가) 로 실패하고, 녹색 경로의 26줄 가운데 한 줄이 빠진다. 궤도 파일의 새 직접 쓰기 경고를 파일 단위 제외로 흘려보내는 뮤턴트 → (다) 로 실패. 보존 지점의 메시지를 지우거나 바꾼 구현 → (나) 로 실패. 성공 항목의 상세를 모두 펼치는 구현 → AC-TCL-015(`inside` 미노출)와 AC-TCL-011(줄 예산)이 잡는다. 보존 줄을 즉시 넘기고 종료 때 또 펼치는 구현 → 2번으로 실패. 안내를 경고로 세는 구현 → 경고 수와 `✓` 조건으로 실패.
- 뮤턴트(v0.2.0): 테스트의 궤도 목록에서 `printer/printer.go` 를 빼 그 파일의 일치를 X9 로 흘리는 테스트 뮤턴트(3회차 감사 M5·M9) → `orbit_printer` 줄이 녹색 경로에서 사라져 빨강. X11 로 이월한 update 행을 파일 단위 제외로 바꿔 분류 작업을 줄이는 테스트 뮤턴트 → `orbit_update_files` 의 (다) 로 실패. `PillErr` 알약(`update_tux.go:120`)을 분류하지 않은 표 → 14개 모양 검색의 미분류 일치로 (가) 실패. P1 의 init 주입 지점에서 쓴 경고가 update 행 P7 과 같은 분류 조각(`file_snapshot.go:95`·`:99`)을 쓰는데, 그 조각을 X11 로만 분류하는 표 → P1 의 (나) 가 조각을 보존 행에서 찾지 못해 실패 — 같은 조각은 보존 행 P1 이 소유하고 X11 은 P7 의 호출 지점만 적는다(design.md §3.3.1). (v0.2.2) 단계가 `nil` 을 돌려주면 그 항목의 캡처를 모두 성공 상세로 접는 구현 → `p17_init_manifest_retrack` 에서 `manifest retrack skip` 0번, `verify.manifest` 줄이 `✓` 라 실패. P17 조각을 X11 로만 분류한 표 → P17 의 (나) 가 조각을 보존 행에서 찾지 못해 실패.
- 계속 발화: `internal/cli` 기본 실행. 검색 모양에 걸리는 새 직접 쓰기 줄이 궤도 파일에 생기면 `inventory_complete` 가 그 자리에서 빨개진다 — 사람이 목록을 기억해 고칠 필요가 없다. 궤도 밖 파일의 일치는 규칙상 X9 라 다른 명령을 고치는 카드가 이 테스트를 빨갛게 만들지 않는다. 궤도 밖 파일이 새로 init·update 에서 불리게 되는 변경은 이 테스트가 보지 못한다(남은 위험, plan.md §H R-3). 이 카드 안에서 새 일치가 생기면 run 단계는 막힘 보고를 내고 manager-spec 이 design.md §3.3.1 에 행을 더한다.

## §D 경계 사례

- 경고와 실패가 한 항목에 같이 오면 `✗` 하나로 끝내고, 경고와 상세를 모두 펼친다.
- 항목 없이 끝나는 실행(위저드 취소)은 머리줄도 그리지 않는다. 머리줄은 첫 항목이 열릴 때 그린다.
- 터미널 폭보다 긴 항목 라벨은 자르지 않는다. 자르기는 범위 밖이며 Gap G-6 로 남긴다(plan.md §I).
- `NO_COLOR` 와 모션 감소가 함께 켜져도 AC-TCL-006·AC-TCL-007 이 각각 성립해야 한다.

## §E 품질 관문

- `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...` 종료 0.
- `go vet ./internal/tui/... ./internal/cli/...` 종료 0.
- 바뀐 패키지 범위 테스트는 §C 의 기준별 명령으로 돈다. 전체 `internal/cli` 실행은 기계 부하 규칙에 따라 자원 임대(`moai slot acquire`) 아래에서 한 번, 또는 CI 에 맡긴다.
- `internal/tui` 커버리지 85% 이상: `go test -cover ./internal/tui` 출력을 그대로 기록.
- `golangci-lint run --timeout=2m` 에서 새 지적 0.

## §F 완료 정의

- AC-TCL-001 부터 AC-TCL-020 까지, AC-TCL-023, AC-TCL-024 가 §A 의 녹색 판정을 만족한다. `[DEFERRED → t1978]` 로 표시한 하위 테스트·조건은 판정에 넣지 않으며, 이 카드의 테스트에 그 하위 테스트를 두지 않는다(건너뜀 하위 테스트로 남기지 않는다).
- AC-TCL-022 는 M6 에서 테스트를 지우기 **전** 마지막 실행 기록(두 명령, 그대로의 출력, 종료 코드, 측정 SHA)과, sync 감사가 판정할 최종 HEAD 에서 두 명령 (1)·(2) 를 다시 돌린 결과(측정 SHA = 최종 HEAD, 40자 — AC-TCL-022 계속 발화)가 **모두** §A 의 녹색 판정을 만족할 때만 성립한다. 기록만으로는 판정하지 않으며, 재실행 측정 SHA 가 최종 HEAD 가 아니면 빨강이다. 지운 뒤의 `[no tests to run]` 은 판정에 쓰지 않는다(v0.2.3, 결정 기록 §15.108).
- AC-TCL-021 은 regression-guard 로 유지된다.
- AC-TCL-013 의 순서 증거 O1·O2·O3·O4(O4a 기록의 측정 HEAD = O1 SHA, O4b 재실행 녹색)가 모두 성립한다.
- §A 의 카드 id 완전성 검사가 M6 에서 빈 출력이다.
- `go test ./internal/tui/golden -run '^(TestManifest|TestGoldenCount)$' -count=1 -v` → `--- PASS: TestManifest `, `--- PASS: TestGoldenCount `.
- progress.md §E.2 에 각 기준의 명령·그대로의 출력·종료 코드·측정 HEAD 가 기록돼 있다.
