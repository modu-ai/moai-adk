---
id: SPEC-WIN-SMARTPATH-001
title: "Windows-safe SmartPATH — GOOS-injected PATH generation for settings.json"
version: "1.0.0"
status: in-progress
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.2.0"
module: "internal/template"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "windows, settings, path, hooks, smartpath, gh-1690"
---

# SPEC-WIN-SMARTPATH-001 — Windows-safe SmartPATH

## HISTORY

- 2026-09-07: draft created (card t515, GH #1690, lane-10). Phase 8 proposal + Phase 10 authoring executed orchestrator-direct (manager-spec spawn failed twice with API 429 — fallback reason recorded in progress.md §F note).

## §A User Story / Problem

Windows 사용자(moai-adk 3.1.2, GH #1690)가 세션을 시작하면 모든 훅이 `Executable not found in $PATH: "bash"` 로 실패한다. 원인은 `.claude/settings.json` 의 `env.PATH` 가 Windows 구분자(`;`)와 POSIX 경로 6개를 섞은 값이기 때문이다:

```
C:\Users\<u>\.local\bin;C:\Users\<u>\go\bin;/usr/local/bin;/usr/local/sbin;/usr/bin;/bin;/usr/sbin;/sbin
```

이 값이 사용자 전역 설정 1곳 + MoAI 프로젝트 2곳에서 바이트 단위로 동일하게 관찰됐다 — 사람이 각각 쓴 값이 아니라 단일 생성 경로가 뿌린 값이다. 본 SPEC 은 그 생성 경로를 고쳐 Windows 에서 올바른 PATH 를 렌더하게 한다.

## §B Evidence Base

### B.1 근본 결함 — GOOS-blind 생성기 (재현 가능, darwin 에서 코드 판독으로 확정)

`internal/template/settings.go` `BuildSmartPATH()` (:44-111):
- `sep := os.PathListSeparator` — Windows 에서 `;`
- `switch runtime.GOOS` (:59-79) 에 `case "darwin"` 과 `default: // linux, etc.` 만 존재 — **Windows 는 default(linux) 분기로 떨어져** `/usr/local/bin`, `/usr/local/sbin` 을 받는다
- :82 의 무조건 POSIX 꼬리 `"/usr/bin", "/bin", "/usr/sbin", "/sbin"` 은 GOOS 분기 밖이라 모든 플랫폼에서 추가된다
- 사용자 항목 2개는 `filepath.Join(homeDir, ...)` — Windows 에서 `C:\Users\<u>\...` 로 올바른 형태

→ Windows 조립 결과가 제보 값과 바이트 단위로 일치한다. `settings_test.go` 전체에 `case "windows"` 사례가 0개다(RED-1 측정: `grep -c 'case "windows"'` = 0 @ 6a46c0edb).

### B.2 Claude Code `env` 의미론 (공식 문서, 2026-09-07 조회)

- `env.PATH` 는 상속된 PATH 를 **치환(replace)**한다: "Claude Code writes each `env` entry into the process environment, replacing the value inherited from the shell" (code.claude.com/docs/en/env-vars). 확장 문법(`%VAR%`, `${env:PATH}`)은 문서화되어 있지 않고 값은 plain text 다. PATH 는 "Variables Claude Code ignores in `env`" 목록에 없다.
- 훅 자식은 부모 환경을 물려받는다(OTEL_* / CLAUDE_CODE_SUBPROCESS_ENV_SCRUB 예외 외). 즉 env.PATH 가 훅 자식이 보는 전체 PATH 다.
- **exec form** 훅은 `command` 를 그 PATH 에서 직접 해석하며 셸이 없다: "Claude Code resolves `command` as an executable on `PATH` and spawns it directly"; "On Windows, exec form requires `command` to resolve to a real executable such as a `.exe`" (code.claude.com/docs/en/hooks).
- 셸 form 은 Windows 에서 Git Bash 를 쓴다(없으면 PowerShell 폴백). CC setup 문서는 Git for Windows 설치를 권장한다.

### B.3 쓰기 표면 — 5 호출점 + 템플릿 (전수, grep 검증 @ 6a46c0edb)

1. `internal/core/project/initializer.go:412` — moai init → 프로젝트 settings (템플릿 경유)
2. `internal/cli/update.go:1019` — `ensureGlobalSettingsEnv` → 사용자 전역 `~/.claude/settings.json` (#598: delete-then-refresh 정책)
3. `internal/cli/update_template_sync.go:275` — template-sync 배포
4. `internal/cli/update_template_sync.go:335` — template-sync 배포
5. `internal/cli/update_clean_install.go:449` — clean reinstall
6. `internal/template/templates/.claude/settings.json.tmpl:418` — `"PATH": "{{jsonEscape .SmartPATH}}"`

생성기 내부를 고치면 5 호출점이 한 번에 수리된다. settings.local.json 작성자(glm.go:976, launcher.go:375 등)는 PATH 를 만지지 않는다(렌즈 전수 확인). 템플릿에 `.GoBinPath` 굽기는 없다(REQ-SWF-011 제거됨) — **env.PATH 가 훅의 `bash`·`moai` 해석을 지배하는 유일 표면이다**.

### B.4 계보 — 의도적 결정들 (drift 아님)

- **#467**: SmartPATH 탄생("터미널 PATH 굽기 금지, 안정적 플랫폼별 목록"). `TestBuildSmartPATH_StableAcrossTerminalPATH` 가 핀.
- **#495**: WSL2 `/mnt/` 상호운용 — `IsWSL2()` + GOOS=linux 게이트가 OS 분기 추가의 저장소 내 선례. #495 제보자는 "PATH 주입 자체를 생략"을 제안했으나 수리는 게이트형이었다.
- **#598**: update.go 의 delete-then-refresh — 전역 PATH 부활은 테스트까지 핀된 의도적 정책(`TestEnsureGlobalSettingsEnv/CleanupMoaiManagedKeys`). 프로젝트 레벨 삭제는 3-way merge 로 존중, 전역 레벨은 부활 — 기록된 비대칭.
- `settings_test.go:244` 주석 "renders identically on every platform, so no per-OS branch is needed" — 본 SPEC 이 windows 분기를 추가하면 이 서술은 정정 대상이다(annotated correction, 원문 유지).
- SPEC-UPDATE-GUARD-EFFICACY-001 은 `BuildSmartPATH` 를 "네 번째 실제-홈 판독자"(os.UserHomeDir 직접 호출, 주입 이음매 부재)로 추후 후보 지명 — 본 SPEC 의 리팩터링이 그 이음매를 착지시킨다.

### B.5 제보자 실측 (GH #1690 — darwin 에서 재현 불가, 미검증이지 반증이 아님)

- Git for Windows 기본 설치는 시스템 PATH(HKLM)에 `Git\cmd` 만 등록 — `bash.exe` 는 없음(bash 는 `Git\bin` 에 있음)
- `go\bin` 디렉터가 실재하지 않음, `.local\bin` 은 시스템 PATH 에 이미 존재
- 우회: settings.json 3곳에서 env.PATH 삭제 + HKCU Path 맨 앞에 `C:\Program Files\Git\bin` 추가 → 훅 정상
- **주의**: 우회 (a)(env.PATH 생략)가 통과한 것은 레지스트리를 함께 고쳤기 때문이다 — 생략만 하면 훅 자식은 시스템 PATH(기본 설치엔 bash 없음)를 물려받는다. 따라서 생략형 수리는 불충분하고, Windows 형태의 올바른 값(실존 bash 디렉터 포함)이 필요하다.

### B.6 재현 가능 / 재현 불가 분할 [HARD]

- **재현 가능(darwin, 코드·주입 테스트로)**: 생성 로직이 무엇을 조립하는가, GOOS 주입 시 각 플랫폼 문자열, 구분자 선택, 실제 템플릿 렌더 결과(jsonEscape), POSIX 혼입 코드 경로의 실재
- **재현 불가(본 머신에서 관측 불가)**: Windows 에서 Claude Code 가 exec form 훅을 실제로 해석해 실행하는 동작, 실제 Git Bash 존재 동태, 제보자의 정확한 환경
- 후자는 전부 **"미검증이지 반증이 아님"** 상태로 기록하며, 본 SPEC 의 어떤 기록도 "Windows 에서 재현했다"고 쓰지 않는다.

## §C Requirements (GEARS)

### REQ-CWSP-001 (Ubiquitous — GOOS 주입 생성기)

The system shall generate settings PATH via an internal, injection-parameterized function — `buildSmartPATHFor(goos string, home string, envLookup func(string) string, stat func(string) bool, wsl2 bool) string` — with `BuildSmartPATH()` preserved as the thin wrapper supplying `runtime.GOOS`, `os.UserHomeDir()`/`HOME`, `os.Getenv`, `os.Stat`, and `IsWSL2()`. The function shall carry a `case "windows"` branch that appends: the two home entries (as produced by `filepath.Join`), `%SystemRoot%\System32` (via `envLookup("SystemRoot")`, only when set), and each EXISTING Git Bash candidate (`C:\Program Files\Git\bin`, `C:\Program Files (x86)\Git\bin`, `%LOCALAPPDATA%\Programs\Git\bin` — appended only when `stat` reports existence), and shall NEVER append the POSIX package-manager or system-tail entries under windows. The join separator remains `os.PathListSeparator`. The darwin and linux branches shall produce output byte-identical to the pre-change implementation (fixture captured at M1 start, provenance-commented). `[AC-CWSP-001, AC-CWSP-002, AC-CWSP-003, AC-CWSP-007]`

### REQ-CWSP-002 (Event-driven — windows 렌더)

When the template context carries a windows-generated SmartPATH, rendering `settings.json.tmpl` shall emit `"PATH"` as a `;`-joined Windows-form value with no POSIX entry, through the real embedded template and its `{{jsonEscape .SmartPATH}}` directive; the directive shall continue to appear exactly once (AC-TPS-014 sentinel). `[AC-CWSP-004]`

### REQ-CWSP-003 (State-driven — 프로브 semantics)

While a Git Bash candidate directory does not exist (`stat` false), that candidate shall not appear in the generated windows PATH; when no candidate exists, the windows PATH shall still carry the home entries (never empty, never POSIX). `[AC-CWSP-002, AC-CWSP-007]`

### REQ-CWSP-004 (Ubiquitous — 기존 불변 보존)

The existing invariant suites shall pass unchanged after the fix: the WSL2 suite (GOOS=linux gate untouched), `TestBuildSmartPATH_StableAcrossTerminalPATH` (#467), `TestBuildSmartPATH_EssentialDirs` (home entries on every platform — windows retains both), `TestSettingsTemplateRequiredEnvVars` (requiredKeys incl. PATH — the template is not modified), `TestSettingsTemplateHookExecForm`, and the jsonEscape exactly-once sentinel (AC-TPS-014 — verified separately in M3, outside AC-CWSP-005's command scope per the F1 repair). The `settings_test.go` comment "no per-OS branch is needed" (:244) shall receive an annotated correction (original kept) whose scope is LIMITED to the env.PATH generation aspect this SPEC changes — the :744 sibling statement about exec-form hook rendering remains true and untouched. `[AC-CWSP-005]`

### REQ-CWSP-005 (Capability-gate — 수리 검증 방식)

Verification of the windows fix shall be by GOOS-injected exact-string table tests and a windows-realistic render test — NOT by `GOOS=windows` cross-build alone, which compiles no tests and is admitted only as a smoke check. `[AC-CWSP-002, AC-CWSP-004, AC-CWSP-006]`

### REQ-CWSP-006 (Capability-gate — 재현성 기록 규율)

Where the SPEC's verdict or progress records claims about Windows behavior, it shall split them into darwin-reproducible facts (generator logic, injected strings, render output) and Windows-unreproducible observations (actual Claude Code exec resolution, real Git Bash dynamics, the reporter's environment), marking the latter "미검증이지 반증이 아님" — and shall never state that Windows behavior was reproduced. `[AC-CWSP-008]`

## §D Acceptance Criteria (two-cell; RED-now measured this session on tree 6a46c0edb unless noted)

### AC-CWSP-001 — windows 분기 존재

- command: `grep -c 'case "windows"' internal/template/settings.go`
- RED-now: `0`, exit 1 (@ 6a46c0edb, measured 2026-09-07). Why red: the id is introduced by this SPEC — 0 means the branch is absent, the exact discriminator for the defect's shape.
- green path: M1 adds the branch → `≥1`, exit 0.

### AC-CWSP-002 — windows 정확-문자열 테이블

- command: `go test ./internal/template/ -run TestBuildSmartPATHWindows -count=1`
- exit code: recorded per run (RED-now: 0; green: 0 — the discrimination lives in the executed-rows evidence below, not the exit code).
- RED-now: stdout contains the observed empty-sweep token `[no tests to run]` with `ok` (vacuous — the sweep selects nothing; the test does not exist today). Why red-shaped: the recorded pre-state, not a pass.
- green path: M2 adds the table (windows rows asserting the byte-exact string — POSIX entries absent, `;` separator, probe-reflecting, AND a never-empty row asserting the all-probes-false scenario still yields the home entries) → `--- PASS` with rows executed. Mutants that re-add a POSIX entry under windows or ignore the stat probe MUST fail these rows.

### AC-CWSP-003 — darwin/linux 바이트 안정성

- command: `go test ./internal/template/ -run 'TestBuildSmartPATH_StableAcrossTerminalPATH|TestBuildSmartPATH_WSL2' -count=1`
- RED-now: `ok` (@ 6a46c0edb, measured 2026-09-07) — the pre-change suites are green and stay green (non-regression guard; co-observed with AC-CWSP-001's 0→≥1 flip on the same file so it is not vacuous).
- green path: unchanged suites stay ok AND the new table's darwin/linux rows assert byte-equality with the M1-captured fixture (provenance: captured from the un-refactored function at M1 start).

### AC-CWSP-004 — windows-realistic 렌더

- command: `go test ./internal/template/ -run TestSettingsRenderWindowsPATH -count=1`
- RED-now: `no tests to run` (absent today; the only windows-shaped render coverage is settings_security_scan_entry_test.go:20 which feeds a POSIX fixture to the windows platform — the recorded coverage hole).
- green path: M2 adds the test — a windows-shaped `;`-joined value rendered through the real embedded template produces valid JSON with no POSIX entry; `TestJsonEscapeInTemplate` (renderer_test.go:307) stays green. (The AC-TPS-014 exactly-once sentinel — `TestTemplateDirectivePreserved`, internal/config/toolpolicy — is OUTSIDE this SPEC's AC command scope; it is verified separately in M3 per the F1 repair, see §E.)

### AC-CWSP-005 — 트립와이어 전수

- command: `go test ./internal/template/ -run 'TestBuildSmartPATH|TestIsWSL2|TestIsUserScopedWindowsPath|TestIsWSL2DrivePath|TestSettingsTemplateRequiredEnvVars|TestSettingsTemplateHookExecForm' -count=1`
- RED-now: `ok` (@ 6a46c0edb, measured 2026-09-07 — the pinned invariants hold today and must survive).
- green path: same single invocation after M2/M3 → `ok`.

### AC-CWSP-006 — 크로스빌드 smoke + 패키지 전체

- command (split — two observables, not one): (a) `GOOS=windows go build ./...` (smoke only — compiles no tests); (b) `go test ./internal/template/... -count=1`
- RED-now: (a) rc 0; (b) `ok` @ 6a46c0edb.
- green path: (a) rc 0 AND (b) `ok` after the change.

### AC-CWSP-007 — 프로브 semantics 깊이

- command: covered by AC-CWSP-002 rows with `stat=false` scenarios
- RED-now: n/a (new-test AC — its RED is AC-CWSP-001's absence grep; this criterion pins the probe branch: a candidate with `stat=false` must not appear).
- green path: table rows assert absence under false probe; a mutant forcing probe-true fails AC-CWSP-002's exact rows.

### AC-CWSP-008 — 재현성 기록 규율

- command: `grep -c '재현했다\|reproduced on Windows' .moai/reports/t515/verdict.md` AND `grep -c '미검증이지 반증' .moai/reports/t515/verdict.md` (two probes, bidirectional)
- RED-now: the verdict does not exist yet — the count probes return file-not-found / 0 (recorded pre-state).
- green path: M4's verdict carries the §B.6 split — the forbidden-phrase probe stays 0 AND the split-marker probe is ≥1 (a verdict silent about reproducibility entirely fails the second probe; both directions are observed).

### Mutant probe record (adoption-depth)

- Mutant A — re-add a POSIX entry under the windows branch: AC-CWSP-002 exact rows fail.
- Mutant B — ignore the stat probe (append all candidates): AC-CWSP-002 rows (probe=false scenario) fail.
- Mutant C — alter the darwin/linux branches: AC-CWSP-003 fixture rows fail (byte-inequality).

## §E Evidence Ledger (5-section)

- **Claim**: 본 SPEC 은 settings.json env.PATH 의 Windows 파손(GH #1690)을 생성기 단일 수리로 닫는다 — GOOS 주입 생성기 + windows 분기(홈 2항목 + SystemRoot\System32 + 실존 프로브된 Git Bash 후보, POSIX 배제) + 정확-문자열 테이블 + windows 렌더 테스트. 5 호출점이 래퍼를 통해 자동 수리된다.
- **Evidence**: §B 전체 + RED-now 측정(본 문서 §D, 트리 6a46c0edb, 2026-09-07 직접 측정) + `.moai/reports/t515/{research-fanout-md.md, lens-external-docs.md, lens-prior-SPEC-memory.md, lens-codebase-precedent.md}` + GH #1690 본문.
- **Baseline-attribution**: worktree `.claude/worktrees/t515` @ `WT-win-path-env`, HEAD `6a46c0edb`, 2026-09-07.
- **Gaps**: Windows 실제 동작(exec 해석·Git Bash 동태·제보자 환경) 미재현 — §B.6 분할 참조. CI 의 windows 러너 존재 여부 미확인. `%VAR%` 비확장은 문서 부재 논증+단일 프로젝트 실측(§22.4) — 폐쇄 소스 CLI 대비 본질적 한계. **결합 기록(F2 수리)**: `TestEnsureGlobalSettingsEnv_FallbackPATHForNonMoaiDirs`(internal/cli/update_test.go:1947-54)가 GOOS 게이트 없이 `/usr/bin`+`/bin` 을 핀한다 — 본 SPEC 의 windows 분기는 그 테스트와 결합되어, 미래에 windows CI 러너가 붙으면 상속 red 를 만든다. 수리는 본 SPEC 범위 밖(update.go 정책 #598 축)이나, 결합의 기록 의무는 여기서 이행한다.
- **Residual-risk**: CC 의 env 적용 의미론은 문서 기반으로 향후 버전에서 변할 수 있다. Git Bash 후보 목록은 일반 설치 위치 3종 — 스쿱/msys2 등 비표준 설치는 프로브 미적중으로 홈 2항목+SystemRoot 만 렌더된다(POSIX 는 여전히 배제되므로 본 결함 축은 닫인다).

## Out of Scope

세부 제외는 `### Out of Scope — <topic>` h3 하위섹션으로 기술한다.

### Out of Scope — update.go 부활 정책(제보 (c))

`moai update` 의 전역 PATH delete-then-refresh 는 #598 의 의도적·테스트 핀 정책이다. 값이 올바르면 부활은 무해해진다(사용자가 지운 것은 깨진 값이었다). 정책 재협상은 별도 결정 사항.

### Out of Scope — 템플릿 훅 형태 변경

exec form("command": "bash") 유지. 셸 form 전환·PowerShell 폴백 설계는 CC 문서 축의 별도 사안.

### Out of Scope — darwin/linux 출력 변경

두 플랫폼 출력은 사전 상태와 바이트 동일이어야 한다(AC-CWSP-003이 금지).

### Out of Scope — WSL2 동작 변경

`IsWSL2()`·`/mnt/` 캡처·필터는 건드리지 않는다(#495 회귀 금지).

### Out of Scope — CI windows 러너

GitHub Actions windows 매트릭스 유무는 본 카드에서 확인하지 않는다(갭으로 기록).

### Out of Scope — CLAUDE_ENV_FILE 계열 대안

Bash tool 전용 탈출구이며 훅에는 문서화된 등가물이 없다 — 채택 검토 자체가 본 카드 밖.
