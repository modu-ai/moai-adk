# progress.md — SPEC-INIT-HARNESS-001

## §E.1 Plan-phase Audit-Ready Signal

- card: t585 · tree: `.claude/worktrees/t585` · branch: `WT-init-harness-q` · HEAD at authoring: `a404132e7` (tree `aebfa5bea`)
- tier: L · artifacts: 6 (spec / plan / acceptance / design / research / progress)
- SPEC ID 자가검증: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS` (2026-09-14 실행, verbatim `PASS`)
- ID 충돌: `.moai/specs/` 디렉터 조회 0건 + 카탈로그 grep 0건 (SPEC-INIT-HARNESS-001)
- frontmatter: 12 캐노닉 필드 전부 + tier/depends_on/related_specs 옵션 — 스네이크 케이스 별칭 0
- depends_on 사전 점검: SPEC-INIT-QUIET-WIZARD-001 `status: completed` (2026-09-14 frontmatter 직접 판독) — 성립

### FO-PLAN-1 (병렬 리서치 팬아웃) 스킵 기록

레인 결정으로 FO-PLAN-1 병렬 조사 팬아웃은 **스킵** — 표준 단일 에이전트 디스커버리(manager-spec 자체 Grep/Read)로 수행. 사유 3개: (1) 리드의 이 공유 코드 표면에 대한 직렬 지시, (2) 카드가 전수 사전 조사를 이미 보유(`.moai/reports/init-tui-audit-20260909.md` — 표면 소속 지도·결함 17건·운영자 결정 포함), (3) 조사 표면이 좁다(배포기·미러·위저드·update·doctor의 하니스 축 한 지점). 대가로 모든 file:line을 이 트리 HEAD `a404132e7`에서 재측정했고 카드 전제 13건 중 2건 반증·1건 드리프트를 잡았다(research.md §1).

### Decision Point 1 (plan 후보 검토 게이트) 레인 노트

DP1은 리드 디스패치 승인 + 카드 본문(운영자 승인 조사 C3의 배출 의도)으로 레인 측 충족 — 본 SPEC의 후보 방향(harness 3-way 배포, codex 단독 본체)은 카드 텍스트가 곧 운영자 의도다. 이 SPEC에 대한 운영자 게이트는 본 산물 + plan-audit 이후의 Implementation Kickoff Approval이며 manager-spec이 요청하지 않는다.

### plan-phase 판단 메모 (리드 보고 항목)

- 카드 전제 반증 2건: commandemit 16종 착지됨(`e7d2a1658`, t503) — both 모드 신규 공사 소멸 / AGENTS.md 결속표 3행 존재(t523 `0755cc7f5`) — 조사의 "0행"은 낡은 측정.
- 명칭 드리프트 1건: 플래그는 `--agent`가 아니라 `--llm`(SPEC-CODEX-WIRING-001) — AC-CW-004 핀 축은 동일.
- 미해결 2건(운영자 소유 — plan.md §I): `--llm codex` 값 의미 변경 수용 여부 / claude 단독 트리밍(디스패치 문구 vs 조사 운영자 결정 "claude 단독=현행" 불일치 — 본 SPEC은 조사 결정 채택).

### iter1 개정 (plan-audit FAIL 0.94 → 수리)

audit iter1 FAIL 0.94 → D2(REQ-IH-003 바이트 동일 모순 → 파일집·로직 보존으로 재작성, REQ-IH-002/008 제외 명시 + AC-IH-005 정렬)+D3(가드 앵커 internal/config/token_budget_guard.go:94 정정)+D4(AC-IH-008 단일 호출형 + 실측 verbatim)+D5(스킬 로더 부분 공개 판정선 추가)+D6(REQ-IH-002 M1-산물 서술) 수리 완료, D7 관례 적합 무수리; MP-7 NC-1/NC-2는 리드 경유 운영자 대기.
NC-1·NC-2 운영자 확정(2026-09-14, 리드 경유) — NC-1 채택(codex-only deployment)·NC-2 현행 유지, plan.md §I 기록 + 본문 마커 RESOLVED 해제; plan_status 유지.

plan_complete_at: 2026-09-14T20:24:31+09:00 (plan 산물 검증 시점 실측 `date` 값)
plan_status: audit-ready

## §E.2 Run-phase Evidence

- run 실행: 2026-09-14, 카드 워크트리 `.claude/worktrees/t585`, 브랜치 `WT-init-harness-q`, plan 착지 HEAD `4564af4e9` 위에서 M1-M5 직렬 실행 (plan.md §F serial 모드 준수)
- **홈 지문 규약 (t583 REQ-IQW-014/016 준용)**: run 단계의 모든 init/update 실행 테스트는 `runInitForAutonomy` / `runInitForAutonomyAtHome`(internal/cli/init_autonomy_wiring_test.go) 경유로 HOME=`t.TempDir()` 격리 + `MOAI_SANDBOX_PROOF`·`MOAI_DISABLE_BYPASS_PERMISSIONS_MODE` scrub. 실제 HOME에는 단 한 번도 쓰지 않았다. update 테스트(update_harness_test.go)는 동일 헬퍼로 init 후 프로젝트 디렉터로 chdir. 슬롯 임대: `moai slot acquire --resource cli-suite-t585` (2026-09-14 21:33·22:33 두 차례 획득 — 무거운 cli 스위트 앞뒤로)
- **E8 — RED 관측 (전부 본 런 verbatim, 구현 전 트리)**:
  - M1 RED (tree `4564af4e9`): `go test ./internal/cli/ -run 'TestInitPersistsHarnessKey' -count=1` → `--- FAIL: TestInitPersistsHarnessKey/flag_absent_records_claude ... llm.harness after init = "", want "claude"` (4 서브테스트 전부) + `--- FAIL: TestInitPersistsHarnessKeyFromWizard ... = "", want "codex"`. 빨간 이유: init에 llm.harness를 쓰는 코드가 전무 — M1의 첫 변경이 정확히 이 키를 씀
  - M2 RED (tree `24bb24c18`): `go test ./internal/cli/ -run 'TestInitCodexOnly...' -count=1` → `TestInitCodexOnlyDeploysNoClaudeSurface`: `.claude exists`, `CLAUDE.md exists`, `.mcp.json exists`, `.claudeignore exists`, `.moai/status_line.sh exists` — 5개 부정 단정 전부 위반 (오늘 `--llm codex`가 claude 전면 배포). `TestInitCodexOnlyRequiredSurfaces`: 재매핑 카탈로그 스킬(`moai-domain-backend` 등 non-core 13종) 심볼릭 링크 부재로 FAIL
  - M3 RED (tree `5a35690a2`): `TestAgentWiringOptions` → `option codex description does not describe deployment` 계열 7건; `TestAgentsDisclosureCompleteness` → `missing disclosures: [agent-spawning output-style slash-commands workflow-scripts skill-loader-non-equivalence]` — acceptance.md 예측과 동일한 5건
  - M4 RED (tree `1fd484643`): `TestUpdateCodexOnlyNoClaudeResurrection` → `.claude/ resurrected by update`, `CLAUDE.md resurrected` / `TestUpdatePreservesHarnessKey` → `llm.harness after update = "claude", want codex`. 빨간 이유: update 재배포 경로에 하니스 인지가 없고, 배포가 claude 템플릿 기본값으로 덮음
  - AC-IH-008 RED: plan 단계 관측치 그대로 — `grep -c "Codex only" README.md` → `0`, exit 1 (tree `a404132e7`, acceptance.md §B 기록)
- **E1 — AC-IH-001..016 매트릭스 (최종 상태, 본 런 관측)**:

| AC | 상태 | 검증 (command → 관측) |
|---|---|---|
| AC-IH-001 | **PASS** | `go test ./internal/cli/ -run 'TestInitPersistsHarnessKey' -count=1` → `ok` (flag-absent/claude/codex/both 4케이스 + wizard 경로) |
| AC-IH-002 | **PASS** | `go test ./internal/cli/ -run 'TestInitCodexOnlyDeploysNoClaudeSurface' -count=1` → `ok` (5 부정 단정) |
| AC-IH-003 | **PASS** | `go test ./internal/cli/ -run 'TestInitCodexOnlyRequiredSurfaces' -count=1` → `ok` (AGENTS.md·배선 3종·TOML 11·게시 16·재매핑 카탈로그 전수·llm.yaml·gitignore) |
| AC-IH-004 | **PASS** | `go test ./internal/template/ -run 'TestCodexOnly' -count=1` → `ok` (listing 무은닉 누설, 실디렉터·symlink 0·SKILL.md 판독·카탈로그 전수, REQ-IH-007 점유 skip) |
| AC-IH-005 | **PASS** | `go test ./internal/cli/ -run 'TestRunInit' -count=1` → `ok` — 기존 claude init 실행 테스트 본문 무변경 green (`git diff`에 해당 파일 없음) |
| AC-IH-006 | **PASS** | `TestRunInit_AgentBothWiresBothSides|TestRunInit_CallsCodexWiring` → `ok` — both 배포에 신규 필터 없음 |
| AC-IH-007 | **PASS** | `go test ./internal/template/ -run 'TestAgentsDisclosureCompleteness' -count=1` → `ok` (9 마커 전수 + 24,576B 상한 이내) |
| AC-IH-008 | **PASS (README)** / docs-site는 sync 위임 | `grep -c "Codex only" README.md` → `1`, exit 0; 4-locale 동기 (ko/ja/zh 동일 절). docs-site 표면은 design.md D8이 sync 단계 실행으로 명시 — manager-docs 인계 |
| AC-IH-009 | **PASS** | `go test ./internal/cli/ -run 'TestUpdateCodexOnlyNoClaudeResurrection' -count=1` → `ok` (부활 0 + codex 표면 갱신) |
| AC-IH-010 | **PASS** | `go test ./internal/cli/ -run 'TestDoctorCodexOnly|TestDowngradeLeaves' -count=1` → `ok` — claude 표면 7검사 INFO 강등, OK/codex 검사 무변경, 기존 `TestDoctorGolden*`·`TestRunHarnessCheck*` 무변경 green |
| AC-IH-011 | **PASS** | `go test ./internal/cli/ -run 'TestInitCodexNonInteractiveParity' -count=1` → `ok` (파일집 diff 0; 런타임 `.moai/state`·`.moai/db/<path-hash>` 제외는 테스트 주석 명시) |
| AC-IH-012 | **PASS** | `go test ./internal/cli/wizard/ -run 'TestAgentWiring' -count=1` → `ok` (3옵션·값 동결·배포 서술·ko/ja/zh 동기) |
| AC-IH-013 | **PASS** | `TestInitAgentFlag*` + closed-set + absent/claude/both → `ok` — init_agent_flag_test.go 본문 무변경 |
| AC-IH-014 | **PASS** | `go test ./internal/cli/ -run 'TestToolEnableCodex' -count=1` → `ok` — tool.go 미접촉 (diff에 부재) |
| AC-IH-015 | **PASS** | `go test ./internal/cli/ -run 'TestUpdatePreservesHarnessKey' -count=1` → `ok` (init codex → update → `codex` 생존) |
| AC-IH-016 | **PASS** | 아래 E2/E3/E5 + cli 전체 스위트 (§E.3) |

- **테스트 본문 변경 2건 (의도된 렌더 핀 — 보존 계약 아님, blocker 아님)**: `TestInitRegroup_SecondGroupGolden` golden 파일은 REQ-IH-013 Title/Desc 변경의 직접 표면이라 갱신(-update-golden), `TestGroupLabel_NotRendered`는 같은 Title 문자열을 어설션 핀으로 갖는 렌더 테스트라 문자열만 갱신. 둘 다 AC-IH-005/006/013이 지키는 "claude 배포 파일집·로직 보존" 또는 AC-CW-* 계약 테스트가 아니며, 본문 어설션 구조는 무변경. E6 커밋에 수반 기록
- **AGENTS.md 예산 수리 기록**: 결속표 보강이 `TestCodexNestedTemplateDiscoveryBudget`(repo-root + template AGENTS.md 합계 ≤ 32,768B)을 490B 초과 → 보강 문구 3회 압축으로 32,768B 이내 회귀 (CodexContractByteCeiling 24,576B는 전 기간 준수). REQ-IH-008 요구 마커 9건은 압축 후에도 전수 유지
- **cli 전체 스위트 관측 (정직 기록)**: 1차 실행(20m 타임아웃) — hang 원인 `TestTodoUndone_EmptiesTheArchiveEntry`가 `[syscall]` 상태로 20분 스톨 (본 SPEC diff 밖, todo_autodone.go — 내 변경 파일 아님). 단독 실행 2.4s 통과 → 테스트 자체 결함 아님, 전체 실행 중 환경 수준 I/O 스톨로 판별. 2차 재시도(30m) — hang 없이 완료, 1건 실패(`TestUpdateLLMYAMLFirstDeployCalm` — 템플릿 verbatim 계약) 발각 → `ApplyHarness` no-op 수리(별도 커밋) 후 단독 PASS + harness persistence 계약 재검증 PASS. 3차 실행(40m 타임아웃, `-cover` 동반) — **전체 PASS, exit 0, coverage 83.6% (1249s)**. 최종 상태: cli 전체 스위트 green 본 런 관측 + CI 재판정은 origin/develop push 시 (lane 규율)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-14T23:30:00+09:00
run_commit_sha: "3f7ef4fed"
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
ac_pass_with_debt: 1   # AC-IH-008 — README PASS, docs-site 절은 design.md D8에 따라 sync 단계(manager-docs) 인계
preserve_list_post_run_count: 6   # resolveAgentWiringWithWizard, codexwiring 3종, mirrorOneSkill 본문, R-011 게시 보호, slimFS, reconfigure 12문항
l44_pre_commit_fetch: "n/a — lane protocol: commits on WT branch are lane-owned; no push (lead batch-pushes develop)"
l44_post_push_fetch: "n/a — NOT PUSHED (lane protocol §4: lane never pushes)"
new_warnings_or_lints_introduced: 0   # golangci-lint run internal/cli/... internal/template/... internal/config/... internal/core/project/... → "0 issues."
cross_platform_build:
  native: "go build ./... → exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... → exit 0"
total_run_phase_files: 25
m1_to_mN_commit_strategy: "per-milestone commits M1..M5, conventional, card id + Authored-By-Agent trailer, SPEC draft→in-progress on M1 (24bb24c18)"
```

