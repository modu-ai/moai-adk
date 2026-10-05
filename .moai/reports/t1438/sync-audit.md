auditor-model: glm-5.3-flash[1m]

# Sync Audit Report — SPEC-INIT-SHRINK-001 (card t1438)

verdict: PASS-WITH-DEBT
audited_sha: 6038f2ec594dc7ec5490137ddf2d8a2d799429a3

- **Overall Verdict**: PASS-WITH-DEBT — 총점 93.5/100 (조화평균). 차단 결함 0건, 소스 검증된 비차단 채무 4건(S-1..S-4) + 문서 완결성 관찰 1건(S-5).
- **Audited SHA**: `6038f2ec594dc7ec5490137ddf2d8a2d799429a3` (branch `WT-moai-init-slim`, 작업 트리 클린 — 감사 창수립 시 `git status --porcelain` 0행 재확인).
- **Close chain**: M1 `d8e4300e4` → M2 `fd877e59a` → M3 `3b04fc13a` → M4 `e22a41bee` → gate-fix `b84d7baca` → card-review-fix `04b2900b0` (15 파일) → sync `c1623e65c` → E.4 백필 `a2b92451f` → E.3 백필 `6038f2ec5`.
- **Artifact references**: `progress.md` §E.2/§E.3/§E.4, `.moai/reports/t1438/card-review.md` (6건 전부 `04b2900b0` 수리), `.moai/reports/t1438/plan-audit-iter5.md` (PASS-WITH-DEBT 0.93, rcpt-91c53431e86cd49b8a2019ab), `CHANGELOG.md` [Unreleased] Added (1 occurrence), README ×4 + docs-site `cli-reference/init.md` ×4.
- **수용 처분 (재지적하지 않음)**: internal/cli 84.9%·internal/template 83.4%·internal/config 83.7% vs 85% 패키지 집계 목표(부분 커버리지 처분 — 신규 코드는 자체 테스트로 전량 커버) · 기존 적색 `TestCodex1718Fixtures_WidenedContent`(develop `7536afe60` 기대 갱신 대기의 base-staleness) · D-18/D-19 기록된 선택 채무 · REQ-008 fail=1 마커-기제 주석(해소 자체는 실증, 마커 그랜트는 `b84d7baca`에서 확대).

## 1. Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | AC 21/21 전수 재관측 GREEN(§2 표 — 판정 근거는 전부 `--- PASS:` 라인). 차감 사유: S-1/S-2의 REQ-006 폴백 비대칭(AC 층은 녹색, REQ 본문 축의 기록 채무) |
| Security (25%) | 94/100 | PASS | 마이그레이션 파일-연산 렌즈 클린(§5) — 심링크 무역참조(분류·아카이브·대상)·프로젝트 루트 밖 무기입·아카이브 우선 제거 순서 확인. 차감: S-2 실패 팔의 댕글 링크 |
| Craft (20%) | 90/100 | PASS | go vet 0 · golangci-lint v2.1.6 `0 issues.` · gofmt 35 파일 0 · 신규 소관 패키지 커버리지 재측정 backup 86.6/deploy 91.6/merge 93.1/plan 95.0/report 92.9/config 83.7(§E.3 수치와 정확히 일치). 집계 목표 미달 3 패키지는 수용 처분 |
| Consistency (15%) | 95/100 | PASS | 경계 grep 0(접촉 비테스트 Go 파일) · `@MX:ANCHOR` ×2 실재 + 신규 파일 `@MX:TODO` 0 · Conventional Commits + `Card: t1438` 트레일러 전 행 · README×4·docs-site×4 로케일 패리티(h2=5/h3=4, 'no-plugin' README 1·docs-site 2) |
| **Harmonic mean** | **93.5/100** | **PASS-WITH-DEBT** | 4/(1/95+1/94+1/90+1/95) = 93.46 → 93.5 |

## 2. AC 재실행 표 — §E.2를 신뢰하지 않고 전수 재관측

모든 실행은 환경세척 복합 형태( `unset MOAI_AUTONOMY_TIER MOAI_CONFIG_SOURCE MOAI_FACTORY_AUTO_DISPATCH MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN MOAI_SESSION_PID && go test …` )로, `-count=1 -v`, 이 트리 6038f2ec5, exit 0. 판정은 `--- PASS:` 라인만으로.

