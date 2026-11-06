verdict: FAIL
score: 0.88
audited_sha: 7eb36bd403691ca8631d875d1fc484c197ff3f66
auditor-model: claude-opus-5-5

# t2135 sync-audit (2회차) — PR #1969 Windows "기타" 실패 분류 보고서 재감사

code verdict: FAIL

- 카드: t2135 (보고서 전용, 제품 변경 없음)
- 감사 대상: `.moai/reports/t2135/progress.md` @ 7eb36bd403 (1회차 대비 변경분 `git diff 665a7c5673..HEAD`)
- 원본 표: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/43472e68-3604-44b1-a008-c14cce281ea1/scratchpad/categorized.tsv` (1335행)

## 반복 이력

| 회차 | 감사 커밋 | 판정 | 점수 | 차단 결함 | 비고 |
|---|---|---|---|---|---|
| 1 | 665a7c5673 | FAIL | 0.77 | F1 G6 귀속, F2 G3 오독, F3 건수 3곳, F4 그룹 합·NOMSG | F5~F8 선택 |
| 2 | 7eb36bd403 | FAIL | 0.88 | R2-1 (NOMSG "나머지 62행은 모두 부모" 주장 오류, 수리에서 생김) | F1~F3 해소, F4 대부분 해소. R2-2~R2-4 선택 |

수리 중단 규칙: R2-1은 직전 감사 이후 변경분 안에 있는 차단 결함이므로 repair-born 1회째입니다. 다음 감사에서도 repair-born 차단 결함이 나오면 패치 대신 범위 축소·재설계 중 하나를 고르고 기록해야 합니다.

## 판정 요약

1회차 차단 결함 F1·F2·F3은 모두 해소됐습니다. 보고서가 내세운 서로소 분할 규칙을 그대로 다시 돌려 보니 G2 15 · G3 8 · G4 12 · G5 10 · G6 5가 정확히 재현되고, 합계도 201입니다. G6·G3 재귀속은 이 트리 코드와 일치하고, 인용한 카드 번호 9개도 모두 존재하며 범위가 맞습니다.

남은 차단 결함은 하나입니다. 보고서는 G1 패키지 밖 NOMSG 62행을 두고 "다른 그룹 실패의 부모 행이라 독립 실패가 아니다"라고 단정합니다. 표를 다시 세어 보면 그중 메시지 없는 독립 실패가 6건 있습니다(internal/cli의 hook sink 계열 5건, homestate 1건). 이 6건은 어떤 그룹에도, 어떤 카드에도 들어가지 않은 채 묻힙니다. 1회차 F4의 필수 수정("나머지 62행이 어느 그룹·카드의 부모인지 적기")도 이 부분은 이행되지 않았습니다. 문장 하나와 행 하나만 고치면 되는 작은 수정입니다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 80/100 | FAIL | 서로소 분할 G2~G6·합계 201 재현, 범주·패키지·단일 패키지 14 일치, F1/F2 귀속 코드와 일치, 카드 번호 일치. NOMSG 62행 주장에 반례 6건(R2-1, 차단) |
| Security (25%) | 100/100 | PASS | 카드 변경 파일은 보고서 3개뿐(`git diff --name-only main...HEAD`). 변경분 비밀값 패턴 탐색 결과 `0` |
| Craft (20%) | 85/100 | PASS | 5절 형식, "추론"·"미재현" 표기, 표 출처 주의 문구 갖춤. G7 규칙 문구가 그대로는 재현되지 않음(R2-2) |
| Consistency (15%) | 90/100 | PASS | 카드 기록 형식·경로 규약 준수, 1회차 결함 번호와 수리 내역 대응 기록 |

가중 합계: 0.4×80 + 0.25×100 + 0.2×85 + 0.15×90 = 87.5 → 0.88. 판정을 가른 것은 점수가 아니라 차단 결함 R2-1입니다.

## 1회차 결함 처리 확인

| 결함 | 상태 | 확인 근거 |
|---|---|---|
| F1 G6 귀속 | 해소 | `internal/session/ccversion.go:167` `versionSegmentRe` = ``(?:^|/)(?:versions|claude-code)/…``, `:173` `trailingBinaryVersionRe` = ``/(…)/claude$``, 접두 `"/versions/"`·접미 `"/"+claudeBinaryName` 슬래시 고정. 테스트 `ccversion_test.go` `installedVersionViaSymlink`는 `t.TempDir()` + `filepath.Join` + `os.Symlink`, PATH만 seam. `doctor_mcp_provider_test.go:54, 68`은 `filepath.Abs` 결과를 이스케이프 없이 JSON에 넣고, `doctor_mcp_provider.go:103-105`는 `json.Unmarshal` 실패 시 "Claude Code state file unreadable; not checked". 보고서의 새 귀속과 일치 |
| F2 G3 오독 | 해소 | `empty_target_guidance_test.go` `assertAllContained`가 `"%s must contain %q, got:\n%s"`. `integration_target_guidance.go:85-87`만 `filepath.Join(root, filepath.FromSlash(gitStrategyRelPath))`. 실패한 8행은 모두 `no_…`/`unparseable_…` 하위 테스트(= unreadable 경우). 보고서는 가설로 표기하고 재현을 첫 단계로 둠 |
| F3 건수 | 해소 | `backlog.lock` 6, `state file unreadable` 3, 단일 패키지 14 모두 재현 |
| F4 그룹 합·NOMSG | 대부분 해소 | 합계 201 재현, "318은 독립 실패 수가 아님" 명시. 다만 62행 처리에서 새 오류(R2-1) |
| F5 G8 | 반영 | G8 삭제 |
| F6 | 기록만 | "G9 안에 있고 따로 분류하지 않음"으로 명시 |
| F7 | 반영 | G4 항목에 경로 표기 가능성 기록 |
| F8 | 반영 | head SHA 출처 주의 문구 추가(SHA 값 자체는 미기재) |

## 발견 사항 (결함 목록)

- **R2-1** [Low] [blocking] [in-diff] [repair-born] progress.md:23 — "The other 62 NOMSG rows are parents of failures in other groups and are not independent failures."는 표와 어긋납니다.
  - 62행 안에 하위 테스트 행이 전혀 없는(=부모가 아닌) 메시지 없는 실패가 있습니다: `internal/cli` `TestHookSinkConcurrentAppendIsIntact`, `TestHookSinkIsNotCreatedWithoutRecords`, `TestHookSinkLevelBoundary/default_config_admits_warn_only`, `TestHookSinkLevelBoundary/lowered_level_admits_info`, `TestHookPathWarnRecordReachesSink`, 그리고 `internal/homestate` `TestOpenFactoryReadonlyNeverCheckpointsWAL`. 독립 실패 6건이며, `TestHookSinkLevelBoundary` 1행은 이들 NOMSG 하위 테스트의 부모입니다.
  - 62행에는 G1 범주(SQLITE 3, DIRSYNC 1)의 부모 4행도 섞여 있어 "다른 그룹"이라는 표현도 정확하지 않습니다.
  - 결과적으로 hook sink 5건짜리 묶음이 어느 카드에도 들어가지 않습니다.
  - **필수 수정**: 해당 문장을 "62행 중 약 55행은 OTHER 33·CLI_WORKTREE 7·HOMEGUARD 5·POSIX 4 등 다른 범주 실패의 부모, 4행은 G1 범주의 부모, 6건은 메시지 없는 독립 실패(hook sink 5 + homestate WAL 1)"로 고치고, 이 6건을 G9 풀이나 별도 행으로 제안 카드 목록에 넣습니다. 숫자는 아래 E6 출력에서 옮기면 됩니다.
- **R2-2** [Low] [optional] [in-diff] progress.md:19 — G7 규칙("`C:` 경로, `RUNNER~1`, spelling/case-fold/raw-path 문구")을 글자 그대로 옮기면 G7 54 · G9 97이 나옵니다. "메시지에 백슬래시가 있음" 조항을 더해야 보고서 숫자 61 · 90이 정확히 나옵니다. 이 조항은 경로 표기가 아닌 4행을 함께 끌어옵니다: `TestAstGrepCmd_UsageErrorIsNotRulesetHeadline`(`stdout="no findings\n"` 이스케이프), `TestPlanAuditTraceability_ZeroDefinitionsIsGap`·`_InlineACsWithoutAcceptance`(awk 경고 "escape sequence `\*'"), `TestMergeStepRealContentConflictAbortsAndReleases`(`"release work\r\n"`, CRLF 줄끝). 이 차이는 G7과 G9를 함께 다루는 탐색 카드의 설계를 바꾸지 않으므로 차단 사유는 아닙니다. 수정 제안: 규칙 칸에 백슬래시 조항을 적고, CRLF 1행과 awk 방언 2행은 경로 표기가 아니라고 표시합니다.
- **R2-3** [Low] [optional] [in-diff] progress.md:16 — `TestDoctorTodoStoreDivergence_ThreeStates` 2행(메시지 `does not carry ".moai/state/todo/backlog.db"`)은 todo 저장소 행인데 규칙상 G9로 떨어집니다. 1회차 F3에서 이 2행을 G4 구성에 넣으라고 했으나 보고서는 테스트 이름 규칙을 택했습니다. 서로소 분할로서는 틀리지 않으며, G4 카드 범위에 "같은 저장소의 doctor 행 2건 포함"을 적어 두면 충분합니다.
- **R2-4** [Low] [optional] [in-diff] progress.md:43 — head SHA 출처 주의 문구는 들어갔지만 SHA 값 자체는 여전히 없습니다. 표가 어느 커밋의 CI 결과인지 보고서만으로는 알 수 없습니다. `gh pr view 1969 --json headRefOid` 결과(또는 해당 CI 실행의 head)를 한 줄 기록하면 됩니다.

