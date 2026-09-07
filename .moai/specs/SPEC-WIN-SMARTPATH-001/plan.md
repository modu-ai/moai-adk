# plan.md — SPEC-WIN-SMARTPATH-001

Card t515 (GH #1690) · worktree `.claude/worktrees/t515` · branch `WT-win-path-env` · base `6a46c0edb` (local develop tip)

## §A Milestones

### M1 — 생성기 리팩터링 + windows 분기

1. **선행 캡처**: 무수정 `BuildSmartPATH()`의 darwin 실출력을 픽스처 상수로 기록(테스트 파일에 provenance 주석: "captured from un-refactored BuildSmartPATH at 6a46c0edb, darwin, HOME=<fixture>").
2. `buildSmartPATHFor(goos, home string, envLookup func(string) string, stat func(string) bool, wsl2 bool) string` 추출 — 기존 분기 로직 이식, 판독은 전부 인자 경유.
3. `case "windows"` 추가: home 2항목 + `%SystemRoot%\System32`(envLookup("SystemRoot") 설정 시) + Git Bash 후보 3곳 중 `stat` true 만. POSIX 패키지 매니저·시스템 꼬리 항목 없음.
4. `BuildSmartPATH()` 래퍼: `runtime.GOOS`, `os.UserHomeDir()`/`HOME`, `os.Getenv`, `os.Stat`, `IsWSL2()` 주입.

### M2 — 테이블 + 렌더 테스트

1. `TestBuildSmartPATHWindows` — windows 정확-문자열 행(프로브 true/false, SystemRoot 유무) `;` 조립·POSIX 부재 바이트 단언.
2. darwin/linux 행 — M1 픽스처와 바이트 동일 단언(비회귀).
3. `TestSettingsRenderWindowsPATH` — windows 형태 SmartPATH 를 실제 임베디드 `settings.json.tmpl` 로 렌더, JSON 유효 + POSIX 부재 + jsonEscape 정상.
4. `settings_test.go:244` "no per-OS branch" 주석 annotated correction(원문 유지) — 스코프는 env.PATH 생성기 측면으로 한정(F6: :744 형제 exec-form 서술은 여전히 참, 무접촉).

### M3 — 트립와이어 + smoke

1. AC-CWSP-005 단일 호출 전수(windows 사례 포함 확인 — empty-sweep 토큰 없음을 결과에서 확인).
2. `GOOS=windows go build ./...` — smoke only(테스트 컴파일 없음을 plan 에 명시; 판정은 M2 몫).
3. `go test ./internal/template/... -count=1` 전체 ok.
4. `go test ./internal/config/toolpolicy/ -run TestTemplateDirectivePreserved -count=1` — AC-TPS-014 jsonEscape exactly-once 센티넬은 본 패키지 AC 명령 범위 밖(F1 수리)이라 별도 확인.

### M4 — 증거/판정 기록

`.moai/reports/t515/verdict.md` — 5섹션 + §B.6 재현성 분할 이행(AC-CWSP-008), progress.md §E.2/§E.3.

## §B Delta markers + PRESERVE

- `[MODIFY] internal/template/settings.go` — 유일 소스 수정(리팩터링+분기)
- `[MODIFY] internal/template/settings_test.go` — 테이블·렌더 테스트·:244 주석 정정
- `[EXISTING-UNTOUCHED] internal/template/templates/**` (settings.json.tmpl 포함 — `{{jsonEscape .SmartPATH}}` 불변)
- `[EXISTING-UNTOUCHED] internal/cli/update.go`·`update_template_sync.go`·`update_clean_install.go`·`initializer.go` (래퍼 시그니처 불변으로 자동 커버)
- `[EXISTING-UNTOUCHED] WSL2 로직(IsWSL2·isWSL2DrivePath·isUserScopedWindowsPath)·darwin/linux 분기 출력`

## §C Risks

| 리스크 | 완화 |
|---|---|
| 프로브가 환경별로 다른 결과 → 렌더가 기계별로 달라짐(#467 위화) | 프로브 대상은 기계 클래스 경로(Program Files 계열) — 사용자 경로 아님. hardcoded_path_audit_test 가 금지하는 것은 사용자 절대경로. 프로브 false 시 미채택이라 굽기 없음 |
| EssentialDirs 핀(home 2항목 every platform) 충돌 | windows 도 home 2항목 유지 — 핀 성립 유지. :244 주석만 정정 |
| GOOS 주입 함수와 래퍼 불일치 | 래퍼는 인자만 공급하는 얇은 층 — table test 가 래퍼 경유 출력도 함께 단언 |
| 크로스빌드를 수리 판정으로 오용 | plan 명시: smoke 만. 판정은 AC-CWSP-002/004 |
| jsonEscape 센티넬 훼손 | 템플릿 무변경 + M3 의 별도 toolpolicy 명령(F1 수리 — AC-005 범위 밖임을 명시) + AC-CWSP-004 는 TestJsonEscapeInTemplate 로 한정 인용 |
| :244 정정의 과잉 서술 | 정정문 스코프를 env.PATH 생성기 측면으로 한정(F6 수리) — :744 형제(exec-form 서술)는 본 SPEC 후에도 참이라 손대지 않음 |
| 생 성 Go 파일 | 테이블 데이터 실수 — 정확 문자열 단언이 곧 검증 |

## §D MX Tag Plan

- `buildSmartPATHFor` 신규 이음매: `@MX:ANCHOR` (fan_in 5 — initializer·update·template-sync×2·clean-install) + `@MX:REASON` (fan-in 계산치) + `@MX:SPEC:SPEC-WIN-SMARTPATH-001`
- windows 분기: `@MX:NOTE` (왜 POSIX 배제인지 — #1690 계보 1줄)
- 기존 태그 이동/삭제 없음. 동기화 보고에 1 added / 0 removed 기록.

## §E AC ↔ Milestone map

| AC | M1 | M2 | M3 | M4 |
|---|---|---|---|---|
| AC-CWSP-001 | ● | | | |
| AC-CWSP-002 | | ● | | |
| AC-CWSP-003 | ●(픽스처) | ●(행) | ●(전수) | |
| AC-CWSP-004 | | ● | | |
| AC-CWSP-005 | | | ● | |
| AC-CWSP-006 | | | ● | |
| AC-CWSP-007 | ●(probe 주입) | ●(행) | | |
| AC-CWSP-008 | | | | ● |

## §F 금지 목록

- `internal/template/templates/**` 무변경 (Template-First 상 템플릿 소스 편집 없음 → `make agents-emit` 불요, `make build` 는 바이너리 재컴파일용으로만)
- `internal/cli/**` 무변경 (래퍼 시그니처 불변)
- WSL2 함수·darwin/linux 출력 무변경
- 커밋은 명시적 pathspec, HEAD 재판독 동반, push 금지(레인 규율), `🗿 MoAI` 푸터
- 측정 인용에 트리+HEAD 병기, 파이프 뒤 `$?` 금지
