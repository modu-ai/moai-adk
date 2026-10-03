# SPEC-CODEX-GATE-SCOPING-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03
tier: M
cycle_type: tdd
artifacts: spec.md, plan.md, acceptance.md, decision-index.md (본 파일 포함 5본)
evidence_copy: .moai/reports/t1404/gate-block-evidence.md (primary 원본 t1395 처분 기록 9건의 본 카드 필요분 사본)
plan_audit: iteration-1 FAIL(D1-D5) 수리 v0.1.1 → iteration-2 FAIL 0.8625(D6-D10, `.moai/reports/t1404/plan-audit-iter2.md`) → 리더 판정 범위 축소 후 3회차 수리 완료(v0.1.2, `.moai/reports/t1404/plan-decision-iter2.md` 지시 범위 한정) — 재판정 대기. v0.1.2 수리 요지: AC-010 대조군 재정의(비-reports 파손 변경 게이트 — D6), AC-012 레이아웃 독립 분할+AC-011 이관(D7), AC-001 행 내용·AC-008 재분류 행 단언·②팔 실입력·AC-007 settings.local.json 삭제(D8), 삼클래스 문언 정렬(D9), §D.0 핀 물리 상태 보충(D10). R 5건 전부 재관측(HEAD 9ef1cbedc + 개정 테스트 파일).
red_tests: internal/cli/codex_review_gate_primary_scope_red_test.go(AC-001·005·007·008) · internal/template/hook_gate_reports_exclude_test.go(AC-010) — plan 단계 저작, iteration-2 D6·D8 개정 후 전부 재관측 적색(acceptance.md §D.0 — 개정 blob 5779dd07·1aa630c3). manager-develop가 M2/M3/M4에서 GREEN 전환(새로 저작하지 않음).

## §E.2 Run-phase Evidence

실행 트리: 카드 워크트리 `.claude/worktrees/t1404b`, 브랜치 `WT-codex-gate-scope`, 기준 HEAD `45a397e72`. 커밋: M1 `697f18570`(설정면 + `draft → in-progress` 전이, `Authored-By-Agent: manager-develop` 트레일러) → M2 `912773c64`(Facet 1) → M3 `dc52d74cb`(Facet 2) → M4 `dc68de6d1`(Facet 3 — 미러·추적본 동반 한 커밋, `make build` 임베드 재컴파일). push 없음(레인 통합 몫).

### GREEN 전환 (acceptance §D.0 장부의 RED 원문 대응 — E8 재확인 축)

- AC-CGSC-001: `go test -count=1 -v ./internal/cli/ -run '^TestCodexReviewGatePrimaryCheckoutSkip$'` → `--- PASS: TestCodexReviewGatePrimaryCheckoutSkip (2.22s)` · `ok github.com/modu-ai/moai-adk/internal/cli 5.514s`. 스킵 행 1건(row.Dir=트리, basis가 primary 사유+트리 경로 운반)·lookups=0·detects=0·무리뷰 — RED-CGSC-001의 세 적색 행 전부 반전.
- AC-CGSC-005: `--- PASS: TestCodexReviewGatePrimaryPolicySharedByBothPaths (1.69s)` — 게이트 경로(행 1·무리뷰)와 체인 경로(Allow/not-applicable/tree_scope 사유/무수신) 동일 정책 통과.
- AC-CGSC-007: `--- PASS: TestReviewableFromPorcelainRuntimeConfigOnlyFalse (0.00s)` — 두 설정 표면 무검사 + `cmd/moai/main.go` 대조군 검사 가능 유지.
- AC-CGSC-008: `--- PASS: TestCodexReviewGateRuntimeDriftFindingsReclassified (0.10s)` — 설정 표면 전용 발견 fail 판정이 ALLOW로 재분류되고 채널 행이 `runtime-managed` 사유와 `.claude/settings.json` 대상 경로를 운반.
- AC-CGSC-010: `go test -count=1 -v -timeout 10m ./internal/template/ -run '^TestSyncPhaseGateExcludesReportsGoFixture$'` → `--- PASS: TestSyncPhaseGateExcludesReportsGoFixture (9.29s)`, 5 서브테스트 전부 PASS — 미추적 팔·tracked ①② 팔·미러 양 팔에서 reports 픽스처 미수집, 대조군(루트 파손 변경)은 계속 게이트. 재실행 `ok ... 10.064s`(부하 하 안정성).

### G·P2 관측 (구현 전후 동일 출력 — baseline `45a397e72` 대비)

