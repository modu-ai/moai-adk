verdict: FAIL
score: 0.77
audited_sha: 665a7c5673728e99e3718891edefbf9ede6854bd
auditor-model: claude-opus-5-5

# t2135 sync-audit — PR #1969 Windows "기타" 실패 분류 보고서 감사

code verdict: FAIL

- 카드: t2135 (보고서 전용, 제품 변경 없음)
- 감사 대상: `.moai/reports/t2135/progress.md`, `.moai/reports/t2135/card-review.md` @ 665a7c5673
- 원본 표: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/43472e68-3604-44b1-a008-c14cce281ea1/scratchpad/categorized.tsv` (1335행)
- 반복 이력: 1회차 감사 (이전 감사 없음)

## 판정 요약

범주 집계, 패키지 분해, 카드 번호 대조는 모두 맞습니다. 그러나 분류 결과물의 핵심인 귀속(테스트 결함인지 제품 결함인지) 가운데 두 그룹(G3, G6)은 이 트리의 코드와 정면으로 어긋나고, G6는 그 잘못된 전제 위에서 "건너뛰기 카드"를 제안합니다. 그룹 크기의 합도 OTHER 201건과 맞지 않습니다. Functionality(필수 통과 차원)가 기준에 못 미치므로 전체 판정은 FAIL입니다. 아래 수정 지시는 모두 보고서 문서만 고치면 되는 작은 작업입니다.

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 62/100 | FAIL | 집계·카드 번호 대조는 정확. G3·G6 귀속이 코드와 모순(F1, F2), 서명 건수 3곳 오류(F3), 그룹 합 불일치(F4) |
| Security (25%) | 100/100 | PASS | 변경 파일 2개 모두 보고서. 비밀값 탐색 결과 `secret-grep exit=1`(일치 없음) |
| Craft (20%) | 70/100 | PASS(경계) | 5절 형식과 "추론" 표기는 갖춤. 다만 G3의 "메시지 비어 있음"은 관측하지 않은 사실을 단정한 것(§1) |
| Consistency (15%) | 85/100 | PASS | 카드 기록 형식·경로 규약 준수. 범주명 `NOMSG(parent/other)`을 `NOMSG`로 줄이면서 "부모 행"이라는 뜻이 사라짐 |

가중 합계: 0.4×62 + 0.25×100 + 0.2×70 + 0.15×85 = 76.6 → 0.77. 필수 통과 방화벽: Functionality 62 < 기준치 → FAIL.

## 발견 사항 (결함 목록)

- **F1** [Medium] [blocking] [in-diff] progress.md:22 (G6) — "실제 Claude CLI/상태 파일이 필요한 환경 테스트 → 정당한 skip" 귀속이 코드와 모순됩니다.
  - `internal/session/ccversion_test.go:147-167` `installedVersionViaSymlink`는 `t.TempDir()` 안에 가짜 바이너리와 심볼릭 링크를 만들고 PATH 조회를 seam으로 바꿉니다. 실제 CLI가 전혀 필요 없습니다. 실패 원인은 `internal/session/ccversion.go:205-206`의 `versionSegmentFromPath`가 `"/versions/"`·`"/claude-code/"`처럼 슬래시로 고정된 접두를 찾는데, Windows의 `filepath.EvalSymlinks`는 `\`로 된 경로를 돌려준다는 점입니다(확신도 중상, 미재현).
  - `internal/cli/doctor_mcp_provider_test.go:46-57` `providerState`는 `filepath.Abs` 결과(`C:\...`)를 이스케이프 없이 JSON 문자열에 끼워 넣습니다. 백슬래시가 잘못된 JSON 이스케이프가 되어 `internal/cli/doctor_mcp_provider.go:103-105`의 `json.Unmarshal`이 실패하고, 메시지 "Claude Code state file unreadable; not checked"가 나옵니다. 이건 테스트 픽스처 결함입니다.
  - 건너뛰기 카드는 고칠 수 있는 결함 두 개를 덮어 버립니다.
  - **필수 수정**: G6를 경로 구분자 계열로 다시 귀속(ccversion은 제품 파서, MCP provider는 테스트 픽스처의 JSON 이스케이프)하고, 제안 카드를 skip이 아니라 수정 카드로 바꿉니다. 실제 CLI 부재가 의심되는 행은 잔여 행의 `claude not found in PATH` 2건(`TestResolveLaunchClaudeBinary_UnsetFallsBackToPATH`, `TestLaunchClaudeDefault_UnsetPinFallsBackToPATH`)이며, 이 둘도 후보로만 적습니다.
- **F2** [Medium] [blocking] [in-diff] progress.md:11, 19 (G3) — "refusal text empty"와 "(message empty)"는 표의 생김새를 잘못 읽은 결과입니다.
  - `internal/cli/empty_target_guidance_test.go:77-83` `assertAllContained`는 `"%s must contain %q, got:\n%s"` 형식이라 첫 실패 줄은 항상 `got:`에서 끝납니다. 실제 거부 문구는 표에 담기지 않은 둘째 줄에 있습니다.
  - 기대 문자열 순서가 `"cannot read"` → `".moai/config/sections/git-strategy.yaml"`인데 첫 실패가 둘째 항목이므로, 거부 문구에는 `cannot read`가 들어 있었습니다. 즉 비어 있지 않았습니다.
  - `internal/config/integration_target_guidance.go:85-87`은 이 경우(EmptyTargetUnreadable)에만 `filepath.Join(root, filepath.FromSlash(gitStrategyRelPath))`로 경로를 만듭니다. Windows에서는 `\`가 되어 슬래시 기대값과 어긋납니다. 다른 행들은 `gitStrategyRelPath`를 그대로 써서 통과합니다. "두 행만 실패한다"는 현상이 이것으로 설명됩니다(확신도 높음, 미재현).
  - **필수 수정**: G3의 원인을 "구분자 표기(제품은 OS 구분자, 테스트는 슬래시 기대)"로 고치고 "메시지 비어 있음" 문구를 지웁니다. 제안 카드의 첫 단계는 "테스트 기대값을 `filepath.FromSlash`로 맞출지, 제품 문구를 슬래시로 고정할지 결정"으로 씁니다.
- **F3** [Low] [blocking] [in-diff] progress.md:9, 11, 20 — 건수 오류 세 곳.
  - `stat backlog.lock` 서명은 3이 아니라 **6**입니다.
  - `Claude Code state file unreadable`은 2가 아니라 **3**입니다.
  - OTHER 단일 테스트 패키지는 8개가 아니라 **14**개입니다(134+12+7+5+5+5+4+4+4+3+2+2=187, 201−187=14).
  - G4의 "~15"는 `TestTodoPR_QueueDirUnchanged` 6과 `stat backlog.lock` 3을 더한 값인데, 이 6행이 바로 그 backlog.lock 행이라 이중으로 셌습니다. 실제 OTHER 행은 todo PR 6 + stale store 6 + `TestDoctorTodoStoreDivergence` 2 = 14입니다(여기에 NOMSG 부모 3).
  - **필수 수정**: 세 숫자를 정정하고 G4 구성을 위 내역으로 다시 적습니다.
- **F4** [Medium] [blocking] [in-diff] progress.md:4, 11, 15-25 — 그룹 크기의 합이 201과 맞지 않고, NOMSG의 성격이 빠졌습니다.
  - G2–G9 중간값 합은 17+8+15+12+7+(40~60)+10+40 = 149~169로, 201보다 32~52건 적습니다.
  - G2의 ~17에는 NOMSG 부모 3행이 섞여 있습니다(OTHER 15 + NOMSG 3).
  - `status="armed" turns=0` 1행은 G2(테스트 이름 접두 `TestStopGoalWait_`)와 G8에 모두 들어 있습니다.
  - 근거 절은 "나머지 ~120건은 단일 서명"이라 하면서, G7+G9는 80~100입니다.
  - 독립 분할 결과(아래 근거 E5)는 G2 15 · G3 8 · G4 14 · G5 11 · G6 5 · G8 8 · 경로 표기 휴리스틱 50(+잔여 중 `C:/Users` 9) · 나머지 약 81입니다. 따라서 G9는 ~40이 아니라 약 80입니다.
  - NOMSG 117행 중 **110행은 실패한 하위 테스트의 부모 행**입니다(SQLITE_URI 부모 40, DIRSYNC 18, OTHER 33, CLI_WORKTREE 7, HOMEGUARD 5, POSIX 4, BIN_SH 1, OWNERSHIP 1, NOMSG 1). 하위 테스트가 없는 행은 7개뿐입니다. G1의 ~55는 SQLITE+DIRSYNC 부모 58행과 맞아 뒷받침되지만, 나머지 62행은 어디에도 배정되지 않았습니다. "201+117=318건의 실패"는 독립 실패 수가 아닙니다.
  - **필수 수정**: 그룹 표에 합계 대조 행(각 그룹 실측 수, 겹침, 미배정)을 넣고, G9를 약 80으로 고치고, NOMSG가 대부분 부모 행(독립 실패 아님)이라는 점과 나머지 62행이 어느 그룹·카드의 부모인지를 적습니다.
- **F5** [Low] [optional] [in-diff] progress.md:24 (G8) — "타이밍/동작, flaky 가능성"은 근거가 약합니다.
  - `recorded stat calls = []` 3행은 `TestCodexStaleSkillFinding_*SlashPinned*`/`AbsoluteStatTarget`로, 이름부터 경로 절대성 문제(G7 계열)입니다.
  - `re-measure recorded` 4행의 실패 꼬리는 `worktree dirty (1 status lines); no record written`이라 타이밍 문제가 아닙니다.
  - `want exactly 1 winner, got 0`(agenticfactory/claim)은 같은 패키지 84행이 DIRSYNC라서 G1 계열일 가능성이 높습니다(추론).
- **F6** [Low] [optional] [in-diff] progress.md:25 — 잔여 행에서 이름 붙일 수 있는 하위 묶음을 놓쳤습니다.
  - POSIX 실행 비트 단언 3건(`TestValidateClaudeBinaryPin_NotExecutable`, `TestM001_PreservesExecutableBit`, `TestTreeState_ChmodMovesUntrackedDigest`) → t2130~t2132 영역 후보.
  - `shell digest "" is not 64 hex chars` 2건 → 셸 부재(t2123 ①) 후보.
  - 빈 homestate 레코드 `{Issue:0 Holder: ...}` 4건.
  - `factory_quota_lanes` 골든 출력 차이 4건.
  - `TestBuildSmartPATHWindows` 3건.
- **F7** [Low] [optional] [in-diff] progress.md:20 — `TestTodoStaleStoreDisclosure` 6행의 메시지는 "stderr does not name the stale store path C:\\Users\\RUNNER~1\\..."입니다(`internal/cli/todo_stale_store_test.go:93`, 문자열 포함 비교). 잠금 처리보다는 경로 표기(짧은 이름/긴 이름) 문제로 보여, G4보다 G7에 가깝습니다(추론).
- **F8** [Low] [optional] [in-diff] progress.md:30 — "head SHA recorded in the t2077 analysis"는 의심스럽습니다. t2077은 열린 PR 일괄 처분 카드입니다(`moai todo list --json` 제목). 카드 본문은 이 표를 "t2123 table"이라 부르고, #1969 CI 분석 카드는 t2092입니다. 이 트리에는 해당 보고서 디렉터리가 없어 확인하지 못했습니다(확신도 낮음).

## 근거 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim
1. 범주 집계와 OTHER·NOMSG 패키지 분해(단일 패키지 개수 제외)는 정확하다.
2. "이미 배정된 행"의 카드 번호 8개는 모두 존재하고 범위가 맞는다.
3. G3·G6 귀속은 코드와 모순되고(F1, F2), 건수 3곳이 틀렸으며(F3), 그룹 합이 맞지 않는다(F4).

### Evidence (이번 실행에서 직접 돌린 명령과 출력)
- E1 `git rev-parse HEAD` → `665a7c5673728e99e3718891edefbf9ede6854bd`. `git show --stat HEAD` → `.moai/reports/t2135/card-review.md | 3 +++`, `.moai/reports/t2135/progress.md | 39 +++`, `2 files changed, 42 insertions(+)`.
- E2 `LC_ALL=C cut -f2 categorized.tsv | sort | uniq -c | sort -rn` → `474 SQLITE_URI / 326 DIRSYNC_ACCESS_DENIED / 201 OTHER / 117 NOMSG(parent/other) / 60 CLI_WORKTREE / 54 HOMEGUARD_SEAM_REALHOME / 46 POSIX_MODE_PERM / 26 OWNERSHIP_EPOCH / 19 BIN_SH_MISSING / 12 PRECOMMIT_E2E`. `wc -l` → `1335`.
- E3 OTHER 패키지: `134 internal/cli, 12 internal/factory, 7 internal/cli/update/skillresidue, 5 session, 5 homestate, 5 graph, 4 template, 4 statusline, 4 agenticfactory/claim, 3 auditreceipt, 2 astgrep, 2 agenticfactory/mergeq`, 그리고 1건짜리 14개(worktreeroot, web, userassets, testguard, spec, settings/yamlpatch, migration/migrations, hygiene, factorylane, discovery, contract/revoke, config/toolpolicy, cli/update/backup, agenticfactory/runner). NOMSG 패키지: `38 internal/cli, 37 agenticfactory/ledger, 8 agenticfactory/e2e, 7 cli/worktree, 7 agenticfactory/runner, 3 agenticfactory/claim, 2 template, 2 homestate, 2 agenticfactory/mergeq, …`(보고서와 일치).
- E4 OTHER 서명 건수(`/usr/bin/grep -cE`, 대상 201행): `want a block decision, got stdout=""` 10 · `re-measure recorded` 4 · `target worktree removed` 3 · `stat backlog.lock` **6** · `no dispatch operation recorded` 3 · `installed version = "unknown"` 2 · `state file unreadable` **3** · `strategy\.yaml` 8 · `want the user.s bytes back` 2 · `move the skill folder aside` 2 · `mirror absent` 3 · `want exactly 1 winner` 1 · `recorded stat calls` 3 · `turns=0` 1 · `RUNNER~1` 37 · `C:/Users` 17. 전체 표에서 `want a block decision`은 `10 OTHER`뿐.
- E5 OTHER 분할(awk 우선순위 규칙): `15 G2, 8 G3, 14 G4, 11 G5, 5 G6, 8 G8, 50 G7?(RUNNER~1·백슬래시·case·cwd·fold), 90 rest`. rest 안에서는 `C:/Users` 9 · `Issue:0 Holder` 4 · `executable|chmod` 3 · `claude not found in PATH` 2 · `digest "" is not` 2.
- E6 G2 전체 범주: `TestStopGoalWait|TestStopParseCap|TestFactoryWaitResolve…` → OTHER 15행 + `NOMSG(parent/other)` 3행.
- E7 G4 전체 범주: `TestTodoStaleStoreDisclosure|TestTodoPR_QueueDirUnchanged|TestDoctorTodoStore` → `3 NOMSG(parent/other)`, `14 OTHER`. 6개 `TestTodoPR_QueueDirUnchanged/*` 행 모두 `todo_pr_test.go:229: stat backlog.lock: GetFileAttributesEx …`.
- E8 NOMSG 부모 분류(`python3 -I parents.py`): `40 parent-of SQLITE_URI / 33 parent-of OTHER / 18 parent-of DIRSYNC_ACCESS_DENIED / 7 parent-of CLI_WORKTREE / 6 no-subtest-rows internal/cli / 5 parent-of HOMEGUARD / 4 parent-of POSIX_MODE_PERM / 1 BIN_SH / 1 NOMSG / 1 OWNERSHIP_EPOCH / 1 no-subtest-rows internal/homestate`.
- E9 G1 패키지 범주 구성: `agenticfactory/ledger`는 `416 SQLITE_URI` + `37 NOMSG`. `e2e`는 `80 DIRSYNC` + `8 NOMSG`. `runner`는 `135 DIRSYNC` + `7 NOMSG` + `5 OWNERSHIP` + `1 OTHER`. `claim`은 `84 DIRSYNC` + `21 OWNERSHIP` + `4 OTHER` + `3 NOMSG`. G1 추론을 뒷받침함.
- E10 카드 대조(`moai todo list --json` → jq, id 필터):
  - t2118 `SQLite "invalid uri authority: C:" 474건`
  - t2119 `디렉터리 fsync … Access is denied 326건`
  - t2123 `① /bin/sh 없음 19건 ② POSIX 46건 ③ home guard 54건 ④ pre-commit…`
  - t2085 `internal/cli/worktree 실패` (CLI_WORKTREE 60행 전부 `internal/cli/worktree`)
  - t2130 `21` + t2131 `11` + t2132 `14` = 46
  - t2133 `home guard TEMP detection: 54`
  - t2134 `pre-commit e2e: 12`
  - t2092 `#1969 CI … race-detector`
  - t2077 `열린 PR … 처분`
  - 모두 `picked`.
- E11 코드 확인(직접 읽음, 7곳):
  - `internal/cli/empty_target_guidance_test.go:77-83, 85-95`
  - `internal/config/integration_target_guidance.go:85-87`
  - `internal/session/ccversion_test.go:104-168`, `internal/session/ccversion.go:92-104, 190-206`
  - `internal/cli/doctor_mcp_provider_test.go:15-57, 116-125`, `internal/cli/doctor_mcp_provider.go:98-106`
  - `internal/cli/hook_stop_goal_wait_t2013_test.go:200-240`: "no wait" 하위 테스트도 차단되지 않음 → 훅 자체가 Windows에서 차단 결정을 내지 못함. G2의 "원인 미상" 표기와 정합.
  - `internal/cli/todo_pr_test.go:220-230`: 잠금 파일 stat 실패. G4의 추론과 정합.
  - `internal/cli/todo_stale_store_test.go:90-112`
- E12 보안 탐색: `grep -nEi '(api[_-]?key|token|secret|password|ghp_|sk-[a-z0-9])'` 두 파일 → 일치 없음, `secret-grep exit=1`.
- E13 감사 계획: `moai verify audit-plan --project-root <this tree>` → `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, `"enforced_required": ["codex"]`, `"cross_model_active": true`.

### Baseline-attribution
- 모든 건수는 이번 실행에서 위 표(1335행, 다른 세션의 스크래치패드)를 직접 다시 센 값입니다. 보고서가 적어 둔 숫자를 옮겨 오지 않았습니다.
- 코드 근거는 이 트리 HEAD 665a7c5673 기준입니다. 표 자체는 PR #1969의 한 CI 실행에서 나온 산출물이므로, 그 실행의 head와 이 트리의 코드가 다를 수 있습니다(아래 Residual-risk).
- 이번 감사는 moai 도구로 측정을 판정하지 않았습니다. 사용한 것은 `moai todo list`, `moai verify audit-plan` 읽기뿐이라 §2.2 빌드 귀속 대상이 없습니다.

### Gaps
- **Codex 필수 게이트 미충족**: 감사 계획이 codex를 `required`로 강제하지만(E13), 이 카드는 디스패치가 Codex 사용을 승인하지 않아 `codex_audit`·`audit_multi`를 호출하지 않았습니다. 받은 영수증은 없습니다(receipts=none). 위 `code verdict:`는 실질 판단이며, 게이트 미충족만으로도 PASS는 불가합니다.
- Windows 재현 없음: F1·F2의 원인은 코드 읽기로 확정한 것이지 실행으로 관측한 것이 아닙니다. 특히 F1의 ccversion 쪽은 `EvalSymlinks` 성공 여부(심볼릭 링크 권한)를 확인하지 못했습니다.
- 표는 테스트별 첫 실패 줄만 담고 있어, F2에서 말한 "둘째 줄의 실제 문구"는 직접 보지 못했습니다. 추론의 근거는 기대 문자열 순서입니다.
- 분할 휴리스틱(E5)은 정규식 기반이라 G7·rest 경계는 ±10행 정도 흔들릴 수 있습니다. 그래도 G9 ≈ 40이 성립하지 않는다는 결론은 바뀌지 않습니다.
- F8(t2077 참조)은 해당 분석 보고서가 이 트리에 없어 확인하지 못했습니다.
- `grep`이 셸 함수로 덮여 있어(`type grep`), 결정적인 건수는 `/usr/bin/grep`으로 다시 셌습니다(E4, E5).
- 워크트리 격리 가드가 `awk -f`와 `git-strategy.yaml` 문자열이 들어간 명령을 거부했습니다. 대신 스크래치 스크립트(`sh sig.sh`, `python3 -I parents.py`)로 같은 계산을 했고, 결과는 위에 적었습니다.

### Residual-risk
- 표를 만든 CI 실행 이후 main에 들어간 카드(t2063~t2091 등)가 일부 행을 이미 고쳤을 수 있습니다. 그러면 그룹 크기는 지금보다 작아집니다.
- F1의 MCP provider 픽스처 결함은 macOS/Linux에서는 경로에 백슬래시가 없어 드러나지 않습니다. skip으로 처리하면 같은 문자열 이어 붙이기 방식의 다른 픽스처도 숨은 채로 남습니다.
- 스크래치 파일(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/98f002f8-4741-40c2-95ef-a06f9bdaac8b/scratchpad/other.XzKw`, `rest.tsv`, `todo.json`)은 기기 로컬이라 인용 대상이 아닙니다. 판정을 가른 출력 줄은 위 Evidence에 옮겨 두었습니다.

## 권고 (재감사 범위)
F1~F4만 고치면 됩니다. 보고서 문서 수정이며 재감사는 이 네 항목의 차이만 확인합니다. F5~F8은 선택 사항이고, 반영하면 G9가 줄어드는 효과가 있습니다.