## 근거 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim
1. 보고서의 서로소 분할은 G2~G6에서 정확히 재현되고 합계 201이 맞다. G7/G9 경계는 규칙 문구 해석에 따라 7행 차이가 난다.
2. 범주 집계, OTHER 패키지 분해, 단일 패키지 14, `backlog.lock` 6, `state file unreadable` 3은 정확하다.
3. F1·F2의 새 귀속은 이 트리의 코드와 일치한다.
4. 제안 카드는 각자 첫 단계(재현 또는 가설 확인)와 크기(행 수)를 갖췄고, 다른 카드로 넘긴 행의 카드 번호 9개는 모두 큐에 있고 범위가 맞다.
5. NOMSG 62행에 대한 주장은 6건에서 틀렸다(R2-1).

### Evidence (이번 실행에서 직접 돌린 명령과 출력, `LC_ALL=C`)
- E1 `git rev-parse HEAD` → `7eb36bd403691ca8631d875d1fc484c197ff3f66`. `git diff --stat 665a7c5673..HEAD` → `progress.md | 73 +++---`, `sync-audit.md | 128 +++`, `2 files changed, 173 insertions(+), 28 deletions(-)`. `git diff --name-only main...HEAD` → `card-review.md`, `progress.md`, `sync-audit.md` (모두 `.moai/reports/t2135/`).
- E2 `wc -l` → `1335`. `cut -f2 | sort | uniq -c` → `474 SQLITE_URI / 326 DIRSYNC_ACCESS_DENIED / 201 OTHER / 117 NOMSG(parent/other) / 60 CLI_WORKTREE / 54 HOMEGUARD_SEAM_REALHOME / 46 POSIX_MODE_PERM / 26 OWNERSHIP_EPOCH / 19 BIN_SH_MISSING / 12 PRECOMMIT_E2E`. 모든 행 열 수 `1335 5`.
- E3 OTHER 패키지(`awk '$2=="OTHER"{print $3}' | sort | uniq -c`) → `134 internal/cli, 12 internal/factory, 7 internal/cli/update/skillresidue, 5 internal/session, 5 internal/homestate, 5 internal/graph, 4 internal/template, 4 internal/statusline, 4 internal/agenticfactory/claim, 3 internal/auditreceipt, 2 internal/astgrep, 2 internal/agenticfactory/mergeq` + 1건짜리. `awk '$1==1' | wc -l` → `14`. `grep -c 'backlog\.lock'` → `6` (6행 모두 `TestTodoPR_QueueDirUnchanged/*`). `grep -c 'state file unreadable'` → `3`.
- E4 서로소 분할 (스크래치 `g.awk`, 첫 일치 규칙: G2 이름 `^(TestStopGoalWait|TestStopParseCap|TestFactoryWaitResolveCloses)` → G3 `^TestEmptyTarget` → G4 `^(TestTodoStaleStoreDisclosure|TestTodoPR_QueueDirUnchanged)` → G5 패키지 `skillresidue` 또는 메시지 `skill\.md|SKILL\.md|skill folder|skills mirror|mirror absent` → G6 메시지 `installed version|state file unreadable` 또는 이름 `^TestCheckMCPProviderDuplicat` → G7 메시지 `C:|RUNNER~1|spelling|case-fold|casefold|fold|raw path|raw-path` → 나머지 G9) → `15 G2 / 8 G3 / 12 G4 / 10 G5 / 5 G6 / 54 G7 / 97 G9`, 합계 `201`.
- E5 G7 해석 변형(G9 97행 대상): 메시지에 백슬래시 → `7`행 (`TestAstGrepCmd_UsageErrorIsNotRulesetHeadline`, `TestMigrationArchiveRefusesSymlinkedArchiveDestination/flow_aborts_before_any_removal`, `TestProductionStringLiteralsUseLeaderLaneVocabulary`, `TestPlanAuditTraceability_ZeroDefinitionsIsGap`, `TestPlanAuditTraceability_InlineACsWithoutAcceptance`, `TestMergeStepRealContentConflictAbortsAndReleases`, `TestPlan_IsReadOnly_MaintenanceLockRepro`). 54+7 = 61, 97−7 = 90 → 보고서 숫자와 정확히 일치. 이 중 4행의 메시지: `stdout="no findings\\n"`, `awk: cmd. line:24: warning: escape sequence `\\*' treated as plain` (2행), `card.txt = "release work\\r\\n"`.
- E6 NOMSG 부모 판정(`python3 -I parents.py`, 같은 패키지에 `<이름>/` 하위 행이 있으면 부모, 하위 행 다수 범주로 분류) → `40 parent-of SQLITE_URI / 33 parent-of OTHER / 18 parent-of DIRSYNC_ACCESS_DENIED / 7 parent-of CLI_WORKTREE / 5 no-subtest internal/cli / 5 parent-of HOMEGUARD_SEAM_REALHOME / 4 parent-of POSIX_MODE_PERM / 2 parent-of BIN_SH_MISSING / 1 parent-of NOMSG(parent/other) / 1 parent-of OWNERSHIP_EPOCH / 1 no-subtest internal/homestate`. G1 네 패키지의 NOMSG 55행 → `37 parent-of SQLITE_URI / 17 parent-of DIRSYNC_ACCESS_DENIED / 1 parent-of OWNERSHIP_EPOCH` (보고서의 G1 55행 귀속을 뒷받침). 부모가 아닌 행 `grep` 확인: `TestHookSinkConcurrentAppendIsIntact`, `TestHookSinkIsNotCreatedWithoutRecords`, `TestHookSinkLevelBoundary/default_config_admits_warn_only`, `TestHookSinkLevelBoundary/lowered_level_admits_info`, `TestHookPathWarnRecordReachesSink`, `TestOpenFactoryReadonlyNeverCheckpointsWAL`, 메시지 모두 `-`. (1회차는 부모 110/비부모 7, 이번 방법은 111/6 — `TestHookSinkLevelBoundary`를 NOMSG 하위 행의 부모로 세느냐의 차이.)
- E7 G2·G3·G4·G5·G6 행 목록을 직접 출력해 확인: G2 15행 모두 세 접두 중 하나, G3 8행 모두 `…/no_git-strategy.yaml` 또는 `…/unparseable_git-strategy.yaml`, 첫 줄이 `got:`에서 끝남. G4 `6 TestTodoPR_QueueDirUnchanged / 5 TestTodoStaleStoreDisclosure_PerVerb / 1 TestTodoStaleStoreDisclosure_AddVerb`. G6 `TestCheckMCPProviderDuplicates_*` 3행(모두 `msg=Claude Code state file unreadable`) + `TestInstalledVersionFromResolvedPath/*` 2행(`installed version = "unknown"`). `TestDoctorTodoStoreDivergence_ThreeStates/*` 2행은 G9.
- E8 코드(직접 읽음): `internal/session/ccversion.go:167, 173, 190-215`; `internal/session/ccversion_test.go:107-168`; `internal/cli/doctor_mcp_provider_test.go:46-69, 117-126`; `internal/cli/doctor_mcp_provider.go:95-106`; `internal/cli/empty_target_guidance_test.go:31-39, 77-95`; `internal/config/integration_target_guidance.go:20-21, 80-102`. 내용은 위 "1회차 결함 처리 확인" 표.
- E9 카드 대조(`moai todo list --json` → `jq`, exit 0): t2118 `SQLite "invalid uri authority: C:" 474건`, t2119 `디렉터리 fsync … Access is denied`, t2123 `① /bin/sh 없음 19건 ② POSIX 권한 비트 단언 46건 ③ home guard…`, t2085 `internal/cli/worktree 실패`, t2130 `21`·t2131 `11`·t2132 `14` POSIX permission-bit (= 46), t2133 `home guard TEMP detection: 54`, t2134 `pre-commit e2e: 12`, t2092 존재. 모두 `picked`.
- E10 보안: `git diff 665a7c5673..HEAD -- progress.md | grep -cEi '(api[_-]?key|token|secret|password|ghp_|sk-[a-z0-9])'` → `0`.
- E11 감사 계획: `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2135` → `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, claude·glm `"gate": "off"`, `"cross_model_active": true`, `"enforced_required": ["codex"]`.

### Baseline-attribution
- 모든 건수는 이번 실행에서 위 1335행 표를 직접 다시 센 값입니다. 보고서나 1회차 감사의 숫자를 옮겨 오지 않았습니다.
- 코드 근거는 이 트리 HEAD 7eb36bd403 기준입니다(1회차 이후 코드 변경 없음, E1). 표 자체는 PR #1969의 한 CI 실행 산출물이라 그 실행의 head와 이 트리 코드가 다를 수 있습니다.
- moai 도구로 측정을 판정하지 않았습니다(`moai todo list`, `moai verify audit-plan` 읽기만). §2.2 빌드 귀속 대상이 없습니다.

### Gaps
- **Codex 필수 게이트 미충족**: 감사 계획이 codex를 `required`로 강제하지만(E11), 이 카드는 디스패치가 Codex 사용을 승인하지 않아 `codex_audit`·`audit_multi`를 포함한 교차 모델 도구를 하나도 호출하지 않았습니다. 받은 영수증은 없습니다(receipts=none). 위 `code verdict:`는 실질 판단이며, 이 게이트만으로도 PASS는 나올 수 없습니다.
- Windows 재현 없음: F1·F2 귀속 확인은 코드 읽기이지 실행 관측이 아닙니다.
- G7 규칙 해석: 보고서 숫자를 재현한 백슬래시 조항은 제 추정이며, 레인이 실제로 쓴 정규식은 보지 못했습니다(R2-2).
- NOMSG 부모 판정은 테스트 이름 접두로만 했습니다. 표에 없는 하위 테스트가 실제 로그에 있었을 가능성은 확인하지 못했습니다.
- 워크트리 격리 가드가 인라인 awk 프로그램과 복합 명령을 거부해, 같은 계산을 스크래치 스크립트(`g.awk`, `run.sh`, `var.sh`, `g9.sh`, `parents.py`, `nosub.py`)로 돌렸습니다. 판정을 가른 출력 줄은 위 Evidence에 옮겼습니다.

### Residual-risk
- 표를 만든 CI 실행 이후 main에 들어간 카드가 일부 행을 이미 고쳤을 수 있습니다. 보고서도 이 점을 Gaps에 적었습니다.
- hook sink 5건은 메시지가 없어 원인 단서가 전혀 없습니다. 카드로 열면 재현이 유일한 첫 단계입니다.
- 스크래치 파일(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/98f002f8-4741-40c2-95ef-a06f9bdaac8b/scratchpad/` 아래)은 기기 로컬이라 인용 대상이 아닙니다.

## 권고 (재감사 범위)
R2-1만 고치면 됩니다. progress.md 23행 문장 하나를 E6 숫자로 바꾸고, 독립 실패 6건을 제안 카드 목록에 넣으십시오. 재감사는 이 차이만 확인합니다. R2-2~R2-4는 선택이며, R2-2를 반영하면 규칙 칸만 읽고도 61/90을 재현할 수 있습니다.