- AC-CGSC-002(연결 워크트리): `--- PASS: TestCodexReviewGate_WTBranchWithoutBaseReviewsWholeTree (2.30s)` — WT- 무베이스 트리 세션은 primary 스킵 미적용, 전체 트리 리뷰·"merge base unavailable" 스코프 행 유지. baseline 동일 테스트 초록.
- AC-CGSC-003(카드 스코프): `--- PASS: TestCodexReviewGate_CardScopeRequestIsCardDiff (1.38s)` · `--- PASS: TestTreeScopeSkip_WTSessionsStillReviewed (3.77s)` — 카드 diff 대상·수신 경로·정책 행 0 pre-SPEC과 동일.
- AC-CGSC-009(공유 리스트 불변): `reviewGateRuntimePrefixes` 6항목 원문 불변(diff 0) + 위 카드 경로 테스트 초록으로 귀속.
- AC-CGSC-012(수집 shape-identical): 기존 `hook_gate_*`·sync 게이트 테스트 `ok github.com/modu-ai/moai-adk/internal/template 76.485s` + state 항목 존재 grep 쌍둔 각 1(baseline과 동일).
- AC-CGSC-004(P2 복원): `go test -v ./internal/config/ -run '^TestNormalizeCodexReviewGatePrimaryScope$'` → `--- PASS: TestNormalizeCodexReviewGatePrimaryScope (0.00s)`(4테스트 `ok ... 0.331s`) + 게이트 레벨 복원 `--- PASS: TestTreeScopePolicy_SameDecisionOnBothPaths (5.94s)` — 노멀라이저는 미지 값을 review가 아닌 skip으로.
- AC-CGSC-006(P2 페일오픈): `--- PASS: TestCodexReviewGateNonGitDirNoPrimarySkip (0.18s)` — 비-git 트리에서 자기게이트 1회 실행·정책 행 0·프로브 리뷰 BLOCK(기존 동작) 유지.
- AC-CGSC-011(P2 쌍둔): `grep -cF ':(top,exclude).moai/reports'` 양 쌍둔 각 1(baseline 0 → ≥1 달성) + 쌍둔 diff 바이트 동일 + 임베드 재컴파일(`make build` exit 0).

### 사전 적색 분류 (baseline — HEAD `45a397e72`, 본 실행)

`unset <레인 env> && go test -count=1 -timeout 30m ./internal/cli/ ./internal/config/ ./internal/template/`: cli `--- FAIL` 4건(RED-CGSC-001·005·007·008 — 계획된 RED) + template `--- FAIL` 1건(RED-CGSC-010 — 계획된 RED) + config `ok 6.443s`. **계획된 RED 5건 외 적색 0** — 회귀 baseline 확정.

### 소관 패키지 재측정 (단위 = 패키지 전체)

`unset <레인 env> && go test -count=1 -timeout 20m ./internal/cli/ -run '<게이트 표면 셀렉터>'` — 4회 실행 전부 `ok`(180.6s → 177.7s → 117.0s → 103.9s, 유일 적색이었던 RED-CGSC-008은 M3 후 반전). `go test -count=1 ./internal/settings/` `ok 1.498s`. 정적: `go vet` 4패키지 0 · `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · `golangci-lint run --timeout=10m`(v2.1.6 = CI 판) `0 issues.` · gofmt 0 · `./bin/moai spec lint SPEC-CODEX-GATE-SCOPING-001` 무발견.

- 커버리지(변경 함수, 게이트 표면 프로파일 — `go test -coverprofile ./internal/cli/ -run '<게이트+porcelain 셀렉터>'`): `HandleCodexReviewGate` 100% · `treeScopeSkipApplies` 100% · `treeScopeSkipRow` 100% · `isTreeRuntimeConfigPath` 100% · `runtimeConfigOnlyFindings` 90.9% · `logRuntimeDriftReclassification` 80.0% · `resolveReviewScope` 93.3% · `isPrimaryCheckoutGit` 77.8% · `reviewScopeGitPath` 100% · `readCodexReviewGatePrimaryScope` 77.8% · `reviewableFromPorcelain` 83.3% · `hasReviewableChanges` 100% — 신규·수정 함수 전부 77.8% 이상, 주요 판정 함수 90-100%.
- 패키지 집계: `config` 83.6% · `settings` 86.8% (측정 — 본 실행, M4 후 트리). `cli`·`template` 패키지 집계는 동시 레인 부하로 커버리지 실행이 CPU 굶주림(듀티 2-16% 실측)에 걸려 미측정 — 부하 완화 후 `unset <레인 env> && go test -count=1 -cover ./internal/config/ ./internal/settings/ ./internal/cli/ ./internal/template/` 재측정. `cli` 변경파일은 위 게이트 표면 프로파일로 판정 보강, `template`은 본 SPEC의 Go 코드 변경이 0(셸 스크립트만 — 스크립트 계약 AC-010/011/012 전부 초록). 85%는 가이드이며 config 83.6% 공백은 본 카드 미접촉 기존 표면.
- E4 경계 grep: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli internal/config internal/settings --include='*.go' | grep -v _test.go | grep -v '// '` = 18행 — **baseline `45a397e72` 동일 형태 18행과 동일(신규 0)**. 전부 기존 스트링 리터럴 헬프 텍스트(harness.go·pr_watch_cmd.go)이며 본 카드 변경 파일 적중 0. C-HRA-008 정적 가드 테스트 초록 유지.
- 미러 중성성: 신규 주석 카드 id 0(`grep -c t1404` = 0), 기존 `t1392` 6건 불변.
- 범위 침범: `git diff 45a397e72..HEAD --stat` 16파일 전부 plan §A.5 범위(설정 3·cli 9·게이트 스크립트 쌍둔 2·spec.md frontmatter 2행·config 테스트 1) + 테스트 적응 5파일. 디버그 잔재 스캔 0.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-10-03"
run_commit_sha: "dc68de6d18ec495ffba02e7b3da47421b36602b9"
run_status: "complete"
ac_pass_count: 12
ac_fail_count: 0
red_first_flip: "R 5건 전부 GREEN(§D.0 장부 원문과 E8 재확인 일치) — M2(001·005) → M3(007·008) → M4(010)"
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "not-run (isolated card worktree — commits land on WT-codex-gate-scope; lane owns integration)"
l44_post_push_fetch: "not-run (no push by design — leader owns develop push)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: "go build ./... exit 0; go vet 0; golangci-lint v2.1.6 0 issues; gofmt 0"
  windows_amd64: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
