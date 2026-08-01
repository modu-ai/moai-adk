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

**선행 조건 (갱신)**: PR #1275는 머지 완료(`83610e03e`)이고 이 브랜치에 머지 반영했다. run-phase는 plan.md §C1의 base 위생 검사를 **진입 시점에 재실행**한다.

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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