- E2: `go build ./...` → exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (본 런, HEAD `3f7ef4fed`)
- E3: 영향 패키지 커버리지 (본 런, `go test -cover`) — config **81.7%** / core/project **88.9%** / cli/wizard **93.6%** / cli **83.6%** / template 전체 **81.8%**, **신규 코드 기준**: harness_fs.go + ApplyHarness 함수 전수 최저 Stat 80%·ApplyHarness 84.2%·신규 파일 평균 ~88% (85% 목표 충족 — apply_harness_test.go가 저커버리지 분지 전부 메움)
- E4: 서브에이전트 경계 grep — run-phase 신규·변경 Go 파일(internal/cli, internal/template)에 AskUserQuestion 매치 **0** (`git diff 4564af4e9..HEAD --name-only | xargs grep -ln AskUserQuestion` → SPEC 문서·README만, Go 소스 0). 참고: `grep -rn 'AskUserQuestion' internal/cli internal/template | grep -v _test.go | grep -v '// '`는 512매치 — 전원 기존 doc 주석·문자열 리터럴·가드 함수명(checkLiteralAskUserQuestion)이며 baseline부터 존재하는 것으로 본 라인은 호출-구문 판정식이 아니라 문서 문자열까지 세는 관측치 (실측치 기록, VCI §2 귀속)
- E5: lint — baseline(tree `4564af4e9`, "0 issues") 대비 NEW 0
- E6: 커밋 — `24bb24c18`(M1) `5a35690a2`(M2) `1fd484643`(M3) `90aafcf1b`(M4) `3f7ef4fed`(M5) · **push 상태: NOT PUSHED (레인 프로토콜 — 리드 일괄)**
- E7: blocker — 없음 (AC-IH-008의 docs-site 부분은 design.md D8이 plan 단계에서 정한 sync 인계이며, run-phase 신규 장애가 아님)

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs가 sync_commit_sha와 함께 기록>_