| # | 명령 (세척 전제 공통) | 판정 출력 (verbatim) |
|---|---|---|
| 1 | `go test ./internal/cli -run '^(TestDefaultDeploySetExcludesSkillsAndCommands\|TestNoPluginPathDeploysFullLocalPayload\|TestDeployModeRecordRoundTrip\|TestMigrationClassification\|TestMigrationLeavesForeignFilesUntouched\|TestMigrationRemovesIdenticalDroppedComponents\|TestMigrationArchivesModifiedBeforeRemoval\|TestMigrationIdempotent\|TestUpdateMigratesLegacyProject\|TestDefaultPathMcpEntryPolicy\|TestResolutionGateHarness\|TestMigrationArchiveRefusesSymlinkedArchiveDestination\|TestMigrationRehomesExistingMirrorEntries\|TestMirrorHealRespectsDeployMode\|TestPreviewManagedCleanup_ModeScoped)$' -count=1 -v` | `--- PASS:` ×15 (전 건) + `ok github.com/modu-ai/moai-adk/internal/cli 8.396s` |
| 2 | `go test ./internal/template -run '^(TestDeployerModeSplitsFileSet\|TestCodexMirrorFollowsDeployMode\|TestEmbeddedSkillAndCommandSourcesRetained)$' -count=1 -v` | `--- PASS:` ×3 + `ok …internal/template 0.391s` |
| 3 | `go test ./internal/cli -run 'Preview' -count=1 -v` | `--- PASS: TestPreviewManagedCleanup (0.00s)`·`_EmptyProject`·`_ModeScoped` 외 3건 = ×6 + `ok … 0.737s` |
| 4 | `go test ./internal/cli -run '^(TestUpdatePluginModeSkipsDroppedRedeploy\|TestUpdateLocalModeKeepsFullScope\|TestUpdateNeverFlipsModeRecord\|TestUpdateForceDoesNotResurrectDropped\|TestNotDemonstratedPreservationEndsAtNextLocalUpdate)$' -count=1 -v` | `--- PASS:` ×5 + `ok … 2.596s` |
| 5 | `go test ./internal/cli -run '^(TestShrinkInitGuidanceOnMissingPlugin\|TestInitDocsDescribeThinDeploy)$' -count=1 -v` | `--- PASS:` ×2 (AC-004, AC-021a) |
| 6 | `go test ./internal/cli -run '^(TestAllFlagDeploysAllTiersLocally\|TestShrinkVerificationNeverReachesRealHome)$' -count=1 -v` | `--- PASS:` ×2 (AC-007, AC-020a) |
| 7 | `go test ./internal/cli -run '^(TestUpdateLLMYAMLFirstDeployCalm\|TestUpdateMirrorHeal_RestoresPathA)$' -count=1 -v` | `--- PASS:` ×2 — §E.2 M4에서 "플립 수변 착지 수정" 처분된 2건의 수정 유지 재확인 |
| 8 | `sh scripts/check-bare-name-resolution.sh --self-check` | `PASS isolation-scrub`·`PASS isolation-scrub-negative-control`·`PASS protected-set-hash`·`PASS protected-set-hash-negative-control`·`PASS shape-lines`·`PASS fixture-present` + `RESULT pass=6 fail=0`, exit 0 (F5의 음성 대조 굉장 — 변조 감지 비공허 실측) |
| 9 | `make commands-emit-check agents-emit-check` | `ok …commandemit 0.323s`·`ok …agentemit 0.344s`, exit 0 (AC-002b) |
| 10 | `go vet ./internal/cli/... ./internal/template/... ./internal/config/...` | 출력 0, exit 0 |
| 11 | `golangci-lint run --timeout=2m ./internal/cli/... ./internal/template/... ./internal/config/...` | `0 issues.` (v2.1.6 — CI 판과 동일 버전) |
| 12 | `git diff --name-only 0a1e105d8..HEAD -- '*.go'` 대상 `gofmt -l` | 빈 출력 (35 파일) |
| 13 | 접촉 비테스트 Go 파일 대상 `grep -ln 'AskUserQuestion\|mcp__askuser'` | 0 적중 (exit 1) — §E.3 boundary_grep 재확인 |
| 14 | `go test -cover -count=1 ./internal/cli/update/... ./internal/config/...` | backup 86.6%·deploy 91.6%·merge 93.1%·plan 95.0%·report 92.9%·config 83.7% — §E.3 기재 수치와 항목별 일치 |