coverage_4_packages: "config 83.6% · settings 86.8% (측정) · cli·template 패키지 집계는 레인 부하로 미측정(부하 완화 후 재측정) — 변경함수 프로파일로 판정 보강(신규·수정 함수 전부 77.8%+, 판정 함수 90-100%)"
total_run_phase_files: 16
m1_to_mN_commit_strategy: "M1 697f18570 -> M2 912773c64 -> M3 dc52d74cb -> M4 dc68de6d1 -> M5 progress record"
fixture_adaptations: "기존 테스트 10지점이 pre-SPEC 리뷰 기본값을 고정하고 있어 REQ-CGSC-004 복원 픽스처로 적응(ownership·scope·stop_fixture·live·selfreview) — 계약상 5 RED 테스트는 무변경"
decision_index_handoff: "Q1·Q3 리더 DECIDED 기록 완료, Q2 EVIDENCE-NEEDED — leader-owned 갱신 가능 상태로 인계(본 파일 미변경)"
ci_verdict: "PENDING — the repository-wide test verdict belongs to the CI run on the integration branch after the lane merges; this report claims the owning-package runs only"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_

## §F Phase 4 Mode Selection

- 입력 파라미터: tier=M · scope=7 소스 파일 + 7 테스트 파일(예상) · 도메인 수=2(internal/cli·internal/config Go + 훅 스크립트/템플릿) · 파일 언어 혼합=Go+shell·markdown · 병렬 이득=LOW(코딩 중심 — Anthropic 코딩 과제 병렬화 주의) · Agent Teams 전제=미충족(명시 요청 없음)
- 모드 평가: direct=미선정(단일 자리수 자리표 이상, 의미 변경 있음) · fanout=미선정(코딩 중심, 도메인 2개로 3 미달) · sweep=미선정(30파일 미달·의미 변형 작업) · agent-team=미선정(명시 요청 없음)
- **Decision: serial** — 마일스톤당 1회 서브에이전트 순차 배차
- 근거: 코딩 중심 작업은 Anthropic 지침상 serial이 기본이고, 마일스톤 M1-M6이 config→handler→scope→템플릿 순으로 의존하는 단일 사슬이라 병렬 분해 이득이 없다. RED 테스트는 plan 단계에서 이미 저작돼 있어 manager-develop는 GREEN 전환 소관(새로 저작하지 않음).
- 경계 메모: M4 착수 전 리더의 C5(pathspec 범위) 한 줄 확정 대기 — C5 답이 오기 전 M4 진행 보류. C1-C7 부채 처분(강화 vs 후속 카드)도 리더 답과 함께 run 배차 범위에 반영.
- C5 처분(2026-10-03 리더): ①②팔 모두 reports 항목만 건다(WCI_EXCLUDES 전체 아님) — M4 보류 해제. C1-C7은 run 범위 밖(후속 카드). 상세: .moai/reports/t1404/kickoff-decision.md 덧붙임.
