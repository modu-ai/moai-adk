verdict: PASS
score: 0.94
audited_sha: 4948957a7b1b488c7d66ac4c42583d5969ebbac0
audit_receipt: rcpt-27c7f33ca4d227284477758a
auditor-model: claude-opus-5-5

# t2135 sync-audit (3회차, 최종) — PR #1969 Windows "기타" 실패 분류 보고서 재감사

code verdict: PASS

- 카드: t2135 (보고서 전용, 제품 변경 없음)
- 감사 대상: `.moai/reports/t2135/progress.md` @ 4948957a7b (2회차 대비 변경분 `git diff 7eb36bd403..HEAD`)
- 원본 표: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/43472e68-3604-44b1-a008-c14cce281ea1/scratchpad/categorized.tsv` (1335행)
- Codex 필수 게이트: 레인이 이 HEAD의 깨끗한 트리에서 호출한 결과로 충족 (`rcpt-27c7f33ca4d227284477758a`, `codex_verdict: pass`, `head_commit: 4948957a7b…`). 이 감사는 교차 모델 도구를 다시 부르지 않았습니다.

## 반복 이력

| 회차 | 감사 커밋 | 판정 | 점수 | 차단 결함 | 비고 |
|---|---|---|---|---|---|
| 1 | 665a7c5673 | FAIL | 0.77 | F1 G6 귀속, F2 G3 오독, F3 건수 3곳, F4 그룹 합·NOMSG | F5~F8 선택 |
| 2 | 7eb36bd403 | FAIL | 0.88 | R2-1 (NOMSG "나머지 62행은 모두 부모" 주장 오류, 수리에서 생김) | F1~F3 해소, F4 대부분 해소. R2-2~R2-4 선택 |
| 3 | 4948957a7b | PASS | 0.94 | 없음 | R2-1 해소, R2-2·R2-3 반영, R2-4는 Gap으로 공개. R3-1~R3-3 선택 |

수리 중단 규칙: 이번 회차에는 차단 결함이 없으므로 repair-born 연속 조건이 성립하지 않습니다.

## 판정 요약

2회차의 유일한 차단 결함 R2-1이 해소됐습니다. 표를 다시 세어 보면 G1 네 패키지 밖 NOMSG 62행 가운데 하위 실패가 없는 독립 실패는 정확히 6건이고, 보고서가 이름을 댄 6건(hook sink 계열 5건, homestate WAL 1건)과 그대로 일치합니다. G1 범주의 부모 4행(SQLite 3, dirsync 1)도 맞습니다. 이 6건은 새 제안 카드 9번으로 들어가 더는 묻히지 않습니다.

선택 결함도 정직하게 처리됐습니다. G7 규칙 칸에 백슬래시 조항과 7행 차이, 경로 표기가 아닌 4행을 적었고(R2-2), doctor 2행을 G4 카드 범위에 넣었으며(R2-3), CI head SHA는 값을 적는 대신 "여기 기록되지 않음"을 Gaps에 밝혔습니다(R2-4). 서로소 분할은 다시 돌려도 합계 201이고, 카드 번호는 1~9로 이어지며, 다른 카드로 넘긴 행 목록은 바뀌지 않았습니다. 변경 파일은 `.moai/reports/t2135/` 아래 두 개뿐입니다.

남은 것은 모두 선택 사항입니다. 가장 눈에 띄는 것은 Gaps의 "Codex was not approved for this card" 문장으로, 이 HEAD에서 Codex 감사가 실제로 돌았으므로 지금은 사실과 다릅니다(R3-1). 분류 내용에는 영향이 없습니다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 93/100 | PASS | R2-1 반례 6건과 G1 부모 4행 재현(E3), 서로소 분할 15/8/12/10/5/61/90 = 201 재현(E4), 카드 1~9 연속·다른 카드 목록 불변(E2) |
| Security (25%) | 100/100 | PASS | 카드 변경 파일은 보고서 3개뿐(E1). 변경분 비밀값 패턴 탐색 `0`(E6) |
| Craft (20%) | 92/100 | PASS | 5절 형식 유지, 규칙 문구로 61/90 재현 가능해짐(R2-2 반영), 미기록 SHA를 Gap으로 공개. Gaps의 Codex 문장이 낡음(R3-1) |
| Consistency (15%) | 92/100 | PASS | 결함 번호(R2-1/R2-3/R2-4)와 수리 내역이 본문에 대응. 117 = 110 + 6 + 1 중 1행이 명시되지 않음(R3-2) |

가중 합계: 0.4×93 + 0.25×100 + 0.2×92 + 0.15×92 = 37.2 + 25 + 18.4 + 13.8 = 94.4 → 0.94. 필수 통과 차원(Functionality, Security) 모두 통과. Codex 필수 게이트는 위 영수증으로 충족.

## 2회차 결함 처리 확인

| 결함 | 상태 | 확인 근거 |
|---|---|---|
| R2-1 NOMSG 62행 | 해소 | progress.md:23이 독립 6건·G1 부모 4행을 정확히 명시(E3), 6건은 카드 9(progress.md:40) |
| R2-2 G7 규칙 문구 | 반영 | progress.md:19에 백슬래시 조항, 54/97 → 61/90, 비경로 4행(`\n` 1, awk 경고 2, CRLF 1) 명시. E4에서 정확히 재현 |
| R2-3 doctor 2행 | 반영 | progress.md:36 G4 카드 범위에 `TestDoctorTodoStoreDivergence` 2행 포함. E4에서 두 행이 G9로 떨어짐을 확인 |
| R2-4 head SHA | Gap으로 공개 | progress.md:51 "not recorded here; take it from `gh pr view 1969 --json headRefOid`". 값은 여전히 없으나 미관측을 숨기지 않음 |

## 발견 사항 (결함 목록)

차단 결함 없음. 아래는 모두 선택 사항입니다.

- **R3-1** [Low] [optional] [touched-file] progress.md:53 — Gaps의 "Codex was not approved for this card: Claude audit only."는 이 HEAD 기준으로 사실과 다릅니다. 영수증 `rcpt-27c7f33ca4d227284477758a`가 `head_commit: 4948957a7b…`, `codex_verdict: pass`로 이 커밋에 대한 Codex 감사를 기록합니다(E5). 커밋 시점에는 맞는 문장이었고 분류 내용과는 무관합니다. 수정 제안: 다음에 이 파일을 고칠 일이 있으면 "Codex audit run on 4948957a7b: pass, receipt rcpt-27c7f33ca4d227284477758a"로 바꿉니다. 이 수정만을 위해 커밋을 새로 만들 필요는 없습니다(커밋하면 영수증의 head 결속이 깨짐).
- **R3-2** [Low] [optional] [touched-file] progress.md:4, 23 — Claim의 "110 of them per the first audit"과 23행의 "6 independent"을 합치면 117 중 1행이 설명되지 않습니다. 그 1행은 `TestHookSinkLevelBoundary` 부모 행으로, 하위 실패가 카드 9의 NOMSG 2건뿐입니다(E3: 부모 111, NOMSG가 아닌 하위를 가진 부모 110). 수정 제안: 23행 끝에 "`TestHookSinkLevelBoundary` 1행은 카드 9 하위 테스트의 부모"를 덧붙입니다.
- **R3-3** [Low] [optional] [in-diff] progress.md:33, 36 — doctor 2행이 카드 2의 G9 90행 풀과 카드 5의 범위 양쪽에 걸칩니다. 중복 처리일 뿐 누락은 아니며, 카드 2 탐색에서 이 2행을 건너뛰라고 적으면 충분합니다.

## 근거 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim
1. R2-1은 해소됐다: 62행 중 독립 실패는 정확히 보고서가 이름 댄 6건이고, G1 부모 4행(SQLite 3, dirsync 1)도 맞다.
2. 서로소 분할은 백슬래시 조항을 넣으면 G2 15 · G3 8 · G4 12 · G5 10 · G6 5 · G7 61 · G9 90 = 201로 재현되고, 빼면 G7 54 · G9 97이며 차이 7행 중 4행이 경로 표기가 아니다(보고서 진술과 일치).
3. 카드 번호는 1~9로 이어지고, "Rows assigned elsewhere" 줄은 2회차와 같다.
4. 이번 변경은 `.moai/reports/t2135/` 아래 두 파일만 건드렸다.
5. Codex 필수 게이트는 이 HEAD에 결속된 pass 영수증으로 충족됐다.

### Evidence (이번 실행에서 직접 돌린 명령과 출력, `LC_ALL=C`)
- E1 `git rev-parse HEAD` → `4948957a7b1b488c7d66ac4c42583d5969ebbac0`. `git status --short` → (빈 출력). `git diff 7eb36bd403..HEAD --stat` → `.moai/reports/t2135/progress.md | 9 +-`, `.moai/reports/t2135/sync-audit.md | 170 ++++++++++++++----------------------`, `2 files changed, 79 insertions(+), 100 deletions(-)`. `git show --stat HEAD`도 같은 두 파일. `git diff --name-only main...HEAD` → `card-review.md`, `progress.md`, `sync-audit.md` (모두 `.moai/reports/t2135/`).
- E2 `git diff 7eb36bd403..HEAD -- progress.md`: 바뀐 줄은 19(G7 행), 23(NOMSG 문장), 36(카드 5), 40(카드 9 추가), 50·51(Gaps 두 줄 추가)뿐. "Rows assigned elsewhere …" 줄은 diff 맥락에 변경 없이 나타남. 제안 카드 번호 `1.`~`9.` 연속.
- E2b `wc -l` → `1335`. `cut -f2 | sort | uniq -c` → `474 SQLITE_URI / 326 DIRSYNC_ACCESS_DENIED / 201 OTHER / 117 NOMSG(parent/other) / 60 CLI_WORKTREE / 54 HOMEGUARD_SEAM_REALHOME / 46 POSIX_MODE_PERM / 26 OWNERSHIP_EPOCH / 19 BIN_SH_MISSING / 12 PRECOMMIT_E2E`.
- E3 NOMSG 부모 판정(awk, 같은 패키지에 `<이름>/` 하위 행이 있으면 부모) → 스크래치 `nomsg.tsv` 117행. G1 네 패키지 밖 `62`. 그중 하위 실패 0건:
  `internal/cli TestHookSinkConcurrentAppendIsIntact 0`, `internal/cli TestHookSinkIsNotCreatedWithoutRecords 0`, `internal/cli TestHookSinkLevelBoundary/default_config_admits_warn_only 0`, `internal/cli TestHookSinkLevelBoundary/lowered_level_admits_info 0`, `internal/cli TestHookPathWarnRecordReachesSink 0`, `internal/homestate TestOpenFactoryReadonlyNeverCheckpointsWAL 0` → 6건, 보고서 목록과 일치. G1 안 독립 실패 `0`.
  62행 하위 범주 분포: `33 OTHER / 7 CLI_WORKTREE / 5 HOMEGUARD_SEAM_REALHOME / 4 POSIX_MODE_PERM / 3 SQLITE_URI / 1 NOMSG / 1 DIRSYNC_ACCESS_DENIED / 1 BIN_SH_MISSING,NOMSG / 1 BIN_SH_MISSING` → G1 범주 부모 3 + 1 = 4. NOMSG 하위만 가진 부모: `internal/cli TestHookSinkLevelBoundary 2 NOMSG(parent/other)`. 부모 총수 `111`, NOMSG가 아닌 하위를 가진 부모 `110`.
- E4 서로소 분할(스크래치 `r3.sh`: 2회차 `g.awk` 규칙 + G9 중 메시지에 백슬래시가 있으면 G7):
  without clause → `15 G2 / 8 G3 / 12 G4 / 10 G5 / 5 G6 / 54 G7 / 97 G9`; with backslash clause → `15 G2 / 8 G3 / 12 G4 / 10 G5 / 5 G6 / 61 G7 / 90 G9`; total `201`.
  조항으로 옮겨진 7행: `TestAstGrepCmd_UsageErrorIsNotRulesetHeadline (stdout="no findings\\n")`, `TestMigrationArchiveRefusesSymlinkedArchiveDestination/flow_aborts_before_any_removal`, `TestProductionStringLiteralsUseLeaderLaneVocabulary (..\\factory\\role.go)`, `TestPlanAuditTraceability_ZeroDefinitionsIsGap`, `TestPlanAuditTraceability_InlineACsWithoutAcceptance`, `TestMergeStepRealContentConflictAbortsAndReleases (card.txt = "release work\\r\\n")`, `TestPlan_IsReadOnly_MaintenanceLockRepro (plan-repo\\\\.git\\\\objects…)`.
  doctor 행 → `G9 TestDoctorTodoStoreDivergence_ThreeStates/divergent`, `G9 TestDoctorTodoStoreDivergence_ThreeStates/unreadable-store`.
- E5 영수증 파일 `.moai/state/audit-receipts/receipts/rcpt-27c7f33ca4d227284477758a.json` 읽음 → `"tool": "codex_audit"`, `"tree_root": "/Users/goos/.moai/worktrees/projects/moai-adk-go/t2135"`, `"created_at": "2026-11-06T18:45:48.469969Z"`, `"codex_verdict": "pass"`, `"head_commit": "4948957a7b1b488c7d66ac4c42583d5969ebbac0"`.
- E6 보안: `git diff 7eb36bd403..HEAD -- progress.md | grep -cEi '(api[_-]?key|token|secret|password|ghp_|sk-[a-z0-9])'` → `0`.
- E7 감사 계획: `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2135` → `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, claude·glm `"gate": "off"`, `"cross_model_active": true`, `"enforced_required": ["codex"]`.