AC 커버: AC-001(a,b)·002(a,b)·003·004·005·006·007·008(b+셀프체크)·009·010·011·012·013·014·015·016·017·018·019·020(a)·021(a, b=프리뷰 패밀리) = **21/21 재관측**.

## 3. card-review 6건 — 코드 검증 (테스트 이름이 아니라 본문으로)

| # | 수정 | 코드 근거 (이 트리, 직접 판독) | 재실행 |
|---|---|---|---|
| F1 | 심링크 아카이브 대상 거부 | `archiveMigrationFile`(`update_migrate.go:215`)이 `rejectSymlinkedArchivePath`(`:278`)를 MkdirAll/WriteFile **이전** `:250`에서 호출. 거부부는 대상 경로에서 프로젝트 루트까지 **모든 부모 구성요소**를 `os.Lstat`로 순회(`ModeSymlink` 적중 시 `ARCHIVE_SYMLINK`). 픽스처가 대상 리프가 아닌 **부모**( `.moai/archive` )에 링크를 심는 것을 확인(`update_migrate_test.go:502-513`) — 부모 도달이 실측됨. 플로우 서브테스트가 제거 0·기록 미기재까지 단언 | `TestMigrationArchiveRefusesSymlinkedArchiveDestination` PASS |
| F2 | 마이그레이션 re-home | `template.RehomeExistingMirrorEntries`(`deployer_mode.go:238`)가 confirmed 팔에서 아카이브 루프(`update_migrate.go:154-159`) **후**·제거 목록 작성(`:185-190`) **전** `:169`에서 호출 — design §3 순서 준수 | `TestMigrationRehomesExistingMirrorEntries` PASS |
| F3 | heal 모드 게이트 | `repairSkillMirrorBestEffortAt`(`update_mirror_heal.go`)이 `config.ReadDeployMode(projectRoot) == "plugin"`이면 즉시 반환 — plugin 기록 프로젝트의 heal 부활 차단, local은 종전대로 | `TestMirrorHealRespectsDeployMode` PASS |
| F4 | preview/execute 공유 계산 | `previewManagedCleanup`(`update_dryrun_preview.go:38`)과 실행 Clean 스텝(`update_template_sync.go:485`)이 모두 `computeRunCleanTargets`(`update_migrate.go:91`)로 라우팅 — 이중 계산 소거 | `TestPreviewManagedCleanup_ModeScoped` PASS (+프리뷰 패밀리 ×6) |
| F5 | 비공허 보호집합 해시 | 스크립트 Pass 1이 `\( … \) -prune -o -print`로 미프룬 엔트리를 출력하고(`:118-135`), Pass 2가 `-o -type f -print0` + `shasum -a 256 /dev/null "$@"`로 **내용 바이트**를 해싱(`:142-170`, 심링크는 type l로 미해싱·미추적) | 셀프체크 `pass=6 fail=0` + 음성 대조 PASS |
| F6 | 재진입 가이던스 정확성 | `init.go:448` 가이던스가 `--no-plugin --force`를 명명하고 force 의미(「plain 재실행은 already-initialized로 실패」)를 기술 | `TestShrinkInitGuidanceOnMissingPlugin` PASS |

## 4. 마감 완결성 (close completeness)

