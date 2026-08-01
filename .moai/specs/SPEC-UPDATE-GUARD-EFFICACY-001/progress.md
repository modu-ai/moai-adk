# SPEC-UPDATE-GUARD-EFFICACY-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase 산출물 3종(spec.md / plan.md / acceptance.md)이 작성되었다. Tier M.

**Plan-phase에서 실제로 실행해 관측한 것** (`.claude/worktrees/e2-data-survival`, PR #1275 브랜치 HEAD `457e1013f`):

- 가드 스코프 커버리지 4종 — `CleanMoaiManagedPaths 66.7%` / `MigrateLegacyMemoryDir 26.9%` / `runCleanReinstall 0.0%` / `BackupMoaiConfig 0.0%`
- update 서브시스템의 비이음매 HOME 호출부 3곳 (`update_clean_install.go` 1건, `update_template_sync.go` 2건) + 범위 밖 `glm.go` 1건
- `snapshotDir` 사용처가 `update_safety_test.go` 한 파일뿐임
- `mergeBackPreserveInventory`의 stat 실패 분기가 `chmod 0o000` + `runtime.GOOS`/`os.Geteuid` 스킵 가드에 의존함
- SPEC ID 정규식 자기 점검: `PASS`

**Plan-phase에서 정정한 선행 기록**: D1의 "`runCleanReinstall`·`BackupMoaiConfig` 0.0%"는 가드 스코프 한정 수치이며, 두 함수 모두 패키지 전체 테스트에서는 이미 구동된다. 진짜 공백은 커버리지가 아니라 사용자 소유 디렉터리 보존 어서션의 도달 범위다 (spec.md §1.3, plan.md §A.1).

### 개정 라운드 1 — plan-auditor FAIL 0.71 (Testability 0.50) 대응

감사 지적 대응과 별개로, **판정 명령을 실제로 실행해** 다음을 확인·정정했다.

**실행해서 잡은 결함 (감사가 지적하지 않은 것 포함)**

| 무엇 | 관측 | 조치 |
|---|---|---|
| AC-UGE-012 변형이 무동작 | 리터럴 `.claude/skills/moai*` 부재(`grep -c` → `0`), `mutation-applied=0` | `defs.SkillsSubdir, "moai*"` 겨냥으로 정정. 재실행: `mutation-applied=1`, 가드 FAIL 계수 `4` |
| AC-UGE-015가 **주석**을 위반으로 계수 | `update_home_radius_test.go 1` — 정체는 `:17`의 "MUST NOT call t.Parallel()" 주석 | 주석 제외로 정정. 재실행: 두 파일 모두 `0` |
| `runCleanReinstall`이 `CleanMoaiManagedPaths`를 부르지 않음 | 해당 파일 grep 0매치 | M4 reinstall 가드의 반증을 `scanDeprecatedPaths` 겨냥으로 재설계 |
| `BackupMoaiConfig` 주 경로가 순수 복사 | `backup.go`에 삭제는 실패 롤백과 `CleanupOldBackups` 회전뿐 | REQ-UGE-010을 두 파괴적 표면으로 재정박 |
| 세 HOME 호출부에 삭제 없음 | `homeDir` → `detectGoBinPathForUpdate` + `WithHomeDir`만 | REQ-UGE-004/005 근거를 테스트 격리로 재작성 |

**실행해서 통과 확인한 판정**

- AC-UGE-007 sentinel HOME 종단간: `before-nonempty=0` / `compile-exit=0` / `--- PASS: TestEnsureGlobalSettingsEnv_HooksRemovalRadius (0.02s)` / `diff-exit=0`
- AC-UGE-008 카나리아: `mutation-applied=1` / `compile-exit=0` / `probe-survived=1`(probe 삭제됨)
- AC-UGE-009F 변형: `mutation-applied=1`, `update_cleanup.go:145`에 1행 삽입
- AC-UGE-010F 변형: `mutation-applied=1`, `backup.go:257`, 변형 트리 `go build` exit 0
- AC-UGE-003 (0) base 위생: `3`
- AC-UGE-013 템플릿 중립성: `0` / `0`
- baseline: `golangci-lint` → `0 issues.`, `merge-base` → `83610e03e`

**plan-phase에서 실행할 수 없었던 것**

- AC-UGE-001/002/004/005(변경 후 방향)/006/006F/006R/009/010/011/011F — 대상 코드·테스트가 아직 없다(run-phase 산출물).
- AC-UGE-002의 Windows **런타임** 통과 — 이 머신은 macOS. `GOOS=windows go vet`까지만 가능하며 런타임은 CI 인용.
- AC-UGE-015의 GNU `xargs` 재현 — 이 머신은 BSD `xargs`. `xargs` 자체를 제거해 차이를 무의미하게 만드는 방식으로 우회했다.
- AC-UGE-011F 변형의 컴파일 — 대상 가드가 없어 미실시. AC에 컴파일 실패 시 등가 변형 지침을 명시했다.

### 개정 라운드 2 — plan-auditor PASS-WITH-DEBT 0.89 대응 (3차 감사 없이 run-phase 진입)

| 지적 | 조치 | 이 트리에서의 관측 |
|---|---|---|
| N1 반증이 원인을 귀속하지 않음 | §A.4에 **원인 귀속 + 경쟁 원인 배제** 의무 신설, 반증 4건(006F/009F/010F/011F)에 적용 | 009F 교란 가설 실측: `backup-stage-abort=0`, `fail-lines=0` — 경로는 실재하나 이 변형에선 미발화 |
| N2 스파이가 REQ-UGE-005를 측정 못함 | AC-UGE-006에 `TemplateContext.HomeDir` 출력 어서션 `(d)` 추가 (스파이는 도달 증거로 유지) | `context.go:239` 단순 대입, `context_test.go:324` 기존 직접 어서트 확인 |
| N3 `xargs` 기전 서술 오류 | "행(hang)" → "stdin이 `/dev/null`로 리다이렉트되어 즉시 `0` 출력" 으로 정정 | GNU 재현 불가(BSD 머신) — 미확인 표기 유지 |
| N4 `while read` 0-매치 공허 | `target-files` 계수 선행 검사 추가 | `target-files=2`. 음성 대조(없는 패턴) → `target-files=0` 으로 공허 탐지 확인 |
| N5 주석 필터가 줄 앵커 한정 | `sed 's\|//.*\|\|'` 로 꼬리 주석까지 제거 | 두 파일 모두 `0`. 블록 주석·문자열 리터럴은 잔여 한계로 명시 |
| N6 DoD 목록 19개 vs 파일 20개 | AC-UGE-004 추가 + 계수 검사 명령 추가 | `grep -c` → `20`, DoD 목록과 파일 집합 `diff` 일치 |
| N7 개정에 버전·HISTORY 없음 | `version: 0.3.0` + HISTORY 2행 추가 | — |

**감사가 대신 해결해 준 것**: AC-UGE-011F 변형의 컴파일 가능성(`go build -overlay` exit 0). 직전 라운드에서 미확인으로 넘겼던 항목이 해소되었다.

### run-phase로 넘기는 미검증 항목 (3차 감사 없음 — 여기가 마지막 plan 기록)

1. **신규·수정 가드 6개 전부의 실제 PASS/FAIL.** 대상 코드가 없어 plan-phase에서 관측 불가. 반증 4건의 FAIL 관측도 마찬가지로 run-phase 몫이다.
2. **AC-UGE-009F의 교란 여부 (신규 픽스처 조건에서).** 위 실측은 `.moai/harness`가 **없는** 기존 픽스처 기준이다. 신규 가드는 그 경로를 실제로 생성하므로 백업 단계 동작이 달라진다 — `backup-stage-abort=0`을 그 조건에서 다시 확인해야 한다.
3. **Windows 런타임.** `GOOS=windows go vet`까지만 가능. AC-UGE-002의 Windows 통과는 CI 결과 인용 필요.
4. **GNU `xargs` 실제 동작.** BSD 머신이라 미재현. 채택한 수정이 `xargs`를 제거하므로 판정은 안전하지만, N3의 기전 서술 자체는 독립 관측이 아니다.
5. **`go test ./...` 전체와 `-race`.** 이번 라운드에서 실행하지 않았다. AC-UGE-014의 FAIL 계수 항목은 **미측정**이다.
6. **`TemplateContext.HomeDir`의 기계 독립성이 전체 렌더 파이프라인 끝까지 유지되는지.** 필드 대입(`context.go:239`)과 기존 직접 어서트(`context_test.go:324`)까지만 확인했다. 렌더 산출물에 그 값이 어떻게 나타나는지는 미추적.
7. **AC-UGE-015의 블록 주석·문자열 리터럴 오탐.** 현 트리에 해당 사례가 없어 계수 `0`이지만, Go 파서 없이는 원리적으로 못 거른다. 위반 발견 시 줄 번호 육안 확인 필요.
8. **AC-UGE-011F 변형의 원인 귀속 실측.** 컴파일 가능성은 확인됐으나, `names-missing-backup`/`rename-subtest-failed` 값은 가드 부재로 미관측.

**선행 조건**: PR #1275는 머지 완료(`83610e03e`)이고 이 브랜치에 머지 반영했다. run-phase는 plan.md §C1의 base 위생 검사를 **진입 시점에 재실행**한다.

### plan-auditor 독립 감사 (iteration 1)

**판정: FAIL** — 종합 0.71 (Tier M 통과선 0.80). Must-pass 7항목은 전부 통과했고, FAIL은 판정 명령의 실효성 부족에서 나온 종합 점수 미달이다.

| 차원 | 점수 | 근거 |
|---|---|---|
| Clarity | 0.75 | REQ-UGE-004·007의 서술된 근거가 코드와 불일치 |
| Completeness | 0.85 | Tier M 3종 + 12필드 + Out of Scope H3 4개 완비 |
| Testability | 0.50 | AC-UGE-012 무동작, 009/010 (d) 약한 대리, 003 교란, 006 공허 |
| Traceability | 0.75 | 13/13 매핑되나 REQ-UGE-013이 §D와 모순 |

Must-pass: MP-1 PASS(REQ-UGE-001~013 연속·중복 0) · MP-2 PASS(13/13 shall 형식) · MP-3 PASS(12필드 canonical) · MP-4 N/A(단일 언어) · MP-5 PASS(`SPEC-UPDATE-DATA-SURVIVAL-001` = `completed` on origin/main) · MP-6 N/A(`syscall` 0건) · MP-7 PASS(`[NEEDS CLARIFICATION]` 0건).

차단 결함 6건 (F1~F6):

1. **F1 — AC-UGE-012의 sed 변형이 무동작 (실행 확인).** `internal/cli/update/deploy/deploy.go`에 리터럴 `.claude/skills/moai*`가 없다. 실제 코드는 `filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*")` (53·54행). 실행 결과 `mutation-applied=0`. REQ-UGE-012의 반증이 실행 불가능하다. 수정: 변형 대상을 `defs.SkillsSubdir, "moai\*"` → `defs.SkillsSubdir, "*"` 로 교체(2행 치환), `mutation-applied=1` 재확인.
2. **F2 — REQ-UGE-013과 §D 추적 매트릭스가 모순.** REQ-UGE-013은 "각 가드"의 실패 관측을 요구하나, §D는 REQ-UGE-013 → AC-UGE-003/008/012만 매핑한다. M4가 신설하는 가드 3종(009/010/011)에는 반증이 하나도 없다. 수정: 세 가드 각각에 변형 반증을 추가하거나 REQ-UGE-013의 범위를 명시적으로 축소.
3. **F3 — AC-UGE-009/010의 (d)가 REQ를 구속하지 못한다.** `grep -cE 'snapshotDir|harness' >= 2`는 스냅샷을 *비교*하는지 검증하지 않는다. spec.md §1.3이 스스로 "진짜 공백은 바이트 동일성 어서션의 도달 범위"라고 정정했는데, (d)는 토큰 등장만 센다. (c)의 커버리지 비교도 baseline과 다른 선택자를 쓰므로 "새 테스트가 그 함수를 부른다"는 (b)의 재확인에 그친다. 수정: 스냅샷 map 비교(`mapsEqual`)의 실패 경로를 변형 반증으로 관측.
4. **F4 — AC-UGE-003의 overlay 반증이 교란된다 (실측).** 현재 `git merge-base origin/main HEAD` = `a64548a2a`(#1275 이전)이며 그 시점 파일에는 `restored %d/%d`가 0건이다. overlay로 되돌리면 M1 이음매 부재가 아니라 **M6 보고 문구 부재** 때문에 FAIL이 난다. §C1 리베이스가 완화하지만 AC 자체에 BASE 건전성 가드가 없다. 수정: `git show "$BASE":...update_preserve_inventory.go | grep -c 'restored %d/%d'`가 3인지 먼저 확인.
5. **F5 — M2→M3 의존이 실재하지 않는다 (코드 확인).** M2가 이음매화하는 세 호출부(`update_clean_install.go:410`, `update_template_sync.go:227·272`)는 모두 `template.WithHomeDir(homeDir)`로 **템플릿 렌더링에 HOME을 읽기만** 한다 — 삭제 반경(`os.RemoveAll`)이 없다. 반면 AC-UGE-007은 `TestEnsureGlobalSettingsEnv_HooksRemovalRadius` 하나만 구동하므로 M2 유무와 무관하게 `ensureGlobalSettingsEnv` 한 곳만 덮는다. REQ-UGE-004의 "HOME 반경을 고정하려는 가드가 그 세 경로에 닿지 못한다"는 전제가 코드와 어긋난다. 수정: M2의 실제 편익(향후 템플릿 렌더링의 HOME 의존 테스트 가능성)으로 근거를 재서술하고, M2-M3 순서 강제를 철회하거나 세 호출부 도달을 실제로 검증하는 AC를 신설.
6. **F6 — AC-UGE-006이 REQ-UGE-005를 공허하게 검증한다.** 현재 `userHomeDirFn`을 재할당하는 테스트는 `glm_tools_test.go`·`update_home_radius_test.go` 2개뿐이고 둘 다 clean-reinstall/template-sync를 구동하지 않는다. 즉 M2가 바꾸는 동작을 관측하는 테스트가 0개이므로 전체 테스트 green은 "변화 없음"의 증거가 아니다. 수정: 주입 활성 상태에서 세 호출부를 지나는 테스트를 하나 추가해 동작 보존을 실측하거나, REQ-UGE-005의 검증 불가를 Gaps로 명시.

부수 결함 4건 (선택):

- **F7 — AC-UGE-015의 `xargs`가 GNU 환경에서 위험.** BSD xargs(macOS)는 빈 입력 시 명령을 실행하지 않지만(실측 확인), GNU xargs는 인자 없이 실행해 `grep`이 stdin을 기다린다. CI는 ubuntu다. 수정: `xargs -r` 또는 `grep -c ... {} \;` 형태.
- **F8 — AC-UGE-007 (e)가 NFR-UGE-002로 오표기.** NFR-UGE-002는 템플릿 중립성이며 `t.Setenv("HOME"` 금지와 무관하다. §D는 올바르게 AC-UGE-013에 매핑한다.
- **F9 — REQ-UGE-007의 근거가 전치 오류.** "before 스냅샷이 비어 있지 않아야 삭제 방향이 판별된다"는 AC-UDS-013(실제 HOME diff) 시절의 논리다. 새 설계에서 sentinel은 *양성* 경로에서 아예 건드려지지 않으며, 비어 있지 않아야 하는 이유는 오직 AC-UGE-008의 카나리아 역할 때문이다.
- **F10 — spec.md §1.1이 저작 시점에 이미 거짓.** PR #1275는 2026-08-01T15:56:27Z에 `83610e03e`로 머지되어 origin/main에 이미 존재한다(SPEC 작성 직전). §C1 절차 자체는 여전히 유효하다.

검증된 것 (실행 관측):
- origin/main 기준 baseline 3종: `update_clean_install.go` 1건 / `update_template_sync.go` 2건 / `glm.go` 1건, `mergeBackPreserveInventory` 본문 `os.Stat(` 1건, 이음매 부재 확인 — acceptance.md §B와 일치.
- §1.3 정정은 **참**: `snapshotDir(` 호출은 `update_safety_test.go` 1개 파일뿐이고, `update_clean_install_config_preserve_test.go`는 settings.json 키·user.yaml·language.yaml·design.yaml 등 *설정값*만 어서트한다.
- NFR-UGE-001 전제 참: `setupToolsTestHome` 47 호출부, `glm_tools_test.go`의 `t.Parallel()` 0건.
- AC-UGE-007/008 설계는 기계 독립성을 실제로 확보한다: `userHomeDir()`가 `$HOME`를 우선 읽으므로(`homedir.go`) 이음매 제거 변형은 sentinel HOME의 probe에 도달한다. 운영자 실제 홈은 건드리지 않는다.
- 범위 규율 준수: 대상 6파일 어느 것도 `internal/template/templates/` 아래가 아니다.
- `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` baseline exit 0.

미검증 (실행 불가):
- AC 판정 명령 전체의 실제 통과 여부 — M1~M4 구현이 아직 없다.
- §B의 `66.7%` / `26.9%` / `94.1%` 수치 — 이 워크트리는 #1275 이전 트리라 재현 불가(측정 시 `CleanMoaiManagedPaths 0.0%` / `MigrateLegacyMemoryDir 0.0%`가 나온다). §C2가 재측정을 의무화하므로 결함은 아니다.
- AC-UGE-002의 Windows 런타임 통과 — 저자가 §E DoD 4항에 정직하게 이월했고, (b)+(d)가 증명 범위를 정확히 한정한다. 은폐 없음.

### plan-auditor 독립 감사 (iteration 2 — 개정본 `0ef5012c4`)

**판정: PASS-WITH-DEBT** — 종합 **0.89** (Tier M 통과선 0.80). iter1 0.71 → iter2 0.89, 점수 회귀 없음(STOP 신호 미발동). 차단급 부채 2건(N1·N2)은 run-phase 진입 전 AC 문구 수정으로 해소 가능.

> 0.89 < 0.90 이므로 Phase 1 Plan Audit Gate **skip-eligible 아님** — run-phase 진입 시 게이트가 재실행된다.

| 차원 | iter1 | iter2 | 근거 |
|---|---|---|---|
| Clarity | 0.75 | 0.90 | 거짓이던 전제마다 정정 주석 + 코드 근거. 잔여: REQ-UGE-005와 AC-UGE-006의 의미 불일치(N2) |
| Completeness | 0.85 | 0.95 | §B에 lint·merge-base·base위생·글로브표현·회전슬라이스 행 추가, 수용 부채 절 신설, B6~B9 추가 |
| Testability | 0.50 | 0.80 | F1/F3/F4 해소를 실행으로 확인. 잔여: 반증 3건의 실패 귀속 부재(N1) 외 3건 |
| Traceability | 0.75 | 0.90 | 13/13 매핑 + 6-가드 반증 완전성 표. 잔여: DoD 열거 19 vs 실제 20(N6) |

Must-pass 7/7: MP-1 PASS(REQ-UGE-001~013, 13개 연속·중복 0) · MP-2 PASS(13/13 shall; REQ-UGE-005는 When+shall not 형태로 오히려 개선) · MP-3 PASS(12 canonical 필드, 거부 alias 0) · MP-4 N/A · MP-5 PASS(`SPEC-UPDATE-DATA-SURVIVAL-001` = `completed`, depends_on 충족) · MP-6 N/A · MP-7 PASS. (MP-6/MP-7 grep은 이 progress.md의 감사 기록 자체와 자기 매칭하므로 산출물 4종 본문 기준으로 판정했다.)

#### 저자가 보고한 3건 — 전부 실행으로 확인(내가 iter1에서 놓친 결함)

1. **`runCleanReinstall`은 `CleanMoaiManagedPaths`를 호출하지 않는다 — 확인.** `grep -n 'CleanMoaiManagedPaths\|MigrateLegacyMemoryDir' internal/cli/update_clean_install.go` → **0매치(exit 1)**. 글로브 확장 변형을 이 가드의 반증으로 재사용했다면 **공허**했을 것이다. 재조준한 파괴 표면도 실재: `update_clean_install.go:265` `scanDeprecatedPaths` → `:322` `os.RemoveAll(abs)`. iter1의 F3에서 나는 "`mapsEqual` 실패 경로를 변형으로 관측하라"고만 적고 어느 변형인지 지목하지 않아 이 공허성을 놓쳤다.
2. **`BackupMoaiConfig` 보존 가드는 구조적으로 준-공허 — 확인.** 본문은 `filepath.Walk` + `ReadFile` + `WriteFile`의 순수 복사이며, 삭제는 오류 롤백 `os.RemoveAll(backupDir)` 3곳뿐이다. 재정박 대상인 `CleanupOldBackups`는 `backups[:len(backups)-keepCount]`를 쓰고, **과거 버그가 주석에 그대로 남아 있다**: `(A prior revision deleted backups[keepCount:] — the newest — which destroyed the most recent restore points on every rotation.)` 역사적 회귀를 재현하는 카나리아로 타당하다.
3. **AC-UGE-015가 주석을 위반으로 계수 — 확인.** 최초 명령 실행 결과 `update_home_radius_test.go 1`, 그 `1`은 **17행 주석**(`// ... this test MUST NOT call t.Parallel().`). 개정 명령(주석 제외) 실행 결과 두 파일 모두 `0`. iter1에서 나는 이 AC를 `xargs` 이식성만 지적하고 통과시켰다.

#### 저자의 REQ-UGE-004 보강 — 정확함(내 결론을 강화)

`internal/runtime/gobin/` `Detect(homeDir)`를 직접 읽어 확인: ① `GOBIN` ② `GOPATH/bin` ③ **`filepath.Join(homeDir, "go", "bin")`** ④ `""`. 즉 `GOBIN`/`GOPATH`가 비면 운영자 실제 홈이 렌더 산출물에 박힌다. 삭제 반경은 없다는 내 F5 결론은 유지되고, "테스트 격리 누수"라는 재서술이 정확하다.

#### 내 iter1 지적 10건의 처분

| # | 처분 | 근거 |
|---|---|---|
| F1 AC-UGE-012 무동작 sed | **해소(실행 확인)** | 재조준 후 `mutation-applied=1`, 변경행 `deploy.go:53-54`, 가드 FAIL 계수 **4** — 내가 직접 재현 |
| F2 REQ-UGE-013 ↔ §D 모순 | **해소** | 6-가드 반증 표 + §D 완전성 점검("반증 없는 신규 가드 0개") |
| F3 AC-UGE-009/010 (d) 토큰 계수 | **해소** | 009(d)는 구조 순서 검사로 교체, 010은 두 파괴 표면으로 재정박, 오선택자 (c) 삭제 |
| F4 AC-UGE-003 base 교란 | **해소(부분 일반화)** | (0) 게이트 추가, 실측 `3`. 단 신규 반증 3건에는 같은 규율 미적용 → **N1** |
| F5 M2→M3 의존 거짓 | **해소** | spec.md·plan.md §A·§F M2 모두 "거짓이었다" 명시 + 독립성 재서술 |
| F6 AC-UGE-006 공허 | **부분 해소** | 006/006F/006R 3분할은 실질 개선. 스파이가 REQ가 요구하는 것을 재지 않음 → **N2** |
| F7 xargs GNU | **해소(내 근거는 부정확)** | `xargs` 제거로 차이 자체를 무의미화. 다만 내 "hang" 메커니즘 주장은 오류 → **N3** |
| F8 NFR-UGE-002 오표기 | **해소** | (e)가 "NFR-UGE-001 계열"로 정정 + 오표기 이력 명시 |
| F9 REQ-UGE-007 근거 전치 | **해소** | "sentinel이 비어 있지 않아야 하는 진짜 이유" 문단이 카나리아 역할로 재서술 |
| F10 §1.1 저작 시점 거짓 | **해소** | 머지 증거 + 정정 이력 기재 |

#### 신규 지적 (iter2)

- **N1 (차단, major) — 신규 반증 3건이 실패를 *귀속*하지 않는다.** AC-UGE-009F/010F/011F는 모두 `grep -cE '^(--- FAIL|FAIL)' ≥1`만 요구한다. 009F를 코드로 추적하면, 주입된 `.moai/harness`는 `os.RemoveAll` 루프 **이전에** `expandDeprecatedBackupTargets`/`backupDeprecatedPaths`를 먼저 지나며, 거기서 `DEPRECATED_BACKUP_FAILED`로 조기 반환해도 똑같이 `--- FAIL`이 찍힌다. 두 원인이 구별되지 않는다 — F4에서 지적한 교란과 같은 계열이다. 저자는 이 규율을 AC-UGE-003에만 적용했다. **수정**: FAIL 출력이 사용자 소유 경로명 또는 스냅샷 불일치를 담는지 함께 어서트.
- **N2 (차단, major) — AC-UGE-006의 스파이가 REQ-UGE-005를 재지 않는다.** REQ-UGE-005는 "*렌더링 결과*가 주입된 홈을 따라야 한다"인데, 호출 계수 스파이는 "이음매가 불렸다"까지만 증명한다(이 SPEC 자신의 §G AP-4). 저자가 출력 어서션을 피한 근거(`gobin.Detect`의 환경 의존)는 **`GoBinPath`에만** 해당한다 — `TemplateContext.HomeDir`은 평범한 대입 필드(`internal/template/context.go:239`)이고 기존 테스트가 이미 직접 어서트한다(`context_test.go:324` `ctx.HomeDir != "/home/tester"`). 기계 독립적이면서 더 강한 판정이 가능한데 과잉 일반화로 포기했다. **수정**: 구성된 컨텍스트의 `HomeDir`이 주입한 sentinel과 같은지 추가 어서트(스파이는 호출부 커버리지용으로 유지).
- **N3 (선택, minor) — 내 F7의 메커니즘이 부정확한 채 산출물에 인용됐다.** BSD/GNU `xargs` 차이(빈 입력 시 BSD 미실행 / GNU 1회 실행, `-r`이 존재하는 이유)는 사실이다. 그러나 GNU `xargs`는 자식 stdin을 `/dev/null`로 리다이렉트하므로 `grep`은 즉시 EOF를 만나 **`0`을 출력하고 종료**한다 — "행(hang)"이 아니다. 저자는 plan.md B9와 AC-UGE-015에 "GNU 환경 재현 미실시"를 명시했으므로 미검증 라벨은 붙어 있다. **수정**: 결과 서술을 "인자 없이 1회 실행되어 허위 결과행을 낸다"로 정정. 채택된 수정(제거)은 어느 쪽이든 안전.
- **N4 (선택, minor) — `while read` 교체가 0매치 공허를 그대로 남긴다.** 무매치 패턴으로 실측: 루프 본문이 한 번도 실행되지 않고 출력이 비며 실패 신호가 없다. §A.1이 `-run`에 대해 금지하는 형태와 동일하고, §A.2는 픽스처의 비어있지 않음 확인을 요구한다. **수정**: 루프 앞에 파일 목록 비어있지 않음 확인(AC-UGE-007 (a) `before-nonempty`의 동형).
- **N5 (선택, minor) — 주석 필터가 행 앵커 전용.** `grep -vE '^\s*//'`는 전체 주석행만 제거하므로 후행 주석(`foo() // t.Parallel()`)과 블록 주석은 여전히 계수된다. 현 트리에서는 무해.
- **N6 (선택, minor) — DoD 열거가 19개, 실제 AC는 20개.** `grep -cE '^#### AC-UGE-' acceptance.md` → **20**. §E DoD 1항의 열거에서 **AC-UGE-004**(mergeBackPreserveInventory 커버리지)가 빠져 있다.
- **N7 (선택, minor) — 개정이 HISTORY·version에 기록되지 않았다.** 3파일 452행 삽입 규모의 개정인데 `version: 0.1.0` 유지, HISTORY 표에 행 추가 없음.

#### `origin/main` 머지(`4ce694296`)의 타당성

**적절하다.** ① 머지 없이는 타깃 파일 3개가 트리에 없어 판정 명령을 실행할 수 없었다(iter1 시점 실측: `update_preserve_partial_test.go` 부재, `restored %d/%d` 0건). ② 이 브랜치는 워크트리이므로 `main-checkout-branch-guard.md`의 주 체크아웃 병합 금지 대상이 아니다. ③ rebase가 아닌 merge를 택한 것이 옳다 — rebase는 기록된 SHA를 고아로 만든다. ④ `merge-base`가 `a64548a2a` → `83610e03e`로 올라가 F4 교란이 해소되었고, 저자는 (0) 게이트를 **남겨** base가 다시 움직이는 경우를 대비했다.

**iter1 측정의 무효화 여부: 없음.** iter1의 F4는 당시 base(`a64548a2a`)에서 `restored %d/%d` = 0이라는 실측에 근거했고 그 관측은 지금도 참이다(재실행 확인). 머지는 그 교란을 제거했을 뿐 측정을 뒤집지 않았다. `origin/main`은 그 후 2커밋 앞서 있으나(`2 3`) merge-base는 `83610e03e`로 고정이라 AC-UGE-013의 타인 변경 배제도 유지된다(실측 `0` / `0`).

#### 이번 패스에서 실행해 관측한 것

| 판정 | 명령 | 관측 |
|---|---|---|
| AC-UGE-003 (0) | `git show $BASE:…update_preserve_inventory.go \| grep -c 'restored %d/%d'` | `3` |
| AC-UGE-005 (c) | 동 baseline | `1` / `2` |
| AC-UGE-001 (c) | 함수창 `os.Stat(` 계수 | `1` |
| AC-UGE-007 (a)~(e) | sentinel HOME 전 과정 | `before-nonempty=0`, `compile-exit=0`, `--- PASS: TestEnsureGlobalSettingsEnv_HooksRemovalRadius (0.00s)`, `diff-exit=0`, `t.Setenv=0`, 이음매 `1` |
| AC-UGE-008 | 이음매 제거 변형 | `mutation-applied=1`, `compile-exit=0`, **`probe-survived=1`** (카나리아 작동) |
| AC-UGE-012 | 재조준 글로브 변형 | `mutation-applied=1`, `deploy.go:53-54`, FAIL 계수 **4** |
| AC-UGE-009F | perl 변형 | `mutation-applied=1`, `144a145` 단일 삽입, 파일 내 동일 패턴 **1곳뿐** |
| AC-UGE-010F | 회전 슬라이스 반전 | `mutation-applied=1`, `backup.go:257` |
| AC-UGE-011F | **변형 컴파일** | `mutation-applied=1`, **`go build -overlay` exit 0** — 저자가 "확인 불가"로 남긴 잔여 한계를 해소 |
| AC-UGE-013 | 템플릿 중립성 | `0` / `0` |
| AC-UGE-014(lint) | `golangci-lint run --timeout=5m` | exit 0, `0 issues.` |
| AC-UGE-015 | 원본/개정 명령 대조 | 원본 `1`(주석 오탐) → 개정 `0` |
| depends_on | `grep '^status:'` | `completed` (Pre-flight 통과) |

#### 저자가 선언한 3대 잔여 한계에 대한 판단

1. **9개 AC 미실행(대상 코드 부재)** — **구조적 plan-phase 한계**, 결함 아님. run-phase가 만들 테스트를 plan-phase가 실행할 수는 없다. 감점 사유로 보지 않는다.
2. **AC-UGE-002 Windows 런타임은 CI 인용만 가능** — **정직한 범위 한정**. (b)+(d)가 "소스에 스킵 없음 + windows 타깃 컴파일"까지로 증명 범위를 명시하고 DoD 4항이 이월을 강제한다. 더 큰 공백을 숨기지 않는다.
3. **AC-UGE-011F 변형 컴파일 미확인** — **해소됨**. 위 표대로 내가 컴파일을 확인했다(exit 0). AC의 대체 변형 안내 문단은 관측 결과를 기록하는 방향으로 정리하면 된다.

#### 미검증 (이번 패스)

- 신규 가드 6종(`TestUpdateSubsystem_HomeSeamReach`, `TestCleanReinstall_PreservesUserArea`, `TestBackupSubsystem_DestructiveSurfaces`, `TestMigrateLegacyMemoryDir_PreservesUserArea` 및 그 서브테스트)의 **실제 PASS/FAIL** — 코드가 아직 없다. AC-UGE-006/006F/009/009F/010/010F/011/011F의 `--- PASS` / FAIL 관측은 전부 run-phase 몫.
- **GNU `xargs` 실동작** — 이 머신은 BSD. N3은 문서 근거에 기반한 추론이며 재현하지 못했다.
- **AC-UGE-002 Windows 런타임** — macOS에서 관측 불가(위 2번과 동일).
- **`go test ./...` 전체** 및 `-race` — 이번 패스에서 실행하지 않았다(AC-UGE-014의 FAIL 계수 항목 미검증).
- `TemplateContext.HomeDir` 어서션이 실제 렌더 파이프라인 끝까지 기계 독립인지 — 필드 대입과 기존 테스트의 직접 어서트까지만 확인했고, 렌더 산출물 전체를 추적하지는 않았다(N2 처방의 근거 범위).

## §E.2 Run-phase Evidence

### §C 선행 조건 재확인 (run-phase 진입 시점)

plan.md §C1의 절차를 진입 시점에 **다시 실행**했다(문서의 문장이 관측을 대체하지 않는다).

```
$ git fetch origin main && git merge-base origin/main HEAD
83610e03eae2f1c9cfb2dc835f995c12de09dfcf
$ git show "$BASE":internal/cli/update_preserve_inventory.go | grep -c 'restored %d/%d'
3
$ test -f … && echo ok  (3종)
safety-ok / partial-ok / radius-ok
$ git branch --show-current ; git rev-parse --short HEAD
feat/SPEC-UPDATE-GUARD-EFFICACY-001 ; 0b31ba8b7
```

### §C2 baseline 재측정 (plan-phase 수치를 옮겨 적지 않고 이 트리에서 실행)

| 항목 | 재측정값 | §B 기록 | 판정 |
|---|---|---|---|
| `userHomeDir()` 비이음매 호출부 | `update_clean_install.go:1` / `update_template_sync.go:2` | 동일 | 일치 |
| stat 이음매 존재 | 매치 0건 | 동일 | 일치 |
| `snapshotDir` 사용처 | `update_safety_test.go` 1개 | 동일 | 일치 |
| `deploy.go` 글로브 | `filepath.Join(defs.ClaudeDir, defs.SkillsSubdir, "moai*")` | 동일 | 일치 |
| 백업 회전 슬라이스 | `backups[:len(backups)-keepCount]` (`backup.go:257`) | 동일 | 일치 |
| `glm.go` 범위 밖 호출부 | `1` | 동일 | 일치 |
| 가드 스코프 커버리지 | `CleanMoaiManagedPaths 66.7%` / `MigrateLegacyMemoryDir 26.9%` / `runCleanReinstall 0.0%` / `BackupMoaiConfig 0.0%` | 동일 | 일치 |
| `mergeBackPreserveInventory` 커버리지 (merge-base 트리에서 실측) | `94.1%` | `94.1%` | 일치 |

**측정 중 발견한 함정 (기록)**: 최초 재측정 시도에서 `84.2%`가 관측되었다. 그 값은 **편집 도중의 트리**를 잰 것이다 — 이음매 선언은 들어갔으나 프로덕션 라우팅은 아직 안 된 RED 상태여서 stat 분기가 도달 불가였다. 배경 작업이 편집과 경쟁한 결과이며 baseline 이동이 아니다. merge-base 트리를 별도 워크트리로 꺼내 재측정해 `94.1%`를 확인했고, §B 수치가 옳다.

### M1 — stat 이음매 (REQ-UGE-001, 002, 003)

**Claim**: `mergeBackPreserveInventory`의 stat 실패 분기가 주입 가능한 이음매로 구동되며, 어떤 플랫폼에서도 `t.Skip`하지 않는다.

**RED 관측 (test-first 증거)** — 이음매를 *선언만* 하고 프로덕션 라우팅 전에 테스트를 실행했다. 컴파일 오류가 아니라 **실제 동작 공백**으로 실패한다.

```
=== RUN   TestMergeBackPreserveInventory_PartialRestore/stat_failure_injected_stat_error
    update_preserve_partial_test.go:97: mergeBackPreserveInventory: expected error, got nil
--- FAIL: TestMergeBackPreserveInventory_PartialRestore (0.03s)
    --- FAIL: TestMergeBackPreserveInventory_PartialRestore/stat_failure_injected_stat_error (0.01s)
    --- PASS: TestMergeBackPreserveInventory_PartialRestore/mkdirall_failure_parent_is_regular_file (0.01s)
    --- PASS: TestMergeBackPreserveInventory_PartialRestore/reports_restored_count (0.02s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.267s
```

**GREEN 관측** — `os.Stat(srcPath)` → `osStatFn(srcPath)` 라우팅 후.

```
--- PASS: TestMergeBackPreserveInventory_PartialRestore (0.01s)
    --- PASS: TestMergeBackPreserveInventory_PartialRestore/stat_failure_injected_stat_error (0.00s)
    --- PASS: TestMergeBackPreserveInventory_PartialRestore/mkdirall_failure_parent_is_regular_file (0.00s)
    --- PASS: TestMergeBackPreserveInventory_PartialRestore/reports_restored_count (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.753s
```

**AC 매트릭스**

| AC | 판정 | 명령 | 관측 출력 |
|---|---|---|---|
| AC-UGE-001 | PASS | (a) `grep -nE '^var [a-zA-Z]*[Ss]tatFn = os\.Stat$'` · (b) 함수 본문 창 `os.Stat(` / `statFn(` 계수 · (c) base 직접 호출 계수 | (a) `54:var osStatFn = os.Stat` · (b) `0` / `1` · (c) `1` |
| AC-UGE-002 | PASS | (a) 존재 grep · (b) 스킵 가드 계수 · (c) PASS 행 · (d) windows vet | (a) `1` · (b) `0` · (c) `--- PASS: …/stat_failure_injected_stat_error (0.00s)` · (d) `win-vet-exit=0` |
| AC-UGE-003 | PASS | overlay 반증 (base 위생 + FAIL 계수 + 원인 귀속) | `(0)=3` · `fail-lines=2` · 원인: `internal/cli/update_preserve_partial_test.go:84:15: undefined: osStatFn` (3행) + `FAIL … [build failed]` |
| AC-UGE-004 | PASS | `go tool cover -func` | `mergeBackPreserveInventory 94.1%` (baseline `94.1%` 이상) |

**AC-UGE-003 원인 귀속 (§A.4)**: 실패 출력이 `undefined: osStatFn`를 3행 지목한다 — 변경 전 코드에 이음매가 없다는 **의도한 원인**을 직접 명명한다. 경쟁 원인(보고 문구 부재)은 (0) 위생 검사 `3`이 배제한다.

**회귀 게이트 (M1)**

```
$ go build ./...                          → build-exit=0
$ GOOS=windows GOARCH=amd64 go build ./... → win-build-exit=0
$ go vet ./...                            → vet-exit=0
$ go test -count=1 ./internal/cli/...     → exit 1, FAIL 계수 4
```

**`./internal/cli/...` FAIL 4건의 귀속 — 이 SPEC의 회귀가 아니다.** 실패는 `TestHookWrapper_LargeStdin_DoesNotExceedTimeout` 한 건이며, 벽시계 임계값 어서션이다.

```
hook_wrapper_load_test.go:55: wrapper execution took 1.086526291s, want < 1s (regression: timeout risk)
```

단독 재실행 3회 전부 PASS(`77.1ms` / `76.3ms` / `77.4ms` — 임계값의 약 1/13):

```
--- PASS: TestHookWrapper_LargeStdin_DoesNotExceedTimeout (0.17s)  ×3
ok  	github.com/modu-ai/moai-adk/internal/cli	1.271s
```

부하 민감형 flaky이며(측정 시점에 백그라운드 커버리지 작업이 동시 실행 중이었다), 이 SPEC의 변경 범위(`git status --porcelain` → `update_preserve_inventory.go` + `update_preserve_partial_test.go` 2개)와 접점이 없다.

**AC-UGE-013 / 015 (M1 시점)**: 템플릿 diff `0` / porcelain `0`; `target-files=3`, 세 파일 모두 `t.Parallel()` 계수 `0`.

**Gaps (M1)**
- AC-UGE-002의 **Windows 런타임** 통과는 이 머신에서 관측 불가. (b)+(d)는 "소스에 스킵 조건이 없고 windows 타깃으로 vet 통과"까지만 증명한다.

### M2 — HOME 이음매 균일화 (REQ-UGE-004, 005)

**Claim**: update 서브시스템의 세 HOME 호출부가 `userHomeDirFn` 이음매를 경유하며, 주입된 홈이 렌더링 입력까지 도달한다.

**RED 관측 (test-first 증거)** — 호출부를 바꾸기 전에 가드를 실행했다. 이 RED는 **부채 자체를 재현**한다.

```
=== RUN   TestUpdateSubsystem_HomeSeamReach/clean_reinstall_deploy_context
    update_home_seam_test.go:79: userHomeDirFn calls = 0; want >= 1 (the deploy-context site does not route through the seam)
    update_home_seam_test.go:90: TemplateContext.HomeDir = "/Users/goos"; want the injected sentinel "/var/folders/…/001" — rendering followed the process $HOME instead of the seam
=== RUN   TestUpdateSubsystem_HomeSeamReach/template_sync_validate_and_deploy
    update_home_seam_test.go:132: userHomeDirFn calls = 1; want >= 2 (Validate Templates + Deploy Templates must both route through the seam)
--- FAIL: TestUpdateSubsystem_HomeSeamReach (3.43s)
    --- FAIL: TestUpdateSubsystem_HomeSeamReach/clean_reinstall_deploy_context (0.65s)
    --- FAIL: TestUpdateSubsystem_HomeSeamReach/template_sync_validate_and_deploy (2.78s)
```

`TemplateContext.HomeDir = "/Users/goos"` — 픽스처를 주입했는데도 **운영자의 실제 홈이 렌더링 입력에 박혔다**. REQ-UGE-004가 지목한 테스트 격리 누수의 직접 관측이다.

**GREEN 관측** — 세 호출부를 `userHomeDirFn()`으로 교체 후.

```
--- PASS: TestUpdateSubsystem_HomeSeamReach (0.33s)
    --- PASS: TestUpdateSubsystem_HomeSeamReach/clean_reinstall_deploy_context (0.06s)
    --- PASS: TestUpdateSubsystem_HomeSeamReach/template_sync_validate_and_deploy (0.27s)
```

**AC 매트릭스**

| AC | 판정 | 명령 | 관측 출력 |
|---|---|---|---|
| AC-UGE-005 | PASS | (a) 비이음매 계수 · (b) 이음매 계수 · (c) baseline · (d) 범위 밖 보존 | (a) `clean_install:0` / `template_sync:0` · (b) `clean_install:1` / `template_sync:2` · (c) `1` / `2` · (d) `glm.go:1` |
| AC-UGE-006 | PASS | (a) 존재 grep · (b) PASS 행 · (c) 계수 어서션 · (d) `.HomeDir` 어서션 | (a) `update_home_seam_test.go:57` · (b) `--- PASS: TestUpdateSubsystem_HomeSeamReach (1.44s)` · (c) `2` · (d) `5` |
| AC-UGE-006F | PASS | overlay 변형 반증 | `mutation-applied=1` (두 파일) · `compile-exit=0` · `fail-lines=4` · `names-reach-assert=3` · `build-failed=0` |
| AC-UGE-006R | PASS | `go test -count=1 ./internal/cli/...` · `go vet ./...` | `cli-test-exit=0`, FAIL 계수 `0` · `vet-exit=0` |

**AC-UGE-006F 원인 귀속 + 경쟁 원인 배제 (§A.4)**: 실패 3행이 모두 도달 어서션(`calls` / `HomeDir`)을 지목한다 — 되돌린 코드에서 계수가 `0`/`1`로 떨어지고 `HomeDir`가 다시 `/Users/goos`가 된다. `compile-exit=0` + `build-failed=0`이 "sed가 코드를 깨뜨려서 FAIL"이라는 경쟁 원인을 배제한다.

#### 실행 중 발견 — 렌더된 `settings.json`의 머신 종속성은 이 SPEC의 이음매로 해소되지 않는다

가드 초안은 "렌더된 `settings.json`에 운영자의 실제 `go/bin` 경로가 없어야 한다"를 함께 어서트했다. **세 호출부를 모두 이음매로 바꾼 뒤에도 이 어서션은 실패했다.**

```
update_home_seam_test.go:145: rendered settings.json embeds the operator's real go/bin path (/Users/goos/go/bin); the seam did not reach the rendering input
```

원인을 실측했다 — 렌더된 PATH는 세 호출부에서 오지 않는다.

```
$ grep -n -A3 'func BuildSmartPATH' internal/template/settings.go
44:func BuildSmartPATH() string {
45:	homeDir, _ := os.UserHomeDir()
…
55:		filepath.Join(homeDir, "go", "bin"),     // Go workspace binaries
```

`template.BuildSmartPATH()`는 `os.UserHomeDir()`을 **직접** 부르는 네 번째 독립 리더이며, `userHomeDirFn` 이음매 밖에 있고 REQ-UGE-004가 지목한 세 호출부에도 속하지 않는다. 따라서 그 어서션은 **이음매가 통제할 수 없는 이유로 실패**하는, REQ가 주장하지 않는 것을 검사하는 과잉 어서션이었다. 가드에서 제거하고 그 자리에 근거를 주석으로 남겼다(테스트 파일 내). 이음매의 도달 증명은 계수 어서션과 site 1의 `TemplateContext.HomeDir` 어서션이 담당한다.

**이월 (§3.5)**: 렌더 산출물 `settings.json`의 PATH는 M2 이후에도 여전히 실행 머신의 홈에 의존한다. 이 SPEC의 범위(세 호출부)를 벗어나므로 고치지 않았고, 별도 SPEC 후보로 남긴다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