### Baseline-attribution
- 모든 건수는 이번 실행에서 위 1335행 표를 직접 다시 센 값입니다. 2회차 숫자를 옮겨 오지 않았습니다(같은 규칙 파일 `g.awk`를 다시 실행).
- 트리 기준은 HEAD 4948957a7b입니다. 표는 PR #1969의 한 CI 실행 산출물이라 그 head와 이 트리는 다를 수 있습니다.
- moai 도구로 측정을 판정하지 않았습니다(`moai verify audit-plan` 읽기만). §2.2 빌드 귀속 대상이 없습니다.

### Gaps
- Codex 감사는 이 감사가 직접 돌리지 않았습니다. 지시대로 레인의 기존 호출 결과를 영수증 파일로만 확인했습니다(E5). Codex의 지적 내용 자체는 보지 못했습니다(findings 0이라는 진술은 디스패치에서 받은 것).
- Windows 재현 없음: 귀속은 여전히 코드 읽기와 메시지 기반 추론입니다(보고서도 같은 Gap 명시).
- NOMSG 부모 판정은 테스트 이름 접두로만 했습니다. 표에 없는 하위 테스트가 실제 로그에 있었을 가능성은 확인하지 못했습니다.
- 백슬래시 조항이 레인이 실제로 쓴 규칙인지는 보지 못했습니다. 보고서 문구대로 적용하면 숫자가 재현된다는 것만 확인했습니다.
- CI head SHA는 이번에도 확인하지 않았습니다(`gh` 호출 안 함).
- 워크트리 격리 가드가 `$(git rev-parse …)`가 든 복합 명령 한 건을 거부해, 같은 내용을 절대 경로로 나눠 다시 실행했습니다. 판정에 영향 없음.

### Residual-risk
- 표를 만든 CI 실행 이후 main에 들어간 카드(t2063~t2091)가 일부 행을 이미 고쳤을 수 있습니다. 보고서도 Gaps에 적었습니다.
- 카드 9의 6건은 메시지가 없어 원인 단서가 없습니다. 재현이 유일한 첫 단계입니다.
- 이 판정 파일을 쓰면 트리가 영수증 시점과 달라집니다. 정지 훅이 영수증을 "이 감사 시작 전 발급"으로 거부할 수 있으며, 리더 말로는 t2138이 그 결함을 다룹니다.
- 스크래치 파일(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/98f002f8-4741-40c2-95ef-a06f9bdaac8b/scratchpad/` 아래 `r3.sh`, `nomsg.tsv`, `r3g.tsv` 등)은 기기 로컬이라 인용 대상이 아닙니다. 판정을 가른 출력 줄은 위 Evidence에 옮겼습니다.

## 권고
차단 결함이 없으므로 이 카드는 감사를 마칠 수 있습니다. R3-1~R3-3은 선택이며, 고치려면 영수증의 head 결속이 깨지는 점을 감안해 다른 변경과 함께 처리하십시오.