| 항목 | 결과 |
|---|---|
| spec.md 상태 전이 | `git show c1623e65c -- .moai/specs/.../spec.md` = 정확히 2행 변경( `-status: in-progress` / `+status: completed` ) — 본문 드리프트 0. 현행 frontmatter `status: completed` 판독 확인 |
| §E.3 run audit-ready | `run_status: audit-ready` + `run_complete_at 2026-10-04T22:40+09:00` 판독. 백필 커밋 `6038f2ec5`(progress.md +15/−1) |
| §E.4 sync_commit_sha | `c1623e65c` 기재 — 백필 `a2b92451f`(progress.md 1행)로 D3 창 이행 확인 |
| CHANGELOG 정확성 | [Unreleased] Added 1 occurrence(grep -c = 1). 청구 3건 스팟 검증 — ① thin 기본 배포 제외 → 재실행 #1·#2 GREEN, ② 기록 영속 + update 모드 추종 → `TestDeployModeRecordRoundTrip` GREEN + `resolveUpdateDeployMode`(`update_template_sync.go:88`)이 `config.ReadDeployMode`로 판독, ③ 마이그레이션 idempotence·외부 보존 → `TestMigrationIdempotent`·`TestMigrationLeavesForeignFilesUntouched` GREEN |
| MX 델타 | `internal/config/deploy_mode.go:41`·`internal/template/apply_deploy_mode.go:38`에 `@MX:ANCHOR` 실재. 신규 파일 9종 `@MX:TODO` 0 |
| DoD 잔여 | decision-index.md `Operator verdict:` 8행(OD-1..8 전건) · REQ-008 판정의 §E.2 기재는 M2 플립의 develop 병합(미병합) 선행 조건과 합치 |

## 5. 보안 렌즈 — 마이그레이션 파일 연산

- **분류의 심링크 무역참조**: `ClassifyMigration`(`migrate_classify.go:172`)이 `fs.WalkDir`의 `d.Type()&fs.ModeSymlink`로 링크를 `plan.Symlinks`로만 기록(`:210-213`)하고 루트 자체도 `os.Lstat`(`:186-192`) — 링크는 분류·아카이브·제거 어디에도 들어가지 않음(REQ-013).
- **아카이브원 심링크 거부**: `archiveMigrationFile`이 소스를 `os.Lstat`로 검사해 링크면 `ARCHIVE_SYMLINK`(`:220-230`) — 링크가 역참조되어 아카이브되지 않음.
- **아카이브 대상 심링크 가드**: 부모 구성요소 순회 거부(§3 F1). 대상 루트는 프로젝트 상대 상수(`.moai/archive/…`, `migrate_classify.go:68-77`)라 트래버설로 프로젝트 루트 밖을 기입할 경로가 없고, 분류 경로는 프로젝트 루트 WalkDir의 clean 상대 경로라 `..` 주입 여지가 없음.
- **백업 선행·제거 중단 순서**: 아카이브 루프(`update_migrate.go:154-159`)가 단 건 실패 시 즉시 반환 — 제거 목록 작성(`:185`)과 Clean 집행( `update_template_sync.go:485` )에 도달하지 못해 "모든 아카이브 성공 후에만 제거"(OD-3)가 성립. F1 플로우 테스트가 제거 0·기록 미기재까지 실측.
- **글로벌 워크의 분류 게이트**: `computeRunCleanTargets`가 plugin 모드에서 드랍 루트를 글로벌 관리 워크에서 벗겨냄(`updateDroppedRootTargets`, `:73-83`) — P-08 백업 면제 함정(템플릿 수반 파일의 무백업 삭제)이 드랍 루트에서 기계적으로 봉쇄되고, 제거는 분류 집합만을 타는 déterministic 리스트가 됨.

## 6. 결함 목록 (S-1..S-5 — 전건 소스 검증 완료, 차단 0)