## §F Phase 4 Mode Selection

- 입력 파라미터: tier=L · scope≈12파일(internal/cli/wizard/{questions,types,translations}.go, internal/cli/init.go, internal/config/defaults.go, internal/template/{deployer,skill_mirror}.go, templates AGENTS.md·llm.yaml, 신규·갱신 테스트) · 도메인 3(CLI 위저드 / config / 템플릿 배포기) · 언어 혼합 Go+YAML+Markdown · concurrency benefit=LOW(coding-heavy) · agent-team 사전요건=미요청(n/a)
- 모드 평가: direct=미선정(다중 파일 의미 변경) / serial=**선정** / fanout=미선정(coding-heavy — Anthropic caveat + 리드 직렬 지시) / sweep=미선정(기계적 균일 변환이 아닌 신규 코드 작업)
- Decision: serial
- 정당화: codex 단독 배포 필터는 단일 표면(deployer+harnessFS 래퍼)에 걸린 신규 코드 작업이다. manager-develop 단일 스폰으로 plan.md M1-M5 마일스퀀스를 직렬 실행한다 — Anthropic의 coding-task 직렬 기본과 리드 디스패치의 직렬 지시(같은 init 마법사 표면 병렬 금지)가 모두 serial을 가리킨다.
- Boundary case: 없음(serial 폴백 도달)
- Implementation Kickoff Approval: PASSED 2026-09-14 — 리드 경유 운영자 확정 2건(plan.md §I: NC-1 `--llm codex` 재정의 수용 / NC-2 claude 현행 유지) + 리드의 run 직행 지시(fanout 일괄 사전승인 정책 — 개별 Kickoff 질문 생략 선언) + plan-audit iter2 PASS 1.00(`.moai/reports/t585/plan-audit-iter2.md`, 로컬).
- Plan Audit Gate skip 판정: verdict PASS · score 1.00 ≥ 0.85(Tier L) · plan-artifact 해시 불변(iter2 감사 이후 해시 대상 5파일 무변경 — §F/§E.*는 해시 대상 아님) 3조건 충족 → run Phase 1 재실행 skip.