- **S-1** [P1→채무] [optional/비차단] `internal/cli/update_migrate.go:169` + `internal/template/deployer_mode.go:263-270` — **실디렉터리 미러 항목이 confirmed 마이그레이션에서 제거된다.** re-home은 **심링크만** 변환한다( `RehomeExistingMirrorEntries` — "a non-link entry is the user's and stays untouched" ), 반면 분류는 `.agents/skills`를 드랍 루트로 걸어( `migrate_classify.go:80-84` ) 템플릿-동일한 실카피 항목(P-12 codex-only 재배치 인구·과거 re-home 산출물)을 identical로 제거 목록에 넣고, plugin 모드 update는 미러를 재배포하지 않는다( `MirrorPolicyNone`, `update_template_sync.go:69-70` ). REQ-006의 무검증 폴백("the local mirror stays deployed on every path with its entries re-homed to real directory copies")과 acceptance.md 에지 케이스("re-homed real directories are in the dropped-root classification scope; loses nothing the plugin does not carry")가 **SPEC 내부에서 긴장**하며 구현은 에지 케이스 쪽을 택했다. Codex 실행 검증이 미생산인 현 시점( §E.2 OD-6 라우팅: "actual execution is not demonstrated" ) codex-only 인구의 로컬 Codex 스킬이 소실된다. 필수 수리: 실행 검증 착지 전까지 실카피 미러 항목을 보존하거나(제거 목록에서 미러 실디렉터리 제외), REQ-006 폴백 본문을 에지 케이스 쪽으로 명시 개정해 긴장을 소거. 완화: 제거 대상은 임베디드 렌더와 동일한 재생성 가능 사본이라 사용자 바이트 손실은 아님(OD-3(a) 논리).
- **S-2** [P2] [optional/비차단] `internal/cli/update_migrate.go:170-172` — **re-home 실패 팔에서 링크가 댕글로 남는다.** `.agents/skills` 쓰기 불가 등으로 `rehomeOneSkill`이 `MirrorModeFailed`를 반환하면 호출부는 경고만 하고 계속하고, 링크는 분류에서 심링크로 보존되지만 그 대상 `.claude/skills/<skill>`은 제거된다 — 결과는 REQ-006의 never-dangling 조항이 금지하는 **끊어진 링크 + plugin 기록**. 조건부 팔(비정상 권한 상태)이고 데이터 손실은 아니나 안전 계약의 무조건 문장에 대한 예외 상태. 필수 수리: `MirrorModeFailed` 시 해당 링크의 제거 대상(분류 identical 중 링크 대상 경로)을 이탈시키거나, 링크 엔트리를 삭제하고 그 사실을 출력.
- **S-3** [P3] [optional] `internal/cli/init.go` (강제 재초기화 경로) — **local→plugin 전환 재초기화가 기존 로컬 페이로드를 방치한다.** `init --no-plugin` 후 `init --force`는 기록을 `plugin`으로 바꾸고 thin 배포를 하지만 기존 `.claude/skills/**`·`.claude/commands/**`를 청소하지 않고(init.go에 CleanMoaiManagedPaths 호출 부재 — update 경로에만 존재), 이후 update의 thin 스코프가 해당 루트를 영구 제외한다. acceptance.md 에지 케이스가 init 재진입 전환을 "문서화된 스위치 표면 — 마이그레이션 결함 아님"으로 범위 밖 선언한 것과 맞물린 위생 격 차이(중복이지 손실이 아님). 권고: t1466 계열 후속으로 전환 시 분류·보존·제거 절차 또는 가이던스 문구 추가.
- **S-4** [P3] [optional] `scripts/check-bare-name-resolution.sh:118,126,145,153` — **보호집합 해시의 `-maxdepth 4`가 플러그인 캐시 내부(깊이 ≥5)를 해시 대상에서 뺀다.** 기존 캐시 존재 하의 깊은 파일 변조는 구조·내용 양팔 모두 미감지. 측정 하네스는 scratch `CLAUDE_CONFIG_DIR`로 격리되어 실홈 기입 자체가 이상 신호이고, 깊이 ≤4의 디렉터리 엔트리 신규 생성은 구조 해시가 잡는(신규 설치는 엔트리 생성을 수반) 절충값이나, 트립와이어 완결성 관점에서는 구멍. 권고: 플러그인 캐시 하위만 예외적으로 깊이 제한 해제 또는 대조 테스트 추가.
- **S-5** [P3] [optional] `acceptance.md` AC-021(b) — 소유 골든 테스트의 현행 이름 핀 부재. AC가 "TestUpdateDryRunPreviewGoldens (or the owning golden test's current name — M4 pins it)"로 위임했으나 §E.2 M4 표에 dry-run 프리뷰 소유 테스트의 핀이 없고, 실제 커버는 `update_observability_test.go`의 `TestPreviewManagedCleanup` 패밀리(재실행 GREEN)가 담당. 문서 완결성 관찰 — 동작 커버는 충분.

## 7. 증거 보고 (5-섹션 형식)

- **Claim**: SPEC-INIT-SHRINK-001의 21개 AC가 6038f2ec5에서 충족되고, card-review 6건이 코드로 수리됐으며, 3-phase 마감 아티팩트가 정확하다.
- **Evidence**: §2 표 14건 + §3 코드 판독 + §4 git show — 전부 이 턴, 이 트리, verbatim 출력 인용.
- **Baseline-attribution**: 모든 재실행은 `unset <13 MOAI_ 변수> &&` 세척 형태의 단일 복합 호출, `-count=1`, HEAD `6038f2ec594dc7ec5490137ddf2d8a2d799429a3`에서 수행. lint는 CI 동일 버전 v2.1.6. §E.3의 미재측정 수치(internal/cli·template 집계 커버리지 %)는 재인용하지 않았고, 재측정한 6개 서브패키지 수치만 대조 인용했다.
- **Gaps**: ① REQ-008의 실측 런타임 계측( `RESULT pass=5 fail=1` 본 실행)은 재실행하지 않았다 — 실 claude/codex 세션 런타임이 필요하고, 할당된 처분( fail=1은 마커 기제, 해소 자체는 세션 로그로 실증 )을 수용하고 하네스 셀프체크+음성 대조만 재관측했다. ② internal/cli 전체 스위트(-timeout 40m) 재실행 없음 — 4-failures 처분을 수용, 플립 관련 패밀리 전건을 셀렉터 재실행으로 대신했다. ③ 세 플랫폼 매트릭스·원격 CI는 리더 일괄 push 후 origin/develop의 몫이다.
- **Residual-risk**: S-1의 codex-only 인구(실행 검증 전 미러 실카피 소실) · S-2의 권한 이상 상태 팔 · 바이너리 래그(수신지가 보고한 설치 바이너리 `0732cc699` — 조상 커밋; 본 감사의 근거는 go test 직접 실행이라 영향 없음) · `TestCodex1718Fixtures_WidenedContent`는 develop 흡수 창에서 해소 예정(해소 불확인 시 에스컬레이션).

## Cross-model receipt

- 도구: `codex_audit` (mode `native`, target `baseBranch`, `project_root` = 본 워크트리 절대경로) — stdio JSON-RPC 드라이버( `sh /tmp/t1438_sa_driver.sh`, initialize → notifications/initialized → tools/call id=10, stdin 유지 ). 요청 3행: `/tmp/t1438_sa_req.jsonl`, 원응답: `/tmp/t1438_sa_out.jsonl` (id=10 라인).
- 결과(2026-10-04): `audit_receipt: rcpt-8ebf466c844bc096ac5b901c`, codex verdict `fail` — findings 4건(P1×1, P2×3), `build_commit 0732cc699` + binary-lag 경고 동봉.
- 소스 검증: 4건 전건을 본 트리 코드로 재판정했다 — S-1(실카피 미러 제거: `rehomeOneSkill`의 비링크 skip + 분류 루트 포함 + `MirrorPolicyNone`으로 코드 확약), S-2(실패 팔 댕글: `MirrorModeFailed` 경고-계속 경로), S-3(init 무청소: init.go의 Clean 호출 부재), S-4(maxdepth 4: 스크립트 리터럴 확인) — 전건 사실관계 **인정**, 전건 비치명(critical 유저 파일 손실 없음, AC 층 전부 녹색)으로 판정해 수렴 규칙(중대 결함 없음 → PASS-WITH-DEBT)을 적용했다. codex의 `fail` 토큰은 발견 존재 시 항상 fail을 내는 도구 시맨틱이며, 본 감사의 수렴 판정은 소스 검증된 중대도 기준이다.

verdict: PASS-WITH-DEBT 93.5 — receipt: codex_audit rcpt-8ebf466c844bc096ac5b901c, converged

## Leader disposition (recorded by the lane, 2026-10-04)

Debt S-1..S-5 accepted as NON-BLOCKING by the leader (mission contract 11c79e1a relay, window nomination message): "채무 S-1~S-5는 비차단으로 받아들이고 판정서 기록만 남겨 주세요." This file is committed at the leader's instruction before the integration window opens (uncommitted verdict = evidence gap).
